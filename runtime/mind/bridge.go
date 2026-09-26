package mind

import (
	"context"
	"fmt"
	"strings"

	"github.com/yecharlot/PrismaTec-Core/core/organism"
	"github.com/yecharlot/PrismaTec-Core/core/policy"
	"github.com/yecharlot/PrismaTec-Core/runtime/execution"
)

// Bridge connects Mind proposals to Core policy + execution without granting Mind authority.
type Bridge struct {
	Mind Mind
	Orgs *organism.Manager
	Exec execution.Engine
}

// Tick: evaluate → authorize via organism policy map → execute only if allowed → record.
func (b *Bridge) Tick(ctx context.Context) (EvaluationResult, ActionResult, error) {
	ev, err := b.Mind.Evaluate(ctx)
	if err != nil {
		return ev, ActionResult{}, err
	}
	if ev.Request == nil {
		res := ActionResult{DecisionID: ev.Decision.ID, Authorized: false, Executed: false, OK: true, Detail: "abstained"}
		_ = b.Mind.RecordResult(res)
		return ev, res, nil
	}
	req := ev.Request
	org, err := b.Orgs.Get(req.OrganismID)
	if err != nil {
		res := ActionResult{RequestID: req.ID, DecisionID: req.DecisionID, Authorized: false, Detail: err.Error()}
		_ = b.Mind.RecordResult(res)
		return ev, res, err
	}

	capAction := mapAction(req.Action)
	eng := policy.FromCapabilityMap(org.Policy.Rules)
	dec := eng.Evaluate(policy.Request{
		Subject: policy.Subject{Type: "*", ID: "mind"},
		Action:  capAction,
		Resource: policy.Resource{Type: "organism", ID: org.ID},
		DefaultDeny: true,
	})
	if !dec.Allowed {
		res := ActionResult{RequestID: req.ID, DecisionID: req.DecisionID, Authorized: false, Executed: false, OK: false, Detail: dec.Reason}
		_ = b.Mind.RecordResult(res)
		return ev, res, nil
	}

	var detail string
	var seq int64
	ok := true
	switch {
	case req.Action == "append_note" || req.Action == "order.note" || strings.HasPrefix(req.Action, "memory."):
		key := "note"
		val := "mind-note"
		if v, okp := req.Params["note"].(string); okp {
			val = v
		}
		o, err := b.Orgs.PutMemory(org.ID, key, val)
		if err != nil {
			ok = false
			detail = err.Error()
		} else {
			detail = "memory.set:" + key
			seq = o.Seq
		}
	case req.Action == "ping" || req.Action == "execute.ping":
		exec := b.Exec
		if exec == nil {
			exec = execution.BuiltinEngine{}
		}
		out, err := exec.Execute(ctx, execution.Request{OrganismID: org.ID, Entry: "ping"})
		if err != nil {
			ok = false
			detail = err.Error()
		} else {
			detail = string(out.Output)
		}
	default:
		ok = false
		detail = fmt.Sprintf("no_executor_for_action:%s", req.Action)
	}

	res := ActionResult{
		RequestID: req.ID, DecisionID: req.DecisionID,
		Authorized: true, Executed: ok, OK: ok, Detail: detail, Seq: seq,
	}
	_ = b.Mind.RecordResult(res)
	return ev, res, nil
}

func mapAction(a string) string {
	switch a {
	case "append_note", "order.note":
		return "memory.write"
	case "ping", "execute.ping":
		return "execute"
	default:
		if strings.HasPrefix(a, "memory.") {
			return "memory.write"
		}
		return a
	}
}
