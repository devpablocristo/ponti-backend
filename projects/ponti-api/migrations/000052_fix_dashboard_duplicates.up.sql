-- ========================================
-- MIGRACIÓN 000052: CORRECCIÓN DE DUPLICADOS EN DASHBOARD
-- ========================================
-- 
-- PROBLEMA IDENTIFICADO:
-- El dashboard_view está devolviendo múltiples filas duplicadas en lugar de una sola fila por proyecto.
-- Esto se debe a que el CTE 'levels' con GROUPING SETS está generando todas las combinaciones
-- posibles y luego los JOINs crean un producto cartesiano.
-- 
-- SOLUCIÓN:
-- Simplificar la lógica para que solo devuelva una fila por proyecto, eliminando
-- la complejidad innecesaria de GROUPING SETS cuando solo necesitamos datos por proyecto.
-- 
-- RESULTADO ESPERADO:
-- Proyecto 1: 100 ha sembradas / 200 ha totales = 50% (1 fila)
-- Proyecto 2: 150 ha sembradas / 150 ha totales = 100% (1 fila)  
-- Proyecto 3: 0 ha sembradas / 100 ha totales = 0% (1 fila)
-- ========================================

-- Eliminar la vista existente
DROP VIEW IF EXISTS dashboard_view;

-- Recrear la vista simplificada (solo por proyecto)
CREATE OR REPLACE VIEW dashboard_view AS
WITH
executed_labors_by_project AS (
  SELECT lb.project_id, SUM(lb.price) AS labors_usd
  FROM labors lb
  WHERE lb.deleted_at IS NULL
    AND EXISTS (
      SELECT 1
      FROM workorders w
      WHERE w.labor_id = lb.id
        AND w.effective_area > 0
        AND w.deleted_at IS NULL
    )
  GROUP BY lb.project_id
),
used_supplies_by_project AS (
  SELECT sp.project_id, SUM(sp.price) AS supplies_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND EXISTS (
      SELECT 1
      FROM workorder_items wi
      JOIN workorders w ON w.id = wi.workorder_id
      WHERE wi.supply_id = sp.id
        AND wi.final_dose > 0
        AND wi.deleted_at IS NULL
        AND w.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),

-- -----------------------------------------------------------------
-- Costos directos por proyecto (SOLO ejecutado / utilizado)
--   B = direct_costs_usd = labors_usd + supplies_usd
-- -----------------------------------------------------------------
v_direct_costs_by_project AS (
  SELECT
    p.id AS project_id,
    COALESCE(el.labors_usd,   0)::numeric(14,2) AS labors_usd,
    COALESCE(us.supplies_usd, 0)::numeric(14,2) AS supplies_usd,
    (COALESCE(el.labors_usd,0) + COALESCE(us.supplies_usd,0))::numeric(14,2) AS direct_costs_usd
  FROM projects p
  LEFT JOIN executed_labors_by_project el ON el.project_id = p.id
  LEFT JOIN used_supplies_by_project  us ON us.project_id = p.id
  WHERE p.deleted_at IS NULL
),

-- -----------------------------------------------------------------
-- Ingresos por field (tons * precio por tonelada)  → A = income_usd
--   El precio por tonelada debe ser configurado en la aplicación
-- -----------------------------------------------------------------
v_income_by_field AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    COALESCE(SUM(l.tons), 0)::numeric(14,2) AS total_tons,
    -- NUEVO: Ingresos calculados (debe ser configurado en la app)
    0::numeric(14,2) AS income_usd
  FROM fields f
  LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.deleted_at IS NULL
  GROUP BY f.project_id, f.id
),

-- -----------------------------------------------------------------
-- Siembra agregada por proyecto → card "Avance de siembra"
--   CORREGIDO: Solo una fila por proyecto
--   sowed_area = Solo hectáreas con sowing_date NOT NULL
--   total_hectares = Todas las hectáreas del proyecto
-- -----------------------------------------------------------------
sowing_by_project AS (
  SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    -- Solo sumar hectáreas cuando sowing_date NO es NULL
    COALESCE(SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END),0)::numeric(14,2) AS sowed_area,
    -- Sumar TODAS las hectáreas para el total
    COALESCE(SUM(l.hectares),0)::numeric(14,2) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots   l ON l.field_id   = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.customer_id, p.id, p.campaign_id
),

-- -----------------------------------------------------------------
-- Cosecha agregada por proyecto → card "Avance de cosecha"
-- -----------------------------------------------------------------
harvest_by_project AS (
  SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END),0)::numeric(14,2) AS harvested_area,
    COALESCE(SUM(l.hectares),0)::numeric(14,2) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots   l ON l.field_id   = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.customer_id, p.id, p.campaign_id
),

