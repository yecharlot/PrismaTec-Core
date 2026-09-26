package mind

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/runtime/zyrion"
)

// TestWorkOrderReferenceApp is Phase 7: observation → decision → authorized note → deny illegal action → persist decision.
func TestWorkOrderReferenceApp(t *testing.T) {
	dir := t.TempDir()
	n, err := core.NewNode(core.Config{DataDir: filepath.Join(dir, "wo"), Name: "workorder-node"})
	if err != nil {
		t.Fatal(err)
	}
	org, err := n.Organisms().Create(organism.CreateOptions{
		Name:         "workorder-42",
		Capabilities: []organism.Capability{"memory.read", "memory.write"},
	})
	if err != nil {
		t.Fatal(err)
	}
	m := New(Config{
		OrganismID: org.ID,
		Rules: zyrion.RuleSet{Version: "wo-1", Rules: []zyrion.Rule{
			{ID: "open-note", Version: "wo-1", When: "order.open", Then: "append_note"},
		}},
	})
	_ = m.Observe(Observation{ID: "obs1", Type: "order.created", Payload: map[string]any{"note": "pump-check"}})
	br := &Bridge{Mind: m, Orgs: n.Organisms()}
	ev, res, err := br.Tick(context.Background())
	if err != nil || !res.Authorized || !res.OK {
		t.Fatalf("note: %+v %v", res, err)
	}
	if err := PersistDecision(n.Organisms(), org.ID, ev.Decision); err != nil {
		t.Fatal(err)
	}
	// reload node data path
	loaded, err := LoadLastDecision(n.Organisms(), org.ID)
	if err != nil || loaded.Selected != "append_note" {
		t.Fatalf("persist: %+v %v", loaded, err)
	}
	// unauthorized close
	m2 := New(Config{OrganismID: org.ID, Rules: zyrion.RuleSet{Version: "wo-1", Rules: []zyrion.Rule{
		{ID: "close", Version: "wo-1", When: "order.open", Then: "order.close"},
	}}})
	_ = m2.Observe(Observation{Type: "order.created"})
	_, res2, _ := (&Bridge{Mind: m2, Orgs: n.Organisms()}).Tick(context.Background())
	if res2.Authorized {
		t.Fatal("order.close must not authorize without capability")
	}
}
