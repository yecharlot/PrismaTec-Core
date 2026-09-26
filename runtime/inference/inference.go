// Package inference defines AI/model providers without hardcoding a vendor (Phase 9).
package inference

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Request is a provider-agnostic inference call.
type Request struct {
	Model   string
	Prompt  string
	System  string
	MaxTokens int
}

// Response from a provider.
type Response struct {
	Text     string
	Provider string
	Model    string
}

// Provider is pluggable (OpenAI, Ollama, local, HTTP, …).
type Provider interface {
	Name() string
	Infer(ctx context.Context, req Request) (Response, error)
}

// Registry of providers.
type Registry struct {
	mu   sync.RWMutex
	by   map[string]Provider
	def  string
}

func NewRegistry() *Registry {
	return &Registry{by: map[string]Provider{}}
}

func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.by[p.Name()] = p
	if r.def == "" {
		r.def = p.Name()
	}
}

func (r *Registry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.by[name]; !ok {
		return fmt.Errorf("unknown provider: %s", name)
	}
	r.def = name
	return nil
}

func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if name == "" {
		name = r.def
	}
	p, ok := r.by[name]
	if !ok {
		return nil, fmt.Errorf("inference provider not found: %s", name)
	}
	return p, nil
}

// EchoProvider is a deterministic local stand-in (no external API).
type EchoProvider struct{}

func (EchoProvider) Name() string { return "echo" }

func (EchoProvider) Infer(ctx context.Context, req Request) (Response, error) {
	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	default:
	}
	text := strings.TrimSpace(req.Prompt)
	if text == "" {
		text = "(empty prompt)"
	}
	return Response{
		Text:     "[echo] " + text,
		Provider: "echo",
		Model:    req.Model,
	}, nil
}

// HTTPProvider is a stub shape for remote models (URL configured later).
type HTTPProvider struct {
	Endpoint string
	Label    string
}

func (h HTTPProvider) Name() string {
	if h.Label != "" {
		return h.Label
	}
	return "http"
}

func (h HTTPProvider) Infer(ctx context.Context, req Request) (Response, error) {
	if h.Endpoint == "" {
		return Response{}, fmt.Errorf("http provider: endpoint not configured")
	}
	// Phase 9: interface only — real HTTP client in a later hardening pass
	return Response{}, fmt.Errorf("http provider not fully wired (endpoint=%s)", h.Endpoint)
}
