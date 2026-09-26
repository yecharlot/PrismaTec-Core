// Package pulse implements the outward communication unit of PrismaTec Core.
//
// A Pulse is an asynchronous, identity-directed stimulus:
//
//	{ target, type, data, meta, source, timestamp }
//
// Philosophy (from Alset-JS + AlsetOS): do not update the whole world when one
// element changes — address a specific organism, node, or (later) UI node.
//
// Phase 4: in-process hub + ring buffer log. Phase 5+ will transport pulses
// over AIP (SSE/WebSocket) to Alset-JS and peers.
package pulse

import (
	"sync"
	"time"
)

// Pulse is the fundamental communication unit toward observers and peers.
type Pulse struct {
	Target    string         `json:"target"` // organism id, node id, or UI key
	Type      string         `json:"type"`
	Data      map[string]any `json:"data,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	Source    string         `json:"source,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Handler receives pulses.
type Handler func(Pulse)

// Hub fans out pulses to subscribers and keeps a recent log for inspect/CLI.
type Hub struct {
	mu       sync.RWMutex
	handlers []Handler
	byTarget map[string][]Handler
	log      []Pulse
	logMax   int
	seq      uint64
}

// NewHub creates a pulse hub with a bounded recent log (default 256).
func NewHub(logMax int) *Hub {
	if logMax <= 0 {
		logMax = 256
	}
	return &Hub{
		byTarget: make(map[string][]Handler),
		logMax:   logMax,
	}
}

// Subscribe registers a handler for all pulses.
func (h *Hub) Subscribe(fn Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers = append(h.handlers, fn)
}

// SubscribeTarget registers a handler for a specific target key.
func (h *Hub) SubscribeTarget(target string, fn Handler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.byTarget[target] = append(h.byTarget[target], fn)
}

// Emit sends a pulse to matching handlers and records it in the log.
func (h *Hub) Emit(p Pulse) {
	if p.Timestamp.IsZero() {
		p.Timestamp = time.Now().UTC()
	}
	h.mu.Lock()
	h.seq++
	if p.Meta == nil {
		p.Meta = map[string]any{}
	}
	p.Meta["seq"] = h.seq
	h.log = append(h.log, p)
	if len(h.log) > h.logMax {
		h.log = h.log[len(h.log)-h.logMax:]
	}
	all := append([]Handler(nil), h.handlers...)
	targeted := append([]Handler(nil), h.byTarget[p.Target]...)
	h.mu.Unlock()

	for _, fn := range all {
		fn(p)
	}
	for _, fn := range targeted {
		fn(p)
	}
}

// EmitType is a convenience builder.
func (h *Hub) EmitType(target, typ, source string, data map[string]any) {
	h.Emit(Pulse{
		Target: target,
		Type:   typ,
		Data:   data,
		Source: source,
	})
}

// Recent returns a copy of the last n pulses (newest last). n<=0 → all in log.
func (h *Hub) Recent(n int) []Pulse {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if n <= 0 || n >= len(h.log) {
		out := make([]Pulse, len(h.log))
		copy(out, h.log)
		return out
	}
	out := make([]Pulse, n)
	copy(out, h.log[len(h.log)-n:])
	return out
}

// Count returns how many pulses are currently retained in the log.
func (h *Hub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.log)
}

// Seq returns the monotonic sequence number of the last emitted pulse.
func (h *Hub) Seq() uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.seq
}
