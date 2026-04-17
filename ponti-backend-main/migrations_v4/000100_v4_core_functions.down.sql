-- ========================================
-- MIGRATION 000100 V4 CORE FUNCTIONS (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP FUNCTION IF EXISTS v4_core.indifference_price_usd_tn(
  total_invested_per_ha numeric, 
  yield_tn_per_ha numeric
);
DROP FUNCTION IF EXISTS v4_core.operating_result_per_ha(
  income_net_per_ha numeric,
  active_total_per_ha numeric
);
DROP FUNCTION IF EXISTS v4_core.active_total_per_ha(
  direct_cost_per_ha numeric,
  rent_per_ha numeric,
  admin_cost_per_ha numeric
);
DROP FUNCTION IF EXISTS v4_core.calculate_rent_per_ha(lease_value numeric);
DROP FUNCTION IF EXISTS v4_core.rent_per_ha(
  lease_type_id bigint,
  lease_type_percent numeric,
  lease_type_value numeric,
  income_net_per_ha numeric,
  cost_per_ha numeric,
  admin_cost_per_ha numeric
);
DROP FUNCTION IF EXISTS v4_core.rent_per_ha(
  lease_type_id integer,
  lease_type_percent numeric,
  lease_type_value numeric,
  income_net_per_ha numeric,
  cost_per_ha numeric,
  admin_cost_per_ha numeric
);
DROP FUNCTION IF EXISTS v4_core.income_net_per_ha(income_net_total numeric, hectares numeric);
DROP FUNCTION IF EXISTS v4_core.income_net_total(tons numeric, net_price_usd numeric);
DROP FUNCTION IF EXISTS v4_core.cost_per_ha(total_cost numeric, hectares numeric);
DROP FUNCTION IF EXISTS v4_core.supply_cost(final_dose numeric, supply_price numeric, effective_area numeric);
DROP FUNCTION IF EXISTS v4_core.labor_cost(labor_price numeric, effective_area numeric);
DROP FUNCTION IF EXISTS v4_core.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric);
DROP FUNCTION IF EXISTS v4_core.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric);
DROP FUNCTION IF EXISTS v4_core.harvested_area(tons numeric, hectares numeric);
DROP FUNCTION IF EXISTS v4_core.seeded_area(sowing_date date, hectares numeric);
DROP FUNCTION IF EXISTS v4_core.dollar_average_for_month(p_project_id bigint, p_date date);
DROP FUNCTION IF EXISTS v4_core.get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR);
DROP FUNCTION IF EXISTS v4_core.calculate_campaign_closing_date(end_date date);
DROP FUNCTION IF EXISTS v4_core.norm_dose(dose numeric, area numeric);
DROP FUNCTION IF EXISTS v4_core.dose_per_ha(total_dose numeric, surface_ha numeric);
DROP FUNCTION IF EXISTS v4_core.units_per_ha(units numeric, area numeric);
DROP FUNCTION IF EXISTS v4_core.per_ha(numeric, double precision);
DROP FUNCTION IF EXISTS v4_core.per_ha(double precision, numeric);
DROP FUNCTION IF EXISTS v4_core.per_ha_dp(double precision, double precision);
DROP FUNCTION IF EXISTS v4_core.per_ha(value numeric, area_ha numeric);
DROP FUNCTION IF EXISTS v4_core.percentage_rounded(part numeric, total numeric);
DROP FUNCTION IF EXISTS v4_core.percentage_capped(part numeric, total numeric);
DROP FUNCTION IF EXISTS v4_core.percentage(part numeric, total numeric);
DROP FUNCTION IF EXISTS v4_core.safe_div_dp(numerator double precision, denominator double precision);
DROP FUNCTION IF EXISTS v4_core.safe_div(numerator numeric, denominator numeric);
DROP FUNCTION IF EXISTS v4_core.coalesce0(double precision);
DROP FUNCTION IF EXISTS v4_core.coalesce0(numeric);

COMMIT;
