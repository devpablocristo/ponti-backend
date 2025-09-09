-- ========================================
-- MIGRACIÓN 000079: REVERTIR CORRECCIÓN DE COSTO ADMINISTRATIVO
-- Entidad: views (Revertir base_admin_costs_view)
-- Funcionalidad: Revertir a la lógica anterior de distribución
-- ========================================

-- Revertir a la vista original base_admin_costs_view
DROP VIEW IF EXISTS base_admin_costs_view;

CREATE VIEW base_admin_costs_view AS
WITH project_total_hectares AS (
  SELECT
    p.id AS project_id,
    COALESCE(SUM(l.hectares), 1) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.id
)
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.hectares,
  p.admin_cost,
  pth.total_hectares,
  CASE 
    WHEN l.hectares > 0 
    THEN p.admin_cost / pth.total_hectares 
    ELSE 0 
  END AS admin_cost_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
JOIN project_total_hectares pth ON pth.project_id = p.id
WHERE l.deleted_at IS NULL;
