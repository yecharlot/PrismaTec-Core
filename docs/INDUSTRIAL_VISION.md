# Visión industrial y de inversor (sin humo)

## Qué es PrismaTec Core hoy

Una **infraestructura ejecutable** para entidades digitales persistentes (Digital Organisms):

- Identidad soberana (NodeID / PeerID / RootCID)
- Estado + memoria content-addressed
- Policy + audit
- Ejecución (WASM) e interfaces de inferencia
- **Red descentralizada** (TCP + **libp2p**)
- Réplica y recovery
- AIP + Control Studio

No es un demo de slides. Es un runtime en Go con tests E2E y transporte P2P real.

## Por qué importa (oportunidad)

El mercado construye “agentes” encima de APIs centralizadas. PrismaTec invierte la pila:

```
Apps del futuro (industria, edge, robots, enterprise)
        ↓
   AIP / Studio / SDKs
        ↓
   PrismaTec Core (organismos, policy, pulse, red)
        ↓
   Tu silicio / tu cloud / tu edge — soberanía
```

**Eficiencia:** un organismo, un RootCID, réplicas solo donde hace falta.  
**Autonomía:** lifecycle + execution + recovery sin “orquestador único”.  
**Seguridad:** least privilege, audit, Noise en libp2p, token AIP.  
**Soberanía:** sin atarte a un LLM vendor ni a un único hyperscaler.

## Qué verá un inversor técnico serio

1. `go test ./...` y `go test ./tests/e2e/ -v` — demos 1–3.  
2. Studio en `:8080/studio/` con datos reales.  
3. libp2p peer-to-peer con el mismo código de replication.  
4. Documentación de cumplimiento: `docs/MANIFESTO_COMPLIANCE.md`.

## Qué aún no debe venderse como hecho

- Flotas globales con DHT/relay de producción  
- SLA 99.99 y certificaciones  
- Marketplace de organismos  

Eso es **roadmap de producto** sobre un Core ya estructurado para no reescribirse.

## Mensaje corto

> PrismaTec Core is the missing OS layer for autonomous digital entities — decentralized by design, industrial in architecture, open in sovereignty.
