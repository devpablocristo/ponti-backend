**Descripción Unificada Del Sistema De IA Propuesto**

El sistema de IA propuesto no debe evolucionar como una extensión específica de `ponti-ai`, sino como una **plataforma agnóstica de asistentes operativos**, reusable para todo el ecosistema. Ponti debe ser la primera aplicación consumidora de esa plataforma, aportando únicamente su especialización de dominio: contexto agrícola, adapters al backend, capabilities de negocio, políticas propias y prompts del dominio.

La separación fundamental del sistema es esta:

- `core`: mecanismos universales del asistente
- `modules`: componentes reutilizables y enchufables
- `apps/ponti`: consumo y especialización de dominio

La inteligencia estructural del sistema no debe quedar enterrada en Ponti. Debe vivir en una plataforma común capaz de servir a otros productos del ecosistema sin forks del mecanismo central. Ponti no debería definir memoria, decisión, permisos, runs, tool execution ni observabilidad. Ponti debería declarar qué capacidades de negocio tiene, qué tools reales expone, qué entidades usa y qué reglas de negocio aplica.

El sistema buscado es un **Assistant Platform** con una sola voz externa y múltiples capacidades internas. Para el usuario existe un único asistente: una sola interfaz mental, una sola manera de interactuar, una sola política de respuesta, una sola forma de consultar, pedir, delegar, confirmar y actuar. Internamente, el sistema no es monolítico: está compuesto por contexto, memoria, capabilities registradas, tools tipadas, decisión, políticas, ejecución durable, auditoría, observabilidad y evaluación continua. Capacidades como explainability, dashboard, stock, labores, reportes, insights y futuras funciones no deben aparecer como asistentes separados, sino como capacidades internas invocadas por un núcleo común.

El problema que este sistema resuelve es el límite actual de Ponti: hoy existen buenas piezas, pero todavía separadas y con demasiado acoplamiento al caso puntual. Existen `insights`, `insight_chat`, chat general, routing heurístico, memoria parcial en dossier, tools de lectura y algo de auditoría, pero todavía falta formalización real de contexto, memoria durable, permisos, decisiones, ejecución y observabilidad. El sistema propuesto unifica todo eso bajo una arquitectura donde el contexto se arma de forma estructurada, la memoria deja de ser acumulación de chat, las decisiones dejan de vivir en prompts implícitos, las acciones requieren política y grants explícitos, las tareas largas tienen ejecución durable y la plataforma deja de estar hardcodeada a Ponti.

El objetivo final es construir un asistente operativo central del ecosistema capaz de entender consultas transversales del negocio, recuperar contexto útil automáticamente, responder con una voz única, explicar decisiones e insights, planificar pasos para resolver objetivos complejos, usar tools de lectura y de acción, actuar sólo cuando tenga permiso explícito, recordar información útil con evidencia y corrección humana, dejar trazabilidad completa y ser reutilizable por otros productos además de Ponti.

La regla arquitectónica rectora es simple:

- todo lo que sea mecanismo va al `core`
- todo lo que sea reusable y enchufable va a `modules`
- todo lo que sea dominio Ponti va a `apps/ponti`

En `core` viven el runtime del asistente, los contratos de contexto, la memoria base, el motor de decisión, el motor de políticas, la ejecución sync/async, el envelope de respuesta y los contratos de auditoría y observabilidad. En `modules` viven piezas como el capability registry, tool registry, eventing, feature flags, eval framework y retrieval semántico opcional. En Ponti viven el `PontiContextAssembler`, las capabilities agrícolas, los adapters a `ponti-backend`, los prompts del dominio y las policies específicas del negocio Ponti.

El flujo universal del sistema es:

1. el usuario interactúa por un canal
2. la request entra a `Assistant API`
3. `Assistant Core` recibe el turno
4. `Context Assembler` arma un contexto estructurado
5. `Memory Core` recupera memoria relevante
6. `Capability Registry` expone las capacidades disponibles
7. `Decision Engine` decide si responder, investigar, planificar, pedir aclaración, proponer acción o ejecutar
8. si hay acción, `Policy Engine` evalúa permisos y riesgo
9. `Run Engine` ejecuta el flujo, inline o async
10. `Response Synthesizer` produce la respuesta final con una sola voz
11. el sistema registra auditoría, métricas, traces y eventos
12. feedback y outcomes alimentan evaluación y mejora continua

