-- ========================================
-- MIGRATION 000090 V4 CORE FUNCTIONS.UP (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

CREATE OR REPLACE FUNCTION v4_core.coalesce0(numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION v4_core.coalesce0(double precision) 
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1, 0)
$$;

CREATE OR REPLACE FUNCTION v4_core.safe_div(numerator numeric, denominator numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION v4_core.safe_div_dp(numerator double precision, denominator double precision) 
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$$;

CREATE OR REPLACE FUNCTION v4_core.percentage(part numeric, total numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div($1, $2) * 100
$$;

CREATE OR REPLACE FUNCTION v4_core.percentage_capped(part numeric, total numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT LEAST(v4_core.safe_div($1, $2) * 100, 100)
$$;

CREATE OR REPLACE FUNCTION v4_core.percentage_rounded(part numeric, total numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div($1, $2) * 100
$$;

CREATE OR REPLACE FUNCTION v4_core.per_ha(value numeric, area_ha numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div($1, $2)
$$;

CREATE OR REPLACE FUNCTION v4_core.per_ha_dp(double precision, double precision) 
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div_dp($1, $2)
$$;

CREATE OR REPLACE FUNCTION v4_core.per_ha(double precision, numeric)
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div_dp($1, $2::double precision)
$$;

CREATE OR REPLACE FUNCTION v4_core.per_ha(numeric, double precision)
RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div_dp($1::double precision, $2)
$$;

CREATE OR REPLACE FUNCTION v4_core.units_per_ha(units numeric, area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.per_ha(units, area)
$$;

CREATE OR REPLACE FUNCTION v4_core.dose_per_ha(total_dose numeric, surface_ha numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div(total_dose, surface_ha)
$$;

CREATE OR REPLACE FUNCTION v4_core.norm_dose(dose numeric, area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN area > 0 THEN dose / area ELSE NULL END
$$;

CREATE OR REPLACE FUNCTION v4_core.calculate_campaign_closing_date(end_date date) 
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT CASE 
    WHEN end_date IS NULL THEN NULL
    ELSE end_date + (get_campaign_closure_days() || ' days')::INTERVAL
  END::date
$$;

CREATE OR REPLACE FUNCTION v4_core.get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR)
RETURNS NUMERIC AS $$
DECLARE
  dollar_value NUMERIC;
BEGIN
  SELECT d.average_value INTO dollar_value
  FROM project_dollar_values d
  WHERE d.project_id = p_project_id 
    AND d.month = p_month
    AND d.deleted_at IS NULL
  LIMIT 1;
  
  RETURN COALESCE(dollar_value, 1.0);
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION v4_core.dollar_average_for_month(p_project_id bigint, p_date date) 
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    (SELECT AVG(d.average_value)
     FROM project_dollar_values d
     WHERE d.project_id = p_project_id
       AND TO_CHAR(p_date, 'YYYY-MM') = d.month
       AND d.deleted_at IS NULL),
    1.0
  )
$$;

CREATE OR REPLACE FUNCTION v4_core.seeded_area(sowing_date date, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN sowing_date IS NOT NULL THEN COALESCE(hectares,0) ELSE 0 END
$$;

CREATE OR REPLACE FUNCTION v4_core.harvested_area(tons numeric, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN tons IS NOT NULL AND tons > 0 THEN COALESCE(hectares,0) ELSE 0 END
$$;

CREATE OR REPLACE FUNCTION v4_core.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div(COALESCE(tons,0), COALESCE(hectares,0))
$$;

CREATE OR REPLACE FUNCTION v4_core.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.safe_div(COALESCE(tons,0), COALESCE(harvested_area,0))
$$;

CREATE OR REPLACE FUNCTION v4_core.labor_cost(labor_price numeric, effective_area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(labor_price,0) * COALESCE(effective_area,0)
$$;

CREATE OR REPLACE FUNCTION v4_core.supply_cost(final_dose numeric, supply_price numeric, effective_area numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(final_dose,0)::numeric * COALESCE(supply_price,0) * COALESCE(effective_area,0)
$$;

CREATE OR REPLACE FUNCTION v4_core.cost_per_ha(total_cost numeric, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.per_ha(total_cost, hectares)
$$;

CREATE OR REPLACE FUNCTION v4_core.income_net_total(tons numeric, net_price_usd numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(tons,0) * COALESCE(net_price_usd,0)
$$;

CREATE OR REPLACE FUNCTION v4_core.income_net_per_ha(income_net_total numeric, hectares numeric) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.per_ha(income_net_total, hectares)
$$;

CREATE OR REPLACE FUNCTION v4_core.rent_per_ha(
  lease_type_id integer,
  lease_type_percent numeric,
  lease_type_value numeric,
  income_net_per_ha numeric,
  cost_per_ha numeric,
  admin_cost_per_ha numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT
    CASE
      WHEN lease_type_id = 1 THEN COALESCE(lease_type_percent,0)/100.0 * COALESCE(income_net_per_ha,0)
      WHEN lease_type_id = 2 THEN COALESCE(lease_type_percent,0)/100.0 *
                               (COALESCE(income_net_per_ha,0) - COALESCE(cost_per_ha,0) - COALESCE(admin_cost_per_ha,0))
      WHEN lease_type_id = 3 THEN COALESCE(lease_type_value,0)
      WHEN lease_type_id = 4 THEN COALESCE(lease_type_value,0) +
                               (COALESCE(lease_type_percent,0)/100.0 * COALESCE(income_net_per_ha,0))
      ELSE 0
    END
$$;

CREATE OR REPLACE FUNCTION v4_core.rent_per_ha(
  lease_type_id bigint,
  lease_type_percent numeric,
  lease_type_value numeric,
  income_net_per_ha numeric,
  cost_per_ha numeric,
  admin_cost_per_ha numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.rent_per_ha(
    lease_type_id::integer, 
    lease_type_percent, 
    lease_type_value,
    income_net_per_ha, 
    cost_per_ha, 
    admin_cost_per_ha
  )
$$;

CREATE OR REPLACE FUNCTION v4_core.calculate_rent_per_ha(lease_value numeric)
RETURNS numeric AS $$
BEGIN
  IF lease_value < 0 THEN
    RETURN 0;
  ELSE
    RETURN lease_value;
  END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION v4_core.active_total_per_ha(
  direct_cost_per_ha numeric,
  rent_per_ha numeric,
  admin_cost_per_ha numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(direct_cost_per_ha,0) + COALESCE(rent_per_ha,0) + COALESCE(admin_cost_per_ha,0)
$$;

CREATE OR REPLACE FUNCTION v4_core.operating_result_per_ha(
  income_net_per_ha numeric,
  active_total_per_ha numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(income_net_per_ha,0) - COALESCE(active_total_per_ha,0)
$$;

CREATE OR REPLACE FUNCTION v4_core.indifference_price_usd_tn(
  total_invested_per_ha numeric, 
  yield_tn_per_ha numeric
) RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v4_core.per_ha(total_invested_per_ha, yield_tn_per_ha)
$$;

COMMIT;
