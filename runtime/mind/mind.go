// Package mind is the cognitive layer: observe → evaluate → propose → record Core results.
// Mind never authorizes or executes; Core does.
package mind

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yecharlot/PrismaTec-Core/runtime/inference"
	"github.com/yecharlot/PrismaTec-Core/runtime/zyrion"
)

// Mind is the cognitive contract for an organism.
type Mind interface {
	OrganismID() string
	Observe(obs Observation) error
	State() CognitiveState
	Evaluate(ctx context.Context) (EvaluationResult, error)
	RecordResult(res ActionResult) error
}

// Config for a Mind instance.
type Config struct {
	OrganismID string
	Rules      zyrion.RuleSet
	Provider   inference.Provider // optional; nil = symbolic only
}

// Engine is the default Mind implementation (Director + Zyrion).
type Engine struct {
	mu       sync.Mutex
	cfg      Config
	beliefs  map[string]Belief
	decisions []Decision
	pending  []ActionRequest
	lastObs  *Observation
	seq      int
}

func New(cfg Config) *Engine {
	if cfg.Rules.Version == "" {
		cfg.Rules.Version = "1"
	}
	return &Engine{
		cfg:     cfg,
		beliefs: map[string]Belief{},
	}
}

func (e *Engine) OrganismID() string { return e.cfg.OrganismID }

func (e *Engine) Observe(obs Observation) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if obs.OrganismID != "" && obs.OrganismID != e.cfg.OrganismID {
		return fmt.Errorf("observation organism mismatch")
	}
	obs.OrganismID = e.cfg.OrganismID
	if obs.At.IsZero() {
		obs.At = time.Now().UTC()
	}
	e.lastObs = &obs
	// Derive simple beliefs from observation type/payload (deterministic).
	e.ingestObservation(obs)
	return nil
}

func (e *Engine) ingestObservation(obs Observation) {
	now := time.Now().UTC()
	switch obs.Type {
	case "order.created", "workorder.open":
		e.beliefs["order.open"] = Belief{Proposition: "order.open", Value: zyrion.Affirmed, Source: obs.Source, At: now, RuleVersion: e.cfg.Rules.Version}
	case "order.closed":
		e.beliefs["order.open"] = Belief{Proposition: "order.open", Value: zyrion.Negated, Source: obs.Source, At: now}
	case "evidence.for":
		prop, _ := obs.Payload["proposition"].(string)
		if prop != "" {
			e.beliefs[prop] = Belief{Proposition: prop, Value: zyrion.Affirmed, Source: obs.Source, At: now}
		}
	case "evidence.against":
		prop, _ := obs.Payload["proposition"].(string)
		if prop != "" {
			e.beliefs[prop] = Belief{Proposition: prop, Value: zyrion.Negated, Source: obs.Source, At: now}
		}
	case "evidence.conflict":
		prop, _ := obs.Payload["proposition"].(string)
		if prop != "" {
			e.beliefs[prop] = Belief{Proposition: prop, Value: zyrion.Unknown, Source: obs.Source, At: now}
		}
	default:
		// unknown observation → mark needs_info
		e.beliefs["observation.understood"] = Belief{Proposition: "observation.understood", Value: zyrion.Unknown, Source: obs.Source, At: now}
	}
}

func (e *Engine) State() CognitiveState {
	e.mu.Lock()
	defer e.mu.Unlock()
	b := make(map[string]Belief, len(e.beliefs))
	for k, v := range e.beliefs {
		b[k] = v
	}
	dec := append([]Decision(nil), e.decisions...)
	pen := append([]ActionRequest(nil), e.pending...)
	return CognitiveState{
		OrganismID: e.cfg.OrganismID,
		Beliefs:    b,
		Decisions:  dec,
		Pending:    pen,
		UpdatedAt:  time.Now().UTC(),
	}
}

// Evaluate runs Director: Zyrion rules → propose or abstain. Optional LLM only as proposal text.
func (e *Engine) Evaluate(ctx context.Context) (EvaluationResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	values := map[string]zyrion.Value{}
	var beliefs []Belief
	for k, b := range e.beliefs {
		values[k] = b.Value
		beliefs = append(beliefs, b)
	}

	// Conflict / unknown on required props → abstain
	candidates := e.cfg.Rules.Match(values)
	e.seq++
	id := fmt.Sprintf("dec-%s-%d", e.cfg.OrganismID, e.seq)
	obsID := ""
	if e.lastObs != nil {
		obsID = e.lastObs.ID
	}

	d := Decision{
		ID:          id,
		OrganismID:  e.cfg.OrganismID,
		ObservationID: obsID,
		Beliefs:     beliefs,
		RuleVersion: e.cfg.Rules.Version,
		Candidates:  candidates,
		At:          time.Now().UTC(),
	}

	// Optional LLM: annotation only, never grants permission
	if e.cfg.Provider != nil {
		resp, err := e.cfg.Provider.Infer(ctx, inference.Request{
			Prompt: "summarize decision candidates", Model: "n/a",
		})
		if err == nil {
			d.ModelProvider = resp.Provider
			d.ModelVersion = resp.Model
			d.DirectorNote = "llm_proposal:" + resp.Text
		} else {
			d.DirectorNote = "llm_unavailable:" + err.Error()
		}
	}

	// Abstain if no candidates or any critical unknown
	if len(candidates) == 0 {
		d.Selected = ""
		d.AbstainReason = "no_matching_rules_or_unknown"
		d.Uncertainty = zyrion.Unknown
		d.DirectorNote = joinNote(d.DirectorNote, "abstain")
		e.decisions = append(e.decisions, d)
		return EvaluationResult{Decision: d, Request: nil}, nil
	}
	// Prefer first matched rule action
	selected := candidates[0]
	// Safety: never treat LLM text as action
	d.Selected = selected
	d.Uncertainty = zyrion.Affirmed
	d.DirectorNote = joinNote(d.DirectorNote, "propose:"+selected)

	req := &ActionRequest{
		ID:         fmt.Sprintf("req-%s-%d", e.cfg.OrganismID, e.seq),
		OrganismID: e.cfg.OrganismID,
		Action:     selected,
		Params:     map[string]any{},
		DecisionID: id,
		Reason:     "rule_match",
	}
	if e.lastObs != nil && e.lastObs.Payload != nil {
		req.Params = e.lastObs.Payload
	}
	e.pending = append(e.pending, *req)
	e.decisions = append(e.decisions, d)
	return EvaluationResult{Decision: d, Request: req}, nil
}

func (e *Engine) RecordResult(res ActionResult) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Update last matching decision
	for i := len(e.decisions) - 1; i >= 0; i-- {
		if e.decisions[i].ID == res.DecisionID || res.DecisionID == "" {
			if res.Authorized {
				e.decisions[i].AuthResult = "authorized"
			} else {
				e.decisions[i].AuthResult = "denied"
			}
			if res.Executed && res.OK {
				e.decisions[i].ExecutionResult = "confirmed:" + res.Detail
			} else if !res.Authorized {
				e.decisions[i].ExecutionResult = "not_executed_denied"
			} else {
				e.decisions[i].ExecutionResult = "failed:" + res.Detail
			}
			e.decisions[i].PolicyResult = e.decisions[i].AuthResult
			break
		}
	}
	// drop pending request
	var left []ActionRequest
	for _, p := range e.pending {
		if p.ID != res.RequestID {
			left = append(left, p)
		}
	}
	e.pending = left
	return nil
}

func joinNote(a, b string) string {
	if a == "" {
		return b
	}
	return a + ";" + b
}