La arquitectura lógica queda así:

```text
Canales / UI / APIs
  -> Assistant API
    -> Assistant Core
      -> Context Assembler
      -> Memory Core
      -> Capability Registry
      -> Tool Registry
      -> Decision Engine
      -> Policy Engine
      -> Run Engine
      -> Response Synthesizer
      -> Audit / Telemetry / Events
      -> Eval Hooks
```

El sistema está pensado para autonomía progresiva. No se busca un agente autónomo sin control, sino un asistente operativo con delegación explícita y trazabilidad. Los modos operativos base son:

- `read`
- `recommend`
- `draft`
- `confirm`
- `act`
- `review`

Eso implica que el asistente primero entiende y consulta, luego recomienda, luego prepara una acción, después pide confirmación si corresponde, sólo ejecuta si tiene grant o consentimiento válido y finalmente reporta qué pasó.

La memoria también debe diseñarse correctamente. No debe ser una transcripción persistida de conversaciones, sino un sistema tipado, con evidencia, scopes, corrección humana, retrieval controlado y separación explícita respecto del aprendizaje del producto. El sistema no debería “aprender solo” en producción mutando comportamiento libremente; debería mejorar por feedback, datasets, evals y análisis de outcomes.

La seguridad tampoco debe vivir en prompts. Debe vivir en la arquitectura: autenticación de servicio y usuario, aislamiento tenant fuerte, validación de contratos, separación entre inputs confiables y no confiables, policy engine para acciones, confirmación humana en operaciones sensibles, degradación explícita y validación estructural de outputs antes de side effects.

El estado final deseado es este:
- un solo asistente visible
- muchas capacidades internas reutilizables
- un core agnóstico del dominio
- Ponti consumiendo el core sin forks estructurales
- memoria útil, corregible y auditable
- acciones seguras bajo grants
- runs durables para tareas largas
- trazabilidad completa
- evaluación continua
- posibilidad real de reutilización en otros productos del ecosistema

En una definición sintética: el sistema de IA propuesto es una **plataforma reusable de asistente operativo**, con una sola voz externa, múltiples capacidades internas declarativas, memoria tipada con evidencia, decisiones gobernadas por políticas, ejecución durable y observabilidad nativa; implementada de forma agnóstica en `core` y `modules`, y consumida por Ponti mediante adapters y especializaciones de dominio.

---

**Tareas Rehechas: pequeñas, comprensibles y secuenciales**

La secuencia correcta es:

1. estabilizar y publicar la base actual
2. fijar la arquitectura de plataforma
3. extraer el runtime reusable
4. formalizar contracts, capabilities y tools
5. formalizar contexto y memoria
6. formalizar decisión, permisos y runs
7. integrar Ponti sobre ese core
8. recién después habilitar acciones reales

Cada tarea debe cerrar con:
- contrato o ADR
- código
- tests
- evidencia observable

---

## **Fase 0. Baseline estable de lo actual**

**T01. Congelar el baseline actual**
Qué hacer:
- snapshot de OpenAPI, schemas, headers, tablas AI y rutas actuales
Cómo auditar:
- existe carpeta/versionado con baseline exportado

**T02. Definir qué comportamiento actual queda soportado**
Qué hacer:
- listar contratos públicos que no deben romperse al empezar la migración
Cómo auditar:
- documento de compatibilidad aprobado

**T03. Reemplazar degradaciones silenciosas por degradación explícita**
Qué hacer:
- clasificar cada fallback actual y definir respuesta estructurada
Cómo auditar:
- tests para `degraded_explicit`

**T04. Publicar esta baseline**
Qué hacer:
- tag/release interna del estado actual estable
Cómo auditar:
- versión publicada y recuperable

---

## **Fase 1. Arquitectura de plataforma**

**T05. Crear ADR `assistant-platform-v1`**
Qué hacer:
- fijar separación `core`, `modules`, `apps/ponti`
Cómo auditar:
- ADR mergeado

**T06. Definir el árbol de paquetes objetivo**
Qué hacer:
- decidir dónde vive cada subsistema
Cómo auditar:
- estructura acordada y documentada

**T07. Definir el envelope universal de respuesta**
Qué hacer:
- schema único para respuestas del asistente
Cómo auditar:
- schema validado por tests

