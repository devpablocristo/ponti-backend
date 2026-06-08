# 0001 Identity Gate — Entity Dedup And Unified Multi-Role Actor

Specification type: baseline current-state observed decision specification.

Status: Observed Implementation Decision.

No formal decision-record approval evidence was verified.

## Observed Implementation Decision: One Identity Per Real Party, Roles As Attribute

A central identity registry (`actors`) holds one row per real party (person/org) per tenant. Roles (`customer`, `provider`, `investor`, `manager`, `contractor`, `biller`, `lessee`) are an attribute stored in `actor_roles`, not a partition. The same party acting as both customer and provider resolves to a single `actors.id` (cross-role unified identity).

The feature is flag-gated by `IdentityGateEnabled()` (env `IDENTITY_GATE`, `sync.Once`, default off); with the flag off the write paths do not resolve against the actor registry.

Evidence:
- `internal/shared/models/base.go` (`IdentityGateEnabled`)
- `migrations_v4/000241_actors_registry.up.sql` (`actors`, `actor_roles` with the 7-role CHECK)
- `internal/actors/usecases/domain/actor.go`

## Observed Implementation Decision: Dedup By Hard Key — TAX_ID, Canonical Legal Name, Alias

Deduplicating keys live in `actor_keys(key_type IN ('TAX_ID','LEGAL_NAME','PERSON_NAME','ALIAS'), key_value, active, source)`. The resolver builds candidate keys in a cascade (TAX_ID first, then name) and dedups against them:
- `TAX_ID` is CUIT/CUIL/DNI normalized to digits only (`NormalizeTaxID`); `TaxIDIsNumeric` requires digits-only.
- Names are canonicalized (`Canonicalize`, ñ-safe) and parsed (`ParseLegalName`) into `LEGAL_NAME` for orgs (core + form, e.g. `acme|SA`) or `PERSON_NAME` for persons.
- `ALIAS` carries alternative names.

Uniqueness is enforced per-tenant by the partial unique index `uq_actor_keys_active(tenant_id, key_type, key_value) WHERE active` — the guarantee lives in the index, not in a trigger, and only active keys are constrained (historical/rotated keys are not). Trigram (`gin_trgm_ops`) indexes back fuzzy search over name keys (migration 241) and over `TAX_ID` (migration 247).

Evidence:
- `migrations_v4/000241_actors_registry.up.sql` (`actor_keys`, `uq_actor_keys_active`, `idx_actor_keys_trgm`)
- `migrations_v4/000247_actor_taxid_trgm.up.sql` (`idx_actor_keys_taxid_trgm`)
- `internal/identity/resolver.go` (`ResolveOrCreateIdentity`, `LookupIdentity`, `candidateKeys`, `TenantFor`)
- `internal/identity/taxid.go` (`NormalizeTaxID`, `TaxIDIsNumeric`)
- `internal/identity/legalname.go` (`ParseLegalName`)
- `internal/identity/canonical.go` (`Canonicalize`)

## Observed Implementation Decision: Identity Is Secondary To The Carrier (actor_id FK, ON DELETE SET NULL)

The 7 carrier tables reference the actor by a nullable FK with `ON DELETE SET NULL`, so the identity does not block hard-delete cascades of the carrier. Which column (and the `actor_roles` rows) carries the role: `customers/investors/managers/providers.actor_id`, `workorders/labors.contractor_actor_id`, `invoices.biller_actor_id`. The CUIT/name is not duplicated onto these tables; it stays in `actor_keys`. Columns are added nullable with `NOT VALID` → `VALIDATE`.

Evidence:
- `migrations_v4/000242_actor_id_fks.up.sql`
- `migrations_v4/000243_actor_tenant_not_null.up.sql`

## Observed Implementation Decision: Strict Create Vs Normal Reuse

`ResolveInput.RejectExisting` selects the create policy. With `reject_existing=true` the use case calls `LookupIdentity` first and returns 409 if a matching identity already exists (by name or tax id) instead of reusing it. Direct creates without the flag follow normal dedup (reuse the existing identity). The HTTP surface lives under the `/actors` group: `POST ""` (ResolveActor), `GET ""` (ListActors with `?status`), `GET /search`, `GET /by-tax-id`, `GET /similar`, `GET /:actor_id`, `PUT /:actor_id` (UpdateActor, name re-key → 409 on collision), `DELETE /:actor_id`, `POST /:actor_id/archive`, `POST /:actor_id/restore`, `PUT /:actor_id/roles`, `PUT /:actor_id/tax-id` (TAX_ID re-key → 409 collision / 400 non-numeric; `actor_id` unchanged).

Evidence:
- `internal/actors/handler.go` (route group)
- `internal/actors/usecases/domain/actor.go` (`RejectExisting`)
- `internal/actors/repository.go` (`if in.RejectExisting`)
- `internal/identity/resolver.go` (`LookupIdentity`)

## UNKNOWN

- Whether this decision was formally approved.
- Decision owner and date.
- Whether actor merge (deduplicating two already-created actors) will be added; it is not implemented.
- Backfill of existing carriers into `actors`/`actor_keys` and production activation of `IDENTITY_GATE`.
