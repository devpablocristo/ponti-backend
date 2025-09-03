-- ========================================
-- MIGRACIÓN 000073: FIX LOT LIST
-- Entidad: lot (Lotes)
-- Funcionalidad: Crear vista fix_lot_list con cálculos correctos
-- ========================================

-- Eliminar la vista si existe
DROP VIEW IF EXISTS fix_lot_list;

-- Crear la vista fix_lot_list con cálculos correctos
CREATE VIEW fix_lot_list AS
WITH
-- =======================
-- CÁLCULO DE ÁREA SEMBRADA (optimizado)
-- =======================
sowing AS (
  SELECT 
    w.lot_id, 
    SUM(w.effective_area) AS sowed_area
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.category_id = 9  -- Siembra
  GROUP BY w.lot_id
),

-- =======================
-- CÁLCULO DE ÁREA COSECHADA (optimizado)
-- =======================
harvesting AS (
  SELECT 
    w.lot_id, 
    SUM(w.effective_area) AS harvested_area,
    MAX(w.date) AS harvest_date
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
    AND lb.category_id = 13  -- Cosecha
  GROUP BY w.lot_id
),

-- =======================
-- CÁLCULO DE COSTOS DIRECTOS (optimizado)
-- =======================
direct_costs AS (
  SELECT 
    w.lot_id,
    SUM(COALESCE(lb.price * w.effective_area, 0)) AS labor_cost,
    SUM(COALESCE(wi.final_dose * s.price * w.effective_area, 0)) AS supply_cost
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  LEFT JOIN workorder_items wi ON w.id = wi.workorder_id AND wi.deleted_at IS NULL
  LEFT JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.deleted_at IS NULL
    AND lb.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
  GROUP BY w.lot_id
),

-- =======================
-- CÁLCULO DE INGRESO NETO (optimizado) - PRECIO REAL DE COMERCIALIZACIÓN
-- =======================
income_net AS (
  SELECT 
    l.id AS lot_id,
    COALESCE(l.tons, 0) * COALESCE(cc.net_price, 0) AS income_net_total
  FROM lots l
  JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN crop_commercializations cc ON cc.project_id = f.project_id 
    AND cc.crop_id = l.current_crop_id 
    AND cc.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
    AND l.tons IS NOT NULL
    AND l.tons > 0
)

