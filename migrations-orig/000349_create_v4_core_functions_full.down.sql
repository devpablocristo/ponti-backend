-- ========================================
-- MIGRACIÓN 000349: Drop v4_core functions (full) (DOWN)
-- ========================================

BEGIN;
DROP FUNCTION IF EXISTS v4_core.indifference_price_usd_tn(total_invested_per_ha double precision, yield_tn_per_ha double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.operating_result_per_ha(income_net_per_ha double precision, active_total_per_ha double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.active_total_per_ha(direct_cost_per_ha double precision, rent_per_ha double precision, admin_cost_per_ha double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.calculate_rent_per_ha(lease_value DOUBLE PRECISION) CASCADE;
DROP FUNCTION IF EXISTS v4_core.rent_per_ha(lease_type_id bigint, lease_type_percent double precision, lease_type_value double precision, income_net_per_ha double precision, cost_per_ha double precision, admin_cost_per_ha double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.rent_per_ha(lease_type_id integer, lease_type_percent double precision, lease_type_value double precision, income_net_per_ha double precision, cost_per_ha double precision, admin_cost_per_ha double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.income_net_per_ha(income_net_total numeric, hectares numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.income_net_total(tons numeric, net_price_usd numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.cost_per_ha(total_cost numeric, hectares numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.supply_cost(final_dose double precision, supply_price numeric, effective_area numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.labor_cost(labor_price numeric, effective_area numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.harvested_area(tons numeric, hectares numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.seeded_area(sowing_date date, hectares numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.dollar_average_for_month(p_project_id bigint, p_date date) CASCADE;
DROP FUNCTION IF EXISTS v4_core.get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR) CASCADE;
DROP FUNCTION IF EXISTS v4_core.calculate_campaign_closing_date(end_date date) CASCADE;
DROP FUNCTION IF EXISTS v4_core.norm_dose(dose numeric, area numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.dose_per_ha(total_dose numeric, surface_ha numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.units_per_ha(units numeric, area numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.per_ha_dp(double precision, double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.per_ha(numeric, double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.per_ha(double precision, numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.per_ha(numeric, numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.percentage_rounded(numeric, numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.percentage_capped(numeric, numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.percentage(numeric, numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.safe_div_dp(double precision, double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.safe_div(numeric, numeric) CASCADE;
DROP FUNCTION IF EXISTS v4_core.coalesce0(double precision) CASCADE;
DROP FUNCTION IF EXISTS v4_core.coalesce0(numeric) CASCADE;
COMMIT;
