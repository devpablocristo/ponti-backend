-- =========================================================
-- Agregar cálculo de crop_incidence a la función del dashboard
-- =========================================================

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
-- 1) MÉTRICAS (5 cards) - calculadas directamente
-- =========================================================
sowing AS (
  SELECT
    COALESCE(SUM(CASE WHEN l.sowing_date IS NOT NULL THEN l.hectares ELSE 0 END), 0)::numeric(14,2)     AS hectares,
    COALESCE(SUM(l.hectares), 0)::numeric(14,2) AS total_hectares
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id, flt f2
  WHERE (f2.customer_id IS NULL OR p.customer_id = f2.customer_id)
    AND (f2.project_id  IS NULL OR p.id  = f2.project_id)
    AND (f2.campaign_id IS NULL OR p.campaign_id = f2.campaign_id)
    AND (f2.field_id    IS NULL OR f.id = f2.field_id)
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
    COALESCE(SUM(CASE WHEN l.tons IS NOT NULL AND l.tons > 0 THEN l.hectares ELSE 0 END), 0)::numeric(14,2) AS hectares,
    COALESCE(SUM(l.hectares), 0)::numeric(14,2) AS total_hectares
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id, flt f2
  WHERE (f2.customer_id IS NULL OR p.customer_id = f2.customer_id)
    AND (f2.project_id  IS NULL OR p.id  = f2.project_id)
    AND (f2.campaign_id IS NULL OR p.campaign_id = f2.campaign_id)
    AND (f2.field_id    IS NULL OR f.id = f2.field_id)
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
    COALESCE(SUM(p.admin_cost), 0)::numeric(14,2)    AS budget_usd,
    COALESCE(SUM(COALESCE(l.price, 0)), 0)::numeric(14,2) AS executed_usd
  FROM projects p
  LEFT JOIN labors l ON l.project_id = p.id, flt f2
  WHERE (f2.customer_id IS NULL OR p.customer_id = f2.customer_id)
    AND (f2.project_id  IS NULL OR p.id  = f2.project_id)
    AND (f2.campaign_id IS NULL OR p.campaign_id = f2.campaign_id)
  GROUP BY p.customer_id, p.id, p.campaign_id
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
    100.0::numeric(14,2) AS progress_pct,
    jsonb_agg(
      jsonb_build_object(
        'investor_id', pi.investor_id,
        'investor_name', i.name,
        'percentage', pi.percentage
      )
    ) AS breakdowns
  FROM projects p
  JOIN project_investors pi ON pi.project_id = p.id
  JOIN investors i ON i.id = pi.investor_id, flt f2
  WHERE (f2.customer_id IS NULL OR p.customer_id = f2.customer_id)
    AND (f2.project_id  IS NULL OR p.id  = f2.project_id)
    AND (f2.campaign_id IS NULL OR p.campaign_id = f2.campaign_id)
),
oper_card AS (
  SELECT
    COALESCE(SUM(l.tons * 200), 0)::numeric(14,2)              AS income_usd,
    COALESCE(SUM(COALESCE(lab.price, 0)), 0)::numeric(14,2) AS direct_costs_executed_usd,
    CASE WHEN SUM(COALESCE(lab.price, 0)) > 0 THEN
      ROUND(((SUM(l.tons * 200) - SUM(COALESCE(lab.price, 0))) / SUM(COALESCE(lab.price, 0)) * 100)::numeric, 2)
    ELSE 0 END AS operating_result_pct,
    COALESCE(SUM(l.tons * 200) - SUM(COALESCE(lab.price, 0)), 0)::numeric(14,2)    AS operating_result_usd
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id
  LEFT JOIN labors lab ON lab.project_id = p.id, flt f2
  WHERE (f2.customer_id IS NULL OR p.customer_id = f2.customer_id)
    AND (f2.project_id  IS NULL OR p.id  = f2.project_id)
    AND (f2.campaign_id IS NULL OR p.campaign_id = f2.campaign_id)
    AND (f2.field_id    IS NULL OR f.id = f2.field_id)
),

