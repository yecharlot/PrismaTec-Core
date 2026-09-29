# PrismaTec-Core — Status (EN)

**Indicative version:** 0.3.x · digital-organism runtime

## What it is

Product engine: identity, organisms, memory, default-deny policy, Pulse, provenance, AIP, Mind, Zyrion, Transport (TCP + libp2p), replication/recovery under test.

## What it is not

Not Pulso Cubano or AbacoPhy. Verticals consume Core through adapters.

## Working

- Multi-node TCP, PeerID, epoch fencing (e2e evidence)
- Mind proposes; Core authorizes and executes
- Studio / AIP for observation

## Industrial maturity still open

- Network partitions + measured RPO/RTO under load
- Inter-node message signing on hostile WAN
- Frozen versioned AIP contract for third parties
