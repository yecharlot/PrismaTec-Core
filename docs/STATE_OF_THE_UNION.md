# PrismaTec Core — State of the Union (honest evaluation)

**Date (UTC):** 2026-09-26  
**Repo:** https://github.com/yecharlot/PrismaTec-Core  
**HEAD (at authorship):** see `git log -1` on main (continuity commit series through `c4e1ca6`+)  
**Audience:** human decision-makers + ChatGPT / external technical reviewers  
**Tone:** evidence over narrative. Claims are labeled **CAN DO NOW**, **CAN DO WITH CAVEATS**, **CANNOT CLAIM YET**.

Related evidence: `docs/EVIDENCE_FOR_REVIEWERS.md` (test logs, package results, continuity criteria).

---

## 1. What we actually have (one paragraph)

PrismaTec Core is a **Go runtime** for **Digital Organisms**: persistent entities with content identity (`RootCID`), host identity (`NodeID`), transport identity (`PeerID`), typed memory, capability/policy gates, pulse/event observation, HTTP+SSE **AIP**, a browser **Control Studio**, optional API token + audit log, execution hooks (builtin + WASM via wazero), inference **interfaces** (echo provider), and **multi-node distribution** over **TCP** and **libp2p**, including **replicate / recover**, **epoch fencing** against dual-primary, and an automated test that **SIGKILLs a real OS process** running the primary and recovers on a second process. It is **not** a finished global industrial fabric, not a full agent OS with continuous autonomy loops, and not a multi-tenant SaaS security boundary.

---

## 2. Capability matrix (self-evaluation)

| Domain | Score (0–10, judgment) | CAN DO NOW | Caveats |
|--------|----------------------|------------|---------|
| Organism model + lifecycle | 8 | Create/start/stop/inspect/persist/restore | App-level “goals/autonomy loops” not built-in |
| Identity separation (RootCID / NodeID / PeerID) | 8 | Enforced in model + network adapters | Peer auth between nodes still weak |
| Memory + CID | 7 | Working/episodic/semantic + local blocks | Not a full IPFS network; semantic is map-level |
| Policy / least privilege | 6 | Default-deny engine, wired on memory paths | Not all actions globally gated; no OPA |
| Observation (Pulse + AIP + Studio) | 7 | Real SSE pulses, Studio UI, no mock organism core | Studio doesn’t auto-send AIP token; no multi-node map in UI |
| Execution | 5 | Builtin ping/echo; WASM instantiate | No rich module ecosystem; limited sandbox policy |
| Inference | 4 | Provider interface + echo | No production LLM provider wired |
| Distribution TCP | 7 | Two nodes, replicate, recover, hard-kill test | Single-machine localhost focus in CI |
| Distribution libp2p | 5 | Message delivery test + Noise | No organism failover E2E on libp2p yet |
| Continuity / fencing | 7 | Epoch fence + SIGKILL process E2E | No partition/chaos suite; no formal RPO/RTO |
| Security (hostile multi-tenant) | 4 | Token, validation, audit JSONL | No signed node messages, mTLS, rate limits |
| Third-party developer product | 5 | USAGE.md, AIP commands, tests | No versioned SDK, stability guarantees, sample vertical app |

**Overall (self):** ~**6.0–6.5 / 10** as *infrastructure runtime prototype with proven multi-node continuity under test* — not 9+ production platform.

---

## 3. What you can build **today** on this stack

### CAN DO NOW (honest product uses)

1. **Local / lab digital organisms** with durable ID + memory + policy checks, driven by CLI or AIP.  
2. **Operator console** (`/studio/`) to create, start, set memory, watch pulses, run ping/infer echo.  
3. **Two-node edge lab**: primary + replica, replicate state, kill primary process, promote replica, keep RootCID + memory, run a command after recovery (as in `TestE2E_ProcessKillContinuity`).  
4. **Experimentation layer** for future agents: store working memory, attach capabilities, gate writes.  
5. **Protocol playground**: same `Transport` interface for TCP vs libp2p without rewriting organisms.

### CAN DO WITH CAVEATS

- **Demo for technical investors:** show tests + Studio + kill/recover story — **if** you frame it as runtime evidence, not “global autonomous OS.”  
- **Internal tools** on a trusted network with `PRISMATEC_AIP_TOKEN` set.  
- **WASM experiments** for small modules (load/instantiate; not a full plugin marketplace).

### CANNOT CLAIM YET

- Drop-in replacement for Kubernetes / Temporal / full agent platforms.  
- Untrusted multi-tenant public internet deployment.  
- Guaranteed continuity under arbitrary network partitions (split-brain beyond epoch unit test).  
- Production LLM agents “out of the box.”  
- Regulatory-grade provenance chain for compliance audits.  
- “Applications of another level” as **shipped verticals** — the Core is a **foundation**, not those apps.

