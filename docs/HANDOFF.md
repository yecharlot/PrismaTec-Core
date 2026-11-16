# HANDOFF — PrismaTec Core

**Última actualización:** 2026-09-26 (Multi-nodo TCP + MANIFESTO_COMPLIANCE)  
**Repositorio local:** `/home/workdir/artifacts/PrismaTec-Core`  
**Módulo Go:** `github.com/yecharlot/PrismaTec-Core`  
**Repo objetivo GitHub:** `github.com/yecharlot/PrismaTec-Core` (aún no publicado necesariamente)

Este documento existe para que **otra IA o desarrollador** pueda continuar el trabajo **sin romper** lo ya construido y **respetando el manifiesto** del proyecto.

## Política de documentación continua (obligatoria)

Cada cambio relevante debe dejar por escrito:

1. **Qué se hizo**
2. **Por qué**
3. **Pensamiento futuro** (implicaciones / riesgos)
4. **Guía de acción** (qué debe hacer la siguiente sesión)

Archivos a mantener:
- `docs/HANDOFF.md` — estado + reglas
- `docs/STATUS.md` — tenemos vs falta para el propósito
- `docs/PHASE{N}_*.md` — detalle por fase
- `docs/JOURNAL.md` — bitácora breve por sesión

Sin documentación actualizada, **no** se considera la fase cerrada.

---

## 1. Manifiesto del proyecto (no negociable)

Leer primero:

- `docs/AUDIT.md` — auditoría de AlsetOS, PrismaTec, Alset-JS-Runtime
- `docs/ARCHITECTURE.md` — arquitectura objetivo
- `docs/MIGRATION.md` — orden de fases obligatorio
- `docs/DECISIONS.md` — ADRs
- `docs/MVP_PLAN.md` — demos de éxito
- `docs/RISK_MAP.md` — qué no hacer

### Principios absolutos

1. **Core = primitives.** Applications (Mind, Gen, Studio completo, apps pharma/finance) **no** van dentro del Core obligatorio.
2. **Digital Organism** es la unidad fundamental (no solo “AI agent”).
3. **Tres identidades separadas e inviolables:**
   - `RootCID` — identidad del contenido / definición del organismo
   - `NodeID` — identidad criptográfica persistente del nodo
   - `PeerID` — identidad de transporte libp2p (aún no implementado)
4. **Separar WHAT de HOW** — interfaces + adapters/providers.
5. **Capability ≠ Policy.**
6. **No construir todo de una vez.** Orden de `MIGRATION.md`.
7. **REAL SYSTEM > DEMO FALSA > MARKETING.**
8. Ciclo por fase: `IMPLEMENT → BUILD → TEST → RUN → VERIFY → DOCUMENT`.
9. No destruir los repos originales (AlsetOS, PrismaTec, Alset-JS-Runtime).
10. Si algo puede esperar, **no** se implementa todavía.

### Criterio de éxito MVP (recordatorio)

1. Create organism → Pulse real → UI Alset-JS actualiza con datos del Core
2. Move organism Node A → Node B
3. Failure recovery con réplica

---

## 2. Estado actual del código

### Fases completadas

| Fase | Nombre | Estado | Doc |
|---|---|---|---|
| 0 | Auditoría | ✅ | `docs/AUDIT.md` + mapas |
| 1 | Core Boot | ✅ | `docs/PHASE1_BOOT.md` |
| 2 | Organism | ✅ | `docs/PHASE2_ORGANISM.md` |
| 3 | Memory + CID | ✅ | `docs/PHASE3_MEMORY_CID.md` |
| 4 | Events + Pulse | ✅ | `docs/PHASE4_PULSE.md` |
| 5 | AIP v1 | ✅ | `docs/PHASE5_AIP.md` |
| 6 | Alset-JS mínimo / Demo 1 panel | ✅ | `docs/PHASE6_ALSET_JS_DEMO.md` |
| 7 | Policy engine | ✅ | `docs/PHASE7_POLICY.md` |
| 8–12 | Execution, Inference, Network, Replication, Recovery | ✅ | `docs/PHASES_8_TO_12.md` |
| 13 | Control Studio | ✅ | `docs/PHASE13_STUDIO.md` |
| 14 | E2E demo checklist | ✅ | `docs/PHASE14_E2E.md` |
| 15 | Security hardening | ✅ | `docs/PHASE15_SECURITY.md` |
| 16 | Vertical de referencia | Parcial/documentado | `docs/PHASE16_VERTICAL.md` |

