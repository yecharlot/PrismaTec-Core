# JOURNAL — Bitácora de sesiones

## 2026-09-26 — Fases 0–4 + política de docs

### Qué se hizo
- Auditoría de AlsetOS, PrismaTec, Alset-JS-Runtime
- Fase 1 Core Boot (identity, node, events, registry, CLI)
- Fase 2 Organism (modelo, lifecycle, RootCID, persistencia, CLI)
- Fase 3 Memory + CID (`cid1:` blocks, working/episodic/semantic)
- Diagnóstico fallo `go-cid`: GOPROXY interno 502 → usar `https://proxy.golang.org,direct`
- Integración `github.com/ipfs/go-cid` + `NewIPFS` / `PutIPFS` (Go 1.25)
- Fase 4 Pulse hub + bridge EventBus→Pulse + CLI `pulse list`
- Documentación: HANDOFF, STATUS, PHASE1–4, política de documentación continua

### Por qué
Cumplir el manifiesto: primitives primero, demo real después, continuidad entre IAs.

### Pensamiento futuro
- AIP debe servir desde `node start` (proceso largo), no desde CLI efímera
- Pulse log RAM es suficiente hasta AIP; persistir solo si hace falta
- No meter libp2p hasta tener interfaces Network limpias

### Guía de acción siguiente
1. `go test ./...` con `GOPROXY=https://proxy.golang.org,direct` si hace falta
2. Fase 5: AIP v1 mínimo (discovery, organism snapshot, pulse stream)
3. Actualizar PHASE5 + HANDOFF + STATUS + esta JOURNAL
4. Luego Fase 6 Alset-JS mínimo → Demo 1

### Verificación
```bash
cd /home/workdir/artifacts/PrismaTec-Core
go test ./...
# ok core, organism, pulse, storage/cid, ...
```

## 2026-09-26 — Fase 5 AIP v1 + ruta GitHub

### Qué se hizo
- `api/aip` tipos + servidor HTTP/SSE
- Endpoints info, organisms, pulses, pulse SSE, commands
- `node start` escucha AIP (PRISMATEC_AIP_ADDR, default :8080)
- Tests API + smoke curl real
- `docs/GITHUB.md` — cómo publicar en github.com/yecharlot/PrismaTec-Core
- `.gitignore` añadido

### Por qué
Cerrar el contrato Core↔clientes antes de UI.

### Pensamiento futuro
Fase 6 = panel mínimo en JS (puede vivir en demos/ o extraer de Alset-JS-Runtime solo Registry+EventSource).

## 2026-09-26 — GitHub publish + Fase 6 Demo 1
- Repo público: https://github.com/yecharlot/PrismaTec-Core
- `demos/aip-panel` + served at `/demo/`
- Documentación PHASE6

## 2026-09-26 — Fase 7 Policy
- `core/policy` Engine (allow/deny, priority, wildcards, conditions)
- Integración organism Manager + AIP policy.check
- Tests + docs PHASE7

## 2026-09-26 — Fases 8–12 batch
- `runtime/execution` + `wasm` (wazero) + inference providers
- `network` LocalTransport fabric
- `core/replication` Replicate + RecoverIfPrimaryDown
- AIP execute/infer
- docs PHASES_8_TO_12.md

## 2026-09-26 — Fase 13 Control Studio
- `demos/studio` Control Studio UI
- Routes /studio/ and / still → studio
- docs PHASE13

## 2026-09-26 — Fase 14 E2E
- `tests/e2e` TestE2E_ManifestChecklist (12 steps) + AIP HTTP
- `prismatec demo e2e`
- docs PHASE14

## 2026-09-26 — Fase 15 Security + USAGE
- `core/audit`, `core/security`
- AIP token + validation + audit API
- docs/USAGE.md, SECURITY.md, PHASE15, README

## 2026-09-26 — Cierre manifiesto: multi-nodo TCP
- `network.TCPTransport` multi-proceso
- PeerID distinto de NodeID
- Node.ReplicateOrganism / RecoverOrganism + Adopt
- Provenance en replicate/recover
- TestE2E_MultiNodeTCP (Demo 2+3)
- CLI/AIP + docs MANIFESTO_COMPLIANCE + USAGE multi-node

## 2026-09-26 — libp2p Transport adapter
- `network.LibP2PTransport` (Noise, /prismatec/aip/1.0.0)
- Node buildTransport tcp|libp2p
- Tests LibP2P + docs NETWORK_LIBP2P + INDUSTRIAL_VISION

## 2026-09-26 — Evidence pack for reviewers
- Ran `go test ./...` green at `976201b`
- Wrote `docs/EVIDENCE_FOR_REVIEWERS.md` for reviewers/auditors

## 2026-09-27 — Phase 16 vertical baseline
- Documenté `PHASE16_VERTICAL.md` (WorkOrder como app de referencia)
- Alineó STATUS/HANDOFF: Core 0–15 hecho; gap = verticales de producto
- Verificación: WorkOrder test + CLI documentados

## 2026-11-16 — Contrato universal de integración v0.1.0

### Qué hice
- Creé `docs/APPLICATION_INTEGRATION_CONTRACT.md`.
- Definí la separación normativa Aplicación → Adaptador → PrismaTEC-Core → Proveedores.
- Establecí contratos para organismos de dominio, identidad, relaciones, episodios, operaciones, idempotencia, atomicidad, concurrencia, autorización, replicación, Mind/Zyrion, errores, versionado y pruebas de conformidad.
- Incorporé Smart-WMS como primera implementación de referencia y documenté las lecciones de receiving, picking, ledger y frontera Core.

### Por qué
Necesito que otras aplicaciones puedan reutilizar las primitivas de PrismaTEC-Core sin copiar la arquitectura del WMS ni trasladar reglas verticales al núcleo.

### Estado y límites
- La especificación es normativa a nivel arquitectónico, pero todavía no es un SDK ni una API implementada.
- No ejecuté pruebas Go ni certifiqué el estado actual de Smart-WMS en esta actualización.
- La conformidad de Smart-WMS debe verificarse contra el código vigente antes de declarar que cumple el contrato.

### Siguiente paso
Auditar la rama vigente de Smart-WMS y construir una matriz requisito → código → prueba → estado, antes de diseñar interfaces públicas adicionales.
