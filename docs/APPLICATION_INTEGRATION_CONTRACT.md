# Contrato Universal de Integración de Aplicaciones con PrismaTEC-Core

**Identificador:** PTC-AIC-001  
**Versión:** 0.1.0 (borrador normativo)  
**Estado:** Especificación arquitectónica propuesta; todavía no equivale a un SDK ni a una API implementada.  
**Ámbito:** Integración de aplicaciones de dominio con PrismaTEC-Core.  
**Primera implementación de referencia:** Smart-WMS, rama `feat/wms-prismatec-genesis`.

---

## 1. Propósito

Defino este contrato para que cualquier aplicación pueda utilizar PrismaTEC-Core como infraestructura común sin tener que conocer sus paquetes internos ni trasladar su lógica de negocio al núcleo.

Quiero que PrismaTEC-Core proporcione los mecanismos generales para que las entidades digitales puedan tener identidad, estado, memoria, historial, capacidades, políticas, procedencia, ejecución y distribución. Cada aplicación debe aportar el significado de esas entidades y las reglas que gobiernan su dominio.

El contrato establece una frontera estable entre:

1. **Aplicación de dominio:** contiene el modelo funcional, los casos de uso, las reglas y los invariantes propios del problema.
2. **Adaptador de integración:** traduce las operaciones y entidades del dominio a contratos de PrismaTEC-Core.
3. **PrismaTEC-Core:** administra las primitivas universales de los organismos digitales.
4. **Proveedores de infraestructura:** implementan almacenamiento, transporte, red y otros mecanismos sustituibles.

Este documento define responsabilidades y garantías esperadas. No afirma que todas estén implementadas hoy.

## 2. Lenguaje normativo y niveles de evidencia

Las palabras **DEBE**, **NO DEBE**, **DEBERÍA**, **PUEDE** y **NO SE RECOMIENDA** expresan requisitos normativos.

Cada capacidad debe etiquetarse en documentación, revisión y pruebas con uno de estos estados:

- **VERIFICADA:** existe evidencia reproducible en código y pruebas ejecutadas.
- **IMPLEMENTADA_NO_VALIDADA:** existe implementación identificable, pero falta evidencia suficiente de funcionamiento.
- **PARCIAL:** solo está cubierta una parte del contrato.
- **DOCUMENTADA:** está descrita, pero no se ha localizado implementación suficiente.
- **PROPUESTA:** forma parte del diseño futuro.
- **DESCONOCIDA:** no hay evidencia suficiente para emitir una conclusión.

No se debe elevar una capacidad de estado por aparecer en un README, una interfaz o un diagrama. La evidencia debe identificar código, prueba, resultado y versión revisada.

## 3. Principios no negociables

1. **Core = primitivas; aplicación = dominio.** El Core no incorpora reglas verticales de WMS, salud, finanzas, comercio u otros sectores.
2. **Separar significado de mecanismo.** La aplicación define qué representa una entidad; el Core resuelve cómo se identifica, persiste, observa y distribuye.
3. **No filtrar internals.** El dominio no importa paquetes internos del Core. Solo el adaptador de infraestructura puede depender de SDKs o contratos públicos.
4. **No duplicar primitivas.** Una aplicación no debe reconstruir por su cuenta identidad, RootCID, historial, autorización, procedencia o replicación cuando el Core ya proporcione esa capacidad.
5. **No asumir capacidades.** El adaptador debe comprobar qué versión y capacidades ofrece el Core en tiempo de ejecución o durante el arranque.
6. **No confundir persistencia con transacción.** Guardar varios organismos no significa que una operación de negocio sea atómica.
7. **No confundir replicación con consenso.** Replicar datos no determina por sí solo quién tiene autoridad para escribir.
8. **No confundir razonamiento con autorización.** Mind y Zyrion pueden producir conclusiones; el Core conserva la autoridad de política y ejecución.
9. **Evolución compatible.** Los contratos versionados deben permitir que las aplicaciones evolucionen sin romper las ya integradas.
10. **Evidencia antes que promesas.** Toda garantía crítica debe tener una prueba que pueda ejecutarse.

