-- Restaurar la vista dashboard_view completa basada en 000050
-- Corregir SOLO el problema de wi.lot_id (workorder_items no tiene lot_id)
-- Mantener TODA la funcionalidad existente

DROP VIEW IF EXISTS dashboard_view;

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
      JOIN workorders w ON w.id = wi.workorder_id  -- CORREGIDO: usar workorders como intermediario
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
-- Dimensión de niveles (todas las combinaciones)
-- -----------------------------------------------------------------
levels AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id),
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Siembra agregada  → card "Avance de siembra"
-- -----------------------------------------------------------------
sowing AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id,
    COALESCE(SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END),0)::numeric(14,2) AS sowed_area,
    COALESCE(SUM(l.hectares),0)::numeric(14,2) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots   l ON l.field_id   = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id),
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Cosecha agregada  → card "Avance de cosecha"
-- -----------------------------------------------------------------
harvest AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id,
    COALESCE(SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END),0)::numeric(14,2) AS harvested_area,
    COALESCE(SUM(l.hectares),0)::numeric(14,2) AS total_hectares
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN lots   l ON l.field_id   = f.id AND l.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id),
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Costos agregados (admin + directos ejecutados)
--   C = budget_cost_usd (= admin_cost en projects)
--   B = executed_costs_usd (= direct_costs_usd de v_direct_costs_by_project)
--   NUEVO: budget_total_usd = costos presupuestados totales (hardcodeado temporalmente)
--   NUEVO: costs_progress_pct = porcentaje de avance de costos
-- -----------------------------------------------------------------
costs_agg AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    COALESCE(SUM(COALESCE(p.admin_cost,0)),0)::numeric(14,2)        AS budget_cost_usd,        -- C
    COALESCE(SUM(COALESCE(dc.labors_usd,0)),0)::numeric(14,2)       AS executed_labors_usd,
    COALESCE(SUM(COALESCE(dc.supplies_usd,0)),0)::numeric(14,2)     AS executed_supplies_usd,
    COALESCE(SUM(COALESCE(dc.direct_costs_usd,0)),0)::numeric(14,2) AS executed_costs_usd,     -- B
    -- NUEVO: Costos presupuestados totales (hardcodeado)
    COALESCE(SUM(20000),0)::numeric(14,2)                            AS budget_total_usd,       -- Presupuesto total por proyecto
    -- NUEVO: Porcentaje de avance de costos
    CASE WHEN COALESCE(SUM(20000),0) > 0
         THEN ROUND(((COALESCE(SUM(COALESCE(dc.direct_costs_usd,0)),0) / NULLIF(SUM(20000),0)) * 100)::numeric, 2)
         ELSE 0 END AS costs_progress_pct
  FROM projects p
  LEFT JOIN v_direct_costs_by_project dc ON dc.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Resultado operativo (RENTABILIDAD) → % rojo = (A-B)/B*100
--   A = income_usd ; B = total_invested_usd (= direct labors + supplies)
-- -----------------------------------------------------------------
operating_result AS (
  WITH income_by_project AS (
    SELECT f.project_id, COALESCE(SUM(vf.income_usd),0)::numeric(14,2) AS income_usd
    FROM v_income_by_field vf
    JOIN fields f ON f.id = vf.field_id AND f.deleted_at IS NULL
    GROUP BY f.project_id
  )
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    COALESCE(SUM(COALESCE(ip.income_usd,0)),0)::numeric(14,2)  AS income_usd,          -- A
    COALESCE(SUM(COALESCE(el.labors_usd,0)),0)::numeric(14,2)  AS direct_labors_usd,
    COALESCE(SUM(COALESCE(us.supplies_usd,0)),0)::numeric(14,2) AS direct_supplies_usd,
    (COALESCE(SUM(COALESCE(el.labors_usd,0)),0)
     + COALESCE(SUM(COALESCE(us.supplies_usd,0)),0))::numeric(14,2) AS total_invested_usd -- B
  FROM projects p
  LEFT JOIN income_by_project         ip ON ip.project_id = p.id
  LEFT JOIN executed_labors_by_project el ON el.project_id = p.id
  LEFT JOIN used_supplies_by_project   us ON us.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Aportes agregados → card "Avance de aportes"
--   Calcula el porcentaje de participación acordada por inversores
--   Basado en project_investors.percentage
--   Retorna información individual de cada inversor
--   SIEMPRE 100% por proyecto (1 proyecto a la vez)
--   Los inversores son por PROYECTO, no por campo
-- -----------------------------------------------------------------
contributions AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS investor_percentage_pct,
    -- SIEMPRE 100% por proyecto
    100.00::numeric(6,2) AS contributions_progress_pct
  FROM projects p
  LEFT JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
  LEFT JOIN investors i ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, pi.investor_id, i.name, pi.percentage),
    (p.customer_id, p.id, pi.investor_id, i.name, pi.percentage),
    (p.customer_id, pi.investor_id, i.name, pi.percentage),
    (pi.investor_id, i.name, pi.percentage)
  )
),

