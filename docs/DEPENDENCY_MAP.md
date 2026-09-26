# DEPENDENCY MAP — PrismaTec Core

## Dependencias técnicas previstas (Go)

| Área | Dependencia | Origen / Notas |
|------|-------------|----------------|
| Network | github.com/libp2p/go-libp2p | AlsetOS + PrismaTec |
| DHT | github.com/libp2p/go-libp2p-kad-dht | AlsetOS |
| CID | github.com/ipfs/go-cid | AlsetOS / PrismaTec |
| Datastore | github.com/ipfs/go-datastore (+ pebble) | AlsetOS |
| WASM | github.com/tetratelabs/wazero | AlsetOS |
| Multiformats | go-multihash, go-multiaddr | AlsetOS |
| Crypto | crypto/ed25519 (stdlib) | AlsetOS |

## Providers / Adapters (interfaces)

| Provider | Primera implementación | Futuras |
|----------|------------------------|---------|
| NetworkProvider | libp2p | WebSocket, HTTP, custom |
| StorageProvider | local + CID blocks | Cloudflare DO, Supabase, IPFS, Postgres |
| ExecutionProvider | WASM (wazero), Lisp | Agent runtime, containers (más tarde) |
| InferenceProvider | (ninguno hardcodeado) | OpenAI, Anthropic, Ollama, local, HTTP |
| UITransport | AIP over SSE / WebSocket | custom |
| IdentityProvider | Ed25519 local | HSM, cloud KMS (más tarde) |
| PolicyProvider | in-memory rules | OPA, custom enterprise |

## Dependencias de Alset-JS (SDK mínimo)

- Solo lo necesario para Demo 1: AlsetRegistry, Pulse handling, AIP client, componentes mínimos.
- Studio completo queda como fase posterior.

## Regla

Ninguna dependencia de aplicación (Mind, Gen, apps concretas) entra en el `go.mod` del Core como requisito obligatorio.
