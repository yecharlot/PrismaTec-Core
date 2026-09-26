# DECISIONS (ADRs) — PrismaTec Core

## ADR-001: Unidad fundamental = Digital Organism

**Estado:** Aceptado  
**Contexto:** Los tres repos hablan de organismos, agentes, genes, nodos, PINs…  
**Decisión:** La unidad canónica es **Digital Organism** (entidad digital persistente, identificable, ejecutable, observable, distribuible).  
No se limita a “AI agent”. Puede ser sensor, robot, workflow, digital twin, dataset, proceso de negocio, etc.  
**Consecuencia:** Core no depende del concepto “AI agent”. Intelligence es un provider.

---

## ADR-002: Tres identidades separadas e inviolables

**Estado:** Aceptado  
**Decisión:**
- `RootCID` — identidad del organismo / definición de contenido
- `NodeID` — identidad persistente del nodo ejecutor
- `PeerID` — identidad de transporte (libp2p)

Un organismo puede moverse entre nodos manteniendo RootCID.  
**Consecuencia:** Nunca mezclar las tres en un solo campo.

---

## ADR-003: Core = primitives; Applications = consumers

**Estado:** Aceptado  
**Decisión:** Mind, Gen, apps de finanzas/pharma, Studio completo, etc. **no** forman parte del Core obligatorio.  
Entran como módulos opcionales o applications externas.  
**Consecuencia:** Evita el monolito y permite que el Core sobreviva a cambios de dominio.

---

## ADR-004: Separar WHAT de HOW (providers / adapters)

**Estado:** Aceptado  
**Decisión:** Todo almacenamiento, red, inferencia, ejecución, UI transport entra por interfaces.  
libp2p, Cloudflare, OpenAI, wazero, etc. son implementaciones.  
**Consecuencia:** El Core no hardcodea proveedores.

---

## ADR-005: Capability ≠ Policy

**Estado:** Aceptado  
**Decisión:**  
- Capability = “puedo hacer X”  
- Policy = “quién puede, sobre qué, bajo qué condiciones”  
La autorización se evalúa en runtime.  
**Consecuencia:** Prepara el sistema para escenarios enterprise reales sin implementar IAM completo ahora.

---

## ADR-006: Pulse + AIP como contrato de experiencia

**Estado:** Aceptado  
**Decisión:** Pulse es el mecanismo de comunicación interno y hacia fuera.  
AIP v1 es el contrato versionado entre Core y clientes (Alset-JS u otros).  
Actualizaciones localizadas (no re-render global).  
**Consecuencia:** La UI de demo debe recibir datos reales del Core.

---

## ADR-007: RootCID debe ser CID de contenido canónico

**Estado:** Aceptado  
**Decisión:** Abandonar gradualmente el `rootcid:sha256...` simple de AlsetOS en favor de CID real generado desde manifiesto serializado canónico.  
**Consecuencia:** Compatibilidad temporal posible; objetivo es CID verificable.

---

## ADR-008: No construir todo de una vez

**Estado:** Aceptado  
**Decisión:** Seguir el orden de fases de MIGRATION.md.  
MVP = Demo 1 (create → pulse → UI real) + Demo 2 (move) + Demo 3 (recovery).  
Todo lo demás espera.  
**Consecuencia:** Sistema real > demo falsa > marketing.

---

## ADR-009: Seguridad desde el diseño (mínimo viable)

**Estado:** Aceptado  
**Decisión:** Desde la primera versión: identity, capability isolation, policy, provenance, least privilege, input validation, resource limits, sandbox para módulos no confiables.  
No se implementa un IAM enterprise completo todavía.  
**Consecuencia:** La API de Policy se diseña para crecer.

---

## ADR-010: Observability como parte del organismo

**Estado:** Aceptado  
**Decisión:** Todo organismo debe poder responder: Who am I? Where am I? What am I doing? What did I do? What am I allowed to do? What happened? Who authorized it?  
Esto se refleja en events, provenance, pulse y UI.
