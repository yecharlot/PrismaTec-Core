package mind

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/yecharlot/PrismaTec-Core/core"
	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/runtime/inference"
	"github.com/yecharlot/PrismaTec-Core/runtime/zyrion"
)

func TestObserveAndAbstainUnknown(t *testing.T) {
	m := New(Config{
		OrganismID: "org-1",
		Rules: zyrion.RuleSet{Version: "1", Rules: []zyrion.Rule{
			{ID: "r1", Version: "1", When: "order.open", Then: "append_note"},
		}},
	})
	_ = m.Observe(Observation{ID: "o1", Type: "noise.unknown", At: time.Now().UTC()})
	ev, err := m.Evaluate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ev.Request != nil {
		t.Fatal("should abstain on unknown")
	}
	if ev.Decision.AbstainReason == "" {
		t.Fatal("expected abstain reason")
	}
}

func TestProposeWhenAffirmed(t *testing.T) {
	m := New(Config{
		OrganismID: "org-1",
		Rules: zyrion.RuleSet{Version: "1", Rules: []zyrion.Rule{
			{ID: "r1", Version: "1", When: "order.open", Then: "append_note"},
		}},
	})
	_ = m.Observe(Observation{ID: "o1", Type: "order.created", At: time.Now().UTC()})
	ev, err := m.Evaluate(context.Background())
	if err != nil || ev.Request == nil || ev.Request.Action != "append_note" {
		t.Fatalf("%+v %v", ev, err)
	}
}

func TestRecordDeniedDoesNotLookLikeSuccess(t *testing.T) {
	m := New(Config{OrganismID: "org-1", Rules: zyrion.RuleSet{Version: "1"}})
	_ = m.Observe(Observation{Type: "order.created"})
	ev, _ := m.Evaluate(context.Background())
	_ = m.RecordResult(ActionResult{DecisionID: ev.Decision.ID, Authorized: false, Executed: false, Detail: "denied"})
	st := m.State()
	if len(st.Decisions) == 0 || st.Decisions[len(st.Decisions)-1].AuthResult != "denied" {
		t.Fatalf("%+v", st.Decisions)
	}
}

func TestBridgeAuthorizeAndDeny(t *testing.T) {
	dir := t.TempDir()
	n, err := core.NewNode(core.Config{DataDir: filepath.Join(dir, "n"), Name: "mind-node"})
	if err != nil {
		t.Fatal(err)
	}
	org, err := n.Organisms().Create(organism.CreateOptions{
		Name:         "wo-1",
		Capabilities: []organism.Capability{"memory.read", "memory.write"},
	})
	if err != nil {
		t.Fatal(err)
	}
	rules := zyrion.RuleSet{Version: "1", Rules: []zyrion.Rule{
		{ID: "r1", Version: "1", When: "order.open", Then: "append_note"},
	}}
	m := New(Config{OrganismID: org.ID, Rules: rules})
	_ = m.Observe(Observation{Type: "order.created", Payload: map[string]any{"note": "field-visit"}})
	br := &Bridge{Mind: m, Orgs: n.Organisms()}
	_, res, err := br.Tick(context.Background())
	if err != nil || !res.Authorized || !res.OK {
		t.Fatalf("authorized path: %+v %v", res, err)
	}
	got, _ := n.Organisms().Get(org.ID)
	if got.Memory.Working["note"] != "field-visit" {
		t.Fatalf("memory %v", got.Memory.Working)
	}

	// Deny: propose execute without capability
	m2 := New(Config{OrganismID: org.ID, Rules: zyrion.RuleSet{Version: "1", Rules: []zyrion.Rule{
		{ID: "r2", Version: "1", When: "order.open", Then: "ping"},
	}}})
	_ = m2.Observe(Observation{Type: "order.created"})
	br2 := &Bridge{Mind: m2, Orgs: n.Organisms()}
	_, res2, _ := br2.Tick(context.Background())
	if res2.Authorized {
		t.Fatal("ping must be denied without execute capability")
	}
}

func TestWorksWithoutLLM(t *testing.T) {
	m := New(Config{OrganismID: "x", Rules: zyrion.RuleSet{Version: "1"}})
	_ = m.Observe(Observation{Type: "order.created"})
	ev, err := m.Evaluate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_ = ev
}

func TestLLMOptionalDoesNotGrantPermission(t *testing.T) {
	m := New(Config{
		OrganismID: "x",
		Provider:   inference.EchoProvider{},
		Rules:      zyrion.RuleSet{Version: "1"}, // no rules → abstain even with LLM
	})
	_ = m.Observe(Observation{Type: "order.created"})
	ev, err := m.Evaluate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ev.Request != nil {
		t.Fatal("LLM must not create actions without rules")
	}
	if ev.Decision.ModelProvider != "echo" {
		t.Fatalf("expected echo annotation, got %q", ev.Decision.ModelProvider)
	}
}
