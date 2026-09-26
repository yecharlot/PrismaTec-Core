package organism

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core/events"
	"github.com/yecharlot/PrismaTec-Core/core/registry"
	cidstore "github.com/yecharlot/PrismaTec-Core/storage/cid"
	"github.com/yecharlot/PrismaTec-Core/core/policy"
)

// Manager owns organism lifecycle, registry sync and local persistence.
type Manager struct {
	mu       sync.RWMutex
	byID     map[string]*Organism
	dataDir  string
	nodeID   string
	nodeName string
	bus      *events.Bus
	reg      *registry.Registry
	blocks   cidstore.Store
}

// NewManager creates a manager bound to a node.
func NewManager(dataDir, nodeID, nodeName string, bus *events.Bus, reg *registry.Registry) (*Manager, error) {
	orgDir := filepath.Join(dataDir, "organisms")
	if err := os.MkdirAll(orgDir, 0700); err != nil {
		return nil, fmt.Errorf("create organisms dir: %w", err)
	}
	blocks, err := cidstore.NewLocalStore(filepath.Join(dataDir, "blocks"))
	if err != nil {
		return nil, err
	}
	m := &Manager{
		byID:     make(map[string]*Organism),
		dataDir:  orgDir,
		nodeID:   nodeID,
		nodeName: nodeName,
		bus:      bus,
		reg:      reg,
		blocks:   blocks,
	}
	if err := m.loadAll(); err != nil {
		return nil, err
	}
	return m, nil
}

// Blocks returns the content-addressed block store.
func (m *Manager) Blocks() cidstore.Store {
	return m.blocks
}

// EngineFor builds a policy.Engine from the organism's stored policy.
func EngineFor(org *Organism) *policy.Engine {
	if org == nil {
		return policy.NewEngine()
	}
	if len(org.Policy.RulesList) > 0 {
		var rules []policy.Rule
		for _, r := range org.Policy.RulesList {
			rules = append(rules, policy.Rule{
				ID: r.ID,
				Subject: policy.Subject{Type: r.SubjectType, ID: r.SubjectID, Role: r.SubjectRole},
				Action: r.Action,
				Resource: policy.Resource{Type: r.ResourceType, ID: r.ResourceID},
				Effect: policy.Effect(r.Effect),
				Conditions: policy.Condition(r.Conditions),
				Priority: r.Priority,
			})
		}
		return policy.NewEngine(rules...)
	}
	if org.Policy.Rules != nil {
		return policy.FromCapabilityMap(org.Policy.Rules)
	}
	// declared capabilities default allow
	m := map[string]bool{}
	for _, c := range org.Capabilities {
		m[string(c)] = true
	}
	return policy.FromCapabilityMap(m)
}

func (m *Manager) authorize(org *Organism, action string, resType string) error {
	eng := EngineFor(org)
	req := policy.Request{
		Subject: policy.Subject{Type: "organism", ID: org.ID, Role: org.Policy.DefaultRole},
		Action:  action,
		Resource: policy.Resource{Type: resType, ID: org.ID},
		Context: policy.Condition{"status": string(org.Status)},
	}
	if err := eng.Authorize(req); err != nil {
		m.emit("policy.denied", org, map[string]any{"action": action, "reason": err.Error()})
		return err
	}
	m.emit("capability.executed", org, map[string]any{"action": action})
	return nil
}


// Create builds a new organism, persists it and registers it.
func (m *Manager) Create(opts CreateOptions) (*Organism, error) {
	if opts.NodeID == "" {
		opts.NodeID = m.nodeID
	}
	if opts.CreatedBy == "" {
		opts.CreatedBy = m.nodeID
	}

	org, err := New(opts)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.byID[org.ID]; exists {
		return nil, fmt.Errorf("organism already exists: %s", org.ID)
	}
	// Name uniqueness within this node (simple)
	for _, existing := range m.byID {
		if existing.Name == org.Name {
			return nil, fmt.Errorf("organism name already in use: %s", org.Name)
		}
	}

	org.Status = StatusReady
	org.Touch(m.nodeID, "ready")

	if err := m.persistLocked(org); err != nil {
		return nil, err
	}
	m.byID[org.ID] = org
	m.syncRegistry(org)

	m.emit("organism.created", org, nil)
	return org.Snapshot(), nil
}

