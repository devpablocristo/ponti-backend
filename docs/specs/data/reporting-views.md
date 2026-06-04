# Reporting Views Baseline Specification

Specification type: baseline current-state data projection specification.

## Verified Reporting Schemas

- `v4_core`
- `v4_ssot`
- `v4_calc`
- `v4_report`

Evidence:
- `migrations_v4/000090_v4_schemas.up.sql`

## Verified Projection Areas

`v4_calc` includes work order metrics, lot base costs/income, field-crop aggregation, dashboard supply/fertilizer totals, and investor contribution calculations.

`v4_report` includes lot metrics/list, labor list/metrics, field-crop reports, summary results, dashboard projections, investor contribution reports, work order list/metrics, and stock consumed by supply.

Evidence:
- `migrations_v4/000120_v4_calc_views.up.sql`
- `migrations_v4/000130_v4_report_views.up.sql`
- later reporting migrations `000140` through `000231`

## Consumers

- Dashboard API.
- Reports API.
- Data Integrity API.
- Lot metrics/list/export.
- Labor metrics/list/export.
- Work order metrics/list/export.

Evidence:
- `internal/dashboard/repository.go`
- `internal/report/repository.go`
- `internal/data-integrity/usecases.go`
- `internal/lot/repository.go`
- `internal/labor/repository.go`
- `internal/work-order/repository.go`

## UNKNOWN

- Whether views are considered contractual public interfaces for external consumers.
- Performance SLAs for reporting views.
- Refresh strategy beyond normal SQL view execution.
