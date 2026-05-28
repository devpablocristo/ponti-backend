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

For controlled cleanup, use the internal command. It is dry-run by default and
does not hard delete or restore archived parents:

```sh
# Local first, against a DB restored from prod/staging.
go run ./cmd/archive-cleanup --dry-run --output table \
    > /tmp/archive-cleanup-local-pre.txt

# Optional tenant scope.
go run ./cmd/archive-cleanup --dry-run --tenant-id "$TENANT_ID" --output json \
    > /tmp/archive-cleanup-tenant-pre.json

# Apply only after reviewing the dry-run.
go run ./cmd/archive-cleanup --apply --output table \
    > /tmp/archive-cleanup-local-apply.txt

# Evidence after remediation.
psql "$PONTI_DB_URL_LOCAL" -f scripts/data-audit/archived_invariants.sql \
    > /tmp/archive-cleanup-local-post.txt
```

The command remediates IA-1 through IA-10 by archiving active children under
archived parents with the same lifecycle cause as the parent. If an archived
parent predates lifecycle metadata, it creates a legacy archive batch with
reason `legacy archive metadata backfill`, backfills the parent metadata, and
uses that cause for affected children. IA-14 is a metadata-only backfill for
already archived rows without `archive_batch_id`.

IA-11 through IA-13 are report-only/manual-review checks. If any of those rows
exist, `--apply` exits non-zero before mutating data so an operator can decide
how to handle actor/investor references explicitly.

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
# 3. If dirty → run archive-cleanup dry-run, review, then apply only with approval.
```

### Cleanup rollout

1. **Local:** restore from production/staging, run `archive-cleanup --dry-run`,
   review the report, run `archive-cleanup --apply`, then run
   `archived_invariants.sql`. Keep pre/apply/post logs as evidence.
2. **Staging:** take a backup first. Run dry-run, attach the report to the
   deployment note, then apply. Smoke archive/restore for customer and project.
3. **Production:** do not run from CI. Require explicit approval, confirmed
   backup, attached dry-run report, a short execution window, and a post-audit
   report saved with the change record.

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
