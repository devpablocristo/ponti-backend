-- ========================================
-- MIGRATION 000197 SUPPLY PARTIAL PRICE FLAG (DOWN)
-- ========================================
-- Revierte el cambio de la migración 000197.

BEGIN;

ALTER TABLE public.supplies
    DROP COLUMN IF EXISTS is_partial_price;

COMMIT;
