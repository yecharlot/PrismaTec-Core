# Red descentralizada: libp2p detrás de `network.Transport`

## Por qué

PrismaTec Core nace como plataforma **descentralizada**. TCP punto-a-punto valida Demo 2/3; **libp2p** es el transporte industrial-compatible (Noise, multiplexado, multiaddr, PeerID nativo) **sin cambiar** organismos ni replication.

```
Organism / Replication  →  network.Transport  →  TCP | LibP2P | (futuro: QUIC-only, relays)
```

## Activación

```bash
export PRISMATEC_TRANSPORT=libp2p
export PRISMATEC_NETWORK_ADDR=/ip4/0.0.0.0/tcp/4001
# peers: full multiaddr with /p2p/<peerID>
export PRISMATEC_PEERS='<remoteNodeID>@/ip4/127.0.0.1/tcp/4002/p2p/<remotePeerID>'
go run ./cmd/prismatec node start
```

Si `PRISMATEC_NETWORK_ADDR` empieza por `/`, el nodo elige libp2p automáticamente.

## Identidades

| ID | Rol |
|----|-----|
| **NodeID** | Identidad de negocio / host PrismaTec (Ed25519 persistente) |
| **PeerID** | Identidad de transporte libp2p (`host.ID()`) |
| **RootCID** | Identidad de contenido del organismo |

Nunca se fusionan.

## Protocolo

- Protocol id: `/prismatec/aip/1.0.0`
- Payload: mismos `network.Message` JSON que TCP (replicate, activate, …)
- Security: **Noise** (go-libp2p default stack)

## Prueba

```bash
go test ./network/ -run LibP2P -v
go test ./tests/e2e/ -v   # TCP multi-node + checklist; libp2p unit en network
```

## Camino industrial (honesto)

| Ya | Siguiente escala |
|----|------------------|
| Host libp2p + streams + Noise | DHT / gossipsub para discovery |
| PeerID real | Identity key persistida en DataDir |
| Mismo Replication/Recover | Circuit relay, AutoNAT, hole punching |
| TCP + libp2p intercambiables | Observabilidad métricas Prometheus |

La soberanía tecnológica viene de: **código abierto, identidades propias, policy local, datos en tu DataDir, transporte P2P sin vendor lock-in de un solo cloud**.
