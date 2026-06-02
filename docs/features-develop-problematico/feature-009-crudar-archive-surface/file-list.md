# file-list.md — feature-009 · CRUDAR archive/restore/hard surface

Rango: `0972e565..777e5f6a`. SOURCE = `develop-problematico~1` (777e5f6a).
Columna `f009` = nº de líneas del hunk relacionadas con la superficie CRUDAR; `noise` = otras features que comparten el archivo.
Leyenda extracción: `partial-hunks` (tomar solo hunks 009 con `git restore -p`), `whole-file` (archivo entero es 009), `do-not-extract-yet` (pertenece a otra feature / build depende de algo previo), `manual-port` (reescribir a mano).

> Nota global: en este repo el `repository.go` con la implementación real de Archive/Restore/HardDelete **NO está en este flist** (lo lleva feature-002). Casi todos los archivos comparten ruido de `platform/*` (feature-001), `GetProtected()` removido (feature-008), `csvexport` (feature-013) y json-tags de dominio (feature-027).

## A) Propios de 009 — partial-hunks (handler + usecases con superficie CRUDAR)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/business-parameters/handler.go | M | handler | rutas archive/restore/hard/archived + runParameterIDAction | partial-hunks | f009=10; ruido platform+GetProtected | medio | alta |
| internal/business-parameters/usecases.go | M | usecase | DeleteParameter→HardDeleteParameter + Archive/Restore | partial-hunks | f009=5 | bajo | alta |
| internal/category/handler.go | M | handler | rutas CRUDAR + runCategoryIDAction | partial-hunks | f009=10 | medio | alta |
| internal/category/usecases.go | M | usecase | rename + archive/restore | partial-hunks | f009=6 | bajo | alta |
| internal/class-type/handler.go | M | handler | rutas CRUDAR + runClassTypeIDAction | partial-hunks | f009=13 | medio | alta |
| internal/class-type/usecases.go | M | usecase | rename + archive/restore | partial-hunks | f009=6 | bajo | alta |
| internal/crop/handler.go | M | handler | rutas CRUDAR | partial-hunks | f009=10 | medio | alta |
| internal/crop/usecases.go | M | usecase | rename + archive/restore | partial-hunks | f009=6 | bajo | alta |
| internal/customer/handler.go | M | handler | HardDeleteCustomer + runCustomerIDAction; baja DELETE /:id legacy | partial-hunks | f009=19; muy representativo | medio | alta |
| internal/customer/usecases.go | M | usecase | DeleteCustomer→HardDeleteCustomer | whole-file (hunk casi puro 009) | f009=7/8 | bajo | alta |
| internal/field/handler.go | M | handler | rutas CRUDAR + runFieldIDAction + ListArchivedFields | partial-hunks | f009=13 | medio | alta |
| internal/field/usecases.go | M | usecase | rename + archive/restore | partial-hunks | f009=6 | bajo | alta |
| internal/investor/handler.go | M | handler | rutas CRUDAR + runInvestorIDAction | partial-hunks | f009=13 | medio | alta |
| internal/investor/usecases.go | M | usecase | rename + archive/restore | partial-hunks | f009=6 | bajo | alta |
| internal/labor/handler.go | M | handler | dual-group CRUDAR + runLaborIDAction + ListArchivedLabors/Global | partial-hunks | f009=16; archivo grande (tot=176) | alto | media |
| internal/labor/usecases.go | M | usecase | DeleteLabor→HardDeleteLabor + archive/restore | partial-hunks | f009=6 | medio | alta |
| internal/lot/handler.go | M | handler | DELETE /:id pasa de archive a hard; +archive/restore/hard/archived + runLotIDAction | partial-hunks | f009=13; MEZCLA con lot-metrics(DONE) y csvexport(013) | alto | media |
| internal/lot/usecases.go | M | usecase | rename + ArchiveLot/RestoreLot/ListArchivedLots | partial-hunks | f009=6; ojo GetMetrics(LotListFilter)=lot-metrics DONE | medio | media |
| internal/manager/handler.go | M | handler | rutas CRUDAR + runManagerIDAction | partial-hunks | f009=13 | medio | alta |
| internal/manager/usecases.go | M | usecase | rename + archive/restore | partial-hunks | f009=6 | bajo | alta |
| internal/provider/handler.go | M | handler | superficie CRUDAR parcial | partial-hunks | f009=10; verificar alcance real | medio | media |
| internal/provider/usecases.go | M | usecase | rename | partial-hunks | f009=6 | bajo | media |
| internal/supply/handler.go | M | handler | supplies + supply-movements + stock-movements CRUDAR; globales /archived | partial-hunks | f009=38; archivo grande (tot=236) | alto | media |
| internal/supply/usecases.go | M | usecase | HardDeleteSupply + ListArchivedSupplies + Archive/Restore | partial-hunks | f009=10 | medio | alta |
| internal/supply/usecases_movement.go | M | usecase | Archive/Restore/HardDelete SupplyMovement + ListArchived | partial-hunks | f009=8 | medio | alta |
| internal/work-order/handler.go | M | handler | rutas CRUDAR + runWorkOrderIDAction + ListArchivedWorkOrders | partial-hunks | f009=15 | medio | alta |
| internal/work-order/usecases.go | M | usecase | DeleteWorkOrderByID→HardDeleteWorkOrder + archive/restore | partial-hunks | f009=6 | bajo | alta |
| internal/work-order-draft/handler.go | M | handler | CRUDAR completo + runWorkOrderDraftIDAction | partial-hunks | f009=10 | medio | alta |
| internal/work-order-draft/usecases.go | M | usecase | rename + archive/restore | partial-hunks | f009=6 | bajo | alta |

