# Phase 5 — AIP v1 (completed)

**Date:** 2026-09-26

## Qué se hizo

- Paquete `api/aip`: contrato versionado **AIP v1**
- Servidor HTTP + SSE montado al arrancar el nodo
- Endpoints:
  - `GET  /aip/v1/info`
  - `GET  /aip/v1/organisms`
  - `GET  /aip/v1/organisms/{id}`
  - `GET  /aip/v1/pulses` (historial reciente JSON)
  - `GET  /aip/v1/pulse` (SSE stream en vivo)
  - `POST /aip/v1/commands` (create | start | stop | memory.set)
  - `GET  /healthz`
- CORS abierto para demos browser locales
- CLI: `prismatec node start` levanta AIP (default `:8080`, env `PRISMATEC_AIP_ADDR`)
- Tests en `api/aip/server_test.go`

## Por qué

Sin un contrato **versionado** y un transporte, Alset-JS no puede consumir el Core. AIP es el puente manifesto:

```
PrismaTec Core ↔ AIP ↔ Alset JS / External clients
```

## Pensamiento futuro

- Auth (tokens) en AIP v1.1 — no bloquear demo local
- WebSocket como alternativa a SSE si el browser lo exige
- Pulse log sigue en RAM del proceso `node start` (correcto para MVP)
- No añadir endpoints de “apps” o Mind aquí (eso es application layer)

## Guía de acción (siguiente)

1. Fase 6: cliente JS mínimo (fetch + EventSource) que pinte un panel con datos de `/aip/v1/*`
2. No implementar Studio completo
3. Verificar Demo 1 end-to-end

## Cómo probar

```bash
export GOPROXY=https://proxy.golang.org,direct
export GOTOOLCHAIN=go1.25.0
export PRISMATEC_DATA_DIR=/tmp/ptc-aip
export PRISMATEC_AIP_ADDR=:8080

go test ./...
go run ./cmd/prismatec node start
# otra terminal:
curl -s http://127.0.0.1:8080/aip/v1/info | jq
curl -s -X POST http://127.0.0.1:8080/aip/v1/commands \
  -H 'Content-Type: application/json' \
  -d '{"aip":"v1","action":"create","params":{"name":"demo","capabilities":["memory.read"]}}'
```