**T08. Definir contratos base del runtime**
Qué hacer:
- `TurnRequest`, `TurnContext`, `TurnResponse`, `RunRecord`
Cómo auditar:
- tipos y tests de serialización

---

## **Fase 2. Núcleo reusable**

**T09. Crear `AssistantCore`**
Qué hacer:
- interfaz única de entrada del runtime
Cómo auditar:
- implementación stub funcional

**T10. Crear `ResponseSynthesizer`**
Qué hacer:
- capa única de salida y voz del asistente
Cómo auditar:
- tests de síntesis con distintos inputs internos

**T11. Crear `ContextBundle` genérico**
Qué hacer:
- contrato reusable de contexto
Cómo auditar:
- fake implementation y tests

**T12. Crear `ContextAssembler` como interfaz**
Qué hacer:
- el core depende de interfaz, no de Ponti
Cómo auditar:
- adapter fake en tests

---

## **Fase 3. Capabilities y tools**

**T13. Definir schema de capability**
Qué hacer:
- contrato declarativo de capabilities
Cómo auditar:
- fixtures válidos e inválidos

**T14. Implementar `CapabilityRegistry`**
Qué hacer:
- registrar y descubrir capabilities
Cómo auditar:
- listado dinámico probado

**T15. Definir schema de tool**
Qué hacer:
- contrato declarativo de tools
Cómo auditar:
- validadores y tests

**T16. Implementar `ToolRegistry`**
Qué hacer:
- registry uniforme para tools
Cómo auditar:
- tests de discoverability y validación

**T17. Inventariar tools actuales de Ponti**
Qué hacer:
- clasificar tools existentes, empezando por `read`
Cómo auditar:
- catálogo Ponti de tools

---

## **Fase 4. Contexto y memoria**

**T18. Definir taxonomía de contexto**
Qué hacer:
- separar turno, memoria corta, memoria durable, permisos, open loops, snapshots
Cómo auditar:
- documento versionado

**T19. Implementar `PontiContextAssembler`**
Qué hacer:
- unir dossier, backend, insights, snapshots, memoria y grants
Cómo auditar:
- tests de ensamblado

**T20. Definir taxonomía de memoria**
Qué hacer:
- tipos mínimos de memoria
Cómo auditar:
- enum/schema versionado

**T21. Crear `memory-core`**
Qué hacer:
- modelos y servicios genéricos de memoria
Cómo auditar:
- package reusable con tests

**T22. Crear storage formal de memoria**
Qué hacer:
- tablas y repositorios
Cómo auditar:
- migraciones y tests de persistencia

**T23. Implementar pipeline de memoria**
Qué hacer:
- `extract -> classify -> validate -> persist`
Cómo auditar:
- tests por etapa

**T24. Implementar retrieval de memoria**
Qué hacer:
- ranking por scope, tipo, recencia, confidence
Cómo auditar:
- tests de ranking y benchmark simple

**T25. Implementar corrección humana**
Qué hacer:
- recordar, olvidar, corregir, confirmar
Cómo auditar:
- API o commands testeados

---

## **Fase 5. Decisión, policy y grants**

**T26. Definir modos de decisión**
Qué hacer:
- `read`, `recommend`, `draft`, `confirm`, `act`, `review`
Cómo auditar:
- tipos compartidos y tests

**T27. Crear catálogo de acciones**
Qué hacer:
- metadata completa de acciones
Cómo auditar:
- catálogo validado

**T28. Implementar `DecisionEngine`**
Qué hacer:
- decidir siguiente estado según intención y contexto
Cómo auditar:
- table tests

**T29. Implementar `PolicyEngine`**
Qué hacer:
- enforcement de permisos y riesgo
Cómo auditar:
- tests de allow/deny

**T30. Definir modelo de grants**
Qué hacer:
- schema de permisos delegables
Cómo auditar:
- contrato versionado

**T31. Crear storage de grants y approvals**
Qué hacer:
- tablas y repositorios
Cómo auditar:
- migraciones y tests

**T32. Implementar evaluación de grants**
Qué hacer:
- toda acción write pasa por grants
Cómo auditar:
- tests negativos y positivos

**T33. Implementar expiración y revocación**
Qué hacer:
- grants temporales y revocables
Cómo auditar:
- tests de revoke/expire

---

## **Fase 6. Runs y observabilidad**

