# Especificación del Dominio: Registro de Entidades

Tipo de especificación: especificación de dominio — estado actual.

## Propósito

Mantener un registro único y deduplicado de personas físicas y jurídicas (actores) que participan en el sistema como clientes, inversores, responsables, proveedores, contratistas, facturadores o arrendatarios. Provee resolución de identidad, deduplicación por clave (CUIT/DNI + nombre legal), asignación de roles y gestión de ciclo de vida (archivar/restaurar) que se propaga en cascada a todas las tablas de entidades vinculadas.

## Alcance

Responsabilidad directa:
- Tabla `actors` y su ciclo de vida (crear, editar, archivar, restaurar, eliminar).
- Tabla `actor_keys`: claves de identidad indexadas por actor (TAX_ID, LEGAL_NAME, PERSON_NAME, ALIAS).
- Tabla `actor_roles`: roles asignados a cada actor.
- Lógica de resolución y deduplicación de identidad (`internal/identity/resolver.go`).
- Cascade de archivar/restaurar hacia las tablas de entidades vinculadas (customers, investors, managers, providers).
- Herramienta de backfill para datos históricos (`cmd/backfill-actors`).

Fuera de su responsabilidad:
- Estructura de las tablas `customers`, `investors`, `managers`, `providers` — solo estampa el FK `actor_id` y coordina el ciclo de vida.
- Referencias de texto libre en `workorders.contractor_actor_id`, `labors.contractor_actor_id`, `invoices.biller_actor_id` — el ciclo de vida de esos registros le corresponde a sus propios dominios.
- Comportamiento de identidad en el frontend/BFF.

## Entidades del Dominio

- `actors`
- `actor_keys`
- `actor_roles`

## APIs

- `POST   /api/v1/actors` — resolver o crear (200 si reutilizó, 201 si creó)
- `GET    /api/v1/actors` — listado con filtro de estado (active | archived | all, por defecto active)
- `GET    /api/v1/actors/search` — búsqueda exacta + trigram por nombre o CUIT
- `GET    /api/v1/actors/similar` — candidatos fuzzy por nombre (orientativo, para detección de duplicados en la UI)
- `GET    /api/v1/actors/by-tax-id` — búsqueda por CUIT/DNI
- `GET    /api/v1/actors/:actor_id`
- `PUT    /api/v1/actors/:actor_id`
- `DELETE /api/v1/actors/:actor_id` — eliminación física
- `POST   /api/v1/actors/:actor_id/archive` — archivar con cascade
- `POST   /api/v1/actors/:actor_id/restore` — restaurar con cascade
- `PUT    /api/v1/actors/:actor_id/roles`
- `PUT    /api/v1/actors/:actor_id/tax-id`

## Dependencias de otros dominios

- Platform, Identity, And Admin: para autenticación y contexto de tenant.

## Quién depende de este dominio

- Portfolio And Master Data: `customers`, `investors`, `managers`, `providers` tienen el FK `actor_id` populado cuando `IDENTITY_GATE=true`.
- Field Operations: `workorders.contractor_actor_id` y `labors.contractor_actor_id` se estampan vía `StampActor` cuando `IDENTITY_GATE=true`.
- Finance And Investor Accounting: `invoices.biller_actor_id` se estampa vía `StampActor` cuando `IDENTITY_GATE=true`.

## Raíces de agregado

- `Actor` (con `Keys` y `Roles` embebidos)

## Reglas de negocio críticas

