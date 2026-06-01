# dependencies.md — feature-009 · CRUDAR archive/restore/hard surface

## Resumen de orden

```
002 (crudar-lifecycle-framework)  ──fuerte──▶  009 (esta)  ──fuerte──▶  FE-014 (master-data pages)
        │                                                          └──▶  FE-006 (ArchivedListPage)
001/008/003/007/013/027  ──débil/coexisten en los mismos archivos──  009
```

## Depende de (upstream)

### Fuerte
- **feature-002 · be-crudar-lifecycle-framework** — Aporta:
  - `internal/<dom>/repository.go` con `ArchiveX/RestoreX/HardDeleteX` (firmas que 009 invoca desde usecases). NO están en mi flist.
  - `internal/shared/models/base.go` con `DeletedAt gorm.DeletedAt`, `DeletedBy`.
  - Helpers `internal/shared/handlers/**`: `RespondNoContent`, `RespondError`, `ParsePaginationParams`, `ParseActor`, `ParseProjectIDParam` usados por los `runXIDAction`.
  - `docs/crudar-lifecycle.md` y `CRUDAR_PLAN.md` (plan maestro).
  - **Sin 002, 009 NO compila.**

### Débil (coexisten en los mismos archivos pero no son requisito lógico de 009)
- **feature-001 · be-platform-tenancy-refactor** — swap `core/* → platform/*` (import `ginmw`, `domainerr`). Aparece en casi todos los handler.go. Hay que **rechazar** estos hunks al extraer 009 (a menos que develop ya esté en platform).
- **feature-008 · identity-tenant-context** — quita `GetProtected() []gin.HandlerFunc` del `MiddlewaresEnginePort`. Aparece junto a 009 en todos los handlers tocados, pero es de 008.
- **feature-013 · be-csv-export** — `internal/shared/csvexport`, reemplazo XLSX→CSV en `lot/handler.go` y `supply/handler.go`. Ruido en esos dos.
- **feature-003 · be-multitenant-db-hardening** y **feature-007 · actor-system** — `tenant_id/TenantID`, `actor_id/ActorID` en `repository/models/*.go`. Esos models NO son de 009.
- **feature-027 · be-cleanup-domain-purity** — json-tags en `usecases/domain/*`. NO de 009.
- **lot-metrics (DONE, #117/#121/#124)** — `GetMetrics(LotListFilter)`, `total_tons` en `lot/usecases.go` y `lot/repository/models/*`. Excluir al extraer lot.

### Inciertas
- `internal/supply/mocks/mock_repository.go`: cambia por 009 (HardDelete*/Archive*) pero también por otras firmas; confianza media. Mejor regenerar mock que portar hunks.
- `internal/provider/handler.go`: f009=10 sugiere superficie CRUDAR, pero hay que confirmar si provider expone ciclo completo (confianza media).

## Bloquea a (downstream)

### Fuerte
- **FE-014 · fe-master-data-pages** — las pages de customers/fields/lots/workorders/etc. consumen `POST /:id/archive`, `DELETE /:id/hard`, `GET /archived`. Si 009 no está, esos endpoints no existen.
- **FE-006 · fe-design-system (ArchivedListPage)** — la lista de archivados depende de `GET /archived`.

### Débil
- Cualquier feature FE que dispare borrado desde una tabla (table-select-filters FE #104 ya DONE asume el contrato de borrado).

## Recursos compartidos (archivos / tipos / config / APIs)

| recurso | compartido con | nota |
|---|---|---|
| `internal/<dom>/handler.go` | 001, 008, 013, 009 | extracción por hunks obligatoria |
| `internal/<dom>/usecases.go` | 001, 027, 009 | mayormente limpio para 009 |
| `internal/<dom>/repository.go` (no en flist) | 002, 003 | implementación CRUDAR real |
| `internal/shared/models/base.go` (no en flist) | 002 | soft-delete |
| `internal/shared/handlers/**` (no en flist) | 002, 008 | Respond*, Parse* |
| `internal/supply/repository_movement.go` | 002, 013, 009 | 560 líneas; mayoría no-009 |
| Contrato HTTP `/archive`,`/restore`,`/hard`,`/archived` | FE-014, FE-006 | API compartida cross-repo |

## Migraciones compartidas

Ninguna propia de 009. El esquema soft-delete (columna `deleted_at`, `archive_batch_id`) lo introducen las migraciones de **feature-002**. 009 no añade ni altera columnas.

## Recomendación de orden (final)

1. **002** en develop (prerequisito de compilación).
2. (Idealmente) **001** en develop si se decide quedarse en `platform/*`; si no, rechazar esos hunks.
3. **009** por entidad (este paquete), BE-first.
4. **FE-014** y **FE-006** después.
