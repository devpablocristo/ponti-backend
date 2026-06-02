# file-list.md — feature-010 projects (BE)

Fuente: `cat /tmp/flists/be-010.txt`. Diff de verdad: `0972e565..777e5f6a`. SOURCE de extracción: `develop-problematico~1` (777e5f6a).

Leyenda extracción: `whole-file` = traer el archivo entero del SOURCE; `partial-hunks` = sólo algunos hunks (archivo compartido); `manual-port` = adaptar a mano; `do-not-extract-yet` = no traer hasta resolver dependencia.

## Propios (núcleo de la feature)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/project/handler.go` | M | handler HTTP | ruteo, renombre `DeleteProject`→`HardDeleteProject`, ruta `/hard`, `runProjectIDAction`, baja `GetProtected()` | whole-file | el archivo entero es coherente a 777e5f6a; cambios localizados al módulo | medio (cambio de contrato de ruta DELETE; afecta FE) | alta |
| `internal/project/handler/dto/project.go` | M | DTO | `ActorID` en DTOs, `CanonicalizeName`, `domainerr.Validation` | whole-file | depende de `shared/text` (004) y de domain con `ActorID` (007/011) | medio (no compila sin deps) | alta |
| `internal/project/repository.go` | M | repository | tenancy scope, actor-sync, lifecycle cascade, hard-delete-bloqueo, `ensure*ForUpdate`, `assertProjectReferencesActive`, `hydrateLegacyActorIDs`, `GetRawAdminCostTotal` | whole-file | 2075 líneas de diff; reescritura completa; traer entero del SOURCE | alto (depende de actor/lifecycle/authz/tenancy; toca grafo entero) | alta |
| `internal/project/repository/models/project.go` | M | modelo GORM | `TenantID`/`ActorID` en modelos, baja de `unique` en Name | whole-file | requiere `uuid` y columnas DB (migraciones de deps) | medio | alta |
| `internal/project/usecases.go` | M | usecases/port | port renombrado (`GetProjectByNameCustomerAndCampaignID`, `HardDeleteProject`) | whole-file | cambio chico (14 líneas), 1:1 con handler/repo | bajo | alta |

## Compartidos (partial-hunks)

(Ninguno dentro de mi flist.) Los puntos de integración compartidos del repo (`go.mod`, `go.sum`, `wire/*`, `cmd/api/*`) NO están en mi flist y son aportados por las features de plataforma/DI/deps. Ver `dependencies.md`.

> Nota: `repository.go` es internamente "mezclado" (tenancy + dominio + actor-sync + lifecycle), pero como archivo es propio del módulo project; al traerlo entero arrastra esas integraciones — por eso depende de que 004/007/008/009 ya estén en `develop`.

## Requeridos por dependencia (NO están en mi flist — referencia)

| path/paquete | feature dueña | por qué lo necesito |
|---|---|---|
| `internal/actor/**` (`legacy_sync.go`, `master_link.go`, `repository.go`, …) | 007 actor-system | `actorsync.SyncLegacyActor/EnsureCustomerFromActor/EnsureLegacyEntityFromActor/RefreshProjectActorMirrors`, constantes `Legacy*`, `Kind*`, `Role*` |
| `internal/shared/lifecycle/**` (`cascade.go`, `lifecycle.go`, `policy.go`, …) | 009 crudar-archive-surface | `CreateArchiveBatch`, `RunCascadeArchive`, `Restore*`, `ArchiveUpdates`, `ApplyCauseScope`, `ActiveRef`, `RequireAllActive`, `ReadRowState`, `CauseFrom*` |
| `internal/shared/authz/**` | 008 identity-tenant-context | `TenantFromContext`, `TenantStrictModeEnabled` |
| `internal/shared/text/propername.go` | 004 shared-text-propername | `CanonicalizeName` |
| `internal/customer/usecases/domain/customer.go` (+manager/investor) | 007 / 011 | campo `ActorID *int64` en los structs de dominio (el DTO/model de project los referencia) |
| `platform/persistence/gorm/go` (paquete `tenancy`) | bump de deps (NO 021, es bump general) | `tenancy.Scope(ctx, db, "<tabla>")` |

## Dudosos

| path | duda | cómo resolver |
|---|---|---|
| `internal/project/repository.go` (hunk `GetRawAdminCostTotal`) | función agregada aquí pero consumida por `internal/data-integrity` (018). ¿Pertenece a 010 o a 018? | Está físicamente en mi archivo y a 777e5f6a; traerla con el whole-file. Confirmar con el agente de 018 que NO la duplique. Comando: `git grep -n GetRawAdminCostTotal 777e5f6a` |

