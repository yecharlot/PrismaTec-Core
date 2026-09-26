# Phase 6 — Alset-JS mínimo / Demo 1 (completed)

**Date:** 2026-09-26

## Qué se hizo

- `demos/aip-panel/` — panel browser (HTML/CSS/JS) sin framework
- Consume **solo** datos reales vía AIP v1 (`/aip/v1/info`, organisms, commands, SSE pulse)
- Actualización **localizada**: solo reescribe campos DOM si el valor cambió; refresh de organismo cuando el pulse `target` coincide
- Servido por el propio Core: `http://localhost:8080/demo/` (misma origen → sin CORS issues)
- Redirect `/` → `/demo/`

## Por qué

Cierra el **Demo 1** del manifiesto:

```
Browser → panel → AIP → Core → Create/Start → Pulse → UI actualiza con datos reales
```

No se portó Alset Studio completo (application layer). Se adoptó la filosofía Alset-JS (PIN / localized resonance) en forma mínima.

## Pensamiento futuro

- Extraer este panel hacia `sdk/js` cuando estabilice el contrato
- Opcional: importar `AlsetPulseCore.js` del repo Alset-JS-Runtime en una fase Studio
- Auth en AIP antes de exponer en red pública

## Guía de acción

1. `go run ./cmd/prismatec node start`
2. Abrir http://localhost:8080/demo/
3. Create → Start → Memory set → observar STATUS y PULSE STREAM
4. Siguiente fase: 7 Policy o 10 Network según prioridad de Demo 2/3

## Verificación

```bash
export GOPROXY=https://proxy.golang.org,direct GOTOOLCHAIN=go1.25.0
go test ./...
PRISMATEC_AIP_ADDR=:8080 go run ./cmd/prismatec node start
# browser: http://127.0.0.1:8080/demo/
```
