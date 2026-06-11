# Actors And Registry Data Model Baseline Specification

Specification type: baseline current-state data specification.

The actors model is the Identity Gate (Pilar 3) central identity registry: one row per real party (person/company) per tenant, with roles as an attribute and uniqueness guaranteed by a partial unique index (not a trigger). It is additive and does not alter existing tables.

## Entities

### `actors`

One row per real party per tenant.

| Column | Type | Notes |
|---|---|---|
| `id` | `bigserial` | PRIMARY KEY |
| `tenant_id` | `uuid NOT NULL` | NOT NULL since migration 243; filled by the resolver with a concrete tenant (ctx OrgID or `default`) |
| `party_type` | `text NOT NULL DEFAULT 'unknown'` | CHECK IN (`org`, `person`, `unknown`) |
| `display_name` | `text NOT NULL` | presentation form |
| `raw_name` | `text NOT NULL` | as entered |
| `status` | `text NOT NULL DEFAULT 'active'` | CHECK IN (`active`, `archived`) |
| `created_at` | `timestamptz NOT NULL DEFAULT now()` | |
| `updated_at` | `timestamptz NOT NULL DEFAULT now()` | |
| `deleted_at` | `timestamptz NULL` | |
| `created_by` | `text NULL` | |
| `updated_by` | `text NULL` | |

Index: `idx_actors_tenant (tenant_id)`.

### `actor_roles`

Roles a party plays, modeled as an attribute (not a partition), so the same party can be customer and provider at once (cross-role unification).

| Column | Type | Notes |
|---|---|---|
| `actor_id` | `bigint NOT NULL` | REFERENCES `actors(id)` ON DELETE CASCADE |
| `role` | `text NOT NULL` | CHECK IN (`customer`, `provider`, `investor`, `manager`, `contractor`, `biller`, `lessee`) |
| `created_at` | `timestamptz NOT NULL DEFAULT now()` | |

PRIMARY KEY `(actor_id, role)`.

### `actor_keys`

Deduplicating keys. Uniqueness lives here, not in a trigger.

| Column | Type | Notes |
|---|---|---|
| `id` | `bigserial` | PRIMARY KEY |
| `actor_id` | `bigint NOT NULL` | REFERENCES `actors(id)` ON DELETE CASCADE |
| `tenant_id` | `uuid NOT NULL` | NOT NULL since migration 243 |
| `key_type` | `text NOT NULL` | CHECK IN (`TAX_ID`, `LEGAL_NAME`, `PERSON_NAME`, `ALIAS`) |
| `key_value` | `text NOT NULL` | |
| `active` | `boolean NOT NULL DEFAULT true` | |
| `source` | `text NOT NULL DEFAULT 'direct'` | CHECK IN (`direct`, `import`, `backfill`) |
| `created_at` | `timestamptz NOT NULL DEFAULT now()` | |

Indexes:

- `uq_actor_keys_active` — `UNIQUE (tenant_id, key_type, key_value) WHERE active`. The dedup guarantee: per-tenant, per-key, role-agnostic. Only active keys participate, so superseded history does not need cleanup.
- `idx_actor_keys_actor (actor_id)`.
- `idx_actor_keys_trgm` — `gin (key_value gin_trgm_ops) WHERE key_type IN ('LEGAL_NAME', 'PERSON_NAME', 'ALIAS')`. Trigram search over name keys (migration 241).
- `idx_actor_keys_taxid_trgm` — `gin (key_value gin_trgm_ops) WHERE key_type = 'TAX_ID'`. Trigram search over tax-id keys for partial CUIT/CUIL lookup (migration 247).

Evidence:
- `migrations_v4/000241_actors_registry.up.sql`
- `migrations_v4/000243_actor_tenant_not_null.up.sql`
- `migrations_v4/000247_actor_taxid_trgm.up.sql`

## Identity FKs On Carrier Tables

The seven identity carriers gained a nullable `actor_id` (`ON DELETE SET NULL`, `NOT VALID` then `VALIDATE`). The role is given by which column references the actor plus `actor_roles`; the party's tax id / name lives in `actor_keys`, not duplicated on these tables.

| Carrier table | Column |
|---|---|
| `customers` | `actor_id` |
| `investors` | `actor_id` |
| `managers` | `actor_id` |
| `providers` | `actor_id` |
| `workorders` | `contractor_actor_id` |
| `labors` | `contractor_actor_id` |
| `invoices` | `biller_actor_id` |

Evidence:
- `migrations_v4/000242_actor_id_fks.up.sql`

## Uniqueness And Dedup Rules

- TAX_ID key = CUIT/CUIL/DNI normalized to digits only; one `TAX_ID` key per party. `TaxIDIsNumeric` rejects non-numeric tax ids.
- Name key = `LEGAL_NAME` for organizations (canonical form `core|FORMA`, e.g. `acme|SA`) or `PERSON_NAME` for persons; `ALIAS` carries alternative names.
- Uniqueness is per-tenant via `uq_actor_keys_active`: the same CUIT or normalized name resolves to a single identity regardless of role (a customer and a provider sharing one CUIT are one identity).
- Resolution cascade (`candidateKeys`): TAX_ID first, then the parsed legal/person name. `Canonicalize` is `ñ`-safe; `ParseLegalName` derives the name key type and value.

Evidence:
- `internal/identity/resolver.go` (`ResolveOrCreateIdentity`, `LookupIdentity`, `candidateKeys`, `TenantFor`)
- `internal/identity/taxid.go` (`NormalizeTaxID`, `TaxIDIsNumeric`)
- `internal/identity/legalname.go` (`ParseLegalName`)
- `internal/identity/canonical.go` (`Canonicalize`)

