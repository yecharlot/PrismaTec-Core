# HANDOFF — PrismaTec Core

**Última actualización:** 2026-09-26 (Fase 5 AIP v1 completada)  
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
10. Si algo puede esperar, **no se implementa todavía**.

### Criterio de éxito MVP (recordatorio)

1. Create organism → Pulse real → UI Alset-JS actualiza con datos del Core  
2. Move organism Node A → Node B  
3. Failure recovery con réplica  

---

## 2. Estado actual del código

### Fases completadas

| Fase | Nombre | Estado | Doc |
|------|--------|--------|-----|
| 0 | Auditoría | ✅ | `docs/AUDIT.md` + mapas |
| 1 | Core Boot | ✅ | `docs/PHASE1_BOOT.md` |
| 2 | Organism | ✅ | `docs/PHASE2_ORGANISM.md` |
| 3 | Memory + CID | ✅ | `docs/PHASE3_MEMORY_CID.md` |
| 4 | Events + Pulse | ✅ | `docs/PHASE4_PULSE.md` |
| 5 | AIP v1 | ✅ | `docs/PHASE5_AIP.md` |
| 6 | Alset-JS mínimo | ⏳ siguiente | Demo 1 |

### Estructura del repo

```
PrismaTec-Core/
├── cmd/prismatec/main.go          # CLI: node + organism
├── core/
│   ├── node.go                    # Node bootstrap + Organisms()
│   ├── node_test.go
│   ├── identity/                  # NodeID Ed25519 (LoadOrCreate, Sign, Verify)
│   ├── events/                    # EventBus in-process
│   ├── registry/                  # OrganismRecord registry
│   └── organism/                  # Modelo + Manager + lifecycle + memory + persist JSON
├── storage/cid/                   # Content-addressed blocks (cid1: + LocalStore)
├── docs/                          # AUDIT, ARCHITECTURE, MIGRATION, ADRs, HANDOFF, PHASEx
├── go.mod                         # module github.com/yecharlot/PrismaTec-Core  (Go 1.25.0)
├── core/pulse/                    # Pulse hub (Phase 4)
├── api/aip/                       # AIP v1 HTTP+SSE (Phase 5)
├── README.md
└── (dirs preparados: network, runtime, api, adapters, sdk, …)
```

### Qué funciona hoy (verificado)

```bash
cd /home/workdir/artifacts/PrismaTec-Core
go test ./...
# ok: core, identity, events, registry, organism

export PRISMATEC_DATA_DIR=/tmp/prismatec-phase2   # o cualquier dir

go run ./cmd/prismatec version
go run ./cmd/prismatec node info
go run ./cmd/prismatec organism create research-agent-01 \
  --cap memory.read --cap inference --deny database.write
go run ./cmd/prismatec organism list
go run ./cmd/prismatec organism start research-agent-01
go run ./cmd/prismatec organism inspect research-agent-01
go run ./cmd/prismatec organism stop research-agent-01
```

- Identidad de nodo **persistente** en `{DataDir}/identity.json`
- Organismos **persistentes** en `{DataDir}/organisms/{id}.json`
- RootCID **determinista** desde manifiesto (SHA-256, prefijo `rootcid:`)
- Eventos emitidos en create/start/stop
- UI de terminal tipo “panel” con datos **reales** del Core

### Modelo Organism actual (resumen)

```go
// core/organism/organism.go
type Organism struct {
  ID, RootCID, Version, Name string
  Status Status  // created|ready|running|stopped|error|recovering
  Capabilities []Capability
  Policy Policy           // Rules map[string]bool  (mínimo; Phase 7 crece)
  Memory MemoryRef        // Working map + Episodes int  (Phase 3 amplía)
  Runtime RuntimeSpec
  Provenance Provenance
  Placement Placement     // Primary = NodeID
  CurrentAction string
  CreatedAt, UpdatedAt time.Time
}
```

`Manager` (`core/organism/manager.go`): Create, Start, Stop, Get, List, SetAction, PutMemory, persistencia, reload, sync registry, emit events.

### CLI actual