## NO traer todavía

| qué | motivo |
|---|---|
| Cualquier intento de aplicar estos 9 archivos sobre `develop` sin antes mergear 004/007/008/009 + bump tenancy | No compila: faltan `internal/actor`, `internal/shared/lifecycle`, `internal/shared/authz`, `internal/shared/text`, y `platform/persistence/gorm/go` (verificado: MISSING on develop) |
| `develop-problematico` (tip) como fuente | su tip es restore/vacío; usar SIEMPRE `develop-problematico~1` (777e5f6a) |

## DONE (ya porteado — no re-extraer)
- Ninguno de mis 9 archivos está marcado DONE. `lot-metrics`/`total_tons` y `tentative-prices` (BE #117/#121/#124) tocaron OTROS archivos del módulo (no estos 9); este paquete no los reintroduce.

## Inventario adicional (completitud)

Tests del módulo `project` que faltaban en las tablas de arriba. Todos son `*_test.go` (mismo paquete `package project`), van con la feature y se traen enteros. El integration test es M (se le agregó cobertura de rutas); los otros tres son A (nuevos). Ningún archivo DELETED/R en este flist (los login `*.tsx`/`useClickOutside` son del FE, no del be).

| path | status (A/M/D/R) | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/project/handler_integration_test.go` | M | test (handler/integración) | cubre el nuevo ruteo: añade `actionCall` al `fakeUseCases`, renombra `DeleteProject`→`HardDeleteProject`, baja `GetProtected()` del `fakeMiddlewares`, y agrega `TestProjectActionRoutesCallExplicitUseCases` (rutas `/archive`, `/restore`, `DELETE /hard`) | whole-file | M de 94 líneas; el archivo entero a 777e5f6a es coherente con `handler.go`/`usecases.go` ya porteados (mismo renombre y baja de `GetProtected`); traerlo entero del SOURCE | bajo (sólo test; pero NO compila si el fake no matchea el port `HardDeleteProject`/sin `GetProtected` — debe ir junto con `handler.go`+`usecases.go`) | alta |
| `internal/project/repository_tenant_test.go` | A | test (repository, sqlite in-memory) | base de tenancy: define los helpers compartidos `projectTenantGormEngine`, `projectTenantContext` y `setupProjectTenantDB` (crea el schema sqlite completo) que reutilizan los otros tests; valida scoping por `tenant_id` | whole-file | nuevo (895 líneas); traer entero del SOURCE | bajo como test, pero arrastra deps de plataforma: `platform/security/go/contextkeys` (`Actor/OrgID/Role/Scopes`) y `shared/domain` + `crop/repository/models`; no compila sin el bump de plataforma (008/tenancy) ni sin `repository.go` con tenancy | alta |
| `internal/project/repository_rename_test.go` | A | test (repository, sqlite in-memory) | valida que el editor propaga el rename del customer compartido por ID sin tocar manager/investor; usa `setupProjectTenantDB`/`projectTenantGormEngine` de `repository_tenant_test.go` y `NewRepository` | whole-file | nuevo (478 líneas); traer entero del SOURCE | bajo como test; depende del helper compartido (traer junto a `repository_tenant_test.go`), de `domainerr` (platform) y de los domain structs `customer/manager/investor/campaign` con `ActorID` (007/011) | alta |
| `internal/project/repository_archived_refs_test.go` | A | test (repository, sqlite in-memory) | cubre `assertProjectReferencesActive` (bloqueo cuando una referencia está archivada): siembra `customers/campaigns/managers/investors/fields/lots/crops/actors` con id activo (1) y archivado (99) | whole-file | nuevo (208 líneas); traer entero del SOURCE | bajo como test; depende de `assertProjectReferencesActive` (existe en el `repository.go` ya porteado), de `domainerr` (platform) y de los domain de los módulos referenciados | alta |

> Nota de orden de extracción: los tres tests A son post-tenancy/lifecycle. `repository_tenant_test.go` es el dueño de los helpers `setupProjectTenantDB`/`projectTenantGormEngine`/`projectTenantContext`; `repository_rename_test.go` los consume — traerlos en el mismo lote o el rename no compila. Todos requieren las mismas deps que `repository.go` (004/007/008/009 + bump plataforma/tenancy y `platform/security/go/contextkeys`).
