-- ========================================
-- MIGRATION 000160 LOT_DATES UNIQUE (DOWN)
-- ========================================

BEGIN;

ALTER TABLE ONLY public.lot_dates
  DROP CONSTRAINT IF EXISTS uq_lot_dates_lot_id_sequence;

COMMIT;
