# Fase 0 — Diagnóstico Mind + Zyrion en PrismaTec Core

**Fecha UTC:** 2026-09-26  
**Commit línea base:** `4d7cfc625b019292c9f96c511929aa0b894c1864`  
**Git:** `main` limpio, alineado con `origin/main`  
**Pruebas:** `go test ./... -count=1` → **0 FAIL** (todos los paquetes con tests en ok)

**Esta fase no modifica código.** Solo inspecciona, mide y propone el cambio mínimo.

---

## A. Diagnóstico — qué existe realmente

### Núcleo operativo (probado)

| Módulo | Rol real | Evidencia |
|--------|----------|-----------|
| `core/organism` | Identidad, RootCID, memoria working/episodic/semantic, lifecycle, Seq, Placement.Epoch | tests + E2E |
| `core/policy` | Default-deny, WHO/CAN/WHAT | tests |
| `core/events` | Bus interno Publish/Subscribe | tests |
| `core/pulse` | Hub de updates localizados | tests + AIP SSE |
| `core/provenance` | Log simple de acciones | poco testeado |
| `runtime/execution` | Engine interface + Builtin | tests |
| `runtime/inference` | **Provider interface** + Echo + HTTP stub | tests; sin vendor hardcode |
| `runtime/wasm` | WASM mínimo (wazero) | tests |
| `api/aip` | HTTP/SSE/commands, códigos error, Idempotency-Key | tests |
| `network` | Transport TCP + libp2p | tests + continuity SIGKILL |
| `core/replication` | Replicate/recover + epoch | tests |

### Ausente / vacío (carpetas sin implementación Mind)

- `runtime/agents/` — vacío  
- `runtime/lisp/` — vacío (LispAI no integrado)  
- `core/capability/`, `core/memory/`, `core/lifecycle/`, `core/execution/` — vacíos o solo placeholders; la lógica real está en `organism`, `runtime/execution`  
- **Mind** — no existe como paquete  
- **Zyrion** — no existe en este repo  
- **Siete órganos** — no hay código  

### Separación ya alineada con la misión

```
Capability (en organismo) → Policy engine → authorize en Manager → Execution / AIP commands
```

Mind **no** debe saltarse esta cadena. El punto de extensión correcto es: **proponer acciones**, no ejecutar.

### Contrato de inferencia reutilizable

`runtime/inference.Provider` ya es el adaptador correcto para LLM opcional:

- `Infer(ctx, Request) (Response, error)`  
- Echo funciona offline  
- HTTP stub no cableado (bien para Fase 6, no Fase 1)

### Flujo de observación existente

- EventBus (`core/events`) + Pulse (`core/pulse`) + AIP SSE  
- AIP actions: create, start, stop, memory.set, execute, infer, policy.check, replicate, recover  

Mind puede **consumir** eventos/pulses y **emitir solicitudes** que Core autorice vía los mismos caminos (`execute`, `memory.set`, o un comando `mind.request` futuro versionado).

---

## B. Decisión — cambio mínimo (plan, no código)

### Principio

**Mind decide y propone. Core gobierna y ejecuta.**

### Ubicación propuesta (menos invasiva)

```
runtime/mind/          # paquete nuevo opcional
  types.go             # Observation, Belief, Decision, ActionRequest, ActionResult
  mind.go              # interfaz Mind + stub
  director.go          # abstención / proponer (Fase 3)
runtime/zyrion/        # paquete nuevo opcional
  ternary.go           # 0|1|2 + Evaluate
  rules.go             # reglas versionadas deterministas
```

**No tocar en Fase 1–4 salvo cableado mínimo:**

- Identidades, persistencia de organismos, TCP/libp2p, replication protocol, AIP paths existentes  
- Formato JSON de organismos en disco  

**Cableado mínimo a Core (Fase 4):**

