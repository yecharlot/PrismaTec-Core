# Phase 16 — Vertical reference (WorkOrder) + product gap

**Status:** baseline present; expand as product layer outside pure Core.

## What landed before this note

- Core Phases 0–15: organism, memory, pulse, AIP, Studio, policy, WASM hooks, network TCP/libp2p, replicate/recover, E2E.
- Reference vertical: `prismatec workorder run` + `runtime/mind/workorder_test.go`.
- Docs: `demos/workorder/README.md`.

## Acceptance (this phase)

| Check | Command / evidence |
|-------|-------------------|
| WorkOrder CLI | `go run ./cmd/prismatec workorder run` |
| Policy deny without capability | Covered in `TestWorkOrderReferenceApp` |
| Decision persistence | `PersistDecision` / `LoadLastDecision` |
| Core still green | `go test ./...` |

## Next product steps (not Core scope creep)

1. **AbacoPhy as organism facade** — map invoice/entry CIDs to organism memory keys (adapter repo or thin module under `demos/`).
2. **Studio panel** for workorder status (real AIP data only).
3. Optional investor deck (slides) only after 1–2 verticals show real data.

## Non-goals

- Embedding full AbacoPhy ledger into Core.
- Public multi-tenant SaaS in this repo.
