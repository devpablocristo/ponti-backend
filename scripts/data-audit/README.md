# Data audit — archived invariants

Scripts to detect rows that violate the hierarchical-integrity invariant:

> A child can not be active under an archived parent. `child.deleted_at IS NULL`
> AND `parent.deleted_at IS NOT NULL` is an illegal state of the domain.

## Usage

Run **read-only** against staging first. Each query reports a check id, the
number of violating rows, and a comma-separated sample of ids:

```sh
psql "$PONTI_DB_URL_STAGING" -f archived_invariants.sql
```

If any check returns rows > 0, decide per case:

1. **Cascade-archive the children.** The most common remediation. Use the
   existing `lifecycle.ArchiveScopedRows` flow from a one-shot migration or
   admin endpoint.
2. **Restore the archived parent.** If the parent was archived by mistake and
   the children are still operationally valid, restore the parent so the
   relationship becomes consistent again.
3. **Hard delete orphans.** Only after operational review; coordinate with the
   data owner.

Re-run the script after remediation to confirm zero rows.

## When to run (mandatory checkpoints)

The Go code path enforces the invariant going forward (the
`assertXReferencesActive` + `RestoreRequiresActiveParent` patterns). This
script catches pre-existing inconsistencies and any regression that bypasses
the BE checks (manual SQL, raw migrations, etc.).

**REQUIRED before applying migration `000233_archived_invariant_triggers.up.sql`**.
The triggers added in that migration reject any new INSERT/UPDATE that would
violate the invariant. If the DB already contains inconsistent rows, the
triggers themselves create fine (they only fire on new writes), but the
first UPDATE on any pre-existing inconsistent row will fail with SQLSTATE
23514. Run the audit + cleanup first to avoid surprises.

### Deploy runbook step

Add this to the release checklist for any release that ships migrations
000233 or later:

```sh
# 1. Audit (read-only)
psql "$PONTI_DB_URL_STAGING" -f scripts/data-audit/archived_invariants.sql \
    > /tmp/audit-staging.log
grep -E '^.+\| *[1-9]' /tmp/audit-staging.log  # any non-zero check fails

# 2. If clean → apply migrations as usual.
# 3. If dirty → coordinate cleanup with data owner before deploy.
```

### Recurring schedule

Beyond release checkpoints, run this script:
- After any large-scale archive/restore operation (bulk imports, manual
  customer archives, data migrations).
- Monthly as part of data hygiene — easy to wire as a cron job that mails
  the report when any check returns non-zero.

### What the BE invariants already guarantee

The triggers in migration 000233, the `assertXReferencesActive` helpers
across repos, and the `RestoreRequiresActiveParent` checks together prevent
*new* invariant violations through any Go code path or transactional SQL.
The audit script is the safety net for:
- Pre-feature data (rows archived before migration 000227 brought
  `archive_batch_id` tracking).
- Raw SQL fixes applied by an operator outside of the BE flow.
- Future regressions if someone bypasses the helpers.
- Migration mistakes (a DDL that inserts rows without going through
  validation).
