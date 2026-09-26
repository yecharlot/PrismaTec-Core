// Package registry keeps track of organisms and nodes known to a Core instance.
package registry

import (
	"fmt"
	"sync"
	"time"
)

// OrganismRecord is a minimal registry entry (full organism model comes in Phase 2).
type OrganismRecord struct {
	ID        string
	RootCID   string
	Name      string
	Status    string
	NodeID    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Registry is an in-memory registry of organisms (and later nodes).
type Registry struct {
	mu         sync.RWMutex
	organisms  map[string]*OrganismRecord
}

// New creates an empty registry.
func New() *Registry {
	return &Registry{
		organisms: make(map[string]*OrganismRecord),
	}
}

// PutOrganism inserts or updates an organism record.
func (r *Registry) PutOrganism(rec *OrganismRecord) error {
	if rec == nil || rec.ID == "" {
		return fmt.Errorf("organism record requires ID")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = now
	}
	rec.UpdatedAt = now
	r.organisms[rec.ID] = rec
	return nil
}

// GetOrganism returns a copy of the organism record or nil.
func (r *Registry) GetOrganism(id string) *OrganismRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.organisms[id]
	if !ok {
		return nil
	}
	cp := *rec
	return &cp
}

// ListOrganisms returns all organism records.
func (r *Registry) ListOrganisms() []*OrganismRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*OrganismRecord, 0, len(r.organisms))
	for _, rec := range r.organisms {
		cp := *rec
		out = append(out, &cp)
	}
	return out
}

// DeleteOrganism removes an organism by ID.
func (r *Registry) DeleteOrganism(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.organisms, id)
}

// Count returns the number of registered organisms.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.organisms)
}