1. `Mind.Observe(event)` suscrito a EventBus o invocado desde Manager tras eventos clave.  
2. `Mind.Propose()` → `ActionRequest`.  
3. Core: `policy.Authorize` + `execution` o rechazo.  
4. `Mind.RecordResult(ActionResult)` solo con resultado **confirmado** por Core.  
5. Persistencia cognitiva: preferir **memoria semantic/episodic del organismo** + CID, o archivo bajo `DataDir/mind/` por organism ID — sin segundo bus.

**AIP:** preferible **sin endpoints nuevos** en Fase 1–5; pruebas unitarias + CLI interna. Si hace falta API: `/aip/v1` action `mind.tick` opcional, campos solo aditivos.

**LLM:** solo Fase 6; Fase 1–5 100% simbólico/determinista.

---

## C. Cambios previstos por fase (archivos)

| Fase | Añadir | Tocar con cuidado | No tocar |
|------|--------|-------------------|----------|
| 0 | este doc | — | todo el código |
| 1 | `runtime/mind/types.go`, `mind.go`, `*_test.go` | — | organism persist format |
| 2 | `runtime/zyrion/*_test.go` | — | policy, network |
| 3 | `runtime/mind/director.go` | — | AIP público |
| 4 | puente delgado en `core/organism` o servicio en `Node` | 1–2 archivos | replication wire format |
| 5 | persist decisión vía memory/CID + provenance | provenance | — |
| 6 | usar `inference.Provider` | — | hardcode OpenAI |
| 7 | `demos/workorder` o CLI tick | demos | Studio rewrite |
| 8 | solo tras 7 + tests | replication solo si justificado | protocol break |

---

## D. Compatibilidad

- Tests baseline verdes: **sí**.  
- Mind como **dependencia opcional** del Node (nil = comportamiento actual).  
- Organismos existentes sin Mind siguen idénticos.  
- Policy/execution sin cambios semánticos en fases tempranas.

---

## E. Pruebas (Fase 0)

**Ejecutado:**

```text
commit: 4d7cfc625b019292c9f96c511929aa0b894c1864
go test ./... -count=1 -timeout 120s
→ all packages ok (incl. tests/e2e ~9.5s)
```

**No ejecutado en esta fase:** ninguna prueba nueva de Mind (aún no hay código).

**Pendiente para fases 1–7:** lista de 13 pruebas de la misión (observación, ternario, abstención, deny policy, sin LLM, etc.).

---

## F. Riesgos

| Riesgo | Mitigación |
|--------|------------|
| Duplicar estado operativo en Mind | Solo referencias + creencias derivadas con version |
| Mezclar ternario 2 “afirmado” con “crítico seguridad” | Tipos separados: `Ternary` vs `SafetyState` |
| Mind ejecuta sin policy | Solo ActionRequest; Core ejecuta |
| Contaminar AIP | Sin endpoints hasta necesidad demostrada |
| Continuidad cognitiva distribuida prematura | Fase 8 solo tras local + tests |
| LispAI / 7 órganos scope creep | Interfaces mínimas; órganos graduales |

---

## G. Próximo paso

**Fase 1 — Contrato de Mind** (interfaces + tipos + tests de contrato, **sin** Zyrion completo ni puente Core todavía), en `runtime/mind/`, cero cambios a persistencia ni red.

Orden confirmado: **1 → 2 → 3 → 4 → 5 → 6 → 7 → 8**, detenerse si rompe baseline.

---

## Mapa de dependencias (integración)

```
                    [AIP / Studio / CLI]
                            │
                         core.Node
                    ┌───────┼───────┐
                    │       │       │
              organism   events   pulse
                    │       │
              policy/auth   │
                    │       ▼
              execution   runtime/mind  ← NUEVO (observa / propone)
                    │         │
                    │    runtime/zyrion ← NUEVO (eval ternaria)
                    │         │
                    │    inference.Provider (opcional, Fase 6)
                    ▼
                 result ──► Mind.RecordResult + provenance
```

---

*Fase 0 completa. Esperando autorización para Fase 1 (solo contrato, sin puente de ejecución todavía).*