## 4. Arquitectura de integración

La arquitectura de referencia es:

```
┌────────────────────────────────────────────────────┐
│                 APLICACIÓN DE DOMINIO              │
│ Entidades · Casos de uso · Reglas · Invariantes    │
└────────────────────────┬───────────────────────────┘
                         │ puertos del dominio
┌────────────────────────▼───────────────────────────┐
│             ADAPTADOR PRISMA TEC                   │
│ Mapeo · Contratos · DTO · Errores · Idempotencia   │
└────────────────────────┬───────────────────────────┘
                         │ contrato público
┌────────────────────────▼───────────────────────────┐
│                 PRISMATEC-CORE                     │
│ Organismos · Identidad · Memoria · Policy · Events │
│ CID/RootCID · Provenance · Execution · Replication │
└────────────────────────┬───────────────────────────┘
                         │ providers
┌────────────────────────▼───────────────────────────┐
│          ALMACENAMIENTO / RED / TRANSPORTE         │
└────────────────────────────────────────────────────┘
```

El adaptador puede ejecutarse dentro del mismo proceso que el Core o comunicarse con un nodo remoto mediante AIP u otro protocolo público. La modalidad debe declararse explícitamente. No se debe documentar un Core remoto cuando la implementación lo enlaza directamente como dependencia Go.

### 4.1 Modalidades permitidas

- **Embebido:** el adaptador instancia o consume el runtime dentro del proceso de la aplicación.
- **Remoto:** el adaptador consume un nodo Core mediante un contrato de red versionado.
- **Híbrido:** una aplicación utiliza más de un nodo o proveedor, manteniendo una única frontera de dominio.

La aplicación no debe mezclar silenciosamente modalidades ni asumir que comparten las mismas garantías operativas.

## 5. Responsabilidades por capa

| Responsabilidad | Aplicación | Adaptador | Core |
|---|---:|---:|---:|
| Definir entidades y vocabulario del dominio | DEBE | Traduce | NO |
| Definir reglas e invariantes de negocio | DEBE | Hace cumplir el flujo técnico | NO |
| Identidad técnica del organismo | NO | Mapea | DEBE |
| NodeID y PeerID | NO | Consume | DEBE |
| RootCID y referencias de contenido | Define el contenido semántico | Serializa/mapa | Calcula, verifica y administra según contrato |
| Historial técnico de episodios | Define el evento de negocio | Traduce | Persiste y expone |
| Transacción de negocio | Define límites e invariantes | Orquesta | Ofrece primitivas disponibles; no inventa semántica |
| Autorización | Declara intención/actor/contexto | Solicita evaluación | Evalúa política y aplica el control de ejecución |
| Replicación y recuperación | Declara requisitos de dominio | Mapea estrategias | Proporciona mecanismos soportados |
| Presentación y UX | DEBE | Puede consumir eventos | No es responsabilidad del Core |

## 6. Contrato de organismo de dominio

Toda aplicación DEBE declarar cómo representa sus entidades persistentes como organismos o como recursos asociados a organismos.

Como mínimo, la declaración lógica de una entidad debe identificar:

- `domain`: identificador estable del dominio, por ejemplo `wms`, `health` o `commerce`.
- `type`: tipo semántico de entidad, por ejemplo `warehouse`, `patient` o `order`.
- `id`: identificador estable dentro del dominio.
- `schema_version`: versión del esquema de datos de dominio.
- `payload`: contenido de negocio serializable.
- `relations`: referencias a otras entidades, cuando existan.
- `metadata`: información técnica no sensible necesaria para indexación, trazabilidad o control.

La clave lógica recomendada es:

```
<domain>.<type>.<stable-id>
```

Ejemplos:

```
wms.warehouse.1
wms.zone.1
wms.bin.1
wms.purchase-order.PO-2026-0001
commerce.customer.C-100
health.patient.P-100
```

