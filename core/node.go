// Package core provides the minimal PrismaTec Core node bootstrap.
package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core/audit"
	"github.com/yecharlot/PrismaTec-Core/core/events"
	"github.com/yecharlot/PrismaTec-Core/core/identity"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/core/pulse"
	"github.com/yecharlot/PrismaTec-Core/core/provenance"
	"github.com/yecharlot/PrismaTec-Core/core/registry"
	"github.com/yecharlot/PrismaTec-Core/core/replication"
	"github.com/yecharlot/PrismaTec-Core/network"
)

// Status of the node lifecycle.
type Status string

const (
	StatusStopped  Status = "stopped"
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusStopping Status = "stopping"
)

// Config for a Core node.
type Config struct {
	DataDir     string // directory for identity and local state
	Name        string // human-readable node name
	NetworkAddr string // TCP host:port OR libp2p multiaddr (/ip4/127.0.0.1/tcp/4001)
	// Peers maps remote NodeID → address (TCP host:port or peerID@multiaddr / full /p2p/ multiaddr)
	Peers map[string]string
	// TransportKind: "tcp" (default when NetworkAddr set without /ip4), "libp2p"
	TransportKind string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "."
	}
	return Config{
		DataDir: filepath.Join(home, ".prismatec", "node"),
		Name:    "node-local-01",
	}
}

// Node is the minimal PrismaTec Core runtime host.
type Node struct {
	cfg       Config
	identity  *identity.Identity
	registry  *registry.Registry
	bus       *events.Bus
	organisms *organism.Manager
	pulses    *pulse.Hub
	audit     *audit.Log
	prov      *provenance.Log
	transport network.Transport
	peerID    string
	repl      *replication.Service
	status    Status
	started   time.Time
	mu        sync.RWMutex
	cancel    context.CancelFunc
}

// NewNode creates a node (does not start it).
func NewNode(cfg Config) (*Node, error) {
	if cfg.DataDir == "" {
		cfg = DefaultConfig()
	}
	if cfg.Name == "" {
		cfg.Name = "node-local-01"
	}
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	idPath := filepath.Join(cfg.DataDir, "identity.json")
	id, err := identity.LoadOrCreate(idPath)
	if err != nil {
		return nil, fmt.Errorf("identity: %w", err)
	}

	reg := registry.New()
	bus := events.NewBus()
	hub := pulse.NewHub(256)
	orgMgr, err := organism.NewManager(cfg.DataDir, string(id.ID), cfg.Name, bus, reg)
	if err != nil {
		return nil, fmt.Errorf("organism manager: %w", err)
	}
	// Bridge internal events → localized pulses (Phase 4).
	bus.SubscribeAll(func(e events.Event) {
		target := string(id.ID)
		if e.Payload != nil {
			if oid, ok := e.Payload["id"].(string); ok && oid != "" {
				target = oid
			}
		}
		hub.Emit(pulse.Pulse{
			Target: target,
			Type:   e.Type,
			Data:   e.Payload,
			Source: e.Source,
			Meta:   map[string]any{"bridged_from": "events"},
		})
	})

	n := &Node{
		cfg:       cfg,
		identity:  id,
		registry:  reg,
		bus:       bus,
		organisms: orgMgr,
		pulses:    hub,
		audit:     audit.New(cfg.DataDir),
		prov:      provenance.New(),
		status:    StatusStopped,
	}
	n.peerID = network.DerivePeerID(string(id.ID), cfg.Name)
	return n, nil
}

// Audit returns the security audit log (Phase 15).
func (n *Node) Audit() *audit.Log {
	return n.audit
}

// PeerID returns the transport identity (distinct from NodeID).
func (n *Node) PeerID() string { return n.peerID }

// Transport returns the network transport (may be nil if offline-only).
func (n *Node) Transport() network.Transport { return n.transport }

// Replication returns the replication service (may be nil).
func (n *Node) Replication() *replication.Service { return n.repl }

// Provenance returns the origin log.
func (n *Node) Provenance() *provenance.Log { return n.prov }

