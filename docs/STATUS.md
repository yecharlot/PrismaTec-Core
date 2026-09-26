# STATUS — Qué tenemos vs propósito final

**Actualizado:** 2026-09-26 (cierre Fases 8–12)

## Propósito final (manifiesto)

Construir **PrismaTec Core**: infraestructura para **Digital Organisms** (entidades digitales persistentes, identificables, ejecutables, observables, distribuibles) con:

Identity · State · Memory · Capability · Policy · Execution · Network · Provenance · Pulse · Interface (AIP)

**No** es un framework de agentes, un frontend, ni un chatbot. Es la capa sobre la que se construyen plataformas de próxima generación.

**Criterio de éxito MVP demostrable:**

1. Browser → Alset-JS → AIP → Core → Create organism → Pulse real → UI localizada  
2. Move organism Node A → Node B  
3. Failure → recovery desde réplica  

---

## Completado (Fases 0–4)

| Fase | Entregable | Estado |
|------|------------|--------|
| 0 | Auditoría 3 repos + mapas | ✅ |
| 1 | Node, Identity (NodeID), EventBus, Registry, CLI node | ✅ |
| 2 | Organism model, lifecycle, RootCID, persist, CLI organism | ✅ |
| 3 | Memory working/episodic/semantic + block store CID | ✅ |
| 3b | go-cid / IPFS CIDv1 disponible + cid1 compat | ✅ |
| 4 | Pulse hub + bridge events→pulse + CLI pulse list | ✅ |
| 5 | AIP v1 HTTP + SSE + commands | ✅ |

### Capacidades reales hoy

- Nodo con **NodeID** Ed25519 persistente  
- **Organismos** create/start/stop/inspect/list con **RootCID** determinista  
- **Policy** mínima (allow/deny por capability)  
- **Memoria** tipada + **bloques** `cid1:` e IPFS CIDv1  
- **Eventos** internos + **Pulses** dirigidos a target (organismo/nodo)  
- CLI `prismatec` operativa  
- Tests unitarios verdes  

### Paths importantes

```
/home/workdir/artifacts/PrismaTec-Core
module: github.com/yecharlot/PrismaTec-Core  (go 1.25.0)
```

---

## Falta para el propósito / MVP

| Fase | Qué falta | Por qué importa |
|------|-----------|-----------------|
| **5 AIP v1** | Contrato versionado + transporte (SSE/WS/HTTP) | Sin esto el browser no habla con el Core |
| **6 Alset-JS mínimo** | Registry + Pulse client + panel real | Demo 1 (UI con datos del Core) |
| **7 Policy** | WHO/CAN/WHAT/CONDITIONS | Enterprise / compliance |
| **8 Execution** | WASM (wazero) / Lisp modules | Organismo “hace” algo sandboxed |
| **9 Inference providers** | Interfaces only | No atarse a un vendor |
| **10 Network** | libp2p adapter detrás de interfaces | Multi-nodo |
| **11 Replication** | Primary + replicas | Base de Demo 2/3 |
| **12 Recovery** | Detect failure → activate replica | Demo 3 inversores |
| **13–14** | Studio + demo E2E | Observabilidad visual |
| **15–16** | Security hardening + docs investor | Solo con demos reales |

### Gap crítico hacia Demo 1

```
[x] Core organism + memory + pulse (in-process)
[x] AIP transport from long-running node (HTTP + SSE)
[ ] Alset-JS receives pulse and updates one PIN
```

**Recomendación:** Fase 6 — cliente JS mínimo contra `http://localhost:8080/aip/v1/*`.

---

## Cómo continuar (cualquier IA)

1. Leer **`docs/HANDOFF.md`** y este **`docs/STATUS.md`**  
2. `cd PrismaTec-Core && go test ./...`  
3. Si `go get` falla: `GOPROXY=https://proxy.golang.org,direct`  
4. Implementar **solo la fase siguiente** en `MIGRATION.md`  
5. Documentar en `docs/PHASE{N}_*.md` + actualizar HANDOFF + STATUS  
6. Ciclo: IMPLEMENT → BUILD → TEST → RUN → VERIFY → DOCUMENT  

### Reglas que no se negocian

- Core ≠ Applications (Mind/Gen fuera del core obligatorio)  
- RootCID / NodeID / PeerID separados  
- No demo falsa (UI debe leer datos del Core)  
- No saltar a red/Studio antes de AIP  

---

## Pensamiento futuro (corto)

- Pulse log debería vivir en el proceso `node start` + AIP, no en CLI one-shot.  
- RootCID puede migrar a CIDv1 de manifiesto cuando se quiera unificar con bloques IPFS.  
- Policy engine debe crecer sin romper `map[string]bool` actual (capa encima).  
- libp2p debe entrar como **adapter**, no como dependencia del tipo Organism.