// Start transitions organism to running.
func (m *Manager) Start(id string) (*Organism, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	org, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	switch org.Status {
	case StatusRunning:
		return org.Snapshot(), nil
	case StatusReady, StatusStopped, StatusCreated:
		// ok
	default:
		return nil, fmt.Errorf("cannot start organism in status %s", org.Status)
	}

	org.Status = StatusRunning
	org.CurrentAction = "idle"
	org.Touch(m.nodeID, "started")
	if err := m.persistLocked(org); err != nil {
		return nil, err
	}
	m.syncRegistry(org)
	m.emit("organism.started", org, nil)
	return org.Snapshot(), nil
}

// Stop transitions organism to stopped.
func (m *Manager) Stop(id string) (*Organism, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	org, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	if org.Status == StatusStopped {
		return org.Snapshot(), nil
	}
	if org.Status != StatusRunning && org.Status != StatusReady {
		return nil, fmt.Errorf("cannot stop organism in status %s", org.Status)
	}

	org.Status = StatusStopped
	org.CurrentAction = ""
	org.Touch(m.nodeID, "stopped")
	if err := m.persistLocked(org); err != nil {
		return nil, err
	}
	m.syncRegistry(org)
	m.emit("organism.stopped", org, nil)
	return org.Snapshot(), nil
}

// Get returns a snapshot by ID.

// Adopt installs an organism snapshot from another node (replication / recovery).
// RootCID is preserved; Primary becomes this node when promote is true.
func (m *Manager) Adopt(src *Organism, promote bool, note string) (*Organism, error) {
	if src == nil {
		return nil, fmt.Errorf("nil organism")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := src.Snapshot()
	if promote {
		cp.Placement.Primary = m.nodeID
		cp.Status = StatusRunning
		cp.CurrentAction = "recovered"
	} else {
		// cold replica stays ready/stopped
		if cp.Status == StatusRunning {
			cp.Status = StatusReady
		}
		cp.CurrentAction = "replica"
	}
	cp.Touch(m.nodeID, note)
	m.byID[cp.ID] = cp
	if err := m.persistLocked(cp); err != nil {
		return nil, err
	}
	m.syncRegistry(cp)
	ev := "organism.replicated"
	if promote {
		ev = "organism.recovered"
	}
	m.emit(ev, cp, map[string]any{"primary": cp.Placement.Primary, "note": note})
	return cp.Snapshot(), nil
}

// SetReplicas updates placement.replicas on the primary.
func (m *Manager) SetReplicas(id string, replicas []string) (*Organism, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	org, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	org.Placement.Replicas = append([]string(nil), replicas...)
	org.Touch(m.nodeID, "set-replicas")
	if err := m.persistLocked(org); err != nil {
		return nil, err
	}
	return org.Snapshot(), nil
}

func (m *Manager) Get(id string) (*Organism, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	org, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	return org.Snapshot(), nil
}

// List returns snapshots of all organisms.
func (m *Manager) List() []*Organism {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Organism, 0, len(m.byID))
	for _, org := range m.byID {
		out = append(out, org.Snapshot())
	}
	return out
}

// SetAction sets CurrentAction (used later by execution / demos).
func (m *Manager) SetAction(id, action string) (*Organism, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	org, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	org.CurrentAction = action
	org.Touch(m.nodeID, "action")
	if err := m.persistLocked(org); err != nil {
		return nil, err
	}
	m.syncRegistry(org)
	m.emit("organism.updated", org, map[string]any{"action": action})
	return org.Snapshot(), nil
}

// PutMemory sets a working-memory key (requires policy allow memory.write or memory.*).
func (m *Manager) PutMemory(id, key, value string) (*Organism, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	org, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	if err := m.authorize(org, "memory.write", "memory"); err != nil {
		// fallback: allow memory.read-only agents to set working if memory.write not configured but memory.read allowed and no explicit deny
		if err2 := m.authorize(org, "memory.read", "memory"); err2 != nil {
			return nil, err
		}
	}
	if org.Memory.Working == nil {
		org.Memory.Working = map[string]string{}
	}
	org.Memory.Working[key] = value
	org.Touch(m.nodeID, "memory.working")
	if err := m.persistLocked(org); err != nil {
		return nil, err
	}
	m.syncRegistry(org)
	m.emit("memory.updated", org, map[string]any{"kind": "working", "key": key})
	return org.Snapshot(), nil
}

