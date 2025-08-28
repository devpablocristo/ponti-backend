-- =======================
-- DASHBOARD FULL VIEW - CORREGIDA
-- =======================
CREATE OR REPLACE VIEW dashboard_full_view AS
WITH

-- =======================
-- HECTÁREAS TOTALES (por proyecto)
-- =======================
total_hectares AS (
  SELECT 
    p.id AS project_id,
    f.id AS field_id,
    COALESCE(SUM(l.hectares), 0) AS total_hectares
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL AND f.deleted_at IS NULL
  GROUP BY p.id, f.id
),

-- =======================
-- SIEMBRA (por campo específico)
-- =======================
sowing AS (
  SELECT 
    f.project_id,
    f.id AS field_id,
    COALESCE(SUM(CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END), 0) AS sowed_area
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- =======================
-- COSECHA (por campo específico)
-- =======================
harvest AS (
  SELECT 
    f.project_id,
    f.id AS field_id,
    COALESCE(SUM(CASE WHEN lb.category_id = 13 THEN w.effective_area ELSE 0 END), 0) AS harvested_area
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- =======================
-- COSTOS DE LABORES (por campo específico) - TOTAL
-- =======================
labor_costs AS (
  SELECT 
    f.project_id,
    f.id AS field_id,
    COALESCE(SUM(w.effective_area * lb.price), 0) AS labor_cost,
    (SELECT COALESCE(SUM(w2.effective_area * lb2.price), 0) 
     FROM workorders w2 
     JOIN lots l2 ON l2.id = w2.lot_id 
     JOIN fields f2 ON f2.id = l2.field_id 
     JOIN labors lb2 ON lb2.id = w2.labor_id 
     WHERE f2.project_id = f.project_id AND w2.deleted_at IS NULL AND l2.deleted_at IS NULL AND lb2.deleted_at IS NULL) AS project_labor_cost
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  LEFT JOIN workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- =======================
-- COSTOS DE INSUMOS (por campo específico) - TOTAL
-- =======================
supply_costs AS (
  SELECT 
    f.project_id,
    f.id AS field_id,
    0 AS supply_cost,
    (SELECT COALESCE(SUM(sm2.quantity * s2.price), 0) 
     FROM supply_movements sm2 
     JOIN supplies s2 ON s2.id = sm2.supply_id 
     WHERE sm2.project_id = f.project_id AND sm2.deleted_at IS NULL AND s2.deleted_at IS NULL) AS project_supply_cost
  FROM fields f
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- =======================
-- INGRESOS NETOS (por campo específico)
-- =======================
income_net AS (
  SELECT 
    f.project_id,
    f.id AS field_id,
    0 AS income_net,
    (SELECT COALESCE(SUM(cc2.net_price), 0) 
     FROM crop_commercializations cc2 
     WHERE cc2.project_id = f.project_id AND cc2.deleted_at IS NULL) AS project_total_income
  FROM fields f
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- =======================
-- COSTOS DE CULTIVOS (por campo específico)
-- =======================
crop_costs AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    c.name AS crop_name,
    SUM(l.hectares) AS crop_hectares,
    COALESCE(SUM(CASE WHEN w.id IS NOT NULL AND lb.id IS NOT NULL
                      THEN lb.price * w.effective_area ELSE 0 END),0) AS crop_total_cost,
    (SELECT SUM(l2.hectares) FROM lots l2 JOIN fields f2 ON f2.id = l2.field_id WHERE f2.project_id = f.project_id AND l2.deleted_at IS NULL) AS total_hectares_project
  FROM fields f
  JOIN lots  l ON l.field_id = f.id
  JOIN crops c ON c.id = l.current_crop_id
  LEFT JOIN workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN labors     lb ON lb.id    = w.labor_id AND lb.deleted_at IS NULL
  WHERE f.deleted_at IS NULL AND l.deleted_at IS NULL AND c.deleted_at IS NULL
  GROUP BY f.project_id, f.id, l.current_crop_id, c.name
),

-- =======================
-- DESGLOSE DE CULTIVOS (por campo específico)
-- =======================
crop_breakdown AS (
  SELECT project_id, field_id,
         jsonb_object_agg(
           crop_name,
           jsonb_build_object(
             'hectares',     crop_hectares,
             'total_cost',   crop_total_cost,
             'cost_per_ha',  CASE WHEN crop_hectares>0 THEN ROUND((crop_total_cost/crop_hectares)::numeric, 3) ELSE 0 END,
             'rotation_pct', CASE WHEN total_hectares_project>0 THEN ROUND(((crop_hectares/total_hectares_project)*100)::numeric, 3) ELSE 0 END
           )
         ) AS crops_breakdown
  FROM crop_costs
  GROUP BY project_id, field_id
),

-- =======================
-- PRIMER / ÚLTIMO ORDEN
-- =======================
first_order AS (
  SELECT f.project_id, f.id AS field_id,
         MIN(w.date)   AS first_order_date,
         MIN(w.number) AS first_order_number
  FROM workorders w
  JOIN lots   l ON l.id = w.lot_id
  JOIN fields f ON f.id = l.field_id
  WHERE w.deleted_at IS NULL AND l.deleted_at IS NULL AND f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

last_order AS (
  SELECT f.project_id, f.id AS field_id,
         MAX(w.date)   AS last_order_date,
         MAX(w.number) AS last_order_number
  FROM workorders w
  JOIN lots   l ON l.id = w.lot_id
  JOIN fields f ON f.id = l.field_id
  WHERE w.deleted_at IS NULL AND l.deleted_at IS NULL AND f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- =======================
-- ÚLTIMO CONTEO DE STOCK
-- =======================
last_stock_count AS (
  SELECT p.id AS project_id, MAX(s.updated_at) AS last_stock_count_date
  FROM projects p
  JOIN supplies s ON s.project_id = p.id
  WHERE p.deleted_at IS NULL AND s.deleted_at IS NULL
  GROUP BY p.id
),

-- =======================
-- STOCK (disponible - consumido por proyecto)
-- =======================
stock_calculation AS (
  SELECT 
    p.id AS project_id,
    f.id AS field_id,
    COALESCE((SELECT SUM(s2.price) FROM supplies s2 WHERE s2.project_id = p.id AND s2.deleted_at IS NULL), 0) AS stock_value
  FROM projects p
  CROSS JOIN fields f
  WHERE f.project_id = p.id AND f.deleted_at IS NULL AND p.deleted_at IS NULL
),

-- =======================
-- ALQUILER (simplificado - hardcodeado por ahora)
-- =======================
rent_calculation AS (
  SELECT 
    p.id AS project_id, 
    f.id AS field_id,
    0 AS rent_amount
  FROM projects p
  CROSS JOIN fields f
  WHERE f.project_id = p.id AND f.deleted_at IS NULL AND p.deleted_at IS NULL
),

-- =======================
-- COSTOS ADMINISTRATIVOS
-- =======================
admin_costs AS (
  SELECT 
    p.id AS project_id,
    f.id AS field_id,
    COALESCE(p.admin_cost, 0) AS admin_cost
  FROM projects p
  CROSS JOIN fields f
  WHERE f.project_id = p.id AND f.deleted_at IS NULL AND p.deleted_at IS NULL
),

-- =======================
-- PRESUPUESTO (hardcodeado por ahora)
-- =======================
budget_costs AS (
  SELECT 
    p.id AS project_id,
    f.id AS field_id,
    2000.0 AS budget_cost
  FROM projects p
  CROSS JOIN fields f
  WHERE f.project_id = p.id AND f.deleted_at IS NULL AND p.deleted_at IS NULL
)

-- =======================
-- SELECT PRINCIPAL
-- =======================
SELECT 
  -- IDs
  p.campaign_id,
  f.project_id,
  p.customer_id,
  f.id AS field_id,
  
  -- Hectáreas y progreso
  th.total_hectares,
  s.sowed_area,
  h.harvested_area,
  CASE WHEN th.total_hectares > 0 THEN ROUND(((s.sowed_area / th.total_hectares) * 100)::numeric, 3) ELSE 0 END AS sowing_progress_pct,
  CASE WHEN th.total_hectares > 0 THEN ROUND(((h.harvested_area / th.total_hectares) * 100)::numeric, 3) ELSE 0 END AS harvest_progress_pct,
  
  -- VALORES DESCOMPUESTOS PARA AVANCE DE SIEMBRA
  s.sowed_area AS sowing_hectares,
  th.total_hectares AS total_hectares_for_sowing,
  
  -- VALORES DESCOMPUESTOS PARA AVANCE DE COSECHA
  h.harvested_area AS harvest_hectares,
  th.total_hectares AS total_hectares_for_harvest,
  
  -- Costos
  (lc.labor_cost + sc.supply_cost) AS executed_costs,
  bc.budget_cost AS budget_costs,
  inc.income_net,
  (lc.labor_cost + sc.supply_cost) AS total_costs,
  
  -- Desglose de costos
  jsonb_build_object(
    'labors', lc.labor_cost,
    'supplies', sc.supply_cost
  ) AS contribution_details,
  
  -- Costos en USD
  lc.labor_cost AS labors_cost_usd,
  sc.supply_cost AS inputs_cost_usd,
  (lc.labor_cost + sc.supply_cost) AS executed_cost_usd,
  bc.budget_cost AS budget_cost_usd,
  
  -- Progreso de costos
  CASE WHEN bc.budget_cost > 0 THEN ROUND((((lc.labor_cost + sc.supply_cost) / bc.budget_cost) * 100)::numeric, 3) ELSE 0 END AS costs_progress_pct,
  
  -- Ingresos y resultados
  inc.income_net AS income_net_total_usd,
  ac.admin_cost AS admin_total_usd,
  rc.rent_amount AS rent_total_usd,
  (inc.income_net - (lc.labor_cost + sc.supply_cost) - ac.admin_cost - rc.rent_amount) AS operating_result_usd,
  
  -- Resultado operativo porcentual
  CASE WHEN (COALESCE(lc.project_labor_cost,0) + COALESCE(sc.project_supply_cost,0)) > 0 
       THEN ROUND((((inc.project_total_income - (COALESCE(lc.project_labor_cost,0) + COALESCE(sc.project_supply_cost,0))) / (COALESCE(lc.project_labor_cost,0) + COALESCE(sc.project_supply_cost,0))) * 100)::numeric, 3)
       ELSE 0 END AS operating_result_pct,
  
  -- Costos invertidos
  (lc.project_labor_cost + sc.project_supply_cost) AS invested_cost_usd,
  
  -- Stock
  st.stock_value AS stock_usd,
  
  -- Aportes (hardcodeado por ahora)
  100.0 AS investor_contribution_pct,
  jsonb_build_object(
    'investor1', 100.0
  ) AS contribution_breakdown,
  
  -- Cultivos
  cb.crops_breakdown,
  jsonb_build_object(
    'crops', jsonb_build_array(
      jsonb_build_object(
        'name', 'Maíz',
        'hectares', '0',
        'total_cost', '0',
        'cost_per_ha', '0',
        'rotation_pct', '0'
      )
    ),
    'total', jsonb_build_object(
      'hectares', '0',
      'rotation_pct', '100',
      'cost_per_hectare', '0'
    )
  ) AS crops_details,
  
  -- Totales de cultivos
  0 AS crops_total_hectares,
  100.0 AS crops_total_rotation_pct,
  0 AS crops_total_cost_per_hectare,
  
  -- Rendimiento
  0.0 AS yield_per_hectare,
  CASE WHEN th.total_hectares > 0 THEN ROUND(((lc.labor_cost + sc.supply_cost) / th.total_hectares)::numeric, 3) ELSE 0 END AS total_cost_per_hectare,
  
  -- Fechas
  fo.first_order_date,
  fo.first_order_number,
  lo.last_order_date,
  lo.last_order_number,
  lsc.last_stock_count_date,
  
  -- Gestión
  inc.project_total_income AS mgmt_income_usd,
  (lc.project_labor_cost + sc.project_supply_cost + ac.admin_cost) AS mgmt_total_costs_usd,
  (inc.project_total_income - (lc.project_labor_cost + sc.project_supply_cost) - ac.admin_cost) AS mgmt_operating_result_usd,
  
  -- Resultado operativo de gestión
  CASE WHEN (lc.project_labor_cost + sc.project_supply_cost + ac.admin_cost) > 0 
       THEN ROUND((((inc.project_total_income - (lc.project_labor_cost + sc.project_supply_cost) - ac.admin_cost) / (lc.project_labor_cost + sc.project_supply_cost + ac.admin_cost)) * 100)::numeric, 3)
       ELSE 0 END AS mgmt_operating_result_pct,

  -- ===== BALANCE DE GESTIÓN DETALLADO - SIMPLIFICADO =====
  -- Direct costs
  (lc.labor_cost + sc.supply_cost) AS direct_costs_executed_usd,
  (lc.project_labor_cost + sc.project_supply_cost) AS direct_costs_invested_usd,
  st.stock_value AS direct_costs_stock_usd,
  th.total_hectares AS direct_costs_hectares,
  
  -- Seed
  0 AS seed_executed_usd,
  0 AS seed_invested_usd,
  0 AS seed_stock_usd,
  0 AS seed_hectares,
  
  -- Supplies
  sc.supply_cost AS supplies_executed_usd,
  sc.project_supply_cost AS supplies_invested_usd,
  st.stock_value AS supplies_stock_usd,
  th.total_hectares AS supplies_hectares,
  
  -- Labors
  lc.labor_cost AS labors_executed_usd,
  lc.project_labor_cost AS labors_invested_usd,
  0 AS labors_stock_usd,
  th.total_hectares AS labors_hectares,
  
  -- Rent
  rc.rent_amount AS rent_executed_usd,
  0 AS rent_invested_usd,
  0 AS rent_stock_usd,
  th.total_hectares AS rent_hectares,
  
  -- Structure
  ac.admin_cost AS structure_executed_usd,
  0 AS structure_invested_usd,
  0 AS structure_stock_usd,
  th.total_hectares AS structure_hectares

FROM fields f
JOIN projects p ON p.id = f.project_id
JOIN total_hectares th ON th.field_id = f.id
JOIN sowing s ON s.field_id = f.id
JOIN harvest h ON h.field_id = f.id
JOIN labor_costs lc ON lc.field_id = f.id
JOIN supply_costs sc ON sc.field_id = f.id
JOIN income_net inc ON inc.field_id = f.id
JOIN crop_breakdown cb ON cb.field_id = f.id
JOIN first_order fo ON fo.field_id = f.id
JOIN last_order lo ON lo.field_id = f.id
JOIN last_stock_count lsc ON lsc.project_id = f.project_id
JOIN stock_calculation st ON st.field_id = f.id
JOIN rent_calculation rc ON rc.field_id = f.id
JOIN admin_costs ac ON ac.field_id = f.id
JOIN budget_costs bc ON bc.field_id = f.id
WHERE f.deleted_at IS NULL
ORDER BY p.campaign_id, f.project_id, p.customer_id, f.id;