La clave lógica de dominio NO DEBE confundirse con RootCID, NodeID o PeerID.

### 6.1 Identidades

- **ID de dominio:** identifica la entidad según las reglas de la aplicación.
- **OrganismID:** identifica el organismo dentro del Core.
- **RootCID:** identifica contenido/manifiesto del organismo según la semántica implementada por Core.
- **NodeID:** identifica criptográficamente el nodo.
- **PeerID:** identifica el extremo de transporte.

No se deben reutilizar estos identificadores entre sí ni derivar autoridad de uno que solo identifica contenido.

### 6.2 Un organismo por entidad o por agregado

El contrato NO impone que cada fila de base de datos se convierta en un organismo.

La aplicación debe elegir una granularidad que preserve coherencia y ciclo de vida:

- entidad independiente cuando tenga identidad, evolución o distribución propia;
- agregado cuando varias entidades deban compartir invariantes y cambios coordinados;
- organismo especializado cuando represente un proceso o recurso con ciclo de vida propio.

La elección debe justificarse y probarse. No se debe crear un organismo por cada campo ni un organismo global gigantesco sin medir sus consecuencias.

## 7. Relaciones y referencias

La aplicación define el significado de las relaciones: pertenencia, dependencia, composición, autorización, transferencia, asociación u otras.

El adaptador DEBE conservar la identidad estable de los extremos y el tipo de relación. El Core PUEDE almacenar o distribuir las referencias mediante sus primitivas disponibles, pero no debe inferir semántica de negocio que no haya sido declarada.

Una referencia a otro organismo no demuestra que el organismo remoto esté disponible, actualizado o autorizado para una operación.

## 8. Estado, episodios y contenido

La aplicación es responsable de definir los estados de negocio y sus transiciones válidas. El Core es responsable de ofrecer los mecanismos técnicos que estén implementados para persistir estado, memoria, eventos y contenido.

El adaptador DEBE distinguir al menos:

- **Estado actual:** proyección útil para consultas y ejecución.
- **Episodio:** hecho ocurrido, con tipo, tiempo, actor/contexto y datos o referencia de contenido.
- **Contenido direccionable:** bytes o documento identificado por CID cuando se use almacenamiento content-addressed.
- **RootCID:** raíz/manifiesto de contenido del organismo, según el contrato de identidad del Core.

No se debe tratar el RootCID como sustituto automático de una versión de negocio, un número de secuencia, una transacción o una prueba de autorización.

### 8.1 Eventos de negocio y eventos técnicos

El dominio define eventos como `wms.receipt.accepted` o `commerce.order.cancelled`.

El Core puede registrar eventos técnicos como creación, cambio de memoria, replicación o recuperación.

El adaptador debe conservar la distinción y evitar convertir eventos técnicos en hechos de negocio sin una regla explícita.

### 8.2 Borrado

Cuando se requiera conservar historia, la aplicación debe representar la eliminación lógica mediante un evento/tombstone y definir sus reglas de retención. La eliminación física de bloques o datos debe seguir una política separada y explícita.

## 9. Operaciones, invariantes e idempotencia

Toda operación que modifique estado debe tener un identificador estable de operación generado en el límite de entrada o comando, no dentro de cada intento de ejecución.

El adaptador DEBE propagar:

- `operation_id`: identidad estable de la intención;
- `actor`: identidad del solicitante;
- `command_type`: operación solicitada;
- `target`: organismo o agregado afectado;
- `expected_version` o condición de concurrencia, cuando aplique;
- `payload_fingerprint`: huella determinista de los parámetros relevantes, o mecanismo equivalente.

Reglas mínimas:

1. Repetir la misma operación con el mismo ID y el mismo contenido no debe aplicar dos veces sus efectos.
2. Reutilizar el mismo ID con contenido distinto debe producir conflicto, no sobrescribir silenciosamente.
3. Un error posterior a un efecto parcial debe dejar evidencia suficiente para detectar y recuperar la operación.
4. Una operación rechazada por una invariante no debe reportarse como completada.
5. El adaptador no debe ignorar errores de operaciones que protegen invariantes del dominio.

