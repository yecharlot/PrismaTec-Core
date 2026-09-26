# Phase 1 — Core Boot (completed)

**Date:** 2026-09-26

## What exists

- Repository layout matching the target architecture
- `go.mod` → `github.com/yecharlot/PrismaTec-Core`
- `core/identity` — Ed25519 NodeID, Save/Load/LoadOrCreate, Sign/Verify (adapted from AlsetOS)
- `core/events` — in-process EventBus (Subscribe / Publish)
- `core/registry` — in-memory organism registry (placeholder for Phase 2)
- `core/node.go` — Node with Start/Stop/Info, persistent identity
- `cmd/prismatec` — CLI: `node start`, `node info`, `version`

## What we reused

- AlsetOS identity model (Ed25519 → `node:<base64url(pubkey)>`)
- Separation of NodeID from RootCID / PeerID
- Lifecycle idea from AlsetOS node

## What we changed

- English package names aligned with architecture map
- Clean interfaces (no libp2p yet)
- EventBus as first-class primitive
- Registry as first-class primitive
- CLI focused on `prismatec` (not `alset-*`)

## How we tested

```bash
go test ./...
# ok  core, core/identity, core/events, core/registry

go run ./cmd/prismatec node start
# prints Name, NodeID, Status=running, DataDir, Organisms=0
# Ctrl+C → clean shutdown
```

Identity is persistent across restarts (same NodeID from same DataDir).

## Next (Phase 2)

Implement canonical Organism model + CLI:

```
prismatec organism create
prismatec organism list
prismatec organism inspect
prismatec organism start
prismatec organism stop
```
