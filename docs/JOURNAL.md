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