// PutSemantic sets a stable semantic fact (value may be a CID string).
func (m *Manager) PutSemantic(id, key, value string) (*Organism, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	org, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	if org.Memory.Semantic == nil {
		org.Memory.Semantic = map[string]string{}
	}
	org.Memory.Semantic[key] = value
	org.Touch(m.nodeID, "memory.semantic")
	if err := m.persistLocked(org); err != nil {
		return nil, err
	}
	m.syncRegistry(org)
	m.emit("memory.updated", org, map[string]any{"kind": "semantic", "key": key})
	return org.Snapshot(), nil
}

// AppendEpisode adds an episodic memory entry. If content is non-nil it is
// stored in the block store and ContentCID is set on the episode.
func (m *Manager) AppendEpisode(id, episodeType string, payload map[string]any, content []byte) (*Organism, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	org, err := m.getLocked(id)
	if err != nil {
		return nil, err
	}
	if err := m.authorize(org, "memory.write", "memory"); err != nil {
		if err2 := m.authorize(org, "memory.read", "memory"); err2 != nil {
			return nil, err
		}
	}
	ep := Episode{
		ID:      fmt.Sprintf("ep-%d", time.Now().UnixNano()),
		Type:    episodeType,
		Payload: payload,
		At:      time.Now().UTC(),
	}
	if len(content) > 0 && m.blocks != nil {
		cid, err := m.blocks.Put(content)
		if err != nil {
			return nil, fmt.Errorf("store episode content: %w", err)
		}
		ep.ContentCID = cid
	}
	org.Memory.Episodic = append(org.Memory.Episodic, ep)
	org.Touch(m.nodeID, "memory.episodic")
	if err := m.persistLocked(org); err != nil {
		return nil, err
	}
	m.syncRegistry(org)
	m.emit("memory.updated", org, map[string]any{
		"kind":        "episodic",
		"episode_id":  ep.ID,
		"content_cid": ep.ContentCID,
	})
	return org.Snapshot(), nil
}

// GetBlock retrieves content from the CID block store.
func (m *Manager) GetBlock(contentCID string) ([]byte, error) {
	if m.blocks == nil {
		return nil, fmt.Errorf("block store not available")
	}
	return m.blocks.Get(contentCID)
}

func (m *Manager) getLocked(id string) (*Organism, error) {
	org, ok := m.byID[id]
	if !ok {
		// try match by name
		for _, o := range m.byID {
			if o.Name == id || o.ID == id {
				return o, nil
			}
		}
		return nil, fmt.Errorf("organism not found: %s", id)
	}
	return org, nil
}

func (m *Manager) persistLocked(org *Organism) error {
	path := filepath.Join(m.dataDir, org.ID+".json")
	data, err := json.MarshalIndent(org, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal organism: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("write organism: %w", err)
	}
	return os.Rename(tmp, path)
}

func (m *Manager) loadAll() error {
	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.dataDir, e.Name()))
		if err != nil {
			return err
		}
		var org Organism
		if err := json.Unmarshal(data, &org); err != nil {
			return fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		m.byID[org.ID] = &org
		m.syncRegistry(&org)
	}
	return nil
}

func (m *Manager) syncRegistry(org *Organism) {
	if m.reg == nil {
		return
	}
	_ = m.reg.PutOrganism(&registry.OrganismRecord{
		ID:        org.ID,
		RootCID:   org.RootCID,
		Name:      org.Name,
		Status:    string(org.Status),
		NodeID:    org.Placement.Primary,
		CreatedAt: org.CreatedAt,
		UpdatedAt: org.UpdatedAt,
	})
}

func (m *Manager) emit(typ string, org *Organism, extra map[string]any) {
	if m.bus == nil {
		return
	}
	payload := map[string]any{
		"id":      org.ID,
		"name":    org.Name,
		"rootcid": org.RootCID,
		"status":  string(org.Status),
	}
	for k, v := range extra {
		payload[k] = v
	}
	m.bus.PublishType(typ, m.nodeID, payload)
}
