-- ========================================
-- MIGRATION 000198 LABOR PARTIAL PRICE FLAG (DOWN)
-- ========================================
-- Revierte el cambio de la migración 000198.

BEGIN;

ALTER TABLE public.labors
    DROP COLUMN IF EXISTS is_partial_price;

COMMIT;
