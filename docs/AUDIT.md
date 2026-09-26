# AUDIT — PrismaTec Core (Fase 0)

**Fecha:** 2026-09-26  
**Objetivo:** Unificación arquitectónica de AlsetOS + PrismaTec + Alset-JS-Runtime  
**Regla:** No destruir los repositorios originales. PrismaTec Core es evolución, no operación destructiva.

---

## 1. Resumen ejecutivo

| Repositorio | Lenguaje principal | Tamaño aprox. | Madurez | Rol conceptual |
|-------------|--------------------|---------------|---------|----------------|
| **AlsetOS** | Go | ~2.8k LOC + tests | Núcleo experimental sólido | Distributed Runtime / Execution Fabric + Digital Organism primitives |
| **PrismaTec** (PrismaTec) | Go | Más grande (HTTP API, Mind, Gen, CF, apps) | Más orientado a producto/demo | Intelligence + Agents + Memory + Coordination + Edge adapters |
| **Alset-JS-Runtime** | JavaScript (ES modules) | Runtime + AIP + Studio + examples | Maduro en reactividad local | Human/Machine Experience (AIP + Pulse + Registry + localized UI) |

**Conclusión clave:**  
Los tres proyectos convergen en el mismo concepto oculto: **Digital Organism / Persistent Autonomous Digital Entity**.  
AlsetOS aporta el *cómo se ejecuta y se mueve*, PrismaTec el *cómo decide y recuerda*, Alset-JS el *cómo se observa e interactúa*.  
PrismaTec Core debe extraer las *primitives* y dejar las *applications* fuera.

---

## 2. Qué existe (inventario de alto nivel)

### 2.1 AlsetOS (`github.com/yecharlot/AlsetOS`)

**Estructura principal:**
```
agente/          autonomia/       boot/           capacidad/
cmd/             eventos/         gene/           identidad/
kernel/          lispai/          manifiesto/     memoria/
mind/            motor/           nodo/           organismo/
organismos/      p2p/             planificador/   pruebas/
pulso/           recuperacion/    recursos/       red/
registro/        rootcid/         sandbox/        wasm/
zyrion/
```

**Conceptos implementados:**
- Organismo (estado: creado/listo/ejecutando/detenido/error)
- RootCID (actualmente SHA256-based `rootcid:...`, no CID IPFS completo aún)
- Identidad Ed25519 → NodeID (persistente)
- PeerID (derivado de libp2p)
- Capacidades (mapa bool + `Requerir`)
- Memoria (map[string]string + persistencia JSON simple)
- Pulse (tipo + origen + contenido)
- Eventos
- LispAI (evaluador)
- Zyrion (lógica ternaria 0/1/2)
- Mind
- Gene / Agente
- WASM (wazero)
- Sandbox
- P2P: libp2p + mDNS + Kademlia DHT + streams + heartbeat + persistencia de organismos
- Recuperación / Placement
- Manifiesto
- Registro
- Recursos / Planificador / Autonomía

**Dependencias clave:** libp2p, go-cid, go-datastore, wazero, multiformats.

**Fortalezas:** Separación clara de identidades (RootCID / NodeID / PeerID), lifecycle de organismo, tests en `pruebas/`, contrato explícito (`Contrato.md`).

**Debilidades / incompleto:**
- Organismo aún muy simple (memoria solo map[string]string)
- RootCID no es CID real de contenido canónico completo
- Pulse y Events aún locales/en proceso
- Poca separación interface vs implementación de red
- No hay Policy engine real (solo capacidad bool)
- No hay AIP ni capa de experiencia humana
- CLI mínima

### 2.2 PrismaTec (`github.com/yecharlot/PrismaTec`)

**Estructura principal (desde README + partial):**
```
cmd/prisma-tec/          → binario del nodo
internal/
  node/                  → corazón HTTP + Mind + Gen + red + persistencia
  lisp/                  → LispAI + evaluar-zyrion
  agents/                → Gen / registry
  gennode/               → daemon, explore, dialogue, cargo, udp_pulse
  persistence/           → Store (local | Supabase | Cloudflare DO)
  pulse/  neural/  sync/ poh/ vero/ ...
cloudflare/alset-gen-worker/
docs/ (HANDOFF, GUIA, tesis Mind, etc.)
```

