# Feature Inventory Baseline Specification

Specification type: baseline current-state feature catalog.

## Platform, Identity, And Admin

| ID | Feature | Status | Evidence |
|---|---|---|---|
| PL-01 | Health, ping, and version endpoints | Implemented | `cmd/api/http_server.go` |
| PL-02 | API key, JWT, and local dev auth | Implemented | `internal/platform/http/middlewares/gin/*`, `migrations_v4/000180_authn_authz_mvp.up.sql`, `migrations_v4/000201_auth_uuid_migration.up.sql` |
| PL-03 | Admin tenant/user/membership management | Implemented | `internal/admin/handler.go`, `internal/admin/idp/*` |

## Runtime, Migration, And Delivery

| ID | Feature | Status | Evidence |
|---|---|---|---|
| PL-04 | SQL migration system | Implemented | `cmd/migrate/*`, `migrations_v4/*`, `Makefile` |
| PL-05 | CI/CD build, deploy, rollback, and DB reset workflows | Implemented | `.github/workflows/*` |
| PL-06 | OpenAPI generation target | Partially Implemented | `Makefile`; generated OpenAPI output not verified |
| PL-07 | Local dev tooling | Implemented | `docker-compose.yml`, `Dockerfile.dev`, `scripts/*`, `Makefile` |

## Portfolio And Master Data

| ID | Feature | Status | Evidence |
|---|---|---|---|
| PF-01 | Customer lifecycle | Implemented | `internal/customer/handler.go`, `internal/customer/usecases.go`, migrations `000020`, `000080` |
| PF-02 | Campaign lookup | Partially Implemented | `internal/campaign/handler.go`; create/get usecases exist but routes expose list only |
| PF-03 | Project lifecycle, search, and dropdown | Implemented | `internal/project/handler.go`, `internal/project/usecases.go`, `internal/project/repository.go` |
| PF-04 | Project field lookup | Implemented | `internal/project/handler.go`, `internal/project/repository.go` |
| PF-05 | Manager lifecycle | Implemented | `internal/manager/handler.go`, migrations `000020`, `000080` |
| PF-06 | Investor lifecycle | Implemented | `internal/investor/handler.go`, migrations `000070`, `000080` |
| PF-07 | Provider lookup | Implemented | `internal/provider/handler.go`, `internal/provider/repository.go` |
| PF-08 | Business parameters CRUD | Implemented | `internal/business-parameters/handler.go`, `migrations_v4/000010_core_tables.up.sql` |
| PF-09 | Category catalog CRUD | Implemented | `internal/category/handler.go`, migrations `000060`, `000080` |
| PF-10 | Class/type catalog CRUD | Implemented | `internal/class-type/handler.go`, migrations `000060`, `000080` |

## Land And Crops

| ID | Feature | Status | Evidence |
|---|---|---|---|
| LC-01 | Field lifecycle | Implemented | `internal/field/handler.go`, migrations `000030`, `000080` |
| LC-02 | Lease type catalog | Implemented | `internal/lease-type/handler.go`, migrations `000030`, `000080` |
| LC-03 | Lot lifecycle/list/export | Implemented | `internal/lot/handler.go`, `internal/lot/excel-service.go` |
| LC-04 | Lot tons/yield metrics | Implemented | `internal/lot/handler.go`, reporting migrations |
| LC-05 | Crop catalog | Implemented | `internal/crop/handler.go`, migrations `000040`, `000080` |

## Field Operations

| ID | Feature | Status | Evidence |
|---|---|---|---|
| OP-01 | Project labor management | Implemented | `internal/labor/handler.go`, `internal/labor/usecases.go` |
| OP-02 | Labor reporting/export/grouping | Implemented | `internal/labor/handler.go`, `internal/labor/excel-service.go`, `v4_report.labor_*` views |
| OP-03 | Work order lifecycle | Implemented | `internal/work-order/handler.go`, `internal/work-order/usecases.go` |
| OP-04 | Work order metrics/export | Implemented | `internal/work-order/handler.go`, `internal/work-order/excel-service.go`, reporting views |
| OP-05 | Investor split payment status | Implemented | `internal/work-order/usecases.go`, migrations `000190`, `000202` |
| OP-06 | Work order duplicate | Stubbed | `internal/work-order/usecases.go`, `internal/work-order/handler.go` |
| OP-07 | Manual work order drafts | Implemented | `internal/work-order-draft/handler.go`, `internal/work-order-draft/usecases.go` |
| OP-08 | Digital and batch work order drafts | Implemented | `internal/work-order-draft/usecases.go`, migrations `000207`, `000230` |
| OP-09 | Draft PDF data | Implemented | `internal/work-order-draft/handler.go`, `internal/work-order-draft/pdf-data.go` |
| OP-10 | Draft publication | Implemented | `internal/work-order-draft/usecases.go` |

