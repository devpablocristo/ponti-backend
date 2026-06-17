# Entity Registry Domain Specification

Specification type: current-state domain specification.

## Purpose

Maintain a single, deduplicated registry of legal and natural persons (actors) that participate in the system as customers, investors, managers, providers, contractors, billers, or lessees. Provides identity resolution, key-based deduplication (CUIT/DNI + legal name), role assignment, and lifecycle management (archive/restore) that cascades to all linked entity tables.

## Boundaries

Owns:
- `actors` table and lifecycle (create, update, archive, restore, hard delete).
- `actor_keys` table: indexed identity keys per actor (TAX_ID, LEGAL_NAME, PERSON_NAME, ALIAS).
- `actor_roles` table: role membership per actor.
- Identity resolution and deduplication logic (`internal/identity/resolver.go`).
- Cascade archive/restore to linked entity tables (customers, investors, managers, providers).
- Backfill tooling for historical data (`cmd/backfill-actors`).

Does not own:
- `customers`, `investors`, `managers`, `providers` table structure — only stamps `actor_id` FK and coordinates lifecycle.
- Text-free actor references in `workorders.contractor_actor_id`, `labors.contractor_actor_id`, `invoices.biller_actor_id` — lifecycle of those records is owned by their respective domains.
- Frontend/BFF identity behavior.

## Owned Entities

- `actors`
- `actor_keys`
- `actor_roles`

## Owned APIs

- `POST   /api/v1/actors` — resolve-or-create (200 reused, 201 created)
- `GET    /api/v1/actors` — list with status filter (active | archived | all, default active)
- `GET    /api/v1/actors/search` — exact + trigram search by name or tax_id
- `GET    /api/v1/actors/similar` — fuzzy name candidates (advisory, for duplicate detection UI)
- `GET    /api/v1/actors/by-tax-id` — lookup by CUIT/DNI
- `GET    /api/v1/actors/:actor_id`
- `PUT    /api/v1/actors/:actor_id`
- `DELETE /api/v1/actors/:actor_id` — hard delete
- `POST   /api/v1/actors/:actor_id/archive` — soft delete with cascade
- `POST   /api/v1/actors/:actor_id/restore` — reactivate with cascade
- `PUT    /api/v1/actors/:actor_id/roles`
- `PUT    /api/v1/actors/:actor_id/tax-id`

## Dependencies On Other Domains

- Platform, Identity, And Admin for auth/tenant context.

## Inbound Dependencies

- Portfolio And Master Data: `customers`, `investors`, `managers`, `providers` have `actor_id` FK populated when `IDENTITY_GATE=true`.
- Field Operations: `workorders.contractor_actor_id`, `labors.contractor_actor_id` stamped via `StampActor` when `IDENTITY_GATE=true`.
- Finance And Investor Accounting: `invoices.biller_actor_id` stamped via `StampActor` when `IDENTITY_GATE=true`.

## Aggregate Roots

- `Actor` (with embedded `Keys` and `Roles`)

## Critical Business Rules

- **Uniqueness**: within a tenant, no two active actors may share the same (key_type, key_value) pair. Enforced by partial unique index `uq_actor_keys_active` on `actor_keys(tenant_id, key_type, key_value) WHERE active`.
- **Resolution cascade**: CUIT takes precedence over legal name. If a CUIT match is found, the name is ignored for deduplication purposes.
- **Archive cascade**: archiving an actor soft-deletes all linked `customers`, `investors`, `managers`, and `providers` rows where `actor_id = id AND deleted_at IS NULL`, within the same transaction. `workorders`, `labors`, and `invoices` are not affected.
- **Restore cascade**: restoring an actor clears `deleted_at` on all linked `customers`, `investors`, `managers`, and `providers` rows where `actor_id = id AND deleted_at IS NOT NULL`, within the same transaction.
- **Restore conflict**: if another active actor already holds one of the archived actor's keys, restore returns 409 CONFLICT. The conflicting key must be resolved manually before restoring.
- **Key deactivation on archive**: all `actor_keys` are set `active = false` on archive, removing the actor from the deduplication pool. A new actor with the same CUIT/name can be created after archiving. When reading an archived actor, `loadActor()` returns all keys regardless of `active` state so that fields like CUIT are visible in edit panels.
- **Identity Gate**: controlled by env var `IDENTITY_GATE=true|false`. When off (default historical behavior), create/update paths in customer/investor/manager/provider/workorder/labor/invoice do NOT resolve actors and do NOT stamp `actor_id`. When on, every write path calls `StampActor` and populates `actor_id`.
- **Backfill**: `cmd/backfill-actors` processes all rows in `customers`, `investors`, `managers`, `providers`, `workorders`, `labors`, `invoices` with a NULL actor FK, creates or reuses the corresponding actor, and stamps the FK. Idempotent; safe to run multiple times.

## Tenant Isolation Requirements

- All actor queries are scoped by `tenant_id` resolved from the request context (OrgID header or fallback to tenant `default`).
- `actor_keys.tenant_id` is always a concrete UUID; NULL tenant_id is not allowed.
- Partial unique index `uq_actor_keys_active` is scoped by `tenant_id`.

## Security Requirements

- Baseline auth applies.
- Reads require `api.read`.
- Mutations require `api.write`.

## Known Limitations

- `actors.status` column (values: active | archived) is not updated by the archive/restore flow; lifecycle state is tracked exclusively via `deleted_at`. The `status` field is currently unused in write paths.
- Cascade restore does not distinguish entities archived independently by a user from entities archived as a side effect of actor archiving. All linked entities with `deleted_at IS NOT NULL` are restored unconditionally.
- Frontend cache must be invalidated after actor archive/restore for downstream selectors (e.g., customer dropdown in project creation) to reflect the updated state without a hard page reload.

## Migrations

- `migrations_v4/000232_add_tenant_id_to_workspace_roots.up.sql` — adds `tenant_id` to `customers`, `campaigns`, `projects`; backfills from tenant `default`.
- `migrations_v4/000233_add_tenant_lifecycle.up.sql` — tenant lifecycle support.
- `migrations_v4/000234_unique_name_per_tenant_roots.up.sql` — unique index `(tenant_id, name)` on `customers` and `campaigns`, scoped per tenant.

## Evidence

- `internal/actors/repository.go`
- `internal/actors/repository_crudar.go`
- `internal/actors/usecases.go`
- `internal/actors/handler.go`
- `internal/actors/handler_crudar.go`
- `internal/identity/resolver.go`
- `internal/shared/models/base.go` — `IdentityGateEnabled()`
- `cmd/backfill-actors/main.go`