// ID returns the persistent NodeID.
func (n *Node) ID() identity.NodeID {
	return n.identity.ID
}

// Name returns the configured node name.
func (n *Node) Name() string {
	return n.cfg.Name
}

// Registry returns the organism registry.
func (n *Node) Registry() *registry.Registry {
	return n.registry
}

// Bus returns the event bus.
func (n *Node) Bus() *events.Bus {
	return n.bus
}

// Identity returns the node identity.
func (n *Node) Identity() *identity.Identity {
	return n.identity
}

// Organisms returns the organism lifecycle manager.
func (n *Node) Organisms() *organism.Manager {
	return n.organisms
}

// Pulses returns the pulse hub.
func (n *Node) Pulses() *pulse.Hub {
	return n.pulses
}

// Status returns current lifecycle status.
func (n *Node) Status() Status {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status
}

// Start boots the node and optional TCP network (multi-node).
func (n *Node) Start(ctx context.Context) error {
	n.mu.Lock()
	if n.status == StatusRunning || n.status == StatusStarting {
		n.mu.Unlock()
		return fmt.Errorf("node already %s", n.status)
	}
	n.status = StatusStarting
	n.mu.Unlock()

	runCtx, cancel := context.WithCancel(ctx)
	n.cancel = cancel

	if n.cfg.NetworkAddr != "" {
		tr, err := n.buildTransport(runCtx)
		if err != nil {
			cancel()
			n.mu.Lock()
			n.status = StatusStopped
			n.mu.Unlock()
			return err
		}
		n.repl = replication.NewService(string(n.identity.ID), tr)
		tr.OnMessage(func(msg network.Message) {
			n.onNetworkMessage(msg)
		})
		if err := tr.Listen(runCtx); err != nil {
			cancel()
			_ = tr.Close()
			n.mu.Lock()
			n.status = StatusStopped
			n.mu.Unlock()
			return err
		}
		n.transport = tr
		if lp, ok := tr.(*network.LibP2PTransport); ok {
			n.peerID = lp.PeerID() // real libp2p peer id
		}
	}

	n.mu.Lock()
	n.status = StatusRunning
	n.started = time.Now().UTC()
	n.mu.Unlock()

	n.bus.PublishType("node.started", string(n.identity.ID), map[string]any{
		"name":    n.cfg.Name,
		"id":      string(n.identity.ID),
		"peer_id": n.peerID,
		"net":     n.cfg.NetworkAddr,
	})
	if n.pulses != nil {
		n.pulses.EmitType(string(n.identity.ID), "node.started", string(n.identity.ID), map[string]any{
			"name": n.cfg.Name, "peer_id": n.peerID,
		})
	}

	go func() {
		<-runCtx.Done()
		if n.transport != nil {
			_ = n.transport.Close()
		}
		n.mu.Lock()
		n.status = StatusStopped
		n.mu.Unlock()
		n.bus.PublishType("node.stopped", string(n.identity.ID), nil)
	}()

	return nil
}


func (n *Node) buildTransport(ctx context.Context) (network.Transport, error) {
	kind := n.cfg.TransportKind
	addr := n.cfg.NetworkAddr
	if kind == "" {
		if len(addr) > 0 && addr[0] == '/' {
			kind = "libp2p"
		} else {
			kind = "tcp"
		}
	}
	switch kind {
	case "libp2p":
		tr, err := network.NewLibP2PTransport(ctx, string(n.identity.ID), addr, nil)
		if err != nil {
			return nil, err
		}
		for id, a := range n.cfg.Peers {
			_ = tr.AddPeer(id, a) // best-effort; dial loop via Send
		}
		return tr, nil
	default:
		tr := network.NewTCPTransport(string(n.identity.ID), n.peerID, addr)
		for id, a := range n.cfg.Peers {
			tr.AddPeer(id, a)
		}
		return tr, nil
	}
}

