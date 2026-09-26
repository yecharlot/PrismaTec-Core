package mind

import (
	"time"

	"github.com/yecharlot/PrismaTec-Core/runtime/zyrion"
)

// Observation from Core (event/pulse/external).
type Observation struct {
	ID         string         `json:"id"`
	OrganismID string         `json:"organism_id"`
	Type       string         `json:"type"`
	Payload    map[string]any `json:"payload,omitempty"`
	Source     string         `json:"source,omitempty"`
	At         time.Time      `json:"at"`
}

// Belief is an epistemic claim with ternary value + evidence refs.
type Belief struct {
	Proposition string        `json:"proposition"`
	Value       zyrion.Value  `json:"value"`
	EvidenceIDs []string      `json:"evidence_ids,omitempty"`
	Source      string        `json:"source,omitempty"`
	At          time.Time     `json:"at"`
	ExpiresAt   time.Time     `json:"expires_at,omitempty"`
	Confidence  float64       `json:"confidence,omitempty"` // auxiliary; does not replace ternary
	RuleVersion string        `json:"rule_version,omitempty"`
}

// ActionRequest is a proposal — never a privileged execution.
type ActionRequest struct {
	ID         string         `json:"id"`
	OrganismID string         `json:"organism_id"`
	Action     string         `json:"action"` // e.g. memory.set, execute, order.note
	Params     map[string]any `json:"params,omitempty"`
	DecisionID string         `json:"decision_id,omitempty"`
	Reason     string         `json:"reason,omitempty"`
}

// ActionResult is confirmed only by Core.
type ActionResult struct {
	RequestID  string `json:"request_id"`
	DecisionID string `json:"decision_id,omitempty"`
	Authorized bool   `json:"authorized"`
	Executed   bool   `json:"executed"`
	OK         bool   `json:"ok"`
	Detail     string `json:"detail,omitempty"`
	Seq        int64  `json:"seq,omitempty"`
}

// Decision is structured audit of a cognitive cycle.
type Decision struct {
	ID              string            `json:"id"`
	OrganismID      string            `json:"organism_id"`
	ObservationID   string            `json:"observation_id,omitempty"`
	Beliefs         []Belief          `json:"beliefs,omitempty"`
	RuleVersion     string            `json:"rule_version,omitempty"`
	ModelProvider   string            `json:"model_provider,omitempty"`
	ModelVersion    string            `json:"model_version,omitempty"`
	Candidates      []string          `json:"candidates,omitempty"`
	Selected        string            `json:"selected,omitempty"` // empty = abstain
	AbstainReason   string            `json:"abstain_reason,omitempty"`
	Uncertainty     zyrion.Value      `json:"uncertainty,omitempty"`
	DirectorNote    string            `json:"director_note,omitempty"`
	PolicyResult    string            `json:"policy_result,omitempty"`
	AuthResult      string            `json:"auth_result,omitempty"`
	ExecutionResult string            `json:"execution_result,omitempty"`
	At              time.Time         `json:"at"`
	ProvenanceNote  string            `json:"provenance_note,omitempty"`
}

// CognitiveState is Mind's derived state (not Core operational truth).
type CognitiveState struct {
	OrganismID string             `json:"organism_id"`
	Beliefs    map[string]Belief  `json:"beliefs"`
	Decisions  []Decision         `json:"decisions,omitempty"`
	Pending    []ActionRequest    `json:"pending,omitempty"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

// EvaluationResult from a cycle before Core execution.
type EvaluationResult struct {
	Decision Decision
	Request  *ActionRequest // nil if abstain
}
