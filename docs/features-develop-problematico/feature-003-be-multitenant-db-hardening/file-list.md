# file-list.md — feature-003 · be-multitenant-db-hardening

Flist exacta (autoritativa) de `/tmp/flists/be-003.txt`: **4 paths, todos `A` (created)**. No hay archivos compartidos, ni `.go`, ni FE en esta flist.

## Propios (núcleo de la feature)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|------|--------|------|-------------------|------------|--------|--------|-----------|
| `migrations_v4/000224_tenant_security_foundation.up.sql` | A | migration SQL (up) | Fase aditiva: +`tenant_id` nullable, backfill al tenant `default`, índices, FK NOT VALID en 32 tablas; siembra roles/permisos por tenant; crea `tenant_invites`, `security_audit_events`, `auth_session_events` | **whole-file** | Archivo nuevo, autónomo, idempotente (`IF NOT EXISTS`/`ON CONFLICT`) | Medio (orden de versión, backfill al default) | Alta (leído completo, 248 líneas) |
| `migrations_v4/000224_tenant_security_foundation.down.sql` | A | migration SQL (down) | Revierte 224: dropea tablas de seguridad, role_permissions/permissions/roles sembrados, columnas/índices/FK `tenant_id` | **whole-file** | Par del up; reversible sin pérdida de datos de dominio | Bajo | Alta (leído completo) |
| `migrations_v4/000225_tenant_constraints_validation.up.sql` | A | migration SQL (up) | Fase estricta: backfill final + abort si nulls, VALIDATE FK, `SET NOT NULL`, reemplaza unicidad global de `name` por unicidad por tenant (`uq_*_tenant_name[_active]`); salta tabla si hay duplicados activos | **whole-file** | Archivo nuevo, autónomo | Medio-Alto (puede saltar unicidad silenciosamente; asume tenancy en código) | Alta (leído completo) |
| `migrations_v4/000225_tenant_constraints_validation.down.sql` | A | migration SQL (down) | Revierte 225 a fase aditiva (dropea `uq_*`, `tenant_id` vuelve nullable); NO borra datos | **whole-file** | Par del up | Bajo | Alta (leído completo) |

## Compartidos (partial-hunks)
Ninguno. Las 4 son migraciones nuevas e independientes; no tocan `wire/*`, `cmd/*`, `go.mod/go.sum`, `Makefile`, `internal/shared/**` ni ningún archivo que sirva a varias features.

> Nota: hay UN borde de "compartido lógico" (no de archivo): la **siembra de roles/permisos** dentro de 224 se solapa conceptualmente con feature-001 (be-platform-tenancy-refactor). No es un hunk compartido (es un archivo nuevo entero), pero coordinar para evitar doble-seed. Está protegido por `ON CONFLICT DO NOTHING`, así que es idempotente.

## Requeridos por dependencia (NO en mi flist — solo referencia)
| path | feature dueña | por qué importa para 003 |
|------|---------------|--------------------------|
| `migrations_v4/000180_authn_authz_mvp.up.sql` | (base, ya en develop) | Crea `auth_tenants/auth_roles/auth_permissions/auth_role_permissions` y roles base `admin/manager/viewer`. **Prerequisito** de 224. Ya está en develop. |
| `internal/...` tenancy (repos tenant-scoped) | feature-001 | 225 asume que el código ya escribe `tenant_id`. Portar 001 antes/junto. |
| `migrations_v4/000226_customer_actor_master_link.up.sql` | feature-004/009 | Usa `customers(tenant_id, actor_id)`, `actors(tenant_id,...)`. **Depende de** 224. |
| `migrations_v4/000231_consolidate_actor_archived_at.up.sql`, `000234_actor_unique_normalized_name.up.sql` | feature-027/009 | Referencian `tenant_id`. Dependen de 224. |

## Dudosos
Ninguno dentro de la flist. Todo es claramente extraíble como whole-file.

## NO traer todavía (do-not-extract-yet) — fuera de flist
- `migrations_v4/000223_actors_safe_migration.*` (feature de actors) — hueco previo en develop; coordina el ordenamiento pero NO es de esta feature.
- `migrations_v4/000226/227/228` — features 004/009; van junto con 224/225 si se decide portar el bloque 223–228 completo para preservar el orden, pero su contenido NO se extrae aquí.
- Cualquier `internal/admin/repository.go` que consuma `tenant_invites`/`security_audit_events` — feature-018 (data-integrity-admin) o 001, no esta.

## Verificación
- `cat /tmp/flists/be-003.txt` -> 4 líneas, todas `A`.
- Confirmado por `git ls-tree 003a9b8f`: 224/225 **NO existen en develop**; develop salta de 222 a 229/230 (hueco 223–228).