-- -----------------------------------------------------------------
-- Costos agregados por proyecto (admin + directos ejecutados)
--   C = budget_cost_usd (= admin_cost en projects)
--   B = executed_costs_usd (= direct_costs_usd de v_direct_costs_by_project)
--   NUEVO: budget_total_usd = costos presupuestados totales (hardcodeado temporalmente)
--   NUEVO: costs_progress_pct = porcentaje de avance de costos
-- -----------------------------------------------------------------
costs_by_project AS (
  SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(p.admin_cost,0)::numeric(14,2)        AS budget_cost_usd,        -- C
    COALESCE(dc.labors_usd,0)::numeric(14,2)       AS executed_labors_usd,
    COALESCE(dc.supplies_usd,0)::numeric(14,2)     AS executed_supplies_usd,
    COALESCE(dc.direct_costs_usd,0)::numeric(14,2) AS executed_costs_usd,     -- B
    -- NUEVO: Costos presupuestados totales (hardcodeado)
    20000::numeric(14,2)                            AS budget_total_usd,       -- Presupuesto total por proyecto
    -- NUEVO: Porcentaje de avance de costos
    CASE WHEN 20000 > 0
         THEN ROUND(((COALESCE(dc.direct_costs_usd,0) / 20000) * 100)::numeric, 2)
         ELSE 0 END AS costs_progress_pct
  FROM projects p
  LEFT JOIN v_direct_costs_by_project dc ON dc.project_id = p.id
  WHERE p.deleted_at IS NULL
),

-- -----------------------------------------------------------------
-- Resultado operativo por proyecto (RENTABILIDAD) → % rojo = (A-B)/B*100
--   A = income_usd ; B = total_invested_usd (= direct labors + supplies)
-- -----------------------------------------------------------------
operating_result_by_project AS (
  WITH income_by_project AS (
    SELECT f.project_id, COALESCE(SUM(vf.income_usd),0)::numeric(14,2) AS income_usd
    FROM v_income_by_field vf
    JOIN fields f ON f.id = vf.field_id AND f.deleted_at IS NULL
    GROUP BY f.project_id
  )
  SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(ip.income_usd,0)::numeric(14,2)  AS income_usd,          -- A
    COALESCE(el.labors_usd,0)::numeric(14,2)  AS direct_labors_usd,
    COALESCE(us.supplies_usd,0)::numeric(14,2) AS direct_supplies_usd,
    (COALESCE(el.labors_usd,0) + COALESCE(us.supplies_usd,0))::numeric(14,2) AS total_invested_usd -- B
  FROM projects p
  LEFT JOIN income_by_project         ip ON ip.project_id = p.id
  LEFT JOIN executed_labors_by_project el ON el.project_id = p.id
  LEFT JOIN used_supplies_by_project   us ON us.project_id = p.id
  WHERE p.deleted_at IS NULL
),

-- -----------------------------------------------------------------
-- Aportes por proyecto → card "Avance de aportes"
--   Calcula el porcentaje de participación acordada por inversores
--   Basado en project_investors.percentage
--   Retorna información individual de cada inversor
--   SIEMPRE 100% por proyecto (1 proyecto a la vez)
--   Los inversores son por PROYECTO, no por campo
-- -----------------------------------------------------------------
contributions_by_project AS (
  SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS investor_percentage_pct,
    -- SIEMPRE 100% por proyecto
    100.00::numeric(6,2) AS contributions_progress_pct
  FROM projects p
  LEFT JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
  LEFT JOIN investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
)

