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

