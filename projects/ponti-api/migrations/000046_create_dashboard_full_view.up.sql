CREATE OR REPLACE FUNCTION get_dashboard_payload(
  p_customer_id  BIGINT DEFAULT NULL,
  p_project_id   BIGINT DEFAULT NULL,
  p_campaign_id  BIGINT DEFAULT NULL,
  p_field_id     BIGINT DEFAULT NULL
) RETURNS JSONB
LANGUAGE sql
STABLE
AS $$
WITH
-- =========================================================
-- 0) Filtros opcionales (si vienen NULL, no filtran nada)
-- =========================================================
flt AS (
  SELECT
    p_customer_id  AS customer_id,
    p_project_id   AS project_id,
    p_campaign_id  AS campaign_id,
    p_field_id     AS field_id
),

-- =========================================================
-- 1) MÉTRICAS (5 cards)
-- =========================================================
sowing AS (
  SELECT
    COALESCE(SUM(v.sowed_area),0)::numeric(14,2)     AS hectares,
    COALESCE(SUM(v.total_hectares),0)::numeric(14,2) AS total_hectares
  FROM dashboard_card_sowing_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    AND (f.field_id    IS NULL OR v.field_id    = f.field_id)
),
sowing_kpi AS (
  SELECT
    CASE WHEN s.total_hectares > 0
         THEN ROUND((s.hectares / s.total_hectares * 100)::numeric, 2)
         ELSE 0 END AS progress_pct,
    s.hectares, s.total_hectares
  FROM sowing s
),
harvest AS (
  SELECT
    COALESCE(SUM(v.harvested_area),0)::numeric(14,2) AS hectares,
    COALESCE(SUM(v.total_hectares),0)::numeric(14,2) AS total_hectares
  FROM dashboard_card_harvest_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    AND (f.field_id    IS NULL OR v.field_id    = f.field_id)
),
harvest_kpi AS (
  SELECT
    CASE WHEN h.total_hectares > 0
         THEN ROUND((h.hectares / h.total_hectares * 100)::numeric, 2)
         ELSE 0 END AS progress_pct,
    h.hectares, h.total_hectares
  FROM harvest h
),
costs AS (
  SELECT
    COALESCE(SUM(v.budget_cost_usd),0)::numeric(14,2)    AS budget_usd,
    COALESCE(SUM(v.executed_costs_usd),0)::numeric(14,2) AS executed_usd
  FROM dashboard_card_costs_progress_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    AND (f.field_id    IS NULL OR v.field_id    = f.field_id)
),
costs_kpi AS (
  SELECT
    CASE WHEN c.budget_usd > 0
         THEN ROUND((c.executed_usd / c.budget_usd * 100)::numeric, 2)
         ELSE 0 END AS progress_pct,
    c.executed_usd, c.budget_usd
  FROM costs c
),
contribs AS (
  SELECT
    COALESCE(AVG(v.contributions_progress_pct),0)::numeric(14,2) AS progress_pct,
    NULLIF(jsonb_agg(v.investors_breakdown) FILTER (WHERE v.investors_breakdown IS NOT NULL), '[]'::jsonb) AS breakdowns
  FROM dashboard_card_contributions_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    AND (f.field_id    IS NULL OR v.field_id    = f.field_id)
),
oper_card AS (
  SELECT
    COALESCE(SUM(v.income_usd),0)::numeric(14,2)              AS income_usd,
    COALESCE(SUM(v.direct_costs_executed_usd),0)::numeric(14,2) AS direct_costs_executed_usd,
    -- % resultado sobre invertido directo
    CASE WHEN SUM(v.direct_costs_executed_usd) IS NOT NULL THEN
      COALESCE(AVG(v.operating_result_pct),0)::numeric(14,2)
    ELSE 0 END AS operating_result_pct,
    COALESCE(SUM(v.operating_result_usd),0)::numeric(14,2)    AS operating_result_usd
  FROM dashboard_card_operating_result_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    AND (f.field_id    IS NULL OR v.field_id    = f.field_id)
),