## A.1) Tests propios de 009

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/customer/handler_delete_test.go | A | test | spy: rutas llaman al usecase correcto y dan 204 | whole-file | creado por 009 | bajo | alta |
| internal/field/handler_actions_test.go | A | test | idem field | whole-file | creado por 009 | bajo | alta |
| internal/investor/handler_actions_test.go | A | test | idem investor | whole-file | creado por 009 | bajo | alta |
| internal/manager/handler_actions_test.go | A | test | idem manager | whole-file | creado por 009 | bajo | alta |
| internal/lot/handler_actions_test.go | A | test | TestLotIDActionHandlers... archive/restore/hard→204 | whole-file | creado por 009 | bajo | alta |
| internal/lot/repository_crudar_test.go | A | test | cobertura CRUDAR empírica + 409 bloqueado | whole-file | creado por 009 (toca repo→depende de 002) | medio | media |
| internal/class-type/usecases_test.go | A | test | usecases archive/restore/hard | whole-file | creado por 009 | bajo | alta |
| internal/customer/repository_harddelete_test.go | M | test | rename Delete→HardDelete en asserts | partial-hunks | f009=7; toca repo (002) | medio | media |
| internal/supply/usecases_delete_test.go | M | test | ajuste a HardDeleteSupply | partial-hunks | pequeño | bajo | alta |
| internal/supply/repository_delete_test.go | M | test | ajuste naming | partial-hunks | f009=1 | bajo | media |
| internal/supply/repository_movement_delete_test.go | M | test | movement delete/archive | partial-hunks | toca repo (002) | medio | media |
| internal/supply/usecases_movement_test.go | M | test | movement archive/restore | partial-hunks | mezcla platform | medio | media |
| internal/work-order-draft/usecases_test.go | M | test | archive/restore/hard draft | partial-hunks | f009=2 | bajo | alta |
| internal/labor/handler_update_labor_test.go | M | test | ajuste rutas labor | partial-hunks | f009=4; mezcla | medio | media |
| internal/lot/handler_export_test.go | M | test | mezcla export(013)+CRUDAR | partial-hunks | f009=2 | medio | baja |

## B) Compartidos (partial) — archivos que sirven a 009 + otra feature

| path | status | tipo | rol 009 | otras features (ruido) | extracción | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/supply/mocks/mock_repository.go | M | mock | renombra/añade HardDelete*/Archive*/Restore* | mezcla firmas de otras | partial-hunks (regenerar mock mejor) | alto | media |
| internal/supply/repository_movement.go | M | repo | hunks Archive/Restore movement (f009=4) | 560 líneas mayormente 002/013 | do-not-extract-yet (va con 002) | alto | media |
| internal/business-parameters/usecases/domain/business_parameter.go | M | domain | — | json-tags 027 + platform | do-not-extract-yet | bajo | media |

