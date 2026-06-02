# dependencies.md — feature-010 projects (BE)

## Resumen de grafo
feature-010 BE es un **consumidor pesado** del actor-system, el lifecycle/archive framework, el tenant-context y el text-canonicalizer, y un **proveedor** para 011, 018 y el FE-010.

## Depende-de

### Fuertes / bloqueantes (sin esto NO compila)
| dep | qué provee | símbolos usados en mis archivos | verificado en develop |
|---|---|---|---|
| **bump `platform/persistence/gorm/go`** (tenancy) | paquete `tenancy` | `tenancy.Scope(ctx, db, "projects"/"fields")` | **MISSING en develop** (`go.mod` no lo lista) |
| **004 shared-text-propername** | `internal/shared/text` | `text.CanonicalizeName(...)`, `sharedtext.CanonicalizeName(...)` | **MISSING en develop** |
| **008 identity-tenant-context** | `internal/shared/authz` + `contextkeys` | `authz.TenantFromContext(ctx)`, `authz.TenantStrictModeEnabled()`, `domainerr.TenantMissing()`, `contextkeys.{Actor,OrgID,Role,Scopes}` (en tests) | **MISSING en develop** |
| **007 actor-system** | `internal/actor` | `actorsync.SyncLegacyActor`, `EnsureCustomerFromActor`, `EnsureLegacyEntityFromActor`, `RefreshProjectActorMirrors`, consts `Legacy{Customers,Managers,Investors}`, `Kind{Organization,Person,Unknown}`, `Role{Cliente,Responsable,Inversor}`, `legacy_actor_map` | **MISSING en develop** |
| **009 crudar-archive-surface** | `internal/shared/lifecycle` | `CreateArchiveBatch`, `RunCascadeArchive`, `RestoreScopedRows`, `ArchiveUpdates`, `RestoreUpdates`, `ApplyCauseScope`, `ActiveRef`, `RequireAllActive`, `ReadRowState`, `ListScopedIDs`, `CauseFrom{Batch,Row}` | **MISSING en develop** |

### Débiles
| dep | nota |
|---|---|
| domain `customer/manager/investor` con `ActorID *int64` | en develop esos structs NO tienen `ActorID`; los aporta 007/011. Mis DTOs/modelos lo referencian. No están en mi flist. |
| `pgconn`/`pgx` y `google/uuid` | ya presentes como deps transitivas en go.mod; bajo riesgo. |

### Inciertas
| dep | duda |
|---|---|
| orden 007 vs 008 vs 009 | actor-sync (007) usa tablas que tenant-context (008) y lifecycle (009) también tocan. Orden sugerido seguro: tenancy-bump → 004 → 008 → 007 → 009 → 010. |

## Bloquea-a

| feature | por qué |
|---|---|
| **011 campaign-dto-projectid** | comparte DTO/domain de project (campaign.project_id). Si 011 se extrae antes, choca con los DTOs de 010. |
| **018 data-integrity-admin** | consume `repository.GetRawAdminCostTotal` (agregado en mi `repository.go`). 018 debe ir DESPUÉS de 010 o coordinar la firma. |
| **FE feature-010** | contrato HTTP: path `DELETE /:id/hard`, campo `actor_id`, 409s. BE-first. |

## Archivos / tipos / config / migraciones / APIs compartidos
- **Tipos compartidos:** `customerdom.Customer`, `managerdom.Manager`, `investordom.Investor` (campo `ActorID`) — propiedad de 007/011.
- **APIs compartidas internas:** `actorsync.*`, `lifecycle.*`, `authz.*`, `text.CanonicalizeName`, `tenancy.Scope`.
- **Migraciones:** tenant_id en todo el grafo, actor_id en customers, tablas actors/legacy_actor_map, columnas archive_* — aportadas por 001/003/007/009. **No** en mi flist.
- **go.mod/go.sum:** bump de `platform/persistence/gorm/go` — fuera de mi flist.
- **API HTTP compartida (cross-repo):** contrato project con FE-010.

## Cross-repo
- FE feature-010 (mismo id). BE-first. El FE debe consumir el nuevo path `/hard`, mandar/leer `actor_id`, y manejar los `409 Conflict`.

## Recomendación de orden
1. bump tenancy plataforma (`platform/persistence/gorm/go`)
2. feature-004 (shared-text)
3. feature-008 (authz / tenant-context)
4. feature-007 (actor-system)
5. feature-009 (lifecycle / archive-surface)
6. **feature-010 BE (este)**
7. feature-011 (campaign-dto-projectid) y feature-018 (data-integrity) — después
8. FE feature-010 — al final del lado FE, una vez mergeado el BE