-- =========================================================
-- 2) BALANCE DE GESTIÓN
-- =========================================================
bal AS (
  SELECT
    COALESCE(SUM(direct_costs_executed_usd),0)::numeric(14,2) AS direct_costs_executed_usd,
    COALESCE(SUM(direct_costs_invested_usd),0)::numeric(14,2) AS direct_costs_invested_usd,
    COALESCE(SUM(stock_usd),0)::numeric(14,2)                 AS stock_usd,
    COALESCE(SUM(rent_usd),0)::numeric(14,2)                  AS rent_usd,
    COALESCE(SUM(structure_usd),0)::numeric(14,2)             AS structure_usd,
    COALESCE(SUM(seed_executed_usd),0)::numeric(14,2)         AS seed_exec,
    COALESCE(SUM(seed_invested_usd),0)::numeric(14,2)         AS seed_inv,
    COALESCE(SUM(supplies_executed_usd),0)::numeric(14,2)     AS supplies_exec,
    COALESCE(SUM(supplies_invested_usd),0)::numeric(14,2)     AS supplies_inv,
    COALESCE(SUM(labors_executed_usd),0)::numeric(14,2)       AS labors_exec,
    COALESCE(SUM(labors_invested_usd),0)::numeric(14,2)       AS labors_inv
  FROM dashboard_balance_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    AND (f.field_id    IS NULL OR v.field_id    = f.field_id)
),

-- =========================================================
-- 3) INCIDENCIA POR CULTIVO
-- =========================================================
ci_rows AS (  -- filas por cultivo
  SELECT
    v.crop_name AS name,
    COALESCE(SUM(v.surface_has),0)::numeric(14,2)       AS hectares,
    ROUND(COALESCE(AVG(v.rotation_pct),0)::numeric,2)   AS rotation_pct,
    ROUND(COALESCE(AVG(v.cost_usd_per_ha),0)::numeric,2) AS cost_usd_per_ha,
    ROUND(COALESCE(AVG(v.incidence_pct),0)::numeric,2)   AS incidence_pct
  FROM dashboard_crop_incidence_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    AND (f.field_id    IS NULL OR v.field_id    = f.field_id)
  GROUP BY v.crop_name
),
ci_total AS (  -- totales del bloque
  SELECT
    COALESCE(SUM(hectares),0)::numeric(14,2) AS hectares,
    CASE WHEN COALESCE(SUM(hectares),0) > 0
         THEN ROUND(100::numeric,2)
         ELSE 0 END                          AS rotation_pct,
    CASE WHEN COALESCE(SUM(hectares),0) > 0
         THEN ROUND( (SUM(hectares * cost_usd_per_ha) / NULLIF(SUM(hectares),0))::numeric , 2 )
         ELSE 0 END                          AS cost_usd_per_hectare
  FROM ci_rows
),

-- =========================================================
-- 4) INDICADORES OPERATIVOS (4 cards)
-- =========================================================
ops AS (
  SELECT
    MIN(first_workorder_date) AS first_workorder_date,
    (ARRAY_AGG(first_workorder_id  ORDER BY first_workorder_date  ASC))[1] AS first_workorder_id,
    MAX(last_workorder_date)  AS last_workorder_date,
    (ARRAY_AGG(last_workorder_id   ORDER BY last_workorder_date   DESC))[1] AS last_workorder_id,
    MAX(last_stock_audit_date) AS last_stock_audit_date,
    MAX(campaign_close_date)  AS campaign_close_date
  FROM dashboard_operational_indicators_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    -- esta vista es por proyecto, sin field_id
)

