# PrismaTec Core

**Infrastructure runtime for persistent, autonomous digital entities (Digital Organisms).**

Not a chatbot framework. Not a UI kit.  
A **core layer**: identity, state, memory, capabilities, policy, execution, network, provenance, pulse, lifecycle — exposed through **AIP** and a **Control Studio**.

**Repository:** https://github.com/yecharlot/PrismaTec-Core

---

## Status (Phases 0–15)

| Area | Status |
|------|--------|
| Node boot, identity (NodeID), events, registry | ✅ |
| Organism lifecycle + RootCID | ✅ |
| Memory + CID store | ✅ |
| Pulse (localized updates) | ✅ |
| AIP v1 HTTP + SSE | ✅ |
| Control Studio `/studio/` | ✅ |
| Policy engine + audit + optional API token | ✅ |
| Execution (builtin + WASM) / inference interfaces | ✅ |
| Network fabric + replicate / recover (tests) | ✅ |
| E2E checklist `tests/e2e` | ✅ |

---

## Quick start

```bash
export GOPROXY=https://proxy.golang.org,direct
export GOTOOLCHAIN=go1.25.0

git clone https://github.com/yecharlot/PrismaTec-Core.git
cd PrismaTec-Core

go test ./...
go run ./cmd/prismatec node start
```

Open **http://127.0.0.1:8080/studio/**

```bash
go run ./cmd/prismatec demo e2e
go test ./tests/e2e/ -v
```

**Full usage guide:** [docs/USAGE.md](docs/USAGE.md)  
**Security:** [docs/SECURITY.md](docs/SECURITY.md)  
**Handoff for contributors/AI:** [docs/HANDOFF.md](docs/HANDOFF.md)

---

## Architecture (one glance)

```
IDENTITY + STATE + MEMORY + CAPABILITY + POLICY
+ EXECUTION + NETWORK + PROVENANCE + PULSE + INTERFACE
                    │
              Digital Organism
                    │
                   AIP
                    │
         Studio / JS / external clients
```

| Identity | Meaning |
|----------|---------|
| **RootCID** | Content / definition |
| **NodeID** | Host (Ed25519, persisted) |
| **PeerID** | Transport (future libp2p adapter) |

---

## Environment

| Variable | Meaning |
|----------|---------|
| `PRISMATEC_DATA_DIR` | State, identity, audit log |
| `PRISMATEC_NODE_NAME` | Display name |
| `PRISMATEC_AIP_ADDR` | Listen address (default `:8080`) |
| `PRISMATEC_AIP_TOKEN` | If set, protects `/aip/*` |

---

## Related repositories (unchanged)

- [AlsetOS](https://github.com/yecharlot/AlsetOS) — distributed execution heritage  
- [Prismatec](https://github.com/yecharlot/Prismatec) — intelligence / edge heritage  
- [Alset-JS-Runtime](https://github.com/yecharlot/Alset-JS-Runtime) — interaction / pulse UI heritage  

PrismaTec Core is an **evolution**, not a destructive merge of those repos.
