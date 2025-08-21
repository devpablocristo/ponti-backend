CREATE OR REPLACE VIEW dashboard_full_view AS
WITH
-- =======================
-- PROGRESO (Siembra/Cosecha)
-- =======================
sowing AS (
  SELECT f.project_id, f.id AS field_id, SUM(w.effective_area) AS sowed_area
  FROM workorders w
  JOIN labors  lb ON lb.id = w.labor_id
  JOIN lots    l  ON l.id  = w.lot_id
  JOIN fields  f  ON f.id  = l.field_id
  WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL
    AND l.deleted_at IS NULL AND f.deleted_at IS NULL
    AND lb.category_id = 1
  GROUP BY f.project_id, f.id
),
harvest AS (
  SELECT f.project_id, f.id AS field_id, SUM(w.effective_area) AS harvested_area
  FROM workorders w
  JOIN labors  lb ON lb.id = w.labor_id
  JOIN lots    l  ON l.id  = w.lot_id
  JOIN fields  f  ON f.id  = l.field_id
  WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL
    AND l.deleted_at IS NULL AND f.deleted_at IS NULL
    AND lb.category_id = 2
  GROUP BY f.project_id, f.id
),

-- =======================
-- COSTOS DIRECTOS (Labores + Insumos)
-- =======================
labor_costs AS (
  SELECT f.project_id, f.id AS field_id, SUM(lb.price * w.effective_area) AS labor_cost
  FROM workorders w
  JOIN labors  lb ON lb.id = w.labor_id
  JOIN lots    l  ON l.id  = w.lot_id
  JOIN fields  f  ON f.id  = l.field_id
  WHERE w.deleted_at IS NULL AND lb.deleted_at IS NULL
    AND l.deleted_at IS NULL AND f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),
