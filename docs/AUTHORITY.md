# Authority semantics (epoch fencing) — ChatGPT-aligned

## What constitutes authority today

| Rule | Behavior |
|------|----------|
| Initial create | `Placement.Epoch = 1`, `Primary = creating NodeID` |
| Successful recovery | `Epoch = prev+1`, `FencedFrom = old primary`, `Primary = recovering node` |
| Stale adopt | Rejected if incoming `Epoch < local Epoch` (`STALE_EPOCH` path / Adopt error) |
| Dual primary same epoch | Rejected if local already `running` under another Primary |

## What is NOT guaranteed

- **2-node majority during partition:** with only two nodes, safe promotion after isolation of the primary requires either accepting unavailability or a **third witness** (ChatGPT A1). Current recover promotes when primary is **not in live peer set** — that is **liveness-oriented**, not partition-safe.
- Votes are **not** multi-node quorum ballots yet; epoch is local monotonic after recovery path.
- Clocks are not used as authority (good); epoch is.

## Safe operational rule (document for operators)

> If you cannot prove higher epoch authority, **do not write**.

For production 2-node HA with automatic failover, plan **3 voters** (2 runtime + 1 witness) before claiming partition safety.

## Confirmed writes (RPO)

`Organism.Seq` increments on each successful `PutMemory` (confirmed local write).  
RPO tests should compare last Seq on A before failure vs Seq on B after recover.
