-- ========================================
-- MIGRATION 000197 SUPPLY PARTIAL PRICE FLAG (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

ALTER TABLE public.supplies
    ADD COLUMN is_partial_price boolean NOT NULL DEFAULT false;

COMMIT;
