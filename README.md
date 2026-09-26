# PrismaTec Core

**Infrastructure runtime for persistent, autonomous and distributed digital entities.**

PrismaTec Core provides identity, state, memory, capabilities, policy, execution, communication, provenance and lifecycle management for autonomous digital entities across local, edge and cloud environments.

This is **not** another agent framework, frontend framework, or chatbot.  
It is the infrastructure layer on which those systems can be built.

---

## Status

**Phase 2 — Organism** (current)

- [x] Phase 1 — Core Boot (identity, events, registry, node)
- [x] Canonical Organism model + RootCID from manifest
- [x] Lifecycle: create / start / stop / inspect / persist / restore
- [x] CLI: `organism create|list|inspect|start|stop`
- [x] Unit tests + end-to-end CLI verification
- [x] `core/identity` (NodeID via Ed25519, persistent)
- [x] `core/events` (EventBus)
- [x] `core/registry` (organism registry)
- [x] Minimal Node lifecycle (`start` / `stop`)
- [x] CLI: `prismatec node start` / `node info` / `version`
- [x] Unit tests passing

**Next:** Phase 3 — Memory + CID

---

## Quick start

```bash
cd PrismaTec-Core

# Run tests
go test ./...

# Start a local node (persists identity under ~/.prismatec/node or PRISMATEC_DATA_DIR)
go run ./cmd/prismatec node start

# Show identity without starting
go run ./cmd/prismatec node info
```

Environment:

| Variable | Meaning |
|----------|---------|
| `PRISMATEC_DATA_DIR` | Directory for identity and local state |
| `PRISMATEC_NODE_NAME` | Human-readable node name |

---

## Architecture (summary)

```
PRISMATEC CORE
      │
  IDENTITY · ORGANISMS · EVENTS · MEMORY · CAPABILITIES
  POLICY · EXECUTION · NETWORK · PROVENANCE · PULSE
      │
     AIP
      │
  Alset JS / External clients
```

See `docs/` for full audit, architecture map, migration plan and ADRs.

---

## Identities (must stay separate)

| Identity | Meaning |
|----------|---------|
| **RootCID** | Content / organism definition identity |
| **NodeID** | Persistent cryptographic identity of the host node |
| **PeerID** | Transport identity (libp2p) — Phase 10+ |

---

## Origins

PrismaTec Core unifies the architectural core of:

- [AlsetOS](https://github.com/yecharlot/AlsetOS) — distributed runtime / organism fabric
- [PrismaTec](https://github.com/yecharlot/PrismaTec) — intelligence, memory, coordination
- [Alset-JS-Runtime](https://github.com/yecharlot/Alset-JS-Runtime) — human/machine experience (AIP, Pulse)

Original repositories are **not** destroyed; this is an evolution.