-- =========================================================
-- 5) COMPOSICIÓN DEL JSON
-- =========================================================
SELECT jsonb_build_object(
  'metrics', jsonb_build_object(
    'sowing', jsonb_build_object(
      'progress_pct', sk.progress_pct,
      'hectares',     sk.hectares,
      'total_hectares', sk.total_hectares
    ),
    'harvest', jsonb_build_object(
      'progress_pct', hk.progress_pct,
      'hectares',     hk.hectares,
      'total_hectares', hk.total_hectares
    ),
    'costs', jsonb_build_object(
      'progress_pct', ck.progress_pct,
      'executed_usd', ck.executed_usd,
      'budget_usd',   ck.budget_usd
    ),
    'investor_contributions', jsonb_build_object(
      'progress_pct', COALESCE(c.progress_pct, 0),
      'breakdown',    COALESCE((SELECT jsonb_agg(x) FROM jsonb_array_elements(c.breakdowns)), NULL)
    ),
    'operating_result', jsonb_build_object(
      'progress_pct',     oc.operating_result_pct,
      'income_usd',       oc.income_usd,
      'total_costs_usd',  oc.direct_costs_executed_usd
    )
  ),

  'management_balance', jsonb_build_object(
    'summary', jsonb_build_object(
      'income_usd',                 oc.income_usd,
      'direct_costs_executed_usd',  b.direct_costs_executed_usd,
      'direct_costs_invested_usd',  b.direct_costs_invested_usd,
      'stock_usd',                  b.stock_usd,
      'rent_usd',                   b.rent_usd,
      'structure_usd',              b.structure_usd,
      'operating_result_usd',       oc.operating_result_usd,
      'operating_result_pct',       oc.operating_result_pct
    ),
    'breakdown', jsonb_build_array(
      jsonb_build_object('label','Seed',      'executed_usd', b.seed_exec,     'invested_usd', b.seed_inv,     'stock_usd', NULL),
      jsonb_build_object('label','Supplies',  'executed_usd', b.supplies_exec, 'invested_usd', b.supplies_inv, 'stock_usd', NULL),
      jsonb_build_object('label','Labors',    'executed_usd', b.labors_exec,   'invested_usd', b.labors_inv,   'stock_usd', 0),
      jsonb_build_object('label','Rent',      'executed_usd', 0,               'invested_usd', b.rent_usd,     'stock_usd', 0),
      jsonb_build_object('label','Structure', 'executed_usd', 0,               'invested_usd', b.structure_usd,'stock_usd', 0)
    ),
    'totals_row', jsonb_build_object(
      'executed_usd', b.seed_exec + b.supplies_exec + b.labors_exec,
      'invested_usd', b.seed_inv  + b.supplies_inv  + b.labors_inv + b.rent_usd + b.structure_usd,
      'stock_usd',    b.stock_usd
    )
  ),

  'crop_incidence', jsonb_build_object(
    'crops', COALESCE((
      SELECT jsonb_agg(
        jsonb_build_object(
          'name', r.name,
          'hectares', r.hectares,
          'rotation_pct', r.rotation_pct,
          'cost_usd_per_ha', r.cost_usd_per_ha,
          'incidence_pct', r.incidence_pct
        )
        ORDER BY r.name
      ) FROM ci_rows r
    ), '[]'::jsonb),
    'total', jsonb_build_object(
      'hectares', ci.hectares,
      'rotation_pct', ci.rotation_pct,
      'cost_usd_per_hectare', ci.cost_usd_per_hectare
    )
  ),

  'operational_indicators', jsonb_build_object(
    'cards', jsonb_build_array(
      jsonb_build_object('key','first_workorder', 'title','Primera orden de trabajo',
                         'date', ops.first_workorder_date, 'workorder_id', ops.first_workorder_id, 'workorder_code', NULL),
      jsonb_build_object('key','last_workorder',  'title','Última orden de trabajo',
                         'date', ops.last_workorder_date,  'workorder_id', ops.last_workorder_id,  'workorder_code', NULL),
      jsonb_build_object('key','last_stock_audit','title','Último arqueo de stock',
                         'date', ops.last_stock_audit_date, 'audit_id', NULL, 'audit_code', NULL),
      jsonb_build_object('key','campaign_close',  'title','Cierre de campaña',
                         'date', ops.campaign_close_date, 'status', CASE WHEN ops.campaign_close_date IS NULL THEN 'pending' ELSE 'closed' END)
    )
  )
)
FROM sowing_kpi sk, harvest_kpi hk, costs_kpi ck, contribs c, oper_card oc, bal b, ci_total ci, ops;
$$;