-- =======================
-- VISTA FINAL CON TODOS LOS CÁLCULOS CORREGIDOS
-- =======================
SELECT
  f.project_id,
  l.field_id,
  p.name AS project_name,
  f.name AS field_name,
  l.id,
  l.name AS lot_name,
  l.variety,
  COALESCE(s.sowed_area, 0) AS sowed_area,
  l.hectares,
  l.season,
  l.updated_at,
  l.tons,
  l.previous_crop_id,
  pc.name AS previous_crop,
  l.current_crop_id,
  cc_crop.name AS current_crop,
  -- CORRECCIÓN: Costo administrativo por hectárea (admin_cost del proyecto ÷ total hectáreas del proyecto)
  CASE 
    WHEN l.hectares > 0 
    THEN p.admin_cost / (
      SELECT COALESCE(SUM(l2.hectares), 1) 
      FROM lots l2 
      JOIN fields f2 ON l2.field_id = f2.id 
      WHERE f2.project_id = p.id AND l2.deleted_at IS NULL
    )
    ELSE 0 
  END AS admin_cost_per_ha,
  COALESCE(h.harvested_area, 0) AS harvested_area,
  h.harvest_date,
  
  -- CORRECCIÓN: Costo USD por hectárea (labor + supplies)
  CASE 
    WHEN COALESCE(s.sowed_area, 0) > 0 
    THEN (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) / s.sowed_area
    ELSE 0 
  END AS cost_usd_per_ha,
  
  -- CORRECCIÓN: Rendimiento usando la fórmula correcta (tons / hectares)
  COALESCE(l.tons / NULLIF(l.hectares, 0), 0) AS yield_tn_per_ha,
  
  -- CORRECCIÓN: Ingreso neto por hectárea
  CASE 
    WHEN l.hectares > 0 
    THEN COALESCE(in_net.income_net_total, 0) / l.hectares
    ELSE 0 
  END AS income_net_per_ha,
  
  -- CORRECCIÓN: Arriendo por hectárea según tipo de arriendo
  CASE 
    WHEN f.lease_type_id = 1 THEN -- % INGRESO NETO
      (COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0) / NULLIF(l.hectares, 0)
    WHEN f.lease_type_id = 2 THEN -- % UTILIDAD
      (COALESCE(f.lease_type_percent, 0) / 100.0) * 
      (COALESCE(in_net.income_net_total, 0) - 
       (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) - 
       (p.admin_cost / (
         SELECT COALESCE(SUM(l2.hectares), 1) 
         FROM lots l2 
         JOIN fields f2 ON l2.field_id = f2.id 
         WHERE f2.project_id = p.id AND l2.deleted_at IS NULL
       ) * l.hectares)) / NULLIF(l.hectares, 0)
    WHEN f.lease_type_id = 3 THEN -- ARRIENDO FIJO
      COALESCE(f.lease_type_value, 0)
    WHEN f.lease_type_id = 4 THEN -- ARRIENDO FIJO + % INGRESO NETO
      COALESCE(f.lease_type_value, 0) + 
      ((COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0) / NULLIF(l.hectares, 0))
    ELSE 0 
  END AS rent_per_ha,
  
  -- CORRECCIÓN: Total activo por hectárea (Costo por ha + arriendo + costo administrativo)
  (
    -- Costo por hectárea (labor + supplies)
    CASE 
      WHEN COALESCE(s.sowed_area, 0) > 0 
      THEN (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) / s.sowed_area
      ELSE 0 
    END +
    -- Arriendo por hectárea
    CASE 
      WHEN f.lease_type_id = 1 THEN -- % INGRESO NETO
        (COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0) / NULLIF(l.hectares, 0)
      WHEN f.lease_type_id = 2 THEN -- % UTILIDAD
        (COALESCE(f.lease_type_percent, 0) / 100.0) * 
        (COALESCE(in_net.income_net_total, 0) - 
         (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) - 
         (p.admin_cost / (
           SELECT COALESCE(SUM(l2.hectares), 1) 
           FROM lots l2 
           JOIN fields f2 ON l2.field_id = f2.id 
           WHERE f2.project_id = p.id AND l2.deleted_at IS NULL
         ) * l.hectares)) / NULLIF(l.hectares, 0)
      WHEN f.lease_type_id = 3 THEN -- ARRIENDO FIJO
        COALESCE(f.lease_type_value, 0)
      WHEN f.lease_type_id = 4 THEN -- ARRIENDO FIJO + % INGRESO NETO
        COALESCE(f.lease_type_value, 0) + 
        ((COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0) / NULLIF(l.hectares, 0))
      ELSE 0 
    END +
    -- Costo administrativo por hectárea
    (p.admin_cost / (
      SELECT COALESCE(SUM(l2.hectares), 1) 
      FROM lots l2 
      JOIN fields f2 ON l2.field_id = f2.id 
      WHERE f2.project_id = p.id AND l2.deleted_at IS NULL
    ))
  ) AS active_total_per_ha,
  
  -- CORRECCIÓN: Resultado operativo por hectárea (Ingreso neto - Activo total)
  CASE 
    WHEN l.hectares > 0 
    THEN (COALESCE(in_net.income_net_total, 0) / l.hectares) - (
      -- Costo por hectárea (labor + supplies)
      CASE 
        WHEN COALESCE(s.sowed_area, 0) > 0 
        THEN (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) / s.sowed_area
        ELSE 0 
      END +
      -- Arriendo por hectárea
      CASE 
        WHEN f.lease_type_id = 1 THEN -- % INGRESO NETO
          (COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0) / NULLIF(l.hectares, 0)
        WHEN f.lease_type_id = 2 THEN -- % UTILIDAD
          (COALESCE(f.lease_type_percent, 0) / 100.0) * 
          (COALESCE(in_net.income_net_total, 0) - 
           (COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)) - 
           (p.admin_cost / (
             SELECT COALESCE(SUM(l2.hectares), 1) 
             FROM lots l2 
             JOIN fields f2 ON l2.field_id = f2.id 
             WHERE f2.project_id = p.id AND l2.deleted_at IS NULL
           ) * l.hectares)) / NULLIF(l.hectares, 0)
        WHEN f.lease_type_id = 3 THEN -- ARRIENDO FIJO
          COALESCE(f.lease_type_value, 0)
        WHEN f.lease_type_id = 4 THEN -- ARRIENDO FIJO + % INGRESO NETO
          COALESCE(f.lease_type_value, 0) + 
          ((COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0) / NULLIF(l.hectares, 0))
        ELSE 0 
      END +
      -- Costo administrativo por hectárea
      (p.admin_cost / (
        SELECT COALESCE(SUM(l2.hectares), 1) 
        FROM lots l2 
        JOIN fields f2 ON l2.field_id = f2.id 
        WHERE f2.project_id = p.id AND l2.deleted_at IS NULL
      ))
    )
    ELSE 0 
  END AS operating_result_per_ha

FROM lots l
JOIN fields f ON f.id = l.field_id AND f.deleted_at IS NULL
JOIN projects p ON p.id = f.project_id AND p.deleted_at IS NULL
LEFT JOIN sowing s ON s.lot_id = l.id
LEFT JOIN harvesting h ON h.lot_id = l.id
LEFT JOIN direct_costs dc ON dc.lot_id = l.id
LEFT JOIN income_net in_net ON in_net.lot_id = l.id
LEFT JOIN crops pc ON pc.id = l.previous_crop_id AND pc.deleted_at IS NULL
LEFT JOIN crops cc_crop ON cc_crop.id = l.current_crop_id AND cc_crop.deleted_at IS NULL
WHERE l.deleted_at IS NULL;

-- Comentarios en español
COMMENT ON VIEW fix_lot_list IS 'Vista corregida para lista de lotes con cálculos correctos';
COMMENT ON COLUMN fix_lot_list.yield_tn_per_ha IS 'Rendimiento correcto: tons / hectares';
COMMENT ON COLUMN fix_lot_list.cost_usd_per_ha IS 'Costo USD por hectárea (labor + supplies)';
COMMENT ON COLUMN fix_lot_list.income_net_per_ha IS 'Ingreso neto por hectárea (precio cultivo * rendimiento)';
