package events

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishSubscribe(t *testing.T) {
	bus := NewBus()
	var count atomic.Int32

	bus.Subscribe("organism.created", func(e Event) {
		count.Add(1)
		if e.Type != "organism.created" {
			t.Errorf("unexpected type %s", e.Type)
		}
	})

	bus.PublishType("organism.created", "test", map[string]any{"id": "org-1"})
	bus.PublishType("other", "test", nil)

	if count.Load() != 1 {
		t.Fatalf("expected 1 handler call, got %d", count.Load())
	}
}

func TestSubscribeAll(t *testing.T) {
	bus := NewBus()
	var count atomic.Int32
	bus.SubscribeAll(func(e Event) { count.Add(1) })
	bus.PublishType("a", "s", nil)
	bus.PublishType("b", "s", nil)
	if count.Load() != 2 {
		t.Fatalf("expected 2, got %d", count.Load())
	}
}

func TestTimestamp(t *testing.T) {
	bus := NewBus()
	var got time.Time
	bus.Subscribe("t", func(e Event) { got = e.Timestamp })
	bus.Publish(Event{Type: "t", Source: "s"})
	if got.IsZero() {
		t.Fatal("timestamp should be set")
	}
}