## Inventory And Stock

| ID | Feature | Status | Evidence |
|---|---|---|---|
| IN-01 | Supply catalog lifecycle | Implemented | `internal/supply/handler.go`, `internal/supply/usecases.go` |
| IN-02 | Supply bulk operations/export | Implemented | `internal/supply/handler.go`, `internal/supply/excel-service.go` |
| IN-03 | Pending supplies | Implemented | `internal/supply/usecases.go`, migration `000211` |
| IN-04 | Supply usage count | Implemented | `internal/supply/handler.go`, repository work order count |
| IN-05 | Supply movements | Implemented | `internal/supply/handler.go`, `internal/supply/usecases_movement.go` |
| IN-06 | Stock movement alias/editor | Implemented | `internal/supply/handler.go` |
| IN-07 | Internal stock transfers | Implemented | `internal/supply/usecases_movement.go` |
| IN-08 | Return movements | Implemented | `internal/supply/usecases_movement.go`, migration `000199` |
| IN-09 | Stock summary/period/export | Implemented | `internal/stock/handler.go`, `internal/stock/usecases.go` |
| IN-10 | Real stock count | Implemented | `internal/stock/usecases.go`, migration `000203` |

## Finance And Investor Accounting

| ID | Feature | Status | Evidence |
|---|---|---|---|
| FN-01 | Project dollar values | Implemented | `internal/dollar/handler.go`, `internal/dollar/usecases.go` |
| FN-02 | Crop commercializations | Implemented | `internal/commercialization/handler.go`, `internal/commercialization/usecases.go` |
| FN-03 | Work order invoices per investor | Implemented | `internal/invoice/handler.go`, `internal/invoice/usecase.go`, migration `000204` |
| FN-04 | Investor allocations | Partially Implemented | allocation tables and repository usage verified; dedicated API semantics UNKNOWN |

## Reporting And Data Integrity

| ID | Feature | Status | Evidence |
|---|---|---|---|
| RP-01 | Dashboard | Implemented | `internal/dashboard/handler.go`, `internal/dashboard/repository.go` |
| RP-02 | Field-crop report | Implemented | `internal/report/handler.go`, `internal/report/repository.go` |
| RP-03 | Investor contribution report | Implemented | `internal/report/handler.go`, `internal/report/repository.go` |
| RP-04 | Summary results report | Implemented | `internal/report/usecases.go`, `internal/report/usecases/validators.go` |
| RP-05 | Data integrity cost checks | Implemented | `internal/data-integrity/handler.go`, `internal/data-integrity/usecases.go` |
| RP-06 | Tentative/partial prices | Implemented | `internal/data-integrity/handler.go`, migrations `000197`, `000198` |

## AI And Business Insights

| ID | Feature | Status | Evidence |
|---|---|---|---|
| AI-01 | AI chat proxy | Partially Implemented | `internal/ai/*`; external AI behavior UNKNOWN |
| AI-02 | AI chat streaming proxy | Partially Implemented | `internal/ai/*`; external AI behavior UNKNOWN |
| AI-03 | AI conversations proxy | Partially Implemented | `internal/ai/*`; local persistence UNKNOWN |
| AI-04 | Insight inbox read/resolve | Implemented | `internal/businessinsights/handler.go`, `internal/businessinsights/repository.go` |
| AI-05 | Negative stock insight generation | Implemented | `internal/businessinsights/service.go`, `internal/stock/usecases.go`; Review/Nexus policy internals UNKNOWN |

## Deprecated Features

No feature was verified as `Deprecated`.