**Conceptos implementados:**
- Nodo HTTP + libp2p
- Alset Mind (7 órganos: dialog, act, mem, self, ethics, curiosity, humor)
- Alset Gen (células/sondas: create, explore, dispatch, memory, return…)
- LispAI + Zyrion
- Memoria por CID + sesiones (X-Mind-Session)
- Persistencia: local / Supabase / Cloudflare Durable Objects
- Apps por CID (`*.app.ans`)
- SSE / Pulse
- Cloudflare Workers (edge)
- API v2
- Admin panel embebido
- Genes con lifecycle, autonomy, publish

**Fortalezas:** Más “producto” (API, sesiones, edge, demos reales), integración Cloudflare, corpus + diálogo, Gen como unidad móvil.

**Debilidades / incompleto:**
- Mezcla fuerte de Core + Applications (finanzas_app, abacophy_app, etc.)
- Mind/Gen son muy específicos de “IA conversacional”
- Menos énfasis en el modelo de Organismo genérico y movimiento/recovery formal
- RootCID / NodeID / PeerID presentes pero menos rigurosamente separados que en AlsetOS
- No hay Policy engine formal (WHO CAN DO WHAT…)
- UI aún tradicional (no Alset-JS)

### 2.3 Alset-JS-Runtime (`github.com/yecharlot/Alset-JS-Runtime`)

**Estructura principal:**
```
src/core/AlsetPulseCore.js
src/components/
docs/AIP-Specification.md
docs/AIP-Extended-Specification.md
docs/AIP-Security-Profile.md
examples/ (basic, forms, telemetry, studio, components)
public/
```

**Conceptos implementados:**
- AlsetRegistry (nodos con identidad persistente)
- Pulse (`{ target, type, data, meta }`)
- Localized Resonance (actualización solo del nodo afectado)
- Persistent Identity Node (PIN)
- Interactive Address Space (IAS)
- Live Mutability
- AIP v1 (especificación formal)
- AlsetStudio (IDE visual)
- Componentes (Column, Row, Text, Button, Card, Input, Theme…)
- Telemetry, Inspector
- Estado reactivo fino (`alsetState`)

**Fortalezas:** Filosofía de UI muy alineada con organismos (identidad estable + estímulos locales). AIP es el contrato natural.

**Debilidades / incompleto:**
- Aún es un runtime frontend; no habla todavía con un Core real de organismos
- Studio es demo visual, no control plane completo
- No hay integración nativa con libp2p / CID / WASM del lado Go

---

## 3. Qué funciona

| Componente | AlsetOS | PrismaTec | Alset-JS | Notas |
|------------|---------|-----------|----------|-------|
| Identidad Ed25519 / NodeID | Sí | Parcial | N/A | AlsetOS más limpio |
| RootCID | Sí (simplificado) | Sí (CID real en persistencia) | N/A | Unificar hacia CID canónico |
| PeerID / libp2p | Sí | Sí | N/A | Reutilizar AlsetOS + PrismaTec |
| Organismo lifecycle | Sí (básico) | Implícito vía Gen/Mind | N/A | Formalizar |
| Capacidades | Sí (bool) | Parcial | N/A | Evolucionar a Capability + Policy |
| Memoria | Sí (JSON) | Sí (CID + sesiones) | N/A | Preferir modelo CID de PrismaTec |
| LispAI / Zyrion | Sí | Sí | N/A | Extraer como módulo de ejecución |
| Mind / Gen | Básico | Avanzado | N/A | Application layer, no Core obligatorio |
| Pulse | Local | SSE + UDP | Core del runtime | Unificar como bus interno + AIP |
| WASM sandbox | Sí (wazero) | — | N/A | Conservar |
| Recovery / Replica | Esqueleto | Parcial | N/A | Prioridad alta para demo |
| AIP / Registry / Localized UI | — | — | Sí | Conservar e integrar |
| Cloudflare / Edge | — | Sí | N/A | Adapter opcional |

---

## 4. Qué está incompleto / deuda

1. **Modelo de Organismo unificado** — AlsetOS tiene struct simple; PrismaTec tiene Gen/Mind; falta un modelo canónico con Identity + State + Memory + Capabilities + Policy + Provenance + Placement + Status.
2. **Policy Engine** — Solo capacidades booleanas. Falta WHO / CAN DO WHAT / TO WHICH / UNDER CONDITIONS.
3. **RootCID real** — Debe ser CID de contenido canónico (manifiesto + estado inicial), no solo hash de nombre.
4. **Separación Interface / Adapter** — Red, Storage, Inference, Execution todavía están acoplados.
5. **AIP ↔ Core** — Alset-JS no está conectado a un nodo real.
6. **Replication + Failure recovery** — Existe esqueleto en AlsetOS; no hay demo end-to-end verificable.
7. **Observability** — Logs, timeline, provenance, health aún fragmentados.
8. **CLI mínima** — `prismatec node start / organism create / inspect`.
9. **Tests de integración cross-repo** — Cada repo tiene tests propios, no hay suite unificada.

