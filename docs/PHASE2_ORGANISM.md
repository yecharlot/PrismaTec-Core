# Phase 2 — Organism (completed)

**Date:** 2026-09-26

## What exists

- `core/organism` — canonical model + RootCID from manifest + lifecycle manager
- Status: `created | ready | running | stopped | error | recovering`
- Fields: ID, RootCID, Name, Capabilities, Policy (minimal), Memory (working map), Placement, Provenance, Runtime, CurrentAction
- Persistence: JSON under `{DataDir}/organisms/{id}.json`
- Events: `organism.created|started|stopped|updated`, `memory.updated`
- Registry sync on every change
- CLI:
  - `prismatec organism create <name> [--cap x] [--deny y]`
  - `prismatec organism list`
  - `prismatec organism inspect <id|name>`
  - `prismatec organism start <id|name>`
  - `prismatec organism stop <id|name>`

## What we reused

- AlsetOS lifecycle idea (creado → listo → ejecutando → detenido)
- Content-addressed RootCID concept (deterministic from definition)
- Capability as authorization boundary (policy still minimal)

## What we changed

- English canonical struct aligned with ARCHITECTURE.md
- RootCID derived from **manifest JSON** (name + version + capabilities + runtime), not only name
- Manager owns create/start/stop/persist/reload
- Policy map ready for Phase 7 (WHO/CAN/WHAT)
- Placement.Primary = NodeID (movement/replicas later)

## How we tested

```bash
go test ./...
# ok core/organism (+ previous packages)

PRISMATEC_DATA_DIR=/tmp/prismatec-phase2 go run ./cmd/prismatec organism create research-agent-01 \
  --cap memory.read --cap inference --deny database.write
go run ./cmd/prismatec organism start research-agent-01
go run ./cmd/prismatec organism inspect research-agent-01
go run ./cmd/prismatec organism stop research-agent-01
go run ./cmd/prismatec organism list
```

Organism survives process restart (reload from disk). Same manifest → same RootCID.

## Next (Phase 3)

Memory + CID:
- Working / episodic / semantic separation
- Optional upgrade RootCID to real multihash CID
- Demonstrate: Organism → Memory → persist → restore
