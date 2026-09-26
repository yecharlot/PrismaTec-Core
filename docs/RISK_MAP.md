# RISK MAP — PrismaTec Core

| ID | Riesgo | Prob. | Impacto | Mitigación |
|----|--------|-------|---------|------------|
| R1 | Intentar implementar todo de una vez | Alta | Crítico | Seguir MIGRATION.md fase por fase. Criterio: cada fase debe BUILD+TEST+RUN. |
| R2 | Mezclar Core con Applications (Mind/Gen) | Alta | Alto | ADR-003. Mind/Gen solo como módulos opcionales o examples. |
| R3 | Demo falsa (UI con datos inventados) | Media | Alto | Criterio de aceptación: todos los datos de la UI vienen del Core vía Pulse/AIP. |
| R4 | RootCID inconsistente entre repos | Media | Alto | ADR-007. Definir formato canónico en Fase 1-3. |
| R5 | Acoplamiento directo Organism ↔ libp2p | Media | Medio | Interfaces Network primero (ADR-004). |
| R6 | Policy demasiado compleja demasiado pronto | Media | Medio | API diseñada para crecer; implementación mínima (bool + conditions simples). |
| R7 | Pérdida de tests existentes | Media | Medio | Portar tests junto con el código; suite de integración nueva. |
| R8 | Clone / tamaño de PrismaTec ralentiza desarrollo | Baja | Bajo | Extraer solo lo necesario; no depender del checkout completo. |
| R9 | Scope creep hacia “OS universal” o “AGI” | Media | Alto | ADR-008 + regla de complejidad: si puede esperar, no se implementa. |
| R10 | Seguridad como afterthought | Media | Alto | ADR-009. Identity + capability isolation + policy + sandbox desde el principio. |

**Principio de contención:**  
Cada vez que se quiera añadir una pieza, preguntar:

1. ¿Es necesaria para el Core?
2. ¿Es una capability?
3. ¿Es un adapter?
4. ¿Es una application?
5. ¿Es solo una demo?
6. ¿Puede esperar?

Si puede esperar → no se implementa todavía.
