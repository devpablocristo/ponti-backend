-- ========================================
-- MIGRACIÓN 000075: REVERTIR WORKORDER_METRICS_VIEW_V2
-- Entidad: workorder (Revertir customer_id y campaign_id de la vista v2)
-- Funcionalidad: Restaurar vista original sin filtros de cliente y campaña
-- ========================================

-- Revertir workorder_metrics_view_v2 a su estado original
DROP VIEW IF EXISTS workorder_metrics_view_v2;

CREATE VIEW workorder_metrics_view_v2 AS
SELECT
  w.project_id,
  w.field_id,
  -- w.customer_id, -- REMOVED
  -- w.campaign_id, -- REMOVED
  SUM(w.effective_area) AS surface_ha,
  SUM(CASE WHEN s.unit_id = 1 THEN wi.final_dose * w.effective_area ELSE 0 END) AS liters,
  SUM(CASE WHEN s.unit_id = 2 THEN wi.final_dose * w.effective_area ELSE 0 END) AS kilograms,
  -- Usar vista base para costos directos
  COALESCE(bdc.direct_cost, 0) AS direct_cost
FROM workorders w
LEFT JOIN workorder_items wi ON wi.workorder_id = w.id
LEFT JOIN supplies s ON s.id = wi.supply_id
LEFT JOIN base_direct_costs_view bdc ON bdc.project_id = w.project_id
  AND bdc.field_id = w.field_id
  AND bdc.lot_id = w.lot_id -- Added lot_id for more precise join
WHERE w.deleted_at IS NULL
GROUP BY w.project_id, w.field_id, bdc.direct_cost; -- Removed customer_id, campaign_id
