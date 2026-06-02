# dependencies.md — feature-003 · be-multitenant-db-hardening

## Resumen de aristas
- **Depende de:** feature-001 (fuerte), migración 180 (fuerte, ya en develop), tablas de dominio base (fuerte, ya en develop).
- **Bloquea a:** feature-004 (226), feature-009 (archive), feature-027 (231/234), y conceptualmente todo el stack multi-tenant que lea/escriba `tenant_id`.

## Depende-de

### Fuertes
| dependencia | tipo | por qué | estado en develop |
|-------------|------|---------|-------------------|
| **feature-001 be-platform-tenancy-refactor** | intra-repo (código) | El header de 225 dice literalmente "Run only after every tenant-owned repository is tenant-scoped and the golden master is green". La capa Go debe escribir `tenant_id` para que la nueva unicidad por tenant tenga sentido y no rompa inserts en runtime. | NO porteado aún (otra feature) |
| **migración 000180_authn_authz_mvp** | intra-repo (DB) | Crea `auth_tenants`, `auth_roles`, `auth_permissions`, `auth_role_permissions` y roles base `admin/manager/viewer`. 224 inserta sobre estas tablas y mapea permisos a esos roles base. | **PRESENTE** (verificado con `git ls-tree 003a9b8f`) |
| **tablas de dominio** (010 core, 020 projects, 070 investors, ...) | intra-repo (DB) | 224 hace `ALTER TABLE` sobre 32 tablas; usa `to_regclass`/`information_schema` para saltar las ausentes, así que es tolerante, pero el valor real depende de que existan. | **PRESENTES** |
| **extensión pgcrypto** | DB | 224 la crea con `IF NOT EXISTS` (para `gen_random_uuid()` de las tablas de seguridad). | Auto-resuelto por la propia migración |

### Débiles
| dependencia | por qué | nota |
|-------------|---------|------|
| Tenant `default` en `auth_tenants` | 224/225 lo crean si no existe (`INSERT ... RETURNING`). | Auto-resuelto. |
| Roles `admin/manager/viewer` | 224 mapea permisos nuevos a ellos; si no existen, simplemente no se mapea (JOIN vacío), sin error. | Vienen de 180, presentes. |

### Inciertas
| dependencia | duda | cómo verificar |
|-------------|------|----------------|
| Columna `name`/`deleted_at` por tabla | 225 ramifica según existan; el comportamiento de unicidad cambia. | `information_schema.columns` por tabla; o leer 225 (ya documentado). |
| Existencia de datos multi-tenant reales pre-224 | Si ya hubiera filas de >1 tenant sin `tenant_id`, el backfill las manda todas al `default`. | Auditar datos antes (ver validation.md). |

## Bloquea-a (esta feature es prerequisito de…)

| feature/migración | tipo | por qué |
|-------------------|------|---------|
| **000226_customer_actor_master_link** (feature-004/009) | DB | Usa `customers(tenant_id, actor_id)`, `actors(tenant_id, normalized_name)`, `c.tenant_id = p.tenant_id`. Requiere que 224 ya agregó `tenant_id`. (Verificado: grep de `tenant_id` en 226.) |
| **000231_consolidate_actor_archived_at** (feature-027) | DB | Referencia `tenant_id` (1 ocurrencia). |
| **000234_actor_unique_normalized_name** (feature-009/027) | DB | Referencia `tenant_id` (7 ocurrencias) — unicidad por tenant de actores. |
| **Stack multi-tenant en código** | App | Cualquier repo Go que filtre por `tenant_id` necesita la columna NOT NULL. |

## Compartidos (archivos/tipos/config/migraciones/APIs)
- **Archivos compartidos:** NINGUNO en la flist (sin `wire/*`, `cmd/*`, `go.mod`, `Makefile`, `internal/shared/**`).
- **Tablas compartidas (lógico):** `auth_roles/auth_permissions/auth_role_permissions` — 224 las siembra, feature-001 también podría tocarlas. Mitigado por `ON CONFLICT DO NOTHING` (idempotente). Coordinar para evitar definiciones divergentes de roles/permisos.
- **Tablas nuevas:** `tenant_invites`, `security_audit_events`, `auth_session_events` — su consumidor (`internal/admin/repository.go` en SOURCE) pertenece a otra feature (001/018); aquí solo se crea el esquema.
- **Config/migraciones compartidas:** el **espacio de numeración de `migrations_v4/`** es el recurso compartido más crítico. 224/225 compiten por orden con 223/226/227/228 (otras features) y con 229/230 (ya en develop). Ver risks.md.

## Cross-repo
- **FE:** ninguna dependencia. Solo-BE. No requiere PR FE antes ni después.

## Recomendación de orden
1. **feature-001** (código tenancy) — antes o en el mismo tren.
2. **feature-003** (esta) — junto con/después de 001.
3. Coordinar el **bloque de migraciones 223–228** para preservar el orden monótono frente a 229/230 ya aplicadas (Camino A de extraction-plan) o renumerar (Camino B).
4. Después: 004 (226), 009/027 (231/234) que dependen de `tenant_id`.

Confianza global: ALTA (dependencias verificadas por grep/ls-tree en SOURCE y develop).