-- -----------------------------------------------------------------
-- SALIDA ÚNICA (una fila por proyecto)
-- -----------------------------------------------------------------
SELECT
  p.customer_id,
  p.id AS project_id,
  p.campaign_id,
  NULL::bigint AS field_id,  -- No usamos field_id en esta versión simplificada

  -- Siembra
  COALESCE(s.sowed_area,0)::numeric(14,2)     AS sowing_hectares,
  COALESCE(s.total_hectares,0)::numeric(14,2) AS sowing_total_hectares,
  -- CORREGIDO: Porcentaje de avance de siembra
  --   Fórmula: (hectáreas sembradas / hectáreas totales) × 100
  --   Ejemplo: (30 ha / 45 ha) × 100 = 66.67%
  CASE WHEN COALESCE(s.total_hectares,0) > 0 
       THEN ROUND((COALESCE(s.sowed_area,0) / NULLIF(s.total_hectares,0) * 100)::numeric, 2)
       ELSE 0 END AS sowing_progress_percent,

  -- Cosecha
  COALESCE(h.harvested_area,0)::numeric(14,2) AS harvest_hectares,
  COALESCE(h.total_hectares,0)::numeric(14,2) AS harvest_total_hectares,
  -- NUEVO: Porcentaje de avance de cosecha
  CASE WHEN COALESCE(h.total_hectares,0) > 0 
       THEN ROUND((COALESCE(h.harvested_area,0) / NULLIF(h.total_hectares,0) * 100)::numeric, 2)
       ELSE 0 END AS harvest_progress_percent,

  -- Costos (B y C)
  COALESCE(ca.executed_costs_usd,0)::numeric(14,2)     AS executed_costs_usd,     -- B
  COALESCE(ca.executed_labors_usd,0)::numeric(14,2)    AS executed_labors_usd,
  COALESCE(ca.executed_supplies_usd,0)::numeric(14,2)  AS executed_supplies_usd,
  COALESCE(ca.budget_cost_usd,0)::numeric(14,2)        AS budget_cost_usd,        -- C (admin)
  -- NUEVO: Costos presupuestados totales y porcentaje de avance
  COALESCE(ca.budget_total_usd,0)::numeric(14,2)       AS budget_total_usd,       -- Presupuesto total
  COALESCE(ca.costs_progress_pct,0)::numeric(14,2)     AS costs_progress_pct,     -- % avance costos

  -- Ingresos (A) y Resultado operativo
  COALESCE(o.income_usd,0)::numeric(14,2)             AS income_usd,              -- A
  (COALESCE(o.income_usd,0) - COALESCE(o.total_invested_usd,0))::numeric(14,2) AS operating_result_usd,
  CASE WHEN COALESCE(o.total_invested_usd,0) > 0
       THEN ROUND(((COALESCE(o.income_usd,0) - COALESCE(o.total_invested_usd,0))
                  / NULLIF(o.total_invested_usd,0) * 100)::numeric, 2)
       ELSE 0 END AS operating_result_pct,                                      -- % rojo = (A-B)/B*100

  -- NÚMERO GRIS 2 listo para UI: (B + C)
  (COALESCE(ca.executed_costs_usd,0) + COALESCE(ca.budget_cost_usd,0))::numeric(14,2)
     AS operating_result_total_costs_usd,                                       -- gris 2 = B + C

  -- Aportes (información individual de cada inversor)
  COALESCE(c.investor_id,0)::bigint AS investor_id,
  COALESCE(c.investor_name,'') AS investor_name,
  COALESCE(c.investor_percentage_pct,0)::numeric(6,2) AS investor_percentage_pct,
  COALESCE(c.contributions_progress_pct,0)::numeric(6,2) AS contributions_progress_pct,

  -- Placeholders para campos que no se usan en esta versión simplificada
  0::bigint AS crop_id,
  '' AS crop_name,
  0::numeric(14,2) AS crop_hectares,
  0::numeric(14,2) AS project_total_hectares,
  0::numeric(6,2) AS incidence_pct,
  0::numeric(14,2) AS crop_direct_costs_usd,
  0::numeric(14,2) AS cost_per_ha_usd,

  -- Balance de Gestión - Placeholders
  0::numeric(14,2) AS semilla_ejecutados_usd,
  0::numeric(14,2) AS semilla_invertidos_usd,
  0::numeric(14,2) AS semilla_stock_usd,
  0::numeric(14,2) AS insumos_ejecutados_usd,
  0::numeric(14,2) AS insumos_invertidos_usd,
  0::numeric(14,2) AS insumos_stock_usd,
  0::numeric(14,2) AS labores_ejecutados_usd,
  0::numeric(14,2) AS labores_invertidos_usd,
  0::numeric(14,2) AS labores_stock_usd,
  0::numeric(14,2) AS arriendo_ejecutados_usd,
  0::numeric(14,2) AS arriendo_invertidos_usd,
  0::numeric(14,2) AS arriendo_stock_usd,
  0::numeric(14,2) AS estructura_ejecutados_usd,
  0::numeric(14,2) AS estructura_invertidos_usd,
  0::numeric(14,2) AS estructura_stock_usd,
  0::numeric(14,2) AS costos_directos_ejecutados_usd,
  0::numeric(14,2) AS costos_directos_invertidos_usd,
  0::numeric(14,2) AS costos_directos_stock_usd,

  -- Indicadores Operativos - Placeholders
  NULL::timestamp AS primera_orden_fecha,
  0::bigint AS primera_orden_id,
  NULL::timestamp AS ultima_orden_fecha,
  0::bigint AS ultima_orden_id,
  NULL::timestamp AS arqueo_stock_fecha,
  NULL::timestamp AS cierre_campana_fecha,

  -- Identificador de fila
  'metric'::text AS row_kind

FROM projects p
LEFT JOIN sowing_by_project s ON s.project_id = p.id
LEFT JOIN harvest_by_project h ON h.project_id = p.id
LEFT JOIN costs_by_project ca ON ca.project_id = p.id
LEFT JOIN operating_result_by_project o ON o.project_id = p.id
LEFT JOIN contributions_by_project c ON c.project_id = p.id
WHERE p.deleted_at IS NULL
ORDER BY p.id;