### Estructura del repo

```
PrismaTec-Core/
├── cmd/prismatec/
├── core/
│   ├── identity/
│   ├── events/
│   ├── registry/
│   └── organism/
├── storage/cid/
├── api/aip/
├── network/
├── runtime/
├── docs/
└── tests/
```

La estructura real debe verificarse contra HEAD antes de asumir que este mapa está completo.

### Verificación

No asumir que una prueba histórica representa el estado actual. Antes de modificar código:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Registrar qué comandos se ejecutaron, commit revisado y resultado. No afirmar verificación local si solo se inspeccionó CI o documentación.

---

## 3. Repos de origen

| Repo | Aporta al Core |
|---|---|
| AlsetOS | identity, organism lifecycle, p2p, wasm, rootcid idea, pulse, tests |
| PrismaTec | CID store, Mind/Gen (apps), Lisp, Cloudflare adapters |
| Alset-JS-Runtime | AIP, Registry UI, Pulse localized, Studio |

Los repos originales no se modifican como consecuencia de cambios en PrismaTEC-Core, salvo petición explícita.

---

## 4. Principios de integración de aplicaciones

La especificación normativa está en `docs/APPLICATION_INTEGRATION_CONTRACT.md`.

- La aplicación conserva sus entidades, reglas, procesos e invariantes.
- El adaptador traduce el dominio a contratos públicos del Core.
- El Core proporciona primitivas universales; no absorbe el dominio.
- El dominio no importa paquetes internos del Core.
- Debe declararse si el Core se consume embebido, remoto o híbrido.
- RootCID, NodeID, PeerID, OrganismID e ID de dominio son conceptos distintos.
- Idempotencia, atomicidad, recuperación y autorización deben probarse, no suponerse.
- Mind/Zyrion pueden razonar, pero no autorizan ni ejecutan por sí mismos.

### Primera implementación de referencia

Smart-WMS (`danielitoCode/smart-wms`, rama `feat/wms-prismatec-genesis`) es la primera implementación a contrastar con el contrato. Sus patrones de CoreStore/organismos son referencia; sus reglas de receiving, picking e inventory ledger pertenecen al WMS.

La revisión previa detectó riesgos de atomicidad en receiving, errores ignorados en picking, idempotencia dependiente de claves estables y una frontera Core documentada de forma distinta a su composición embebida. Esos hallazgos deben revalidarse contra el HEAD actual antes de considerarlos resueltos o vigentes.

---

## 5. Qué NO debe hacerse

- No convertir PrismaTEC-Core en un monolito de aplicaciones verticales.
- No trasladar inventario, recepción, picking, pacientes, cuentas o pedidos al núcleo.
- No crear SDKs que expongan structs o paquetes internos inestables.
- No prometer transacciones entre organismos, consenso, disponibilidad o RPO/RTO sin implementación y pruebas.
- No equiparar RootCID con identidad de usuario, autoridad o versión de negocio.
- No confundir réplica disponible con réplica autorizada para convertirse en primaria.
- No asumir que AIP v1 ya cubre todas las garantías del contrato universal.

---

## 6. Guía de continuidad

1. Leer este HANDOFF, `docs/DECISIONS.md`, `docs/STATUS.md` y `docs/APPLICATION_INTEGRATION_CONTRACT.md`.
2. Verificar HEAD y ejecutar baseline de pruebas.
3. Auditar Smart-WMS vigente y construir matriz requisito → código → prueba → estado.
4. Resolver primero discrepancias entre contrato, implementación y documentación.
5. Diseñar interfaces públicas mínimas solo después de comprobar qué abstracciones son realmente universales.
6. Validar el contrato con una segunda aplicación de dominio distinto.
7. Actualizar HANDOFF, STATUS, JOURNAL y documentos de arquitectura en cada cambio relevante.

---

## 7. Decisión y estado de 2026-11-16

Se creó el contrato universal de integración v0.1.0 en `docs/APPLICATION_INTEGRATION_CONTRACT.md` y se enlazó desde README y ARCHITECTURE.

**Estado:** especificación arquitectónica propuesta; no equivale a SDK, API ni certificación de conformidad. No se ejecutaron pruebas Go ni se modificó lógica de runtime en esta actualización.

**Siguiente paso:** auditar la rama actual de Smart-WMS y completar una matriz de conformidad con evidencia verificable. No comenzar implementando interfaces adicionales antes de esa auditoría.

**Fin del handoff.**
