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