-- =========================================================
-- 2) CÁLCULO DE CROP_INCIDENCE
-- =========================================================
crop_data AS (
  SELECT
    c.id as crop_id,
    c.name as crop_name,
    COALESCE(SUM(l.hectares), 0)::numeric(14,2) as hectares,
    COALESCE(SUM(l.tons), 0)::numeric(14,2) as tons,
    COALESCE(SUM(l.tons * 200), 0)::numeric(14,2) as income_usd,
    COALESCE(SUM(COALESCE(lab.price, 0)), 0)::numeric(14,2) as costs_usd
  FROM projects p
  JOIN fields f ON f.project_id = p.id
  JOIN lots l ON l.field_id = f.id
  JOIN crops c ON l.current_crop_id = c.id
  LEFT JOIN labors lab ON lab.project_id = p.id, flt f2
  WHERE (f2.customer_id IS NULL OR p.customer_id = f2.customer_id)
    AND (f2.project_id  IS NULL OR p.id  = f2.project_id)
    AND (f2.campaign_id IS NULL OR p.campaign_id = f2.campaign_id)
    AND (f2.field_id    IS NULL OR f.id = f2.field_id)
    AND c.deleted_at IS NULL
  GROUP BY c.id, c.name
),
crop_incidence AS (
  SELECT
    jsonb_agg(
      jsonb_build_object(
        'crop_id', cd.crop_id,
        'crop_name', cd.crop_name,
        'hectares', cd.hectares,
        'tons', cd.tons,
        'income_usd', cd.income_usd,
        'costs_usd', cd.costs_usd,
        'cost_usd_per_hectare', CASE 
          WHEN cd.hectares > 0 THEN ROUND((cd.costs_usd / cd.hectares)::numeric, 2)
          ELSE 0 
        END
      )
    ) as crops,
    COALESCE(SUM(cd.hectares), 0)::numeric(14,2) as total_hectares,
    100.0::numeric(14,2) as rotation_pct,
    CASE 
      WHEN SUM(cd.hectares) > 0 THEN ROUND((SUM(cd.costs_usd) / SUM(cd.hectares))::numeric, 2)
      ELSE 0 
    END as cost_usd_per_hectare
  FROM crop_data cd
)

-- =========================================================
-- 3) COMPOSICIÓN DEL JSON COMPLETO
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
      'breakdown',    COALESCE(c.breakdowns, NULL)
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
      'direct_costs_executed_usd',  oc.direct_costs_executed_usd,
      'direct_costs_invested_usd',  0,
      'stock_usd',                  0,
      'rent_usd',                   0,
      'structure_usd',              0,
      'operating_result_usd',       oc.operating_result_usd,
      'operating_result_pct',       oc.operating_result_pct
    ),
    'breakdown', jsonb_build_array(
      jsonb_build_object('label','Seed',      'executed_usd', 0, 'invested_usd', 0, 'stock_usd', NULL),
      jsonb_build_object('label','Supplies',  'executed_usd', 0, 'invested_usd', 0, 'stock_usd', NULL),
      jsonb_build_object('label','Labors',    'executed_usd', oc.direct_costs_executed_usd, 'invested_usd', 0, 'stock_usd', 0),
      jsonb_build_object('label','Rent',      'executed_usd', 0, 'invested_usd', 0, 'stock_usd', 0),
      jsonb_build_object('label','Structure', 'executed_usd', 0, 'invested_usd', 0, 'stock_usd', 0)
    ),
    'totals_row', jsonb_build_object(
      'executed_usd', oc.direct_costs_executed_usd,
      'invested_usd', 0,
      'stock_usd',    0
    )
  ),

  'crop_incidence', jsonb_build_object(
    'crops', COALESCE(ci.crops, '[]'::jsonb),
    'total', jsonb_build_object(
      'hectares', ci.total_hectares,
      'rotation_pct', ci.rotation_pct,
      'cost_usd_per_hectare', ci.cost_usd_per_hectare
    )
  ),

  'operational_indicators', jsonb_build_object(
    'cards', jsonb_build_array(
      jsonb_build_object('key','first_workorder', 'title','Primera orden de trabajo',
                         'date', NULL, 'workorder_id', NULL, 'workorder_code', NULL),
      jsonb_build_object('key','last_workorder',  'title','Última orden de trabajo',
                         'date', NULL, 'workorder_id', NULL, 'workorder_code', NULL),
      jsonb_build_object('key','last_stock_audit','title','Último arqueo de stock',
                         'date', NULL, 'audit_id', NULL, 'audit_code', NULL),
      jsonb_build_object('key','campaign_close',  'title','Cierre de campaña',
                         'date', NULL, 'status', 'pending')
    )
  )
)
FROM sowing_kpi sk, harvest_kpi hk, costs_kpi ck, contribs c, oper_card oc, crop_incidence ci;
$$;
