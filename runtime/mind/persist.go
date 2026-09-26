package mind

import (
	"encoding/json"
	"fmt"

	"github.com/yecharlot/PrismaTec-Core/core/organism"
)

const decisionMemoryKey = "mind.last_decision"

// PersistDecision stores the last structured decision in organism semantic memory (JSON).
// Does not invent execution results — only what Mind already recorded.
func PersistDecision(orgs *organism.Manager, organismID string, d Decision) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	o, err := orgs.Get(organismID)
	if err != nil {
		return err
	}
	if o.Memory.Semantic == nil {
		// PutSemantic path
	}
	_, err = orgs.PutSemantic(organismID, decisionMemoryKey, string(b))
	if err != nil {
		return fmt.Errorf("persist decision: %w", err)
	}
	return nil
}

// LoadLastDecision reads mind.last_decision from semantic memory.
func LoadLastDecision(orgs *organism.Manager, organismID string) (*Decision, error) {
	o, err := orgs.Get(organismID)
	if err != nil {
		return nil, err
	}
	raw, ok := o.Memory.Semantic[decisionMemoryKey]
	if !ok || raw == "" {
		return nil, fmt.Errorf("no persisted decision")
	}
	var d Decision
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return nil, err
	}
	return &d, nil
}
