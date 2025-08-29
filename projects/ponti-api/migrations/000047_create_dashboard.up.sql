-- =========================================================
-- ÚNICA VISTA: dashboard_view (sin FULL JOIN, Cloud SQL friendly)
--  - Global / Customer / Project / Campaign / Field (GROUPING SETS)
--  - Incluye: siembra, cosecha, costos (budget / ejecutado),
--             ingresos, contribuciones, resultado operativo (USD y %)
--  - Expone labors/supplies y breakdown normalizado por inversor
--  - row_kind: 'metric' (métricas) | 'investor' (filas por inversor)
-- =========================================================
CREATE OR REPLACE VIEW dashboard_view AS
WITH
-- -----------------------------------------------------------------
-- Costos directos por proyecto (labors + supplies)
-- -----------------------------------------------------------------
v_direct_costs_by_project AS (
  SELECT
    p.id AS project_id,
    COALESCE((SELECT SUM(lb.price) FROM labors lb WHERE lb.project_id = p.id AND lb.deleted_at IS NULL), 0)::numeric(14,2) AS labors_usd,
    COALESCE((SELECT SUM(sp.price) FROM supplies sp WHERE sp.project_id = p.id AND sp.deleted_at IS NULL), 0)::numeric(14,2) AS supplies_usd,
    (
      COALESCE((SELECT SUM(lb.price) FROM labors lb WHERE lb.project_id = p.id AND lb.deleted_at IS NULL), 0)
      +
      COALESCE((SELECT SUM(sp.price) FROM supplies sp WHERE sp.project_id = p.id AND sp.deleted_at IS NULL), 0)
    )::numeric(14,2) AS direct_costs_usd
  FROM projects p
  WHERE p.deleted_at IS NULL
),

