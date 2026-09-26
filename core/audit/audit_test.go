package audit

import (
	"path/filepath"
	"testing"
)

func TestAuditFile(t *testing.T) {
	dir := t.TempDir()
	l := New(dir)
	l.Record(Event{Type: "command", Action: "create", OK: true})
	if l.Count() != 1 {
		t.Fatal(l.Count())
	}
	if _, err := filepath.Glob(filepath.Join(dir, "audit.jsonl")); err != nil {
		t.Fatal(err)
	}
	if len(l.Recent(10)) != 1 {
		t.Fatal("recent")
	}
}
