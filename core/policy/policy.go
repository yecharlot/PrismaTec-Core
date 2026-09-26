// Package policy implements a minimal, evolvable authorization engine.
//
// Model (Phase 7):
//
//	WHO  can  DO WHAT  on  WHICH resource  under  CONDITIONS
//
// This is intentionally small for the Core. Enterprise IAM can plug in later
// via a PolicyProvider without changing callers of Evaluate.
package policy

import (
	"fmt"
	"strings"
	"time"
)

// Effect of a rule.
type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

// Subject is WHO (agent, organism, node, role, or "*").
type Subject struct {
	ID   string `json:"id,omitempty"`   // e.g. organism id or "agent:A"
	Role string `json:"role,omitempty"` // e.g. "researcher"
	Type string `json:"type,omitempty"` // organism|agent|node|user|"*"
}

// Resource is WHICH (optional scope).
type Resource struct {
	Type string `json:"type,omitempty"` // e.g. "memory", "database", "organism"
	ID   string `json:"id,omitempty"`   // specific id or "*"
}

// Condition is a simple key/value constraint (Phase 7 subset).
// Examples: status=running, env=local
type Condition map[string]string

// Rule is one authorization statement.
type Rule struct {
	ID         string    `json:"id,omitempty"`
	Subject    Subject   `json:"subject"`
	Action     string    `json:"action"` // capability / verb, e.g. memory.read
	Resource   Resource  `json:"resource,omitempty"`
	Effect     Effect    `json:"effect"`
	Conditions Condition `json:"conditions,omitempty"`
	Priority   int       `json:"priority,omitempty"` // higher wins
}

// Request is what Evaluate answers.
type Request struct {
	Subject    Subject
	Action     string
	Resource   Resource
	Context    Condition // runtime facts (status, env, ...)
	DefaultDeny bool     // if no rule matches → deny when true (default)
}

// Decision is the result of Evaluate.
type Decision struct {
	Allowed bool   `json:"allowed"`
	Effect  Effect `json:"effect"`
	RuleID  string `json:"rule_id,omitempty"`
	Reason  string `json:"reason"`
	At      time.Time `json:"at"`
}

// Engine holds ordered rules and evaluates requests.
type Engine struct {
	Rules []Rule
}

// NewEngine returns an empty engine (default deny).
func NewEngine(rules ...Rule) *Engine {
	return &Engine{Rules: append([]Rule(nil), rules...)}
}

// FromCapabilityMap builds legacy Phase 2 map[string]bool into allow/deny rules
// for subject type organism (or any subject matching "*").
func FromCapabilityMap(caps map[string]bool) *Engine {
	var rules []Rule
	i := 0
	for action, allowed := range caps {
		i++
		eff := EffectDeny
		if allowed {
			eff = EffectAllow
		}
		rules = append(rules, Rule{
			ID:      fmt.Sprintf("legacy-%d", i),
			Subject: Subject{Type: "*"},
			Action:  action,
			Effect:  eff,
			Priority: 0,
		})
	}
	return NewEngine(rules...)
}

// Add appends a rule.
func (e *Engine) Add(r Rule) {
	e.Rules = append(e.Rules, r)
}

// Evaluate returns allow/deny. Deny rules with equal or higher priority beat allow.
// Algorithm: collect matching rules; if any deny at max priority among matches → deny;
// else if any allow → allow; else default deny.
func (e *Engine) Evaluate(req Request) Decision {
	now := time.Now().UTC()
	if req.Action == "" {
		return Decision{Allowed: false, Effect: EffectDeny, Reason: "empty action", At: now}
	}
	var matched []Rule
	for _, r := range e.Rules {
		if ruleMatches(r, req) {
			matched = append(matched, r)
		}
	}
	if len(matched) == 0 {
		return Decision{
			Allowed: false,
			Effect:  EffectDeny,
			Reason:  "no matching rule (default deny)",
			At:      now,
		}
	}
	// highest priority first; deny wins ties
	bestPri := matched[0].Priority
	for _, r := range matched {
		if r.Priority > bestPri {
			bestPri = r.Priority
		}
	}
	var top []Rule
	for _, r := range matched {
		if r.Priority == bestPri {
			top = append(top, r)
		}
	}
	for _, r := range top {
		if r.Effect == EffectDeny {
			return Decision{
				Allowed: false,
				Effect:  EffectDeny,
				RuleID:  r.ID,
				Reason:  fmt.Sprintf("denied by rule %s", r.ID),
				At:      now,
			}
		}
	}
	for _, r := range top {
		if r.Effect == EffectAllow {
			return Decision{
				Allowed: true,
				Effect:  EffectAllow,
				RuleID:  r.ID,
				Reason:  fmt.Sprintf("allowed by rule %s", r.ID),
				At:      now,
			}
		}
	}
	return Decision{Allowed: false, Effect: EffectDeny, Reason: "no allow rule at top priority", At: now}
}

// Authorize is Evaluate + error if denied.
func (e *Engine) Authorize(req Request) error {
	d := e.Evaluate(req)
	if d.Allowed {
		return nil
	}
	return &Denied{Decision: d}
}

// Denied is returned when authorization fails.
type Denied struct {
	Decision Decision
}

func (d *Denied) Error() string {
	return fmt.Sprintf("policy denied: %s", d.Decision.Reason)
}

func ruleMatches(r Rule, req Request) bool {
	if !actionMatch(r.Action, req.Action) {
		return false
	}
	if !subjectMatch(r.Subject, req.Subject) {
		return false
	}
	if !resourceMatch(r.Resource, req.Resource) {
		return false
	}
	if !conditionsMatch(r.Conditions, req.Context) {
		return false
	}
	return true
}

func actionMatch(ruleAction, reqAction string) bool {
	if ruleAction == "*" || ruleAction == reqAction {
		return true
	}
	// prefix match: "memory.*" vs "memory.read"
	if strings.HasSuffix(ruleAction, ".*") {
		prefix := strings.TrimSuffix(ruleAction, ".*")
		return strings.HasPrefix(reqAction, prefix+".") || reqAction == prefix
	}
	return false
}

func subjectMatch(rule, req Subject) bool {
	if rule.Type != "" && rule.Type != "*" && req.Type != "" && rule.Type != req.Type {
		return false
	}
	if rule.ID != "" && rule.ID != "*" && req.ID != "" && rule.ID != req.ID {
		return false
	}
	if rule.Role != "" && rule.Role != "*" && req.Role != "" && rule.Role != req.Role {
		return false
	}
	// empty rule fields = match any
	return true
}

func resourceMatch(rule, req Resource) bool {
	if rule.Type != "" && rule.Type != "*" && req.Type != "" && rule.Type != req.Type {
		return false
	}
	if rule.ID != "" && rule.ID != "*" && req.ID != "" && rule.ID != req.ID {
		return false
	}
	return true
}

func conditionsMatch(rule Condition, ctx Condition) bool {
	if len(rule) == 0 {
		return true
	}
	if ctx == nil {
		ctx = Condition{}
	}
	for k, v := range rule {
		if ctx[k] != v {
			return false
		}
	}
	return true
}
