# Domain Baseline Specifications

Specification type: baseline current-state specifications.

These files define current verified domain ownership and boundaries. Future feature specs must reference these domain specs.

## Domain Specs

- `platform-identity-admin.md`
- `portfolio-master-data.md`
- `land-crops.md`
- `field-operations.md`
- `inventory-stock.md`
- `finance-investor-accounting.md`
- `reporting-integrity.md`
- `ai-insights.md`
- `runtime-delivery.md`

## Rules

- Domain ownership is derived from current code, handlers, usecases, repositories, and migrations.
- Aggregate roots are inferred from current module boundaries and persistence ownership; the code does not declare formal DDD aggregates.
- Missing evidence is marked `UNKNOWN`.
