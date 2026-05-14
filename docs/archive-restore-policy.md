# Archive / Restore Policy

## Archive

Any archive operation that can affect more than one row must:

1. Open a transaction.
2. Create an `archive_batches` row.
3. Archive the root row.
4. Archive child rows according to the entity policy.
5. Write the same `archive_batch_id` and origin metadata to cascaded children.
6. Commit the transaction.

## Restore

Any restore operation must:

1. Open a transaction.
2. Validate that the root row is archived.
3. Validate that the parent is active when required.
4. Restore the root row.
5. Restore children only when their archive cause matches the root operation.
6. Clear archive metadata on restored rows.
7. Commit the transaction.

The golden rule:

```txt
Do not restore by relationship.
Restore by cause.
```

## Hard Delete

Hard delete is physical deletion and must be blocked for active rows.

Default rule:

```txt
hard delete allowed only when deleted_at IS NOT NULL
```

Entities with business dependencies must block hard delete until those
dependencies are removed or a specific policy explicitly allows physical
cascade.

