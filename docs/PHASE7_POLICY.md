# Phase 7 — Policy Engine (completed)

**Date:** 2026-09-26

## Qué se hizo

### `core/policy`

Motor de autorización:

```text
WHO (Subject) can DO WHAT (Action) on WHICH (Resource) under CONDITIONS
```

- `Rule` con Effect allow/deny, Priority (deny gana en empate)
- Wildcards: `memory.*`, subject/resource `*`
- `Evaluate` / `Authorize` → `Decision` o error `Denied`
- `FromCapabilityMap` para migrar `map[string]bool` de Fases 2–6

### Integración

- `organism.EngineFor(org)` construye engine desde `Policy.RulesList` o `Rules` legacy
- `Manager.authorize` en `PutMemory` / `AppendEpisode`
- Eventos: `policy.denied`, `capability.executed`
- AIP: `POST /aip/v1/commands` action `policy.check`

### Tests

- `core/policy` unitarios (priority, wildcard, conditions, default deny)
- `TestPolicyDenyMemoryWrite` en organism

## Por qué

El manifiesto exige separación Capability ≠ Policy y least privilege desde el diseño, sin IAM enterprise completo todavía.

## Pensamiento futuro

- PolicyProvider adapter (OPA, cloud IAM)
- Firmas en reglas / audit trail persistente
- `requires approval` flow (humano en el loop) como rule effect futuro

## Guía de acción

1. `go test ./...`
2. Crear organismo con `--deny memory.write` y verificar deny
3. Siguiente: Fase 8 Execution (WASM) o Fase 10 Network según Demo 2/3

## Ejemplo

```bash
# policy.check via AIP
curl -s -X POST http://127.0.0.1:8080/aip/v1/commands \
  -H 'Content-Type: application/json' \
  -d '{"aip":"v1","action":"policy.check","params":{"id":"ORG_ID","action":"memory.write"}}'
```
