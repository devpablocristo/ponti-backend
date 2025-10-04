-- ========================================
-- MIGRACIÓN 000121: CREATE DASHBOARD SEPARATED VIEWS (UP)
-- ========================================
-- 
-- Propósito: Crear vistas independientes del dashboard (una por card/módulo)
-- Dependencias: Requiere v3_core_ssot (000113), v3_workorder_ssot (000114), v3_lot_ssot (000115),
--               v3_dashboard_ssot (000116), v3_workorder_metrics (000117), v3_workorder_list (000118),
--               v3_lot_metrics (000119), y v3_lot_list (000120)
-- Arquitectura: 5 cards + 3 módulos = 8 vistas
-- Fecha: 2025-10-04
-- Autor: Sistema
-- 
-- Nota: Vistas SOLO ensamblan, NO calculan (usan funciones SSOT)

BEGIN;

-- ========================================
-- ELIMINAR VISTAS ANTIGUAS
-- ========================================
DROP VIEW IF EXISTS v3_dashboard CASCADE;
DROP VIEW IF EXISTS v3_dashboard_contributions_progress CASCADE;
DROP VIEW IF EXISTS v3_dashboard_management_balance CASCADE;
DROP VIEW IF EXISTS v3_dashboard_crop_incidence CASCADE;

-- ========================================
-- SECCIÓN 1: VISTAS DE LAS 5 CARDS
-- ========================================

-- ========================================
-- CARD 1: v3_dashboard_sowing_metrics
-- ========================================
-- Propósito: Métricas de avance de siembra
-- Campos: 3 (hectares, total_hectares, progress_pct)
-- Funciones SSOT: 1 (v3_core_ssot.percentage)
CREATE OR REPLACE VIEW public.v3_dashboard_sowing_metrics AS
WITH lot_metrics_base AS (
  SELECT
    project_id,
    hectares,
    sowed_area_ha
  FROM public.v3_lot_metrics
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- Superficie sembrada
  COALESCE(SUM(lm.sowed_area_ha), 0)::double precision AS sowing_hectares,
  -- Superficie total
  COALESCE(SUM(lm.hectares), 0)::double precision AS sowing_total_hectares,
  -- Porcentaje de progreso (usa v3_core_ssot)
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.sowed_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS sowing_progress_pct
  
FROM public.projects p
LEFT JOIN lot_metrics_base lm ON lm.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

COMMENT ON VIEW public.v3_dashboard_sowing_metrics IS 'Card 1: Métricas de avance de siembra por proyecto';

-- ========================================
-- CARD 2: v3_dashboard_harvest_metrics
-- ========================================
-- Propósito: Métricas de avance de cosecha
-- Campos: 3 (hectares, total_hectares, progress_pct)
-- Funciones SSOT: 1 (v3_core_ssot.percentage)
CREATE OR REPLACE VIEW public.v3_dashboard_harvest_metrics AS
WITH lot_metrics_base AS (
  SELECT
    project_id,
    hectares,
    harvested_area_ha
  FROM public.v3_lot_metrics
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- Superficie cosechada
  COALESCE(SUM(lm.harvested_area_ha), 0)::double precision AS harvest_hectares,
  -- Superficie total
  COALESCE(SUM(lm.hectares), 0)::double precision AS harvest_total_hectares,
  -- Porcentaje de progreso (usa v3_core_ssot)
  v3_core_ssot.percentage(
    COALESCE(SUM(lm.harvested_area_ha), 0)::numeric,
    COALESCE(SUM(lm.hectares), 0)::numeric
  ) AS harvest_progress_pct
  
FROM public.projects p
LEFT JOIN lot_metrics_base lm ON lm.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id;

COMMENT ON VIEW public.v3_dashboard_harvest_metrics IS 'Card 2: Métricas de avance de cosecha por proyecto';

-- ========================================
-- CARD 3: v3_dashboard_costs_metrics
-- ========================================
-- Propósito: Métricas de avance de costos
-- Campos: 3 (executed_costs, budget_cost, progress_pct)
-- Funciones SSOT: 1 (v3_core_ssot.percentage)
-- ⚠️ NOTA: budget_cost_usd está HARDCODEADO (admin_cost * 10) de forma TEMPORAL
-- TODO: Definir fórmula correcta para calcular el presupuesto dinámicamente
CREATE OR REPLACE VIEW public.v3_dashboard_costs_metrics AS
WITH lot_metrics_base AS (
  SELECT
    project_id,
    sowed_area_ha,
    direct_cost_per_ha_usd
  FROM public.v3_lot_metrics
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- Costo ejecutado (promedio ponderado por ha sembrada)
  COALESCE(
    SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 
    0
  )::double precision AS executed_costs_usd,
  
  -- ⚠️ PRESUPUESTO HARDCODEADO TEMPORAL ⚠️
  -- TODO: Definir fórmula correcta para calcular el presupuesto dinámicamente
  -- Por ahora: admin_cost * 10 (valor temporal hasta definir cálculo)
  (p.admin_cost * 10)::double precision AS budget_cost_usd,
  
  -- Porcentaje de progreso (usa v3_core_ssot)
  v3_core_ssot.percentage(
    COALESCE(
      SUM(lm.direct_cost_per_ha_usd * lm.sowed_area_ha) / NULLIF(SUM(lm.sowed_area_ha), 0), 
      0
    )::numeric,
    (p.admin_cost * 10)::numeric  -- Mismo valor hardcodeado
  ) AS costs_progress_pct
  
FROM public.projects p
LEFT JOIN lot_metrics_base lm ON lm.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost;

COMMENT ON VIEW public.v3_dashboard_costs_metrics IS 'Card 3: Métricas de avance de costos (⚠️ budget_cost_usd HARDCODEADO TEMPORAL)';

-- ========================================
-- CARD 4: v3_dashboard_operating_result
-- ========================================
-- Propósito: Métricas de resultado operativo
-- Campos: 4 (income, result, total_costs, result_pct)
-- Funciones SSOT: 4 (income_net_total_for_lot, operating_result_total_for_project, direct_costs_total_for_project, renta_pct)
CREATE OR REPLACE VIEW public.v3_dashboard_operating_result AS
WITH lot_metrics_base AS (
  SELECT
    project_id,
    lot_id
  FROM public.v3_lot_metrics
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) as total_hectares
  FROM public.v3_lot_metrics
  GROUP BY project_id
)
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  
  -- Ingresos (suma de ingresos netos por lote - usa v3_core_ssot)
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(lm.lot_id)), 0) AS income_usd,
  -- Resultado operativo (usa v3_dashboard_ssot)
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  -- Total costos = directos + arriendo + admin
  (COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) + 
   COALESCE(p.admin_cost * ph.total_hectares, 0) + 
   COALESCE((SELECT f.lease_type_value * ph.total_hectares 
             FROM fields f 
             WHERE f.project_id = p.id AND f.deleted_at IS NULL 
             LIMIT 1), 0))::double precision AS operating_result_total_costs_usd,
  -- Porcentaje de rentabilidad (usa v3_lot_ssot.renta_pct)
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    (COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) + 
     COALESCE(p.admin_cost * ph.total_hectares, 0) + 
     COALESCE((SELECT f.lease_type_value * ph.total_hectares 
               FROM fields f 
               WHERE f.project_id = p.id AND f.deleted_at IS NULL 
               LIMIT 1), 0))::double precision
  ) AS operating_result_pct
  