-- -----------------------------------------------------------------
-- Incidencia de Costos por Cultivo
--   Filas: Cultivos del proyecto/campo
--   Columnas: Superficie (Has), Incidencia %, Costo por Ha por cultivo
-- -----------------------------------------------------------------
v_crop_incidence AS (
  SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    c.id AS crop_id,
    c.name AS crop_name,
    
    -- 1. Superficie (Has): Suma de superficies asociadas al cultivo
    COALESCE(SUM(l.hectares), 0)::numeric(14,2) AS crop_hectares,
    
    -- 2. Hectáreas totales del proyecto
    COALESCE(SUM(SUM(l.hectares)) OVER (PARTITION BY p.id), 0)::numeric(14,2) AS project_total_hectares,
    
    -- 3. Costos directos del cultivo (labores + insumos ejecutados)
    COALESCE(SUM(
      CASE 
        WHEN w.labor_id IS NOT NULL THEN lb.price  -- Labores ejecutadas
        WHEN wi.supply_id IS NOT NULL THEN sp.price  -- Insumos utilizados
        ELSE 0 
      END
    ), 0)::numeric(14,2) AS crop_direct_costs_usd,
    
    -- 4. Incidencia %: (Hectáreas del cultivo / Hectáreas totales) × 100
    CASE 
      WHEN SUM(SUM(l.hectares)) OVER (PARTITION BY p.id) > 0 
      THEN ROUND((SUM(l.hectares) / SUM(SUM(l.hectares)) OVER (PARTITION BY p.id) * 100)::numeric, 2)
      ELSE 0 
    END AS incidence_pct,
    
    -- 5. Costo por Ha por cultivo: Costos directos / Hectáreas del cultivo
    CASE 
      WHEN SUM(l.hectares) > 0 
      THEN ROUND((SUM(
        CASE 
          WHEN w.labor_id IS NOT NULL THEN lb.price
          WHEN wi.supply_id IS NOT NULL THEN sp.price
          ELSE 0 
        END
      ) / SUM(l.hectares))::numeric, 2)
      ELSE 0 
    END AS cost_per_ha_usd
    
  FROM projects p
  JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  JOIN crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
  
  -- Labores ejecutadas en órdenes de trabajo
  LEFT JOIN workorders w ON w.lot_id = l.id 
                         AND w.effective_area > 0 
                         AND w.deleted_at IS NULL
  LEFT JOIN labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  
  -- Insumos utilizados en órdenes de trabajo (CORREGIDO: usar workorders como intermediario)
  LEFT JOIN workorder_items wi ON wi.workorder_id = w.id 
                               AND wi.final_dose > 0 
                               AND wi.deleted_at IS NULL
  LEFT JOIN supplies sp ON sp.id = wi.supply_id AND sp.deleted_at IS NULL
  
  WHERE p.deleted_at IS NULL
  GROUP BY p.customer_id, p.id, p.campaign_id, c.id, c.name
),

