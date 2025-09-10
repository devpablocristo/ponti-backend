-- ========================================
-- MIGRACIÓN 000079: CORREGIR CÁLCULO DE COSTO ADMINISTRATIVO
-- Entidad: views (Corregir base_admin_costs_view)
-- Funcionalidad: Corregir el cálculo para que admin_cost sea costo por hectárea, no total
-- ========================================

-- Corregir la vista base_admin_costs_view
-- El problema: está dividiendo admin_cost entre total_hectares
-- La solución: admin_cost ya es el costo por hectárea
DROP VIEW IF EXISTS base_admin_costs_view CASCADE;

CREATE VIEW base_admin_costs_view AS
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.hectares,
  p.admin_cost, -- Este ya es el costo por hectárea
  p.admin_cost AS admin_cost_per_ha -- Usar directamente el admin_cost como costo por hectárea
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
WHERE l.deleted_at IS NULL;

-- Recrear fix_lot_list que depende de base_admin_costs_view
CREATE VIEW fix_lot_list AS
SELECT
  f.project_id,
  l.field_id,
  p.customer_id,
  p.campaign_id,
  p.name AS project_name,
  f.name AS field_name,
  l.id,
  l.name AS lot_name,
  l.variety,
  CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END AS sowed_area,
  l.hectares,
  l.season,
  l.updated_at,
  l.tons,
  l.previous_crop_id,
  pc.name AS previous_crop,
  l.current_crop_id,
  cc_crop.name AS current_crop,
  -- Usar vista base para costo administrativo (CORREGIDO)
  bac.admin_cost_per_ha,
  l.hectares AS harvested_area,
  -- Función para calcular fecha de cosecha
  CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN CURRENT_DATE ELSE NULL END AS harvest_date,
  -- Campos simplificados por ahora
  0 AS cost_usd_per_ha,
  0 AS yield_tn_per_ha,
  0 AS income_net_per_ha,
  0 AS rent_per_ha,
  0 AS active_total_per_ha,
  0 AS operating_result_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN crops pc ON pc.id = l.previous_crop_id AND pc.deleted_at IS NULL
LEFT JOIN crops cc_crop ON cc_crop.id = l.current_crop_id AND cc_crop.deleted_at IS NULL
LEFT JOIN base_admin_costs_view bac ON bac.lot_id = l.id
WHERE l.deleted_at IS NULL;