## C) Requeridos por dependencia (NO en mi flist, referenciar)

| path | dueño | por qué importa a 009 |
|---|---|---|
| internal/<dominio>/repository.go (customer, lot, supply, field, manager, work-order, ...) | feature-002 | implementan ArchiveX/RestoreX/HardDeleteX; sin esto 009 no compila |
| internal/shared/models/base.go | feature-002 | `DeletedAt gorm.DeletedAt`, soft-delete |
| internal/shared/handlers/** (RespondNoContent, ParsePaginationParams, RespondError) | feature-002/008 | usados por los helpers runXIDAction |
| docs/crudar-lifecycle.md | feature-002/009 | documenta alias legacy por recurso |
| CRUDAR_PLAN.md | feature-002 | plan maestro (810 líneas, commit 3727a2e3) |

## D) Dudosos (revisar antes de decidir)

| path | duda | comando para resolver |
|---|---|---|
| internal/provider/handler.go | ¿provider es CRUDAR full o solo parcial? | `git -C <repo> diff 0972e565..777e5f6a -- internal/provider/handler.go` |
| internal/lot/usecases.go | separar ArchiveLot/RestoreLot (009) de GetMetrics(LotListFilter) (lot-metrics DONE) | revisar hunks individuales con `restore -p` |
| internal/supply/handler_update_supply_test.go | f009=13 pero archivo de update; ver qué hunks son CRUDAR | `git -C <repo> diff ... -- internal/supply/handler_update_supply_test.go` |

## E) NO traer todavía (ruido de otras features pese a estar en mi flist)

Todos con `f009=0`: estos archivos cambiaron por **otras** features y NO deben entrar en el PR de 009.

| path | feature dueña | qué cambió |
|---|---|---|
| internal/*/repository/models/*.go (business_parameter, category, classtype, crop, customer, dollar, field, investor, invoice, leasetype, lot, lot-table, manager, base de labor/provider/stock, supply, supply_movement, work_order, work_order_*, dashboard mappers) | 001/003/004/007 | tenant_id/TenantID, actor_id/ActorID, unique-constraint, mappers |
| internal/*/usecases/domain/*.go (business_parameter, category, customer, investor, leasetype, manager, provider, entry_type, work_order_list, work_order_metrics) | 027 / lot-metrics(DONE) | json-tags, total_tons |
| internal/dollar/handler.go, internal/dollar/usecases.go, internal/dollar/repository/models/dollar.go | 008 / 017 | solo `GetProtected()` removido |
| internal/commercialization/handler.go, usecases.go, repository/models | 008 | solo `GetProtected()` removido |
| internal/invoice/handler.go, usecase.go, dto, repository/models, repository_test.go | otra | `DeleteInvoice` (clave compuesta, NO CRUDAR) + platform |
| internal/stock/** (handler, dto, base.go, repository_test.go, usecases.go, domain/filter.go) | 001/013 | platform + csvexport + filtros |
| internal/report/** (handler, models, usecases, validators, project_info.go) | otra (projects/reports) | project_info, validators |
| internal/dashboard/** (handler, mappers, repository_aggregate_test.go, usecases_test.go) | 015 (dashboard) | agregados/tests dashboard |
| internal/businessinsights/** (handler, service, service_test) | 001 | solo platform import |
| internal/customer/handler/dto/{requests,responses}.go | 007/027 | actor/dto |
| internal/investor/handler/dto/responses.go, internal/manager/handler/dto/responses.go | 007 | actor |
| internal/supply/handler/dto/** (create/get/update), internal/provider/handler/dto/get_providers.go | 013 | csvexport/get DTOs |
| internal/invoice/handler/dto/post_invoice.go | otra | platform |
| internal/lot/repository/models/{lot,lot-table}.go | lot-metrics(DONE)/001 | total_tons/tenant |
| internal/lot/repository_upsert_test.go, internal/lot/validations.go | 013/001 | csvexport/platform |
| internal/labor/repository/models/{base,labor_item}.go | 001/007 | tenant/actor |
| internal/work-order/{repository_list_test.go, usecases/domain/work_order_list.go, work_order_metrics.go} | lot-metrics/008 | metrics |
| internal/stock/handler_update_real_stock_test.go, internal/supply/handler/dto/create/create_request_test.go | 001/013 | platform/csv |

## Inventario adicional (completitud)

Paths que faltaban en las tablas anteriores (verificados contra el diff `0972e565..777e5f6a`). Tras muestrear cada hunk, **ninguno aporta superficie CRUDAR propia**: son modelos/dominio/DTO que cambian por tenancy (001), actor (007), json-tags/errores platform (027/001) o limpieza puntual. Donde el modelo **también** suelta el `unique` sobre `Name`, ese hunk SÍ es relevante a 009 (el rename-on-archive necesita liberar el índice único), por eso van como `partial-hunks` y se aclara abajo.

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/businessinsights/service.go | M | service | — | do-not-extract-yet | f009=0; solo migración import core→platform (reviewclient→governanceclient) + `domainerr` (001) | bajo | alta |
| internal/businessinsights/service_test.go | M | test | — | do-not-extract-yet | f009=0; idem service.go, renombra reviewclient→governanceclient en stubs (001) | bajo | alta |
| internal/category/repository/models/category.go | M | model | drop `unique` en Name + tenant_id | partial-hunks | mezcla: `TenantID` (001) es ruido, pero quitar `unique` del Name SÍ es 009 (rename-on-archive) | medio | media |
| internal/category/usecases/domain/category.go | M | domain | nuevo `ListFilters{TypeID}` | do-not-extract-yet | f009=0; struct de filtros para ListCategories, no toca CRUDAR (feature de listado) | bajo | media |
| internal/class-type/repository/models/classtype.go | M | model | drop `unique` en Name + remueve TenantID (catálogo global) | partial-hunks | quitar `unique` es 009; el comentario aclara que ClassType es global (sin tenant_id) — fix 001 | medio | media |
| internal/commercialization/repository/models/commercialization.go | M | model | + tenant_id | do-not-extract-yet | f009=0; solo `TenantID` (001), sin CRUDAR (comercialización no tiene archive) | bajo | alta |
| internal/crop/repository/models/crop.go | M | model | drop `uniqueIndex` en Name + tenant_id | partial-hunks | `TenantID` (001) ruido; quitar `uniqueIndex idx_crops_name` SÍ es 009 | medio | media |
| internal/customer/handler/dto/requests.go | M | dto | + ActorID + CanonicalizeName | do-not-extract-yet | f009=0; actor_id en create/update + text.CanonicalizeName (007), no CRUDAR | bajo | alta |
| internal/customer/repository/models/customer.go | M | model | drop `unique` en Name + tenant_id + ActorID | partial-hunks | mezcla 001(tenant)+007(actor); quitar `unique` del Name SÍ es 009 | medio | media |
| internal/customer/usecases/domain/customer.go | M | domain | + ActorID en Customer/ListedCustomer | do-not-extract-yet | f009=0; solo campo ActorID (007); `ArchivedAt` ya existía | bajo | alta |
| internal/dashboard/repository/models/mappers.go | M | mapper | borra investorContributionsToDomain (dead code) | do-not-extract-yet | f009=0; elimina mapper `//nolint:unused` (dashboard/015) | bajo | alta |
| internal/field/repository/models/field.go | M | model | + tenant_id en Field y FieldInvestor | do-not-extract-yet | f009=0; solo `TenantID` (001); Field no cambia constraint de Name aquí | bajo | media |
| internal/investor/repository/models/investor.go | M | model | drop `unique` en Name + tenant_id + ActorID(`-`) | partial-hunks | 001+007 ruido; quitar `unique` del Name SÍ es 009 | medio | media |
| internal/investor/usecases/domain/investor.go | M | domain | + ActorID | do-not-extract-yet | f009=0; solo campo ActorID (007); `ArchivedAt` preexistente | bajo | alta |
| internal/labor/repository/models/labor_item.go | M | model | + TableName() v4_report.labor_list | do-not-extract-yet | f009=0; fija TableName de la vista (fix de query, no CRUDAR) | bajo | alta |
| internal/lease-type/repository/models/leasetype.go | M | model | drop `unique` en Name + tenant_id | partial-hunks | `TenantID` (001) ruido; quitar `unique` del Name SÍ es 009 | medio | media |
| internal/lease-type/usecases/domain/leasetype.go | M | domain | quita json-tags id/name | do-not-extract-yet | f009=0; solo remueve `json:"..."` (027) | bajo | alta |
| internal/lot/repository/models/lot-table.go | M | model | + tenant_id en LotDates | do-not-extract-yet | f009=0; solo `TenantID` (001) + import decimal reordenado | bajo | alta |
| internal/lot/repository/models/lot.go | M | model | + tenant_id en Lot | do-not-extract-yet | f009=0; solo `TenantID` (001); Lot no toca `unique` de Name aquí | bajo | media |
| internal/manager/repository/models/manager.go | M | model | drop `unique` en Name + tenant_id + ActorID(`-`) | partial-hunks | 001+007 ruido; quitar `unique` del Name SÍ es 009 | medio | media |
| internal/manager/usecases/domain/manager.go | M | domain | + ActorID | do-not-extract-yet | f009=0; solo campo ActorID (007); `ArchivedAt` preexistente | bajo | alta |
| internal/provider/usecases/domain/provider.go | M | domain | + ActorID + quita json-tags | do-not-extract-yet | f009=0; ActorID (007) + remueve `json:"..."` (027) | bajo | alta |
| internal/report/repository/models/investor-contribution.go | M | model | fmt.Errorf→domainerr | do-not-extract-yet | f009=0; migración a `platform/errors/domainerr` (001) | bajo | alta |
| internal/report/usecases/validators.go | M | usecase | fmt.Errorf→domainerr.Validation | do-not-extract-yet | f009=0; migración a `platform/errors/domainerr` (001) | bajo | alta |
| internal/stock/handler/dto/update_close_date.go | M | dto | import core→platform domainerr | do-not-extract-yet | f009=0; solo cambia path de import (001) | bajo | alta |
| internal/supply/handler/dto/create/create_supply_movement_request.go | M | dto | import core→platform domainerr | do-not-extract-yet | f009=0; solo cambia path de import (001) | bajo | alta |
| internal/supply/handler/dto/get/get_providers_response.go | M | dto | + ActorID en respuesta provider | do-not-extract-yet | f009=0; expone ActorID (007) | bajo | alta |
| internal/supply/handler/dto/get/get_supply_movement_response.go | M | dto | + ProjectID + rearmado de campos investor/provider/supply | do-not-extract-yet | f009=0; enriquece respuesta de movimientos (013/supply), no CRUDAR | bajo | media |
| internal/supply/handler/dto/update/update_supply_movement_request.go | M | dto | import core→platform domainerr | do-not-extract-yet | f009=0; solo cambia path de import (001) | bajo | alta |
| internal/supply/repository/models/supply.go | M | model | + tenant_id + log "event" key | do-not-extract-yet | f009=0; `TenantID` (001) + tweak de slog.Warn; sin CRUDAR | bajo | alta |
| internal/supply/repository/models/supply_movement.go | M | model | + tenant_id | do-not-extract-yet | f009=0; solo `TenantID` (001) | bajo | alta |
| internal/supply/usecases/domain/entry_type.go | M | domain | borra newline final | do-not-extract-yet | f009=0; cambio cosmético (whitespace) | bajo | alta |
| internal/supply/usecases_movement_import.go | M | usecase | import core→platform + mensajes a inglés | do-not-extract-yet | f009=0; migración import (001) + i18n de mensajes de validación | bajo | alta |
| internal/work-order-draft/repository/models/work_order_draft.go | M | model | + tenant_id | do-not-extract-yet | f009=0; solo `TenantID` (001) | bajo | alta |
| internal/work-order/repository/models/work_order.go | M | model | drop `uniqueIndex` en Number + tenant_id (WorkOrder/Item) | partial-hunks | `TenantID` (001) ruido; quitar `uniqueIndex` de Number SÍ es 009 (rename-on-archive) | medio | media |
| internal/work-order/repository/models/work_order_investor_split.go | M | model | + tenant_id | do-not-extract-yet | f009=0; solo `TenantID` (001) | bajo | alta |