```
prismatec node start|info
prismatec organism create|list|inspect|start|stop
prismatec version|help
```

Env: `PRISMATEC_DATA_DIR`, `PRISMATEC_NODE_NAME`

---

## 3. Repos de origen (referencia, no modificar como fuente de verdad del Core)

| Repo | Path local (si clonado) | Aporta al Core |
|------|-------------------------|----------------|
| AlsetOS | `/home/workdir/artifacts/AlsetOS` | identity, organism lifecycle, p2p, wasm, rootcid idea, pulse, tests |
| PrismaTec | `/home/workdir/artifacts/Prismatec-repo` (partial) | CID store, Mind/Gen (apps), Lisp, Cloudflare adapters |
| Alset-JS-Runtime | `/home/workdir/artifacts/Alset-JS-Runtime` | AIP, Registry UI, Pulse localized, Studio |

**Nota:** El repo GitHub se llama `PrismaTec` (T mayúscula), no `Prismatec`.

---

## 4. Qué NO está hecho todavía (no inventar como si existiera)

- [x] Memory tipada (working / episodic / semantic) + CID de artefactos (`cid1:` + sha256 blocks)
- [ ] CID multihash IPFS real (ipfs/go-cid) — pendiente cuando red de módulos funcione; RootCID sigue `rootcid:`
- [x] EventBus → Pulse (in-process hub + bridge); transporte red/UI en Fase 5+
- [ ] AIP v1 + conexión Alset-JS
- [ ] Policy engine real (WHO/CAN/WHAT/CONDITIONS)
- [ ] WASM / Lisp execution
- [ ] libp2p / DHT / Gossip
- [ ] Replication + recovery demos
- [ ] Studio / HTTP API / SSE
- [ ] Publicación a GitHub del repo Core

---

## 5. Fase 3 — Memory + CID (COMPLETADA — ver PHASE3_MEMORY_CID.md)

Resumen: working/episodic/semantic + `storage/cid` LocalStore + CLI memory.*

### Histórico de especificación (referencia)

### Objetivo

```
Organism
  → Memory (working / episodic / semantic)
  → CID de contenido cuando aplique
  → Persist
  → Restore
```

### Requisitos

1. **No romper** la API pública de `Organism` / `Manager` / CLI existente.
2. Ampliar `MemoryRef` (o tipo `Memory`) con:
   - `Working map[string]string` (ya existe)
   - `Episodic []Episode` (eventos con timestamp, tipo, payload/CID)
   - `Semantic map[string]string` o refs a CID (hechos estables)
3. Almacenamiento de bloques/contenido direccionable bajo `storage/` o `core` mínimo:
   - `Put(content []byte) (cid string, error)`
   - `Get(cid string) ([]byte, error)`
4. RootCID: **puede** empezar a usar `github.com/ipfs/go-cid` + multihash, pero mantener compatibilidad o migración clara. Si se cambia el formato, documentarlo y tests de regresión.
5. CLI mínima adicional (opcional pero útil):
   - `prismatec organism memory set <id> <key> <value>`
   - `prismatec organism memory get <id> <key>`
   - o `inspect` ya muestra working keys
6. Tests:
   - mismo contenido → mismo CID
   - PutMemory / episodic append → persist → reload
   - organism restore conserva memoria
7. Documentar en `docs/PHASE3_MEMORY_CID.md`

### Qué reutilizar

- AlsetOS `memoria` (simple JSON) — solo como idea
- PrismaTec persistencia CID/bloques — como referencia de diseño
- `Manager.PutMemory` ya existe — extender, no duplicar

### Qué NO hacer en Fase 3

- No integrar libp2p
- No integrar Alset-JS todavía
- No implementar Mind/Gen
- No reescribir el modelo Organism desde cero
- No añadir dependencias innecesarias

---

## 6. Orden de fases restantes (obligatorio)

```
3  Memory + CID
4  Events + Pulse (bus interno unificado hacia fuera)
5  AIP v1
6  Alset-JS mínimo (Demo 1)
7  Capabilities + Policy
8  Execution (WASM / Lisp modules)
9  Intelligence providers (interfaces)
10 Network (interfaces + libp2p)
11 Replication
12 Failure / Recovery (Demo 3)
13 Studio progresivo
14 End-to-end demo
15 Security hardening
16 Documentation + investor package (solo con demos reales)
```

