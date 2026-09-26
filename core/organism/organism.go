package organism

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Capability is something an organism is declared able to do.
// Authorization is decided later by Policy (Phase 7).
type Capability string

// Policy is a minimal placeholder for Phase 7.
// Format evolves toward: WHO can DO WHAT to WHICH under CONDITIONS.
type Policy struct {
	Rules map[string]bool `json:"rules,omitempty"` // capability name → allowed
}

// Episode is a time-ordered memory event (episodic memory).
type Episode struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Payload    map[string]any `json:"payload,omitempty"`
	ContentCID string         `json:"content_cid,omitempty"`
	At         time.Time      `json:"at"`
}

// MemoryRef holds working, episodic and semantic memory.
// Large payloads should be stored via storage/cid and referenced by ContentCID.
type MemoryRef struct {
	Working  map[string]string `json:"working,omitempty"`
	Episodic []Episode         `json:"episodic,omitempty"`
	Semantic map[string]string `json:"semantic,omitempty"`
}

// Provenance tracks origin and last change (expanded later).
type Provenance struct {
	CreatedBy string `json:"created_by,omitempty"`
	UpdatedBy string `json:"updated_by,omitempty"`
	Note      string `json:"note,omitempty"`
}

// Placement describes where the organism runs (replication in Phase 11–12).
type Placement struct {
	Primary  string   `json:"primary,omitempty"`
	Replicas []string `json:"replicas,omitempty"`
}

// RuntimeSpec declares how the organism may execute (WASM/Lisp/Agent later).
type RuntimeSpec struct {
	Kind string `json:"kind,omitempty"` // "none" | "wasm" | "lisp" | "agent"
}

// Manifest is the canonical definition used to derive RootCID.
type Manifest struct {
	Version      string       `json:"version"`
	Name         string       `json:"name"`
	Capabilities []Capability `json:"capabilities,omitempty"`
	Runtime      RuntimeSpec  `json:"runtime,omitempty"`
}

// Organism is the fundamental unit of PrismaTec Core.
// RootCID identifies content; NodeID (on Placement.Primary) identifies the host.
type Organism struct {
	ID           string       `json:"id"`
	RootCID      string       `json:"root_cid"`
	Version      string       `json:"version"`
	Name         string       `json:"name"`
	Status       Status       `json:"status"`
	Capabilities []Capability `json:"capabilities,omitempty"`
	Policy       Policy       `json:"policy,omitempty"`
	Memory       MemoryRef    `json:"memory,omitempty"`
	Runtime      RuntimeSpec  `json:"runtime,omitempty"`
	Provenance   Provenance   `json:"provenance,omitempty"`
	Placement    Placement    `json:"placement,omitempty"`
	CurrentAction string      `json:"current_action,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// CreateOptions configures a new organism.
type CreateOptions struct {
	Name         string
	Capabilities []Capability
	Policy       Policy
	Runtime      RuntimeSpec
	CreatedBy    string // usually NodeID
	NodeID       string // placement primary
}

// New creates an organism in StatusCreated with a deterministic RootCID from its manifest.
func New(opts CreateOptions) (*Organism, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("organism name is required")
	}
	if opts.Runtime.Kind == "" {
		opts.Runtime.Kind = "none"
	}
	if opts.Policy.Rules == nil {
		opts.Policy.Rules = map[string]bool{}
	}
	// Default: every declared capability is allowed unless policy says otherwise.
	for _, c := range opts.Capabilities {
		if _, ok := opts.Policy.Rules[string(c)]; !ok {
			opts.Policy.Rules[string(c)] = true
		}
	}

	manifest := Manifest{
		Version:      "0.1.0",
		Name:         opts.Name,
		Capabilities: opts.Capabilities,
		Runtime:      opts.Runtime,
	}
	rootCID, err := RootCIDFromManifest(manifest)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	id := deriveID(opts.Name, rootCID)

	return &Organism{
		ID:           id,
		RootCID:      rootCID,
		Version:      manifest.Version,
		Name:         opts.Name,
		Status:       StatusCreated,
		Capabilities: opts.Capabilities,
		Policy:       opts.Policy,
		Memory: MemoryRef{
			Working:  map[string]string{},
			Semantic: map[string]string{},
		},
		Runtime:      opts.Runtime,
		Provenance: Provenance{
			CreatedBy: opts.CreatedBy,
			UpdatedBy: opts.CreatedBy,
			Note:      "created",
		},
		Placement: Placement{
			Primary: opts.NodeID,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// RootCIDFromManifest computes a content-addressed RootCID from the canonical manifest.
// Phase 3 may upgrade this to a full multihash CID; the contract (deterministic, content-based) stays.
func RootCIDFromManifest(m Manifest) (string, error) {
	if m.Name == "" {
		return "", fmt.Errorf("manifest requires name")
	}
	if m.Version == "" {
		m.Version = "0.1.0"
	}
	canonical, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshal manifest: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return "rootcid:" + hex.EncodeToString(sum[:]), nil
}

func deriveID(name, rootCID string) string {
	// Stable, readable id: name + short hash of rootcid
	h := sha256.Sum256([]byte(rootCID))
	short := hex.EncodeToString(h[:6])
	return name + "-" + short
}

// HasCapability reports whether the capability is declared.
func (o *Organism) HasCapability(c Capability) bool {
	for _, x := range o.Capabilities {
		if x == c {
			return true
		}
	}
	return false
}

// Allowed reports whether policy permits a capability (default deny if unknown).
func (o *Organism) Allowed(c Capability) bool {
	if o.Policy.Rules == nil {
		return false
	}
	return o.Policy.Rules[string(c)]
}

// Touch updates UpdatedAt and provenance note.
func (o *Organism) Touch(by, note string) {
	o.UpdatedAt = time.Now().UTC()
	if by != "" {
		o.Provenance.UpdatedBy = by
	}
	if note != "" {
		o.Provenance.Note = note
	}
}

// Snapshot returns a deep-ish copy safe for read-only use.
func (o *Organism) Snapshot() *Organism {
	if o == nil {
		return nil
	}
	cp := *o
	if o.Capabilities != nil {
		cp.Capabilities = append([]Capability(nil), o.Capabilities...)
	}
	if o.Policy.Rules != nil {
		cp.Policy.Rules = make(map[string]bool, len(o.Policy.Rules))
		for k, v := range o.Policy.Rules {
			cp.Policy.Rules[k] = v
		}
	}
	if o.Memory.Working != nil {
		cp.Memory.Working = make(map[string]string, len(o.Memory.Working))
		for k, v := range o.Memory.Working {
			cp.Memory.Working[k] = v
		}
	}
	if o.Memory.Episodic != nil {
		cp.Memory.Episodic = append([]Episode(nil), o.Memory.Episodic...)
	}
	if o.Memory.Semantic != nil {
		cp.Memory.Semantic = make(map[string]string, len(o.Memory.Semantic))
		for k, v := range o.Memory.Semantic {
			cp.Memory.Semantic[k] = v
		}
	}
	if o.Placement.Replicas != nil {
		cp.Placement.Replicas = append([]string(nil), o.Placement.Replicas...)
	}
	return &cp
}
