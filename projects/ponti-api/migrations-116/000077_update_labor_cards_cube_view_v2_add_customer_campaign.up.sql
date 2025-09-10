-- ========================================
-- MIGRACIÓN 000077: ACTUALIZAR LABOR_CARDS_CUBE_VIEW_V2
-- Entidad: labor (Agregar customer_id y campaign_id a la vista v2)
-- Funcionalidad: Incluir filtros de cliente y campaña en la vista de métricas de labor
-- ========================================

-- Actualizar labor_cards_cube_view_v2 para incluir customer_id y campaign_id
DROP VIEW IF EXISTS labor_cards_cube_view_v2;

CREATE VIEW labor_cards_cube_view_v2 AS
WITH labor_metrics AS (
  SELECT
    bdc.project_id,
    bdc.field_id,
    p.customer_id,
    p.campaign_id,
    -- Usar vista base para superficie y costos labor
    SUM(w.effective_area) AS surface_ha,
    SUM(bdc.labor_cost) AS total_labor_cost,
    CASE 
      WHEN SUM(w.effective_area) > 0 
      THEN SUM(bdc.labor_cost) / SUM(w.effective_area) 
      ELSE 0 
    END AS labor_cost_per_ha
  FROM base_direct_costs_view bdc
  JOIN workorders w ON w.project_id = bdc.project_id 
    AND w.field_id = bdc.field_id 
    AND w.lot_id = bdc.lot_id
  JOIN projects p ON p.id = w.project_id AND p.deleted_at IS NULL
  WHERE w.deleted_at IS NULL 
    AND w.effective_area > 0
  GROUP BY bdc.project_id, bdc.field_id, p.customer_id, p.campaign_id
)
SELECT
  project_id,
  field_id,
  customer_id,
  campaign_id,
  'project+field' AS level,
  surface_ha,
  total_labor_cost AS net_total_cost,
  labor_cost_per_ha AS avg_cost_per_ha
FROM labor_metrics
UNION ALL
SELECT
  project_id,
  NULL AS field_id,
  customer_id,
  campaign_id,
  'project' AS level,
  SUM(surface_ha) AS surface_ha,
  SUM(total_labor_cost) AS net_total_cost,
  CASE 
    WHEN SUM(surface_ha) > 0 
    THEN SUM(total_labor_cost) / SUM(surface_ha) 
    ELSE 0 
  END AS avg_cost_per_ha
FROM labor_metrics
GROUP BY project_id, customer_id, campaign_id
UNION ALL
SELECT
  NULL AS project_id,
  NULL AS field_id,
  customer_id,
  campaign_id,
  'global' AS level,
  SUM(surface_ha) AS surface_ha,
  SUM(total_labor_cost) AS net_total_cost,
  CASE 
    WHEN SUM(surface_ha) > 0 
    THEN SUM(total_labor_cost) / SUM(surface_ha) 
    ELSE 0 
  END AS avg_cost_per_ha
FROM labor_metrics
GROUP BY customer_id, campaign_id;