---

## 7. Cómo trabajar sin romper nada

1. **Leer** este HANDOFF + `DECISIONS.md` + la `PHASEx` de la fase anterior.
2. **Correr** `go test ./...` antes de tocar código (baseline verde).
3. Implementar en paquetes existentes o crear bajo la estructura ya definida (`storage/cid`, `core/memory`, etc.).
4. **No** renombrar paquetes públicos (`core`, `organism`, `identity`) sin necesidad.
5. Tras cada cambio: `go test ./...` + smoke CLI.
6. Actualizar:
   - `docs/PHASE{N}_*.md`
   - este `docs/HANDOFF.md` (sección estado + fecha)
   - `README.md` status
7. Si fallas: STOP → DIAGNOSE → FIX → TEST. No apilar capas rotas.
8. Preguntar siempre: ¿es Core, capability, adapter, application o demo? Si puede esperar → no.

### Comandos de verificación baseline

```bash
cd /home/workdir/artifacts/PrismaTec-Core
go test ./...
PRISMATEC_DATA_DIR=/tmp/ptc-smoke go run ./cmd/prismatec organism create smoke-agent
PRISMATEC_DATA_DIR=/tmp/ptc-smoke go run ./cmd/prismatec organism inspect smoke-agent
```

---

## 8. Decisiones técnicas ya tomadas (no revertir sin ADR)

| Tema | Decisión |
|------|----------|
| Lenguaje Core | Go |
| NodeID | Ed25519, formato `node:` + base64url(pubkey) |
| RootCID actual | `rootcid:` + hex(sha256(manifest JSON canónico)) |
| Persistencia local | JSON files bajo DataDir |
| Policy v0 | `map[string]bool` por capability name |
| Package names | inglés (`organism`, no `organismo`) |
| CLI name | `prismatec` |
| Event bus | in-process, tipos string |

---

## 9. Contacto con el manifiesto de producto

Posicionamiento técnico provisional:

> PrismaTec Core is an infrastructure runtime for persistent, autonomous and distributed digital entities.

No usar frases tipo “AGI OS”, “revolutionary AI”, “new internet” en docs técnicos.

Mercados futuros (solo diseño, no implementación): agents, pharma, industrial, robotics, enterprise agent control plane, etc.

---

## 10. Checklist para la siguiente IA al retomar

- [ ] Leer `docs/HANDOFF.md` completo
- [ ] Leer `docs/DECISIONS.md` y `docs/MIGRATION.md`
- [ ] `go test ./...` verde
- [ ] Identificar fase actual (si Fase 3 incompleta → terminarla)
- [ ] Implementar solo el alcance de esa fase
- [ ] Tests + CLI smoke
- [ ] Escribir `docs/PHASE{N}_*.md`
- [ ] Actualizar sección “Estado actual” de este HANDOFF
- [ ] No avanzar dos fases en un solo salto sin verificación

---

## 11. Archivos clave para no perder el hilo

| Archivo | Por qué |
|---------|---------|
| `docs/HANDOFF.md` | Continuidad entre sesiones/IAs |
| `docs/AUDIT.md` | Qué hay en los 3 repos origen |
| `docs/ARCHITECTURE.md` | Mapa objetivo |
| `docs/MIGRATION.md` | Orden de trabajo |
| `docs/DECISIONS.md` | ADRs |
| `docs/PHASE1_BOOT.md` | Qué se hizo en Fase 1 |
| `docs/PHASE2_ORGANISM.md` | Qué se hizo en Fase 2 |
| `core/organism/*.go` | Corazón del modelo actual |
| `core/node.go` | Punto de entrada del runtime |
| `cmd/prismatec/main.go` | CLI |

---

**Fin del handoff.**  
Fase 3 completada. Siguiente: Fase 4 Events + Pulse. Actualizar este HANDOFF al cerrar cada fase.