### 9.1 Atomicidad de negocio

El Core no debe prometer una transacción ACID entre organismos si no ofrece explícitamente esa garantía.

Cuando una operación de dominio modifica varios recursos, la aplicación define el límite de consistencia. El adaptador implementa una de estas estrategias, justificando la elección:

- transacción atómica proporcionada por un almacenamiento que realmente la soporte;
- agregado único que encapsule la invariante;
- saga/flujo durable con estados intermedios, compensación y recuperación;
- protocolo de coordinación explícito cuando sea imprescindible.

La estrategia elegida debe sobrevivir a reinicios y reintentos. Un mutex en memoria no constituye por sí solo una garantía de atomicidad distribuida.

## 10. Concurrencia y autoridad de escritura

Toda escritura concurrente debe declarar cómo se resuelven las versiones y quién tiene autoridad para aceptar cambios.

El adaptador debe utilizar controles de versión, compare-and-swap, fencing o mecanismos equivalentes cuando estén disponibles.

Replicar una versión no autoriza a dos nodos a escribir simultáneamente. La política de autoridad debe estar separada del transporte y del descubrimiento de peers.

El contrato no exige consenso distribuido universal. Cada aplicación debe declarar su modelo de consistencia y sus límites: escritor único, autoridad por agregado, multi-writer con resolución explícita u otro modelo validado.

## 11. Identidad, capacidades, políticas y seguridad

La aplicación declara el actor y la intención de cada operación. El adaptador traduce esa intención a la evaluación de política del Core.

- **Capability:** declara una acción que un organismo puede realizar.
- **Policy:** determina quién puede ejecutar qué acción, sobre qué recurso y bajo qué condiciones.
- **Authorization:** decisión efectiva en tiempo de ejecución.
- **Execution:** solo puede continuar cuando la autorización aplicable lo permite.

Requisitos:

1. El adaptador NO DEBE considerar una capability como autorización suficiente.
2. Las operaciones sensibles DEBEN pasar por el punto de autorización definido.
3. La ausencia de política o de evidencia suficiente debe tratarse según el modelo fail-closed del Core; nunca se debe inventar un permiso.
4. El contexto de autorización debe incluir actor, acción, recurso y condiciones relevantes.
5. La aplicación conserva sus reglas de acceso de dominio; el Core no sustituye automáticamente su modelo de roles o cumplimiento normativo.
6. Las credenciales y secretos no deben almacenarse en episodios o payloads sin protección específica.
7. La identidad del organismo no equivale a la identidad de la persona o usuario que solicita una acción.

## 12. Replicación, disponibilidad y recuperación

La aplicación debe declarar qué organismos requieren replicación, qué pérdida de datos tolera y qué comportamiento espera ante desconexión.

El adaptador debe distinguir:

- contenido anunciado;
- contenido localmente disponible;
- contenido recuperado y verificado;
- versión aceptada como vigente;
- autoridad para promover una réplica;
- estado de recuperación pendiente o completada.

El Core debe verificar integridad de contenido mediante sus mecanismos CID disponibles. La existencia de un RootCID anunciado no prueba que todos sus bloques estén disponibles.

No se debe afirmar disponibilidad distribuida, RPO/RTO o failover automático sin pruebas que midan esos resultados.

La aplicación debe definir qué operaciones pueden ejecutarse desconectadas, cuáles deben bloquearse y cómo se reconcilian al reconectar. La resolución automática de conflictos no se presupone.

## 13. Mind y Zyrion

Mind y Zyrion son capacidades cognitivas opcionales, no dependencias obligatorias de toda aplicación.

La aplicación puede aportar hechos, premisas, reglas de dominio y evidencia. Mind puede producir inferencias trazables. Zyrion clasifica el resultado conforme a la semántica definida para este ecosistema:

- `1 = True`
- `0 = False`
- `2 = Indeterminado`

