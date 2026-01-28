-- ========================================
-- MIGRACIÓN 000334: Consolidar v3_calc dominio -> v3_core_ssot (UP)
-- ========================================
--
-- Propósito: Unificar funciones de dominio en v3_calc para que deleguen en v3_core_ssot.
-- Enfoque: v3_calc queda como wrapper de v3_core_ssot.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Helpers de fechas
CREATE OR REPLACE FUNCTION v3_calc.calculate_campaign_closing_date(end_date date)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT v3_core_ssot.calculate_campaign_closing_date(end_date)
$$;

-- Helpers por hectárea / dosis
CREATE OR REPLACE FUNCTION v3_calc.dose_per_ha(total_dose numeric, surface_ha numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.dose_per_ha(total_dose, surface_ha)
$$;

CREATE OR REPLACE FUNCTION v3_calc.units_per_ha(units numeric, area numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.units_per_ha(units, area)
$$;

CREATE OR REPLACE FUNCTION v3_calc.norm_dose(dose numeric, area numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.norm_dose(dose, area)
$$;

-- Áreas
CREATE OR REPLACE FUNCTION v3_calc.seeded_area(sowing_date date, hectares numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.seeded_area(sowing_date, hectares)
$$;

CREATE OR REPLACE FUNCTION v3_calc.harvested_area(tons numeric, hectares numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.harvested_area(tons, hectares)
$$;

-- Rendimientos
CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.yield_tn_per_ha_over_hectares(tons, hectares)
$$;

CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.yield_tn_per_ha_over_harvested(tons, harvested_area)
$$;

-- Costos
CREATE OR REPLACE FUNCTION v3_calc.labor_cost(labor_price numeric, effective_area numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.labor_cost(labor_price, effective_area)
$$;

CREATE OR REPLACE FUNCTION v3_calc.supply_cost(final_dose double precision, supply_price numeric, effective_area numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.supply_cost(final_dose, supply_price, effective_area)
$$;

CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha(total_cost numeric, hectares numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.cost_per_ha(total_cost, hectares)
$$;

-- Ingresos
CREATE OR REPLACE FUNCTION v3_calc.income_net_total(tons numeric, net_price_usd numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.income_net_total(tons, net_price_usd)
$$;

CREATE OR REPLACE FUNCTION v3_calc.income_net_per_ha(income_net_total numeric, hectares numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.income_net_per_ha(income_net_total, hectares)
$$;

-- Renta por ha (sobrecargas)
CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha(
  lease_type_id integer,
  lease_type_percent double precision,
  lease_type_value double precision,
  income_net_per_ha double precision,
  cost_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.rent_per_ha(
    lease_type_id,
    lease_type_percent,
    lease_type_value,
    income_net_per_ha,
    cost_per_ha,
    admin_cost_per_ha
  )
$$;

CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha(
  lease_type_id bigint,
  lease_type_percent double precision,
  lease_type_value double precision,
  income_net_per_ha double precision,
  cost_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.rent_per_ha(
    lease_type_id,
    lease_type_percent,
    lease_type_value,
    income_net_per_ha,
    cost_per_ha,
    admin_cost_per_ha
  )
$$;

-- Activo total y resultado operativo
CREATE OR REPLACE FUNCTION v3_calc.active_total_per_ha(
  direct_cost_per_ha double precision,
  rent_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.active_total_per_ha(direct_cost_per_ha, rent_per_ha, admin_cost_per_ha)
$$;

CREATE OR REPLACE FUNCTION v3_calc.operating_result_per_ha(
  income_net_per_ha double precision,
  active_total_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.operating_result_per_ha(income_net_per_ha, active_total_per_ha)
$$;

-- Precio de indiferencia
CREATE OR REPLACE FUNCTION v3_calc.indifference_price_usd_tn(
  total_invested_per_ha double precision,
  yield_tn_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_core_ssot.indifference_price_usd_tn(total_invested_per_ha, yield_tn_per_ha)
$$;

COMMIT;
