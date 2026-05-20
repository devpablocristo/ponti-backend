# CRUDAR Lifecycle

CRUDAR in Ponti means:

- Create
- Read
- Update
- Archive
- Restore
- Hard delete

The lifecycle state is physical and database-backed:

```txt
active   = deleted_at IS NULL
archived = deleted_at IS NOT NULL
deleted  = row physically removed
```

`DELETE` HTTP is reserved for hard delete. Archive and restore must use explicit
endpoints such as `POST /v1/{resource}/{id}/archive` and
`POST /v1/{resource}/{id}/restore`.

Hard delete must require the row to be archived first unless an internal-only
case explicitly documents a different policy.

## Design Rules

- Do not create a generic entity table.
- Each business entity keeps its own table.
- Each entity declares its lifecycle capabilities.
- Each parent/child relationship declares archive/restore/delete policy.
- Frontend must not invent lifecycle rules; backend/usecase policy is the source
  of truth.

## Archive Cause

`deleted_at` tells us the state. `archive_batch_id` tells us the cause.

Rows affected by the same archive operation share:

- `archive_batch_id`
- `archive_origin_entity`
- `archive_origin_id`

Restore must use that cause metadata. It must not restore children just because
they belong to the restored parent.

## Legacy DELETE endpoints

The canonical endpoint for hard delete is `DELETE /v1/{resource}/{id}/hard`.
Some resources also expose the bare `DELETE /v1/{resource}/{id}` as a legacy
alias toward hard delete (kept for older clients):

| Resource    | `DELETE /:id` exposed | Alias behavior |
|-------------|-----------------------|----------------|
| customer    | yes                   | hard delete    |
| project     | yes                   | hard delete    |
| lot         | yes                   | hard delete    |
| work-order  | yes                   | hard delete    |
| supply      | yes                   | hard delete    |
| field       | yes                   | hard delete    |
| manager     | yes                   | hard delete    |
| investor    | no                    | —              |
| campaign    | no                    | —              |
| labor       | yes (dual)            | see below      |

New code must call `DELETE /:id/hard` explicitly. The legacy alias exists only
for backwards compatibility and may be removed in future versions.

## Resolved asymmetries (2026-05-20)

Three historical asymmetries between resources were corrected:

- **lot.DELETE `/:id`** previously aliased to archive (soft delete). It now
  aliases to hard delete, matching customer/project/work-order/supply.
- **field.DELETE `/:id`** and **manager.DELETE `/:id`** did not exist; both
  now expose the legacy alias toward hard delete for uniformity.

## Pending asymmetry: labor dual routing

The labor resource exposes two route groups:

- `/v1/projects/:project_id/labors/*` — project-scoped operations (create,
  list, update, project-archived listing, project-delete legacy).
- `/v1/labors/*` — global operations (archive, restore, hard delete, metrics,
  reporting, global-archived listing).

Some endpoints exist in both groups with related but distinct semantics
(archived listing, delete). Frontend currently calls a mix of the two plus
some paths under `/v1/labors/projects/...` that do not exist server-side.

Harmonization is pending a coordinated FE+BE change: pick the canonical
group, migrate FE callers, deprecate the duplicates with a `Deprecation`
response header, and remove them after a deprecation window.

