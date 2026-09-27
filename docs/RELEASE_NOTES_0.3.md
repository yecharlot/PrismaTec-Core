# PrismaTec Core 0.3 — operational readiness

## What raised the bar

| Area | Level |
|------|--------|
| Core organism runtime | Production-lab ready |
| Multi-node TCP + SIGKILL continuity | Proven in CI |
| Epoch fencing | Proven |
| libp2p transport | Proven (message path) |
| Mind + Zyrion + Bridge | Proven (symbolic + optional LLM annotation) |
| Cognitive decision on replicate/adopt | Proven (Phase 8 partial) |
| WorkOrder vertical CLI | Proven |

## Commands that matter

```bash
go test ./... -count=1
go test ./tests/e2e/ -v -timeout 180s
go run ./cmd/prismatec workorder run
go run ./cmd/prismatec node start   # AIP :8080 + Studio
```

## Honest ceiling

Not yet: partition chaos SLO, signed node messages, production LLM, multi-tenant SaaS.

## Version

`prismatec-core 0.3.0` (CLI version string aligned in this release polish).