FROM public.projects p
LEFT JOIN lot_metrics_base lm ON lm.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_operating_result IS 'Card 4: Métricas de resultado operativo por proyecto';

-- ========================================
-- CARD 5: v3_dashboard_contributions_progress
-- ========================================
-- Propósito: Avance de aportes por inversor
-- Campos: 4 (investor_id, investor_name, percentage, progress)
-- Funciones SSOT: 0 (datos directos de tablas)
CREATE OR REPLACE VIEW public.v3_dashboard_contributions_progress AS
SELECT
  p.id AS project_id,
  pi.investor_id AS investor_id,
  i.name AS investor_name,
  pi.percentage AS investor_percentage_pct,
  pi.percentage::numeric AS contributions_progress_pct
FROM public.projects p
JOIN public.project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
JOIN public.investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
WHERE p.deleted_at IS NULL;

COMMENT ON VIEW public.v3_dashboard_contributions_progress IS 'Card 5: Avance de aportes de inversores por proyecto';

-- ========================================
-- SECCIÓN 2: VISTAS DE MÓDULOS ADICIONALES (NO CARDS)
-- ========================================

-- ========================================
-- MÓDULO 6: v3_dashboard_operational_indicators
-- ========================================
-- Propósito: Fechas e indicadores operativos
-- Campos: 6 (start, end, closing, first_wo, last_wo, last_stock)
-- Funciones SSOT: 6 (todas de v3_dashboard_ssot + v3_core_ssot.calculate_campaign_closing_date)
CREATE OR REPLACE VIEW public.v3_dashboard_operational_indicators AS
SELECT
  p.id AS project_id,
  
  -- Fechas operativas (usan v3_dashboard_ssot - SIN CTEs)
  v3_dashboard_ssot.first_workorder_date_for_project(p.id) AS start_date,
  v3_dashboard_ssot.last_workorder_date_for_project(p.id) AS end_date,
  v3_core_ssot.calculate_campaign_closing_date(
    v3_dashboard_ssot.last_workorder_date_for_project(p.id)
  ) AS campaign_closing_date,
  
  -- IDs de workorders (usan v3_dashboard_ssot - SIN CTEs)
  v3_dashboard_ssot.first_workorder_id_for_project(p.id) AS first_workorder_id,
  v3_dashboard_ssot.last_workorder_id_for_project(p.id) AS last_workorder_id,
  
  -- Fecha último arqueo (usa v3_dashboard_ssot - SIN CTEs)
  v3_dashboard_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
  
