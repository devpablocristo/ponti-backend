-- ========================================
-- MIGRATION 000217 DASHBOARD STOCK COUNT FROM REAL STOCK (UP)
-- ========================================
-- Ajusta "Arqueo de stock" para usar la misma fuente que la pantalla STOCK:
-- el ultimo conteo manual de stock de campo guardado en public.stocks.

BEGIN;

CREATE OR REPLACE FUNCTION v4_ssot.last_manual_stock_entry_date_for_project(p_project_id bigint)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT MAX(s.updated_at::date)
  FROM public.stocks s
  WHERE s.project_id = p_project_id
    AND s.close_date IS NULL
    AND s.has_real_stock_count = TRUE
$$;

COMMIT;
