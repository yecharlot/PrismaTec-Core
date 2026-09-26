// Package execution defines how organisms run modules (Phase 8).
//
//	Organism → Capability → Execution Engine → WASM | Builtin | Lisp | External
package execution

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Request to run a module.
type Request struct {
	OrganismID string
	Module     string // name or path key
	Entry      string // function / entrypoint
	Input      []byte
	Limits     Limits
}

// Limits bound untrusted execution.
type Limits struct {
	Timeout   time.Duration
	MaxMemory int // advisory
}

// Result of execution.
type Result struct {
	Output   []byte
	Logs     string
	Duration time.Duration
	Engine   string
}

// Engine runs modules in isolation as far as the implementation allows.
type Engine interface {
	Name() string
	Execute(ctx context.Context, req Request) (Result, error)
}

// Registry maps engine kind → Engine.
type Registry struct {
	mu      sync.RWMutex
	engines map[string]Engine
}

func NewRegistry() *Registry {
	return &Registry{engines: map[string]Engine{}}
}

func (r *Registry) Register(e Engine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.engines[e.Name()] = e
}

func (r *Registry) Get(name string) (Engine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.engines[name]
	if !ok {
		return nil, fmt.Errorf("execution engine not found: %s", name)
	}
	return e, nil
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.engines))
	for k := range r.engines {
		out = append(out, k)
	}
	return out
}

// BuiltinEngine is a safe no-sandbox demo engine (echo / set-action style).
type BuiltinEngine struct{}

func (BuiltinEngine) Name() string { return "builtin" }

func (BuiltinEngine) Execute(ctx context.Context, req Request) (Result, error) {
	start := time.Now()
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}
	switch req.Entry {
	case "", "echo":
		return Result{Output: req.Input, Engine: "builtin", Duration: time.Since(start)}, nil
	case "ping":
		return Result{Output: []byte("pong"), Engine: "builtin", Duration: time.Since(start)}, nil
	default:
		return Result{}, fmt.Errorf("builtin entry unknown: %s", req.Entry)
	}
}