supply_costs AS (
  SELECT f.project_id, f.id AS field_id,
         SUM(wi.final_dose * s.price * w.effective_area) AS supply_cost
  FROM workorders w
  JOIN lots l ON l.id = w.lot_id
  JOIN fields f ON f.id = l.field_id
  JOIN workorder_items wi ON w.id = wi.workorder_id
  JOIN supplies s ON s.id = wi.supply_id
  WHERE w.deleted_at IS NULL AND l.deleted_at IS NULL AND f.deleted_at IS NULL
    AND wi.deleted_at IS NULL AND s.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- =======================
-- SEMILLAS vs NO SEMILLAS
-- =======================
seed_costs AS (
  SELECT f.project_id, f.id AS field_id,
         SUM(wi.final_dose * s.price * w.effective_area) AS seed_cost
  FROM workorders w
  JOIN lots l ON l.id = w.lot_id
  JOIN fields f ON f.id = l.field_id
  JOIN workorder_items wi ON w.id = wi.workorder_id
  JOIN supplies s ON s.id = wi.supply_id
  JOIN types t ON t.id = s.type_id
  WHERE w.deleted_at IS NULL AND l.deleted_at IS NULL AND f.deleted_at IS NULL
    AND wi.deleted_at IS NULL AND s.deleted_at IS NULL AND t.deleted_at IS NULL
    AND t.id = 1
  GROUP BY f.project_id, f.id
),
non_seed_supply_costs AS (
  SELECT f.project_id, f.id AS field_id,
         SUM(wi.final_dose * s.price * w.effective_area) AS non_seed_supply_cost
  FROM workorders w
  JOIN lots l ON l.id = w.lot_id
  JOIN fields f ON f.id = l.field_id
  JOIN workorder_items wi ON w.id = wi.workorder_id
  JOIN supplies s ON s.id = wi.supply_id
  JOIN types t ON t.id = s.type_id
  WHERE w.deleted_at IS NULL AND l.deleted_at IS NULL AND f.deleted_at IS NULL
    AND wi.deleted_at IS NULL AND s.deleted_at IS NULL AND t.deleted_at IS NULL
    AND t.id <> 1
  GROUP BY f.project_id, f.id
),

-- =======================
-- TONELADAS y HECTÁREAS
-- =======================
harvest_tons AS (
  SELECT f.project_id, f.id AS field_id, SUM(l.tons) AS total_tons
  FROM lots l
  JOIN fields f ON f.id = l.field_id
  WHERE l.deleted_at IS NULL AND f.deleted_at IS NULL AND l.tons IS NOT NULL
  GROUP BY f.project_id, f.id
),
total_hectares AS (
  SELECT f.project_id, f.id AS field_id, COALESCE(SUM(l.hectares),0) AS total_hectares
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- =======================
-- COSTO PRESUPUESTADO (placeholder)
-- =======================
budget_costs AS (
  SELECT f.project_id, f.id AS field_id, 2000.0 AS budget_cost
  FROM fields f
  WHERE f.deleted_at IS NULL
),

-- =======================
-- INGRESOS (evitar duplicados por lote)
-- =======================
field_crops AS (
  SELECT f.project_id, f.id AS field_id, l.current_crop_id
  FROM fields f
  JOIN lots l ON l.field_id = f.id
  WHERE f.deleted_at IS NULL AND l.deleted_at IS NULL
  GROUP BY f.project_id, f.id, l.current_crop_id
),
cc_by_crop AS (
  SELECT project_id, crop_id, COALESCE(SUM(net_price),0) AS income_net_total
  FROM crop_commercializations
  WHERE deleted_at IS NULL
  GROUP BY project_id, crop_id
),
income_net AS (
  SELECT fc.project_id, fc.field_id, COALESCE(SUM(cc.income_net_total),0) AS income_net_total
  FROM field_crops fc
  JOIN cc_by_crop cc
    ON cc.project_id = fc.project_id
   AND cc.crop_id    = fc.current_crop_id
  GROUP BY fc.project_id, fc.field_id
),

-- =======================
-- CONTRIBUCIONES DE INVERSORES (clave estable)
-- =======================
investor_contributions AS (
  SELECT p.id AS project_id,
         jsonb_object_agg(
           pi.investor_id,
           jsonb_build_object('name', i.name, 'percentage', pi.percentage)
         ) AS contribution_breakdown
  FROM projects p
  JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
  JOIN investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.id
),

-- =======================
-- INCIDENCIA DE COSTOS POR CULTIVO (con denominador por campo)
-- =======================
crop_costs AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    l.current_crop_id,
    c.name AS crop_name,
    l.hectares AS crop_hectares,
    COALESCE(SUM(CASE WHEN w.id IS NOT NULL AND lb.id IS NOT NULL
                      THEN lb.price * w.effective_area ELSE 0 END),0) AS crop_total_cost,
    SUM(l.hectares) OVER (PARTITION BY f.project_id, f.id) AS total_hectares_field
  FROM fields f
  JOIN lots  l ON l.field_id = f.id
  JOIN crops c ON c.id = l.current_crop_id
  LEFT JOIN workorders w ON w.lot_id = l.id AND w.deleted_at IS NULL
  LEFT JOIN labors     lb ON lb.id    = w.labor_id AND lb.deleted_at IS NULL
  WHERE f.deleted_at IS NULL AND l.deleted_at IS NULL AND c.deleted_at IS NULL
  GROUP BY f.project_id, f.id, l.current_crop_id, c.name, l.hectares
),
crop_breakdown AS (
  SELECT project_id, field_id,
         jsonb_object_agg(
           crop_name,
           jsonb_build_object(
             'hectares',     crop_hectares,
             'total_cost',   crop_total_cost,
             'cost_per_ha',  CASE WHEN crop_hectares>0 THEN crop_total_cost/crop_hectares ELSE 0 END,
             'rotation_pct', CASE WHEN total_hectares_field>0 THEN (crop_hectares/total_hectares_field)*100 ELSE 0 END
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
-- STOCK (pre-aggregado por proyecto para evitar duplicados)
-- =======================
stock_by_project AS (
  SELECT s.project_id, COALESCE(SUM(s.price),0) AS stock_value
  FROM supplies s
  WHERE s.deleted_at IS NULL
  GROUP BY s.project_id
),
stock_calculation AS (
  SELECT f.project_id, f.id AS field_id, COALESCE(sb.stock_value,0) AS stock_value
  FROM fields f
  LEFT JOIN stock_by_project sb ON sb.project_id = f.project_id
  WHERE f.deleted_at IS NULL
),

-- =======================
-- ALQUILER (placeholder)
-- =======================
rent_calculation AS (
  SELECT f.project_id, f.id AS field_id, 0.0 AS rent_total
  FROM fields f
  WHERE f.deleted_at IS NULL
)

SELECT
  -- IDENTIFICADORES (placeholder parametrizable)
  1 AS campaign_id,
  th.project_id,
  1 AS customer_id,
  th.field_id,

  -- HECTÁREAS / PROGRESO
  COALESCE(th.total_hectares, 0) AS total_hectares,
  COALESCE(s.sowed_area, 0)     AS sowed_area,
  COALESCE(h.harvested_area, 0) AS harvested_area,

  CASE WHEN COALESCE(th.total_hectares,0)>0 THEN (COALESCE(s.sowed_area,0)/th.total_hectares)*100 ELSE 0 END AS sowing_progress_pct,
  COALESCE(s.sowed_area,0)       AS sowing_hectares,
  COALESCE(th.total_hectares,0)  AS total_hectares_for_sowing,

  CASE WHEN COALESCE(th.total_hectares,0)>0 THEN (COALESCE(h.harvested_area,0)/th.total_hectares)*100 ELSE 0 END AS harvest_progress_pct,
  COALESCE(h.harvested_area,0)   AS harvest_hectares,
  COALESCE(th.total_hectares,0)  AS total_hectares_for_harvest,

  -- COSTOS vs PRESUPUESTO
  CASE WHEN COALESCE(bc.budget_cost,0)>0
       THEN ((COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0))/bc.budget_cost)*100
       ELSE 0 END AS costs_progress_pct,
  (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)) AS executed_costs,
  COALESCE(bc.budget_cost,0) AS budget_costs,

  -- CONTRIBUCIONES
  CASE
    WHEN ic.contribution_breakdown IS NOT NULL AND ic.contribution_breakdown <> '{}'::jsonb
    THEN (SELECT COALESCE(SUM((value->>'percentage')::int),0) FROM jsonb_each(ic.contribution_breakdown))
    ELSE 0
  END AS investor_contribution_pct,
  COALESCE(ic.contribution_breakdown, '{}'::jsonb) AS contribution_details,
  COALESCE(ic.contribution_breakdown, '{}'::jsonb) AS contribution_breakdown,

  -- RESULTADO OPERATIVO (%)
  CASE WHEN (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0))>0
       THEN ((COALESCE(inc.income_net_total,0)-(COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)))
             /(COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)))*100
       ELSE 0 END AS operating_result_pct,

  -- BASES
  COALESCE(inc.income_net_total,0) AS income_net,
  (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)) AS total_costs,

  -- DESGLOSADO DIRECTO
  COALESCE(lc.labor_cost,0)           AS labors_cost_usd,
  COALESCE(sc.supply_cost,0)          AS inputs_cost_usd,
  COALESCE(seed.seed_cost,0)          AS seed_cost_usd,
  COALESCE(ns.non_seed_supply_cost,0) AS non_seed_supply_cost_usd,
  (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)) AS executed_cost_usd,

  -- PRESUPUESTO e INGRESOS
  COALESCE(bc.budget_cost,0)   AS budget_cost_usd,
  COALESCE(inc.income_net_total,0) AS income_net_total_usd,

  -- ADMINISTRACIÓN / ALQUILER
  COALESCE(p.admin_cost,0) AS admin_total_usd,
  COALESCE(rc.rent_total,0) AS rent_total_usd,

  -- RESULTADO (USD)
  (COALESCE(inc.income_net_total,0)-(COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0))) AS operating_result_usd,

  -- INVERTIDO / STOCK
  (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)) AS invested_cost_usd,
  COALESCE(stock_calc.stock_value,0) AS stock_usd,

  -- COMPATIBILIDAD DOMINIO
  COALESCE(inc.income_net_total,0) AS mgmt_income_usd,
  (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)+COALESCE(p.admin_cost,0)) AS mgmt_total_costs_usd,
  (COALESCE(inc.income_net_total,0)-(COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)+COALESCE(p.admin_cost,0))) AS mgmt_operating_result_usd,
  CASE WHEN (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)+COALESCE(p.admin_cost,0))>0
       THEN ((COALESCE(inc.income_net_total,0)-(COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)+COALESCE(p.admin_cost,0)))
            /(COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0)+COALESCE(p.admin_cost,0)))*100
       ELSE 0 END AS mgmt_operating_result_pct,

  -- INCIDENCIA POR CULTIVO
  COALESCE(cb.crops_breakdown,'{}'::jsonb) AS crops_breakdown,
  COALESCE(cb.crops_breakdown,'[]'::jsonb) AS crops_details,
  COALESCE(th.total_hectares,0) AS crops_total_hectares,
  100.0 AS crops_total_rotation_pct,
  CASE WHEN COALESCE(th.total_hectares,0)>0
       THEN (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0))/th.total_hectares
       ELSE 0 END AS crops_total_cost_per_hectare,

  -- RENDIMIENTO / COSTOS por HA
  CASE WHEN COALESCE(h.harvested_area,0)>0
       THEN COALESCE(ht.total_tons,0)/h.harvested_area
       ELSE 0 END AS yield_per_hectare,
  CASE WHEN COALESCE(th.total_hectares,0)>0
       THEN (COALESCE(lc.labor_cost,0)+COALESCE(sc.supply_cost,0))/th.total_hectares
       ELSE 0 END AS total_cost_per_hectare,

  -- INDICADORES
  fo.first_order_date, fo.first_order_number,
  lo.last_order_date,  lo.last_order_number,
  lsc.last_stock_count_date

