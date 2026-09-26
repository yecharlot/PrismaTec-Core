# MVP PLAN — PrismaTec Core

**Prioridad absoluta:** REAL SYSTEM > DEMO FALSA > MARKETING

---

## Demo 1 — Create → Pulse → Localized UI (obligatorio)

```
User opens browser
        ↓
Alset JS Runtime (mínimo: Registry + Pulse + 1 panel)
        ↓
AIP v1
        ↓
PrismaTec Core Node
        ↓
Create Organism
        ↓
Assign Identity + Capabilities + Memory
        ↓
Start organism
        ↓
Organism emits Pulse (status / action / memory count…)
        ↓
Browser receives Pulse
        ↓
UI updates locally with REAL data from Core
```

**Pantalla objetivo (datos reales):**
```
┌─────────────────────────────────────────────┐
│             PRISMATEC CORE                 │
├─────────────────────────────────────────────┤
│  ORGANISM                                   │
│  ID: research-agent-01                     │
│  RootCID: bafy...                           │
│  Node: node-local-01                        │
│  STATUS       ● RUNNING                     │
│  MEMORY       12 episodes                   │
│  CAPABILITIES 4                             │
│  PULSE        128 events                    │
│  CURRENT ACTION  "Analyzing sample..."      │
│  POLICY                                      │
│  ✓ memory.read                              │
│  ✓ inference                                │
│  ✕ database.write                           │
└─────────────────────────────────────────────┘
```

---

## Demo 2 — Organism Movement

```
Node A ──(organism)──▶ Node B
RootCID, estado, memoria, provenance se conservan.
UI muestra Before / After.
```

---

## Demo 3 — Failure / Recovery

```
Primary = Node A
Replica = Node B
Apagar Node A
→ detect failure
→ locate replica
→ verify RootCID / state
→ activate organism on Node B
→ continue
UI muestra: NODE A OFFLINE → RECOVERING → NODE B ACTIVE → SUCCESS
```

---

## Checklist de aceptación MVP

- [ ] `prismatec node start` arranca y muestra NodeID
- [ ] `prismatec organism create` crea organismo con RootCID real
- [ ] `prismatec organism inspect` muestra estado real
- [ ] Organismo emite Pulse
- [ ] Alset-JS recibe Pulse vía AIP y actualiza solo el nodo afectado
- [ ] Datos de la UI provienen del Core (no hardcodeados)
- [ ] Tests unitarios de identity / organism / memory / pulse pasan
- [ ] (Opcional inmediato) movimiento de organismo entre dos nodos locales
- [ ] (Opcional inmediato) recovery desde réplica

---

## Qué NO entra en el MVP

- Kubernetes, blockchain, AGI, full cloud platform
- IAM enterprise completo
- Mind/Gen completos como dependencia del Core
- Studio completo
- Universal AI / universal frontend / universal OS
- Hardcoding de OpenAI, Cloudflare, AWS, etc.

---

## Siguiente acción después de estos documentos

**Fase 1 — Core Boot**

1. Crear estructura de repositorio `prismatec-core`
2. Inicializar `go.mod`
3. Implementar `core/identity` (basado en AlsetOS)
4. Implementar Node mínimo + Registry + EventBus
5. CLI: `prismatec node start`
6. Tests + verificación

Solo después de que eso funcione se avanza a Organism.
