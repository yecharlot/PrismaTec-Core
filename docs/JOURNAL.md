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

### Guía de acción
1. Publicar repo en GitHub siguiendo docs/GITHUB.md
2. Fase 6 Demo 1
3. GOPROXY=https://proxy.golang.org,direct y GOTOOLCHAIN=go1.25.0


## 2026-09-26 — GitHub publish + Fase 6 Demo 1

### Qué se hizo
- Repo público: https://github.com/yecharlot/PrismaTec-Core
- demos/aip-panel + served at /demo/
- Documentación PHASE6

### Por qué
Demo 1 del manifiesto + visibilidad del código en GitHub.

### Seguridad
El token de GitHub usado para el push quedó expuesto en el chat: **revocar y generar uno nuevo** en GitHub → Settings → Developer settings → Tokens.

### Guía de acción
1. Revocar token expuesto
2. Abrir demo en browser con node start
3. Fase 7 o 10 según prioridad


## 2026-09-26 — Fase 7 Policy

### Qué se hizo
- core/policy Engine (allow/deny, priority, wildcards, conditions)
- Integración organism Manager + AIP policy.check
- Tests + docs PHASE7

### Por qué
Authorization real antes de execution/network.

### Guía de acción
Push a GitHub; siguiente Fase 8 o 10.


## 2026-09-26 — Fases 8–12 batch

### Qué se hizo
- runtime/execution + wasm (wazero) + inference providers
- network LocalTransport fabric
- core/replication Replicate + RecoverIfPrimaryDown
- AIP execute/infer
- docs PHASES_8_TO_12.md

### Por qué
Cerrar el camino técnico a Demos 2/3 sin esperar libp2p completo.

### Guía de acción
Push GitHub; opcional CLI multi-node; Phase 13 Studio.


## 2026-09-26 — Fase 13 Control Studio

### Qué se hizo
- demos/studio Control Studio UI
- Routes /studio/ and / still → studio
- docs PHASE13

### Por qué
Plano de control observable sobre Core real.

### Guía de acción
Push GitHub; node start → /studio/
