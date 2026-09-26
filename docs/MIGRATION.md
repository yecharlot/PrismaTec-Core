# MIGRATION MAP — PrismaTec Core

**Regla de oro:**  
```
IMPLEMENT → BUILD → TEST → RUN → VERIFY → DOCUMENT
```
Si falla: STOP → DIAGNOSE → FIX → TEST AGAIN.  
No acumular capas rotas.

---

## Orden de ejecución obligatorio

| Step | Nombre | Objetivo | Criterio de salida |
|------|--------|----------|--------------------|
| 0 | AUDIT | Documentos de auditoría | AUDIT.md + ARCHITECTURE.md + este archivo |
| 1 | Core Boot | Node + Identity + Registry + EventBus + Storage mínimo | `prismatec node start` arranca |
| 2 | Organism | create / start / stop / inspect / update / persist / restore | CLI mínima funciona |
| 3 | Memory + CID | Working + Episodic + Semantic + CID | Organismo guarda y restaura memoria por CID |
| 4 | Events + Pulse | EventBus interno + emisión de Pulse | Eventos se registran y emiten |
| 5 | AIP v1 | Contrato versionado mínimo | Browser puede hablar con Core vía AIP |
| 6 | Alset-JS mínimo | Registry + Pulse + UI localizada | Demo 1 (create → pulse → UI update real) |
| 7 | Capabilities + Policy | Separación Capability / Policy + autorización básica | Policy deniega acciones no permitidas |
| 8 | Execution | WASM + Lisp (módulos) | Organismo puede ejecutar módulo sandboxed |
| 9 | Intelligence providers | Interfaces InferenceProvider | No hardcodear OpenAI/etc. |
| 10 | Network | Interfaces + libp2p adapter | Dos nodos se descubren |
| 11 | Replication | Primary + Replica | Organismo replicado |
| 12 | Failure / Recovery | Detect → locate → verify → activate | Demo 3 funciona |
| 13 | Studio (progresivo) | Control plane visual | Observación de nodos/organismos |
| 14 | End-to-end demo | 12 pasos del checklist | Demo ejecutable y real |
| 15 | Security hardening | Sandbox, least privilege, audit, signed events | Tests de seguridad pasan |
| 16 | Documentation + Investor package | docs/investor/ | Solo después de demos reales |

---

## Fase 0 — Auditoría (COMPLETADA)

- [x] Clonar / inspeccionar AlsetOS
- [x] Clonar / inspeccionar Alset-JS-Runtime
- [x] Inspeccionar PrismaTec (README + estructura + partial)
- [x] Generar AUDIT.md
- [x] Generar ARCHITECTURE.md
- [x] Generar MIGRATION.md (este documento)
- [ ] Generar DECISIONS.md, DEPENDENCY_MAP.md, RISK_MAP.md, MVP_PLAN.md

---

## Fase 1 — Core Boot

**Qué existe:** Identidad y Node en AlsetOS.  
**Qué reutilizamos:** `identidad`, estructura de `nodo`/`p2p`.  
**Qué cambiamos:** Interfaces limpias, sin acoplar libp2p todavía.  
**Nueva interfaz:**
```go
type Node interface {
    ID() NodeID
    Start(ctx context.Context) error
    Stop() error
    Registry() Registry
}
```
**Cómo testear:** `go test ./core/...` + `prismatec node start` imprime NodeID y escucha.

---

## Fase 2 — Organism

**Qué existe:** `organismo.Organismo` simple en AlsetOS.  
**Qué reutilizamos:** Lifecycle states + tests.  
**Qué cambiamos:** Modelo canónico (ver ARCHITECTURE.md).  
**CLI mínima:**
```
prismatec organism create --name research-agent-01
prismatec organism list
prismatec organism inspect <id>
prismatec organism start <id>
prismatec organism stop <id>
```

---

## Fase 3 — Memory + CID

Preferir el modelo de CID/bloques de PrismaTec + tipado (working/episodic/semantic).  
RootCID debe ser CID real de manifiesto canónico.

---

## Fase 4–6 — Pulse + AIP + Alset-JS

Unificar Pulse de los tres repos.  
AIP v1 debe ser mínimo y versionado.  
Alset-JS solo se porta lo necesario para el primer demo (no todo el Studio).

---

## Fase 7–9 — Capabilities / Policy / Execution / Intelligence

Policy API diseñada para crecer (WHO/CAN/WHAT/CONDITIONS).  
Implementación inicial simple.  
WASM (wazero) y Lisp como providers.

---

## Fase 10–12 — Network / Replication / Recovery

Interfaces primero → libp2p adapter.  
Demo de movimiento de organismo.  
Demo de fallo de nodo + recuperación desde réplica.

---

## Reglas de migración de código

1. Identificar origen
2. Adaptar a interfaces
3. Testear
4. Documentar
5. Integrar
6. Eliminar duplicación

**Nunca:** `copy everything`.  
**Siempre:** `extract → normalize → integrate → test`.

---

## Criterio de éxito de la primera versión demostrable

```
USER → PRISMATEC STUDIO (mínimo) → AIP → CORE NODE
         → ORGANISM (state + memory + capabilities)
         → EVENTS / PULSE / AUDIT
         → CID / STORAGE
```

Y después:
```
NODE A ──organism──▶ NODE B   (con recovery real)
```
