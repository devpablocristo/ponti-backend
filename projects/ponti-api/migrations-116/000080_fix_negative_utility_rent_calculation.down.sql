-- ========================================
-- ROLLBACK: MIGRACIÓN 000080
-- ========================================
-- Revertir el fix de utilidad negativa en arriendo por porcentaje

-- Recrear la vista base_lease_calculations_view sin el fix
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
    WHEN f.lease_type_id = 2 THEN -- % UTILIDAD (SIN FIX: permite utilidad negativa)
      (COALESCE(f.lease_type_percent, 0) / 100.0) *
      (bin.income_net_per_ha - (bdc.direct_cost / NULLIF(l.hectares, 0)) - bac.admin_cost_per_ha)
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
