# Phase 15 — Security hardening (completed)

**Date:** 2026-09-26

## Qué se hizo

- `core/audit` — log JSONL + buffer (`audit.jsonl` en DataDir)
- `core/security` — validación de name/id/action/memory; token AIP
- AIP: middleware auth si `PRISMATEC_AIP_TOKEN`; audit de comandos y denegaciones
- `GET /aip/v1/audit`
- Docs: **USAGE.md**, **SECURITY.md**, README actualizado

## Por qué

Antes de “vender” la plataforma hace falta un mínimo de isolation, audit y auth opcional sin romper el demo local.

## Pensamiento futuro

- Studio envía token automáticamente  
- Firmas en mensajes de red  
- Rate limits  

## Guía de acción

Ver `docs/USAGE.md` y `docs/SECURITY.md`.  
Siguiente: Phase 16 documentation / investor package (completar si USAGE+SECURITY ya cubren uso; package comercial opcional).