FROM public.projects p
WHERE p.deleted_at IS NULL;

COMMENT ON VIEW public.v3_dashboard_operational_indicators IS 'Módulo: Fechas e indicadores operativos por proyecto';

-- ========================================
-- MÓDULO 7: v3_dashboard_management_balance
-- ========================================
-- Propósito: Balance de gestión (ejecutados/invertidos/stock)
-- Nota: Mantener estructura actual, actualizar para usar v3_dashboard_ssot
CREATE OR REPLACE VIEW public.v3_dashboard_management_balance AS
WITH lots_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id AS project_id,
    l.hectares AS hectares
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.deleted_at IS NULL
),
project_hectares AS (
  SELECT
    project_id,
    SUM(hectares) as total_hectares
  FROM lots_base
  GROUP BY project_id
)
SELECT
  p.id AS project_id,
  
  -- Ingresos y Resultado (usa v3_lot_ssot y v3_dashboard_ssot)
  COALESCE(SUM(v3_lot_ssot.income_net_total_for_lot(lb.lot_id)), 0) AS income_usd,
  v3_dashboard_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
  v3_lot_ssot.renta_pct(
    v3_dashboard_ssot.operating_result_total_for_project(p.id),
    (COALESCE(v3_dashboard_ssot.direct_costs_total_for_project(p.id), 0) + 
     COALESCE(p.admin_cost * ph.total_hectares, 0) + 
     COALESCE((SELECT f.lease_type_value * ph.total_hectares 
               FROM fields f 
               WHERE f.project_id = p.id AND f.deleted_at IS NULL 
               LIMIT 1), 0))::double precision
  ) AS operating_result_pct,
  
  -- Costos Directos (usa v3_dashboard_ssot)
  v3_dashboard_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
  COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM supply_movements sm
    JOIN supplies s ON s.id = sm.supply_id
    WHERE sm.project_id = p.id 
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial')
  ), 0) + COALESCE(SUM(v3_dashboard_ssot.labor_cost_for_lot_mb(lb.lot_id)), 0) AS costos_directos_invertidos_usd,
  (COALESCE((
    SELECT SUM(sm.quantity * s.price)
    FROM supply_movements sm
    JOIN supplies s ON s.id = sm.supply_id
    WHERE sm.project_id = p.id 
      AND sm.deleted_at IS NULL
      AND s.deleted_at IS NULL
      AND sm.movement_type IN ('Stock', 'Remito oficial')
  ), 0) + COALESCE(SUM(v3_dashboard_ssot.labor_cost_for_lot_mb(lb.lot_id)), 0)) - 
  v3_dashboard_ssot.direct_costs_total_for_project(p.id) AS costos_directos_stock_usd,
  
  -- Semillas (usa v3_dashboard_ssot._mb)
  COALESCE(SUM(v3_dashboard_ssot.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semillas_ejecutados_usd,
  v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
  v3_dashboard_ssot.seeds_invested_for_project_mb(p.id) - 
    COALESCE(SUM(v3_dashboard_ssot.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semillas_stock_usd,
  
  -- Agroquímicos (usa v3_dashboard_ssot._mb)
  COALESCE(SUM(v3_dashboard_ssot.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS agroquimicos_ejecutados_usd,
  v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
  v3_dashboard_ssot.agrochemicals_invested_for_project_mb(p.id) - 
    COALESCE(SUM(v3_dashboard_ssot.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS agroquimicos_stock_usd,
  
  -- Labores (usa v3_dashboard_ssot._mb)
  COALESCE(SUM(v3_dashboard_ssot.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_ejecutados_usd,
  COALESCE(SUM(v3_dashboard_ssot.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_invertidos_usd,
  
  -- Arriendo (según tipo de arriendo)
  COALESCE((
    SELECT CASE 
      WHEN f.lease_type_id IN (3, 4) THEN f.lease_type_value * ph.total_hectares
      ELSE 0
    END
    FROM public.fields f
    WHERE f.project_id = p.id AND f.deleted_at IS NULL
    LIMIT 1
  ), 0)::double precision AS arriendo_ejecutados_usd,
  COALESCE((SELECT f.lease_type_value * ph.total_hectares
            FROM public.fields f
            WHERE f.project_id = p.id AND f.deleted_at IS NULL
            LIMIT 1), 0)::double precision AS arriendo_invertidos_usd,
  
  -- Estructura
  COALESCE(p.admin_cost * ph.total_hectares, 0)::double precision AS estructura_ejecutados_usd,
  COALESCE(p.admin_cost * ph.total_hectares, 0)::double precision AS estructura_invertidos_usd,
  
  -- Costos calculados (para compatibilidad)
  COALESCE(SUM(v3_dashboard_ssot.supply_cost_seeds_for_lot_mb(lb.lot_id)), 0) AS semilla_cost,
  COALESCE(SUM(v3_dashboard_ssot.supply_cost_agrochemicals_for_lot_mb(lb.lot_id)), 0) AS insumos_cost,
  COALESCE(SUM(v3_dashboard_ssot.labor_cost_for_lot_mb(lb.lot_id)), 0) AS labores_cost

FROM public.projects p
LEFT JOIN lots_base lb ON lb.project_id = p.id
LEFT JOIN project_hectares ph ON ph.project_id = p.id
WHERE p.deleted_at IS NULL
GROUP BY 
  p.id, 
  p.admin_cost, 
  ph.total_hectares;

COMMENT ON VIEW public.v3_dashboard_management_balance IS 'Módulo 6: Balance de gestión con ejecutados/invertidos/stock';

-- ========================================
-- MÓDULO 8: v3_dashboard_crop_incidence
-- ========================================
-- Propósito: Incidencia de costos por cultivo
-- Nota: Mantener estructura actual, actualizar para usar v3_core_ssot y v3_dashboard_ssot
CREATE OR REPLACE VIEW public.v3_dashboard_crop_incidence AS
WITH lot_base AS (
  SELECT
    l.id AS lot_id,
    f.project_id AS project_id,
    l.current_crop_id AS current_crop_id,
    c.name AS crop_name,
    l.hectares AS hectares
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  WHERE l.deleted_at IS NULL AND l.hectares IS NOT NULL AND l.hectares > 0
),
project_totals AS (
  SELECT 
    project_id,
    SUM(hectares)::numeric AS total_project_hectares
  FROM lot_base
  GROUP BY project_id
),
by_crop AS (
  SELECT 
    project_id, 
    current_crop_id, 
    crop_name, 
    SUM(hectares)::numeric AS crop_hectares
  FROM lot_base
  WHERE current_crop_id IS NOT NULL
  GROUP BY project_id, current_crop_id, crop_name
),
crop_percentages AS (
  SELECT
    bc.project_id,
    bc.current_crop_id,
    bc.crop_name,
    bc.crop_hectares,
    pt.total_project_hectares,
    -- Calcular porcentaje base usando v3_core_ssot
    v3_core_ssot.percentage_rounded(bc.crop_hectares, pt.total_project_hectares) AS base_percentage,
    ROW_NUMBER() OVER (PARTITION BY bc.project_id ORDER BY bc.crop_name) AS crop_order,
    COUNT(*) OVER (PARTITION BY bc.project_id) AS total_crops
  FROM by_crop bc
  JOIN project_totals pt ON pt.project_id = bc.project_id
),
project_sums AS (
  SELECT
    project_id,
    SUM(base_percentage) AS total_percentage
  FROM crop_percentages
  GROUP BY project_id
),
adjusted_percentages AS (
  SELECT
    cp.project_id,
    cp.current_crop_id,
    cp.crop_name,
    cp.crop_hectares,
    cp.base_percentage,
    ps.total_percentage,
    CASE 
      -- Si suma > 99% Y es el último cultivo, ajustar para 100%
      WHEN ps.total_percentage > 99.000 AND cp.crop_order = cp.total_crops THEN
        100.000 - COALESCE((
          SELECT SUM(base_percentage) 
          FROM crop_percentages cp2 
          WHERE cp2.project_id = cp.project_id 
            AND cp2.crop_order < cp.crop_order
        ), 0)
      ELSE
        cp.base_percentage
    END AS crop_incidence_pct
  FROM crop_percentages cp
  JOIN project_sums ps ON ps.project_id = cp.project_id
)
SELECT
  ap.project_id,
  ap.current_crop_id,
  ap.crop_name,
  ap.crop_hectares,
  ap.crop_incidence_pct,
  -- Costo por ha usando v3_dashboard_ssot
  v3_dashboard_ssot.cost_per_ha_for_crop_ssot(ap.project_id, ap.current_crop_id)::numeric AS cost_per_ha_usd
FROM adjusted_percentages ap
ORDER BY ap.project_id, ap.crop_name;

COMMENT ON VIEW public.v3_dashboard_crop_incidence IS 'Módulo 7: Incidencia de costos por cultivo';

COMMIT;

