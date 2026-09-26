// Package events provides a minimal in-process EventBus for PrismaTec Core.
package events

import (
	"sync"
	"time"
)

// Event is a typed message emitted by the core.
type Event struct {
	Type      string
	Source    string
	Payload   map[string]any
	Timestamp time.Time
}

// Handler receives events.
type Handler func(Event)

// Bus is a simple publish/subscribe event bus.
type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	all      []Handler
}

// NewBus creates an empty event bus.
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for a specific event type.
func (b *Bus) Subscribe(eventType string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], h)
}

// SubscribeAll registers a handler for every event.
func (b *Bus) SubscribeAll(h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.all = append(b.all, h)
}

// Publish emits an event to matching subscribers.
func (b *Bus) Publish(e Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, h := range b.handlers[e.Type] {
		h(e)
	}
	for _, h := range b.all {
		h(e)
	}
}

// PublishType is a convenience for simple events.
func (b *Bus) PublishType(eventType, source string, payload map[string]any) {
	b.Publish(Event{
		Type:    eventType,
		Source:  source,
		Payload: payload,
	})
}
