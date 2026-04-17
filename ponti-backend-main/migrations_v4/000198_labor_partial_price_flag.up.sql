-- ========================================
-- MIGRATION 000198 LABOR PARTIAL PRICE FLAG (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

ALTER TABLE public.labors
    ADD COLUMN is_partial_price boolean NOT NULL DEFAULT false;

COMMIT;
