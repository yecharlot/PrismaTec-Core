package policy

import "testing"

func TestAllowDenyPriority(t *testing.T) {
	e := NewEngine(
		Rule{ID: "allow-read", Subject: Subject{Type: "*"}, Action: "memory.read", Effect: EffectAllow, Priority: 1},
		Rule{ID: "deny-write", Subject: Subject{Type: "*"}, Action: "memory.write", Effect: EffectDeny, Priority: 1},
	)
	if !e.Evaluate(Request{Action: "memory.read"}).Allowed {
		t.Fatal("read should allow")
	}
	if e.Evaluate(Request{Action: "memory.write"}).Allowed {
		t.Fatal("write should deny")
	}
	if e.Evaluate(Request{Action: "inference"}).Allowed {
		t.Fatal("default deny")
	}
}

func TestDenyWinsSamePriority(t *testing.T) {
	e := NewEngine(
		Rule{ID: "a", Action: "x", Effect: EffectAllow, Priority: 5, Subject: Subject{Type: "*"}},
		Rule{ID: "d", Action: "x", Effect: EffectDeny, Priority: 5, Subject: Subject{Type: "*"}},
	)
	if e.Evaluate(Request{Action: "x"}).Allowed {
		t.Fatal("deny should win")
	}
}

func TestWildcardAction(t *testing.T) {
	e := NewEngine(Rule{ID: "m", Action: "memory.*", Effect: EffectAllow, Subject: Subject{Type: "*"}})
	if !e.Evaluate(Request{Action: "memory.read"}).Allowed {
		t.Fatal("memory.* should match memory.read")
	}
}

func TestSubjectAndConditions(t *testing.T) {
	e := NewEngine(Rule{
		ID: "role", Subject: Subject{Role: "researcher"}, Action: "database.read",
		Effect: EffectAllow, Conditions: Condition{"status": "running"},
	})
	d := e.Evaluate(Request{
		Subject: Subject{Role: "researcher"},
		Action:  "database.read",
		Context: Condition{"status": "running"},
	})
	if !d.Allowed {
		t.Fatalf("%+v", d)
	}
	d = e.Evaluate(Request{
		Subject: Subject{Role: "researcher"},
		Action:  "database.read",
		Context: Condition{"status": "stopped"},
	})
	if d.Allowed {
		t.Fatal("condition should fail")
	}
}

func TestFromCapabilityMap(t *testing.T) {
	e := FromCapabilityMap(map[string]bool{"inference": true, "database.write": false})
	if !e.Evaluate(Request{Action: "inference"}).Allowed {
		t.Fatal("inference")
	}
	if e.Evaluate(Request{Action: "database.write"}).Allowed {
		t.Fatal("write denied")
	}
}


func TestAuthorizeError(t *testing.T) {
	e := NewEngine()
	err := e.Authorize(Request{Action: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}
