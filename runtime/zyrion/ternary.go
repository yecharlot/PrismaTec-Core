// Package zyrion provides deterministic ternary evaluation (0/1/2).
// Epistemic values only — never authorization or safety-critical halt codes.
package zyrion

import "fmt"

// Value is ternary epistemic state.
//
//	0 Negated  — sufficient evidence against
//	1 Unknown  — insufficient or conflicting evidence
//	2 Affirmed — sufficient evidence for
//
// Do NOT overload 2 as a critical safety absorb state; use Safety separately.
type Value int

const (
	Negated  Value = 0
	Unknown  Value = 1
	Affirmed Value = 2
)

func (v Value) String() string {
	switch v {
	case Negated:
		return "negated"
	case Unknown:
		return "unknown"
	case Affirmed:
		return "affirmed"
	default:
		return fmt.Sprintf("invalid(%d)", int(v))
	}
}

// Safety is orthogonal to ternary belief (critical halt / escalate / ok).
type Safety int

const (
	SafetyOK       Safety = 0
	SafetyEscalate Safety = 1
	SafetyBlock    Safety = 2 // absorb critical — not the same as Affirmed
)

// Evidence supports or opposes a proposition.
type Evidence struct {
	ID     string `json:"id"`
	For    bool   `json:"for"` // true = supports affirmation
	Weight int    `json:"weight"`
	Source string `json:"source,omitempty"`
	Note   string `json:"note,omitempty"`
}

// Proposition evaluated under evidence.
type Proposition struct {
	ID       string     `json:"id"`
	Text     string     `json:"text"`
	Evidence []Evidence `json:"evidence,omitempty"`
}

// Evaluate returns ternary value from weighted evidence.
// No evidence → Unknown. Both sides non-zero → Unknown (conflict).
// Only for → Affirmed. Only against → Negated.
func Evaluate(p Proposition) Value {
	var forW, againstW int
	for _, e := range p.Evidence {
		w := e.Weight
		if w <= 0 {
			w = 1
		}
		if e.For {
			forW += w
		} else {
			againstW += w
		}
	}
	if forW == 0 && againstW == 0 {
		return Unknown
	}
	if forW > 0 && againstW > 0 {
		return Unknown // conflict
	}
	if forW > 0 {
		return Affirmed
	}
	return Negated
}

// And ternary: Affirmed only if both Affirmed; Negated if any Negated; else Unknown.
func And(a, b Value) Value {
	if a == Negated || b == Negated {
		return Negated
	}
	if a == Affirmed && b == Affirmed {
		return Affirmed
	}
	return Unknown
}

// Or ternary: Negated only if both Negated; Affirmed if any Affirmed; else Unknown.
func Or(a, b Value) Value {
	if a == Affirmed || b == Affirmed {
		return Affirmed
	}
	if a == Negated && b == Negated {
		return Negated
	}
	return Unknown
}

// Not: Affirmed↔Negated, Unknown stays Unknown.
func Not(a Value) Value {
	switch a {
	case Affirmed:
		return Negated
	case Negated:
		return Affirmed
	default:
		return Unknown
	}
}

// Rule is a versioned deterministic production.
type Rule struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	When    string `json:"when"` // proposition id that must be Affirmed
	Then    string `json:"then"` // action name proposed (not executed)
}

// RuleSet holds versioned rules.
type RuleSet struct {
	Version string `json:"version"`
	Rules   []Rule `json:"rules"`
}

// Match returns Then actions whose When proposition is Affirmed.
func (rs RuleSet) Match(values map[string]Value) []string {
	var out []string
	for _, r := range rs.Rules {
		if values[r.When] == Affirmed {
			out = append(out, r.Then)
		}
	}
	return out
}
