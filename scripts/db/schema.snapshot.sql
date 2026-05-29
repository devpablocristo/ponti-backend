--
-- PostgreSQL database dump
--


-- Dumped from database version 16.11 (Debian 16.11-1.pgdg12+1)
-- Dumped by pg_dump version 16.11 (Debian 16.11-1.pgdg12+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: v4_calc; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA v4_calc;


--
-- Name: v4_core; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA v4_core;


--
-- Name: v4_report; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA v4_report;


--
-- Name: v4_ssot; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA v4_ssot;


--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: unaccent; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS unaccent WITH SCHEMA public;


--
-- Name: EXTENSION unaccent; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION unaccent IS 'text search dictionary that removes accents';


--
-- Name: get_business_parameter(character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_business_parameter(p_key character varying) RETURNS character varying
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value FROM public.business_parameters WHERE key = p_key);
END;
$$;


--
-- Name: get_business_parameter_decimal(character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_business_parameter_decimal(p_key character varying) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value::decimal FROM public.business_parameters WHERE key = p_key);
END;
$$;


--
-- Name: get_business_parameter_integer(character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_business_parameter_integer(p_key character varying) RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value::integer FROM public.business_parameters WHERE key = p_key);
END;
$$;


--
-- Name: get_campaign_closure_days(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_campaign_closure_days() RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN public.get_business_parameter_integer('campaign_closure_days');
END;
$$;


--
-- Name: get_default_fx_rate(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_default_fx_rate() RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN public.get_business_parameter_decimal('default_fx_rate');
END;
$$;


--
-- Name: get_iva_percentage(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_iva_percentage() RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN public.get_business_parameter_decimal('iva_percentage');
END;
$$;


--
-- Name: update_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$;


--
-- Name: active_total_per_ha(numeric, numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.active_total_per_ha(direct_cost_per_ha numeric, rent_per_ha numeric, admin_cost_per_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(direct_cost_per_ha,0) + COALESCE(rent_per_ha,0) + COALESCE(admin_cost_per_ha,0)
$$;


--
-- Name: calculate_campaign_closing_date(date); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.calculate_campaign_closing_date(end_date date) RETURNS date
    LANGUAGE sql STABLE
    AS $$
  SELECT CASE 
    WHEN end_date IS NULL THEN NULL
    ELSE end_date + (get_campaign_closure_days() || ' days')::INTERVAL
  END::date
$$;


--
-- Name: calculate_rent_per_ha(numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.calculate_rent_per_ha(lease_value numeric) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  IF lease_value < 0 THEN
    RETURN 0;
  ELSE
    RETURN lease_value;
  END IF;
END;
$$;


--
-- Name: coalesce0(double precision); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.coalesce0(double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT COALESCE($1, 0)
$_$;


--
-- Name: coalesce0(numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.coalesce0(numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT COALESCE($1, 0)
$_$;


--
-- Name: cost_per_ha(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.cost_per_ha(total_cost numeric, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.per_ha(total_cost, hectares)
$$;


--
-- Name: dollar_average_for_month(bigint, date); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.dollar_average_for_month(p_project_id bigint, p_date date) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT AVG(d.average_value)
     FROM project_dollar_values d
     WHERE d.project_id = p_project_id
       AND TO_CHAR(p_date, 'YYYY-MM') = d.month
       AND d.deleted_at IS NULL),
    1.0
  )
$$;


--
-- Name: dose_per_ha(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.dose_per_ha(total_dose numeric, surface_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.safe_div(total_dose, surface_ha)
$$;


--
-- Name: get_project_dollar_value(bigint, character varying); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.get_project_dollar_value(p_project_id bigint, p_month character varying) RETURNS numeric
    LANGUAGE plpgsql STABLE
    AS $$
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
$$;


--
-- Name: harvested_area(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.harvested_area(tons numeric, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN tons IS NOT NULL AND tons > 0 THEN COALESCE(hectares,0) ELSE 0 END
$$;


--
-- Name: income_net_per_ha(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.income_net_per_ha(income_net_total numeric, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.per_ha(income_net_total, hectares)
$$;


--
-- Name: income_net_total(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.income_net_total(tons numeric, net_price_usd numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(tons,0) * COALESCE(net_price_usd,0)
$$;


--
-- Name: indifference_price_usd_tn(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.indifference_price_usd_tn(total_invested_per_ha numeric, yield_tn_per_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.per_ha(total_invested_per_ha, yield_tn_per_ha)
$$;


--
-- Name: labor_cost(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.labor_cost(labor_price numeric, effective_area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(labor_price,0) * COALESCE(effective_area,0)
$$;


--
-- Name: norm_dose(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.norm_dose(dose numeric, area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN area > 0 THEN dose / area ELSE NULL END
$$;


--
-- Name: operating_result_per_ha(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.operating_result_per_ha(income_net_per_ha numeric, active_total_per_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(income_net_per_ha,0) - COALESCE(active_total_per_ha,0)
$$;


--
-- Name: per_ha(double precision, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.per_ha(double precision, numeric) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v4_core.safe_div_dp($1, $2::double precision)
$_$;


--
-- Name: per_ha(numeric, double precision); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.per_ha(numeric, double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v4_core.safe_div_dp($1::double precision, $2)
$_$;


--
-- Name: per_ha(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.per_ha(value numeric, area_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v4_core.safe_div($1, $2)
$_$;


--
-- Name: per_ha_dp(double precision, double precision); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.per_ha_dp(double precision, double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v4_core.safe_div_dp($1, $2)
$_$;


--
-- Name: percentage(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.percentage(part numeric, total numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v4_core.safe_div($1, $2) * 100
$_$;


--
-- Name: percentage_capped(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.percentage_capped(part numeric, total numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT LEAST(v4_core.safe_div($1, $2) * 100, 100)
$_$;


--
-- Name: percentage_rounded(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.percentage_rounded(part numeric, total numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v4_core.safe_div($1, $2) * 100
$_$;


--
-- Name: rent_per_ha(integer, numeric, numeric, numeric, numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.rent_per_ha(lease_type_id integer, lease_type_percent numeric, lease_type_value numeric, income_net_per_ha numeric, cost_per_ha numeric, admin_cost_per_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
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


--
-- Name: rent_per_ha(bigint, numeric, numeric, numeric, numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.rent_per_ha(lease_type_id bigint, lease_type_percent numeric, lease_type_value numeric, income_net_per_ha numeric, cost_per_ha numeric, admin_cost_per_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.rent_per_ha(
    lease_type_id::integer, 
    lease_type_percent, 
    lease_type_value,
    income_net_per_ha, 
    cost_per_ha, 
    admin_cost_per_ha
  )
$$;


--
-- Name: safe_div(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.safe_div(numerator numeric, denominator numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$_$;


--
-- Name: safe_div_dp(double precision, double precision); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.safe_div_dp(numerator double precision, denominator double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$_$;


--
-- Name: seeded_area(date, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.seeded_area(sowing_date date, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN sowing_date IS NOT NULL THEN COALESCE(hectares,0) ELSE 0 END
$$;


--
-- Name: supply_cost(numeric, numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.supply_cost(final_dose numeric, supply_price numeric, effective_area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(final_dose,0)::numeric * COALESCE(supply_price,0) * COALESCE(effective_area,0)
$$;


--
-- Name: units_per_ha(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.units_per_ha(units numeric, area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.per_ha(units, area)
$$;


--
-- Name: yield_tn_per_ha_over_harvested(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.safe_div(COALESCE(tons,0), COALESCE(harvested_area,0))
$$;


--
-- Name: yield_tn_per_ha_over_hectares(numeric, numeric); Type: FUNCTION; Schema: v4_core; Owner: -
--

CREATE FUNCTION v4_core.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.safe_div(COALESCE(tons,0), COALESCE(hectares,0))
$$;


--
-- Name: active_total_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.active_total_per_ha_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_core.active_total_per_ha(
           v4_ssot.cost_per_ha_for_lot(p_lot_id),
           v4_ssot.rent_per_ha_for_lot(p_lot_id),
           v4_ssot.admin_cost_per_ha_for_lot(p_lot_id)
         )
$$;


--
-- Name: admin_cost_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.admin_cost_per_ha_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(p.admin_cost, 0)::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: admin_cost_prorated_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.admin_cost_prorated_per_ha_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT CASE
           WHEN t.total_hectares > 0 THEN COALESCE(p.admin_cost, 0)::numeric / t.total_hectares
           ELSE 0::numeric
         END
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  CROSS JOIN LATERAL (
    SELECT v4_ssot.total_hectares_for_project(f.project_id)::numeric AS total_hectares
  ) t
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: admin_cost_total_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.admin_cost_total_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT p.admin_cost * v4_ssot.total_hectares_for_project(p_project_id)
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::numeric
$$;


--
-- Name: agrochemicals_invested_for_project_mb(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.agrochemicals_invested_for_project_mb(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 2
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::numeric
$$;


--
-- Name: board_price_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.board_price_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(cc.board_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;


--
-- Name: commercial_cost_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.commercial_cost_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(cc.board_price * (cc.commercial_cost / 100.0), 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;


--
-- Name: cost_per_ha_for_crop(bigint, bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.cost_per_ha_for_crop(p_project_id bigint, p_crop_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_core.per_ha(
    v4_ssot.total_costs_for_crop(p_project_id, p_crop_id),
    (SELECT COALESCE(SUM(l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  )
$$;


--
-- Name: cost_per_ha_for_crop_ssot(bigint, bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.cost_per_ha_for_crop_ssot(p_project_id bigint, p_crop_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_ssot.cost_per_ha_for_crop(p_project_id, p_crop_id)
$$;


--
-- Name: cost_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.cost_per_ha_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_core.safe_div(
           COALESCE(v4_ssot.direct_cost_for_lot(p_lot_id), 0)::numeric,
           v4_ssot.lot_hectares(p_lot_id)
         )
$$;


--
-- Name: crop_incidence_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.crop_incidence_for_project(p_project_id bigint) RETURNS TABLE(current_crop_id bigint, crop_name text, crop_hectares numeric, crop_incidence_pct numeric)
    LANGUAGE sql STABLE
    AS $$
  WITH lot_base AS (
    SELECT
      l.current_crop_id,
      c.name AS crop_name,
      l.hectares
    FROM public.lots l
    JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
    LEFT JOIN public.crops c ON c.id = l.current_crop_id AND c.deleted_at IS NULL
    WHERE f.project_id = p_project_id 
      AND l.deleted_at IS NULL 
      AND l.hectares IS NOT NULL 
      AND l.hectares > 0
  ),
  project_totals AS (
    SELECT SUM(hectares)::numeric AS total_project_hectares
    FROM lot_base
  ),
  by_crop AS (
    SELECT 
      current_crop_id, 
      crop_name, 
      SUM(hectares)::numeric AS crop_hectares
    FROM lot_base
    WHERE current_crop_id IS NOT NULL
    GROUP BY current_crop_id, crop_name
  ),
  crop_percentages AS (
    SELECT
      bc.current_crop_id,
      bc.crop_name,
      bc.crop_hectares,
      pt.total_project_hectares,
      v4_core.percentage_rounded(bc.crop_hectares, pt.total_project_hectares) AS base_percentage,
      ROW_NUMBER() OVER (ORDER BY bc.crop_name) AS crop_order,
      COUNT(*) OVER () AS total_crops
    FROM by_crop bc
    CROSS JOIN project_totals pt
  ),
  project_sums AS (
    SELECT SUM(base_percentage) AS total_percentage
    FROM crop_percentages
  )
  SELECT
    cp.current_crop_id,
    cp.crop_name,
    cp.crop_hectares,
    CASE 
      WHEN ps.total_percentage > 99.000 AND cp.crop_order = cp.total_crops THEN
        100.000 - COALESCE((
          SELECT SUM(base_percentage) 
          FROM crop_percentages cp2 
          WHERE cp2.crop_order < cp.crop_order
        ), 0)
      ELSE
        cp.base_percentage
    END AS crop_incidence_pct
  FROM crop_percentages cp
  CROSS JOIN project_sums ps
  ORDER BY cp.crop_name
$$;


--
-- Name: direct_cost_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.direct_cost_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(v4_ssot.labor_cost_for_lot(p_lot_id), 0)::numeric
       + COALESCE(v4_ssot.supply_cost_for_lot(p_lot_id), 0)
$$;


--
-- Name: direct_cost_per_ha_usd(numeric, numeric, numeric); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.direct_cost_per_ha_usd(p_labor_cost_usd numeric, p_supply_cost_usd numeric, p_sowed_area_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v4_core.safe_div(
    COALESCE(p_labor_cost_usd, 0) + COALESCE(p_supply_cost_usd, 0),
    COALESCE(p_sowed_area_ha, 0)
  )
$$;


--
-- Name: direct_cost_usd(numeric, numeric); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.direct_cost_usd(p_labor_cost_usd numeric, p_supply_cost_usd numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(p_labor_cost_usd, 0) + COALESCE(p_supply_cost_usd, 0)
$$;


--
-- Name: direct_costs_invested_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.direct_costs_invested_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(lb.price * l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     JOIN public.labors lb ON lb.project_id = f.project_id AND lb.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::numeric
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    +
    v4_ssot.supply_cost_received_for_project(p_project_id)
  , 0)::numeric
$$;


--
-- Name: direct_costs_invested_for_project_mb(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.direct_costs_invested_for_project_mb(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::numeric
$$;


--
-- Name: direct_costs_total_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.direct_costs_total_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.direct_cost_for_lot(l.id)), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id
       AND l.deleted_at IS NULL)
  , 0)::numeric
$$;


--
-- Name: first_workorder_date_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.first_workorder_date_for_project(p_project_id bigint) RETURNS date
    LANGUAGE sql STABLE
    AS $$
  SELECT MIN(date)
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;


--
-- Name: first_workorder_id_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.first_workorder_id_for_project(p_project_id bigint) RETURNS bigint
    LANGUAGE sql STABLE
    AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;


--
-- Name: first_workorder_number_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.first_workorder_number_for_project(p_project_id bigint) RETURNS text
    LANGUAGE sql STABLE
    AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date ASC, id ASC
  LIMIT 1
$$;


--
-- Name: freight_cost_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.freight_cost_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(cc.freight_cost, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc 
    ON cc.project_id = f.project_id 
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id 
    AND l.deleted_at IS NULL
  LIMIT 1
$$;


--
-- Name: harvested_area_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.harvested_area_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_core.harvested_area(
           v4_ssot.lot_tons(p_lot_id)::numeric,
           v4_ssot.lot_hectares(p_lot_id)::numeric
         )
$$;


--
-- Name: income_net_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.income_net_per_ha_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_core.safe_div(
           COALESCE(v4_ssot.income_net_total_for_lot(p_lot_id), 0),
           v4_ssot.lot_hectares(p_lot_id)::numeric
         )
$$;


--
-- Name: income_net_total_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.income_net_total_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(l.tons, 0)::numeric * COALESCE(v4_ssot.net_price_usd_for_lot(l.id), 0)::numeric
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: kilograms_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.kilograms_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    SUM(CASE WHEN s.unit_id = 2 THEN (wi.final_dose * w.effective_area) ELSE 0 END), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;


--
-- Name: labor_cost_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.labor_cost_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
$$;


--
-- Name: labor_cost_pre_harvest_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.labor_cost_pre_harvest_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(SUM(lb.price * w.effective_area), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  JOIN public.categories cat ON cat.id = lb.category_id
  WHERE w.deleted_at IS NULL
    AND w.effective_area > 0
    AND lb.price IS NOT NULL
    AND w.lot_id = p_lot_id
    AND cat.type_id = 4
    AND cat.name != 'Cosecha'  -- EXCLUIR COSECHA
$$;


--
-- Name: last_stock_count_date_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.last_stock_count_date_for_project(p_project_id bigint) RETURNS date
    LANGUAGE sql STABLE
    AS $$
  SELECT MAX(close_date)
  FROM public.stocks
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;


--
-- Name: last_workorder_date_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.last_workorder_date_for_project(p_project_id bigint) RETURNS date
    LANGUAGE sql STABLE
    AS $$
  SELECT MAX(date)
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
$$;


--
-- Name: last_workorder_id_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.last_workorder_id_for_project(p_project_id bigint) RETURNS bigint
    LANGUAGE sql STABLE
    AS $$
  SELECT id
  FROM public.workorders
  WHERE project_id = p_project_id 
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;


--
-- Name: last_workorder_number_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.last_workorder_number_for_project(p_project_id bigint) RETURNS text
    LANGUAGE sql STABLE
    AS $$
  SELECT number::text
  FROM public.workorders
  WHERE project_id = p_project_id
    AND deleted_at IS NULL
  ORDER BY date DESC, id DESC
  LIMIT 1
$$;


--
-- Name: lease_executed_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.lease_executed_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(SUM(
    CASE
      WHEN f.lease_type_id IN (3, 4) THEN f.lease_type_value * l.hectares
      ELSE 0
    END
  ), 0)::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
$$;


--
-- Name: lease_invested_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.lease_invested_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares), 0)::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
$$;


--
-- Name: liters_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.liters_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    SUM(CASE WHEN s.unit_id = 1 THEN (wi.final_dose * w.effective_area) ELSE 0 END), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;


--
-- Name: lot_hectares(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.lot_hectares(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(l.hectares, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: lot_tons(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.lot_tons(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(l.tons, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: net_price_usd_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.net_price_usd_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(cc.net_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc
    ON cc.project_id = f.project_id
   AND cc.crop_id = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id
    AND l.deleted_at IS NULL
  LIMIT 1
$$;


--
-- Name: operating_result_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.operating_result_per_ha_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_core.operating_result_per_ha(
           v4_ssot.income_net_per_ha_for_lot(p_lot_id),
           v4_ssot.active_total_per_ha_for_lot(p_lot_id)
         )
$$;


--
-- Name: operating_result_total_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.operating_result_total_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  WITH project_totals AS (
    SELECT
      p.id,
      p.admin_cost,
      COALESCE(SUM(l.hectares), 0)::numeric AS total_hectares
    FROM public.projects p
    LEFT JOIN public.fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.id = p_project_id AND p.deleted_at IS NULL
    GROUP BY p.id, p.admin_cost
  ),
  lease_cost AS (
    SELECT
      COALESCE(
        SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares),
        0
      )::numeric AS total_lease
    FROM public.lots l
    JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
    WHERE f.project_id = p_project_id AND l.deleted_at IS NULL
  )
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.income_net_total_for_lot(l.id)), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    v4_ssot.direct_costs_total_for_project(p_project_id)
    -
    (SELECT total_lease FROM lease_cost)
    -
    (SELECT COALESCE(admin_cost * total_hectares, 0)::numeric FROM project_totals)
  , 0)::numeric
$$;


--
-- Name: rent_fixed_only_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.rent_fixed_only_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT
    CASE
      WHEN f.lease_type_id = 1 THEN 0
      
      WHEN f.lease_type_id = 2 THEN 0
      
      WHEN f.lease_type_id = 3 THEN COALESCE(f.lease_type_value, 0)
      
      WHEN f.lease_type_id = 4 THEN COALESCE(f.lease_type_value, 0)
      
      ELSE 0
    END::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: rent_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.rent_per_ha_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_core.rent_per_ha(
           f.lease_type_id,
           f.lease_type_percent,
           f.lease_type_value,
           v4_ssot.income_net_per_ha_for_lot(p_lot_id),
           v4_ssot.cost_per_ha_for_lot(p_lot_id),
           v4_ssot.admin_cost_per_ha_for_lot(p_lot_id)
         )::numeric
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: rent_per_ha_for_lot_fixed(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.rent_per_ha_for_lot_fixed(p_lot_id bigint) RETURNS numeric
    LANGUAGE plpgsql STABLE
    AS $$
DECLARE
  calculated_rent numeric;
BEGIN
  calculated_rent := v4_ssot.rent_per_ha_for_lot(p_lot_id);
  
  IF calculated_rent < 0 THEN
    RETURN 0;
  ELSE
    RETURN calculated_rent;
  END IF;
END;
$$;


--
-- Name: renta_pct(numeric, numeric); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.renta_pct(operating_result_total_usd numeric, total_costs_usd numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN COALESCE(total_costs_usd,0) > 0
              THEN (COALESCE(operating_result_total_usd,0) / total_costs_usd) * 100
              ELSE 0 END
$$;


--
-- Name: seeded_area_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.seeded_area_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(SUM(
    CASE WHEN lb.category_id = 9 THEN w.effective_area ELSE 0 END
  ), 0)::numeric
  FROM public.workorders w
  JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
$$;


--
-- Name: seeds_invested_for_project_mb(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.seeds_invested_for_project_mb(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.categories c ON c.id = s.category_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND c.type_id = 1
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::numeric
$$;


--
-- Name: stock_value_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.stock_value_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::numeric
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id 
       AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    -
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
     WHERE w.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::numeric
$$;


--
-- Name: stock_value_for_project_mb(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.stock_value_for_project_mb(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type IN ('Stock', 'Remito oficial'))
  , 0)::numeric - v4_ssot.direct_costs_total_for_project(p_project_id)
$$;


--
-- Name: supply_cost_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.supply_cost_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM((wi.final_dose)::numeric * s.price * (w.effective_area)::numeric), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.lot_id = p_lot_id)
    +
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     JOIN public.workorders w ON w.lot_id = p_lot_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND w.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = w.project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::numeric
$$;


--
-- Name: supply_cost_for_lot_base(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.supply_cost_for_lot_base(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    SUM(wi.total_used * s.price), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area IS NOT NULL
    AND w.effective_area > 0
$$;


--
-- Name: supply_cost_for_lot_by_category(bigint, text); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.supply_cost_for_lot_by_category(p_lot_id bigint, p_category_name text) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    SUM(wi.total_used * s.price), 0
  )::numeric
  FROM public.workorders w
  JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  JOIN public.categories c ON c.id = s.category_id
  WHERE w.lot_id = p_lot_id
    AND c.name = p_category_name
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
    AND wi.total_used > 0
    AND s.price IS NOT NULL
$$;


--
-- Name: supply_cost_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.supply_cost_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM((wi.final_dose)::numeric * s.price * (w.effective_area)::numeric), 0)::numeric
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.project_id = p_project_id)
    +
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::numeric
$$;


--
-- Name: supply_cost_received_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.supply_cost_received_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::numeric
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno entrada'
       AND sm.is_entry = true
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::numeric
$$;


--
-- Name: supply_movements_invested_total_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.supply_movements_invested_total_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT SUM(sm.quantity * s.price)
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.project_id = p_project_id 
       AND sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.is_entry = TRUE
       AND sm.movement_type IN ('Stock', 'Remito oficial', 'Movimiento interno', 'Movimiento interno entrada'))
  , 0)::numeric
$$;


--
-- Name: surface_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.surface_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(SUM(w.effective_area), 0)::numeric
  FROM public.workorders w
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
    AND w.effective_area > 0
$$;


--
-- Name: total_budget_cost_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.total_budget_cost_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(p.admin_cost * 10, 0)::numeric
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;


--
-- Name: total_costs_for_crop(bigint, bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.total_costs_for_crop(p_project_id bigint, p_crop_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.direct_cost_for_lot(l.id)), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id
       AND l.current_crop_id = p_crop_id
       AND l.deleted_at IS NULL)
  , 0)::numeric
$$;


--
-- Name: total_costs_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.total_costs_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    v4_ssot.direct_costs_total_for_project(p_project_id) + 
    v4_ssot.lease_invested_for_project(p_project_id) + 
    v4_ssot.admin_cost_total_for_project(p_project_id)
  , 0)::numeric
$$;


--
-- Name: total_hectares_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.total_hectares_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(SUM(l.hectares), 0)::numeric
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
$$;


--
-- Name: total_invested_cost_for_project(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.total_invested_cost_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT COALESCE(SUM(v4_ssot.direct_cost_for_lot(l.id)), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(v4_ssot.rent_per_ha_for_lot(l.id) * l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    (SELECT COALESCE(SUM(v4_ssot.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::numeric
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::numeric
$$;


--
-- Name: yield_tn_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v4_ssot; Owner: -
--

CREATE FUNCTION v4_ssot.yield_tn_per_ha_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v4_core.per_ha(
           v4_ssot.lot_tons(p_lot_id),
           v4_ssot.lot_hectares(p_lot_id)
         )
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_cost_investors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.admin_cost_investors (
    project_id bigint NOT NULL,
    investor_id bigint NOT NULL,
    percentage integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: business_parameters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business_parameters (
    id integer NOT NULL,
    key character varying(100) NOT NULL,
    value character varying(255) NOT NULL,
    type character varying(20) NOT NULL,
    category character varying(50) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint,
    deleted_at timestamp with time zone
);


--
-- Name: business_parameters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.business_parameters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: business_parameters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.business_parameters_id_seq OWNED BY public.business_parameters.id;


--
-- Name: campaigns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.campaigns (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: campaigns_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.campaigns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: campaigns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.campaigns_id_seq OWNED BY public.campaigns.id;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    id bigint NOT NULL,
    name character varying(250) NOT NULL,
    type_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.categories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;


--
-- Name: crop_commercializations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.crop_commercializations (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    crop_id bigint NOT NULL,
    board_price numeric(12,2) NOT NULL,
    freight_cost numeric(12,2) NOT NULL,
    commercial_cost numeric(18,6) NOT NULL,
    net_price numeric(12,2) NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    deleted_at timestamp without time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: crop_commercializations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.crop_commercializations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: crop_commercializations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.crop_commercializations_id_seq OWNED BY public.crop_commercializations.id;


--
-- Name: crops; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.crops (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: crops_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.crops_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: crops_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.crops_id_seq OWNED BY public.crops.id;


--
-- Name: customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customers (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: customers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.customers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.customers_id_seq OWNED BY public.customers.id;


--
-- Name: field_investors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.field_investors (
    field_id bigint NOT NULL,
    investor_id bigint NOT NULL,
    percentage integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fields (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    project_id bigint NOT NULL,
    lease_type_id bigint NOT NULL,
    lease_type_percent numeric(18,6),
    lease_type_value numeric(18,6),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: fields_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fields_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fields_id_seq OWNED BY public.fields.id;


--
-- Name: fx_rates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fx_rates (
    id integer NOT NULL,
    currency_pair character varying(10) NOT NULL,
    rate numeric(10,4) NOT NULL,
    effective_date date NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: fx_rates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fx_rates_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fx_rates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fx_rates_id_seq OWNED BY public.fx_rates.id;


--
-- Name: investors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.investors (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: investors_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.investors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: investors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.investors_id_seq OWNED BY public.investors.id;


--
-- Name: invoices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoices (
    id bigint NOT NULL,
    work_order_id bigint NOT NULL,
    number character varying NOT NULL,
    company character varying(100) NOT NULL,
    date timestamp without time zone NOT NULL,
    status character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: invoices_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoices_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoices_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.invoices_id_seq OWNED BY public.invoices.id;


--
-- Name: labor_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.labor_categories (
    id integer NOT NULL,
    name text NOT NULL,
    type_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by character varying(255),
    updated_by character varying(255),
    deleted_by character varying(255)
);


--
-- Name: labor_categories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.labor_categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: labor_categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.labor_categories_id_seq OWNED BY public.labor_categories.id;


--
-- Name: labor_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.labor_types (
    id integer NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by character varying(255),
    updated_by character varying(255),
    deleted_by character varying(255)
);


--
-- Name: labor_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.labor_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: labor_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.labor_types_id_seq OWNED BY public.labor_types.id;


--
-- Name: labors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.labors (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    name text NOT NULL,
    category_id integer NOT NULL,
    price numeric(12,2) NOT NULL,
    contractor_name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: labors_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.labors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: labors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.labors_id_seq OWNED BY public.labors.id;


--
-- Name: lease_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lease_types (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: lease_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lease_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lease_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.lease_types_id_seq OWNED BY public.lease_types.id;


--
-- Name: lot_dates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lot_dates (
    id bigint NOT NULL,
    lot_id bigint NOT NULL,
    sowing_date date,
    harvest_date date,
    sequence integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: lot_dates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lot_dates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lot_dates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.lot_dates_id_seq OWNED BY public.lot_dates.id;


--
-- Name: lots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lots (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    field_id bigint NOT NULL,
    hectares numeric(18,6) NOT NULL,
    previous_crop_id bigint NOT NULL,
    current_crop_id bigint NOT NULL,
    season character varying(20) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint,
    variety text,
    sowing_date date,
    tons numeric
);


--
-- Name: lots_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lots_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lots_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.lots_id_seq OWNED BY public.lots.id;


--
-- Name: managers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.managers (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: managers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.managers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: managers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.managers_id_seq OWNED BY public.managers.id;


--
-- Name: project_dollar_values; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_dollar_values (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    year bigint NOT NULL,
    month character varying(20) NOT NULL,
    start_value numeric(12,2) NOT NULL,
    end_value numeric(12,2) NOT NULL,
    average_value numeric(12,2) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: project_dollar_values_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.project_dollar_values_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: project_dollar_values_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.project_dollar_values_id_seq OWNED BY public.project_dollar_values.id;


--
-- Name: project_investors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_investors (
    project_id bigint NOT NULL,
    investor_id bigint NOT NULL,
    percentage integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: project_managers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project_managers (
    project_id bigint NOT NULL,
    manager_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    customer_id bigint NOT NULL,
    campaign_id bigint NOT NULL,
    admin_cost numeric(15,3) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint,
    planned_cost numeric(12,2)
);


--
-- Name: COLUMN projects.admin_cost; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.projects.admin_cost IS 'Costo administrativo del proyecto en USD con 3 decimales de precisión';


--
-- Name: projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;


--
-- Name: providers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.providers (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by character varying(255),
    updated_by character varying(255),
    deleted_by character varying(255)
);


--
-- Name: providers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.providers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: providers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.providers_id_seq OWNED BY public.providers.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: stocks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stocks (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    supply_id bigint NOT NULL,
    investor_id bigint NOT NULL,
    close_date date,
    real_stock_units numeric(15,3) NOT NULL,
    initial_units numeric(15,3) NOT NULL,
    year_period integer NOT NULL,
    month_period integer NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint,
    units_entered numeric(15,3) DEFAULT 0 NOT NULL,
    units_consumed numeric(15,3) DEFAULT 0 NOT NULL
);


--
-- Name: stocks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.stocks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: stocks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.stocks_id_seq OWNED BY public.stocks.id;


--
-- Name: supplies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.supplies (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    name character varying(100) NOT NULL,
    price numeric(18,6) NOT NULL,
    unit_id integer,
    category_id integer,
    type_id integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: supplies_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.supplies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: supplies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.supplies_id_seq OWNED BY public.supplies.id;


--
-- Name: supply_movements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.supply_movements (
    id bigint NOT NULL,
    stock_id bigint NOT NULL,
    quantity numeric(15,3) NOT NULL,
    movement_type text NOT NULL,
    movement_date timestamp without time zone NOT NULL,
    reference_number text NOT NULL,
    is_entry boolean NOT NULL,
    project_id bigint NOT NULL,
    project_destination_id bigint NOT NULL,
    supply_id bigint NOT NULL,
    investor_id bigint NOT NULL,
    provider_id bigint NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint,
    CONSTRAINT chk_supply_movements_movement_type CHECK ((movement_type = ANY (ARRAY['Stock'::text, 'Movimiento interno'::text, 'Remito oficial'::text, 'Movimiento interno entrada'::text])))
);


--
-- Name: supply_movements_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.supply_movements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: supply_movements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.supply_movements_id_seq OWNED BY public.supply_movements.id;


--
-- Name: types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.types (
    id bigint NOT NULL,
    name character varying(250) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.types_id_seq OWNED BY public.types.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    token_hash text NOT NULL,
    refresh_tokens text[],
    id_rol bigint,
    is_verified boolean,
    active boolean,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp without time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: workorder_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workorder_items (
    id bigint NOT NULL,
    workorder_id bigint NOT NULL,
    supply_id bigint NOT NULL,
    total_used numeric(18,6) NOT NULL,
    final_dose numeric(18,6) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: workorder_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.workorder_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workorder_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workorder_items_id_seq OWNED BY public.workorder_items.id;


--
-- Name: workorders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.workorders (
    id bigint NOT NULL,
    number character varying(100),
    project_id bigint NOT NULL,
    field_id bigint NOT NULL,
    lot_id bigint NOT NULL,
    crop_id bigint NOT NULL,
    labor_id bigint NOT NULL,
    contractor character varying(100),
    observations text,
    date date NOT NULL,
    investor_id bigint NOT NULL,
    effective_area numeric(18,6) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: workorders_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.workorders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: workorders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.workorders_id_seq OWNED BY public.workorders.id;


--
-- Name: dashboard_fertilizers_invested_by_project; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.dashboard_fertilizers_invested_by_project AS
 SELECT sm.project_id,
    COALESCE(sum((sm.quantity * s.price)), (0)::numeric) AS fertilizantes_invertidos_usd
   FROM ((public.supply_movements sm
     JOIN public.supplies s ON ((s.id = sm.supply_id)))
     JOIN public.categories c ON ((s.category_id = c.id)))
  WHERE ((sm.deleted_at IS NULL) AND (s.deleted_at IS NULL) AND (sm.is_entry = true) AND (sm.movement_type = ANY (ARRAY['Stock'::text, 'Remito oficial'::text, 'Movimiento interno'::text, 'Movimiento interno entrada'::text])) AND (c.type_id = 3))
  GROUP BY sm.project_id;


--
-- Name: dashboard_supply_costs_by_project; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.dashboard_supply_costs_by_project AS
 SELECT p.id AS project_id,
    COALESCE(sum(v4_ssot.supply_cost_for_lot_by_category(l.id, 'Semilla'::text)), (0)::numeric) AS semillas_ejecutados_usd,
    COALESCE(sum((((((v4_ssot.supply_cost_for_lot_by_category(l.id, 'Coadyuvantes'::text) + v4_ssot.supply_cost_for_lot_by_category(l.id, 'Curasemillas'::text)) + v4_ssot.supply_cost_for_lot_by_category(l.id, 'Herbicidas'::text)) + v4_ssot.supply_cost_for_lot_by_category(l.id, 'Insecticidas'::text)) + v4_ssot.supply_cost_for_lot_by_category(l.id, 'Fungicidas'::text)) + v4_ssot.supply_cost_for_lot_by_category(l.id, 'Otros Insumos'::text))), (0)::numeric) AS agroquimicos_ejecutados_usd,
    COALESCE(sum(v4_ssot.supply_cost_for_lot_by_category(l.id, 'Fertilizantes'::text)), (0)::numeric) AS fertilizantes_ejecutados_usd
   FROM ((public.projects p
     LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.id;


--
-- Name: field_crop_lot_base; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.field_crop_lot_base AS
 SELECT f.project_id,
    f.id AS field_id,
    l.current_crop_id AS crop_id,
    l.id AS lot_id,
    l.hectares AS surface_ha,
    v4_ssot.seeded_area_for_lot(l.id) AS seeded_area_ha,
    l.tons,
    v4_ssot.seeded_area_for_lot(l.id) AS sowed_area_ha,
    v4_ssot.seeded_area_for_lot(l.id) AS sown_area_ha
   FROM (public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
  WHERE ((l.deleted_at IS NULL) AND (l.current_crop_id IS NOT NULL));


--
-- Name: field_crop_supply_costs_by_lot; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.field_crop_supply_costs_by_lot AS
 SELECT project_id,
    field_id,
    crop_id,
    lot_id,
    surface_ha,
    seeded_area_ha,
    tons,
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla'::text) AS semillas_usd,
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas'::text) AS curasemillas_usd,
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas'::text) AS herbicidas_usd,
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas'::text) AS insecticidas_usd,
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas'::text) AS fungicidas_usd,
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes'::text) AS coadyuvantes_usd,
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes'::text) AS fertilizantes_usd,
    v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos'::text) AS otros_insumos_usd,
    (((((((v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Semilla'::text) + v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Curasemillas'::text)) + v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Herbicidas'::text)) + v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Insecticidas'::text)) + v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Fungicidas'::text)) + v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Coadyuvantes'::text)) + v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Fertilizantes'::text)) + v4_ssot.supply_cost_for_lot_by_category(lot_id, 'Otros Insumos'::text)) AS total_insumos_usd,
    seeded_area_ha AS sowed_area_ha,
    seeded_area_ha AS sown_area_ha
   FROM v4_calc.field_crop_lot_base;


--
-- Name: field_crop_aggregated; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.field_crop_aggregated AS
 SELECT project_id,
    field_id,
    crop_id,
    min(lot_id) AS sample_lot_id,
    sum(tons) AS production_tn,
    sum(seeded_area_ha) AS seeded_area_ha,
    sum(surface_ha) AS surface_ha,
    sum(v4_ssot.labor_cost_for_lot(lot_id)) AS labor_costs_usd,
    sum(total_insumos_usd) AS supply_costs_usd,
    sum((v4_ssot.labor_cost_for_lot(lot_id) + total_insumos_usd)) AS direct_cost_usd,
    sum((v4_ssot.rent_fixed_only_for_lot(lot_id) * surface_ha)) AS rent_fixed_usd,
    sum((v4_ssot.rent_per_ha_for_lot(lot_id) * surface_ha)) AS rent_total_usd,
    sum((v4_ssot.admin_cost_prorated_per_ha_for_lot(lot_id) * surface_ha)) AS administration_usd,
    sum(seeded_area_ha) AS sowed_area_ha,
    sum(seeded_area_ha) AS sown_area_ha
   FROM v4_calc.field_crop_supply_costs_by_lot
  GROUP BY project_id, field_id, crop_id;


--
-- Name: field_crop_labor_costs_by_lot; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.field_crop_labor_costs_by_lot AS
 SELECT project_id,
    field_id,
    crop_id,
    lot_id,
    seeded_area_ha,
    surface_ha,
    COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text = 'Siembra'::text) AND (cat.type_id = 4))), (0)::numeric) AS siembra_usd,
    COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text = 'Pulverización'::text) AND (cat.type_id = 4))), (0)::numeric) AS pulverizacion_usd,
    COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text = 'Riego'::text) AND (cat.type_id = 4))), (0)::numeric) AS riego_usd,
    COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text = 'Cosecha'::text) AND (cat.type_id = 4))), (0)::numeric) AS cosecha_usd,
    COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text <> ALL ((ARRAY['Siembra'::character varying, 'Pulverización'::character varying, 'Riego'::character varying, 'Cosecha'::character varying])::text[])) AND (cat.type_id = 4))), (0)::numeric) AS otras_labores_usd,
    ((((COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text = 'Siembra'::text) AND (cat.type_id = 4))), (0)::numeric) + COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text = 'Pulverización'::text) AND (cat.type_id = 4))), (0)::numeric)) + COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text = 'Riego'::text) AND (cat.type_id = 4))), (0)::numeric)) + COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text = 'Cosecha'::text) AND (cat.type_id = 4))), (0)::numeric)) + COALESCE(( SELECT sum((lab.price * w.effective_area)) AS sum
           FROM ((public.workorders w
             JOIN public.labors lab ON ((lab.id = w.labor_id)))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lab.price IS NOT NULL) AND ((cat.name)::text <> ALL ((ARRAY['Siembra'::character varying, 'Pulverización'::character varying, 'Riego'::character varying, 'Cosecha'::character varying])::text[])) AND (cat.type_id = 4))), (0)::numeric)) AS total_labores_usd,
    seeded_area_ha AS sowed_area_ha,
    seeded_area_ha AS sown_area_ha
   FROM v4_calc.field_crop_lot_base lb;


--
-- Name: field_crop_metrics_lot_base; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.field_crop_metrics_lot_base AS
 SELECT f.project_id,
    f.id AS field_id,
    f.name AS field_name,
    l.current_crop_id,
    c.name AS crop_name,
    l.id AS lot_id,
    l.hectares,
    l.tons,
    COALESCE(v4_ssot.seeded_area_for_lot(l.id), (0)::numeric) AS seeded_area_ha,
    COALESCE(v4_ssot.harvested_area_for_lot(l.id), (0)::numeric) AS harvested_area_ha,
    COALESCE(v4_ssot.yield_tn_per_ha_for_lot(l.id), (0)::numeric) AS yield_tn_per_ha,
    COALESCE(v4_ssot.labor_cost_for_lot(l.id), (0)::numeric) AS labor_cost_usd,
    COALESCE(v4_ssot.supply_cost_for_lot_base(l.id), (0)::numeric) AS supply_cost_usd,
    COALESCE(v4_ssot.net_price_usd_for_lot(l.id), (0)::numeric) AS net_price_usd,
    COALESCE(v4_ssot.rent_per_ha_for_lot(l.id), (0)::numeric) AS rent_per_ha,
    COALESCE(v4_ssot.admin_cost_prorated_per_ha_for_lot(l.id), (0)::numeric) AS admin_per_ha,
    COALESCE(v4_ssot.board_price_for_lot(l.id), (0)::numeric) AS board_price,
    COALESCE(v4_ssot.freight_cost_for_lot(l.id), (0)::numeric) AS freight_cost,
    COALESCE(v4_ssot.commercial_cost_for_lot(l.id), (0)::numeric) AS commercial_cost,
    COALESCE(v4_ssot.seeded_area_for_lot(l.id), (0)::numeric) AS sowed_area_ha,
    COALESCE(v4_ssot.seeded_area_for_lot(l.id), (0)::numeric) AS sown_area_ha
   FROM (((public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
     JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
     LEFT JOIN public.crops c ON (((c.id = l.current_crop_id) AND (c.deleted_at IS NULL))))
  WHERE ((l.deleted_at IS NULL) AND (l.current_crop_id IS NOT NULL) AND (l.hectares > (0)::numeric));


--
-- Name: field_crop_metrics_aggregated; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.field_crop_metrics_aggregated AS
 SELECT project_id,
    field_id,
    field_name,
    current_crop_id,
    crop_name,
    sum(hectares) AS superficie_total,
    sum(seeded_area_ha) AS superficie_sembrada_ha,
    sum(harvested_area_ha) AS area_cosechada_ha,
    sum(tons) AS produccion_tn,
        CASE
            WHEN (sum(seeded_area_ha) > (0)::numeric) THEN (sum((yield_tn_per_ha * seeded_area_ha)) / sum(seeded_area_ha))
            ELSE (0)::numeric
        END AS rendimiento_tn_ha,
        CASE
            WHEN (sum(tons) > (0)::numeric) THEN (sum((board_price * tons)) / sum(tons))
            ELSE (0)::numeric
        END AS precio_bruto_usd_tn,
        CASE
            WHEN (sum(tons) > (0)::numeric) THEN (sum((freight_cost * tons)) / sum(tons))
            ELSE (0)::numeric
        END AS gasto_flete_usd_tn,
        CASE
            WHEN (sum(tons) > (0)::numeric) THEN (sum((commercial_cost * tons)) / sum(tons))
            ELSE (0)::numeric
        END AS gasto_comercial_usd_tn,
        CASE
            WHEN (sum(tons) > (0)::numeric) THEN (sum((net_price_usd * tons)) / sum(tons))
            ELSE (0)::numeric
        END AS precio_neto_usd_tn,
    sum((tons * net_price_usd)) AS ingreso_neto_total
   FROM v4_calc.field_crop_metrics_lot_base lb
  GROUP BY project_id, field_id, field_name, current_crop_id, crop_name;


--
-- Name: investor_contribution_categories; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.investor_contribution_categories AS
 WITH lot_base AS (
         SELECT f.project_id,
            l.id AS lot_id,
            l.hectares,
            COALESCE(( SELECT sum(w.effective_area) AS sum
                   FROM ((public.workorders w
                     JOIN public.labors lab ON ((w.labor_id = lab.id)))
                     JOIN public.categories cat ON ((lab.category_id = cat.id)))
                  WHERE ((w.lot_id = l.id) AND (w.deleted_at IS NULL) AND ((cat.name)::text = 'Siembra'::text) AND (cat.type_id = 4))), (0)::numeric) AS seeded_area_ha
           FROM ((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
          WHERE (l.deleted_at IS NULL)
        ), seed_area AS (
         SELECT lb.project_id,
            COALESCE(sum(lb.seeded_area_ha), (0)::numeric) AS total_seeded_area_ha,
            COALESCE(sum((v4_ssot.rent_fixed_only_for_lot(lb.lot_id) * lb.hectares)), (0)::numeric) AS rent_capitalizable_total_usd,
            COALESCE(sum((v4_ssot.admin_cost_prorated_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::numeric) AS administration_total_usd
           FROM lot_base lb
          GROUP BY lb.project_id
        ), labor_totals AS (
         SELECT lb.project_id,
            COALESCE(sum(
                CASE
                    WHEN ((cat.name)::text = ANY ((ARRAY['Pulverización'::character varying, 'Otras Labores'::character varying])::text[])) THEN (lab.price * w.effective_area)
                    ELSE (0)::numeric
                END), (0)::numeric) AS general_labors_total_usd,
            COALESCE(sum(
                CASE
                    WHEN ((cat.name)::text = 'Siembra'::text) THEN (lab.price * w.effective_area)
                    ELSE (0)::numeric
                END), (0)::numeric) AS sowing_total_usd,
            COALESCE(sum(
                CASE
                    WHEN ((cat.name)::text = 'Riego'::text) THEN (lab.price * w.effective_area)
                    ELSE (0)::numeric
                END), (0)::numeric) AS irrigation_total_usd
           FROM (((lot_base lb
             JOIN public.workorders w ON (((w.lot_id = lb.lot_id) AND (w.deleted_at IS NULL))))
             JOIN public.labors lab ON (((lab.id = w.labor_id) AND (lab.deleted_at IS NULL))))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE (cat.type_id = 4)
          GROUP BY lb.project_id
        ), invested_totals AS (
         SELECT p.project_id,
            v4_ssot.seeds_invested_for_project_mb(p.project_id) AS seeds_total_usd,
            v4_ssot.agrochemicals_invested_for_project_mb(p.project_id) AS agrochemicals_total_usd,
            ( SELECT COALESCE(sum((sm.quantity * s.price)), (0)::numeric) AS "coalesce"
                   FROM ((public.supply_movements sm
                     JOIN public.supplies s ON (((s.id = sm.supply_id) AND (s.deleted_at IS NULL))))
                     JOIN public.categories cat ON ((cat.id = s.category_id)))
                  WHERE ((sm.project_id = p.project_id) AND (sm.deleted_at IS NULL) AND (sm.is_entry = true) AND (sm.movement_type = ANY (ARRAY['Stock'::text, 'Remito oficial'::text, 'Movimiento interno'::text, 'Movimiento interno entrada'::text])) AND (cat.type_id = 3))) AS fertilizers_total_usd
           FROM ( SELECT DISTINCT lot_base.project_id
                   FROM lot_base) p
        )
 SELECT it.project_id,
    it.agrochemicals_total_usd,
    COALESCE(it.fertilizers_total_usd, (0)::numeric) AS fertilizers_total_usd,
    it.seeds_total_usd,
    COALESCE(lt.general_labors_total_usd, (0)::numeric) AS general_labors_total_usd,
    COALESCE(lt.sowing_total_usd, (0)::numeric) AS sowing_total_usd,
    COALESCE(lt.irrigation_total_usd, (0)::numeric) AS irrigation_total_usd,
    COALESCE(sa.rent_capitalizable_total_usd, (0)::numeric) AS rent_capitalizable_total_usd,
    COALESCE(sa.administration_total_usd, (0)::numeric) AS administration_total_usd,
    COALESCE(sa.total_seeded_area_ha, (0)::numeric) AS total_seeded_area_ha
   FROM ((invested_totals it
     LEFT JOIN labor_totals lt ON ((lt.project_id = it.project_id)))
     LEFT JOIN seed_area sa ON ((sa.project_id = it.project_id)));


--
-- Name: investor_real_contributions; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.investor_real_contributions AS
 WITH investor_base AS (
         SELECT pi.project_id,
            pi.investor_id,
            i.name AS investor_name,
            pi.percentage AS share_pct_agreed
           FROM (public.project_investors pi
             JOIN public.investors i ON (((i.id = pi.investor_id) AND (i.deleted_at IS NULL))))
          WHERE (pi.deleted_at IS NULL)
        ), admin_config AS (
         SELECT admin_cost_investors.project_id,
            (count(*) FILTER (WHERE (admin_cost_investors.deleted_at IS NULL)) > 0) AS has_custom_admin
           FROM public.admin_cost_investors
          GROUP BY admin_cost_investors.project_id
        ), category_totals AS (
         SELECT investor_contribution_categories.project_id,
            investor_contribution_categories.agrochemicals_total_usd,
            investor_contribution_categories.fertilizers_total_usd,
            investor_contribution_categories.seeds_total_usd,
            investor_contribution_categories.general_labors_total_usd,
            investor_contribution_categories.sowing_total_usd,
            investor_contribution_categories.irrigation_total_usd,
            investor_contribution_categories.rent_capitalizable_total_usd,
            investor_contribution_categories.administration_total_usd,
            investor_contribution_categories.total_seeded_area_ha,
            (((((((investor_contribution_categories.agrochemicals_total_usd + investor_contribution_categories.fertilizers_total_usd) + investor_contribution_categories.seeds_total_usd) + investor_contribution_categories.general_labors_total_usd) + investor_contribution_categories.sowing_total_usd) + investor_contribution_categories.irrigation_total_usd) + investor_contribution_categories.rent_capitalizable_total_usd) + investor_contribution_categories.administration_total_usd) AS total_contributions_usd
           FROM v4_calc.investor_contribution_categories
        ), investor_agrochemicals_real AS (
         SELECT sm.project_id,
            sm.investor_id,
            COALESCE(sum((sm.quantity * s.price)), (0)::numeric) AS agrochemicals_real_usd
           FROM ((public.supply_movements sm
             JOIN public.supplies s ON (((s.id = sm.supply_id) AND (s.deleted_at IS NULL))))
             JOIN public.categories cat ON ((cat.id = s.category_id)))
          WHERE ((sm.deleted_at IS NULL) AND (sm.is_entry = true) AND (sm.movement_type = ANY (ARRAY['Stock'::text, 'Remito oficial'::text, 'Movimiento interno'::text, 'Movimiento interno entrada'::text])) AND (cat.type_id = 2) AND ((cat.name)::text = ANY ((ARRAY['Coadyuvantes'::character varying, 'Curasemillas'::character varying, 'Herbicidas'::character varying, 'Insecticidas'::character varying, 'Fungicidas'::character varying, 'Otros Insumos'::character varying])::text[])))
          GROUP BY sm.project_id, sm.investor_id
        ), investor_fertilizers_real AS (
         SELECT sm.project_id,
            sm.investor_id,
            COALESCE(sum((sm.quantity * s.price)), (0)::numeric) AS fertilizers_real_usd
           FROM ((public.supply_movements sm
             JOIN public.supplies s ON (((s.id = sm.supply_id) AND (s.deleted_at IS NULL))))
             JOIN public.categories cat ON ((cat.id = s.category_id)))
          WHERE ((sm.deleted_at IS NULL) AND (sm.is_entry = true) AND (sm.movement_type = ANY (ARRAY['Stock'::text, 'Remito oficial'::text, 'Movimiento interno'::text, 'Movimiento interno entrada'::text])) AND (cat.type_id = 3))
          GROUP BY sm.project_id, sm.investor_id
        ), investor_seeds_real AS (
         SELECT sm.project_id,
            sm.investor_id,
            COALESCE(sum((sm.quantity * s.price)), (0)::numeric) AS seeds_real_usd
           FROM ((public.supply_movements sm
             JOIN public.supplies s ON (((s.id = sm.supply_id) AND (s.deleted_at IS NULL))))
             JOIN public.categories cat ON ((cat.id = s.category_id)))
          WHERE ((sm.deleted_at IS NULL) AND (sm.is_entry = true) AND (sm.movement_type = ANY (ARRAY['Stock'::text, 'Remito oficial'::text, 'Movimiento interno'::text, 'Movimiento interno entrada'::text])) AND (cat.type_id = 1))
          GROUP BY sm.project_id, sm.investor_id
        ), investor_general_labors_real AS (
         SELECT w.project_id,
            w.investor_id,
            COALESCE(sum((lab.price * w.effective_area)), (0)::numeric) AS general_labors_real_usd
           FROM ((public.workorders w
             JOIN public.labors lab ON (((w.labor_id = lab.id) AND (lab.deleted_at IS NULL))))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.deleted_at IS NULL) AND (cat.type_id = 4) AND ((cat.name)::text = ANY ((ARRAY['Pulverización'::character varying, 'Otras Labores'::character varying])::text[])))
          GROUP BY w.project_id, w.investor_id
        ), investor_sowing_real AS (
         SELECT w.project_id,
            w.investor_id,
            COALESCE(sum((lab.price * w.effective_area)), (0)::numeric) AS sowing_real_usd
           FROM ((public.workorders w
             JOIN public.labors lab ON (((w.labor_id = lab.id) AND (lab.deleted_at IS NULL))))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.deleted_at IS NULL) AND (cat.type_id = 4) AND ((cat.name)::text = 'Siembra'::text))
          GROUP BY w.project_id, w.investor_id
        ), investor_irrigation_real AS (
         SELECT w.project_id,
            w.investor_id,
            COALESCE(sum((lab.price * w.effective_area)), (0)::numeric) AS irrigation_real_usd
           FROM ((public.workorders w
             JOIN public.labors lab ON (((w.labor_id = lab.id) AND (lab.deleted_at IS NULL))))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.deleted_at IS NULL) AND (cat.type_id = 4) AND ((cat.name)::text = 'Riego'::text))
          GROUP BY w.project_id, w.investor_id
        ), investor_rent_real AS (
         SELECT f.project_id,
            fi.investor_id,
            sum((((v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares) * (COALESCE(fi.percentage, 0))::numeric) / (100)::numeric)) AS rent_real_usd
           FROM ((public.fields f
             JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
             LEFT JOIN public.field_investors fi ON (((fi.field_id = f.id) AND (fi.deleted_at IS NULL))))
          WHERE (f.deleted_at IS NULL)
          GROUP BY f.project_id, fi.investor_id
        ), investor_admin_real AS (
         SELECT ib_1.project_id,
            ib_1.investor_id,
                CASE
                    WHEN COALESCE(ac.has_custom_admin, false) THEN ((ct_1.administration_total_usd * (COALESCE(aci.percentage, 0))::numeric) / (100)::numeric)
                    ELSE ((ct_1.administration_total_usd * (ib_1.share_pct_agreed)::numeric) / (100)::numeric)
                END AS admin_real_usd
           FROM (((investor_base ib_1
             JOIN category_totals ct_1 ON ((ct_1.project_id = ib_1.project_id)))
             LEFT JOIN admin_config ac ON ((ac.project_id = ib_1.project_id)))
             LEFT JOIN public.admin_cost_investors aci ON (((aci.project_id = ib_1.project_id) AND (aci.investor_id = ib_1.investor_id) AND (aci.deleted_at IS NULL))))
        )
 SELECT ib.project_id,
    ib.investor_id,
    ib.investor_name,
    ib.share_pct_agreed,
    COALESCE(agro.agrochemicals_real_usd, (0)::numeric) AS agrochemicals_real_usd,
    COALESCE(fert.fertilizers_real_usd, (0)::numeric) AS fertilizers_real_usd,
    COALESCE(seed.seeds_real_usd, (0)::numeric) AS seeds_real_usd,
    COALESCE(glabor.general_labors_real_usd, (0)::numeric) AS general_labors_real_usd,
    COALESCE(sow.sowing_real_usd, (0)::numeric) AS sowing_real_usd,
    COALESCE(irrig.irrigation_real_usd, (0)::numeric) AS irrigation_real_usd,
    COALESCE(rri.rent_real_usd, (0)::numeric) AS rent_real_usd,
    COALESCE(ia.admin_real_usd, (0)::numeric) AS administration_real_usd,
    (((((((COALESCE(agro.agrochemicals_real_usd, (0)::numeric) + COALESCE(fert.fertilizers_real_usd, (0)::numeric)) + COALESCE(seed.seeds_real_usd, (0)::numeric)) + COALESCE(glabor.general_labors_real_usd, (0)::numeric)) + COALESCE(sow.sowing_real_usd, (0)::numeric)) + COALESCE(irrig.irrigation_real_usd, (0)::numeric)) + COALESCE(rri.rent_real_usd, (0)::numeric)) + COALESCE(ia.admin_real_usd, (0)::numeric)) AS total_real_contribution_usd,
    ct.total_contributions_usd AS project_total_contributions_usd,
        CASE
            WHEN (ct.total_contributions_usd > (0)::numeric) THEN (((((((((COALESCE(agro.agrochemicals_real_usd, (0)::numeric) + COALESCE(fert.fertilizers_real_usd, (0)::numeric)) + COALESCE(seed.seeds_real_usd, (0)::numeric)) + COALESCE(glabor.general_labors_real_usd, (0)::numeric)) + COALESCE(sow.sowing_real_usd, (0)::numeric)) + COALESCE(irrig.irrigation_real_usd, (0)::numeric)) + COALESCE(rri.rent_real_usd, (0)::numeric)) + COALESCE(ia.admin_real_usd, (0)::numeric)) / ct.total_contributions_usd) * (100)::numeric)
            ELSE (0)::numeric
        END AS contributions_progress_pct
   FROM (((((((((investor_base ib
     JOIN category_totals ct ON ((ct.project_id = ib.project_id)))
     LEFT JOIN investor_agrochemicals_real agro ON (((agro.project_id = ib.project_id) AND (agro.investor_id = ib.investor_id))))
     LEFT JOIN investor_fertilizers_real fert ON (((fert.project_id = ib.project_id) AND (fert.investor_id = ib.investor_id))))
     LEFT JOIN investor_seeds_real seed ON (((seed.project_id = ib.project_id) AND (seed.investor_id = ib.investor_id))))
     LEFT JOIN investor_general_labors_real glabor ON (((glabor.project_id = ib.project_id) AND (glabor.investor_id = ib.investor_id))))
     LEFT JOIN investor_sowing_real sow ON (((sow.project_id = ib.project_id) AND (sow.investor_id = ib.investor_id))))
     LEFT JOIN investor_irrigation_real irrig ON (((irrig.project_id = ib.project_id) AND (irrig.investor_id = ib.investor_id))))
     LEFT JOIN investor_rent_real rri ON (((rri.project_id = ib.project_id) AND (rri.investor_id = ib.investor_id))))
     LEFT JOIN investor_admin_real ia ON (((ia.project_id = ib.project_id) AND (ia.investor_id = ib.investor_id))));


--
-- Name: workorder_metrics_raw; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.workorder_metrics_raw AS
 WITH base AS (
         SELECT w.id AS workorder_id,
            w.project_id,
            w.field_id,
            w.lot_id,
            w.effective_area,
            lb.price AS labor_price
           FROM (public.workorders w
             JOIN public.labors lb ON (((lb.id = w.labor_id) AND (lb.deleted_at IS NULL))))
          WHERE ((w.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
        ), surface AS (
         SELECT base.project_id,
            base.field_id,
            base.lot_id,
            sum(base.effective_area) AS surface_ha
           FROM base
          GROUP BY base.project_id, base.field_id, base.lot_id
        ), labor_costs AS (
         SELECT base.project_id,
            base.field_id,
            base.lot_id,
            sum((base.labor_price * base.effective_area)) AS labor_cost_usd
           FROM base
          GROUP BY base.project_id, base.field_id, base.lot_id
        ), supply_metrics AS (
         SELECT b.project_id,
            b.field_id,
            b.lot_id,
            sum(
                CASE
                    WHEN (s.unit_id = 1) THEN (wi.final_dose * b.effective_area)
                    ELSE (0)::numeric
                END) AS liters,
            sum(
                CASE
                    WHEN (s.unit_id = 2) THEN (wi.final_dose * b.effective_area)
                    ELSE (0)::numeric
                END) AS kilograms,
            sum(v4_core.supply_cost((wi.final_dose)::numeric, (s.price)::numeric, (b.effective_area)::numeric)) AS supplies_cost_usd
           FROM ((base b
             LEFT JOIN public.workorder_items wi ON (((wi.workorder_id = b.workorder_id) AND (wi.deleted_at IS NULL))))
             LEFT JOIN public.supplies s ON (((s.id = wi.supply_id) AND (s.deleted_at IS NULL))))
          GROUP BY b.project_id, b.field_id, b.lot_id
        )
 SELECT COALESCE(sur.project_id, lc.project_id, sm.project_id) AS project_id,
    COALESCE(sur.field_id, lc.field_id, sm.field_id) AS field_id,
    COALESCE(sur.lot_id, lc.lot_id, sm.lot_id) AS lot_id,
    COALESCE(sur.surface_ha, (0)::numeric) AS surface_ha,
    COALESCE(sm.liters, (0)::numeric) AS liters,
    COALESCE(sm.kilograms, (0)::numeric) AS kilograms,
    COALESCE(lc.labor_cost_usd, (0)::numeric) AS labor_cost_usd,
    COALESCE(sm.supplies_cost_usd, (0)::numeric) AS supplies_cost_usd,
    (COALESCE(lc.labor_cost_usd, (0)::numeric) + COALESCE(sm.supplies_cost_usd, (0)::numeric)) AS direct_cost_usd,
    v4_core.cost_per_ha((COALESCE(lc.labor_cost_usd, (0)::numeric) + COALESCE(sm.supplies_cost_usd, (0)::numeric)), COALESCE(sur.surface_ha, (0)::numeric)) AS avg_cost_per_ha_usd,
    v4_core.per_ha(COALESCE(sm.liters, (0)::numeric), COALESCE(sur.surface_ha, (0)::numeric)) AS liters_per_ha,
    v4_core.per_ha(COALESCE(sm.kilograms, (0)::numeric), COALESCE(sur.surface_ha, (0)::numeric)) AS kilograms_per_ha
   FROM ((surface sur
     FULL JOIN labor_costs lc USING (project_id, field_id, lot_id))
     FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id));


--
-- Name: lot_base_costs; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.lot_base_costs AS
 WITH raw AS (
         SELECT f.project_id,
            f.id AS field_id,
            l.current_crop_id,
            l.id AS lot_id,
            l.name AS lot_name,
            COALESCE(l.hectares, (0)::numeric) AS hectares,
            COALESCE(l.tons, (0)::numeric) AS tons,
            l.sowing_date
           FROM (public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
          WHERE (l.deleted_at IS NULL)
        ), areas AS (
         SELECT r.lot_id,
            v4_ssot.seeded_area_for_lot(r.lot_id) AS seeded_area_ha,
            v4_ssot.harvested_area_for_lot(r.lot_id) AS harvested_area_ha
           FROM raw r
        ), costs AS (
         SELECT workorder_metrics_raw.lot_id,
            max(COALESCE(workorder_metrics_raw.labor_cost_usd, (0)::numeric)) AS labor_cost_usd,
            max(COALESCE(workorder_metrics_raw.supplies_cost_usd, (0)::numeric)) AS supplies_cost_usd,
            max(COALESCE(workorder_metrics_raw.direct_cost_usd, (0)::numeric)) AS direct_cost_usd
           FROM v4_calc.workorder_metrics_raw
          GROUP BY workorder_metrics_raw.lot_id
        ), ssot_values AS (
         SELECT r.lot_id,
            v4_ssot.yield_tn_per_ha_for_lot(r.lot_id) AS yield_tn_per_ha,
            v4_ssot.income_net_total_for_lot(r.lot_id) AS income_net_total_usd,
            v4_ssot.rent_per_ha_for_lot(r.lot_id) AS rent_per_ha_usd,
            v4_ssot.rent_fixed_only_for_lot(r.lot_id) AS rent_fixed_per_ha_usd,
            v4_ssot.admin_cost_per_ha_for_lot(r.lot_id) AS admin_cost_per_ha_usd
           FROM raw r
        ), derived AS (
         SELECT r.project_id,
            r.field_id,
            r.current_crop_id,
            r.lot_id,
            r.lot_name,
            r.hectares,
            r.tons,
            r.sowing_date,
            COALESCE(a.seeded_area_ha, (0)::numeric) AS seeded_area_ha,
            COALESCE(a.harvested_area_ha, (0)::numeric) AS harvested_area_ha,
            COALESCE(c.labor_cost_usd, (0)::numeric) AS labor_cost_usd,
            COALESCE(c.supplies_cost_usd, (0)::numeric) AS supplies_cost_usd,
            COALESCE(c.direct_cost_usd, (0)::numeric) AS direct_cost_usd,
            COALESCE(s.yield_tn_per_ha, (0)::numeric) AS yield_tn_per_ha,
            COALESCE(s.income_net_total_usd, (0)::numeric) AS income_net_total_usd,
            COALESCE(s.rent_per_ha_usd, (0)::numeric) AS rent_per_ha_usd,
            COALESCE(s.rent_fixed_per_ha_usd, (0)::numeric) AS rent_fixed_per_ha_usd,
            COALESCE(s.admin_cost_per_ha_usd, (0)::numeric) AS admin_cost_per_ha_usd
           FROM (((raw r
             LEFT JOIN areas a ON ((a.lot_id = r.lot_id)))
             LEFT JOIN costs c ON ((c.lot_id = r.lot_id)))
             LEFT JOIN ssot_values s ON ((s.lot_id = r.lot_id)))
        )
 SELECT project_id,
    field_id,
    current_crop_id,
    lot_id,
    lot_name,
    hectares,
    tons,
    sowing_date,
    seeded_area_ha,
    harvested_area_ha,
    yield_tn_per_ha,
    labor_cost_usd,
    supplies_cost_usd,
    direct_cost_usd,
    income_net_total_usd,
    v4_core.per_ha(income_net_total_usd, hectares) AS income_net_per_ha_usd,
    v4_core.per_ha(direct_cost_usd, hectares) AS direct_cost_per_ha_usd,
    rent_per_ha_usd,
    rent_fixed_per_ha_usd,
    admin_cost_per_ha_usd,
    seeded_area_ha AS sowed_area_ha,
    seeded_area_ha AS sown_area_ha
   FROM derived d;


--
-- Name: lot_base_income; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.lot_base_income AS
 SELECT project_id,
    field_id,
    current_crop_id,
    lot_id,
    lot_name,
    hectares,
    tons,
    seeded_area_ha,
    yield_tn_per_ha,
    COALESCE(v4_ssot.net_price_usd_for_lot(lot_id), (0)::numeric) AS net_price_usd_tn,
    income_net_total_usd,
    income_net_per_ha_usd,
    seeded_area_ha AS sowed_area_ha,
    seeded_area_ha AS sown_area_ha
   FROM v4_calc.lot_base_costs c;


--
-- Name: workorder_metrics; Type: VIEW; Schema: v4_calc; Owner: -
--

CREATE VIEW v4_calc.workorder_metrics AS
 WITH base AS (
         SELECT w.id AS workorder_id,
            w.project_id,
            w.field_id,
            w.lot_id,
            w.effective_area,
            lb.price AS labor_price
           FROM (public.workorders w
             JOIN public.labors lb ON (((lb.id = w.labor_id) AND (lb.deleted_at IS NULL))))
          WHERE ((w.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
        ), surface AS (
         SELECT base.project_id,
            base.field_id,
            base.lot_id,
            sum(base.effective_area) AS surface_ha
           FROM base
          GROUP BY base.project_id, base.field_id, base.lot_id
        ), labor_costs AS (
         SELECT base.project_id,
            base.field_id,
            base.lot_id,
            sum((base.labor_price * base.effective_area)) AS labor_cost_usd
           FROM base
          GROUP BY base.project_id, base.field_id, base.lot_id
        ), supply_metrics AS (
         SELECT b.project_id,
            b.field_id,
            b.lot_id,
            sum(
                CASE
                    WHEN (s.unit_id = 1) THEN (wi.final_dose * b.effective_area)
                    ELSE (0)::numeric
                END) AS liters,
            sum(
                CASE
                    WHEN (s.unit_id = 2) THEN (wi.final_dose * b.effective_area)
                    ELSE (0)::numeric
                END) AS kilograms,
            sum((COALESCE(wi.total_used, (0)::numeric) * COALESCE(s.price, (0)::numeric))) AS supplies_cost_usd
           FROM ((base b
             LEFT JOIN public.workorder_items wi ON (((wi.workorder_id = b.workorder_id) AND (wi.deleted_at IS NULL))))
             LEFT JOIN public.supplies s ON (((s.id = wi.supply_id) AND (s.deleted_at IS NULL))))
          GROUP BY b.project_id, b.field_id, b.lot_id
        )
 SELECT COALESCE(sur.project_id, lc.project_id, sm.project_id) AS project_id,
    COALESCE(sur.field_id, lc.field_id, sm.field_id) AS field_id,
    COALESCE(sur.lot_id, lc.lot_id, sm.lot_id) AS lot_id,
    COALESCE(sur.surface_ha, (0)::numeric) AS surface_ha,
    COALESCE(sm.liters, (0)::numeric) AS liters,
    COALESCE(sm.kilograms, (0)::numeric) AS kilograms,
    COALESCE(lc.labor_cost_usd, (0)::numeric) AS labor_cost_usd,
    COALESCE(sm.supplies_cost_usd, (0)::numeric) AS supplies_cost_usd,
    (COALESCE(lc.labor_cost_usd, (0)::numeric) + COALESCE(sm.supplies_cost_usd, (0)::numeric)) AS direct_cost_usd,
    v4_core.cost_per_ha((COALESCE(lc.labor_cost_usd, (0)::numeric) + COALESCE(sm.supplies_cost_usd, (0)::numeric)), COALESCE(sur.surface_ha, (0)::numeric)) AS avg_cost_per_ha_usd,
    v4_core.per_ha(COALESCE(sm.liters, (0)::numeric), COALESCE(sur.surface_ha, (0)::numeric)) AS liters_per_ha,
    v4_core.per_ha(COALESCE(sm.kilograms, (0)::numeric), COALESCE(sur.surface_ha, (0)::numeric)) AS kilograms_per_ha
   FROM ((surface sur
     FULL JOIN labor_costs lc USING (project_id, field_id, lot_id))
     FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id));


--
-- Name: dashboard_contributions_progress; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_contributions_progress AS
 SELECT project_id,
    investor_id,
    investor_name,
    share_pct_agreed AS investor_percentage_pct,
    contributions_progress_pct
   FROM v4_calc.investor_real_contributions;


--
-- Name: dashboard_crop_incidence; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_crop_incidence AS
 SELECT p.id AS project_id,
    ci.current_crop_id,
    ci.crop_name,
    ci.crop_hectares,
    ci.crop_incidence_pct,
    v4_ssot.cost_per_ha_for_crop_ssot(p.id, ci.current_crop_id) AS cost_per_ha_usd
   FROM (public.projects p
     CROSS JOIN LATERAL v4_ssot.crop_incidence_for_project(p.id) ci(current_crop_id, crop_name, crop_hectares, crop_incidence_pct))
  WHERE (p.deleted_at IS NULL)
  ORDER BY p.id, ci.crop_name;


--
-- Name: dashboard_crop_incidence_field; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_crop_incidence_field AS
 WITH field_totals AS (
         SELECT field_crop_aggregated.project_id,
            field_crop_aggregated.field_id,
            sum(field_crop_aggregated.surface_ha) AS total_hectares
           FROM v4_calc.field_crop_aggregated
          GROUP BY field_crop_aggregated.project_id, field_crop_aggregated.field_id
        ), by_crop AS (
         SELECT field_crop_aggregated.project_id,
            field_crop_aggregated.field_id,
            field_crop_aggregated.crop_id AS current_crop_id,
            sum(field_crop_aggregated.surface_ha) AS crop_hectares
           FROM v4_calc.field_crop_aggregated
          GROUP BY field_crop_aggregated.project_id, field_crop_aggregated.field_id, field_crop_aggregated.crop_id
        )
 SELECT p.id AS project_id,
    p.customer_id,
    p.campaign_id,
    bc.field_id,
    bc.current_crop_id,
    c.name AS crop_name,
    bc.crop_hectares,
    v4_core.percentage(bc.crop_hectares, COALESCE(ft.total_hectares, (0)::numeric)) AS crop_incidence_pct,
    v4_ssot.cost_per_ha_for_crop_ssot(p.id, bc.current_crop_id) AS cost_per_ha_usd
   FROM (((by_crop bc
     JOIN public.projects p ON (((p.id = bc.project_id) AND (p.deleted_at IS NULL))))
     LEFT JOIN field_totals ft ON (((ft.project_id = bc.project_id) AND (ft.field_id = bc.field_id))))
     LEFT JOIN public.crops c ON (((c.id = bc.current_crop_id) AND (c.deleted_at IS NULL))))
  ORDER BY bc.project_id, bc.field_id, c.name;


--
-- Name: dashboard_management_balance; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_management_balance AS
 SELECT p.id AS project_id,
    COALESCE(sum(v4_ssot.income_net_total_for_lot(l.id)), (0)::numeric) AS income_usd,
    v4_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
    v4_ssot.renta_pct(v4_ssot.operating_result_total_for_project(p.id), v4_ssot.total_costs_for_project(p.id)) AS operating_result_pct,
    v4_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
    (v4_ssot.supply_movements_invested_total_for_project(p.id) + COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), (0)::numeric)) AS costos_directos_invertidos_usd,
    ((v4_ssot.supply_movements_invested_total_for_project(p.id) + COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), (0)::numeric)) - v4_ssot.direct_costs_total_for_project(p.id)) AS costos_directos_stock_usd,
    COALESCE(sc.semillas_ejecutados_usd, (0)::numeric) AS semillas_ejecutados_usd,
    v4_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
    (v4_ssot.seeds_invested_for_project_mb(p.id) - COALESCE(sc.semillas_ejecutados_usd, (0)::numeric)) AS semillas_stock_usd,
    COALESCE(sc.agroquimicos_ejecutados_usd, (0)::numeric) AS agroquimicos_ejecutados_usd,
    v4_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
    (v4_ssot.agrochemicals_invested_for_project_mb(p.id) - COALESCE(sc.agroquimicos_ejecutados_usd, (0)::numeric)) AS agroquimicos_stock_usd,
    COALESCE(sc.fertilizantes_ejecutados_usd, (0)::numeric) AS fertilizantes_ejecutados_usd,
    COALESCE(fi.fertilizantes_invertidos_usd, (0)::numeric) AS fertilizantes_invertidos_usd,
    (COALESCE(fi.fertilizantes_invertidos_usd, (0)::numeric) - COALESCE(sc.fertilizantes_ejecutados_usd, (0)::numeric)) AS fertilizantes_stock_usd,
    COALESCE(sum(v4_ssot.labor_cost_for_lot(l.id)), (0)::numeric) AS labores_ejecutados_usd,
    COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), (0)::numeric) AS labores_invertidos_usd,
    v4_ssot.lease_invested_for_project(p.id) AS arriendo_ejecutados_usd,
    v4_ssot.lease_executed_for_project(p.id) AS arriendo_invertidos_usd,
    v4_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
    v4_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
    COALESCE(sc.semillas_ejecutados_usd, (0)::numeric) AS semilla_cost,
    COALESCE(sc.agroquimicos_ejecutados_usd, (0)::numeric) AS insumos_cost,
    COALESCE(sum(v4_ssot.labor_cost_for_lot(l.id)), (0)::numeric) AS labores_cost,
    COALESCE(sc.fertilizantes_ejecutados_usd, (0)::numeric) AS fertilizantes_cost
   FROM ((((public.projects p
     LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
     LEFT JOIN v4_calc.dashboard_supply_costs_by_project sc ON ((sc.project_id = p.id)))
     LEFT JOIN v4_calc.dashboard_fertilizers_invested_by_project fi ON ((fi.project_id = p.id)))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.id, sc.semillas_ejecutados_usd, sc.agroquimicos_ejecutados_usd, sc.fertilizantes_ejecutados_usd, fi.fertilizantes_invertidos_usd;


--
-- Name: dashboard_management_balance_field; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_management_balance_field AS
 WITH lots_base AS (
         SELECT p.id AS project_id,
            p.customer_id,
            p.campaign_id,
            f.id AS field_id,
            l.id AS lot_id,
            l.hectares
           FROM ((public.projects p
             JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
             JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
          WHERE (p.deleted_at IS NULL)
        ), income_totals AS (
         SELECT lots_base.project_id,
            lots_base.field_id,
            sum(v4_ssot.income_net_total_for_lot(lots_base.lot_id)) AS income_usd
           FROM lots_base
          GROUP BY lots_base.project_id, lots_base.field_id
        ), direct_costs AS (
         SELECT lots_base.project_id,
            lots_base.field_id,
            sum(v4_ssot.direct_cost_for_lot(lots_base.lot_id)) AS direct_costs_usd
           FROM lots_base
          GROUP BY lots_base.project_id, lots_base.field_id
        ), rent_totals AS (
         SELECT lots_base.project_id,
            lots_base.field_id,
            sum((v4_ssot.rent_per_ha_for_lot(lots_base.lot_id) * lots_base.hectares)) AS rent_total_usd
           FROM lots_base
          GROUP BY lots_base.project_id, lots_base.field_id
        ), admin_totals AS (
         SELECT lots_base.project_id,
            lots_base.field_id,
            sum((v4_ssot.admin_cost_per_ha_for_lot(lots_base.lot_id) * lots_base.hectares)) AS admin_total_usd
           FROM lots_base
          GROUP BY lots_base.project_id, lots_base.field_id
        ), supply_costs AS (
         SELECT field_crop_supply_costs_by_lot.project_id,
            field_crop_supply_costs_by_lot.field_id,
            sum(field_crop_supply_costs_by_lot.semillas_usd) AS semillas_usd,
            sum((((field_crop_supply_costs_by_lot.herbicidas_usd + field_crop_supply_costs_by_lot.insecticidas_usd) + field_crop_supply_costs_by_lot.fungicidas_usd) + field_crop_supply_costs_by_lot.coadyuvantes_usd)) AS agroquimicos_usd,
            sum(field_crop_supply_costs_by_lot.fertilizantes_usd) AS fertilizantes_usd
           FROM v4_calc.field_crop_supply_costs_by_lot
          GROUP BY field_crop_supply_costs_by_lot.project_id, field_crop_supply_costs_by_lot.field_id
        ), labor_costs AS (
         SELECT field_crop_labor_costs_by_lot.project_id,
            field_crop_labor_costs_by_lot.field_id,
            sum(field_crop_labor_costs_by_lot.total_labores_usd) AS labor_total_usd
           FROM v4_calc.field_crop_labor_costs_by_lot
          GROUP BY field_crop_labor_costs_by_lot.project_id, field_crop_labor_costs_by_lot.field_id
        )
 SELECT lb.project_id,
    lb.customer_id,
    lb.campaign_id,
    lb.field_id,
    COALESCE(it.income_usd, (0)::numeric) AS income_usd,
    (((COALESCE(it.income_usd, (0)::numeric) - COALESCE(dc.direct_costs_usd, (0)::numeric)) - COALESCE(rt.rent_total_usd, (0)::numeric)) - COALESCE(ad.admin_total_usd, (0)::numeric)) AS operating_result_usd,
    v4_ssot.renta_pct((((COALESCE(it.income_usd, (0)::numeric) - COALESCE(dc.direct_costs_usd, (0)::numeric)) - COALESCE(rt.rent_total_usd, (0)::numeric)) - COALESCE(ad.admin_total_usd, (0)::numeric)), ((COALESCE(dc.direct_costs_usd, (0)::numeric) + COALESCE(rt.rent_total_usd, (0)::numeric)) + COALESCE(ad.admin_total_usd, (0)::numeric))) AS operating_result_pct,
    COALESCE(dc.direct_costs_usd, (0)::numeric) AS costos_directos_ejecutados_usd,
    COALESCE(dc.direct_costs_usd, (0)::numeric) AS costos_directos_invertidos_usd,
    (0)::numeric AS costos_directos_stock_usd,
    COALESCE(sc.semillas_usd, (0)::numeric) AS semillas_ejecutados_usd,
    COALESCE(sc.semillas_usd, (0)::numeric) AS semillas_invertidos_usd,
    (0)::numeric AS semillas_stock_usd,
    COALESCE(sc.agroquimicos_usd, (0)::numeric) AS agroquimicos_ejecutados_usd,
    COALESCE(sc.agroquimicos_usd, (0)::numeric) AS agroquimicos_invertidos_usd,
    (0)::numeric AS agroquimicos_stock_usd,
    COALESCE(sc.fertilizantes_usd, (0)::numeric) AS fertilizantes_ejecutados_usd,
    COALESCE(sc.fertilizantes_usd, (0)::numeric) AS fertilizantes_invertidos_usd,
    (0)::numeric AS fertilizantes_stock_usd,
    COALESCE(lc.labor_total_usd, (0)::numeric) AS labores_ejecutados_usd,
    COALESCE(lc.labor_total_usd, (0)::numeric) AS labores_invertidos_usd,
    COALESCE(rt.rent_total_usd, (0)::numeric) AS arriendo_ejecutados_usd,
    COALESCE(rt.rent_total_usd, (0)::numeric) AS arriendo_invertidos_usd,
    COALESCE(ad.admin_total_usd, (0)::numeric) AS estructura_ejecutados_usd,
    COALESCE(ad.admin_total_usd, (0)::numeric) AS estructura_invertidos_usd,
    COALESCE(sc.semillas_usd, (0)::numeric) AS semilla_cost,
    COALESCE(sc.agroquimicos_usd, (0)::numeric) AS insumos_cost,
    COALESCE(lc.labor_total_usd, (0)::numeric) AS labores_cost,
    COALESCE(sc.fertilizantes_usd, (0)::numeric) AS fertilizantes_cost
   FROM ((((((( SELECT DISTINCT lots_base.project_id,
            lots_base.customer_id,
            lots_base.campaign_id,
            lots_base.field_id
           FROM lots_base) lb
     LEFT JOIN income_totals it ON (((it.project_id = lb.project_id) AND (it.field_id = lb.field_id))))
     LEFT JOIN direct_costs dc ON (((dc.project_id = lb.project_id) AND (dc.field_id = lb.field_id))))
     LEFT JOIN rent_totals rt ON (((rt.project_id = lb.project_id) AND (rt.field_id = lb.field_id))))
     LEFT JOIN admin_totals ad ON (((ad.project_id = lb.project_id) AND (ad.field_id = lb.field_id))))
     LEFT JOIN supply_costs sc ON (((sc.project_id = lb.project_id) AND (sc.field_id = lb.field_id))))
     LEFT JOIN labor_costs lc ON (((lc.project_id = lb.project_id) AND (lc.field_id = lb.field_id))));


--
-- Name: lot_metrics; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.lot_metrics AS
 WITH project_totals AS (
         SELECT lot_base_costs.project_id,
            sum(lot_base_costs.hectares) AS total_hectares
           FROM v4_calc.lot_base_costs
          GROUP BY lot_base_costs.project_id
        ), field_totals AS (
         SELECT lot_base_costs.field_id,
            sum(lot_base_costs.hectares) AS total_hectares
           FROM v4_calc.lot_base_costs
          GROUP BY lot_base_costs.field_id
        )
 SELECT c.project_id,
    c.field_id,
    c.lot_id,
    c.lot_name,
    c.hectares,
    c.seeded_area_ha,
    c.harvested_area_ha,
    c.yield_tn_per_ha,
    c.tons,
    c.sowing_date,
    c.labor_cost_usd,
    c.supplies_cost_usd AS supply_cost_usd,
    c.direct_cost_usd,
    c.income_net_total_usd,
    c.income_net_per_ha_usd,
    c.rent_per_ha_usd,
    c.admin_cost_per_ha_usd,
    c.direct_cost_per_ha_usd,
    ((c.direct_cost_per_ha_usd + c.rent_per_ha_usd) + c.admin_cost_per_ha_usd) AS active_total_per_ha_usd,
    (c.income_net_per_ha_usd - ((c.direct_cost_per_ha_usd + c.rent_per_ha_usd) + c.admin_cost_per_ha_usd)) AS operating_result_per_ha_usd,
    (c.rent_per_ha_usd * c.hectares) AS rent_total_usd,
    (c.admin_cost_per_ha_usd * c.hectares) AS admin_total_usd,
    (((c.direct_cost_per_ha_usd + c.rent_per_ha_usd) + c.admin_cost_per_ha_usd) * c.hectares) AS active_total_usd,
    ((c.income_net_per_ha_usd - ((c.direct_cost_per_ha_usd + c.rent_per_ha_usd) + c.admin_cost_per_ha_usd)) * c.hectares) AS operating_result_total_usd,
    c.direct_cost_usd AS direct_cost_total_usd,
    COALESCE(pt.total_hectares, (0)::numeric) AS project_total_hectares,
    COALESCE(ft.total_hectares, (0)::numeric) AS field_total_hectares,
    c.seeded_area_ha AS sowed_area_ha,
    c.seeded_area_ha AS sown_area_ha
   FROM ((v4_calc.lot_base_costs c
     LEFT JOIN project_totals pt ON ((pt.project_id = c.project_id)))
     LEFT JOIN field_totals ft ON ((ft.field_id = c.field_id)));


--
-- Name: dashboard_metrics; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_metrics AS
 WITH lot_data AS (
         SELECT lm.project_id,
            lm.lot_id,
            lm.hectares,
            lm.seeded_area_ha,
            lm.harvested_area_ha,
            lm.direct_cost_per_ha_usd
           FROM v4_report.lot_metrics lm
        ), project_hectares AS (
         SELECT lot_metrics.project_id,
            sum(lot_metrics.hectares) AS total_hectares
           FROM v4_report.lot_metrics
          GROUP BY lot_metrics.project_id
        ), rent_fixed_ssot AS (
         SELECT f.project_id,
            sum((v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)) AS rent_fixed_total_usd
           FROM (public.fields f
             JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
          WHERE (f.deleted_at IS NULL)
          GROUP BY f.project_id
        )
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(sum(ld.seeded_area_ha), (0)::numeric) AS sowing_hectares,
    COALESCE(sum(ld.hectares), (0)::numeric) AS sowing_total_hectares,
    v4_core.percentage(COALESCE(sum(ld.seeded_area_ha), (0)::numeric), COALESCE(sum(ld.hectares), (0)::numeric)) AS sowing_progress_pct,
    COALESCE(sum(ld.harvested_area_ha), (0)::numeric) AS harvest_hectares,
    COALESCE(sum(ld.hectares), (0)::numeric) AS harvest_total_hectares,
    v4_core.percentage(COALESCE(sum(ld.harvested_area_ha), (0)::numeric), COALESCE(sum(ld.hectares), (0)::numeric)) AS harvest_progress_pct,
    COALESCE((sum((ld.direct_cost_per_ha_usd * ld.seeded_area_ha)) / NULLIF(sum(ld.seeded_area_ha), (0)::numeric)), (0)::numeric) AS executed_costs_usd,
    COALESCE(p.planned_cost, (0)::numeric) AS budget_cost_usd,
    v4_core.percentage(COALESCE((sum((ld.direct_cost_per_ha_usd * ld.seeded_area_ha)) / NULLIF(sum(ld.seeded_area_ha), (0)::numeric)), (0)::numeric), COALESCE(p.planned_cost, (0)::numeric)) AS costs_progress_pct,
    COALESCE(sum(v4_ssot.income_net_total_for_lot(ld.lot_id)), (0)::numeric) AS operating_result_income_usd,
    v4_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
    ((COALESCE(v4_ssot.direct_costs_total_for_project(p.id), (0)::numeric) + COALESCE((p.admin_cost * ph.total_hectares), (0)::numeric)) + COALESCE(rfs.rent_fixed_total_usd, (0)::numeric)) AS operating_result_total_costs_usd,
    v4_ssot.renta_pct(v4_ssot.operating_result_total_for_project(p.id), ((COALESCE(v4_ssot.direct_costs_total_for_project(p.id), (0)::numeric) + COALESCE((p.admin_cost * ph.total_hectares), (0)::numeric)) + COALESCE(rfs.rent_fixed_total_usd, (0)::numeric))) AS operating_result_pct,
    COALESCE(ph.total_hectares, (0)::numeric) AS project_total_hectares
   FROM (((public.projects p
     LEFT JOIN lot_data ld ON ((ld.project_id = p.id)))
     LEFT JOIN project_hectares ph ON ((ph.project_id = p.id)))
     LEFT JOIN rent_fixed_ssot rfs ON ((rfs.project_id = p.id)))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares, rfs.rent_fixed_total_usd;


--
-- Name: dashboard_metrics_field; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_metrics_field AS
 WITH lot_data AS (
         SELECT lm.project_id,
            lm.field_id,
            lm.lot_id,
            lm.hectares,
            lm.seeded_area_ha,
            lm.harvested_area_ha,
            lm.direct_cost_per_ha_usd
           FROM v4_report.lot_metrics lm
        ), field_hectares AS (
         SELECT lot_metrics.project_id,
            lot_metrics.field_id,
            sum(lot_metrics.hectares) AS total_hectares
           FROM v4_report.lot_metrics
          GROUP BY lot_metrics.project_id, lot_metrics.field_id
        ), project_hectares AS (
         SELECT lot_metrics.project_id,
            sum(lot_metrics.hectares) AS total_hectares
           FROM v4_report.lot_metrics
          GROUP BY lot_metrics.project_id
        ), rent_fixed_ssot AS (
         SELECT f.project_id,
            f.id AS field_id,
            sum((v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)) AS rent_fixed_total_usd
           FROM (public.fields f
             JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
          WHERE (f.deleted_at IS NULL)
          GROUP BY f.project_id, f.id
        ), direct_costs AS (
         SELECT lot_metrics.project_id,
            lot_metrics.field_id,
            sum(v4_ssot.direct_cost_for_lot(lot_metrics.lot_id)) AS direct_costs_total_usd
           FROM v4_report.lot_metrics
          GROUP BY lot_metrics.project_id, lot_metrics.field_id
        ), admin_costs AS (
         SELECT lot_metrics.project_id,
            lot_metrics.field_id,
            sum((v4_ssot.admin_cost_per_ha_for_lot(lot_metrics.lot_id) * lot_metrics.hectares)) AS admin_total_usd
           FROM v4_report.lot_metrics
          GROUP BY lot_metrics.project_id, lot_metrics.field_id
        ), rent_totals AS (
         SELECT lot_metrics.project_id,
            lot_metrics.field_id,
            sum((v4_ssot.rent_per_ha_for_lot(lot_metrics.lot_id) * lot_metrics.hectares)) AS rent_total_usd
           FROM v4_report.lot_metrics
          GROUP BY lot_metrics.project_id, lot_metrics.field_id
        ), income_totals AS (
         SELECT lot_metrics.project_id,
            lot_metrics.field_id,
            sum(v4_ssot.income_net_total_for_lot(lot_metrics.lot_id)) AS income_net_total_usd
           FROM v4_report.lot_metrics
          GROUP BY lot_metrics.project_id, lot_metrics.field_id
        )
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    ld.field_id,
    COALESCE(sum(ld.seeded_area_ha), (0)::numeric) AS sowing_hectares,
    COALESCE(sum(ld.hectares), (0)::numeric) AS sowing_total_hectares,
    v4_core.percentage(COALESCE(sum(ld.seeded_area_ha), (0)::numeric), COALESCE(sum(ld.hectares), (0)::numeric)) AS sowing_progress_pct,
    COALESCE(sum(ld.harvested_area_ha), (0)::numeric) AS harvest_hectares,
    COALESCE(sum(ld.hectares), (0)::numeric) AS harvest_total_hectares,
    v4_core.percentage(COALESCE(sum(ld.harvested_area_ha), (0)::numeric), COALESCE(sum(ld.hectares), (0)::numeric)) AS harvest_progress_pct,
    COALESCE((sum((ld.direct_cost_per_ha_usd * ld.seeded_area_ha)) / NULLIF(sum(ld.seeded_area_ha), (0)::numeric)), (0)::numeric) AS executed_costs_usd,
    (COALESCE(p.planned_cost, (0)::numeric) * v4_core.safe_div(COALESCE(fh.total_hectares, (0)::numeric), COALESCE(ph.total_hectares, (0)::numeric))) AS budget_cost_usd,
    v4_core.percentage(COALESCE((sum((ld.direct_cost_per_ha_usd * ld.seeded_area_ha)) / NULLIF(sum(ld.seeded_area_ha), (0)::numeric)), (0)::numeric), (COALESCE(p.planned_cost, (0)::numeric) * v4_core.safe_div(COALESCE(fh.total_hectares, (0)::numeric), COALESCE(ph.total_hectares, (0)::numeric)))) AS costs_progress_pct,
    COALESCE(it.income_net_total_usd, (0)::numeric) AS operating_result_income_usd,
    (((COALESCE(it.income_net_total_usd, (0)::numeric) - COALESCE(dc.direct_costs_total_usd, (0)::numeric)) - COALESCE(rt.rent_total_usd, (0)::numeric)) - COALESCE(ac.admin_total_usd, (0)::numeric)) AS operating_result_usd,
    ((COALESCE(dc.direct_costs_total_usd, (0)::numeric) + COALESCE(ac.admin_total_usd, (0)::numeric)) + COALESCE(rfs.rent_fixed_total_usd, (0)::numeric)) AS operating_result_total_costs_usd,
    v4_ssot.renta_pct((((COALESCE(it.income_net_total_usd, (0)::numeric) - COALESCE(dc.direct_costs_total_usd, (0)::numeric)) - COALESCE(rt.rent_total_usd, (0)::numeric)) - COALESCE(ac.admin_total_usd, (0)::numeric)), ((COALESCE(dc.direct_costs_total_usd, (0)::numeric) + COALESCE(ac.admin_total_usd, (0)::numeric)) + COALESCE(rfs.rent_fixed_total_usd, (0)::numeric))) AS operating_result_pct,
    COALESCE(ph.total_hectares, (0)::numeric) AS project_total_hectares
   FROM ((((((((public.projects p
     LEFT JOIN lot_data ld ON ((ld.project_id = p.id)))
     LEFT JOIN field_hectares fh ON (((fh.project_id = p.id) AND (fh.field_id = ld.field_id))))
     LEFT JOIN project_hectares ph ON ((ph.project_id = p.id)))
     LEFT JOIN rent_fixed_ssot rfs ON (((rfs.project_id = p.id) AND (rfs.field_id = ld.field_id))))
     LEFT JOIN direct_costs dc ON (((dc.project_id = p.id) AND (dc.field_id = ld.field_id))))
     LEFT JOIN admin_costs ac ON (((ac.project_id = p.id) AND (ac.field_id = ld.field_id))))
     LEFT JOIN rent_totals rt ON (((rt.project_id = p.id) AND (rt.field_id = ld.field_id))))
     LEFT JOIN income_totals it ON (((it.project_id = p.id) AND (it.field_id = ld.field_id))))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, p.planned_cost, ph.total_hectares, fh.total_hectares, rfs.rent_fixed_total_usd, dc.direct_costs_total_usd, ac.admin_total_usd, rt.rent_total_usd, it.income_net_total_usd, ld.field_id;


--
-- Name: dashboard_operational_indicators; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_operational_indicators AS
 SELECT id AS project_id,
    v4_ssot.first_workorder_date_for_project(id) AS start_date,
    v4_ssot.last_workorder_date_for_project(id) AS end_date,
    v4_core.calculate_campaign_closing_date(v4_ssot.last_workorder_date_for_project(id)) AS campaign_closing_date,
    v4_ssot.first_workorder_number_for_project(id) AS first_workorder_id,
    v4_ssot.last_workorder_number_for_project(id) AS last_workorder_id,
    v4_ssot.last_stock_count_date_for_project(id) AS last_stock_count_date
   FROM public.projects p
  WHERE (deleted_at IS NULL);


--
-- Name: dashboard_operational_indicators_field; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.dashboard_operational_indicators_field AS
 SELECT p.id AS project_id,
    p.customer_id,
    p.campaign_id,
    f.id AS field_id,
    min(w.date) AS start_date,
    max(w.date) AS end_date,
    v4_core.calculate_campaign_closing_date(max(w.date)) AS campaign_closing_date,
    min((w.number)::text) AS first_workorder_id,
    max((w.number)::text) AS last_workorder_id,
    v4_ssot.last_stock_count_date_for_project(p.id) AS last_stock_count_date
   FROM ((public.projects p
     JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.workorders w ON (((w.field_id = f.id) AND (w.deleted_at IS NULL))))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.id, p.customer_id, p.campaign_id, f.id;


--
-- Name: field_crop_cultivos; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.field_crop_cultivos AS
 WITH lot_base AS (
         SELECT f.project_id,
            f.id AS field_id,
            f.name AS field_name,
            l.current_crop_id AS crop_id,
            c.name AS crop_name,
            l.id AS lot_id,
            l.hectares,
            l.tons,
            v4_ssot.seeded_area_for_lot(l.id) AS seeded_area_ha,
            v4_ssot.harvested_area_for_lot(l.id) AS harvested_area_ha
           FROM ((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             JOIN public.crops c ON (((c.id = l.current_crop_id) AND (c.deleted_at IS NULL))))
          WHERE ((l.deleted_at IS NULL) AND (l.current_crop_id IS NOT NULL))
        )
 SELECT project_id,
    field_id,
    field_name,
    crop_id AS current_crop_id,
    crop_name,
    sum(hectares) AS superficie_total,
    sum(seeded_area_ha) AS superficie_sembrada_ha,
    sum(harvested_area_ha) AS area_cosechada_ha,
    sum(tons) AS produccion_tn,
    v4_ssot.board_price_for_lot(min(lot_id)) AS precio_bruto_usd_tn,
    v4_ssot.freight_cost_for_lot(min(lot_id)) AS gasto_flete_usd_tn,
    v4_ssot.commercial_cost_for_lot(min(lot_id)) AS gasto_comercial_usd_tn,
    v4_ssot.net_price_usd_for_lot(min(lot_id)) AS precio_neto_usd_tn,
    v4_core.safe_div(sum(tons), sum(hectares)) AS rendimiento_tn_ha,
    (v4_core.safe_div(sum(tons), sum(hectares)) * v4_ssot.net_price_usd_for_lot(min(lot_id))) AS ingreso_neto_por_ha
   FROM lot_base
  GROUP BY project_id, field_id, field_name, crop_id, crop_name;


--
-- Name: field_crop_economicos; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.field_crop_economicos AS
 SELECT project_id,
    field_id,
    crop_id AS current_crop_id,
    (labor_costs_usd + supply_costs_usd) AS gastos_directos_usd,
    v4_core.safe_div((labor_costs_usd + supply_costs_usd), surface_ha) AS gastos_directos_usd_ha,
    ((production_tn * v4_ssot.net_price_usd_for_lot(sample_lot_id)) - (labor_costs_usd + supply_costs_usd)) AS margen_bruto_usd,
    ((v4_core.safe_div(production_tn, surface_ha) * v4_ssot.net_price_usd_for_lot(sample_lot_id)) - v4_core.safe_div((labor_costs_usd + supply_costs_usd), surface_ha)) AS margen_bruto_usd_ha,
    rent_fixed_usd AS arriendo_usd,
    v4_core.safe_div(rent_fixed_usd, surface_ha) AS arriendo_usd_ha,
    administration_usd AS adm_estructura_usd,
    v4_core.safe_div(administration_usd, surface_ha) AS adm_estructura_usd_ha,
    ((((production_tn * v4_ssot.net_price_usd_for_lot(sample_lot_id)) - (labor_costs_usd + supply_costs_usd)) - rent_total_usd) - administration_usd) AS resultado_operativo_usd,
    ((((v4_core.safe_div(production_tn, surface_ha) * v4_ssot.net_price_usd_for_lot(sample_lot_id)) - v4_core.safe_div((labor_costs_usd + supply_costs_usd), surface_ha)) - v4_core.safe_div(rent_total_usd, surface_ha)) - v4_core.safe_div(administration_usd, surface_ha)) AS resultado_operativo_usd_ha
   FROM v4_calc.field_crop_aggregated;


--
-- Name: field_crop_insumos; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.field_crop_insumos AS
 WITH supply_costs AS (
         SELECT field_crop_supply_costs_by_lot.project_id,
            field_crop_supply_costs_by_lot.field_id,
            field_crop_supply_costs_by_lot.crop_id,
            field_crop_supply_costs_by_lot.lot_id,
            field_crop_supply_costs_by_lot.seeded_area_ha,
            field_crop_supply_costs_by_lot.surface_ha,
            field_crop_supply_costs_by_lot.semillas_usd,
            field_crop_supply_costs_by_lot.curasemillas_usd,
            field_crop_supply_costs_by_lot.herbicidas_usd,
            field_crop_supply_costs_by_lot.insecticidas_usd,
            field_crop_supply_costs_by_lot.fungicidas_usd,
            field_crop_supply_costs_by_lot.coadyuvantes_usd,
            field_crop_supply_costs_by_lot.fertilizantes_usd,
            field_crop_supply_costs_by_lot.otros_insumos_usd,
            field_crop_supply_costs_by_lot.total_insumos_usd
           FROM v4_calc.field_crop_supply_costs_by_lot
        )
 SELECT project_id,
    field_id,
    crop_id AS current_crop_id,
    COALESCE(sum(semillas_usd), (0)::numeric) AS semillas_total_usd,
    COALESCE(sum(curasemillas_usd), (0)::numeric) AS curasemillas_total_usd,
    COALESCE(sum(herbicidas_usd), (0)::numeric) AS herbicidas_total_usd,
    COALESCE(sum(insecticidas_usd), (0)::numeric) AS insecticidas_total_usd,
    COALESCE(sum(fungicidas_usd), (0)::numeric) AS fungicidas_total_usd,
    COALESCE(sum(coadyuvantes_usd), (0)::numeric) AS coadyuvantes_total_usd,
    COALESCE(sum(fertilizantes_usd), (0)::numeric) AS fertilizantes_total_usd,
    COALESCE(sum(otros_insumos_usd), (0)::numeric) AS otros_insumos_total_usd,
    v4_core.safe_div(COALESCE(sum(semillas_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS semillas_usd_ha,
    v4_core.safe_div(COALESCE(sum(curasemillas_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS curasemillas_usd_ha,
    v4_core.safe_div(COALESCE(sum(herbicidas_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS herbicidas_usd_ha,
    v4_core.safe_div(COALESCE(sum(insecticidas_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS insecticidas_usd_ha,
    v4_core.safe_div(COALESCE(sum(fungicidas_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS fungicidas_usd_ha,
    v4_core.safe_div(COALESCE(sum(coadyuvantes_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS coadyuvantes_usd_ha,
    v4_core.safe_div(COALESCE(sum(fertilizantes_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS fertilizantes_usd_ha,
    v4_core.safe_div(COALESCE(sum(otros_insumos_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS otros_insumos_usd_ha,
    COALESCE(sum(total_insumos_usd), (0)::numeric) AS total_insumos_usd,
    v4_core.safe_div(COALESCE(sum(total_insumos_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS total_insumos_usd_ha
   FROM supply_costs
  GROUP BY project_id, field_id, crop_id;


--
-- Name: field_crop_labores; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.field_crop_labores AS
 SELECT project_id,
    field_id,
    crop_id AS current_crop_id,
    COALESCE(sum(siembra_usd), (0)::numeric) AS siembra_total_usd,
    COALESCE(sum(pulverizacion_usd), (0)::numeric) AS pulverizacion_total_usd,
    COALESCE(sum(riego_usd), (0)::numeric) AS riego_total_usd,
    COALESCE(sum(cosecha_usd), (0)::numeric) AS cosecha_total_usd,
    COALESCE(sum(otras_labores_usd), (0)::numeric) AS otras_labores_total_usd,
    v4_core.safe_div(COALESCE(sum(siembra_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS siembra_usd_ha,
    v4_core.safe_div(COALESCE(sum(pulverizacion_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS pulverizacion_usd_ha,
    v4_core.safe_div(COALESCE(sum(riego_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS riego_usd_ha,
    v4_core.safe_div(COALESCE(sum(cosecha_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS cosecha_usd_ha,
    v4_core.safe_div(COALESCE(sum(otras_labores_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS otras_labores_usd_ha,
    COALESCE(sum(total_labores_usd), (0)::numeric) AS total_labores_usd,
    v4_core.safe_div(COALESCE(sum(total_labores_usd), (0)::numeric), COALESCE(sum(surface_ha), (1)::numeric)) AS total_labores_usd_ha
   FROM v4_calc.field_crop_labor_costs_by_lot
  GROUP BY project_id, field_id, crop_id;


--
-- Name: field_crop_metrics; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.field_crop_metrics AS
 WITH aggregated_base AS (
         SELECT field_crop_metrics_aggregated.project_id,
            field_crop_metrics_aggregated.field_id,
            field_crop_metrics_aggregated.field_name,
            field_crop_metrics_aggregated.current_crop_id,
            field_crop_metrics_aggregated.crop_name,
            field_crop_metrics_aggregated.superficie_total,
            field_crop_metrics_aggregated.superficie_sembrada_ha,
            field_crop_metrics_aggregated.area_cosechada_ha,
            field_crop_metrics_aggregated.produccion_tn,
            field_crop_metrics_aggregated.rendimiento_tn_ha,
            field_crop_metrics_aggregated.precio_bruto_usd_tn,
            field_crop_metrics_aggregated.gasto_flete_usd_tn,
            field_crop_metrics_aggregated.gasto_comercial_usd_tn,
            field_crop_metrics_aggregated.precio_neto_usd_tn,
            field_crop_metrics_aggregated.ingreso_neto_total
           FROM v4_calc.field_crop_metrics_aggregated
        ), costs AS (
         SELECT field_crop_aggregated.project_id,
            field_crop_aggregated.field_id,
            field_crop_aggregated.crop_id,
            field_crop_aggregated.labor_costs_usd,
            field_crop_aggregated.supply_costs_usd,
            field_crop_aggregated.rent_total_usd AS arriendo_total_usd,
            field_crop_aggregated.administration_usd AS admin_total_usd
           FROM v4_calc.field_crop_aggregated
        )
 SELECT a.project_id,
    a.field_id,
    a.field_name,
    a.current_crop_id,
    a.crop_name,
    a.superficie_total AS superficie_ha,
    a.produccion_tn,
    a.superficie_sembrada_ha AS area_sembrada_ha,
    a.area_cosechada_ha,
    a.rendimiento_tn_ha,
    a.precio_bruto_usd_tn,
    a.gasto_flete_usd_tn,
    a.gasto_comercial_usd_tn,
    a.precio_neto_usd_tn,
    a.ingreso_neto_total AS ingreso_neto_usd,
    v4_core.safe_div(a.ingreso_neto_total, a.superficie_total) AS ingreso_neto_usd_ha,
    c.labor_costs_usd AS costos_labores_usd,
    v4_core.safe_div(c.labor_costs_usd, a.superficie_total) AS costos_labores_usd_ha,
    c.supply_costs_usd AS costos_insumos_usd,
    v4_core.safe_div(c.supply_costs_usd, a.superficie_total) AS costos_insumos_usd_ha,
    (c.labor_costs_usd + c.supply_costs_usd) AS total_costos_directos_usd,
    v4_core.safe_div((c.labor_costs_usd + c.supply_costs_usd), a.superficie_total) AS costos_directos_usd_ha,
    ((a.ingreso_neto_total - c.labor_costs_usd) - c.supply_costs_usd) AS margen_bruto_usd,
    v4_core.safe_div(((a.ingreso_neto_total - c.labor_costs_usd) - c.supply_costs_usd), a.superficie_total) AS margen_bruto_usd_ha,
    c.arriendo_total_usd AS arriendo_usd,
    v4_core.safe_div(c.arriendo_total_usd, a.superficie_total) AS arriendo_usd_ha,
    c.admin_total_usd AS administracion_usd,
    v4_core.safe_div(c.admin_total_usd, a.superficie_total) AS administracion_usd_ha,
    ((((a.ingreso_neto_total - c.labor_costs_usd) - c.supply_costs_usd) - c.arriendo_total_usd) - c.admin_total_usd) AS resultado_operativo_usd,
    v4_core.safe_div(((((a.ingreso_neto_total - c.labor_costs_usd) - c.supply_costs_usd) - c.arriendo_total_usd) - c.admin_total_usd), a.superficie_total) AS resultado_operativo_usd_ha,
    (((c.labor_costs_usd + c.supply_costs_usd) + c.arriendo_total_usd) + c.admin_total_usd) AS total_invertido_usd,
    v4_core.safe_div((((c.labor_costs_usd + c.supply_costs_usd) + c.arriendo_total_usd) + c.admin_total_usd), a.superficie_total) AS total_invertido_usd_ha,
        CASE
            WHEN ((((c.labor_costs_usd + c.supply_costs_usd) + c.arriendo_total_usd) + c.admin_total_usd) > (0)::numeric) THEN ((((((a.ingreso_neto_total - c.labor_costs_usd) - c.supply_costs_usd) - c.arriendo_total_usd) - c.admin_total_usd) / (((c.labor_costs_usd + c.supply_costs_usd) + c.arriendo_total_usd) + c.admin_total_usd)) * (100)::numeric)
            ELSE (0)::numeric
        END AS renta_pct,
        CASE
            WHEN ((a.precio_neto_usd_tn > (0)::numeric) AND (a.superficie_total > (0)::numeric)) THEN (((((c.labor_costs_usd + c.supply_costs_usd) + c.arriendo_total_usd) + c.admin_total_usd) / a.superficie_total) / a.precio_neto_usd_tn)
            ELSE (0)::numeric
        END AS rinde_indiferencia_usd_tn
   FROM (aggregated_base a
     LEFT JOIN costs c ON (((c.project_id = a.project_id) AND (c.field_id = a.field_id) AND (c.crop_id = a.current_crop_id))));


--
-- Name: field_crop_rentabilidad; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.field_crop_rentabilidad AS
 SELECT project_id,
    field_id,
    crop_id AS current_crop_id,
    ((direct_cost_usd + rent_fixed_usd) + administration_usd) AS total_invertido_usd,
    v4_core.safe_div(((direct_cost_usd + rent_fixed_usd) + administration_usd), surface_ha) AS total_invertido_usd_ha,
    v4_ssot.renta_pct(((((production_tn * v4_ssot.net_price_usd_for_lot(sample_lot_id)) - direct_cost_usd) - rent_fixed_usd) - administration_usd), ((direct_cost_usd + rent_fixed_usd) + administration_usd)) AS renta_pct,
    v4_core.safe_div(v4_core.safe_div(((direct_cost_usd + rent_fixed_usd) + administration_usd), surface_ha), v4_ssot.net_price_usd_for_lot(sample_lot_id)) AS rinde_indiferencia_total_usd_tn
   FROM v4_calc.field_crop_aggregated;


--
-- Name: investor_contribution_categories; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.investor_contribution_categories AS
 SELECT project_id,
    agrochemicals_total_usd,
    fertilizers_total_usd,
    seeds_total_usd,
    general_labors_total_usd,
    sowing_total_usd,
    irrigation_total_usd,
    rent_capitalizable_total_usd,
    administration_total_usd,
    total_seeded_area_ha
   FROM v4_calc.investor_contribution_categories;


--
-- Name: investor_project_base; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.investor_project_base AS
 SELECT p.id AS project_id,
    p.name AS project_name,
    p.customer_id,
    c.name AS customer_name,
    p.campaign_id,
    cam.name AS campaign_name,
    COALESCE(sum(l.hectares), (0)::numeric) AS surface_total_ha,
    COALESCE(sum((v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)), (0)::numeric) AS lease_fixed_total_usd,
        CASE
            WHEN (COALESCE(sum((v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)), (0)::numeric) > (0)::numeric) THEN true
            ELSE false
        END AS lease_is_fixed,
        CASE
            WHEN (COALESCE(sum(l.hectares), (0)::numeric) > (0)::numeric) THEN (COALESCE(sum((v4_ssot.rent_fixed_only_for_lot(l.id) * l.hectares)), (0)::numeric) / sum(l.hectares))
            ELSE (0)::numeric
        END AS lease_per_ha_usd,
    COALESCE(sum((v4_ssot.admin_cost_prorated_per_ha_for_lot(l.id) * l.hectares)), (0)::numeric) AS admin_total_usd,
        CASE
            WHEN (COALESCE(sum(l.hectares), (0)::numeric) > (0)::numeric) THEN (COALESCE(sum((v4_ssot.admin_cost_prorated_per_ha_for_lot(l.id) * l.hectares)), (0)::numeric) / sum(l.hectares))
            ELSE (0)::numeric
        END AS admin_per_ha_usd
   FROM ((((public.projects p
     JOIN public.customers c ON (((p.customer_id = c.id) AND (c.deleted_at IS NULL))))
     JOIN public.campaigns cam ON (((p.campaign_id = cam.id) AND (cam.deleted_at IS NULL))))
     LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.id, p.name, p.customer_id, c.name, p.campaign_id, cam.name;


--
-- Name: investor_contribution_data; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.investor_contribution_data AS
 WITH investor_base AS (
         SELECT pi.project_id,
            pi.investor_id,
            i.name AS investor_name,
            pi.percentage AS share_pct_agreed
           FROM (public.project_investors pi
             JOIN public.investors i ON (((i.id = pi.investor_id) AND (i.deleted_at IS NULL))))
          WHERE (pi.deleted_at IS NULL)
        ), project_surface_data AS (
         SELECT investor_project_base.project_id,
            max(investor_project_base.surface_total_ha) AS surface_total_ha
           FROM v4_report.investor_project_base
          GROUP BY investor_project_base.project_id
        ), investor_harvest_real AS (
         SELECT w.project_id,
            w.investor_id,
            COALESCE(sum((lab.price * w.effective_area)), (0)::numeric) AS harvest_real_usd
           FROM ((public.workorders w
             JOIN public.labors lab ON (((w.labor_id = lab.id) AND (lab.deleted_at IS NULL))))
             JOIN public.categories cat ON ((cat.id = lab.category_id)))
          WHERE ((w.deleted_at IS NULL) AND (cat.type_id = 4) AND ((cat.name)::text = 'Cosecha'::text))
          GROUP BY w.project_id, w.investor_id
        ), harvest_totals AS (
         SELECT psd.project_id,
            COALESCE(sum(hr.harvest_real_usd), (0)::numeric) AS total_harvest_usd,
                CASE
                    WHEN (COALESCE(psd.surface_total_ha, (0)::numeric) > (0)::numeric) THEN (COALESCE(sum(hr.harvest_real_usd), (0)::numeric) / psd.surface_total_ha)
                    ELSE (0)::numeric
                END AS total_harvest_usd_ha
           FROM (project_surface_data psd
             LEFT JOIN investor_harvest_real hr ON ((hr.project_id = psd.project_id)))
          GROUP BY psd.project_id, psd.surface_total_ha
        )
 SELECT pb.project_id,
    pb.project_name,
    pb.customer_id,
    pb.customer_name,
    pb.campaign_id,
    pb.campaign_name,
    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'share_pct', (irc.share_pct_agreed)::numeric) ORDER BY irc.investor_id) AS jsonb_agg
           FROM v4_calc.investor_real_contributions irc
          WHERE (irc.project_id = pb.project_id)) AS investor_headers,
    jsonb_build_object('surface_total_ha', pb.surface_total_ha, 'lease_fixed_total_usd', pb.lease_fixed_total_usd, 'lease_is_fixed', pb.lease_is_fixed, 'lease_per_ha_usd', pb.lease_per_ha_usd, 'admin_total_usd', pb.admin_total_usd, 'admin_per_ha_usd', pb.admin_per_ha_usd) AS general_project_data,
    ( SELECT jsonb_agg(jsonb_build_object('key', cat.key, 'sort_index', cat.sort_index, 'type', cat.type, 'label', cat.label, 'total_usd', cat.total_usd, 'total_usd_ha', cat.total_usd_ha, 'investors', cat.investors, 'requires_manual_attribution', cat.requires_manual_attribution, 'attribution_note', cat.attribution_note) ORDER BY cat.sort_index) AS jsonb_agg
           FROM ( SELECT 'agrochemicals'::text AS key,
                    1 AS sort_index,
                    'pre_harvest'::text AS type,
                    'Agroquímicos'::text AS label,
                    cc_1.agrochemicals_total_usd AS total_usd,
                        CASE
                            WHEN (pb.surface_total_ha > (0)::numeric) THEN (cc_1.agrochemicals_total_usd / pb.surface_total_ha)
                            ELSE (0)::numeric
                        END AS total_usd_ha,
                    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'amount_usd', irc.agrochemicals_real_usd, 'share_pct',
                                CASE
                                    WHEN (cc_1.agrochemicals_total_usd > (0)::numeric) THEN ((irc.agrochemicals_real_usd / cc_1.agrochemicals_total_usd) * (100)::numeric)
                                    ELSE (0)::numeric
                                END) ORDER BY irc.investor_id) AS jsonb_agg
                           FROM v4_calc.investor_real_contributions irc
                          WHERE (irc.project_id = pb.project_id)) AS investors,
                    false AS requires_manual_attribution,
                    NULL::text AS attribution_note
                   FROM v4_report.investor_contribution_categories cc_1
                  WHERE (cc_1.project_id = pb.project_id)
                UNION ALL
                 SELECT 'fertilizers'::text AS text,
                    2,
                    'pre_harvest'::text AS text,
                    'Fertilizantes'::text AS text,
                    cc_1.fertilizers_total_usd,
                        CASE
                            WHEN (pb.surface_total_ha > (0)::numeric) THEN (cc_1.fertilizers_total_usd / pb.surface_total_ha)
                            ELSE (0)::numeric
                        END AS "case",
                    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'amount_usd', irc.fertilizers_real_usd, 'share_pct',
                                CASE
                                    WHEN (cc_1.fertilizers_total_usd > (0)::numeric) THEN ((irc.fertilizers_real_usd / cc_1.fertilizers_total_usd) * (100)::numeric)
                                    ELSE (0)::numeric
                                END) ORDER BY irc.investor_id) AS jsonb_agg
                           FROM v4_calc.investor_real_contributions irc
                          WHERE (irc.project_id = pb.project_id)) AS jsonb_agg,
                    false,
                    NULL::text
                   FROM v4_report.investor_contribution_categories cc_1
                  WHERE (cc_1.project_id = pb.project_id)
                UNION ALL
                 SELECT 'seeds'::text AS text,
                    3,
                    'pre_harvest'::text AS text,
                    'Semilla'::text AS text,
                    cc_1.seeds_total_usd,
                        CASE
                            WHEN (pb.surface_total_ha > (0)::numeric) THEN (cc_1.seeds_total_usd / pb.surface_total_ha)
                            ELSE (0)::numeric
                        END AS "case",
                    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'amount_usd', irc.seeds_real_usd, 'share_pct',
                                CASE
                                    WHEN (cc_1.seeds_total_usd > (0)::numeric) THEN ((irc.seeds_real_usd / cc_1.seeds_total_usd) * (100)::numeric)
                                    ELSE (0)::numeric
                                END) ORDER BY irc.investor_id) AS jsonb_agg
                           FROM v4_calc.investor_real_contributions irc
                          WHERE (irc.project_id = pb.project_id)) AS jsonb_agg,
                    false,
                    NULL::text
                   FROM v4_report.investor_contribution_categories cc_1
                  WHERE (cc_1.project_id = pb.project_id)
                UNION ALL
                 SELECT 'general_labors'::text AS text,
                    4,
                    'pre_harvest'::text AS text,
                    'Labores Generales'::text AS text,
                    cc_1.general_labors_total_usd,
                        CASE
                            WHEN (pb.surface_total_ha > (0)::numeric) THEN (cc_1.general_labors_total_usd / pb.surface_total_ha)
                            ELSE (0)::numeric
                        END AS "case",
                    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'amount_usd', irc.general_labors_real_usd, 'share_pct',
                                CASE
                                    WHEN (cc_1.general_labors_total_usd > (0)::numeric) THEN ((irc.general_labors_real_usd / cc_1.general_labors_total_usd) * (100)::numeric)
                                    ELSE (0)::numeric
                                END) ORDER BY irc.investor_id) AS jsonb_agg
                           FROM v4_calc.investor_real_contributions irc
                          WHERE (irc.project_id = pb.project_id)) AS jsonb_agg,
                    false,
                    NULL::text
                   FROM v4_report.investor_contribution_categories cc_1
                  WHERE (cc_1.project_id = pb.project_id)
                UNION ALL
                 SELECT 'sowing'::text AS text,
                    5,
                    'pre_harvest'::text AS text,
                    'Siembra'::text AS text,
                    cc_1.sowing_total_usd,
                        CASE
                            WHEN (pb.surface_total_ha > (0)::numeric) THEN (cc_1.sowing_total_usd / pb.surface_total_ha)
                            ELSE (0)::numeric
                        END AS "case",
                    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'amount_usd', irc.sowing_real_usd, 'share_pct',
                                CASE
                                    WHEN (cc_1.sowing_total_usd > (0)::numeric) THEN ((irc.sowing_real_usd / cc_1.sowing_total_usd) * (100)::numeric)
                                    ELSE (0)::numeric
                                END) ORDER BY irc.investor_id) AS jsonb_agg
                           FROM v4_calc.investor_real_contributions irc
                          WHERE (irc.project_id = pb.project_id)) AS jsonb_agg,
                    false,
                    NULL::text
                   FROM v4_report.investor_contribution_categories cc_1
                  WHERE (cc_1.project_id = pb.project_id)
                UNION ALL
                 SELECT 'irrigation'::text AS text,
                    6,
                    'pre_harvest'::text AS text,
                    'Riego'::text AS text,
                    cc_1.irrigation_total_usd,
                        CASE
                            WHEN (pb.surface_total_ha > (0)::numeric) THEN (cc_1.irrigation_total_usd / pb.surface_total_ha)
                            ELSE (0)::numeric
                        END AS "case",
                    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'amount_usd', irc.irrigation_real_usd, 'share_pct',
                                CASE
                                    WHEN (cc_1.irrigation_total_usd > (0)::numeric) THEN ((irc.irrigation_real_usd / cc_1.irrigation_total_usd) * (100)::numeric)
                                    ELSE (0)::numeric
                                END) ORDER BY irc.investor_id) AS jsonb_agg
                           FROM v4_calc.investor_real_contributions irc
                          WHERE (irc.project_id = pb.project_id)) AS jsonb_agg,
                    false,
                    NULL::text
                   FROM v4_report.investor_contribution_categories cc_1
                  WHERE (cc_1.project_id = pb.project_id)
                UNION ALL
                 SELECT 'capitalizable_lease'::text AS text,
                    7,
                    'pre_harvest'::text AS text,
                    'Arriendo Capitalizable'::text AS text,
                    cc_1.rent_capitalizable_total_usd,
                        CASE
                            WHEN (pb.surface_total_ha > (0)::numeric) THEN (cc_1.rent_capitalizable_total_usd / pb.surface_total_ha)
                            ELSE (0)::numeric
                        END AS "case",
                    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'amount_usd', irc.rent_real_usd, 'share_pct',
                                CASE
                                    WHEN (cc_1.rent_capitalizable_total_usd > (0)::numeric) THEN ((irc.rent_real_usd / cc_1.rent_capitalizable_total_usd) * (100)::numeric)
                                    ELSE (0)::numeric
                                END) ORDER BY irc.investor_id) AS jsonb_agg
                           FROM v4_calc.investor_real_contributions irc
                          WHERE (irc.project_id = pb.project_id)) AS jsonb_agg,
                    true,
                    'Requiere atribución manual por inversor'::text
                   FROM v4_report.investor_contribution_categories cc_1
                  WHERE (cc_1.project_id = pb.project_id)
                UNION ALL
                 SELECT 'administration_structure'::text AS text,
                    8,
                    'pre_harvest'::text AS text,
                    'Administración y Estructura'::text AS text,
                    cc_1.administration_total_usd,
                        CASE
                            WHEN (pb.surface_total_ha > (0)::numeric) THEN (cc_1.administration_total_usd / pb.surface_total_ha)
                            ELSE (0)::numeric
                        END AS "case",
                    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'amount_usd', irc.administration_real_usd, 'share_pct',
                                CASE
                                    WHEN (cc_1.administration_total_usd > (0)::numeric) THEN ((irc.administration_real_usd / cc_1.administration_total_usd) * (100)::numeric)
                                    ELSE (0)::numeric
                                END) ORDER BY irc.investor_id) AS jsonb_agg
                           FROM v4_calc.investor_real_contributions irc
                          WHERE (irc.project_id = pb.project_id)) AS jsonb_agg,
                    true,
                    'Requiere atribución manual por inversor'::text
                   FROM v4_report.investor_contribution_categories cc_1
                  WHERE (cc_1.project_id = pb.project_id)) cat) AS contribution_categories,
    ( SELECT jsonb_agg(jsonb_build_object('investor_id', irc.investor_id, 'investor_name', irc.investor_name, 'agreed_share_pct', irc.share_pct_agreed, 'agreed_usd', ((irc.project_total_contributions_usd * (irc.share_pct_agreed)::numeric) / (100)::numeric), 'actual_usd', irc.total_real_contribution_usd, 'share_pct', irc.contributions_progress_pct, 'adjustment_usd', (irc.total_real_contribution_usd - ((irc.project_total_contributions_usd * (irc.share_pct_agreed)::numeric) / (100)::numeric))) ORDER BY irc.investor_id) AS jsonb_agg
           FROM v4_calc.investor_real_contributions irc
          WHERE (irc.project_id = pb.project_id)) AS investor_contribution_comparison,
    jsonb_build_object('rows', jsonb_build_array(jsonb_build_object('key', 'harvest', 'type', 'harvest', 'total_usd', COALESCE(ht.total_harvest_usd, (0)::numeric), 'total_us_ha', COALESCE(ht.total_harvest_usd_ha, (0)::numeric), 'investors', COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ib.investor_id, 'investor_name', ib.investor_name, 'amount_usd', COALESCE(hr.harvest_real_usd, (0)::numeric), 'share_pct',
                CASE
                    WHEN (COALESCE(ht.total_harvest_usd, (0)::numeric) > (0)::numeric) THEN ((COALESCE(hr.harvest_real_usd, (0)::numeric) / COALESCE(ht.total_harvest_usd, (0)::numeric)) * (100)::numeric)
                    ELSE (0)::numeric
                END) ORDER BY ib.investor_id) AS jsonb_agg
           FROM (investor_base ib
             LEFT JOIN investor_harvest_real hr ON (((hr.project_id = ib.project_id) AND (hr.investor_id = ib.investor_id))))
          WHERE (ib.project_id = pb.project_id)), '[]'::jsonb)), jsonb_build_object('key', 'totals', 'type', 'totals', 'total_usd', COALESCE(ht.total_harvest_usd, (0)::numeric), 'total_us_ha', COALESCE(ht.total_harvest_usd_ha, (0)::numeric), 'investors', COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ib.investor_id, 'investor_name', ib.investor_name, 'amount_usd', COALESCE(hr.harvest_real_usd, (0)::numeric), 'share_pct',
                CASE
                    WHEN (COALESCE(ht.total_harvest_usd, (0)::numeric) > (0)::numeric) THEN ((COALESCE(hr.harvest_real_usd, (0)::numeric) / COALESCE(ht.total_harvest_usd, (0)::numeric)) * (100)::numeric)
                    ELSE (0)::numeric
                END) ORDER BY ib.investor_id) AS jsonb_agg
           FROM (investor_base ib
             LEFT JOIN investor_harvest_real hr ON (((hr.project_id = ib.project_id) AND (hr.investor_id = ib.investor_id))))
          WHERE (ib.project_id = pb.project_id)), '[]'::jsonb))), 'footer_payment_agreed', COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ib.investor_id, 'investor_name', ib.investor_name, 'amount_usd', ((COALESCE(ht.total_harvest_usd, (0)::numeric) * (ib.share_pct_agreed)::numeric) / (100)::numeric), 'share_pct', (ib.share_pct_agreed)::numeric) ORDER BY ib.investor_id) AS jsonb_agg
           FROM investor_base ib
          WHERE (ib.project_id = pb.project_id)), '[]'::jsonb), 'footer_payment_adjustment', COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ib.investor_id, 'investor_name', ib.investor_name, 'amount_usd', (COALESCE(hr.harvest_real_usd, (0)::numeric) - ((COALESCE(ht.total_harvest_usd, (0)::numeric) * (ib.share_pct_agreed)::numeric) / (100)::numeric)), 'share_pct', (ib.share_pct_agreed)::numeric) ORDER BY ib.investor_id) AS jsonb_agg
           FROM (investor_base ib
             LEFT JOIN investor_harvest_real hr ON (((hr.project_id = ib.project_id) AND (hr.investor_id = ib.investor_id))))
          WHERE (ib.project_id = pb.project_id)), '[]'::jsonb)) AS harvest_settlement
   FROM ((v4_report.investor_project_base pb
     JOIN v4_report.investor_contribution_categories cc ON ((cc.project_id = pb.project_id)))
     LEFT JOIN harvest_totals ht ON ((ht.project_id = pb.project_id)))
  ORDER BY pb.project_id;


--
-- Name: investor_distributions; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.investor_distributions AS
 SELECT project_id,
    investor_id,
    investor_name,
    share_pct_agreed,
    ((project_total_contributions_usd * (share_pct_agreed)::numeric) / (100)::numeric) AS agreed_contribution_usd,
    total_real_contribution_usd AS real_contribution_usd,
    (total_real_contribution_usd - ((project_total_contributions_usd * (share_pct_agreed)::numeric) / (100)::numeric)) AS adjustment_usd,
    agrochemicals_real_usd,
    fertilizers_real_usd,
    seeds_real_usd,
    general_labors_real_usd,
    sowing_real_usd,
    irrigation_real_usd,
    rent_real_usd,
    administration_real_usd,
    project_total_contributions_usd
   FROM v4_calc.investor_real_contributions irc
  ORDER BY project_id, investor_id;


--
-- Name: labor_list; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.labor_list AS
 SELECT w.id AS workorder_id,
    w.number AS workorder_number,
    w.date,
    w.project_id,
    p.name AS project_name,
    w.field_id,
    f.name AS field_name,
    w.lot_id,
    l.name AS lot_name,
    w.crop_id,
    c.name AS crop_name,
    w.labor_id,
    lb.name AS labor_name,
    lb.category_id AS labor_category_id,
    cat.name AS labor_category_name,
    w.contractor,
    lb.contractor_name,
    w.effective_area AS surface_ha,
    lb.price AS cost_per_ha,
    (lb.price * w.effective_area) AS total_labor_cost,
    v4_core.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month,
    (lb.price)::numeric AS usd_cost_ha,
    (lb.price * w.effective_area) AS usd_net_total,
    w.investor_id,
    i.name AS investor_name
   FROM (((((((public.workorders w
     JOIN public.projects p ON (((p.id = w.project_id) AND (p.deleted_at IS NULL))))
     JOIN public.fields f ON (((f.id = w.field_id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.id = w.lot_id) AND (l.deleted_at IS NULL))))
     LEFT JOIN public.crops c ON (((c.id = w.crop_id) AND (c.deleted_at IS NULL))))
     JOIN public.labors lb ON (((lb.id = w.labor_id) AND (lb.deleted_at IS NULL))))
     LEFT JOIN public.categories cat ON (((cat.id = lb.category_id) AND (cat.deleted_at IS NULL))))
     LEFT JOIN public.investors i ON (((i.id = w.investor_id) AND (i.deleted_at IS NULL))))
  WHERE ((w.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric) AND (lb.price IS NOT NULL));


--
-- Name: labor_metrics; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.labor_metrics AS
 WITH wo AS (
         SELECT w.id AS workorder_id,
            w.project_id,
            w.field_id,
            w.date,
            (w.effective_area)::numeric AS effective_area,
            (lb.price)::numeric AS labor_price_per_ha
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric) AND (lb.price IS NOT NULL))
        ), agg AS (
         SELECT wo.project_id,
            wo.field_id,
            count(DISTINCT wo.workorder_id) AS total_workorders,
            sum(wo.effective_area) AS surface_ha,
            sum(v4_core.labor_cost(wo.labor_price_per_ha, wo.effective_area)) AS total_labor_cost,
            min(wo.date) AS first_workorder_date,
            max(wo.date) AS last_workorder_date
           FROM wo
          GROUP BY wo.project_id, wo.field_id
        )
 SELECT project_id,
    field_id,
    surface_ha,
    total_labor_cost,
    v4_core.cost_per_ha(total_labor_cost, surface_ha) AS avg_labor_cost_per_ha,
    total_workorders,
    first_workorder_date,
    last_workorder_date
   FROM agg a;


--
-- Name: lot_list; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.lot_list AS
 SELECT f.project_id,
    p.name AS project_name,
    f.id AS field_id,
    f.name AS field_name,
    l.id,
    l.name AS lot_name,
    l.variety,
    l.season,
    l.previous_crop_id,
    prev_crop.name AS previous_crop,
    l.current_crop_id,
    curr_crop.name AS current_crop,
    l.hectares,
    l.updated_at,
    COALESCE(lm.seeded_area_ha, (0)::numeric) AS seeded_area_ha,
    COALESCE(lm.harvested_area_ha, (0)::numeric) AS harvested_area_ha,
    COALESCE(lm.yield_tn_per_ha, (0)::numeric) AS yield_tn_per_ha,
    COALESCE(lm.direct_cost_per_ha_usd, (0)::numeric) AS cost_usd_per_ha,
    COALESCE(lm.income_net_per_ha_usd, (0)::numeric) AS income_net_per_ha_usd,
    COALESCE(lm.rent_per_ha_usd, (0)::numeric) AS rent_per_ha_usd,
    COALESCE(lm.admin_cost_per_ha_usd, (0)::numeric) AS admin_cost_per_ha_usd,
    COALESCE(lm.active_total_per_ha_usd, (0)::numeric) AS active_total_per_ha_usd,
    COALESCE(lm.operating_result_per_ha_usd, (0)::numeric) AS operating_result_per_ha_usd,
    COALESCE(lm.income_net_total_usd, (0)::numeric) AS income_net_total_usd,
    COALESCE(lm.direct_cost_total_usd, (0)::numeric) AS direct_cost_total_usd,
    COALESCE(lm.rent_total_usd, (0)::numeric) AS rent_total_usd,
    COALESCE(lm.admin_total_usd, (0)::numeric) AS admin_total_usd,
    COALESCE(lm.active_total_usd, (0)::numeric) AS active_total_usd,
    COALESCE(lm.operating_result_total_usd, (0)::numeric) AS operating_result_total_usd,
    l.sowing_date AS lot_sowing_date,
    NULL::date AS lot_harvest_date,
    l.tons,
    ( SELECT min(w.date) AS min
           FROM (public.workorders w
             JOIN public.labors lb ON (((lb.id = w.labor_id) AND (lb.deleted_at IS NULL))))
          WHERE ((w.lot_id = l.id) AND (w.deleted_at IS NULL))) AS raw_sowing_date,
    COALESCE(lm.seeded_area_ha, (0)::numeric) AS sowed_area_ha,
    COALESCE(lm.seeded_area_ha, (0)::numeric) AS sown_area_ha
   FROM (((((public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
     JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
     LEFT JOIN public.crops prev_crop ON (((prev_crop.id = l.previous_crop_id) AND (prev_crop.deleted_at IS NULL))))
     LEFT JOIN public.crops curr_crop ON (((curr_crop.id = l.current_crop_id) AND (curr_crop.deleted_at IS NULL))))
     LEFT JOIN v4_report.lot_metrics lm ON ((lm.lot_id = l.id)))
  WHERE (l.deleted_at IS NULL);


--
-- Name: stock_consumed_by_supply; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.stock_consumed_by_supply AS
 SELECT wo.project_id,
    woi.supply_id,
    COALESCE(sum(woi.total_used), (0)::numeric) AS consumed
   FROM (public.workorder_items woi
     JOIN public.workorders wo ON ((wo.id = woi.workorder_id)))
  WHERE ((wo.deleted_at IS NULL) AND (woi.deleted_at IS NULL) AND (woi.supply_id IS NOT NULL))
  GROUP BY wo.project_id, woi.supply_id;


--
-- Name: summary_results; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.summary_results AS
 WITH by_crop AS (
         SELECT field_crop_metrics.project_id,
            field_crop_metrics.current_crop_id,
            field_crop_metrics.crop_name,
            sum(field_crop_metrics.superficie_ha) AS surface_ha,
            sum(field_crop_metrics.ingreso_neto_usd) AS net_income_usd,
            sum(field_crop_metrics.total_costos_directos_usd) AS direct_costs_usd,
            sum(field_crop_metrics.arriendo_usd) AS rent_usd,
            sum(field_crop_metrics.administracion_usd) AS structure_usd,
            sum(field_crop_metrics.total_invertido_usd) AS total_invested_usd,
            sum(field_crop_metrics.resultado_operativo_usd) AS operating_result_usd
           FROM v4_report.field_crop_metrics
          WHERE (field_crop_metrics.current_crop_id IS NOT NULL)
          GROUP BY field_crop_metrics.project_id, field_crop_metrics.current_crop_id, field_crop_metrics.crop_name
        ), project_totals AS (
         SELECT by_crop.project_id,
            sum(by_crop.surface_ha) AS total_surface_ha,
            sum(by_crop.net_income_usd) AS total_net_income_usd,
            sum(by_crop.direct_costs_usd) AS total_direct_costs_usd,
            sum(by_crop.rent_usd) AS total_rent_usd,
            sum(by_crop.structure_usd) AS total_structure_usd,
            sum(by_crop.total_invested_usd) AS total_invested_project_usd,
            sum(by_crop.operating_result_usd) AS total_operating_result_usd
           FROM by_crop
          GROUP BY by_crop.project_id
        )
 SELECT bc.project_id,
    bc.current_crop_id,
    bc.crop_name,
    bc.surface_ha,
    bc.net_income_usd,
    bc.direct_costs_usd,
    bc.rent_usd,
    bc.structure_usd,
    bc.total_invested_usd,
    bc.operating_result_usd,
        CASE
            WHEN (bc.total_invested_usd > (0)::numeric) THEN ((bc.operating_result_usd / bc.total_invested_usd) * (100)::numeric)
            ELSE (0)::numeric
        END AS crop_return_pct,
    pt.total_surface_ha,
    pt.total_net_income_usd,
    pt.total_direct_costs_usd,
    pt.total_rent_usd,
    pt.total_structure_usd,
    pt.total_invested_project_usd,
    pt.total_operating_result_usd,
        CASE
            WHEN (pt.total_invested_project_usd > (0)::numeric) THEN ((pt.total_operating_result_usd / pt.total_invested_project_usd) * (100)::numeric)
            ELSE (0)::numeric
        END AS project_return_pct
   FROM (by_crop bc
     JOIN project_totals pt ON ((pt.project_id = bc.project_id)));


--
-- Name: workorder_list; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.workorder_list AS
 SELECT w.id,
    w.number,
    w.project_id,
    w.field_id,
    p.name AS project_name,
    f.name AS field_name,
    l.name AS lot_name,
    w.date,
    c.name AS crop_name,
    lb.name AS labor_name,
    cat_lb.name AS labor_category_name,
    t.name AS type_name,
    w.contractor,
    w.effective_area AS surface_ha,
    s.name AS supply_name,
    wi.total_used AS consumption,
    cat.name AS category_name,
    wi.final_dose AS dose_per_ha,
    s.price AS unit_price,
        CASE
            WHEN ((wi.final_dose IS NOT NULL) AND (s.price IS NOT NULL)) THEN v4_core.cost_per_ha(((wi.final_dose)::numeric * s.price), (1)::numeric)
            ELSE (0)::numeric
        END AS supply_cost_per_ha,
    v4_core.supply_cost((wi.final_dose)::numeric, (s.price)::numeric, (w.effective_area)::numeric) AS supply_total_cost
   FROM ((((((((((public.workorders w
     JOIN public.projects p ON (((p.id = w.project_id) AND (p.deleted_at IS NULL))))
     JOIN public.fields f ON (((f.id = w.field_id) AND (f.deleted_at IS NULL))))
     JOIN public.lots l ON (((l.id = w.lot_id) AND (l.deleted_at IS NULL))))
     JOIN public.crops c ON (((c.id = w.crop_id) AND (c.deleted_at IS NULL))))
     JOIN public.labors lb ON (((lb.id = w.labor_id) AND (lb.deleted_at IS NULL))))
     JOIN public.categories cat_lb ON (((cat_lb.id = lb.category_id) AND (cat_lb.deleted_at IS NULL))))
     LEFT JOIN public.workorder_items wi ON (((wi.workorder_id = w.id) AND (wi.deleted_at IS NULL))))
     LEFT JOIN public.supplies s ON (((s.id = wi.supply_id) AND (s.deleted_at IS NULL))))
     LEFT JOIN public.types t ON (((t.id = s.type_id) AND (t.deleted_at IS NULL))))
     LEFT JOIN public.categories cat ON (((cat.id = s.category_id) AND (cat.deleted_at IS NULL))))
  WHERE (w.deleted_at IS NULL);


--
-- Name: VIEW workorder_list; Type: COMMENT; Schema: v4_report; Owner: -
--

COMMENT ON VIEW v4_report.workorder_list IS 'Listado detallado de workorders usando funciones v4_core';


--
-- Name: workorder_metrics; Type: VIEW; Schema: v4_report; Owner: -
--

CREATE VIEW v4_report.workorder_metrics AS
 SELECT project_id,
    field_id,
    lot_id,
    surface_ha,
    liters,
    kilograms,
    labor_cost_usd,
    supplies_cost_usd,
    direct_cost_usd,
    avg_cost_per_ha_usd,
    liters_per_ha,
    kilograms_per_ha
   FROM v4_calc.workorder_metrics;


--
-- Name: business_parameters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_parameters ALTER COLUMN id SET DEFAULT nextval('public.business_parameters_id_seq'::regclass);


--
-- Name: campaigns id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.campaigns ALTER COLUMN id SET DEFAULT nextval('public.campaigns_id_seq'::regclass);


--
-- Name: categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: crop_commercializations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crop_commercializations ALTER COLUMN id SET DEFAULT nextval('public.crop_commercializations_id_seq'::regclass);


--
-- Name: crops id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crops ALTER COLUMN id SET DEFAULT nextval('public.crops_id_seq'::regclass);


--
-- Name: customers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers ALTER COLUMN id SET DEFAULT nextval('public.customers_id_seq'::regclass);


--
-- Name: fields id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields ALTER COLUMN id SET DEFAULT nextval('public.fields_id_seq'::regclass);


--
-- Name: fx_rates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fx_rates ALTER COLUMN id SET DEFAULT nextval('public.fx_rates_id_seq'::regclass);


--
-- Name: investors id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.investors ALTER COLUMN id SET DEFAULT nextval('public.investors_id_seq'::regclass);


--
-- Name: invoices id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices ALTER COLUMN id SET DEFAULT nextval('public.invoices_id_seq'::regclass);


--
-- Name: labor_categories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_categories ALTER COLUMN id SET DEFAULT nextval('public.labor_categories_id_seq'::regclass);


--
-- Name: labor_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_types ALTER COLUMN id SET DEFAULT nextval('public.labor_types_id_seq'::regclass);


--
-- Name: labors id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labors ALTER COLUMN id SET DEFAULT nextval('public.labors_id_seq'::regclass);


--
-- Name: lease_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lease_types ALTER COLUMN id SET DEFAULT nextval('public.lease_types_id_seq'::regclass);


--
-- Name: lot_dates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates ALTER COLUMN id SET DEFAULT nextval('public.lot_dates_id_seq'::regclass);


--
-- Name: lots id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots ALTER COLUMN id SET DEFAULT nextval('public.lots_id_seq'::regclass);


--
-- Name: managers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.managers ALTER COLUMN id SET DEFAULT nextval('public.managers_id_seq'::regclass);


--
-- Name: project_dollar_values id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_dollar_values ALTER COLUMN id SET DEFAULT nextval('public.project_dollar_values_id_seq'::regclass);


--
-- Name: projects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);


--
-- Name: providers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers ALTER COLUMN id SET DEFAULT nextval('public.providers_id_seq'::regclass);


--
-- Name: stocks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks ALTER COLUMN id SET DEFAULT nextval('public.stocks_id_seq'::regclass);


--
-- Name: supplies id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supplies ALTER COLUMN id SET DEFAULT nextval('public.supplies_id_seq'::regclass);


--
-- Name: supply_movements id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements ALTER COLUMN id SET DEFAULT nextval('public.supply_movements_id_seq'::regclass);


--
-- Name: types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.types ALTER COLUMN id SET DEFAULT nextval('public.types_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: workorder_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorder_items ALTER COLUMN id SET DEFAULT nextval('public.workorder_items_id_seq'::regclass);


--
-- Name: workorders id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders ALTER COLUMN id SET DEFAULT nextval('public.workorders_id_seq'::regclass);


--
-- Name: admin_cost_investors pk_admin_cost_investors; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT pk_admin_cost_investors PRIMARY KEY (project_id, investor_id);


--
-- Name: business_parameters pk_business_parameters; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_parameters
    ADD CONSTRAINT pk_business_parameters PRIMARY KEY (id);


--
-- Name: campaigns pk_campaigns; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT pk_campaigns PRIMARY KEY (id);


--
-- Name: categories pk_categories; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT pk_categories PRIMARY KEY (id);


--
-- Name: crop_commercializations pk_crop_commercializations; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT pk_crop_commercializations PRIMARY KEY (id);


--
-- Name: crops pk_crops; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT pk_crops PRIMARY KEY (id);


--
-- Name: customers pk_customers; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT pk_customers PRIMARY KEY (id);


--
-- Name: field_investors pk_field_investors; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT pk_field_investors PRIMARY KEY (field_id, investor_id);


--
-- Name: fields pk_fields; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT pk_fields PRIMARY KEY (id);


--
-- Name: fx_rates pk_fx_rates; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fx_rates
    ADD CONSTRAINT pk_fx_rates PRIMARY KEY (id);


--
-- Name: investors pk_investors; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT pk_investors PRIMARY KEY (id);


--
-- Name: invoices pk_invoices; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT pk_invoices PRIMARY KEY (id);


--
-- Name: labor_categories pk_labor_categories; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_categories
    ADD CONSTRAINT pk_labor_categories PRIMARY KEY (id);


--
-- Name: labor_types pk_labor_types; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_types
    ADD CONSTRAINT pk_labor_types PRIMARY KEY (id);


--
-- Name: labors pk_labors; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT pk_labors PRIMARY KEY (id);


--
-- Name: lease_types pk_lease_types; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT pk_lease_types PRIMARY KEY (id);


--
-- Name: lot_dates pk_lot_dates; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT pk_lot_dates PRIMARY KEY (id);


--
-- Name: lots pk_lots; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT pk_lots PRIMARY KEY (id);


--
-- Name: managers pk_managers; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT pk_managers PRIMARY KEY (id);


--
-- Name: project_dollar_values pk_project_dollar_values; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT pk_project_dollar_values PRIMARY KEY (id);


--
-- Name: project_investors pk_project_investors; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT pk_project_investors PRIMARY KEY (project_id, investor_id);


--
-- Name: project_managers pk_project_managers; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT pk_project_managers PRIMARY KEY (project_id, manager_id);


--
-- Name: projects pk_projects; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT pk_projects PRIMARY KEY (id);


--
-- Name: providers pk_providers; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT pk_providers PRIMARY KEY (id);


--
-- Name: stocks pk_stocks; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT pk_stocks PRIMARY KEY (id);


--
-- Name: supplies pk_supplies; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT pk_supplies PRIMARY KEY (id);


--
-- Name: supply_movements pk_supply_movements; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT pk_supply_movements PRIMARY KEY (id);


--
-- Name: types pk_types; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.types
    ADD CONSTRAINT pk_types PRIMARY KEY (id);


--
-- Name: users pk_users; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT pk_users PRIMARY KEY (id);


--
-- Name: workorder_items pk_workorder_items; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT pk_workorder_items PRIMARY KEY (id);


--
-- Name: workorders pk_workorders; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT pk_workorders PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: business_parameters uq_business_parameters_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_parameters
    ADD CONSTRAINT uq_business_parameters_key UNIQUE (key);


--
-- Name: campaigns uq_campaigns_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT uq_campaigns_name UNIQUE (name);


--
-- Name: crops uq_crops_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT uq_crops_name UNIQUE (name);


--
-- Name: customers uq_customers_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT uq_customers_name UNIQUE (name);


--
-- Name: fx_rates uq_fx_rates_pair_date; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fx_rates
    ADD CONSTRAINT uq_fx_rates_pair_date UNIQUE (currency_pair, effective_date);


--
-- Name: investors uq_investors_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT uq_investors_name UNIQUE (name);


--
-- Name: invoices uq_invoices_work_order; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT uq_invoices_work_order UNIQUE (work_order_id);


--
-- Name: labor_types uq_labor_types_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_types
    ADD CONSTRAINT uq_labor_types_name UNIQUE (name);


--
-- Name: lease_types uq_lease_types_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT uq_lease_types_name UNIQUE (name);


--
-- Name: managers uq_managers_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT uq_managers_name UNIQUE (name);


--
-- Name: project_dollar_values uq_project_dollar_values_period; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT uq_project_dollar_values_period UNIQUE (project_id, year, month);


--
-- Name: providers uq_providers_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT uq_providers_name UNIQUE (name);


--
-- Name: types uq_types_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.types
    ADD CONSTRAINT uq_types_name UNIQUE (name);


--
-- Name: users uq_users_username; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT uq_users_username UNIQUE (username);


--
-- Name: idx_admin_cost_investors_investor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_admin_cost_investors_investor_id ON public.admin_cost_investors USING btree (investor_id);


--
-- Name: idx_business_parameters_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_business_parameters_category ON public.business_parameters USING btree (category);


--
-- Name: idx_business_parameters_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_business_parameters_key ON public.business_parameters USING btree (key);


--
-- Name: idx_campaigns_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_campaigns_created_by ON public.campaigns USING btree (created_by);


--
-- Name: idx_campaigns_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_campaigns_deleted_by ON public.campaigns USING btree (deleted_by);


--
-- Name: idx_campaigns_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_campaigns_updated_by ON public.campaigns USING btree (updated_by);


--
-- Name: idx_categories_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_categories_type_id ON public.categories USING btree (type_id);


--
-- Name: idx_crop_commercializations_crop_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crop_commercializations_crop_id ON public.crop_commercializations USING btree (crop_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_crop_commercializations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crop_commercializations_deleted_at ON public.crop_commercializations USING btree (deleted_at);


--
-- Name: idx_crop_commercializations_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crop_commercializations_notdel ON public.crop_commercializations USING btree (project_id, crop_id, net_price) WHERE (deleted_at IS NULL);


--
-- Name: idx_crop_commercializations_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crop_commercializations_project_id ON public.crop_commercializations USING btree (project_id);


--
-- Name: idx_crops_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crops_created_by ON public.crops USING btree (created_by);


--
-- Name: idx_crops_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crops_deleted_by ON public.crops USING btree (deleted_by);


--
-- Name: idx_crops_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crops_notdel ON public.crops USING btree (id, name) WHERE (deleted_at IS NULL);


--
-- Name: idx_crops_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crops_updated_by ON public.crops USING btree (updated_by);


--
-- Name: idx_customers_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_created_by ON public.customers USING btree (created_by);


--
-- Name: idx_customers_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_deleted_by ON public.customers USING btree (deleted_by);


--
-- Name: idx_customers_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_updated_by ON public.customers USING btree (updated_by);


--
-- Name: idx_field_investors_investor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_field_investors_investor_id ON public.field_investors USING btree (investor_id);


--
-- Name: idx_fields_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fields_created_by ON public.fields USING btree (created_by);


--
-- Name: idx_fields_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fields_deleted_by ON public.fields USING btree (deleted_by);


--
-- Name: idx_fields_lease_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fields_lease_type_id ON public.fields USING btree (lease_type_id);


--
-- Name: idx_fields_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fields_notdel ON public.fields USING btree (id, project_id, lease_type_id, lease_type_value, lease_type_percent) WHERE (deleted_at IS NULL);


--
-- Name: idx_fields_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fields_project_id ON public.fields USING btree (project_id);


--
-- Name: idx_fields_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fields_updated_by ON public.fields USING btree (updated_by);


--
-- Name: idx_fx_rates_effective_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fx_rates_effective_date ON public.fx_rates USING btree (effective_date);


--
-- Name: idx_investors_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_investors_created_by ON public.investors USING btree (created_by);


--
-- Name: idx_investors_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_investors_deleted_by ON public.investors USING btree (deleted_by);


--
-- Name: idx_investors_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_investors_updated_by ON public.investors USING btree (updated_by);


--
-- Name: idx_labor_categories_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_categories_type_id ON public.labor_categories USING btree (type_id);


--
-- Name: idx_labors_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labors_category_id ON public.labors USING btree (category_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_labors_harvest; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labors_harvest ON public.labors USING btree (id, category_id) WHERE ((deleted_at IS NULL) AND (category_id = 2));


--
-- Name: idx_labors_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labors_notdel ON public.labors USING btree (id, price) WHERE (deleted_at IS NULL);


--
-- Name: idx_labors_notdel_price; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labors_notdel_price ON public.labors USING btree (id, category_id, price) WHERE (deleted_at IS NULL);


--
-- Name: idx_labors_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labors_project_id ON public.labors USING btree (project_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_labors_sowing; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labors_sowing ON public.labors USING btree (id, category_id) WHERE ((deleted_at IS NULL) AND (category_id = 1));


--
-- Name: idx_lease_types_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lease_types_created_by ON public.lease_types USING btree (created_by);


--
-- Name: idx_lease_types_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lease_types_deleted_by ON public.lease_types USING btree (deleted_by);


--
-- Name: idx_lease_types_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lease_types_updated_by ON public.lease_types USING btree (updated_by);


--
-- Name: idx_lot_dates_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_dates_created_by ON public.lot_dates USING btree (created_by);


--
-- Name: idx_lot_dates_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_dates_deleted_by ON public.lot_dates USING btree (deleted_by);


--
-- Name: idx_lot_dates_lot_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_dates_lot_id ON public.lot_dates USING btree (lot_id);


--
-- Name: idx_lot_dates_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_dates_updated_by ON public.lot_dates USING btree (updated_by);


--
-- Name: idx_lots_composite_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_composite_notdel ON public.lots USING btree (field_id, current_crop_id, previous_crop_id, tons, hectares) WHERE ((deleted_at IS NULL) AND (hectares > (0)::numeric));


--
-- Name: idx_lots_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_created_by ON public.lots USING btree (created_by);


--
-- Name: idx_lots_current_crop_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_current_crop_id ON public.lots USING btree (current_crop_id);


--
-- Name: idx_lots_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_deleted_by ON public.lots USING btree (deleted_by);


--
-- Name: idx_lots_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_field_id ON public.lots USING btree (field_id);


--
-- Name: idx_lots_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_notdel ON public.lots USING btree (id, field_id, current_crop_id, previous_crop_id, tons, hectares) WHERE (deleted_at IS NULL);


--
-- Name: idx_lots_previous_crop_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_previous_crop_id ON public.lots USING btree (previous_crop_id);


--
-- Name: idx_lots_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_updated_by ON public.lots USING btree (updated_by);


--
-- Name: idx_managers_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_managers_created_by ON public.managers USING btree (created_by);


--
-- Name: idx_managers_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_managers_deleted_by ON public.managers USING btree (deleted_by);


--
-- Name: idx_managers_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_managers_updated_by ON public.managers USING btree (updated_by);


--
-- Name: idx_project_dollar_values_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_dollar_values_project_id ON public.project_dollar_values USING btree (project_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_project_investors_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_investors_created_by ON public.project_investors USING btree (created_by);


--
-- Name: idx_project_investors_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_investors_deleted_by ON public.project_investors USING btree (deleted_by);


--
-- Name: idx_project_investors_investor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_investors_investor_id ON public.project_investors USING btree (investor_id);


--
-- Name: idx_project_investors_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_investors_updated_by ON public.project_investors USING btree (updated_by);


--
-- Name: idx_project_managers_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_managers_created_by ON public.project_managers USING btree (created_by);


--
-- Name: idx_project_managers_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_managers_deleted_by ON public.project_managers USING btree (deleted_by);


--
-- Name: idx_project_managers_manager_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_managers_manager_id ON public.project_managers USING btree (manager_id);


--
-- Name: idx_project_managers_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_managers_updated_by ON public.project_managers USING btree (updated_by);


--
-- Name: idx_projects_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_active ON public.projects USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_projects_campaign_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_campaign_id ON public.projects USING btree (campaign_id);


--
-- Name: idx_projects_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_created_by ON public.projects USING btree (created_by);


--
-- Name: idx_projects_customer_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_customer_id ON public.projects USING btree (customer_id);


--
-- Name: idx_projects_deleted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_deleted_by ON public.projects USING btree (deleted_by);


--
-- Name: idx_projects_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_notdel ON public.projects USING btree (id, admin_cost) WHERE (deleted_at IS NULL);


--
-- Name: idx_projects_updated_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_updated_by ON public.projects USING btree (updated_by);


--
-- Name: idx_stocks_investor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stocks_investor_id ON public.stocks USING btree (investor_id);


--
-- Name: idx_stocks_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stocks_project_id ON public.stocks USING btree (project_id);


--
-- Name: idx_stocks_supply_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stocks_supply_id ON public.stocks USING btree (supply_id);


--
-- Name: idx_supplies_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_supplies_category_id ON public.supplies USING btree (category_id);


--
-- Name: idx_supplies_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_supplies_notdel ON public.supplies USING btree (id, price) WHERE (deleted_at IS NULL);


--
-- Name: idx_supplies_type_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_supplies_type_id ON public.supplies USING btree (type_id);


--
-- Name: idx_supplies_units_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_supplies_units_notdel ON public.supplies USING btree (id, price, unit_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_supply_movements_investor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_supply_movements_investor_id ON public.supply_movements USING btree (investor_id);


--
-- Name: idx_supply_movements_provider_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_supply_movements_provider_id ON public.supply_movements USING btree (provider_id);


--
-- Name: idx_supply_movements_supply_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_supply_movements_supply_id ON public.supply_movements USING btree (supply_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_workorder_items_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorder_items_notdel ON public.workorder_items USING btree (workorder_id, supply_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_workorder_items_supply_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorder_items_supply_id ON public.workorder_items USING btree (supply_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_workorder_items_supply_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorder_items_supply_notdel ON public.workorder_items USING btree (workorder_id, supply_id, final_dose) WHERE (deleted_at IS NULL);


--
-- Name: idx_workorder_items_v2_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorder_items_v2_notdel ON public.workorder_items USING btree (workorder_id, supply_id, total_used, final_dose) WHERE (deleted_at IS NULL);


--
-- Name: idx_workorder_items_workorder_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorder_items_workorder_id ON public.workorder_items USING btree (workorder_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_workorders_composite; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorders_composite ON public.workorders USING btree (project_id, field_id, labor_id, effective_area) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));


--
-- Name: idx_workorders_crop_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorders_crop_id ON public.workorders USING btree (crop_id);


--
-- Name: idx_workorders_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorders_field_id ON public.workorders USING btree (field_id);


--
-- Name: idx_workorders_grouping; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorders_grouping ON public.workorders USING btree (project_id, field_id, effective_area) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));


--
-- Name: idx_workorders_labor_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorders_labor_id ON public.workorders USING btree (labor_id);


--
-- Name: idx_workorders_lot_composite; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorders_lot_composite ON public.workorders USING btree (lot_id, labor_id, effective_area, date) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));


--
-- Name: idx_workorders_lot_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorders_lot_notdel ON public.workorders USING btree (lot_id, effective_area, date) WHERE (deleted_at IS NULL);


--
-- Name: idx_workorders_metrics_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorders_metrics_notdel ON public.workorders USING btree (project_id, field_id, labor_id, effective_area) WHERE (deleted_at IS NULL);


--
-- Name: users trg_users_set_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_users_set_timestamp BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: admin_cost_investors fk_admin_cost_investors_investor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT fk_admin_cost_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);


--
-- Name: admin_cost_investors fk_admin_cost_investors_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.admin_cost_investors
    ADD CONSTRAINT fk_admin_cost_investors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: campaigns fk_campaigns_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: campaigns fk_campaigns_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: campaigns fk_campaigns_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT fk_campaigns_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: categories fk_categories_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT fk_categories_type FOREIGN KEY (type_id) REFERENCES public.types(id);


--
-- Name: crop_commercializations fk_crop_commercializations_crop; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT fk_crop_commercializations_crop FOREIGN KEY (crop_id) REFERENCES public.crops(id);


--
-- Name: crop_commercializations fk_crop_commercializations_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT fk_crop_commercializations_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: crops fk_crops_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: crops fk_crops_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: crops fk_crops_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT fk_crops_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: customers fk_customers_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: customers fk_customers_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: customers fk_customers_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: field_investors fk_field_investors_field; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT fk_field_investors_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON DELETE CASCADE;


--
-- Name: field_investors fk_field_investors_investor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.field_investors
    ADD CONSTRAINT fk_field_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);


--
-- Name: fields fk_fields_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: fields fk_fields_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: fields fk_fields_lease_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_lease_type FOREIGN KEY (lease_type_id) REFERENCES public.lease_types(id);


--
-- Name: fields fk_fields_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: fields fk_fields_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_fields_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: investors fk_investors_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: investors fk_investors_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: investors fk_investors_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT fk_investors_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: invoices fk_invoices_work_order; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT fk_invoices_work_order FOREIGN KEY (work_order_id) REFERENCES public.workorders(id) ON DELETE CASCADE;


--
-- Name: labor_categories fk_labor_categories_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_categories
    ADD CONSTRAINT fk_labor_categories_type FOREIGN KEY (type_id) REFERENCES public.labor_types(id);


--
-- Name: labors fk_labors_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT fk_labors_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: labors fk_labors_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT fk_labors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: lease_types fk_lease_types_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: lease_types fk_lease_types_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: lease_types fk_lease_types_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT fk_lease_types_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: lot_dates fk_lot_dates_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: lot_dates fk_lot_dates_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: lot_dates fk_lot_dates_lot; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_lot FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON DELETE CASCADE;


--
-- Name: lot_dates fk_lot_dates_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: lots fk_lots_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: lots fk_lots_current_crop; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_current_crop FOREIGN KEY (current_crop_id) REFERENCES public.crops(id);


--
-- Name: lots fk_lots_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: lots fk_lots_field; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON DELETE CASCADE;


--
-- Name: lots fk_lots_previous_crop; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_previous_crop FOREIGN KEY (previous_crop_id) REFERENCES public.crops(id);


--
-- Name: lots fk_lots_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT fk_lots_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: managers fk_managers_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: managers fk_managers_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: managers fk_managers_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT fk_managers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: project_dollar_values fk_project_dollar_values_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT fk_project_dollar_values_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: project_investors fk_project_investors_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: project_investors fk_project_investors_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: project_investors fk_project_investors_investor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);


--
-- Name: project_investors fk_project_investors_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: project_investors fk_project_investors_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT fk_project_investors_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: project_managers fk_project_managers_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: project_managers fk_project_managers_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: project_managers fk_project_managers_manager; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_manager FOREIGN KEY (manager_id) REFERENCES public.managers(id);


--
-- Name: project_managers fk_project_managers_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: project_managers fk_project_managers_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT fk_project_managers_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: projects fk_projects_campaign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_campaign FOREIGN KEY (campaign_id) REFERENCES public.campaigns(id);


--
-- Name: projects fk_projects_created_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_created_by FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: projects fk_projects_customer; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_customer FOREIGN KEY (customer_id) REFERENCES public.customers(id);


--
-- Name: projects fk_projects_deleted_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_deleted_by FOREIGN KEY (deleted_by) REFERENCES public.users(id);


--
-- Name: projects fk_projects_updated_by; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT fk_projects_updated_by FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: stocks fk_stocks_investor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_stocks_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: stocks fk_stocks_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_stocks_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: stocks fk_stocks_supply; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_stocks_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: supplies fk_supplies_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT fk_supplies_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: supplies fk_supplies_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT fk_supplies_type FOREIGN KEY (type_id) REFERENCES public.types(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: supply_movements fk_supply_movements_investor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_supply_movements_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: supply_movements fk_supply_movements_provider; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_supply_movements_provider FOREIGN KEY (provider_id) REFERENCES public.providers(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: supply_movements fk_supply_movements_supply; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_supply_movements_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorder_items fk_workorder_items_supply; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT fk_workorder_items_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorder_items fk_workorder_items_workorder; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT fk_workorder_items_workorder FOREIGN KEY (workorder_id) REFERENCES public.workorders(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workorders fk_workorders_crop; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_crop FOREIGN KEY (crop_id) REFERENCES public.crops(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorders fk_workorders_field; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorders fk_workorders_labor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_labor FOREIGN KEY (labor_id) REFERENCES public.labors(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorders fk_workorders_lot; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_lot FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorders fk_workorders_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT fk_workorders_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--


