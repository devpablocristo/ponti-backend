-- =========================================================
-- MIGRACIÓN 000047: ROLLBACK - Restaurar get_dashboard_payload original
-- =========================================================
-- FECHA: 2025-01-29
-- DESCRIPCIÓN: Restaurar la función get_dashboard_payload original sin crop_incidence real
-- =========================================================

-- Restaurar la función get_dashboard_payload original (sin crop_incidence real)
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
-- 1) MÉTRICAS (5 cards) - usando solo las vistas que existen
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
    COALESCE(jsonb_agg(v.investors_breakdown) FILTER (WHERE v.investors_breakdown IS NOT NULL), '[]'::jsonb) AS breakdowns
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
    CASE WHEN SUM(v.direct_costs_executed_usd) IS NOT NULL AND SUM(v.direct_costs_executed_usd) > 0 THEN
      COALESCE(AVG(v.operating_result_pct),0)::numeric(14,2)
    ELSE 0 END AS operating_result_pct,
    COALESCE(SUM(v.operating_result_usd),0)::numeric(14,2)    AS operating_result_usd
  FROM dashboard_card_operating_result_view v, flt f
  WHERE (f.customer_id IS NULL OR v.customer_id = f.customer_id)
    AND (f.project_id  IS NULL OR v.project_id  = f.project_id)
    AND (f.campaign_id IS NULL OR v.campaign_id = f.campaign_id)
    AND (f.field_id    IS NULL OR v.field_id    = f.field_id)
)

-- =========================================================
-- 2) COMPOSICIÓN DEL JSON SIMPLIFICADO - SIEMPRE DEVUELVE UN JSON VÁLIDO
-- =========================================================
SELECT COALESCE(
  jsonb_build_object(
    'metrics', jsonb_build_object(
      'sowing', jsonb_build_object(
        'progress_pct', COALESCE(sk.progress_pct, 0),
        'hectares',     COALESCE(sk.hectares, 0),
        'total_hectares', COALESCE(sk.total_hectares, 0)
      ),
      'harvest', jsonb_build_object(
        'progress_pct', COALESCE(hk.progress_pct, 0),
        'hectares',     COALESCE(hk.hectares, 0),
        'total_hectares', COALESCE(hk.total_hectares, 0)
      ),
      'costs', jsonb_build_object(
        'progress_pct', COALESCE(ck.progress_pct, 0),
        'executed_usd', COALESCE(ck.executed_usd, 0),
        'budget_usd',   COALESCE(ck.budget_usd, 0)
      ),
      'investor_contributions', jsonb_build_object(
        'progress_pct', COALESCE(c.progress_pct, 0),
        'breakdown',    COALESCE(c.breakdowns, '[]'::jsonb)
      ),
      'operating_result', jsonb_build_object(
        'progress_pct',     COALESCE(oc.operating_result_pct, 0),
        'income_usd',       COALESCE(oc.income_usd, 0),
        'total_costs_usd',  COALESCE(oc.direct_costs_executed_usd, 0)
      )
    ),

    'management_balance', jsonb_build_object(
      'summary', jsonb_build_object(
        'income_usd',                 COALESCE(oc.income_usd, 0),
        'direct_costs_executed_usd',  COALESCE(oc.direct_costs_executed_usd, 0),
        'direct_costs_invested_usd',  0,
        'stock_usd',                  0,
        'rent_usd',                   0,
        'structure_usd',              0,
        'operating_result_usd',       COALESCE(oc.operating_result_usd, 0),
        'operating_result_pct',       COALESCE(oc.operating_result_pct, 0)
      ),
      'breakdown', jsonb_build_array(
        jsonb_build_object('label','Seed',      'executed_usd', 0, 'invested_usd', 0, 'stock_usd', NULL),
        jsonb_build_object('label','Supplies',  'executed_usd', 0, 'invested_usd', 0, 'stock_usd', NULL),
        jsonb_build_object('label','Labors',    'executed_usd', COALESCE(oc.direct_costs_executed_usd, 0), 'invested_usd', 0, 'stock_usd', 0),
        jsonb_build_object('label','Rent',      'executed_usd', 0, 'invested_usd', 0, 'stock_usd', 0),
        jsonb_build_object('label','Structure', 'executed_usd', 0, 'invested_usd', 0, 'stock_usd', 0)
      ),
      'totals_row', jsonb_build_object(
        'executed_usd', COALESCE(oc.direct_costs_executed_usd, 0),
        'invested_usd', 0,
        'stock_usd',    0
      )
    ),

    'crop_incidence', jsonb_build_object(
      'crops', '[]'::jsonb,
      'total', jsonb_build_object(
        'hectares', 0,
        'rotation_pct', 0,
        'cost_usd_per_hectare', 0
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
  ),
  -- JSON por defecto si todo falla
  '{"metrics":{"sowing":{"progress_pct":0,"hectares":0,"total_hectares":0},"harvest":{"progress_pct":0,"hectares":0,"total_hectares":0},"costs":{"progress_pct":0,"executed_usd":0,"budget_usd":0},"investor_contributions":{"progress_pct":0,"breakdown":[]},"operating_result":{"progress_pct":0,"income_usd":0,"total_costs_usd":0}},"management_balance":{"summary":{"income_usd":0,"direct_costs_executed_usd":0,"direct_costs_invested_usd":0,"stock_usd":0,"rent_usd":0,"structure_usd":0,"operating_result_usd":0,"operating_result_pct":0},"breakdown":[{"label":"Seed","executed_usd":0,"invested_usd":0,"stock_usd":null},{"label":"Supplies","executed_usd":0,"invested_usd":0,"stock_usd":null},{"label":"Labors","executed_usd":0,"invested_usd":0,"stock_usd":0},{"label":"Rent","executed_usd":0,"invested_usd":0,"stock_usd":0},{"label":"Structure","executed_usd":0,"invested_usd":0,"stock_usd":0}],"totals_row":{"executed_usd":0,"invested_usd":0,"stock_usd":0}},"crop_incidence":{"crops":[],"total":{"hectares":0,"rotation_pct":0,"cost_usd_per_hectare":0}},"operational_indicators":{"cards":[{"key":"first_workorder","title":"Primera orden de trabajo","date":null,"workorder_id":null,"workorder_code":null},{"key":"last_workorder","title":"Última orden de trabajo","date":null,"workorder_code":null},{"key":"last_stock_audit","title":"Último arqueo de stock","date":null,"audit_id":null,"audit_code":null},{"key":"campaign_close","title":"Cierre de campaña","date":null,"status":"pending"}]}}'::jsonb
) as payload
FROM sowing_kpi sk, harvest_kpi hk, costs_kpi ck, contribs c, oper_card oc;
$$;
