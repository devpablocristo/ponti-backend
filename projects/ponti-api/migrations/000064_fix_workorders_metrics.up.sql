-- ========================================
-- MIGRACIÓN 000069: FIX WORKORDER METRICS
-- Entidad: workorder (Órdenes de Trabajo)
-- Funcionalidad: Reemplazar workorder_metrics_view con nueva lógica de cálculo
-- ========================================

-- Eliminar la vista antigua
DROP VIEW IF EXISTS workorder_metrics_view;

-- Crear la nueva vista workorder_metrics_view usando la lógica mejorada
CREATE VIEW workorder_metrics_view AS
SELECT
  w.project_id,
  w.field_id,
  p.customer_id,
  p.campaign_id,
  
  -- Superficie total en hectáreas
  SUM(w.effective_area) AS surface_ha,
  
  -- Litros: final_dose para Herbicidas(4), Fungicidas(6), Insecticidas(5), Coadyuvantes(2)
  SUM(COALESCE(wi.final_dose, 0))
    FILTER (WHERE s.category_id IN (2,4,5,6)) AS liters,
  
  -- Kilogramos: final_dose para Fertilizantes(8), Semilla(1), Curasemillas(3)
  SUM(COALESCE(wi.final_dose, 0))
    FILTER (WHERE s.category_id IN (1,3,8)) AS kilograms,
  
  -- Costo directo: usando la lógica mejorada de v_calc_workorders
  SUM(
    COALESCE(lb.price * w.effective_area, 0) +  -- Costo labor por hectárea
    COALESCE(wi.final_dose * s.price * w.effective_area, 0)  -- Costo supplies por hectárea
  ) AS direct_cost

FROM workorders w
  JOIN projects p ON p.id = w.project_id AND p.deleted_at IS NULL
  JOIN labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  LEFT JOIN workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
WHERE w.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
GROUP BY w.project_id, w.field_id, p.customer_id, p.campaign_id;

-- Comentarios en español
COMMENT ON VIEW workorder_metrics_view IS 'Vista optimizada para métricas de workorders con cálculos mejorados';
COMMENT ON COLUMN workorder_metrics_view.surface_ha IS 'Superficie total en hectáreas';
COMMENT ON COLUMN workorder_metrics_view.liters IS 'Consumo total en litros (Herbicidas, Fungicidas, Insecticidas, Coadyuvantes)';
COMMENT ON COLUMN workorder_metrics_view.kilograms IS 'Consumo total en kilogramos (Fertilizantes, Semilla, Curasemillas)';
COMMENT ON COLUMN workorder_metrics_view.direct_cost IS 'Costo directo total (labor + supplies) en USD';
