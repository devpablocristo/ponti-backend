# dependencies.md — feature-019 · be-local-tooling-db-scripts

## Resumen direccional

- **depende-de:** ninguna feature de forma **bloqueante de merge**.
- **bloquea-a:** ninguna feature. Es tooling hoja; nadie compila contra estos scripts.
- Todas las dependencias son **soft** (runtime de DB o compilación de un único script).

## Intra-repo

### Soft (compilación de un solo archivo)
| feature | recurso | quién lo usa | tipo |
| --- | --- | --- | --- |
| 012 ai-companion-integration | paquete `internal/axis` (`NewCompanionClient`, `Config`, `CallContext`, `ChatRequest`) | `scripts/smoke-companion/main.go` | **fuerte para ese archivo** — sin 012 ese `.go` no compila. Por eso es manual-port/condicionado. |

### Soft (runtime de la DB, no compilación)
| feature | recurso | quién lo usa | tipo |
| --- | --- | --- | --- |
| 001 be-platform-tenancy-refactor | `platform/persistence/gorm/go/tenancy`, `domainerr.TenantMissing()`, baja de `authz.MaybeTenantScope`/`TenantScope`/`TenantWhere` | `scripts/lint-tenant-leaks.sh` (grep sobre `internal/`) | débil — el lint corre igual; sin 001 da falsos positivos/negativos pero no rompe |
| 003 be-multitenant-db-hardening | columna `tenant_id`, strict mode | `tenant_isolation_audit.sql`, `multi_tenant_golden_master.sql` | débil — SQL falla en runtime si la columna no existe; no afecta el merge |
| 007 actor-system | tablas `actors`, `actor_roles`, `legacy_actor_map`, función `normalize_actor_name()`, `auth_tenants` | `actors_backfill_sync.sql`, `actors_golden_master.sql` | débil — SQL falla en runtime si no existen; no afecta merge |
| 018 data-integrity-admin | comando `cmd/archive-cleanup` | `scripts/data-audit/README.md` (solo lo documenta, no lo invoca en build) | muy débil — doc únicamente |

### Incierta
- Hunk del `Makefile` `bin-build`/`run` `cmd/`→`cmd/api`: podría depender de un cleanup de
  `cmd/` que vive en otra feature (021/023). Si al partir hunks da conflicto, dejarlo allí.
- `db_migrate_up.sh` (v2→v4): asume `migrations_v4/` (lo provee el set de migraciones del
  refactor; no es esta feature). Cambio trivial de comentario; sin impacto de build.

## Cross-repo

Ninguna. Feature solo-BE. En el repo `web` no hay carpeta ni cambios → "sin cambios FE".
Nota: `Makefile`/`run_ponti_local.sh` mencionan `web/` y `axis/` por path/texto, pero no
modifican archivos de esos repos.

## Archivos / tipos / config / migraciones / APIs compartidos

- **Archivo compartido (partial-hunks):** `Makefile` (raíz). Lo tocan también 001, 020/022, 024,
  y el cleanup de seed/env. Riesgo de arrastre si se extrae entero. Ver file-list.md y extraction-plan.md.
- **Tipos compartidos:** `internal/axis.*` (propiedad de 012) consumidos por `smoke-companion`.
- **Config:** los scripts leen `.env` (DB_*) y variables `SRC_*`/`PONTI_AI_DSN`/`PONTI_WEB_DIR`;
  no definen config nueva del repo.
- **Migraciones:** ninguna compartida (los `.sql` no son migraciones).
- **APIs:** `smoke-companion` consume `POST /v1/chat` del servicio axis (externo), no define endpoints.

## Recomendación de orden

1. (Pre) Tener 012 en `develop` → habilita `smoke-companion/main.go`. Si no, postergar ese único archivo.
2. Mergear feature-019 **independiente** (no espera a 001/003/007/018 para el merge; sí para *usar* el tooling).
3. Idealmente 001/003/007 ya en `develop` para que los `.sql`/lint sean funcionalmente útiles.
4. Coordinar con feature-020 (CI) para enganchar `lint-tenant-leaks.sh` al pipeline, y con 024 para el target `openapi:` del Makefile (que NO se trae aquí).
