-- ========================================
-- MIGRATION 000160 LOT_DATES UNIQUE (UP)
-- ========================================

BEGIN;

ALTER TABLE ONLY public.lot_dates
  ADD CONSTRAINT uq_lot_dates_lot_id_sequence UNIQUE (lot_id, sequence);

COMMIT;
