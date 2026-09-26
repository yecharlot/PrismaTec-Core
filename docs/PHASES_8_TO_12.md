# Phases 8–12 (completed)

**Date:** 2026-09-26

## Overview

| Phase | Package | What |
|-------|---------|------|
| 8 Execution | `runtime/execution`, `runtime/wasm` | Engine interface, builtin + wazero WASM |
| 9 Inference | `runtime/inference` | Provider interface, Echo + HTTP stub |
| 10 Network | `network` | Transport/Discovery-ready interfaces, **LocalTransport** fabric |
| 11 Replication | `core/replication` | Primary + replicas over Transport |
| 12 Recovery | `core/replication` | Promote replica when primary offline |

## Why local fabric instead of full libp2p first

libp2p is the intended production adapter (AlsetOS). Phase 10 delivers **clean interfaces** + a working multi-node fabric in-process so Demo 2/3 logic is testable without blocking on DHT wiring. libp2p adapter can implement the same `Transport` later without changing replication/recovery.

## Phase 8 — Execution

```text
Organism → execute → Engine (builtin | wasm)
```

- `execution.Registry` of engines
- `BuiltinEngine`: `echo`, `ping`
- `wasm.Engine`: wazero, load bytes or file, call export (or instantiate-only)

**AIP:** `action: execute` (builtin)

## Phase 9 — Inference

- `inference.Provider` / `Registry`
- `EchoProvider` (deterministic local)
- `HTTPProvider` stub (explicit “not fully wired”)

**AIP:** `action: infer`

## Phase 10 — Network

```go
Transport: Listen, Send, Broadcast, OnMessage, Peers
```

`LocalTransport` + `Fabric` = multiple nodes in one process (or tests).

## Phase 11–12 — Replicate + Recover

```text
Primary node-A  --replicate-->  Replica node-B
node-A offline  -->  B.RecoverIfPrimaryDown  -->  Primary=B, RootCID preserved
```

Verified by `TestReplicateAndRecover`.

## Tests

```bash
export GOPROXY=https://proxy.golang.org,direct GOTOOLCHAIN=go1.25.0
go test ./runtime/... ./network/... ./core/replication/... ./...
```

## Future

- libp2p adapter implementing `network.Transport`
- Wire replication into long-running `node start` + CLI `organism move|recover`
- Sandbox resource limits on WASM
- Real HTTP/OpenAI providers behind same interface

## Action guide

1. Keep interfaces stable
2. Next product focus: CLI/AIP for multi-node demo or Studio (Phase 13)
3. Do not couple Organism struct to libp2p
