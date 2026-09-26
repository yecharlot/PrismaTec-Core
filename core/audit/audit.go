// Package audit provides an append-only security event log (Phase 15).
package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Event is one security-relevant fact.
type Event struct {
	Time       time.Time      `json:"time"`
	Type       string         `json:"type"` // policy.denied|command|auth|execution|...
	Actor      string         `json:"actor,omitempty"`
	OrganismID string         `json:"organism_id,omitempty"`
	Action     string         `json:"action,omitempty"`
	OK         bool           `json:"ok"`
	Detail     string         `json:"detail,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
}

// Log is a process-local, optionally file-backed audit trail.
type Log struct {
	mu   sync.Mutex
	path string
	buf  []Event
	max  int
}

// New creates an in-memory log; if dataDir != "", also appends JSONL to audit.jsonl.
func New(dataDir string) *Log {
	l := &Log{max: 2000}
	if dataDir != "" {
		l.path = filepath.Join(dataDir, "audit.jsonl")
	}
	return l
}

// Record appends an event.
func (l *Log) Record(e Event) {
	if l == nil {
		return
	}
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf = append(l.buf, e)
	if len(l.buf) > l.max {
		l.buf = l.buf[len(l.buf)-l.max:]
	}
	if l.path != "" {
		f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return
		}
		_ = json.NewEncoder(f).Encode(e)
		_ = f.Close()
	}
}

// Recent returns the last n events (newest last).
func (l *Log) Recent(n int) []Event {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if n <= 0 || n > len(l.buf) {
		n = len(l.buf)
	}
	out := make([]Event, n)
	copy(out, l.buf[len(l.buf)-n:])
	return out
}

// Count returns buffered events.
func (l *Log) Count() int {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buf)
}
