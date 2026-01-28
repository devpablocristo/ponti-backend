-- ========================================
-- MIGRACIÓN 000334: Consolidar v3_calc dominio -> v3_core_ssot (DOWN)
-- ========================================
--
-- Propósito: Revertir wrappers y restaurar definiciones locales en v3_calc.
-- Nota: Comentarios en español, código en inglés.
--
BEGIN;

-- Helpers de fechas
CREATE OR REPLACE FUNCTION v3_calc.calculate_campaign_closing_date(date)
RETURNS date
LANGUAGE sql STABLE AS $$
  SELECT CASE
    WHEN $1 IS NULL THEN NULL
    ELSE $1 + (get_campaign_closure_days() || ' days')::INTERVAL
  END::date
$$;

-- Helpers por hectárea / dosis
CREATE OR REPLACE FUNCTION v3_calc.dose_per_ha(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.units_per_ha(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.per_ha($1, $2)
$$;

CREATE OR REPLACE FUNCTION v3_calc.norm_dose(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $2 > 0 THEN $1 / $2 ELSE NULL END
$$;

-- Áreas
CREATE OR REPLACE FUNCTION v3_calc.seeded_area(date, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $1 IS NOT NULL THEN COALESCE($2,0) ELSE 0 END
$$;

CREATE OR REPLACE FUNCTION v3_calc.harvested_area(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT CASE WHEN $1 IS NOT NULL AND $1 > 0 THEN COALESCE($2,0) ELSE 0 END
$$;

-- Rendimientos
CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_over_hectares(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div(COALESCE($1,0), COALESCE($2,0))
$$;

CREATE OR REPLACE FUNCTION v3_calc.yield_tn_per_ha_over_harvested(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.safe_div(COALESCE($1,0), COALESCE($2,0))
$$;

-- Costos
CREATE OR REPLACE FUNCTION v3_calc.labor_cost(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1,0) * COALESCE($2,0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.supply_cost(double precision, numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1,0)::numeric * COALESCE($2,0) * COALESCE($3,0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.cost_per_ha(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.per_ha($1, $2)
$$;

-- Ingresos
CREATE OR REPLACE FUNCTION v3_calc.income_net_total(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE($1,0) * COALESCE($2,0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.income_net_per_ha(numeric, numeric)
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.per_ha($1, $2)
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

CREATE OR REPLACE FUNCTION v3_calc.rent_per_ha(
  lease_type_id bigint,
  lease_type_percent double precision,
  lease_type_value double precision,
  income_net_per_ha double precision,
  cost_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.rent_per_ha(lease_type_id::integer, lease_type_percent, lease_type_value,
                          income_net_per_ha, cost_per_ha, admin_cost_per_ha)
$$;

-- Activo total y resultado operativo
CREATE OR REPLACE FUNCTION v3_calc.active_total_per_ha(
  direct_cost_per_ha double precision,
  rent_per_ha double precision,
  admin_cost_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(direct_cost_per_ha,0) + COALESCE(rent_per_ha,0) + COALESCE(admin_cost_per_ha,0)
$$;

CREATE OR REPLACE FUNCTION v3_calc.operating_result_per_ha(
  income_net_per_ha double precision,
  active_total_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(income_net_per_ha,0) - COALESCE(active_total_per_ha,0)
$$;

-- Precio de indiferencia
CREATE OR REPLACE FUNCTION v3_calc.indifference_price_usd_tn(
  total_invested_per_ha double precision,
  yield_tn_per_ha double precision
) RETURNS double precision
LANGUAGE sql IMMUTABLE AS $$
  SELECT v3_calc.per_ha_dp(total_invested_per_ha, yield_tn_per_ha)
$$;

COMMIT;