**T34. Definir modelo de run**
Qué hacer:
- `run` y `run_steps`
Cómo auditar:
- contracts versionados

**T35. Implementar `run-engine` sync**
Qué hacer:
- ejecución trazable inline
Cómo auditar:
- tests de lifecycle

**T36. Implementar `run-engine` async**
Qué hacer:
- trabajos largos con retry y resume
Cómo auditar:
- tests de persistencia y reanudación

**T37. Estandarizar IDs de correlación**
Qué hacer:
- `request_id`, `conversation_id`, `run_id`, `tenant_id`, `project_id`, `capability_id`
Cómo auditar:
- logs y traces consistentes

**T38. Crear modelo de eventos**
Qué hacer:
- eventos del agente versionados
Cómo auditar:
- contrato validado

**T39. Instrumentar métricas básicas**
Qué hacer:
- latencia, errores, tool failures, policy denials, tokens/costo
Cómo auditar:
- dashboard mínimo operativo

---

## **Fase 7. Integración Ponti**

**T40. Convertir `insight_chat` en capability Ponti**
Qué hacer:
- explainability ya no como superficie separada interna
Cómo auditar:
- capability registrada y compatible con legado

**T41. Convertir capacidades Ponti restantes**
Qué hacer:
- dashboard, labors, supplies, lots, stock, campaigns, reports
Cómo auditar:
- todas discoverables en registry

**T42. Crear adapters Ponti de tools**
Qué hacer:
- encapsular integración con `ponti-backend`
Cómo auditar:
- core desacoplado de Ponti

**T43. Unificar la surface del asistente**
Qué hacer:
- una sola entrada UX/API
Cómo auditar:
- endpoint/flujo principal único

**T44. Unificar la síntesis de salida**
Qué hacer:
- misma voz y mismo envelope para todas las capabilities
Cómo auditar:
- snapshots homogéneos

**T45. Limpiar legado restante**
Qué hacer:
- reducir naming/documentación residual vieja
Cómo auditar:
- búsqueda residual controlada

---

## **Fase 8. Acciones reales**

**T46. Elegir primeras acciones low-risk**
Qué hacer:
- definir set inicial de acciones Ponti
Cómo auditar:
- catálogo MVP aprobado

**T47. Implementar `draft` de acciones**
Qué hacer:
- ninguna acción write ejecuta directa
Cómo auditar:
- contrato `draft` probado

**T48. Implementar `confirm -> act`**
Qué hacer:
- ejecutar sólo con confirmación o grant
Cómo auditar:
- tests de flujo completo

**T49. Registrar outcomes**
Qué hacer:
- persistir resultado final de acciones
Cómo auditar:
- tabla/eventos con outcome

---

## **Fase 9. Evals y seguridad**

**T50. Crear dataset inicial de evals**
Qué hacer:
- casos reales de Ponti
Cómo auditar:
- corpus versionado

**T51. Crear scorecard**
Qué hacer:
- factualidad, routing, tools, memoria, policy, actions
Cómo auditar:
- reporte reproducible

**T52. Integrar evals al release**
Qué hacer:
- bloquear regresiones
Cómo auditar:
- gate en CI/CD

**T53. Blindar tenant isolation**
Qué hacer:
- enforcement fuerte en tablas AI críticas
Cómo auditar:
- tests cross-tenant

**T54. Separar inputs confiables y no confiables**
Qué hacer:
- fronteras explícitas entre user text, tool outputs y contexto
Cómo auditar:
- contratos y validadores

**T55. Validar outputs antes de side effects**
Qué hacer:
- ninguna acción write sin validación estructural y reglas de negocio
Cómo auditar:
- tests de rechazo

---

**Orden final recomendado**
1. `T01-T04`
2. `T05-T08`
3. `T09-T12`
4. `T13-T17`
5. `T18-T25`
6. `T26-T33`
7. `T34-T39`
8. `T40-T45`
9. `T46-T49`
10. `T50-T55`

**Criterio global de terminado**
Esto queda bien cuando:
- existe un core reusable que puede correr con otro adapter no-Ponti
- Ponti consume el core sin forks estructurales
- el usuario ve un solo asistente
- las capabilities y tools son declarativas y discoverables
- la memoria es tipada, corregible y auditable
- las acciones requieren grants o confirmación
- los runs sobreviven fallos
- la degradación es explícita
- todo queda medido, trazado y evaluado