- **Unicidad**: dentro de un tenant, no pueden existir dos actores activos con el mismo par (key_type, key_value). Se garantiza con el índice único parcial `uq_actor_keys_active` sobre `actor_keys(tenant_id, key_type, key_value) WHERE active`.
- **Cascada de resolución**: el CUIT tiene prioridad sobre el nombre legal. Si se encuentra un match por CUIT, el nombre se ignora para la deduplicación.
- **Cascade al archivar**: archivar un actor hace soft-delete de todos los registros vinculados en `customers`, `investors`, `managers` y `providers` donde `actor_id = id AND deleted_at IS NULL`, dentro de la misma transacción. Las tablas `workorders`, `labors` e `invoices` no se ven afectadas.
- **Cascade al restaurar**: restaurar un actor limpia `deleted_at` en todos los registros vinculados en `customers`, `investors`, `managers` y `providers` donde `actor_id = id AND deleted_at IS NOT NULL`, dentro de la misma transacción.
- **Conflicto al restaurar**: si otro actor activo ya tiene una de las claves del actor archivado, el restore devuelve 409 CONFLICT. La clave en conflicto debe resolverse manualmente antes de poder restaurar.
- **Desactivación de claves al archivar**: todas las `actor_keys` pasan a `active = false` al archivar, sacando al actor del pool de deduplicación. Después de archivar, se puede crear un nuevo actor con el mismo CUIT o nombre. Al leer un actor archivado, `loadActor()` devuelve todas sus claves sin importar el estado `active`, para que campos como el CUIT sean visibles en los paneles de edición.
- **Identity Gate**: controlado por la variable de entorno `IDENTITY_GATE=true|false`. Cuando está desactivado (comportamiento histórico por defecto), los paths de creación/edición de customer/investor/manager/provider/workorder/labor/invoice NO resuelven actores ni estampan `actor_id`. Cuando está activado, cada escritura llama a `StampActor` y popula `actor_id`.
- **Backfill**: `cmd/backfill-actors` procesa todos los registros en `customers`, `investors`, `managers`, `providers`, `workorders`, `labors`, `invoices` con FK de actor en NULL, crea o reutiliza el actor correspondiente, y estampa el FK. Es idempotente; se puede correr más de una vez sin riesgo.

## Aislamiento por tenant

- Todas las consultas de actores están filtradas por `tenant_id` resuelto desde el contexto del request (header OrgID o fallback al tenant `default`).
- `actor_keys.tenant_id` siempre es un UUID concreto; no se permite NULL.
- El índice único parcial `uq_actor_keys_active` está acotado por `tenant_id`.

## Requisitos de seguridad

- Aplica la autenticación base del sistema.
- Las lecturas requieren permiso `api.read`.
- Las modificaciones requieren permiso `api.write`.

## Limitaciones conocidas

- La columna `actors.status` (valores: active | archived) no se actualiza en el flujo de archivar/restaurar; el estado del ciclo de vida se trackea exclusivamente vía `deleted_at`. El campo `status` actualmente no se usa en los paths de escritura.
- El cascade de restaurar no distingue entre entidades archivadas manualmente por el usuario y entidades archivadas como efecto del archivado del actor. Todos los registros vinculados con `deleted_at IS NOT NULL` se restauran incondicionalmente.
- El cache del frontend debe invalidarse después de archivar/restaurar un actor para que los selectores (por ejemplo, el dropdown de clientes en la creación de proyectos) reflejen el estado actualizado sin necesidad de recargar la página manualmente.

## Migraciones relacionadas

- `migrations_v4/000232_add_tenant_id_to_workspace_roots.up.sql` — agrega `tenant_id` a `customers`, `campaigns`, `projects`; backfill desde el tenant `default`.
- `migrations_v4/000233_add_tenant_lifecycle.up.sql` — soporte de ciclo de vida por tenant.
- `migrations_v4/000234_unique_name_per_tenant_roots.up.sql` — índice único `(tenant_id, name)` en `customers` y `campaigns`, acotado por tenant.

## Evidencia de código

- `internal/actors/repository.go`
- `internal/actors/repository_crudar.go`
- `internal/actors/usecases.go`
- `internal/actors/handler.go`
- `internal/actors/handler_crudar.go`
- `internal/identity/resolver.go`
- `internal/shared/models/base.go` — función `IdentityGateEnabled()`
- `cmd/backfill-actors/main.go`
