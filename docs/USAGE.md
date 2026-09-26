# Cómo usar PrismaTec Core

Guía práctica para desarrolladores y operadores.  
Repositorio: https://github.com/yecharlot/PrismaTec-Core

---

## 1. Requisitos

- Go **1.25+** (o `GOTOOLCHAIN=go1.25.0`)
- Red para descargar módulos la primera vez

```bash
export GOPROXY=https://proxy.golang.org,direct
export GOTOOLCHAIN=go1.25.0
git clone https://github.com/yecharlot/PrismaTec-Core.git
cd PrismaTec-Core
go test ./...
```

---

## 2. Arrancar un nodo

```bash
# directorio de datos (identidad, organismos, audit)
export PRISMATEC_DATA_DIR=$HOME/.prismatec/node
export PRISMATEC_NODE_NAME=node-local-01

# opcional: proteger AIP con token (recomendado fuera de localhost)
# export PRISMATEC_AIP_TOKEN=cambia-este-secreto

go run ./cmd/prismatec node start
```

Salida típica:

- Nodo en ejecución  
- **AIP** en `http://localhost:8080` (o `PRISMATEC_AIP_ADDR`)  
- **Control Studio:** http://127.0.0.1:8080/studio/  
- Demo panel: http://127.0.0.1:8080/demo/

| Variable | Default | Uso |
|----------|---------|-----|
| `PRISMATEC_DATA_DIR` | `~/.prismatec/node` | Estado local |
| `PRISMATEC_NODE_NAME` | `node-local-01` | Nombre legible |
| `PRISMATEC_AIP_ADDR` | `:8080` | Bind HTTP/AIP |
| `PRISMATEC_AIP_TOKEN` | (vacío) | Si se define, exige Bearer token en `/aip/*` |

---

## 3. CLI

```bash
go run ./cmd/prismatec version
go run ./cmd/prismatec node info
go run ./cmd/prismatec organism create research-agent --cap memory.read --cap inference
go run ./cmd/prismatec organism list
go run ./cmd/prismatec organism inspect <id-or-name>
go run ./cmd/prismatec organism start <id>
go run ./cmd/prismatec organism stop <id>
go run ./cmd/prismatec pulse list 20
go run ./cmd/prismatec demo e2e
```

---

## 4. AIP v1 (API HTTP)

Base: `http://127.0.0.1:8080`

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/healthz` | Liveness |
| GET | `/aip/v1/info` | Nodo |
| GET | `/aip/v1/organisms` | Lista |
| GET | `/aip/v1/organisms/{id}` | Detalle |
| GET | `/aip/v1/pulses` | Historial reciente |
| GET | `/aip/v1/pulse` | **SSE** stream |
| GET | `/aip/v1/audit` | Eventos de seguridad |
| POST | `/aip/v1/commands` | Comandos |

### Comandos (`POST /aip/v1/commands`)

```json
{"aip":"v1","action":"create","params":{"name":"agent-01","capabilities":["memory.read","inference"]}}
{"aip":"v1","action":"start","params":{"id":"..."}}
{"aip":"v1","action":"stop","params":{"id":"..."}}
{"aip":"v1","action":"memory.set","params":{"id":"...","key":"focus","value":"x"}}
{"aip":"v1","action":"execute","params":{"id":"...","entry":"ping"}}
{"aip":"v1","action":"infer","params":{"prompt":"hello"}}
{"aip":"v1","action":"policy.check","params":{"id":"...","action":"memory.write"}}
```

Con token:

```bash
curl -s -H "Authorization: Bearer $PRISMATEC_AIP_TOKEN" \
  http://127.0.0.1:8080/aip/v1/info
```

---

## 5. Control Studio

1. `go run ./cmd/prismatec node start`  
2. Abrir http://127.0.0.1:8080/studio/  
3. Connect → Create organism → Start → Memory → ver Pulse  

No usa datos mock: todo pasa por AIP.

---

## 6. Conceptos clave

| Concepto | Significado |
|----------|-------------|
| **Organism** | Unidad digital persistente (agente, servicio, robot lógico…) |
| **RootCID** | Identidad de **contenido** (definición) |
| **NodeID** | Identidad criptográfica del **host** |
| **Pulse** | Actualización localizada (UI/clientes reaccionan por `target`) |
| **Policy** | WHO can DO WHAT under CONDITIONS |
| **AIP** | Contrato versionado Core ↔ clientes |

---

## 7. Tests y demo E2E

```bash
go test ./...
go test ./tests/e2e/ -count=1 -v   # 12 pasos del manifiesto
go run ./cmd/prismatec demo e2e
```

---

## 8. Documentación en el repo

| Archivo | Contenido |
|---------|-----------|
| `docs/USAGE.md` | Esta guía |
| `docs/SECURITY.md` | Seguridad y hardening |
| `docs/ARCHITECTURE.md` | Mapa arquitectónico |
| `docs/HANDOFF.md` | Continuidad para otra IA/dev |
| `docs/PHASE*.md` | Historial por fase |
| `docs/STATUS.md` | Qué hay / qué falta |

---

## 9. Qué no es (todavía)

- No es un framework de chatbots  
- libp2p multi-host en producción: adapter pendiente (hay fabric local para tests)  
- Modelos LLM reales: interfaz lista (`inference`), provider `echo` por defecto  
