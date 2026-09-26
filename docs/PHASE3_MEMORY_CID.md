# Phase 3 — Memory + CID (completed)

**Date:** 2026-09-26

## Goal achieved

```
Organism → Memory (working / episodic / semantic) → CID blocks → Persist → Restore
```

## What was added

### `storage/cid`

- `cid.New(content) string` → `cid1:` + hex(sha256)
- `LocalStore` under `{DataDir}/blocks/` (sharded by first 2 hex chars)
- Interface: `Put` / `Get` / `Has`
- Integrity check on Get
- **Note:** Not IPFS CIDv1 yet (go-cid download failed in sandbox). Format is stable and documented for later upgrade without changing `Store` interface.

### `core/organism` memory model

```go
type Episode struct {
  ID, Type string
  Payload map[string]any
  ContentCID string  // optional block CID
  At time.Time
}

type MemoryRef struct {
  Working  map[string]string
  Episodic []Episode
  Semantic map[string]string
}
```

### Manager APIs

- `PutMemory` (working) — already existed, kept
- `PutSemantic`
- `AppendEpisode(id, type, payload, content []byte)` — stores blob in block store when content non-empty
- `GetBlock(contentCID)`
- `Blocks() cidstore.Store`

### CLI

```
prismatec organism memory set <id> <key> <value>
prismatec organism memory semantic <id> <key> <value>
prismatec organism memory episode <id> <type> [content...]
```

Inspect shows: `N working · M episodes · K semantic`

## What was NOT broken

- All Phase 1–2 CLI commands still work
- RootCID format unchanged (`rootcid:` + sha256 of manifest)
- Organism JSON schema extended only (additive fields)
- `go test ./...` green including `storage/cid`

## Verification

```bash
go test ./...

export PRISMATEC_DATA_DIR=/tmp/prismatec-phase3
go run ./cmd/prismatec organism create lab-agent --cap memory.read --cap inference
go run ./cmd/prismatec organism memory set lab-agent focus "sample A"
go run ./cmd/prismatec organism memory semantic lab-agent protocol PCR-v2
go run ./cmd/prismatec organism memory episode lab-agent observation "sample A positive"
go run ./cmd/prismatec organism inspect lab-agent
```

Episode line includes `cid=cid1:...` and block survives process restart via reload of Manager.

## Next phase

**Phase 4 — Events + Pulse**  
Unify internal EventBus toward outward Pulse (preparation for AIP / Alset-JS).

Do **not** skip to network or Studio.
