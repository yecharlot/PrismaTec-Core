// Package replication handles primary/replica placement and recovery (Phases 11–12).
package replication

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/network"
)

// Snapshot is a transferable organism state.
type Snapshot struct {
	Organism  organism.Organism `json:"organism"`
	Primary   string            `json:"primary"`
	Replicas  []string          `json:"replicas"`
	Version   int64             `json:"version"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Service coordinates replicate / failover over a Transport.
type Service struct {
	mu       sync.RWMutex
	nodeID   string
	transport network.Transport
	// local replicas stored by organism id
	replicas map[string]*Snapshot
	// online override for simulation (nodeID → online); nil = use transport only
	onlineFn func(nodeID string) bool
}

func NewService(nodeID string, tr network.Transport) *Service {
	return &Service{
		nodeID:    nodeID,
		transport: tr,
		replicas:  map[string]*Snapshot{},
	}
}

// SetOnlineFunc allows tests to simulate node failure independent of transport map.
func (s *Service) SetOnlineFunc(fn func(string) bool) {
	s.onlineFn = fn
}

func (s *Service) isOnline(id string) bool {
	if s.onlineFn != nil {
		return s.onlineFn(id)
	}
	if id == s.nodeID {
		if lt, ok := s.transport.(*network.LocalTransport); ok {
			return lt.IsOnline()
		}
		return true
	}
	for _, p := range s.transport.Peers() {
		if p.NodeID == id {
			return true
		}
	}
	return false
}

// Replicate sends a snapshot to replica node ids and records placement.
func (s *Service) Replicate(org *organism.Organism, replicaNodeIDs []string) (*Snapshot, error) {
	if org == nil {
		return nil, fmt.Errorf("nil organism")
	}
	snap := &Snapshot{
		Organism:  *org.Snapshot(),
		Primary:   org.Placement.Primary,
		Replicas:  append([]string(nil), replicaNodeIDs...),
		Version:   time.Now().UnixNano(),
		UpdatedAt: time.Now().UTC(),
	}
	if snap.Primary == "" {
		snap.Primary = s.nodeID
		snap.Organism.Placement.Primary = s.nodeID
	}
	snap.Organism.Placement.Replicas = append([]string(nil), replicaNodeIDs...)

	payload, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}
	for _, rid := range replicaNodeIDs {
		msg := network.Message{Type: "organism.replicate", Payload: payload}
		if err := s.transport.Send(nil, rid, msg); err != nil {
			return nil, fmt.Errorf("replicate to %s: %w", rid, err)
		}
	}
	return snap, nil
}

// HandleMessage stores incoming replica snapshots.
func (s *Service) HandleMessage(msg network.Message) {
	if msg.Type != "organism.replicate" && msg.Type != "organism.activate" {
		return
	}
	var snap Snapshot
	if err := json.Unmarshal(msg.Payload, &snap); err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg.Type == "organism.replicate" {
		// store as cold replica
		cp := snap
		s.replicas[snap.Organism.ID] = &cp
		return
	}
	if msg.Type == "organism.activate" {
		cp := snap
		cp.Primary = s.nodeID
		cp.Organism.Placement.Primary = s.nodeID
		cp.Organism.Status = organism.StatusRunning
		s.replicas[snap.Organism.ID] = &cp
	}
}

// GetReplica returns a stored replica snapshot if any.
func (s *Service) GetReplica(orgID string) *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if snap, ok := s.replicas[orgID]; ok {
		cp := *snap
		return &cp
	}
	return nil
}

// ListReplicas returns organism ids held as replicas on this node.
func (s *Service) ListReplicas() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.replicas))
	for id := range s.replicas {
		out = append(out, id)
	}
	return out
}

// RecoverResult describes a failover attempt.
type RecoverResult struct {
	OrganismID string
	OldPrimary string
	NewPrimary string
	Success    bool
	Reason     string
}

// RecoverIfPrimaryDown promotes a local replica when primary is offline.
func (s *Service) RecoverIfPrimaryDown(orgID string) (RecoverResult, error) {
	s.mu.Lock()
	snap, ok := s.replicas[orgID]
	if !ok || snap == nil {
		s.mu.Unlock()
		return RecoverResult{OrganismID: orgID, Success: false, Reason: "no local replica"}, fmt.Errorf("no local replica for %s", orgID)
	}
	cp := *snap
	s.mu.Unlock()

	old := cp.Primary
	if s.isOnline(old) {
		return RecoverResult{
			OrganismID: orgID,
			OldPrimary: old,
			Success:    false,
			Reason:     "primary still online",
		}, nil
	}

	// verify root cid present
	if cp.Organism.RootCID == "" {
		return RecoverResult{OrganismID: orgID, Success: false, Reason: "missing RootCID"}, fmt.Errorf("invalid snapshot")
	}

	cp.Primary = s.nodeID
	cp.Organism.Placement.Primary = s.nodeID
	cp.Organism.Status = organism.StatusRunning
	cp.Organism.CurrentAction = "recovered"
	cp.UpdatedAt = time.Now().UTC()
	cp.Version = time.Now().UnixNano()

	s.mu.Lock()
	s.replicas[orgID] = &cp
	s.mu.Unlock()

	// notify others
	payload, _ := json.Marshal(cp)
	_ = s.transport.Broadcast(nil, network.Message{Type: "organism.activate", Payload: payload})

	return RecoverResult{
		OrganismID: orgID,
		OldPrimary: old,
		NewPrimary: s.nodeID,
		Success:    true,
		Reason:     "promoted replica after primary offline",
	}, nil
}
