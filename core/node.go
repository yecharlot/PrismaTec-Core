// Package core provides the minimal PrismaTec Core node bootstrap.
package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core/events"
	"github.com/yecharlot/PrismaTec-Core/core/identity"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/core/pulse"
	"github.com/yecharlot/PrismaTec-Core/core/registry"
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
	DataDir string // directory for identity and local state
	Name    string // human-readable node name
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

	return &Node{
		cfg:       cfg,
		identity:  id,
		registry:  reg,
		bus:       bus,
		organisms: orgMgr,
		pulses:    hub,
		status:    StatusStopped,
	}, nil
}

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

// Start boots the node.
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

	n.mu.Lock()
	n.status = StatusRunning
	n.started = time.Now().UTC()
	n.mu.Unlock()

	n.bus.PublishType("node.started", string(n.identity.ID), map[string]any{
		"name": n.cfg.Name,
		"id":   string(n.identity.ID),
	})
	if n.pulses != nil {
		n.pulses.EmitType(string(n.identity.ID), "node.started", string(n.identity.ID), map[string]any{
			"name": n.cfg.Name,
		})
	}

	// Background: keep context alive until Stop or parent cancel.
	go func() {
		<-runCtx.Done()
		n.mu.Lock()
		n.status = StatusStopped
		n.mu.Unlock()
		n.bus.PublishType("node.stopped", string(n.identity.ID), nil)
	}()

	return nil
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
	return map[string]any{
		"name":       n.cfg.Name,
		"node_id":    string(n.identity.ID),
		"status":     string(n.status),
		"started_at": n.started,
		"organisms":  count,
		"pulses":     pulseCount,
		"data_dir":   n.cfg.DataDir,
	}
}