---

## 5. Qué debe conservarse (KEEP)

### De AlsetOS
- Separación RootCID / NodeID / PeerID
- Package `organismo`, `identidad`, `capacidad`, `pulso`, `rootcid`, `p2p`, `wasm`, `sandbox`, `recuperacion`
- Contrato conceptual (`Contrato.md`)
- Uso de wazero + libp2p + go-cid
- Tests unitarios de identidades y lifecycle

### De PrismaTec
- Persistencia por CID + bloques
- Sesiones de memoria
- LispAI + evaluar-zyrion (motor de ejecución)
- Cloudflare Durable Objects / Workers como *adapter*
- Concepto de Gen como organismo móvil (pero generalizado)
- API HTTP + SSE como transporte de Pulse/AIP

### De Alset-JS-Runtime
- AlsetRegistry + Pulse + Localized Resonance
- AIP Specification v1
- Componentes reactivos y Studio (como capa de experiencia)
- Filosofía “no re-render global”

---

## 6. Qué debe trasladarse (MOVE)

| Origen | Componente | Destino en Core | Acción |
|--------|------------|-----------------|--------|
| AlsetOS | organismo, identidad, rootcid, capacidad, pulso, eventos, p2p, wasm, recuperacion | `core/` + `network/` + `runtime/` | MOVE + normalizar interfaces |
| AlsetOS | lispai, zyrion, mind, gene, agente | `runtime/` (opcional) | MOVE como módulos de ejecución |
| PrismaTec | persistence (CID/blocks), lisp engine | `storage/` + `runtime/lisp/` | MOVE |
| PrismaTec | Cloudflare worker | `adapters/cloudflare/` | MOVE como adapter |
| PrismaTec | Mind / Gen / apps | Applications / ejemplos | NO al Core (solo interfaces) |
| Alset-JS | Registry, Pulse, AIP, core components | `sdk/js/` + `api/aip/` | MOVE mínimo para demo |

---

## 7. Qué debe reescribirse (REWRITE)

- **Organism model** → struct canónico con campos estables (ID, RootCID, Identity, State, MemoryRef, Capabilities, Policy, RuntimeSpec, Provenance, Placement, Status, timestamps).
- **Capability + Policy** → separar Capability (qué puede) de Policy (quién/cuándo/condiciones).
- **Network interface** → `Transport / Discovery / Messaging / Replication / Presence` → adapter libp2p.
- **Pulse bus interno** → unificar el Pulse de los tres repos en un EventBus + Pulse emitter que alimente AIP.
- **RootCID** → generación canónica desde manifiesto serializado (CID real).
- **AIP v1** → contrato versionado mínimo (identity, organism, pulse, commands, events, errors).

---

## 8. Qué está duplicado

| Concepto | AlsetOS | PrismaTec | Alset-JS | Decisión |
|----------|---------|-----------|----------|----------|
| Pulse | Sí (simple) | Sí (SSE/UDP) | Sí (core) | Unificar en Core EventBus + AIP |
| LispAI / Zyrion | Sí | Sí | — | Una sola implementación (preferir la más completa) |
| Memoria | map JSON | CID + sesiones | — | Modelo CID + working/episodic/semantic |
| Identidad nodo | Ed25519 | Presente | — | AlsetOS como base |
| Registry | — | agents/registry | AlsetRegistry | Distintos (node vs UI) — ambos necesarios |
| Mind / Gen | Básico | Avanzado | — | Application, no Core |

---

## 9. Qué pertenece realmente al Core

```
IDENTITY (RootCID, NodeID, PeerID)
ORGANISM (modelo + lifecycle)
STATE
MEMORY (working / episodic / semantic + CID)
CAPABILITY
POLICY
EXECUTION (WASM, Lisp, Agent, External Inference — como providers)
NETWORK (interfaces + libp2p adapter)
PROVENANCE
LIFECYCLE
EVENTS / EVENT BUS
PULSE
REGISTRY (nodos + organismos)
STORAGE (CID / blocks / local / distributed)
AIP (contrato)
```

---

## 10. Qué pertenece a Applications / módulos opcionales

