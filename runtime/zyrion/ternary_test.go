package zyrion

import "testing"

func TestEvaluateEmptyUnknown(t *testing.T) {
	if Evaluate(Proposition{ID: "p"}) != Unknown {
		t.Fatal("empty → unknown")
	}
}

func TestEvaluateAffirmedNegatedConflict(t *testing.T) {
	if Evaluate(Proposition{Evidence: []Evidence{{For: true}}}) != Affirmed {
		t.Fatal("for")
	}
	if Evaluate(Proposition{Evidence: []Evidence{{For: false}}}) != Negated {
		t.Fatal("against")
	}
	if Evaluate(Proposition{Evidence: []Evidence{{For: true}, {For: false}}}) != Unknown {
		t.Fatal("conflict")
	}
}

func TestAndOrNot(t *testing.T) {
	if And(Affirmed, Affirmed) != Affirmed || And(Affirmed, Unknown) != Unknown || And(Negated, Affirmed) != Negated {
		t.Fatal("and")
	}
	if Or(Negated, Negated) != Negated || Or(Unknown, Negated) != Unknown || Or(Affirmed, Negated) != Affirmed {
		t.Fatal("or")
	}
	if Not(Affirmed) != Negated || Not(Unknown) != Unknown {
		t.Fatal("not")
	}
}

func TestRuleMatch(t *testing.T) {
	rs := RuleSet{Version: "1", Rules: []Rule{{ID: "r1", Version: "1", When: "order.open", Then: "append_note"}}}
	got := rs.Match(map[string]Value{"order.open": Affirmed})
	if len(got) != 1 || got[0] != "append_note" {
		t.Fatalf("%v", got)
	}
	if len(rs.Match(map[string]Value{"order.open": Unknown})) != 0 {
		t.Fatal("unknown must not match")
	}
}
