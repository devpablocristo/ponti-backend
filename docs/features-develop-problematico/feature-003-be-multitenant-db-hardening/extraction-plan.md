# extraction-plan.md — feature-003 · be-multitenant-db-hardening

## Coordenadas
- **repo:** ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base:** `develop` (tip `003a9b8f`)
- **SOURCE de extracción:** `develop-problematico~1` (SHA **777e5f6a**). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **rango fuente-de-verdad (diff):** `0972e565..777e5f6a`
- **rama sugerida:** `pr/feature-003-be-multitenant-db-hardening-be`
- **orden cross-repo:** **Solo-BE** (no hay PR FE). Respecto a otras features BE: **BE-first**, pero DESPUÉS de feature-001 (ver dependencias previas).

## PR title
`feat(db): multi-tenant DB hardening — tenant_id NOT NULL + constraints (224/225)`

## PR description (sugerida)
```
Endurece el esquema multi-tenant en dos migraciones:

- 000224_tenant_security_foundation (aditiva, reversible): agrega tenant_id
  nullable a 32 tablas de dominio, backfill al tenant `default`, índices y FK
  NOT VALID hacia auth_tenants; siembra roles/permisos por tenant
  (saas_superadmin, tenant_owner, tenant_admin, tenant_manager, tenant_viewer)
  y crea tablas tenant_invites, security_audit_events, auth_session_events.
- 000225_tenant_constraints_validation (estricta): backfill final + abort si
  quedan nulls, VALIDATE FK, SET NOT NULL, y reemplaza la unicidad global de
  `name` por unicidad por tenant (uq_*_tenant_name[_active]). Salta la tabla
  (RAISE NOTICE + CONTINUE) si detecta nombres duplicados activos.

Ambas con down reversible (225 vuelve a fase aditiva; 224 revierte todo sin
borrar datos de dominio).

Depende de: feature-001 (tenancy refactor del código) — 225 asume repos
tenant-scoped. Prerequisito ya presente: migración 180 (auth_tenants/roles).

ATENCIÓN ORDENAMIENTO: ver sección "Riesgo de numeración".
```

## ⚠️ Riesgo de numeración (LEER ANTES DE TODO)
- develop ya tiene migraciones **229 y 230** aplicadas (contenido idéntico a SOURCE — verificado con `git diff 003a9b8f 777e5f6a` = sin cambios).
- develop tiene un **hueco 223–228**.
- golang-migrate (`migrate/migrate:v4.17.1`, ver `scripts/db/db_migrate_up.sh`) trackea **una sola versión entera** en `schema_migrations`. Sobre una DB ya en versión 230, un `migrate up` **NO aplicará** archivos 224/225 (números menores). Quedarían "muertos" en disco.
- Además 226/231/234 (otras features) **dependen de `tenant_id`** que crea 224.

### Dos caminos posibles (elegir con el humano)
**Camino A — Portar el bloque 223–228 en orden (RECOMENDADO):** extraer 223,224,225,226,227,228 (cada uno con su feature dueña: 003 aporta 224/225; 004/009/027 aportan el resto) y mergearlos ANTES de que la DB pase de 222. Sólo viable si la DB destino aún no superó 222, o si se hace un reset. Preserva el orden natural.
**Camino B — Renumerar 224/225 a > 230 (p.ej. 235/236):** si la DB destino ya está en 230. Implica renombrar los 4 archivos. PERO entonces 226/231/234 (que dependen de 224) deben renumerarse a algo > esa versión, manteniendo 224<226, 224<231, 224<234. Más frágil; coordinar con dueños de 004/009/027.

> Esta es una decisión de orquestación que excede la flist de 003. Documentado aquí para que el humano no aplique 224/225 a ciegas sobre una DB en versión 230.

## Pasos ordenados
1. **(previo) feature-001:** portar/mergear el tenancy refactor de código, o al menos verificar que va en el mismo tren. 225 lo asume.
2. **(previo) decidir Camino A o B** de numeración (sección anterior).
3. Crear rama desde develop.
4. Traer los 4 archivos enteros desde SOURCE.
5. (Si Camino B) renombrar los 4 archivos a la nueva numeración.
6. `git diff --check` (whitespace) y revisión visual.
7. Validar localmente con `make db-verify` sobre una DB con datos representativos (ver validation.md) — clave: detectar duplicados de nombre activos ANTES de que 225 los salte.
8. PR a develop.

