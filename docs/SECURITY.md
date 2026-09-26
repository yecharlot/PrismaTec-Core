# Security (Phase 15)

PrismaTec Core applies **least privilege** and **default deny** as design defaults. Hardening is progressive; this document describes what is enforced today.

---

## Principles

1. **RootCID ≠ NodeID ≠ PeerID** — never conflate content, host, and transport identity.  
2. **Capability ≠ Policy** — declaring a skill does not grant unbounded rights.  
3. **Default deny** — if no policy rule matches, access is denied.  
4. **Audit** — security-relevant actions are recorded.  
5. **Validate inputs** — names, ids, actions, and memory payloads are bounded.

---

## What is enforced

| Control | Behavior |
|---------|----------|
| **Policy engine** | allow/deny, priority, wildcards (`memory.*`), conditions |
| **Authorize on memory** | `PutMemory` / episodes check policy |
| **Input validation** | name, id, action, memory key/value size (max 64KiB value) |
| **AIP token** | if `PRISMATEC_AIP_TOKEN` is set, `/aip/*` requires `Authorization: Bearer …` or `X-PrismaTec-Token` |
| **Audit log** | in-memory + `audit.jsonl` under `PRISMATEC_DATA_DIR` (mode 0600) |
| **Audit API** | `GET /aip/v1/audit` |
| **Data dir** | created with `0700` |
| **WASM** | wazero sandbox; timeout default 5s |
| **Auth failures** | audited as `auth.denied` |

---

## Enabling API authentication

```bash
export PRISMATEC_AIP_TOKEN="long-random-secret"
go run ./cmd/prismatec node start
```

```bash
curl -H "Authorization: Bearer long-random-secret" http://127.0.0.1:8080/aip/v1/info
```

Without the env var, AIP stays open for **local development** (Studio on same origin).

Static UI (`/studio/`, `/demo/`) and `/healthz` do not require the token so the browser can load the app; **API calls under `/aip/` do** when the token is set.

> Note: Studio’s `fetch` does not yet attach the token automatically. For token mode, call AIP with curl/SDK or extend the Studio client.

---

## Audit trail

Path: `$PRISMATEC_DATA_DIR/audit.jsonl`

Example event types:

- `command` — AIP command accepted/rejected  
- `auth.denied` — bad or missing token  
- (organism layer) `policy.denied` / `capability.executed` via event bus → pulse  

```bash
curl -s http://127.0.0.1:8080/aip/v1/audit | jq
```

---

## What is NOT done yet (honest backlog)

- Mutual TLS / SPIFFE  
- Signed Pulse messages end-to-end  
- Multi-tenant isolation across untrusted users  
- Rate limiting / DoS protection  
- Full WASM fuel/memory caps beyond timeout  
- Secret management integration  

Treat current Core as **trusted-operator local/edge** unless you add network controls and `PRISMATEC_AIP_TOKEN`.

---

## Reporting

Prefer fixing issues via the GitHub repo issues for `yecharlot/PrismaTec-Core`.
