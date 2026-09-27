package mind

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/runtime/zyrion"
)

// Phase 8 (partial): cognitive decision in semantic memory survives replicate snapshot path.
func TestCognitiveDecisionSurvivesReplicate(t *testing.T) {
	dir := t.TempDir()
	n, err := core.NewNode(core.Config{DataDir: filepath.Join(dir, "n"), Name: "cog"})
	if err != nil {
		t.Fatal(err)
	}
	org, err := n.Organisms().Create(organism.CreateOptions{
		Name:         "cog-org",
		Capabilities: []organism.Capability{"memory.read", "memory.write"},
	})
	if err != nil {
		t.Fatal(err)
	}
	m := New(Config{
		OrganismID: org.ID,
		Rules: zyrion.RuleSet{Version: "1", Rules: []zyrion.Rule{
			{ID: "r1", Version: "1", When: "order.open", Then: "append_note"},
		}},
	})
	_ = m.Observe(Observation{Type: "order.created", Payload: map[string]any{"note": "x"}})
	br := &Bridge{Mind: m, Orgs: n.Organisms()}
	ev, res, err := br.Tick(context.Background())
	if err != nil || !res.OK {
		t.Fatalf("%+v %v", res, err)
	}
	if err := PersistDecision(n.Organisms(), org.ID, ev.Decision); err != nil {
		t.Fatal(err)
	}
	got, err := n.Organisms().Get(org.ID)
	if err != nil {
		t.Fatal(err)
	}
	snap := got.Snapshot()
	if snap.Memory.Semantic[decisionMemoryKey] == "" {
		t.Fatal("semantic mind.last_decision missing on snapshot")
	}
	// Adopt onto a second manager simulates replica receiving organism blob
	n2, err := core.NewNode(core.Config{DataDir: filepath.Join(dir, "n2"), Name: "cog2"})
	if err != nil {
		t.Fatal(err)
	}
	adopted, err := n2.Organisms().Adopt(snap, false, "replica")
	if err != nil {
		t.Fatal(err)
	}
	if adopted.Memory.Semantic[decisionMemoryKey] == "" {
		t.Fatal("decision not present after adopt/replicate path")
	}
	loaded, err := LoadLastDecision(n2.Organisms(), org.ID)
	if err != nil || loaded.Selected != "append_note" {
		t.Fatalf("%+v %v", loaded, err)
	}
}
