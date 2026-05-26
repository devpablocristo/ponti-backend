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

## Reference validation pattern (active = referenceable)

Two invariants protect the operational domain end-to-end:

1. **Archived = no existe**. An entity with `deleted_at IS NOT NULL` can not be
   used, selected, referenced, or counted by any non-archived flow.
2. **Hierarchical integrity**. A child can not be active under an archived
   parent. Restoring a child must reject if its parent is still archived.

Both invariants are enforced **inside the create/update/restore transaction**
at the repository layer, not at the handler or UI. The helper that implements
the check is `lifecycle.RequireAllActive` (see
[`internal/shared/lifecycle/lifecycle.go`](../internal/shared/lifecycle/lifecycle.go)).

### Canonical pattern: `assertXReferencesActive`

Each repository that creates or updates an entity with FKs to other
lifecycle-managed entities exposes a private function at the end of the file:

```go
func assertXReferencesActive(tx *gorm.DB, x *domain.X) error {
    if x == nil {
        return nil
    }
    refs := []lifecycle.ActiveRef{
        {Table: "y", Label: "y", ID: x.YID},
        {Table: "z", Label: "z", ID: x.ZID},
    }
    return lifecycle.RequireAllActive(tx, refs)
}
```

The helper short-circuits when `id <= 0` (new rows, optional FKs), runs a
point query per ref, and returns the first `domainerr.Conflict("<label> is
archived")` it finds. `translateBackendError` on the FE maps those messages
to Spanish toasts.

Live implementations (all share the same shape):

| Repository | Function | Validates |
|---|---|---|
| `internal/customer/repository.go` | `assertCustomerReferencesActive` | `actor_id` |
| `internal/project/repository.go` | `assertProjectReferencesActive` | `customer_id`, `customer_actor_id`, `campaign_id`, each `manager.actor_id`, each `investor.actor_id`, each field/lot/crop in the graph |
| `internal/manager/repository.go` | `assertManagerReferencesActive` | `actor_id` |
| `internal/investor/repository.go` | `assertInvestorReferencesActive` | `actor_id` |
| `internal/field/repository.go` | `assertFieldReferencesActive` | `project_id`, `lease_type_id`, nested `crop_id` for each lot |
| `internal/lot/repository.go` | `assertLotReferencesActive` | `field_id`, `previous_crop_id`, `current_crop_id` |
| `internal/labor/repository.go` | `assertLaborReferencesActive` | `project_id`, `category_id` |
| `internal/supply/repository.go` | `assertSupplyReferencesActive` | `project_id`, `category_id`, `type_id` |
| `internal/supply/repository_movement.go` | `assertSupplyMovementReferencesActive` | `project_id`, `project_destination_id`, `supply_id`, `investor_id` + `actor_id`, `provider_id` |
| `internal/work-order/repository.go` | `assertWorkOrderReferencesActive` (pre-existing, the template for the others) | `project_id`, `field_id`, `lot_id`, `crop_id`, `labor_id`, `investor_id`, splits and items |

Every function has a test file `<repo>_archived_refs_test.go` covering: all
refs active (happy path), each ref archived individually (rejects with
correct label), zero IDs treated as no-op, nil input safe.

### Canonical pattern: restore requires active parent

When an entity declares `RestoreRequiresActiveParent: true` in
[`policy.go`](../internal/shared/lifecycle/policy.go), the `RestoreX()`
function must validate every parent is active before clearing the soft-delete
columns. The check uses `lifecycle.RequireActive` and the error message
states the dependency explicitly so the user can resolve the order:

```go
if err := lifecycle.RequireActive(tx, "projects", "project", wo.ProjectID); err != nil {
    return domainerr.Conflict("cannot restore work order while project is archived; restore the project first")
}
if err := lifecycle.RequireActive(tx, "fields", "field", wo.FieldID); err != nil {
    return domainerr.Conflict("cannot restore work order while field is archived; restore the field first")
}
```

Restoring a child does **not** restore its archived parents automatically —
that can revive sibling rows in inconsistent state. The user (or the
explicit cascade-restore of the root parent) must restore them in order.

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