- Alset Mind (órganos específicos)
- Alset Gen (células exploradoras)
- Apps concretas (finanzas, pharma, etc.)
- Studio completo (puede vivir como `prismatec-studio`)
- Cloudflare Durable Objects (adapter)
- Supabase (adapter)
- Inference providers concretos (OpenAI, Ollama…)
- UI frameworks adicionales

---

## 11. Tabla de migración (Source → Destination → Action)

| Source Repo | Component | Destination | Action |
|-------------|-----------|-------------|--------|
| AlsetOS | identidad | core/identity | KEEP + minor refactor |
| AlsetOS | rootcid | core/identity + storage/cid | REFACTOR (CID real) |
| AlsetOS | organismo | core/organism | REWRITE (modelo canónico) |
| AlsetOS | capacidad | core/capability | KEEP + extend |
| AlsetOS | memoria | core/memory | REFACTOR (CID + tipos) |
| AlsetOS | pulso / eventos | core/events + core/pulse | REFACTOR + unificar |
| AlsetOS | p2p / red | network/ + adapters | MOVE + interface |
| AlsetOS | wasm / sandbox | runtime/wasm | KEEP |
| AlsetOS | recuperacion | core/lifecycle + network | KEEP + demo |
| AlsetOS | lispai / zyrion / mind / gene / agente | runtime/ | OPTIONAL MODULE |
| PrismaTec | persistence / CID store | storage/ | MOVE |
| PrismaTec | lisp engine | runtime/lisp | MOVE (elegir la mejor) |
| PrismaTec | Cloudflare | adapters/cloudflare | OPTIONAL |
| PrismaTec | Mind / Gen / apps | examples/ o apps externas | DROP from Core |
| Alset-JS | AlsetRegistry + Pulse + AIP | api/aip + sdk/js | MOVE mínimo |
| Alset-JS | Studio | demos/studio o repo separado | OPTIONAL later |
| Alset-JS | components | sdk/js | KEEP for demo |

---

## 12. Riesgos principales

| Riesgo | Probabilidad | Impacto | Mitigación |
|--------|--------------|---------|------------|
| Intentar portar todo de una vez | Alta | Alto | Seguir fases estrictas; MVP mínimo |
| Mezclar Core con Applications (Mind/Gen) | Alta | Alto | Regla: Core = primitives only |
| RootCID inconsistente entre repos | Media | Alto | Definir formato canónico en Fase 1 |
| Acoplamiento libp2p → Organism | Media | Medio | Interface Network primero |
| Demo falsa (UI sin datos reales) | Media | Alto | Criterio de éxito: datos vienen del Core |
| Pérdida de tests existentes | Media | Medio | Portar tests junto con el código |
| Complejidad de Policy demasiado pronto | Media | Medio | API diseñada para evolucionar; implementación mínima al inicio |

---

## 13. MVP Plan (lo que debe funcionar primero)

**Demo 1 (obligatorio):**
```
Browser → Alset-JS (mínimo) → AIP → PrismaTec Core Node
  → Create Organism
  → Assign Identity + Capabilities + Memory
  → Start
  → Organism emits Pulse
  → Browser recibe Pulse
  → UI se actualiza localmente (datos reales)
```

**Demo 2:**
```
Organism en Node A → Move → Node B
(RootCID, estado, memoria, provenance se conservan)
```

**Demo 3:**
```
Primary = Node A, Replica = Node B
Apagar Node A → detect → activate replica → continue
```

Todo lo demás (Mind completo, Gen, Studio full, multi-cloud, IAM enterprise) espera.

---

## 14. Próximos documentos generados

- `docs/ARCHITECTURE.md` — mapa de arquitectura objetivo
- `docs/MIGRATION.md` — plan de fases detallado
- `docs/DECISIONS.md` — decisiones arquitectónicas (ADRs)
- `docs/DEPENDENCY_MAP.md` — dependencias y providers
- `docs/RISK_MAP.md` — riesgos y mitigaciones
- `docs/MVP_PLAN.md` — plan concreto de MVP

---

**Estado de esta auditoría:** Completa a nivel de repositorios públicos disponibles.  
Código fuente de AlsetOS y Alset-JS-Runtime inspeccionado en profundidad.  
PrismaTec inspeccionado vía README, estructura y partial checkout (clone completo ralentizado por tamaño).  

**Siguiente paso obligatorio:** ARCHITECTURE MAP + MIGRATION MAP antes de cualquier implementación.
