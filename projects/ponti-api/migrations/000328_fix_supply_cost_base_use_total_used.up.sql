-- =============================================================================
-- MIGRACIÓN 000328: Ajustar costo de insumos base a total_used (SSOT)
-- =============================================================================
--
-- Propósito: Alinear métricas de workorders con exportación (Excel)
-- Regla: Costo de insumos ejecutado = SUM(total_used * price)
-- Impacto: v3_workorder_metrics y todo lo que lo consume (dashboard, lot_metrics, summary)
-- Nota: Comentarios en español, código en inglés
--

BEGIN;

-- Reemplazar la función base para que use total_used (coincide con exportación)
CREATE OR REPLACE FUNCTION v3_lot_ssot.supply_cost_for_lot_base(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(wi.total_used * s.price), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
$$;

COMMIT;
