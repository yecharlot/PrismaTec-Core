# PrismaTec-Core — Estado (ES)

**Versión orientativa:** 0.3.x · runtime de organismos digitales

## Qué es

Motor de producto: identidad, organismos, memoria, policy default-deny, Pulse, provenance, AIP, Mind, Zyrion, Transport (TCP + libp2p), réplica/recuperación bajo prueba.

## Qué NO es

No es Pulso Cubano ni AbacoPhy. Los verticales consumen Core vía adaptadores.

## Operativo

- Multi-nodo TCP, PeerID, fencing por época (evidencia en tests e2e)
- Mind propone; Core autoriza y ejecuta
- Studio / AIP para observación

## Pendiente de madurez industrial

- Particiones de red + RPO/RTO medidos en carga
- Firmas de mensajes entre nodos en WAN hostil
- Contrato AIP versionado y congelado para terceros
