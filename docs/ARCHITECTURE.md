# ARCHITECTURE MAP — PrismaTec Core

**Versión:** 0.1 (post-auditoría)  
**Principio rector:** Core proporciona primitives. Applications consumen primitives.

---

## 1. Visión de capas

```
                    FUTURE APPLICATIONS
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
       AI                ROBOTICS            WEB / ENTERPRISE
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                    PRISMATEC CORE
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
     Identity            Execution           Memory
        │                   │                   │
     Policy             Network             Events
        │                   │                   │
   Provenance          Capabilities          Pulse
                            │
                       Host Systems
                   (Linux / Windows / macOS / Edge)
```

---

## 2. Arquitectura objetivo del repositorio

```
prismatec-core/
│
├── cmd/
│   └── prismatec/                 # CLI: node start, organism create/inspect/...
│
├── core/
│   ├── identity/                  # RootCID, NodeID, PeerID (separados)
│   ├── organism/                  # Modelo canónico + lifecycle
│   ├── state/                     # Estado observable
│   ├── memory/                    # Working / Episodic / Semantic + CID refs
│   ├── capability/                # Qué puede hacer un organismo
│   ├── policy/                    # WHO CAN DO WHAT TO WHICH UNDER CONDITIONS
│   ├── execution/                 # Interfaces de ejecución
│   ├── provenance/                # Quién autorizó / qué cambió
│   ├── lifecycle/                 # create → start → stop → move → recover
│   ├── events/                    # EventBus interno
│   ├── pulse/                     # Emisor de Pulse (hacia AIP / UI / red)
│   └── registry/                  # Registro de nodos y organismos
│
├── network/
│   ├── interfaces.go              # Transport, Discovery, Messaging, Replication, Presence
│   ├── libp2p/                    # Adapter libp2p (mDNS, DHT, Gossip, streams)
│   ├── dht/
│   ├── discovery/
│   ├── gossip/
│   └── protocols/
│
├── runtime/
│   ├── wasm/                      # wazero sandbox
│   ├── lisp/                      # LispAI + Zyrion (módulo opcional)
│   └── agents/                    # Interfaces de agentes (no Mind/Gen concretos)
│
├── storage/
│   ├── cid/                       # Generación y verificación de CID
│   ├── blocks/                    # Bloques de contenido
│   ├── local/                     # Persistencia local
│   └── distributed/               # Replicación de bloques
│
├── api/
│   ├── http/                      # REST mínimo
│   ├── websocket/
│   ├── sse/                       # Transporte de Pulse
│   └── aip/                       # Contrato AIP v1
│
├── adapters/
│   ├── cloudflare/                # Durable Objects / Workers (opcional)
│   ├── supabase/                  # Opcional
│   └── external/
│
├── sdk/
│   ├── go/
│   └── js/                        # Mínimo para conectar Alset-JS
│
├── examples/
├── demos/                         # Demo 1, 2, 3
├── docs/
└── tests/
```

---

## 3. Modelo de Digital Organism (canónico)

```go
// Conceptual — no implementación final todavía
type Organism struct {
    ID            string
    RootCID       string          // identidad del contenido / definición
    Version       string

    Identity      Identity        // NodeID / claves asociadas
    State         State
    Memory        MemoryRef       // refs a CID + working memory
    Capabilities  []Capability
    Policy        Policy
    Runtime       RuntimeSpec     // WASM / Lisp / Agent / External
    Provenance    Provenance
    Placement     Placement       // Primary + Replicas
    Status        Status          // Created | Ready | Running | Stopped | Error | Recovering

    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

**Identidades separadas (inviolable):**
- `RootCID` → identidad del organismo / contenido
- `NodeID`  → identidad persistente del nodo que lo ejecuta
- `PeerID`  → identidad de transporte (libp2p)

Un organismo puede moverse de Node A → Node B manteniendo RootCID.

---

## 4. Flujo de Pulse + AIP

```
Organism (Core)
    │
    ▼
EventBus / Pulse Emitter
    │
    ├──→ logs / audit / provenance
    ├──→ memoria
    ├──→ red (Gossip / streams)
    └──→ AIP transport (SSE / WebSocket / HTTP)
            │
            ▼
        Alset-JS Runtime
            │
            ▼
        AlsetRegistry (PIN)
            │
            ▼
        Localized update (solo el nodo UI afectado)
```

---

## 5. Separación Capability vs Policy

```
Capability          →  "puedo hacer X" (declaración)
Policy              →  "quién puede, sobre qué, bajo qué condiciones"
Authorization       →  decisión en tiempo de ejecución
Execution           →  solo si Authorization = allow
```

Ejemplo:
```
Capability: database.read
Policy:     AgentA may database.read on resource "patients" when role=researcher
```

---

## 6. Providers / Adapters (future-proof)

Todo lo que no es primitiva entra por interface:

- `StorageProvider`
- `NetworkProvider` (libp2p es la primera implementación)
- `InferenceProvider` (OpenAI, Ollama, local, custom…)
- `ExecutionProvider` (WASM, Lisp, Agent…)
- `IdentityProvider`
- `PolicyProvider`
- `PersistenceProvider`
- `UITransport` (AIP)

---

## 7. Escenarios soportados por diseño (sin implementar todo)

- Local / Edge / Cloud / Hybrid
- Organism movement
- Primary + Replicas + Failure recovery
- Pharma / Biotech (experiment organism + approval gates)
- Industrial / Robotics / Drones (local policy cuando red cae)
- Enterprise Agent Control Plane
- Digital twins, workflows, sensors, datasets (no solo AI agents)

---

## 8. Criterio de no-monolito

Aunque sea un solo repositorio al inicio, las responsabilidades están desacopladas.  
Posteriormente se podrá separar en:
```
prismatec-core
prismatec-node
prismatec-sdk
prismatec-js
prismatec-studio
```

---

## 9. Principio más importante

**Separar WHAT de HOW.**

```
Organism.StoreMemory(...)     →  no sabe si es SQLite / IPFS / Cloudflare
Organism.SendMessage(...)     →  no sabe si es libp2p / WebSocket / HTTP
Organism.Execute(...)         →  no sabe si es WASM / Lisp / External model
```

El Core habla en primitives. Los adapters implementan el how.
