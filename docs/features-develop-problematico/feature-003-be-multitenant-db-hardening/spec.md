# spec.md — feature-003 · be-multitenant-db-hardening

## Identidad
- **id:** feature-003
- **slug:** be-multitenant-db-hardening
- **nombre:** Multi-tenant DB hardening
- **tipo:** migration (solo migraciones SQL; sin código Go en este recorte)
- **repo:** Backend Go — ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **existe-en-FE:** NO (Solo-BE). En FE no hay carpeta para esta feature; mencionar en su `cross-repo-map` como "sin cambios FE".
- **existe-en-BE:** SÍ (esta extracción).

## Resumen
Endurece el esquema multi-tenant agregando la columna `tenant_id` a ~32 tablas de dominio, haciéndola obligatoria con FK e índices, y sembrando el modelo de roles/permisos por tenant (`saas_superadmin`, `tenant_owner`, etc.) más tablas de seguridad (`tenant_invites`, `security_audit_events`, `auth_session_events`). Se entrega en dos migraciones diseñadas como fases: **224 = aditiva/backfill reversible** y **225 = validación estricta (NOT NULL + índices únicos por tenant)**.

## Objetivo
Garantizar aislamiento por tenant a nivel de base de datos: que toda fila de dominio pertenezca a un tenant, con integridad referencial (`FK -> auth_tenants`) y unicidad de nombre acotada por tenant. Es el cimiento de datos sobre el que se apoyan las features de tenancy (001) y el resto del stack multi-tenant.

## Problema que resuelve
Hasta ahora las tablas de dominio no tenían `tenant_id`; el aislamiento dependía solo de la capa de aplicación. Sin columna ni constraint, datos de distintos tenants podían convivir sin garantía de DB y los índices de unicidad de `name` eran globales (no por tenant). Esta feature lleva el aislamiento al motor.

## Alcance en este repo (BE)
SOLO 4 archivos (mi flist exacta — ver `file-list.md`):
- `migrations_v4/000224_tenant_security_foundation.up.sql` (248 líneas)
- `migrations_v4/000224_tenant_security_foundation.down.sql`
- `migrations_v4/000225_tenant_constraints_validation.up.sql`
- `migrations_v4/000225_tenant_constraints_validation.down.sql`

Qué hace cada una:

### 224 — tenant_security_foundation (aditiva, reversible)
1. `CREATE EXTENSION IF NOT EXISTS pgcrypto`.
2. Resuelve/crea el tenant `default` en `public.auth_tenants` (lo inserta si no existe).
3. Sobre un array de **32 tablas tenant-owned** (`customers`, `projects`, `campaigns`, `fields`, `lots`, `lot_dates`, `workorders`, `workorder_items`, `workorder_investor_splits`, `workorder_supply_items`, `work_order_drafts`, `project_managers`, `project_investors`, `admin_cost_investors`, `field_investors`, `labors`, `supplies`, `supply_movements`, `stock_movements`, `stocks`, `invoices`, `investors`, `managers`, `providers`, `crops`, `categories`, `class_types`, `lease_types`, `business_parameters`, `crop_commercializations`, `project_dollar_values`):
   - `ADD COLUMN IF NOT EXISTS tenant_id uuid` (nullable).
   - Backfill al tenant `default`.
   - Índices `idx_<t>_tenant_id`, `idx_<t>_tenant_id_id`, `idx_<t>_tenant_name`.
   - `FK <t>_tenant_id_fkey -> auth_tenants(id) ON DELETE RESTRICT` creada **NOT VALID** (no valida filas existentes todavía).
4. Re-deriva `tenant_id` de tablas hijas a partir del padre (`project_managers/project_investors/admin_cost_investors` desde `projects`; `field_investors` desde `fields`).
5. Siembra roles `saas_superadmin`, `tenant_owner`, `tenant_admin`, `tenant_manager`, `tenant_viewer` y permisos por recurso (`customers.read/write/archive`, `projects.*`, `lots.*`, `workorders.*`, `labors.*`, `supplies.*`, `stock.*`, `actors.read/write/archive/merge`, `admin.tenants/users/memberships`, `exports.run`, `imports.run`, `ai.use`) con `ON CONFLICT DO NOTHING`. Mapea permisos a roles (incluye roles base `admin/manager/viewer` que vienen de la migración 180).
6. Crea tablas de seguridad: `tenant_invites`, `security_audit_events`, `auth_session_events` (todas con `IF NOT EXISTS` e índices propios).

### 225 — tenant_constraints_validation (estricta)
Por cada una de las 32 tablas que tenga columna `tenant_id`:
1. Backfill final de nulls al tenant `default`; si quedan nulls -> `RAISE EXCEPTION` (aborta).
2. `VALIDATE CONSTRAINT <t>_tenant_id_fkey`.
3. `ALTER COLUMN tenant_id SET NOT NULL`.
4. Reconstruye índice `(tenant_id, id)`.
5. Si la tabla tiene `name`: **dropea índices/constraints únicos globales sobre `name`** y crea índice único por tenant sobre `lower(btrim(name))` — `uq_<t>_tenant_name_active` (parcial `WHERE deleted_at IS NULL`) si hay `deleted_at`, o `uq_<t>_tenant_name` si no. **Si detecta duplicados activos -> `RAISE NOTICE` y `CONTINUE`** (salta esa tabla, no falla).

