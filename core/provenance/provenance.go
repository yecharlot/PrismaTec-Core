// Package provenance records origin of state changes (manifest REQUIREMENT).
package provenance

import (
	"sync"
	"time"
)

// Record is one auditable origin fact.
type Record struct {
	Time       time.Time `json:"time"`
	OrganismID string    `json:"organism_id"`
	Action     string    `json:"action"`
	Actor      string    `json:"actor"` // node or agent id
	Detail     string    `json:"detail,omitempty"`
	RootCID    string    `json:"root_cid,omitempty"`
}

// Log is process-local provenance trail.
type Log struct {
	mu  sync.Mutex
	buf []Record
}

func New() *Log { return &Log{} }

func (l *Log) Add(r Record) {
	if l == nil {
		return
	}
	if r.Time.IsZero() {
		r.Time = time.Now().UTC()
	}
	l.mu.Lock()
	l.buf = append(l.buf, r)
	if len(l.buf) > 5000 {
		l.buf = l.buf[len(l.buf)-5000:]
	}
	l.mu.Unlock()
}

func (l *Log) Recent(n int) []Record {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if n <= 0 || n > len(l.buf) {
		n = len(l.buf)
	}
	out := make([]Record, n)
	copy(out, l.buf[len(l.buf)-n:])
	return out
}
