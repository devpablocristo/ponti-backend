# Customers / Projects Lifecycle

Ponti keeps the domain relationship:

```txt
Actor    = identity master
Customer = operational customer (legacy view of actor with role="cliente")
Project  = operational unit under customer
```

Rules:

- `customers.actor_id` links customer to the identity master.
- `projects.customer_id` remains the project relationship.
- Do not create `actor.projects`.
- Do not move projects directly to actors.

## Customer ↔ Actor relationship

The `customers` table is a **legacy view** of actors with `role="cliente"`. Every customer
created via `POST /customers` is automatically linked to an actor through
`EnsureCustomerFromActor()` (in `internal/actor/master_link.go`), called from
`internal/customer/repository.go:CreateCustomer()` inside the same transaction.

Behavior:

- If `actor_id` is provided in the request: the customer is linked to that existing actor.
- If `actor_id` is NOT provided: an actor is found by normalized name or created with
  `actor_kind="organization"` and `role="cliente"`, then linked.
- The `legacy_actor_map` table maintains a bridge `(tenant_id, source_table='customers',
  source_id) → actor_id` so lookups still work for legacy customers that pre-date the
  sync (where `customers.actor_id` may still be `NULL`).
- The mirror column `projects.customer_actor_id` is automatically refreshed from
  `customers.actor_id` whenever the customer-actor link changes.

**Important caveat for tests**: `actorSyncDisabled()` returns `true` when the GORM
driver is `sqlite` (see `internal/actor/legacy_sync.go:638`). This means the sync logic
is skipped entirely in unit tests using sqlite in-memory. Verifying the
customer-always-has-actor invariant requires integration tests against postgres.

## Tech debt: migrate projects to reference actors directly

The current model maintains two related tables (`customers` + `actors`) for backwards
compatibility. The cleaner long-term design is for `projects` to reference actors
directly via `actor_id NOT NULL` and to deprecate the `customers` table.

**Scope of that migration (when prioritized):**

- Data migration: move `customers.actor_id` → `projects.actor_id` directly (using
  `legacy_actor_map` as fallback for unlinked legacy customers).
- Schema: `ALTER TABLE projects DROP COLUMN customer_id, ADD COLUMN actor_id NOT NULL`.
- BE: refactor `internal/project/` to operate on `actor_id`; delete `internal/customer/`
  package (or keep as a thin read-only view).
- FE: refactor `hooks/useDatabase/projects/` and `CustomerEditor` to reference actor
  instead of customer.
- Tests: update all tests that assume `customer_id`.

Estimated effort: 5-10 days of a senior developer. Out of scope for incremental sprints.

## Archive Customer

Archiving a customer:

- creates an archive batch with root `customers`;
- archives the customer;
- archives only active projects of that customer;
- marks those projects with the same archive batch and origin
  `customers/{customer_id}`;
- runs in one transaction.

## Restore Customer

Restoring a customer:

- restores the customer;
- restores only projects archived by the same customer archive batch;
- does not restore projects archived manually before;
- does not restore projects archived by another cause.

## Archive Project

Archiving a project:

- creates an archive batch with root `projects`;
- archives the project;
- archives project-owned rows according to project policy;
- does not archive the customer.

## Restore Project

Restoring a project:

- requires the customer to be active;
- restores only rows archived by that project archive batch;
- blocks if the customer is archived.