`.down.sql`:
- 224 down: dropea las 3 tablas de seguridad, borra role_permissions/permissions/roles sembrados, y revierte columnas/índices/FK `tenant_id` de las 32 tablas.
- 225 down: dropea los índices únicos `uq_*` y revierte `tenant_id` a nullable (vuelve a la fase aditiva 224). NO borra datos.

## Alcance en el otro repo (FE)
Ninguno. Sin cambios FE. La existencia de `tenant_id` en DB es transparente para el FE en este recorte.

## Fuera de alcance
- Código Go que escriba/lea `tenant_id` (eso es feature-001 be-platform-tenancy-refactor; ver dependencies.md). En esta flist NO hay archivos `.go`.
- Las migraciones 226 (`customer_actor_master_link`), 231, 234 (que usan `tenant_id`): pertenecen a otras features (004 / 009 / 027) aunque dependen de la 224.
- Tablas `tenant_invites`/`security_audit_events`/`auth_session_events`: se CREAN aquí pero el handler/repo que las consume (`internal/admin/repository.go` en SOURCE) NO está en esta flist.

## Comportamiento esperado
- Tras 224: las tablas tienen `tenant_id` nullable backfilleado al `default`; el sistema sigue funcionando idéntico (fase no disruptiva).
- Tras 225: `tenant_id NOT NULL` en todas; unicidad de nombre por tenant; FK validadas. Inserts sin tenant fallan a nivel DB.

## Estado en dp~1 (SHA 777e5f6a)
- Las 4 migraciones existen y son coherentes entre sí (224 prepara, 225 endurece).
- **NO están en `develop`** (tip 003a9b8f): develop tiene migraciones hasta 222 y luego 229/230, con un **hueco 223–228**. Las 224/225 encajan numéricamente en ese hueco.
- 224/225 son **prerequisito de migraciones ya presentes en develop e introducidas en SOURCE** que referencian `tenant_id` (226, 231, 234). Ver riesgo de ordenamiento en risks.md.

## Criterios de aceptación
1. `migrate up` aplica 224 luego 225 sin error en una DB que ya tiene la migración 180 (auth) y las tablas de dominio.
2. Tras aplicar, todas las 32 tablas presentes tienen `tenant_id NOT NULL` con FK validada.
3. Existen roles `saas_superadmin/tenant_*` y permisos por recurso.
4. Existen tablas `tenant_invites`, `security_audit_events`, `auth_session_events`.
5. `migrate down` (225 luego 224) vuelve al estado previo sin pérdida de datos de dominio.
6. La numeración no colisiona con las migraciones que develop ya aplicó (229/230) — ver decisión.

## Endpoints/modelos/UI/DB/tests afectados
- **Endpoints:** ninguno en esta flist.
- **Modelos/DTOs Go:** ninguno en esta flist (los structs con `tenant_id` van en feature-001).
- **UI:** ninguna.
- **DB:** 32 tablas alteradas (+ `tenant_id`, índices, FK); 3 tablas nuevas; siembra de roles/permisos. Ver `file-list.md`.
- **Tests:** ninguno en la flist. Validación = `make db-verify` (db-reset + migrate-up + validate + schema-snapshot + schema-diff). Ver validation.md.

## Dependencias
- **Intra-repo:** migración 180 (`authn_authz_mvp`, crea `auth_tenants/auth_roles/auth_permissions/auth_role_permissions` y roles base `admin/manager/viewer`) — **ya en develop**. Tablas de dominio (010, 020, 070, ...) — ya en develop.
- **Feature 001 (be-platform-tenancy-refactor):** dependencia declarada. 225 ("Run only after every tenant-owned repository is tenant-scoped") asume que la capa Go ya escribe `tenant_id`. Sin 001, 225 igual corre pero la app puede insertar filas que ya no validan unicidad por tenant. Fuerte.
- **Cross-repo:** ninguna (Solo-BE).

## Riesgos (resumen — detalle en risks.md)
- **Funcional:** 225 puede saltar la unicidad por tenant si hay nombres duplicados activos (RAISE NOTICE + CONTINUE) -> endurecimiento parcial silencioso.
- **Técnico (ALTO):** colisión/ordenamiento de versiones de migrate. develop ya aplicó 229/230; insertar 224/225 (números menores) NO se auto-aplica con golang-migrate sobre una DB ya en versión 230.
- **Datos:** backfill al tenant `default`; si hay datos de varios tenants reales pre-existentes, todo cae al default.

## DECISIÓN recomendada
**Extraer tal cual, pero ARREGLAR EL ORDENAMIENTO antes de mergear (renumerar) + coordinar con 001.**

Las migraciones son sólidas y bien diseñadas (aditiva + estricta, reversibles). Pero develop tiene un hueco 223–228 y ya aplicó 229/230, lo que rompe el orden monótono que asume golang-migrate. Recomendado:
1. Portar 001 (tenancy refactor del código) ANTES o junto, porque 225 lo asume.
2. Resolver la numeración: idealmente portar el bloque 223–228 completo (junto con sus features 004/009/027) en orden, de modo que 224/225 entren ANTES que 229/230. Si eso no es viable, renumerar 224/225 a números > 230 (p.ej. 231+), pero entonces hay que verificar que 226/231/234 que dependen de `tenant_id` no exijan que 224 vaya antes. Ver extraction-plan.md.
3. NO mergear 225 solo en BE sin tener verificado que no hay duplicados de nombre activos (correr el chequeo de validation.md primero).

Confianza: ALTA en el contenido de las migraciones; MEDIA-ALTA en el diagnóstico de ordenamiento (verificado con `git ls-tree` sobre develop y SOURCE).
