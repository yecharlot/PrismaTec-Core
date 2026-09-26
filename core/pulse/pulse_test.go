package pulse

import (
	"sync/atomic"
	"testing"
)

func TestEmitSubscribe(t *testing.T) {
	h := NewHub(10)
	var n atomic.Int32
	h.Subscribe(func(p Pulse) { n.Add(1) })
	h.EmitType("org-1", "organism.started", "node:x", map[string]any{"id": "org-1"})
	if n.Load() != 1 {
		t.Fatalf("got %d", n.Load())
	}
	if h.Count() != 1 || h.Seq() != 1 {
		t.Fatalf("count=%d seq=%d", h.Count(), h.Seq())
	}
}

func TestSubscribeTarget(t *testing.T) {
	h := NewHub(10)
	var hit atomic.Int32
	h.SubscribeTarget("org-1", func(p Pulse) { hit.Add(1) })
	h.EmitType("org-1", "x", "s", nil)
	h.EmitType("org-2", "x", "s", nil)
	if hit.Load() != 1 {
		t.Fatalf("target filter failed: %d", hit.Load())
	}
}

func TestRecentRing(t *testing.T) {
	h := NewHub(3)
	for i := 0; i < 5; i++ {
		h.EmitType("t", "tick", "s", map[string]any{"i": i})
	}
	if h.Count() != 3 {
		t.Fatalf("count %d", h.Count())
	}
	r := h.Recent(2)
	if len(r) != 2 {
		t.Fatalf("recent len %d", len(r))
	}
}
