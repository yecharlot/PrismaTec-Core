# Phase 13 — PrismaTec Control Studio (completed)

**Date:** 2026-09-26

## Qué se hizo

- `demos/studio/` — **PrismaTec Control Studio** (control plane UI)
- Vistas: Overview, Organisms, Pulse, Execution, Policy
- Datos **reales** vía AIP v1 + SSE
- Servido en `http://localhost:8080/studio/` (redirect `/` → `/studio/`)
- Panel Demo 1 sigue en `/demo/`

## Observa / controla

| Área | UI |
|------|-----|
| Node + health | Overview |
| Organisms list + inspector | Organisms |
| Pulse stream | Pulse |
| Execute / Infer | Execution |
| Policy check | Policy |

## Por qué

El manifiesto pide un Studio como plano de control: nodos, organismos, memoria, pulse, capabilities, policy, execution — sin portar aún todo Alset Studio legacy.

## Pensamiento futuro

- Network / replicas / recovery views cuando multi-nodo esté en AIP
- Auth antes de exponer en red
- Opcional: integrar componentes de Alset-JS-Runtime

## Guía de acción

```bash
go run ./cmd/prismatec node start
# browser → http://127.0.0.1:8080/studio/
```

Siguiente: Phase 14 E2E demo checklist o Phase 15 security.