-- -----------------------------------------------------------------
-- Ingresos por field (tons * 200)
-- -----------------------------------------------------------------
v_income_by_field AS (
  SELECT
    f.project_id,
    f.id AS field_id,
    COALESCE(SUM(l.tons * 200), 0)::numeric(14,2) AS income_usd
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
-- Siembra agregada
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
-- Cosecha agregada
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
-- Costos agregados (budget + ejecutado desglosado)
-- -----------------------------------------------------------------
costs_agg AS (
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    COALESCE(SUM(COALESCE(p.admin_cost,0)),0)::numeric(14,2)                AS budget_cost_usd,
    COALESCE(SUM(COALESCE(dc.labors_usd,0)),0)::numeric(14,2)               AS executed_labors_usd,
    COALESCE(SUM(COALESCE(dc.supplies_usd,0)),0)::numeric(14,2)             AS executed_supplies_usd,
    COALESCE(SUM(COALESCE(dc.direct_costs_usd,0)),0)::numeric(14,2)         AS executed_costs_usd
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
-- Contribuciones: filas por inversor (nivel proyecto; field_id NULL)
-- -----------------------------------------------------------------
contrib_rows AS (
  SELECT
    p.customer_id,
    p.id           AS project_id,
    p.campaign_id,
    NULL::bigint   AS field_id,
    pi.investor_id,
    COALESCE(i.name,'')                     AS investor_name,
    COALESCE(pi.percentage,0)::numeric(6,2) AS investor_percentage,
    COALESCE(pi.percentage,0)::numeric(6,2) AS investor_contribution_pct
  FROM projects p
  JOIN project_investors pi ON pi.project_id = p.id AND pi.deleted_at IS NULL
  LEFT JOIN investors i      ON i.id = pi.investor_id AND i.deleted_at IS NULL
  WHERE p.deleted_at IS NULL
),

-- Métrica de contribuciones por nivel (100 si hay inversores; 0 si no)
contrib_metric AS (
  SELECT
    CASE WHEN GROUPING(customer_id)=1 THEN NULL ELSE customer_id END AS customer_id,
    CASE WHEN GROUPING(project_id)=1  THEN NULL ELSE project_id  END AS project_id,
    CASE WHEN GROUPING(campaign_id)=1 THEN NULL ELSE campaign_id END AS campaign_id,
    CASE WHEN GROUPING(field_id)=1    THEN NULL ELSE field_id    END AS field_id,
    CASE WHEN COUNT(*) > 0 THEN 100.0::numeric(14,2) ELSE 0::numeric(14,2) END AS contrib_progress_pct
  FROM contrib_rows
  GROUP BY GROUPING SETS (
    (customer_id, project_id, campaign_id, field_id),
    (customer_id, project_id, campaign_id),
    (customer_id, project_id),
    (customer_id),
    ()
  )
),

-- -----------------------------------------------------------------
-- Resultado operativo (ingresos vs SOLO labors)
-- -----------------------------------------------------------------
operating_result AS (
  WITH income_by_project AS (
    SELECT f.project_id, COALESCE(SUM(vf.income_usd),0)::numeric(14,2) AS income_usd
    FROM v_income_by_field vf
    JOIN fields f ON f.id = vf.field_id AND f.deleted_at IS NULL
    GROUP BY f.project_id
  ),
  labors_by_project AS (
    SELECT lb.project_id, COALESCE(SUM(lb.price),0)::numeric(14,2) AS labors_usd
    FROM labors lb
    WHERE lb.deleted_at IS NULL
    GROUP BY lb.project_id
  )
  SELECT
    CASE WHEN GROUPING(p.customer_id)=1 THEN NULL ELSE p.customer_id END AS customer_id,
    CASE WHEN GROUPING(p.id)=1          THEN NULL ELSE p.id          END AS project_id,
    CASE WHEN GROUPING(p.campaign_id)=1 THEN NULL ELSE p.campaign_id END AS campaign_id,
    COALESCE(SUM(COALESCE(ip.income_usd,0)),0)::numeric(14,2)  AS income_usd,
    COALESCE(SUM(COALESCE(lb.labors_usd,0)),0)::numeric(14,2)  AS direct_labors_usd
  FROM projects p
  LEFT JOIN income_by_project ip ON ip.project_id = p.id
  LEFT JOIN labors_by_project lb ON lb.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY GROUPING SETS (
    (p.customer_id, p.id, p.campaign_id),
    (p.customer_id, p.id),
    (p.customer_id),
    ()
  )
)

-- -----------------------------------------------------------------
-- SALIDA ÚNICA (base = levels) con UNION ALL: METRIC + INVESTOR
-- -----------------------------------------------------------------
-- 1) Filas METRIC
SELECT
  lvl.customer_id,
  lvl.project_id,
  lvl.campaign_id,
  lvl.field_id,

  COALESCE(s.sowed_area,0)::numeric(14,2)     AS sowing_hectares,
  COALESCE(s.total_hectares,0)::numeric(14,2) AS sowing_total_hectares,

  COALESCE(h.harvested_area,0)::numeric(14,2) AS harvest_hectares,
  COALESCE(h.total_hectares,0)::numeric(14,2) AS harvest_total_hectares,

  COALESCE(ca.budget_cost_usd,0)::numeric(14,2)        AS budget_cost_usd,
  COALESCE(ca.executed_costs_usd,0)::numeric(14,2)     AS executed_costs_usd,
  COALESCE(ca.executed_labors_usd,0)::numeric(14,2)    AS executed_labors_usd,
  COALESCE(ca.executed_supplies_usd,0)::numeric(14,2)  AS executed_supplies_usd,

  COALESCE(o.income_usd,0)::numeric(14,2)             AS income_usd,
  COALESCE(o.direct_labors_usd,0)::numeric(14,2)      AS direct_labors_usd,
  (COALESCE(o.income_usd,0) - COALESCE(o.direct_labors_usd,0))::numeric(14,2) AS operating_result_usd,
  CASE WHEN COALESCE(o.direct_labors_usd,0) > 0
       THEN ROUND(((COALESCE(o.income_usd,0) - COALESCE(o.direct_labors_usd,0))
                  / NULLIF(o.direct_labors_usd,0) * 100)::numeric, 2)
       ELSE 0 END AS operating_result_pct,

  COALESCE(cm.contrib_progress_pct,0)::numeric(14,2) AS contributions_progress_pct,

  -- columnas de inversor vacías
  NULL::bigint        AS investor_id,
  NULL::text          AS investor_name,
  NULL::numeric(6,2)  AS investor_percentage,
  NULL::numeric(6,2)  AS investor_contribution_pct,

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
 AND h.field_id    IS NOT DISTINCT FROM lvl.field_id
LEFT JOIN costs_agg ca
  ON ca.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND ca.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND ca.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN operating_result o
  ON o.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND o.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND o.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
LEFT JOIN contrib_metric cm
  ON cm.customer_id IS NOT DISTINCT FROM lvl.customer_id
 AND cm.project_id  IS NOT DISTINCT FROM lvl.project_id
 AND cm.campaign_id IS NOT DISTINCT FROM lvl.campaign_id
 AND cm.field_id    IS NOT DISTINCT FROM lvl.field_id

UNION ALL

-- 2) Filas INVESTOR (métricas NULL para no duplicar agregados)
SELECT
  cr.customer_id,
  cr.project_id,
  cr.campaign_id,
  cr.field_id,

  NULL::numeric(14,2) AS sowing_hectares,
  NULL::numeric(14,2) AS sowing_total_hectares,

  NULL::numeric(14,2) AS harvest_hectares,
  NULL::numeric(14,2) AS harvest_total_hectares,

  NULL::numeric(14,2) AS budget_cost_usd,
  NULL::numeric(14,2) AS executed_costs_usd,
  NULL::numeric(14,2) AS executed_labors_usd,
  NULL::numeric(14,2) AS executed_supplies_usd,

  NULL::numeric(14,2) AS income_usd,
  NULL::numeric(14,2) AS direct_labors_usd,
  NULL::numeric(14,2) AS operating_result_usd,
  NULL::numeric(14,2) AS operating_result_pct,

  100.0::numeric(14,2) AS contributions_progress_pct,

  cr.investor_id,
  cr.investor_name,
  cr.investor_percentage,
  cr.investor_contribution_pct,

  'investor'::text AS row_kind
FROM contrib_rows cr;
