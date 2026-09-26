// Package aip defines AIP v1 — Alset Interaction Protocol contract for PrismaTec Core.
//
// AIP is the versioned boundary between Core and clients (Alset-JS, SDKs, tools).
// Keep this package free of heavy dependencies; transport lives in server.go.
package aip

import "time"

const Version = "v1"

// Envelope is the common wrapper for AIP messages.
type Envelope struct {
	AIP       string `json:"aip"` // "v1"
	Type      string `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// NodeInfo is returned by GET /aip/v1/info
type NodeInfo struct {
	AIP        string `json:"aip"`
	Name       string `json:"name"`
	NodeID     string `json:"node_id"`
	Status     string `json:"status"`
	Organisms  int    `json:"organisms"`
	Pulses     int    `json:"pulses"`
	PulseSeq   uint64 `json:"pulse_seq"`
	DataDir    string `json:"data_dir,omitempty"`
}

// OrganismView is a client-safe snapshot of an organism.
type OrganismView struct {
	ID            string            `json:"id"`
	RootCID       string            `json:"root_cid"`
	Name          string            `json:"name"`
	Status        string            `json:"status"`
	NodeID        string            `json:"node_id"`
	Capabilities  []string          `json:"capabilities,omitempty"`
	Policy        map[string]bool   `json:"policy,omitempty"`
	WorkingKeys   int               `json:"working_keys"`
	Episodes      int               `json:"episodes"`
	SemanticKeys  int               `json:"semantic_keys"`
	CurrentAction string            `json:"current_action,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// PulseMessage is streamed over SSE / returned in pulse lists.
type PulseMessage struct {
	AIP       string         `json:"aip"`
	Target    string         `json:"target"`
	Type      string         `json:"type"`
	Data      map[string]any `json:"data,omitempty"`
	Meta      map[string]any `json:"meta,omitempty"`
	Source    string         `json:"source,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Command is a minimal client → core request (Phase 5 subset).
type Command struct {
	AIP    string         `json:"aip"`
	Action string         `json:"action"` // create|start|stop|memory.set
	Params map[string]any `json:"params,omitempty"`
}

// CommandResult is the response to a Command.
type CommandResult struct {
	AIP     string         `json:"aip"`
	OK      bool           `json:"ok"`
	Error   string         `json:"error,omitempty"`
	Organism *OrganismView `json:"organism,omitempty"`
}

// ErrorBody is a standard JSON error.
type ErrorBody struct {
	AIP   string `json:"aip"`
	Error string `json:"error"`
}