## Strict-Create Vs Reuse

`ResolveInput.RejectExisting`:

- `reject_existing = true` (strict create): `LookupIdentity` runs first; if a matching name or CUIT already exists, the request is rejected (409) instead of reusing.
- without the flag (comboboxes / direct creates): normal dedup reuses the existing identity.

Evidence:
- `internal/identity/resolver.go`

## Catalog Name Dedup (Track B)

Non-actor catalogs use the `prevent_duplicate_name` trigger (migration 240 pattern, escape hatch `app.bypass_name_dedup = 'on'`):

- `crops`, `types`, `lease_types`: per-tenant unique upgraded from exact to normalized name (`trg_prevent_dup_name` -> `prevent_duplicate_name`).
- `categories`: normalized-name dedup within `(tenant, type_id)` via dedicated `prevent_duplicate_category_name`, because names legitimately repeat across `type_id`.
- `business_parameters`: excluded — deduped by `key` identifier; normalizing the name would collapse distinct keys.

Evidence:
- `migrations_v4/000244_catalog_name_dedup.up.sql`

## Actors And Registry Endpoints

Actors module (`internal/actors`), group `/actors`, core base `/api/v1`:

- `POST ""` `ResolveActor` — body `{name, tax_id?, role, allow_create, reject_existing}`.
- `GET ""` `ListActors` — `?status=active|archived|all&page&per_page`.
- `GET /search` `SearchActors` — `?q&field=name|tax_id&limit`.
- `GET /by-tax-id` `GetByTaxID` — `?tax_id`.
- `GET /similar` `SimilarActors` — `?name&limit`.
- `GET /:actor_id` `GetActor`.
- `PUT /:actor_id` `UpdateActor` — `{display_name, party_type}`; rotates the name key (409 if the name is used by another identity).
- `DELETE /:actor_id` `DeleteActor` — hard delete.
- `POST /:actor_id/archive` `ArchiveActor`.
- `POST /:actor_id/restore` `RestoreActor`.
- `PUT /:actor_id/roles` `SetActorRoles` — `{roles:[]}`.
- `PUT /:actor_id/tax-id` `SetActorTaxID` — `{tax_id}`; re-key (rotates TAX_ID, 409 if it collides, 400 if non-numeric, `actor_id` unchanged).

Registry module (`internal/registry`, read-only search + alias), group `/registry`:

- `GET ""` `Search` — `?q&type&status&page&per_page`; returns `{data:[{entity_type,id,name,tax,roles,archived}], page_info{...}}` by UNION (actors once + 4 catalogs); `type` in `{all, customer, provider, investor, manager, contractor, biller, lessee, crops, types, lease-types, campaigns}`; tenant-scoped flag-gated.
- `PUT /actors/:actor_id/aliases` `SetActorAliases` — `{aliases:[]}`; rotates ALIAS keys, 409 on collision.

Evidence:
- `internal/actors/handler.go`
- `internal/registry/handler.go`

## Feature Flags

- `IdentityGateEnabled()` gates the Identity Gate (Pilar 3).
- `TenantEnforcementEnabled()` gates physical tenant_id filtering (Pilar 1).

Both are env-driven via `sync.Once`.

Evidence:
- `internal/shared/models/base.go`

## Migrations 241–247 Summary

| Migration | What it does |
|---|---|
| `000241_actors_registry` | Creates `actors`, `actor_roles`, `actor_keys`; adds `uq_actor_keys_active` partial unique index and the name-keys trigram index. |
| `000242_actor_id_fks` | Adds nullable `*_actor_id` columns on the 7 carrier tables with FKs `ON DELETE SET NULL` (`NOT VALID` then `VALIDATE`) plus per-column indexes. |
| `000243_actor_tenant_not_null` | Defensive backfill of `tenant_id` to `default`, then `SET NOT NULL` on `actors.tenant_id` and `actor_keys.tenant_id`. |
| `000244_catalog_name_dedup` | Name-dedup triggers on `crops`/`types`/`lease_types` (normalized, per-tenant) and `categories` (per `(tenant, type_id)`); `business_parameters` excluded. |
| `000245_tenant_id_default_not_null` | Defensive backfill, `SET DEFAULT 'default'`, and `SET NOT NULL` on `tenant_id` for 11 tables: `customers`, `campaigns`, `projects`, `managers`, `investors`, `providers`, `crops`, `types`, `lease_types`, `business_parameters`, `categories`. |
| `000246_fine_permissions` | Seeds fine permissions `users:manage` (-> `admin`) and `invites:write` (-> `admin`, `tenant_owner`); idempotent. |
| `000247_actor_taxid_trgm` | Adds the trigram index on `actor_keys` where `key_type = 'TAX_ID'` for partial CUIT/CUIL search. |

Evidence:
- `migrations_v4/000241_actors_registry.up.sql`
- `migrations_v4/000242_actor_id_fks.up.sql`
- `migrations_v4/000243_actor_tenant_not_null.up.sql`
- `migrations_v4/000244_catalog_name_dedup.up.sql`
- `migrations_v4/000245_tenant_id_default_not_null.up.sql`
- `migrations_v4/000246_fine_permissions.up.sql`
- `migrations_v4/000247_actor_taxid_trgm.up.sql`

## UNKNOWN

- Actor merge: not implemented.
- Backfill of existing carrier rows into `actors`/`actor_keys` (plan part V.7): pending; with the Identity Gate off the tables are empty.
- Fields/projects/supplies/labors as registry entity types: future scope.
