# STATUS — PrismaTEC-Core

**Actualizado:** 2026-11-16  
**Referencia normativa de integración:** `docs/APPLICATION_INTEGRATION_CONTRACT.md` v0.1.0  
**Propósito:** infraestructura para organismos digitales persistentes, identificables, observables y distribuibles.

> Este documento resume el estado descrito por la documentación del repositorio. No sustituye la ejecución de pruebas ni una auditoría de código contra HEAD.

## 1. Dirección del proyecto

PrismaTEC-Core proporciona primitivas comunes. Las aplicaciones conservan sus dominios, modelos, reglas e invariantes. La integración se realiza mediante adaptadores y contratos públicos.

No es responsabilidad del Core implementar inventarios, recepción, picking, expedientes clínicos, contabilidad, ventas ni otros dominios verticales.

## 2. Capacidades que el repositorio documenta

| Área | Estado declarado en la documentación | Límite que debe conservarse |
|---|---|---|
| Node e identidad | Implementado | NodeID no es RootCID ni PeerID |
| Organismos y ciclo de vida | Implementado | No implica autonomía general |
| RootCID y memoria CID | Implementado | RootCID no es versión de negocio ni autorización |
| Events y Pulse | Implementado | Un evento no sustituye una fuente durable de verdad |
| AIP HTTP/SSE | Implementado | AIP v1 no equivale todavía al contrato universal de dominio |
| Policy y auditoría | Implementado parcialmente | Revisar cobertura de todas las operaciones sensibles |
| Execution / WASM | Implementado con alcance limitado | No implica sandbox industrial completo |
| Inference / Mind | Interfaces y capacidades documentadas | No afirmar razonamiento general validado |
| TCP / libp2p | Transporte documentado | La cobertura E2E y la seguridad varían por transporte |
| Replicación / recuperación | Implementado para escenarios documentados | No implica consenso ni tolerancia a toda partición |
| Contrato de integración | Especificación arquitectónica v0.1.0 | No es todavía SDK ni API completa |

Para detalles de evidencia y límites consultar `docs/STATE_OF_THE_UNION.md`, `docs/EVIDENCE_FOR_REVIEWERS.md`, `docs/SECURITY.md` y las pruebas correspondientes.

## 3. Contrato universal de integración

Creado: `docs/APPLICATION_INTEGRATION_CONTRACT.md`.

Define:
- responsabilidades de Aplicación, Adaptador, Core y proveedores;
- declaración de tipos de organismos y relaciones;
- separación de ID de dominio, OrganismID, RootCID, NodeID y PeerID;
- estado, episodios, contenido y tombstones;
- operation_id, idempotencia, concurrencia y atomicidad;
- autorización y separación Capability/Policy;
- replicación, disponibilidad y recuperación;
- Mind/Zyrion como razonamiento sin autoridad de ejecución;
- versionado, errores y pruebas de conformidad.

**Estado:** documentado como propuesta normativa. Ninguna garantía se considera implementada por el mero hecho de aparecer en el contrato.

## 4. Aplicación de referencia

Smart-WMS, rama `feat/wms-prismatec-genesis`, es el primer caso de estudio.

La aplicación mantiene warehouse, inventory, receiving y picking como conceptos de dominio. El adaptador traduce registros y operaciones a organismos, episodios y referencias CID.

Antes de declarar conformidad:
1. revisar el HEAD vigente de Smart-WMS;
2. mapear cada requisito del contrato a código y pruebas;
3. corregir los problemas de atomicidad, idempotencia y manejo de errores que sigan presentes;
4. verificar la frontera real entre Core embebido o remoto.

## 5. Brechas de producto y arquitectura

- [ ] Matriz de conformidad Smart-WMS: requisito → código → prueba → evidencia.
- [ ] Interfaces públicas mínimas del adaptador, sin exponer paquetes internos.
- [ ] Contrato de operación estable e idempotente.
- [ ] Pruebas de conformidad reutilizables.
- [ ] Validación del contrato con una segunda aplicación de dominio distinto.
- [ ] Alinear README, HANDOFF, STATUS y evidencia técnica con resultados reales de CI.

## 6. Reglas de continuidad

1. Core ≠ aplicación vertical.
2. La aplicación es dueña de sus reglas e invariantes.
3. No importar paquetes internos del Core desde el dominio.
4. Declarar integración embebida, remota o híbrida.
5. No confundir replicación con consenso ni razonamiento con autorización.
6. No prometer atomicidad, RPO/RTO o recuperación sin pruebas.
7. Cada cambio relevante actualiza HANDOFF, STATUS, JOURNAL y documentos afectados.
8. Ciclo de trabajo: IMPLEMENT → BUILD → TEST → RUN → VERIFY → DOCUMENT.

## 7. Próxima actividad

Auditar Smart-WMS vigente y construir la matriz de conformidad antes de diseñar o implementar más interfaces públicas. El objetivo es extraer únicamente abstracciones realmente universales, sin copiar la lógica del WMS al Core.
