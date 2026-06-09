# Technical Debt Log

This directory tracks technical debt found while fixing production bugs. Entries
are evidence-first notes, not immediate implementation mandates.

## Work Orders: `unit_price` In Aggregated Rows

Status: open.

`unit_price` is a numeric unit price for one component row. In the current report
view it comes from `supplies.price` for supply rows and `labors.price` for labor
rows. It is not a currency discriminator and must not be used to decide between
USD, ARS, or any other currency.

The field is meaningful in a component-level row. It becomes ambiguous when Core
aggregates multiple `v4_report.workorder_list` component rows into one physical
work-order row, because a single order can include several supplies and labor
costs with different unit prices. The current compatibility behavior keeps the
first non-zero supply unit price. That avoids breaking the API, but the value
does not fully describe a multi-component row.

Future options:

- Keep `unit_price` only in component/detail endpoints.
- Return `null` or a display marker such as `Mixto` when a row has multiple
  component prices.
- Split summary listing fields from component detail fields.

## Work Orders: Report View Component Rows

Status: open/managed by aggregation.

`v4_report.workorder_list` emits multiple rows for one physical work order or
draft when the order has supply and labor components. The public list, filter
rows, and export contracts are physical-order lists, so Core must aggregate
those component rows by physical `id` before pagination and mapping.

This behavior is intentional in this branch. A row such as `D-1905555.1` can
have one supply component and one labor component in the view, but it must appear
once in `/work-orders`.

## Digital Multi-Lot Drafts Are Physical Split Rows

Status: accepted limitation.

The system does not currently have a real multi-lot work-order aggregate. A
mobile digital batch with several lots persists one physical draft per lot:
`D-n.1`, `D-n.2`, and so on. The batch input `total_used` is the total entered
for the batch; Core distributes it across the physical lot rows.

This keeps compatibility with the existing schema and list UI. A future model
change should be designed separately instead of inferred from these split rows.

## Migration Ledger Sensitivity

Status: open operational risk.

The labor lookup fix depends on `000232_labor_pending_changes`; the historical
consumption repair depends on `000233_fix_multilot_workorder_consumption`.
Validation must target the active DB and the official migration ledger. Manual
commands against the wrong DB name can create false confidence without changing
the runtime database.

For this local stack, the active DB observed during the fix was
`new_ponti_db_develop_local` with `schema_migrations.version = 233` and
`dirty = false`.
