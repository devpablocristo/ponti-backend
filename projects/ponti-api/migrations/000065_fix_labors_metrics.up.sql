-- ========================================
-- MIGRACIÓN 000070: FIX LABORS METRICS
-- Entidad: labor (Labores)
-- Funcionalidad: Corregir labor_cards_cube_view para calcular correctamente el total USD neto
-- ========================================

-- Eliminar la vista actual
DROP VIEW IF EXISTS labor_cards_cube_view;

-- Crear la vista corregida con el cálculo correcto del total USD neto
CREATE VIEW labor_cards_cube_view AS
SELECT
  w.project_id,
  w.field_id,
  CASE
    WHEN GROUPING(w.project_id)=0 AND GROUPING(w.field_id)=0 THEN 'project+field'
    WHEN GROUPING(w.project_id)=0 AND GROUPING(w.field_id)=1 THEN 'project'
    WHEN GROUPING(w.project_id)=1 AND GROUPING(w.field_id)=0 THEN 'field'
    ELSE 'global'
  END AS level,
  
  -- Superficie total en hectáreas
  SUM(w.effective_area) AS surface_ha,
  
  -- CORRECCIÓN: Total USD neto = precio labor * superficie (por cada workorder)
  SUM(lb.price * w.effective_area) AS total_labor_cost,
  
  -- Costo promedio por hectárea
  CASE
    WHEN SUM(w.effective_area) > 0
      THEN SUM(lb.price * w.effective_area) / SUM(w.effective_area)
    ELSE 0
  END AS labor_cost_per_ha

FROM workorders w
JOIN labors lb ON lb.id = w.labor_id
WHERE
  w.deleted_at IS NULL
  AND lb.deleted_at IS NULL
  AND w.effective_area IS NOT NULL
  AND w.effective_area > 0
  AND lb.price IS NOT NULL
GROUP BY GROUPING SETS (
  (w.project_id, w.field_id),
  (w.project_id),
  (w.field_id),
  ()
);

-- Comentarios en español
COMMENT ON VIEW labor_cards_cube_view IS 'Vista corregida para métricas de labores con cálculo correcto del total USD neto';
COMMENT ON COLUMN labor_cards_cube_view.surface_ha IS 'Superficie total en hectáreas';
COMMENT ON COLUMN labor_cards_cube_view.total_labor_cost IS 'Total USD neto (precio labor × superficie por cada workorder)';
COMMENT ON COLUMN labor_cards_cube_view.labor_cost_per_ha IS 'Costo promedio por hectárea';