-- -----------------------------------------------------------------
-- Balance de Gestión - Semillas
-- -----------------------------------------------------------------
v_semilla_ejecutados AS (
  SELECT 
    sp.project_id,
    COALESCE(SUM(sp.price), 0)::numeric(14,2) AS semilla_ejecutados_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND EXISTS (
      SELECT 1
      FROM workorder_items wi
      JOIN workorders w ON w.id = wi.workorder_id  -- CORREGIDO: usar workorders como intermediario
      WHERE wi.supply_id = sp.id
        AND wi.final_dose > 0
        AND wi.deleted_at IS NULL
        AND w.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),

v_semilla_invertidos AS (
  SELECT 
    sp.project_id,
    COALESCE(SUM(sp.price), 0)::numeric(14,2) AS semilla_invertidos_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
  GROUP BY sp.project_id
),

-- -----------------------------------------------------------------
-- Balance de Gestión - Insumos (excluyendo semillas)
-- -----------------------------------------------------------------
v_insumos_ejecutados AS (
  SELECT 
    sp.project_id,
    COALESCE(SUM(sp.price), 0)::numeric(14,2) AS insumos_ejecutados_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
    AND EXISTS (
      SELECT 1
      FROM workorder_items wi
      JOIN workorders w ON w.id = wi.workorder_id  -- CORREGIDO: usar workorders como intermediario
      WHERE wi.supply_id = sp.id
        AND wi.final_dose > 0
        AND wi.deleted_at IS NULL
        AND w.deleted_at IS NULL
    )
  GROUP BY sp.project_id
),

v_insumos_invertidos AS (
  SELECT 
    sp.project_id,
    COALESCE(SUM(sp.price), 0)::numeric(14,2) AS insumos_invertidos_usd
  FROM supplies sp
  WHERE sp.deleted_at IS NULL
  GROUP BY sp.project_id
),

-- -----------------------------------------------------------------
-- Balance de Gestión - Labores
-- -----------------------------------------------------------------
v_labores_ejecutados AS (
  SELECT 
    lb.project_id,
    COALESCE(SUM(lb.price), 0)::numeric(14,2) AS labores_ejecutados_usd
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

v_labores_invertidos AS (
  SELECT 
    lb.project_id,
    COALESCE(SUM(lb.price), 0)::numeric(14,2) AS labores_invertidos_usd
  FROM labors lb
  WHERE lb.deleted_at IS NULL
  GROUP BY lb.project_id
),

-- -----------------------------------------------------------------
-- Balance de Gestión - Arriendo (solo invertidos)
-- -----------------------------------------------------------------
v_arriendo_invertidos AS (
  SELECT 
    p.id AS project_id,
    -- Total comercializaciones × 30%
    (COALESCE(SUM(cc.net_price), 0) * 0.30)::numeric(14,2) AS arriendo_invertidos_usd
  FROM projects p
  LEFT JOIN crop_commercializations cc ON cc.project_id = p.id AND cc.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.id
),

-- -----------------------------------------------------------------
-- Balance de Gestión - Estructura (solo invertidos)
-- -----------------------------------------------------------------
v_estructura_invertidos AS (
  SELECT 
    p.id AS project_id,
    -- Gastos administrativos fijos del proyecto
    COALESCE(p.admin_cost, 0)::numeric(14,2) AS estructura_invertidos_usd
  FROM projects p
  WHERE p.deleted_at IS NULL
),

-- -----------------------------------------------------------------
-- Balance de Gestión - Agregación final
-- -----------------------------------------------------------------
v_balance_gestion AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    CASE WHEN GROUPING(f.id)=1          THEN NULL ELSE f.id          END AS field_id,
    
    -- Semilla
    COALESCE(SUM(se.semilla_ejecutados_usd), 0)::numeric(14,2) AS semilla_ejecutados_usd,
    COALESCE(SUM(si.semilla_invertidos_usd), 0)::numeric(14,2) AS semilla_invertidos_usd,
    (COALESCE(SUM(si.semilla_invertidos_usd), 0) - COALESCE(SUM(se.semilla_ejecutados_usd), 0))::numeric(14,2) AS semilla_stock_usd,
    
    -- Insumos
    COALESCE(SUM(ie.insumos_ejecutados_usd), 0)::numeric(14,2) AS insumos_ejecutados_usd,
    COALESCE(SUM(ii.insumos_invertidos_usd), 0)::numeric(14,2) AS insumos_invertidos_usd,
    (COALESCE(SUM(ii.insumos_invertidos_usd), 0) - COALESCE(SUM(ie.insumos_ejecutados_usd), 0))::numeric(14,2) AS insumos_stock_usd,
    
    -- Labores
    COALESCE(SUM(le.labores_ejecutados_usd), 0)::numeric(14,2) AS labores_ejecutados_usd,
    COALESCE(SUM(li.labores_invertidos_usd), 0)::numeric(14,2) AS labores_invertidos_usd,
    (COALESCE(SUM(li.labores_invertidos_usd), 0) - COALESCE(SUM(le.labores_ejecutados_usd), 0))::numeric(14,2) AS labores_stock_usd,
    
    -- Arriendo (solo invertidos)
    0::numeric(14,2) AS arriendo_ejecutados_usd,  -- NO se calcula
    COALESCE(SUM(ai.arriendo_invertidos_usd), 0)::numeric(14,2) AS arriendo_invertidos_usd,  -- Total comercializaciones × 30%
    0::numeric(14,2) AS arriendo_stock_usd,  -- NO se calcula
    
    -- Estructura (solo invertidos)
    0::numeric(14,2) AS estructura_ejecutados_usd,  -- NO se calcula
    COALESCE(SUM(ei.estructura_invertidos_usd), 0)::numeric(14,2) AS estructura_invertidos_usd,  -- projects.admin_cost
    0::numeric(14,2) AS estructura_stock_usd,  -- NO se calcula
    
    -- Totales del Balance de Gestión (incluyendo arriendo y estructura)
    (COALESCE(SUM(se.semilla_ejecutados_usd), 0) + COALESCE(SUM(ie.insumos_ejecutados_usd), 0) + 
     COALESCE(SUM(le.labores_ejecutados_usd), 0))::numeric(14,2) AS costos_directos_ejecutados_usd,
    (COALESCE(SUM(si.semilla_invertidos_usd), 0) + COALESCE(SUM(ii.insumos_invertidos_usd), 0) + 
     COALESCE(SUM(li.labores_invertidos_usd), 0) + COALESCE(SUM(ai.arriendo_invertidos_usd), 0) + 
     COALESCE(SUM(ei.estructura_invertidos_usd), 0))::numeric(14,2) AS costos_directos_invertidos_usd,
    (COALESCE(SUM(si.semilla_invertidos_usd), 0) + COALESCE(SUM(ii.insumos_invertidos_usd), 0) + 
     COALESCE(SUM(li.labores_invertidos_usd), 0) + COALESCE(SUM(ai.arriendo_invertidos_usd), 0) + 
     COALESCE(SUM(ei.estructura_invertidos_usd), 0) - 
     COALESCE(SUM(se.semilla_ejecutados_usd), 0) - COALESCE(SUM(ie.insumos_ejecutados_usd), 0) - 
     COALESCE(SUM(le.labores_ejecutados_usd), 0))::numeric(14,2) AS costos_directos_stock_usd
    
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN v_semilla_ejecutados se ON se.project_id = p.id
  LEFT JOIN v_semilla_invertidos si ON si.project_id = p.id
  LEFT JOIN v_insumos_ejecutados ie ON ie.project_id = p.id
  LEFT JOIN v_insumos_invertidos ii ON ii.project_id = p.id
  LEFT JOIN v_labores_ejecutados le ON le.project_id = p.id
  LEFT JOIN v_labores_invertidos li ON li.project_id = p.id
  -- JOINs para arriendo y estructura
  LEFT JOIN v_arriendo_invertidos ai ON ai.project_id = p.id
  LEFT JOIN v_estructura_invertidos ei ON ei.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id, f.id),
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Indicadores Operativos
--   4 elementos: Primera orden, Última orden, Arqueo de stock, Cierre de campaña
-- -----------------------------------------------------------------
v_indicadores_operativos AS (
  SELECT
    p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    
    -- 1. Primera orden de trabajo
    MIN(w.date) AS primera_orden_fecha,
    MIN(w.id) AS primera_orden_id,
    
    -- 2. Última orden de trabajo
    MAX(w.date) AS ultima_orden_fecha,
    MAX(w.id) AS ultima_orden_id,
    
    -- 3. Arqueo de stock (fecha del último inventario)
    -- Por ahora se deja NULL hasta que se implemente la funcionalidad
    NULL::timestamp AS arqueo_stock_fecha,
    
    -- 4. Cierre de campaña (fecha estimada o real)
    -- Por ahora se deja NULL hasta que se implemente la funcionalidad
    NULL::timestamp AS cierre_campana_fecha
    
  FROM projects p
  LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
  LEFT JOIN workorders w ON w.project_id = p.id AND w.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
  GROUP BY p.customer_id, p.id, p.campaign_id
)

-- -----------------------------------------------------------------
-- SALIDA ÚNICA (base = levels)
-- -----------------------------------------------------------------
SELECT
  lvl.customer_id,
  lvl.project_id,
  lvl.campaign_id,
  lvl.field_id,

  -- Siembra
  COALESCE(s.sowed_area,0)::numeric(14,2)     AS sowing_hectares,
  COALESCE(s.total_hectares,0)::numeric(14,2) AS sowing_total_hectares,

  -- Cosecha
  COALESCE(h.harvested_area,0)::numeric(14,2) AS harvest_hectares,
  COALESCE(h.total_hectares,0)::numeric(14,2) AS harvest_total_hectares,

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

  -- NUEVO: Incidencia de Costos por Cultivo
  COALESCE(ci.crop_id,0)::bigint AS crop_id,
  COALESCE(ci.crop_name,'') AS crop_name,
  COALESCE(ci.crop_hectares,0)::numeric(14,2) AS crop_hectares,           -- Superficie (Has)
  COALESCE(ci.project_total_hectares,0)::numeric(14,2) AS project_total_hectares,
  COALESCE(ci.incidence_pct,0)::numeric(6,2) AS incidence_pct,            -- Incidencia %
  COALESCE(ci.crop_direct_costs_usd,0)::numeric(14,2) AS crop_direct_costs_usd,  -- Costos directos del cultivo
  COALESCE(ci.cost_per_ha_usd,0)::numeric(14,2) AS cost_per_ha_usd,      -- Costo por Ha por cultivo

  -- NUEVO: Balance de Gestión - Semilla
  COALESCE(bg.semilla_ejecutados_usd,0)::numeric(14,2) AS semilla_ejecutados_usd,
  COALESCE(bg.semilla_invertidos_usd,0)::numeric(14,2) AS semilla_invertidos_usd,
  COALESCE(bg.semilla_stock_usd,0)::numeric(14,2) AS semilla_stock_usd,
  
  -- NUEVO: Balance de Gestión - Insumos
  COALESCE(bg.insumos_ejecutados_usd,0)::numeric(14,2) AS insumos_ejecutados_usd,
  COALESCE(bg.insumos_invertidos_usd,0)::numeric(14,2) AS insumos_invertidos_usd,
  COALESCE(bg.insumos_stock_usd,0)::numeric(14,2) AS insumos_stock_usd,
  
  -- NUEVO: Balance de Gestión - Labores
  COALESCE(bg.labores_ejecutados_usd,0)::numeric(14,2) AS labores_ejecutados_usd,
  COALESCE(bg.labores_invertidos_usd,0)::numeric(14,2) AS labores_invertidos_usd,
  COALESCE(bg.labores_stock_usd,0)::numeric(14,2) AS labores_stock_usd,
  
  -- NUEVO: Balance de Gestión - Arriendo (solo invertidos)
  0::numeric(14,2) AS arriendo_ejecutados_usd,  -- NO se calcula
  COALESCE(bg.arriendo_invertidos_usd,0)::numeric(14,2) AS arriendo_invertidos_usd,  -- Total comercializaciones × 30%
  0::numeric(14,2) AS arriendo_stock_usd,  -- NO se calcula
  
  -- NUEVO: Balance de Gestión - Estructura (solo invertidos)
  0::numeric(14,2) AS estructura_ejecutados_usd,  -- NO se calcula
  COALESCE(bg.estructura_invertidos_usd,0)::numeric(14,2) AS estructura_invertidos_usd,  -- projects.admin_cost
  0::numeric(14,2) AS estructura_stock_usd,  -- NO se calcula
  
  -- NUEVO: Totales del Balance de Gestión
  COALESCE(bg.costos_directos_ejecutados_usd,0)::numeric(14,2) AS costos_directos_ejecutados_usd,
  COALESCE(bg.costos_directos_invertidos_usd,0)::numeric(14,2) AS costos_directos_invertidos_usd,
  COALESCE(bg.costos_directos_stock_usd,0)::numeric(14,2) AS costos_directos_stock_usd,

  -- NUEVO: Indicadores Operativos
  io.primera_orden_fecha,
  io.primera_orden_id,
  io.ultima_orden_fecha,
  io.ultima_orden_id,
  io.arqueo_stock_fecha,
  io.cierre_campana_fecha,

  -- Identificador de fila
  'metric'::text AS row_kind

FROM levels lvl
LEFT JOIN sowing s
  ON s.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND s.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND s.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
 AND s.field_id    IS NOT DISTINCT FROM lvl.field_id
LEFT JOIN harvest h
  ON h.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND h.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND h.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
 AND s.field_id    IS NOT DISTINCT FROM lvl.field_id
LEFT JOIN costs_agg ca
  ON ca.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND ca.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND ca.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN operating_result o
  ON o.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND o.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND o.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN contributions c
  ON c.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND c.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND c.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN v_crop_incidence ci
  ON ci.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND ci.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND ci.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN v_balance_gestion bg
  ON bg.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND bg.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND bg.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN v_indicadores_operativos io
  ON io.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND io.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND io.campaign_id IS NOT DISTINCT FROM lvl.campaign_id;
