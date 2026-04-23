-- ========================================
-- MIGRATION 000217 DASHBOARD STOCK COUNT FROM REAL STOCK (DOWN)
-- ========================================

BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.last_manual_stock_entry_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MAX(sm.movement_date::date)
  FROM public.supply_movements sm
  WHERE sm.project_id = p_project_id
    AND sm.deleted_at IS NULL
    AND sm.movement_type = 'Stock'
$$;

COMMIT;
