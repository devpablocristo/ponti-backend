-- ========================================
-- MIGRATION 000199 WORK ORDER DRAFTS (DOWN)
-- ========================================

BEGIN;

DROP TABLE IF EXISTS public.work_order_draft_investor_splits;
DROP TABLE IF EXISTS public.work_order_draft_items;
DROP TABLE IF EXISTS public.work_order_drafts;

COMMIT;
