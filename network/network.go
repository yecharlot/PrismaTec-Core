// Package network defines transport interfaces (Phase 10).
// libp2p (or others) plug in as adapters; organisms never import them directly.
package network

import (
	"context"
	"sync"
	"time"
)

// NodeAddr is a logical network address (not necessarily multiaddr yet).
type NodeAddr struct {
	NodeID string
	Addrs  []string // e.g. "local://node-a", later /ip4/.../tcp/...
}

// Message is a signed-ready payload between nodes.
type Message struct {
	From    string
	To      string // empty = broadcast
	Type    string
	Payload []byte
	SentAt  time.Time
}

// Handler receives messages.
type Handler func(Message)

// Transport is the core messaging interface.
type Transport interface {
	ID() string
	Listen(ctx context.Context) error
	Close() error
	Send(ctx context.Context, to string, msg Message) error
	Broadcast(ctx context.Context, msg Message) error
	OnMessage(h Handler)
	Peers() []NodeAddr
}

// Discovery finds peers.
type Discovery interface {
	Advertise(ctx context.Context, id string) error
	FindPeers(ctx context.Context) ([]NodeAddr, error)
}

// Presence tracks online nodes.
type Presence interface {
	Heartbeat(ctx context.Context) error
	Online() []string
	IsOnline(nodeID string) bool
}

// --- Local in-process fabric (multi-node demos without libp2p) ---

var defaultFabric = NewFabric()

// Fabric is a process-local switchboard connecting LocalTransport instances.
type Fabric struct {
	mu    sync.RWMutex
	nodes map[string]*LocalTransport
}

func NewFabric() *Fabric {
	return &Fabric{nodes: map[string]*LocalTransport{}}
}

func DefaultFabric() *Fabric { return defaultFabric }

// LocalTransport implements Transport over an in-memory fabric.
type LocalTransport struct {
	id      string
	fabric  *Fabric
	mu      sync.RWMutex
	handlers []Handler
	online  bool
}

func NewLocalTransport(nodeID string, fabric *Fabric) *LocalTransport {
	if fabric == nil {
		fabric = DefaultFabric()
	}
	t := &LocalTransport{id: nodeID, fabric: fabric}
	fabric.mu.Lock()
	fabric.nodes[nodeID] = t
	fabric.mu.Unlock()
	return t
}

func (t *LocalTransport) ID() string { return t.id }

func (t *LocalTransport) Listen(ctx context.Context) error {
	t.mu.Lock()
	t.online = true
	t.mu.Unlock()
	return nil
}

func (t *LocalTransport) Close() error {
	t.mu.Lock()
	t.online = false
	t.mu.Unlock()
	t.fabric.mu.Lock()
	delete(t.fabric.nodes, t.id)
	t.fabric.mu.Unlock()
	return nil
}

func (t *LocalTransport) OnMessage(h Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handlers = append(t.handlers, h)
}

func (t *LocalTransport) deliver(msg Message) {
	t.mu.RLock()
	hs := append([]Handler(nil), t.handlers...)
	online := t.online
	t.mu.RUnlock()
	if !online {
		return
	}
	for _, h := range hs {
		h(msg)
	}
}

func (t *LocalTransport) Send(ctx context.Context, to string, msg Message) error {
	msg.From = t.id
	msg.To = to
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	t.fabric.mu.RLock()
	peer, ok := t.fabric.nodes[to]
	t.fabric.mu.RUnlock()
	if !ok {
		return ErrPeerOffline
	}
	peer.deliver(msg)
	return nil
}

func (t *LocalTransport) Broadcast(ctx context.Context, msg Message) error {
	msg.From = t.id
	if msg.SentAt.IsZero() {
		msg.SentAt = time.Now().UTC()
	}
	t.fabric.mu.RLock()
	defer t.fabric.mu.RUnlock()
	for id, peer := range t.fabric.nodes {
		if id == t.id {
			continue
		}
		peer.deliver(msg)
	}
	return nil
}

func (t *LocalTransport) Peers() []NodeAddr {
	t.fabric.mu.RLock()
	defer t.fabric.mu.RUnlock()
	var out []NodeAddr
	for id, peer := range t.fabric.nodes {
		if id == t.id {
			continue
		}
		peer.mu.RLock()
		on := peer.online
		peer.mu.RUnlock()
		if on {
			out = append(out, NodeAddr{NodeID: id, Addrs: []string{"local://" + id}})
		}
	}
	return out
}

// IsOnline reports whether this transport is listening.
func (t *LocalTransport) IsOnline() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.online
}

// ErrPeerOffline is returned when the destination is not on the fabric / offline.
var ErrPeerOffline = errPeerOffline{}

type errPeerOffline struct{}

func (errPeerOffline) Error() string { return "peer offline or unknown" }
