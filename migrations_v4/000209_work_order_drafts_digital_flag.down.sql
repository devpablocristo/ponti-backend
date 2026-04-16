-- ========================================
-- MIGRATION 000201 WORK ORDER DRAFTS DIGITAL FLAG (DOWN)
-- ========================================

BEGIN;

DROP INDEX IF EXISTS idx_work_order_drafts_is_digital;

ALTER TABLE public.work_order_drafts
    DROP COLUMN IF EXISTS is_digital;

COMMIT;
