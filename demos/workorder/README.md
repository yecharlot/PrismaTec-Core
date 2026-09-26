# Reference app: WorkOrder Organism (ChatGPT C7)

**Scope:** 1–2 week vertical — **not** a full product.

## Idea

A field **work order** is a Digital Organism:

| Field | Mapping |
|-------|---------|
| Definition | RootCID |
| Current host | Placement.Primary / NodeID |
| Client, equipment, status, notes | Working + episodic memory |
| Actions | Capabilities + policy |

## Suggested capabilities

- `order.read`, `order.note`, `order.status`, `order.close`

## Acceptance flow (uses Core as-is)

```bash
# Node A
go run ./cmd/prismatec node start   # with NETWORK peers configured

# Via AIP or Studio:
# 1. create workorder organism
# 2. memory.set status=open, client=...
# 3. replicate to B
# 4. SIGKILL A (or stop process)
# 5. recover on B
# 6. policy.check / memory.set note
# 7. assert Seq and RootCID
```

## What Core guarantees vs what the app guarantees

| Guarantee | Owner |
|-----------|--------|
| RootCID stability across recover | **Core** (tested) |
| Memory keys after replicate/recover | **Core** (tested) |
| Business meaning of `status=closed` | **App** |
| Immutable audit of commercial history | **App** (unless you add hash chain) |
| Partition-safe dual-primary | **Not Core yet** (see AUTHORITY.md) |

Implement the thin HTTP/CLI glue in a follow-up; Core already supplies create/memory/replicate/recover/policy.
