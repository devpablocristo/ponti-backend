# Land And Crops Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Model productive land and crop structure: fields, lease types, lots, lot dates, crops, hectares, seasons, and tons/yield inputs.

## Boundaries

Owns:
- Field lifecycle.
- Lease type catalog.
- Lot lifecycle and lot date structure.
- Crop catalog.
- Lot metrics/export surfaces.

Does not own:
- Work order execution.
- Supply consumption.
- Commercialization pricing.
- Reporting And Data Integrity projections derived from lots.

## Owned Entities

- `fields`
- `lease_types`
- `lots`
- `lot_dates`
- `crops`

## Owned APIs

- `/api/v1/fields*`
- `/api/v1/lease-types*`
- `/api/v1/lots*`
- `/api/v1/crops*`

## Dependencies On Other Domains

- Portfolio And Master Data: `projects` and `investors`.
- Finance And Investor Accounting: `field_investors` allocation records.
- Platform, Identity, And Admin auth.

## Inbound Dependencies

- Field Operations depends on fields/lots/crops.
- Finance And Investor Accounting commercializations depend on crops.
- Reporting And Data Integrity read field/lot/crop data.

## Outbound Dependencies

- PostgreSQL.
- XLSX export for lots.

## Aggregate Roots

- `Field`
- `Lot`
- `Crop`
- `LeaseType`

## Critical Business Rules

- Lot name must be non-empty, length 2-255, no consecutive spaces, and contain valid business-name characters.
- Hectares must be greater than 0 and not exceed 10000.
- Tons must be greater than or equal to 0 and not exceed 10000.
- Season must match `YYYY` or `YYYY-YYYY`.
- Field/crop IDs must be positive.
- Field-to-project consistency is validated by shared workspace helpers where used.

## Tenant Isolation Requirements

- Isolation is inherited through `fields.project_id -> projects`.
- Workspace filters use `customer_id`, `project_id`, `campaign_id`, and optional `field_id`.
- Physical `tenant_id` is not verified on land/crop tables.

## Security Requirements

- Baseline auth applies.
- Reads/exports require `api.read`.
- Mutations require `api.write`.

## Evidence

- `internal/field/handler.go`
- `internal/field/usecases.go`
- `internal/lot/handler.go`
- `internal/lot/usecases.go`
- `internal/lot/validations.go`
- `internal/crop/handler.go`
- `internal/lease-type/handler.go`
- `internal/shared/filters/workspace.go`
- `migrations_v4/000030_fields_lots_tables.up.sql`
- `migrations_v4/000040_crops_tables.up.sql`
- `migrations_v4/000080_constraints_fks_indexes.up.sql`