El estado `2` es un estado sumidero que activa otro nivel de razonamiento con las mismas reglas, habilitando el comportamiento fractal. No debe reinterpretarse como verdadero, falso o permiso.

La salida cognitiva debe conservar referencias a premisas, reglas, versiones y evidencias utilizadas. Una inferencia no concede permisos ni ejecuta operaciones. El flujo normativo es:

```
Hechos del dominio
      ↓
Mind: inferencia trazable
      ↓
Zyrion: clasificación 0 / 1 / 2
      ↓
Core Policy: autorización independiente
      ↓
Adaptador: ejecución de operación autorizada
      ↓
Episodio / provenance / resultado
```

La aplicación puede funcionar sin Mind/Zyrion y debe conservar un comportamiento seguro cuando estas capacidades no estén disponibles.

## 14. Errores y observabilidad

El adaptador debe traducir errores técnicos a errores estables del contrato sin borrar la causa original.

Como mínimo, debe distinguir:

- entrada inválida;
- entidad u organismo inexistente;
- versión obsoleta o conflicto concurrente;
- operación duplicada;
- reutilización conflictiva de operation_id;
- no autorizado;
- capacidad no disponible;
- proveedor no disponible;
- contenido no encontrado;
- integridad inválida;
- operación incompleta/requiere recuperación;
- error interno no clasificado.

Cada operación debe permitir correlacionar solicitud, operación, organismo, episodio y resultado mediante identificadores estables. Los logs no deben exponer secretos ni datos personales innecesarios.

Pulse/AIP pueden transportar observaciones y actualizaciones, pero la aplicación no debe usar un evento de notificación como única fuente de verdad para reconstruir estado crítico.

## 15. Versionado y compatibilidad

El contrato tiene su propia versión, independiente de la versión del Core y de la versión del esquema de dominio.

Cada aplicación debe declarar:

- versión mínima y máxima del contrato que soporta;
- versión/capacidades del Core probadas;
- versión de sus esquemas de dominio;
- política de migración de datos y organismos;
- comportamiento ante capacidades ausentes.

Cambios compatibles deben añadirse sin cambiar el significado de campos existentes. Cambios incompatibles requieren nueva versión mayor o una migración explícita. El adaptador debe rechazar versiones incompatibles de forma visible.

AIP v1 es una interfaz pública existente documentada en el repositorio, pero este documento no afirma que AIP implemente ya todos los campos y garantías aquí descritos. La evolución del contrato de aplicación debe mapearse a AIP/SDK cuando esas superficies estén implementadas.

## 16. Pruebas de conformidad obligatorias

Una aplicación no puede declararse integrada conforme solo porque arranca o puede crear un organismo.

Debe demostrar, al menos:

| Prueba | Resultado exigido |
|---|---|
| Declaración de dominio | Los tipos y esquemas se identifican sin introducir reglas en Core |
| Creación/lectura | El organismo puede crearse y recuperarse con identidad estable |
| Identidad | ID de dominio, OrganismID, RootCID, NodeID y PeerID no se confunden |
| Evolución | Cambios de estado generan versión/evidencia según el contrato |
| Historial | Episodios pueden correlacionarse con la operación de dominio |
| Idempotencia | Reintento idéntico no duplica efectos |
| Conflicto | Mismo operation_id con contenido distinto se rechaza |
| Invariantes | Fallos en pasos intermedios no dejan estado de negocio falso |
| Autorización | Operación denegada no produce efectos |
| Recuperación | Reinicio no pierde ni duplica operaciones confirmadas |
| Replicación | Integridad y versión se verifican en el escenario declarado |
| Compatibilidad | Contrato y capacidades incompatibles se detectan |
| Aislamiento | El dominio no importa paquetes internos del Core |

Las pruebas deben incluir fallos inducidos, reinicios y concurrencia donde el caso de uso lo requiera. Deben documentarse comandos, entorno, resultado y limitaciones.

## 17. Implementación de referencia: Smart-WMS

