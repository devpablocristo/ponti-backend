BEGIN;

-- No-op by design.
-- The up migration rewrites duplicated per-lot consumption to the canonical
-- value derived from final_dose * effective_area. The original duplicated
-- group totals are not recoverable without an external backup.

COMMIT;
