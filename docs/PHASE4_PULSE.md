# Phase 4 — Events + Pulse (completed)

**Date:** 2026-09-26

## Why this phase

The manifesto requires a **localized** communication path:

```
Core → Pulse → organism | node | (later UI node)
```

Without Pulse, Alset-JS and AIP cannot receive real updates. Events (internal bus) already existed; Phase 4 makes them **outward-addressable** as Pulses.

## What we did

### `core/pulse`

- `Pulse{ Target, Type, Data, Meta, Source, Timestamp }`
- `Hub`: Subscribe / SubscribeTarget / Emit / Recent / Count / Seq
- Ring buffer log (default 256) for CLI and future AIP backfill

### Bridge

In `core/node.go` on `NewNode`:

```
EventBus.SubscribeAll → Pulse.Hub.Emit
```

- Target = organism `id` from payload when present, else NodeID
- Meta includes `bridged_from: events` and monotonic `seq`

### CLI

```
prismatec pulse list [n]
```

(In-memory log is per process in Phase 4; list also emits `organism.observed` for loaded organisms so the command is useful across CLI invocations.)

### CID fix (Phase 3 residual)

| Problem | Cause | Solution |
|---------|--------|----------|
| `go get go-cid` 502 | `GOPROXY=http://35.245.43.102/go/` broken | Use `GOPROXY=https://proxy.golang.org,direct` |
| go-cid needs Go ≥ 1.25 | Module requirement | Toolchain upgraded to **go 1.25.0** (works in env) |
| Dual format | Compatibility | Keep `cid1:` default; add `NewIPFS` / `PutIPFS` / Valid for IPFS CIDv1 |

## Tests

```bash
go test ./...
# core/pulse + TestEventBridgeToPulse in core
# storage/cid includes TestNewIPFSAndPutIPFS
```

## What we did NOT do (correctly deferred)

- AIP transport (SSE/WebSocket) → Phase 5
- Alset-JS connection → Phase 6
- Network gossip of pulses → Phase 10+

## Future thought

Pulse log is RAM-only. Before multi-process demos, either:

1. Keep pulses in the long-running `node start` process and attach AIP there, or
2. Persist a small pulse journal under DataDir (optional, not required for MVP if node stays up).

## Action guide (next AI)

1. Read `docs/HANDOFF.md`
2. `go test ./...`
3. Start **Phase 5 — AIP v1** (versioned contract, minimal HTTP/SSE or WebSocket)
4. Do not implement Studio or libp2p yet
