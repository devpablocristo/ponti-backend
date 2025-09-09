-- ========================================
-- MIGRACIÓN 000074: CORREGIR VISTAS BASE PARA DASHBOARD
-- ========================================
-- Propósito: Aplicar correcciones de migrations-116 a las vistas base
-- Incluye: Fixes de admin_cost, utilidad negativa y consistencia de áreas
-- ========================================

-- ========================================
-- 1. CORREGIR BASE_ADMIN_COSTS_VIEW
-- ========================================
-- Problema: admin_cost se estaba dividiendo entre total_hectares incorrectamente
-- Solución: admin_cost ya es el costo por hectárea, usarlo directamente
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

-- ========================================
-- 2. CORREGIR BASE_LEASE_CALCULATIONS_VIEW
-- ========================================
-- Problema: Se calculaba utilidad negativa en arriendo por porcentaje
-- Solución: Implementar regla "No se calcula utilidad negativa!" con GREATEST(..., 0)
DROP VIEW IF EXISTS base_lease_calculations_view;

CREATE VIEW base_lease_calculations_view AS
SELECT
  l.id AS lot_id,
  l.field_id,
  f.project_id,
  l.hectares,
  f.lease_type_id,
  f.lease_type_percent,
  f.lease_type_value,
  -- Ingreso neto por hectárea (desde base_income_net_view)
  bin.income_net_per_ha,
  -- Costo por hectárea (desde base_direct_costs_view)
  bdc.direct_cost / NULLIF(l.hectares, 0) AS cost_per_ha,
  -- Costo administrativo por hectárea (desde base_admin_costs_view)
  bac.admin_cost_per_ha,
  -- Función para convertir porcentaje a decimal
  CASE
    WHEN f.lease_type_id = 1 THEN -- % INGRESO NETO
      (COALESCE(f.lease_type_percent, 0) / 100.0) * bin.income_net_per_ha
    WHEN f.lease_type_id = 2 THEN -- % UTILIDAD (CON FIX: No se calcula utilidad negativa)
      (COALESCE(f.lease_type_percent, 0) / 100.0) *
      GREATEST(
        (bin.income_net_per_ha - (bdc.direct_cost / NULLIF(l.hectares, 0)) - bac.admin_cost_per_ha),
        0
      )
    WHEN f.lease_type_id = 3 THEN -- ARRIENDO FIJO
      COALESCE(f.lease_type_value, 0)
    WHEN f.lease_type_id = 4 THEN -- ARRIENDO FIJO + % INGRESO NETO
      COALESCE(f.lease_type_value, 0) + 
      ((COALESCE(f.lease_type_percent, 0) / 100.0) * bin.income_net_per_ha)
    ELSE 0
  END AS rent_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
LEFT JOIN base_income_net_view bin ON bin.lot_id = l.id
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
LEFT JOIN base_admin_costs_view bac ON bac.lot_id = l.id
WHERE l.deleted_at IS NULL;

-- ========================================
-- 3. RECREAR FIX_LOT_LIST CON CORRECCIONES
-- ========================================
-- Actualizar fix_lot_list para usar las vistas base corregidas
DROP VIEW IF EXISTS fix_lot_list;

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
  byc.hectares AS harvested_area,
  -- Función para calcular fecha de cosecha
  CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN CURRENT_DATE ELSE NULL END AS harvest_date,
  -- Usar vista base para costo USD por hectárea
  bdc.direct_cost / NULLIF(l.hectares, 0) AS cost_usd_per_ha,
  -- Usar vista base para rendimiento
  byc.yield_tn_per_ha,
  -- Usar vista base para ingreso neto por hectárea
  bin.income_net_per_ha,
  -- Usar vista base para arriendo por hectárea (CORREGIDO)
  blc.rent_per_ha,
  -- Usar vista base para total activo por hectárea
  bat.active_total_per_ha,
  -- Usar vista base para resultado operativo por hectárea
  bor.operating_result_per_ha
FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN crops pc ON pc.id = l.previous_crop_id AND pc.deleted_at IS NULL
LEFT JOIN crops cc_crop ON cc_crop.id = l.current_crop_id AND cc_crop.deleted_at IS NULL
-- Usar vistas base corregidas
LEFT JOIN base_direct_costs_view bdc ON bdc.lot_id = l.id
LEFT JOIN base_yield_calculations_view byc ON byc.lot_id = l.id
LEFT JOIN base_income_net_view bin ON bin.lot_id = l.id
LEFT JOIN base_admin_costs_view bac ON bac.lot_id = l.id
LEFT JOIN base_lease_calculations_view blc ON blc.lot_id = l.id
LEFT JOIN base_active_total_view bat ON bat.lot_id = l.id
LEFT JOIN base_operating_result_view bor ON bor.lot_id = l.id
WHERE l.deleted_at IS NULL;
