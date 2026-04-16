BEGIN;

DROP INDEX IF EXISTS idx_supplies_pending_notdel;

ALTER TABLE public.supplies
    DROP COLUMN IF EXISTS is_pending;

COMMIT;
