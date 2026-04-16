-- ========================================
-- MIGRATION 000201 WORK ORDER DRAFTS DIGITAL FLAG (UP)
-- ========================================

BEGIN;

ALTER TABLE public.work_order_drafts
    ADD COLUMN IF NOT EXISTS is_digital boolean;

CREATE INDEX IF NOT EXISTS idx_work_order_drafts_is_digital
    ON public.work_order_drafts (is_digital);

COMMIT;