Smart-WMS es la primera evidencia de cómo aterrizar el patrón:

- El WMS conserva su lenguaje y sus reglas: warehouse, zone, bin, product, receiving, picking e inventory.
- Su adaptador/CoreStore traduce registros de negocio a organismos y episodios.
- Las features de dominio consumen una abstracción de almacenamiento, no deberían manipular CID/RootCID directamente.
- El ledger de inventario representa una invariante del WMS, no una primitiva universal del Core.
- La compatibilidad HTTP permite que los clientes existentes convivan con el backend gemelo.

### 17.1 Lecciones que el contrato convierte en requisitos

1. La idempotencia debe comenzar en el límite de la solicitud y conservar el mismo operation_id durante todos los reintentos.
2. Recepciones, picking y otros flujos multi-recurso deben declarar su atomicidad o recuperación; no basta con una secuencia de escrituras.
3. No se pueden ignorar errores del ledger cuando protegen invariantes de inventario.
4. Un ledger global puede ser correcto para un primer alcance, pero exige medir contención, crecimiento y write amplification antes de escalar.
5. La frontera Core debe ser real: si el Core está embebido, se declara embebido; si es remoto, se consume por contrato público. No se documenta una frontera de servicio que el código no respeta.
6. La aplicación no debe trasladar sus reglas de negocio al Core para obtener persistencia o replicación.

La implementación de Smart-WMS debe volver a auditarse contra esta versión del contrato antes de declararse conforme. Las observaciones anteriores son evidencia de diseño y revisión, no certificación de la rama actual.

## 18. Estructura recomendada para nuevas aplicaciones

La estructura exacta puede variar por lenguaje y arquitectura, pero debe mantener estas fronteras:

```
application/
  domain/
    entities/
    valueobjects/
    rules/
    events/
  application/
    commands/
    queries/
    usecases/
  ports/
    organism_repository.go
    operation_store.go
    authorization.go
  adapters/
    prismatec/
      organism_mapper.go
      episode_mapper.go
      operation_adapter.go
      policy_adapter.go
  transport/
  tests/
    conformance/
```

Los nombres son ilustrativos, no una exigencia de paquetes Go. El dominio no debe importar `core/organism`, `storage/cid` ni otros paquetes internos de PrismaTEC-Core. Esos detalles pertenecen al adaptador.

## 19. Criterio de aceptación de una integración

Consideraré que una aplicación está integrada correctamente cuando:

1. sus reglas y entidades permanecen dentro del dominio;
2. el adaptador traduce sus operaciones mediante contratos públicos;
3. el Core conserva identidad, memoria, procedencia y capacidades técnicas sin absorber el vocabulario vertical;
4. las operaciones críticas son idempotentes, recuperables y autorizadas;
5. la modalidad embebida/remota y el modelo de consistencia están documentados;
6. las pruebas de conformidad pasan y la evidencia se puede reproducir;
7. otra aplicación puede implementar el mismo contrato sin copiar código específico de la primera.

## 20. Estado de esta especificación y próximos pasos

Esta versión establece la frontera arquitectónica y los requisitos de conformidad. No crea por sí sola interfaces Go, endpoints AIP, SDK ni garantías transaccionales nuevas.

Orden recomendado de evolución:

1. Auditar Smart-WMS actualizado y mapear cada sección del contrato a código y pruebas.
2. Definir interfaces públicas mínimas del adaptador, sin exponer internals del Core.
3. Formalizar el sobre de operación: operation_id, actor, target, expected_version, payload y resultado.
4. Implementar pruebas de conformidad reutilizables.
5. Probar una segunda aplicación de dominio para comprobar que el contrato es realmente universal.
6. Solo después estabilizar SDKs y garantías de compatibilidad.

**Decisión rectora:** PrismaTEC-Core debe permitir que cada aplicación construya su propio organismo de dominio sin tener que reconstruir la infraestructura de existencia, identidad, memoria, evolución y distribución; y sin que el Core tenga que convertirse en esa aplicación.
