# Phase 14 — End-to-end demonstration (completed)

**Date:** 2026-09-26

## Qué se hizo

### Automated checklist (`tests/e2e`)

`TestE2E_ManifestChecklist` ejecuta los 12 pasos del manifiesto con código real:

1. Create organism  
2. Identity (ID + RootCID)  
3. Memory  
4. Capabilities  
5. Start  
6. Observe  
7. Command (builtin execute ping)  
8. Pulse (organism.started)  
9. Persist / restore  
10. Replicate (node-A → node-B)  
11. Failure (primary offline)  
12. Recover (promote replica, RootCID preserved)  

También `TestE2E_AIP_HTTP` cubre create/start/memory/execute/infer/policy vía HTTP.

### CLI

```bash
go run ./cmd/prismatec demo e2e
go test ./tests/e2e/ -count=1 -v
```

### Studio path (manual)

```bash
go run ./cmd/prismatec node start
# http://127.0.0.1:8080/studio/
```

## Por qué

El criterio de éxito del proyecto es **sistema real demostrable**, no marketing. Esta fase formaliza y automatiza ese criterio.

## Pensamiento futuro

- Multi-proceso con dos `node start` + adapter libp2p  
- Script que abra Studio y valide SSE en CI (headless)  
- Phase 15: security hardening sobre este camino E2E  

## Guía de acción

```bash
export GOPROXY=https://proxy.golang.org,direct GOTOOLCHAIN=go1.25.0
go test ./tests/e2e/ -count=1 -v
go test ./...
go run ./cmd/prismatec demo e2e
```