## Archivos enteros vs parciales
- **Enteros (4/4):** las 4 migraciones. No hay extracción parcial.
- **Parciales:** ninguno.

## Migraciones / tests a incluir
- Incluir: `000224_*.up/down.sql`, `000225_*.up/down.sql`.
- Tests: ninguno en la flist. Validación vía `make db-verify` / `db-validate`.

## Dependencias previas
- Migración 180 (auth) — ya en develop. OK.
- feature-001 (código tenancy) — coordinar para que vaya antes/junto.
- Tablas de dominio (010/020/070/...) — ya en develop. OK.

## Coordinación con el otro repo
- **FE:** sin cambios. No hay PR FE asociado. En el `cross-repo-map` del FE marcar "feature-003: sin cambios FE".

## Comandos git SUGERIDOS (para un humano — NO ejecutar desde el agente)
```bash
# 0) partir de develop
git checkout develop
git pull
git checkout -b pr/feature-003-be-multitenant-db-hardening-be

# 1) traer los 4 archivos enteros desde SOURCE (whole-file)
git checkout develop-problematico~1 -- \
  migrations_v4/000224_tenant_security_foundation.up.sql \
  migrations_v4/000224_tenant_security_foundation.down.sql \
  migrations_v4/000225_tenant_constraints_validation.up.sql \
  migrations_v4/000225_tenant_constraints_validation.down.sql

# 2) (SOLO si Camino B) renumerar — ejemplo a 235/236, ajustar según decisión:
#   git mv migrations_v4/000224_tenant_security_foundation.up.sql   migrations_v4/000235_tenant_security_foundation.up.sql
#   git mv migrations_v4/000224_tenant_security_foundation.down.sql migrations_v4/000235_tenant_security_foundation.down.sql
#   git mv migrations_v4/000225_tenant_constraints_validation.up.sql   migrations_v4/000236_tenant_constraints_validation.up.sql
#   git mv migrations_v4/000225_tenant_constraints_validation.down.sql migrations_v4/000236_tenant_constraints_validation.down.sql

# 3) sanity
git diff --check
git status

# 4) revisar contenido vs SOURCE (read-only)
git show develop-problematico~1:migrations_v4/000224_tenant_security_foundation.up.sql | head -160
```
No hay archivos mezclados, así que NO se necesita `git restore -p`. Si por la renumeración hubiera que tocar un índice/registro de migraciones, hacerlo manual (no aplica aquí: golang-migrate descubre por nombre de archivo, no hay índice).

## Qué NO traer
- 223, 226, 227, 228 (otras features) — salvo que se elija Camino A y se coordine con sus dueños.
- Cualquier `.go` (eso es feature-001).
- `develop-problematico` tip (vacío/restore).

## Qué podría romperse
- `migrate up` no aplica 224/225 si la DB ya está en versión 230 (Camino B obligatorio).
- 225 falla con `RAISE EXCEPTION` si quedan `tenant_id` nulls tras backfill (tabla sin relación derivable y sin default aplicado) — improbable porque hace `UPDATE ... WHERE tenant_id IS NULL` antes, pero posible si la tabla no estaba en 224.
- 225 deja unicidad por tenant SIN crear (solo NOTICE) si hay nombres duplicados activos -> integridad más débil de lo esperado, sin error visible.
- Si feature-001 no fue, la app puede seguir insertando con `tenant_id` del default y violar la nueva unicidad por tenant en runtime (error 23505 al usuario).

## Cómo detectar extracción incompleta
- `git ls-tree HEAD migrations_v4/ | grep -E "00022[45]"` debe listar los 4 archivos (o sus equivalentes renumerados).
- `make db-verify` debe correr hasta el final sin abortar.
- Grep de `tenant_id` post-migración: las 32 tablas deben tenerlo NOT NULL.

## Qué validar antes del PR
- Ver validation.md (checklist completa). Mínimo: `make db-verify` verde, sin duplicados de nombre activos, FKs validadas.

## Qué hacer después de mergear
- Confirmar que el PR de feature-001 (código tenancy) está mergeado o en el mismo tren.
- Avisar a dueños de 004/009/027 (226/231/234) que `tenant_id` ya existe en DB.
- Actualizar el snapshot de schema (`make db-schema-snapshot`) si el repo lo versiona.
