# Cumplimiento del manifiesto y visión Genesis

**Fecha:** 2026-09-26  
**Pregunta:** ¿Se cumple el manifiesto completo y la visión de plataforma distribuida?

## Respuesta operativa

**Sí en el alcance Core demostrable y distribuido (multi-nodo TCP real),** con evidencia automatizada:

| Criterio de éxito del manifiesto | Evidencia |
|----------------------------------|-----------|
| 1. Create → Pulse → UI real | Studio `/studio/` + AIP SSE + `TestE2E_AIP_HTTP` |
| 2. Move / replicate Node A → B | `TestE2E_MultiNodeTCP` + `organism replicate` + red TCP |
| 3. Failure → recover réplica | Mismo test + `organism recover` + RootCID preservado |

```bash
go test ./tests/e2e/ -count=1 -v
# incluye TestE2E_ManifestChecklist y TestE2E_MultiNodeTCP
```

## Componentes Genesis implementados

| Pieza | Implementación |
|-------|----------------|
| Digital Organism | `core/organism` |
| RootCID | Manifest hash / content id |
| NodeID | Ed25519 persistente |
| **PeerID** | `network.DerivePeerID` (distinto de NodeID) |
| Memory | working / episodic / semantic + CID |
| Capability + Policy | engine + authorize |
| Execution | builtin + WASM (wazero) |
| Inference | provider interface + echo |
| Network | **TCP multi-proceso** + fabric local tests |
| Replication / Recovery | `core/replication` + `Node.ReplicateOrganism` / `RecoverOrganism` |
| Pulse | hub + bridge + SSE |
| AIP | v1 HTTP/SSE/commands |
| Provenance | `core/provenance` en replicate/recover |
| Audit / security | token, validation, audit.jsonl |
| Interface | Control Studio |

## Cómo reproducir Demo 2 y 3 (no solo tests)

```bash
go run ./cmd/prismatec demo multinode
# o el procedimiento en docs/USAGE.md § Multi-node
```

Variables:

- `PRISMATEC_NETWORK_ADDR` — escucha TCP del nodo  
- `PRISMATEC_PEERS` — `nodeID@host:port,...`

## Qué significa “plataforma distribuida de próxima generación” aquí

**Garantizado ahora:**

- Entidades digitales **persistentes, identificables, ejecutables, observables**
- **Distribuibles** entre nodos reales (procesos/TCP), no solo memoria compartida
- Recuperables tras caída del primary
- Observables por UI y API versionada
- Policy + audit en el camino crítico

**Aún evolutivo (no bloquea el cumplimiento del MVP del manifiesto):**

- Adapter **libp2p** de producción (la interfaz `Transport` ya permite sustituir TCP)
- Providers LLM reales, Studio con token automático, mTLS, rate limits
- Carga industrial / edge fleet a escala

Esos puntos son **endurecimiento y escala**, no la definición mínima de Genesis en el propio manifiesto (organismo + identidades + pulse + red + réplica + recovery + AIP).

## Veredicto

El Core **deja de ser especulación** respecto a las promesas centrales del manifiesto y a los tres criterios de éxito.  
Quien ejecute `go test ./tests/e2e/ -v` y el flujo multi-nodo documentado **verifica** el sistema, no un mock.