FROM total_hectares th
JOIN projects p ON p.id = th.project_id
LEFT JOIN sowing            s    ON s.project_id    = th.project_id AND s.field_id    = th.field_id
LEFT JOIN harvest           h    ON h.project_id    = th.project_id AND h.field_id    = th.field_id
LEFT JOIN labor_costs       lc   ON lc.project_id   = th.project_id AND lc.field_id   = th.field_id
LEFT JOIN supply_costs      sc   ON sc.project_id   = th.project_id AND sc.field_id   = th.field_id
LEFT JOIN seed_costs        seed ON seed.project_id = th.project_id AND seed.field_id = th.field_id
LEFT JOIN non_seed_supply_costs ns ON ns.project_id = th.project_id AND ns.field_id   = th.field_id
LEFT JOIN harvest_tons      ht   ON ht.project_id   = th.project_id AND ht.field_id   = th.field_id
LEFT JOIN budget_costs      bc   ON bc.project_id   = th.project_id AND bc.field_id   = th.field_id
LEFT JOIN income_net        inc  ON inc.project_id  = th.project_id AND inc.field_id  = th.field_id
LEFT JOIN investor_contributions ic ON ic.project_id = th.project_id
LEFT JOIN crop_breakdown    cb   ON cb.project_id   = th.project_id AND cb.field_id   = th.field_id
LEFT JOIN stock_calculation stock_calc ON stock_calc.project_id = th.project_id AND stock_calc.field_id = th.field_id
LEFT JOIN first_order       fo   ON fo.project_id   = th.project_id AND fo.field_id   = th.field_id
LEFT JOIN last_order        lo   ON lo.project_id   = th.project_id AND lo.field_id   = th.field_id
LEFT JOIN last_stock_count  lsc  ON lsc.project_id  = th.project_id
LEFT JOIN rent_calculation  rc   ON rc.project_id   = th.project_id AND rc.field_id   = th.field_id
WHERE p.deleted_at IS NULL;