func (n *Node) onNetworkMessage(msg network.Message) {
	if n.repl == nil {
		return
	}
	if msg.Type != "organism.replicate" && msg.Type != "organism.activate" {
		return
	}
	n.repl.HandleMessage(msg)
	var snap replication.Snapshot
	if err := json.Unmarshal(msg.Payload, &snap); err != nil {
		return
	}
	promote := msg.Type == "organism.activate"
	note := "replicated"
	if promote {
		note = "recovered-activate"
	}
	_, err := n.organisms.Adopt(&snap.Organism, promote, note)
	if err == nil && n.prov != nil {
		n.prov.Add(provenance.Record{
			OrganismID: snap.Organism.ID, Action: note, Actor: string(n.identity.ID),
			RootCID: snap.Organism.RootCID, Detail: msg.From,
		})
	}
}

// ReplicateOrganism copies organism to remote replica node IDs over the network.
func (n *Node) ReplicateOrganism(orgID string, replicaNodeIDs []string) error {
	if n.repl == nil || n.transport == nil {
		return fmt.Errorf("network not enabled (set PRISMATEC_NETWORK_ADDR)")
	}
	org, err := n.organisms.Get(orgID)
	if err != nil {
		return err
	}
	org.Placement.Primary = string(n.identity.ID)
	if _, err := n.organisms.SetReplicas(orgID, replicaNodeIDs); err != nil {
		return err
	}
	org, _ = n.organisms.Get(orgID)
	_, err = n.repl.Replicate(org, replicaNodeIDs)
	if err != nil {
		return err
	}
	if n.prov != nil {
		n.prov.Add(provenance.Record{
			OrganismID: org.ID, Action: "replicate", Actor: string(n.identity.ID),
			RootCID: org.RootCID, Detail: fmt.Sprintf("%v", replicaNodeIDs),
		})
	}
	return nil
}

// RecoverOrganism promotes a local replica if primary is unreachable.
func (n *Node) RecoverOrganism(orgID string) (*replication.RecoverResult, error) {
	if n.repl == nil {
		return nil, fmt.Errorf("network/replication not enabled")
	}
	// mark primary offline if not in peers
	n.repl.SetOnlineFunc(func(id string) bool {
		if id == string(n.identity.ID) {
			return true
		}
		if n.transport == nil {
			return false
		}
		for _, p := range n.transport.Peers() {
			if p.NodeID == id {
				return true
			}
		}
		return false
	})
	res, err := n.repl.RecoverIfPrimaryDown(orgID)
	if err != nil {
		return &res, err
	}
	if res.Success {
		snap := n.repl.GetReplica(orgID)
		if snap != nil {
			_, _ = n.organisms.Adopt(&snap.Organism, true, "recovered")
			if n.prov != nil {
				n.prov.Add(provenance.Record{
					OrganismID: orgID, Action: "recover", Actor: string(n.identity.ID),
					RootCID: snap.Organism.RootCID, Detail: res.Reason,
				})
			}
		}
	}
	return &res, nil
}

// Stop shuts the node down.
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.status != StatusRunning && n.status != StatusStarting {
		return nil
	}
	n.status = StatusStopping
	if n.cancel != nil {
		n.cancel()
		n.cancel = nil
	}
	if n.transport != nil {
		_ = n.transport.Close()
	}
	n.status = StatusStopped
	return nil
}

// Info returns a snapshot for CLI / observability.
func (n *Node) Info() map[string]any {
	n.mu.RLock()
	defer n.mu.RUnlock()
	count := 0
	if n.organisms != nil {
		count = len(n.organisms.List())
	}
	pulseCount := 0
	if n.pulses != nil {
		pulseCount = n.pulses.Count()
	}
	peers := 0
	netAddr := n.cfg.NetworkAddr
	if n.transport != nil {
		peers = len(n.transport.Peers())
	}
	return map[string]any{
		"name":         n.cfg.Name,
		"node_id":      string(n.identity.ID),
		"peer_id":      n.peerID,
		"status":       string(n.status),
		"started_at":   n.started,
		"organisms":    count,
		"pulses":       pulseCount,
		"data_dir":     n.cfg.DataDir,
		"network_addr": netAddr,
		"peers":        peers,
		"transport":    n.cfg.TransportKind,
	}
}
