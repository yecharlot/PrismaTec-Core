# Mind + Zyrion integration (Phases 1–7)

**Principle:** Mind decides and proposes. Core governs and executes.

## Packages

| Package | Role |
|---------|------|
| `runtime/zyrion` | Ternary 0/1/2 + rules (epistemic only) |
| `runtime/mind` | Observe → Evaluate → ActionRequest → RecordResult |
| `runtime/mind.Bridge` | Policy + PutMemory/Execute only after authorize |

## Flow

```
Observation → Mind beliefs (Zyrion) → Director propose|abstain
    → Bridge maps action → policy.Evaluate
    → Core PutMemory / Execute
    → RecordResult (denied stays denied)
    → optional PersistDecision → semantic memory
```

## LLM

Optional `inference.Provider` on `mind.Config`. Annotations only; **cannot** invent actions without matching rules.

## Tests

```bash
go test ./runtime/zyrion/ ./runtime/mind/ -v
go test ./...   # baseline must stay green
```

## Not claimed

- Seven full organs as production subsystems  
- Distributed cognitive continuity (Phase 8 partial only)  
- LLM as authority  
- Partition-safe mind state  

## Phase status

| Phase | Status |
|-------|--------|
| 0 Diagnosis | done |
| 1 Contract | done |
| 2 Zyrion | done |
| 3 Director/abstain | done |
| 4 Core bridge | done |
| 5 Persist decision | done (semantic key) |
| 6 Optional LLM | done (echo annotation) |
| 7 WorkOrder test | done |
| 8 Cognitive replication | **not done** — no protocol change |
