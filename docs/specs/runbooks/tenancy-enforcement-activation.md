# Tenancy Enforcement Activation Runbook

Specification type: activation procedure + precondition checklist for the multi-tenancy / identity feature flags.

Scope: how and when to turn ON `TENANT_ENFORCEMENT` and `IDENTITY_GATE`. The data model and isolation rules (the "what") live in `../data/tenant-isolation.md` and `../system/security-and-tenancy.md`; this runbook is the "how/when".

## The Flags

Two independent, env-driven, default-OFF flags resolved once at startup:

- `TENANT_ENFORCEMENT` — when ON, list/get/mutation queries are physically scoped by `tenant_id` (the caller's resolved tenant). When OFF, `tenant_id` is never referenced and all rows are visible (single-tenant behavior).
- `IDENTITY_GATE` — when ON, write-paths resolve against the actor registry (`actors`/`actor_keys`) for dedup. When OFF, that resolution is skipped.

Evidence:
- `internal/shared/models/base.go:56` (`TenantEnforcementEnabled`) — `v == "1" || EqualFold(v,"true")`, absent ⇒ false.
- `internal/shared/models/base.go:72` (`IdentityGateEnabled`) — same, absent ⇒ false.
- `internal/shared/filters/workspace.go` — `ScopeTenant` / `TenantProjectScope` / `TenantFieldScope` (no-op when the flag is off).

## Verified Current State (2026-06-17, local DB = prod copy)

- **Flags are OFF everywhere.** Local `.env` had them `true` (prod-like, wrong for single-tenant); dev/prod deploys do **not** set the vars at all ⇒ default false. That is why dev/prod "just work" and a flags-on local showed empty lists.
  - Evidence: `.github/workflows/deploy-dev.yml:124` (ENV_VARS sets `AUTH_ENABLED=false`, no `TENANT_ENFORCEMENT`/`IDENTITY_GATE`).
- **Backfill of master/catalog `tenant_id` is COMPLETE.** 0 NULLs across projects, customers, crops, campaigns, providers, categories, investors, managers, lease_types, types, business_parameters. All rows assigned to the `'default'` tenant.
  - Evidence: `UPDATE ... SET tenant_id = (SELECT id FROM auth_tenants WHERE name='default')` in `migrations_v4/000237_tenant_id_master_entities.up.sql` and `migrations_v4/000238_tenant_id_catalogs.up.sql`.
- **Only one real tenant exists.** Every scoped table has exactly 1 distinct `tenant_id` (`'default'`); the 5 `auth_memberships` all point to it.
- **`actors` / `actor_keys` are EMPTY** ⇒ Identity Gate has no dedup registry to honor yet.
- **Child entities have no `tenant_id`** (fields, lots, supplies, work_orders) — they are scoped indirectly via project (`TenantProjectScope`/`TenantFieldScope`), not their own column.

## Recommendation: DO NOT activate yet

Turning enforcement ON today buys nothing (a single tenant has nothing to isolate from) and adds a real failure mode: any caller path that fails to resolve the correct tenant returns empty lists (the exact symptom observed locally). Keep both flags OFF until there is a second real tenant AND the checklist below is green.

## Preconditions to activate `TENANT_ENFORCEMENT`

- [ ] A second real tenant exists with its own `auth_tenants` row and `auth_memberships` mapping its users.
- [ ] **Tenant resolution is bulletproof on every entry path** — verify each resolves to the correct tenant (not empty / not `'default'`): Identity Platform JWT login, BFF (`X-API-Key` + `X-User-Id`), API keys, background jobs, AI service calls.
- [ ] **Dual-write confirmed on all creates** — new rows get the caller's `tenant_id` (never NULL/default-by-accident). Re-verify after any new create path is added.
- [ ] **Child-entity scoping verified end-to-end** — fields/lots/supplies/work_orders return correctly when filtered via project ownership under enforcement.
- [ ] `AUTH_PLATFORM_ADMIN_SUBJECTS` allowlist set so platform-admin (cross-tenant) operations still work.
- [ ] FE tenant context/switcher ready (platform-admin UI for tenant management is a pending frontend item).
- [ ] Dry-run in **dev** first with enforcement ON; confirm NO list silently empties for each tenant's users.

## Preconditions to activate `IDENTITY_GATE`

- [ ] `actors` / `actor_keys` backfilled (the dedup registry is populated). Activating against an empty registry breaks write-paths.
- [ ] Identity-gate dedup verified in dev with the backfilled registry.

## Activation Procedure

1. Set the env var(s) in the deploy config (add to the `ENV_VARS` string), per environment: dev → staging → prod.
   - File: `.github/workflows/deploy-dev.yml` (and `deploy-staging.yml` / `deploy-prod.yml`), append `,TENANT_ENFORCEMENT=true` (and/or `,IDENTITY_GATE=true`).
2. Deploy. `RUN_MIGRATIONS_ON_STARTUP=true` is already set; no schema change is required to flip a flag.
3. Verify (below). Roll forward one environment at a time; do not enable in prod until dev+staging are clean.

Local toggling for testing: edit `.env` (`TENANT_ENFORCEMENT=true`), then **recreate** the container (not `restart` — `env_file` is only re-read on recreate): `docker compose -f docker-compose.yml up -d --force-recreate --no-deps ponti-api`.

## Verification

- For a user of each tenant, list every major entity (projects, customers, catalogs, work-orders) and confirm the counts match that tenant's data — and that cross-tenant rows are NOT visible.
- Confirm a create writes the correct `tenant_id`.
- Confirm platform-admin can still operate cross-tenant.

## Rollback

- Remove the env var (or set `=false`) and redeploy. No schema rollback is needed — the flag only gates query behavior; `tenant_id` data stays intact.

## Related

- `../data/tenant-isolation.md`, `../system/security-and-tenancy.md` (the "what").