---

## 4. Counterweight to prior ChatGPT reviews

| ChatGPT emphasis | Our counterweight (evidence) | Residual agreement |
|------------------|------------------------------|--------------------|
| “Declared TCP ≠ proven” | `TestE2E_MultiNodeTCP` + process kill E2E **PASS** | Still need partition chaos |
| “Recovery only flips status” | Memory + execute after recover **asserted** | — |
| “No hard kill” | `TestE2E_ProcessKillContinuity` **SIGKILL** | — |
| “No dual-primary control” | `Placement.Epoch` + `TestE2E_AntiDualPrimary` | Not full Raft/quorum |
| “Security thin” | Token + audit + validation exist | **Agree** — still thin for hostile nets |
| “Need real app on Core” | **Agree** — no vertical app in-repo | Highest product gap |
| “Docs contradictory” | Partially fixed; keep README/STATUS synced | Ongoing discipline |
| Distribution score ~50 | Should rise on **test evidence**; not to 90 | Score ≠ production readiness |

**Shared conclusion with ChatGPT’s better reviews:**  
Architecture is real; **trust for critical third-party apps** still requires contracts, chaos, security, and a reference application.

---

## 5. What “polish today” should mean (priority order)

1. Keep **evidence docs** and README **aligned** with HEAD (no stale “PeerID future”).  
2. Do **not** add random features; close gaps that change **trust**:  
   - partition test (optional next)  
   - libp2p organism failover E2E  
   - one **reference app** (even tiny)  
3. Investor narrative: only claims in §3 CAN DO NOW.

---

## 6. Explicit requests to ChatGPT (help where we are weak)

Please review this document + `docs/EVIDENCE_FOR_REVIEWERS.md` and help on:

### A. Continuity / distributed systems (we may stay weak without external design pressure)

1. Propose a **minimal** anti-split-brain protocol beyond single epoch integer (leases? compare-and-swap via AIP? single writer lock file?) suitable for **2–5 nodes**, not full Raft unless justified.  
2. Define **RPO/RTO** measurement methodology we can automate in Go tests.  
3. Design a **network partition test** outline (even if we implement later).

### B. Security (we are intentionally incomplete)

4. Threat model for “trusted LAN” vs “semi-hostile WAN.”  
5. Minimum message-signing scheme between nodes compatible with `network.Message`.  
6. What **not** to build yet (avoid security theatre).

### C. Product / developer experience (outside pure runtime coding)

7. Outline a **reference application** (one vertical, 1–2 week scope) that *must* use Organism + replicate/recover + policy — so the Core is not only infra demos.  
8. AIP **contract** suggestions: versioning, error codes, idempotency keys.  
9. Investor one-pager structure that stays honest given §3.

### D. Challenge our scores

10. Re-score the matrix in §2 with your own numbers and **force disagreement** where our 7–8 ratings are optimistic.

---

## 7. One-sentence truth

**We have a working, test-backed organism runtime with real multi-process kill/recover and basic fencing — enough to build serious labs and demos; not enough to sell as a finished planetary autonomous platform.**

---

*End of State of the Union. Re-run `go test ./tests/e2e/ -v -timeout 180s` after any claim change.*

---

## 8. Response to ChatGPT evaluation (2026-09-26)

**We accept:** global **6–6.5** as experimental runtime; continuity is **scenario-specific** not general; security **3–4**; third-party DX **4**; libp2p organism path unproven.

**Score adjustments we adopt:** Identity 7, Lifecycle 7, Continuity 6, Execution ≤5, Inference ≤3, Security ≤3–4.

**Implemented after this feedback:**

| Item | Status |
|------|--------|
| Authority semantics documented | `docs/AUTHORITY.md` |
| Confirmed-write `Seq` for RPO | `Organism.Seq` on PutMemory |
| AIP error codes + Idempotency-Key | `api/aip` |
| WorkOrder reference outline | `demos/workorder/README.md` |
| Honest claim language | this section + AUTHORITY |

**Still open (need design + time, ChatGPT roadmap):**

1. 3-voter / witness quorum (2-node cannot have HA + partition safety)  
2. Partition proxy tests  
3. Message signatures Ed25519 on `network.Message`  
4. Full WorkOrder app glue  
5. OpenAPI freeze  

**Claim we will publish:**

> Experimental Go runtime for persistent digital entities. Automated evidence shows TCP multi-node replicate/recover, including hard-kill of a primary OS process, post-recover execution, and epoch fencing. Partition safety, WAN security, and production continuity SLOs are **not** claimed.
