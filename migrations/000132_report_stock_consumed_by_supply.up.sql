-- Crea vista para consumos de insumos por proyecto
CREATE OR REPLACE VIEW v4_report.stock_consumed_by_supply AS
SELECT
  wo.project_id,
  woi.supply_id,
  COALESCE(SUM(woi.total_used), 0) AS consumed
FROM public.workorder_items woi
JOIN public.workorders wo ON wo.id = woi.workorder_id
WHERE wo.deleted_at IS NULL
  AND woi.deleted_at IS NULL
  AND woi.supply_id IS NOT NULL
GROUP BY wo.project_id, woi.supply_id;
