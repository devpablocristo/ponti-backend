--
-- PostgreSQL database dump
--

-- Dumped from database version 16.3 (Debian 16.3-1.pgdg120+1)
-- Dumped by pg_dump version 16.3 (Debian 16.3-1.pgdg120+1)

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
-- Name: calc; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA calc;


--
-- Name: calc_common; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA calc_common;


--
-- Name: report; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA report;


--
-- Name: movement_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.movement_type AS ENUM (
    'Stock',
    'Movimiento interno',
    'Remito oficial'
);


--
-- Name: norm_dose(numeric, numeric); Type: FUNCTION; Schema: calc; Owner: -
--

CREATE FUNCTION calc.norm_dose(dose numeric, area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN area > 0 THEN dose / area ELSE NULL END
$$;


--
-- Name: calculate_campaign_closing_date(date); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_campaign_closing_date(end_date date) RETURNS date
    LANGUAGE plpgsql
    AS $$
BEGIN
  IF end_date IS NULL THEN
    RETURN NULL;
  END IF;
  
  RETURN end_date + (get_campaign_closure_days() || ' days')::INTERVAL;
END;
$$;


--
-- Name: calculate_cost_per_ha(numeric, numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_cost_per_ha(p_total_cost numeric, p_hectares numeric) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN CASE 
        WHEN COALESCE(p_hectares, 0) > 0 
        THEN COALESCE(p_total_cost, 0) / p_hectares 
        ELSE 0 
    END;
END;
$$;


--
-- Name: calculate_harvested_area(numeric, numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_harvested_area(p_tons numeric, p_hectares numeric) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN CASE WHEN p_tons IS NOT NULL AND p_tons > 0 THEN COALESCE(p_hectares, 0) ELSE 0 END;
END;
$$;


--
-- Name: calculate_labor_cost(numeric, numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_labor_cost(p_labor_price numeric, p_effective_area numeric) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN COALESCE(p_labor_price * p_effective_area, 0);
END;
$$;


--
-- Name: calculate_sowed_area(date, numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_sowed_area(p_sowing_date date, p_hectares numeric) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN CASE WHEN p_sowing_date IS NOT NULL THEN COALESCE(p_hectares, 0) ELSE 0 END;
END;
$$;


--
-- Name: calculate_supply_cost(double precision, numeric, numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_supply_cost(p_final_dose double precision, p_supply_price numeric, p_effective_area numeric) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN COALESCE(p_final_dose * p_supply_price * p_effective_area, 0);
END;
$$;


--
-- Name: calculate_yield(numeric, numeric); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.calculate_yield(p_tons numeric, p_hectares numeric) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN CASE 
        WHEN COALESCE(p_hectares, 0) > 0 
        THEN COALESCE(p_tons, 0) / p_hectares 
        ELSE 0 
    END;
END;
$$;


--
-- Name: get_app_parameter(character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_app_parameter(p_key character varying) RETURNS character varying
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value FROM app_parameters WHERE key = p_key);
END;
$$;


--
-- Name: get_app_parameter_decimal(character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_app_parameter_decimal(p_key character varying) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value::DECIMAL FROM app_parameters WHERE key = p_key);
END;
$$;


--
-- Name: get_app_parameter_integer(character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_app_parameter_integer(p_key character varying) RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value::INTEGER FROM app_parameters WHERE key = p_key);
END;
$$;


--
-- Name: get_campaign_closure_days(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_campaign_closure_days() RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN get_app_parameter_integer('campaign_closure_days');
END;
$$;


--
-- Name: get_default_fx_rate(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_default_fx_rate() RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN get_app_parameter_decimal('default_fx_rate');
END;
$$;


--
-- Name: get_iva_percentage(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_iva_percentage() RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN get_app_parameter_decimal('iva_percentage');
END;
$$;


--
-- Name: get_project_dollar_value(bigint, character varying); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_project_dollar_value(p_project_id bigint, p_month character varying) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (
    SELECT average_value 
    FROM project_dollar_values 
    WHERE project_id = p_project_id 
      AND month = p_month 
      AND deleted_at IS NULL
    LIMIT 1
  );
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


SET default_tablespace = '';

SET default_table_access_method = heap;

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
    deleted_by bigint,
    effective_area_ha numeric(18,6) GENERATED ALWAYS AS (effective_area) STORED
);


--
-- Name: workorder_supply; Type: VIEW; Schema: calc_common; Owner: -
--

CREATE VIEW calc_common.workorder_supply AS
 SELECT w.project_id,
    w.field_id,
    w.lot_id,
    sum(wi.final_dose) AS dose_total
   FROM (public.workorder_items wi
     JOIN public.workorders w ON ((w.id = wi.workorder_id)))
  WHERE ((wi.deleted_at IS NULL) AND (w.deleted_at IS NULL))
  GROUP BY w.project_id, w.field_id, w.lot_id;


--
-- Name: workorder_surface; Type: VIEW; Schema: calc_common; Owner: -
--

CREATE VIEW calc_common.workorder_surface AS
 SELECT project_id,
    field_id,
    lot_id,
    sum(effective_area) AS surface_ha
   FROM public.workorders w
  WHERE (deleted_at IS NULL)
  GROUP BY project_id, field_id, lot_id;


--
-- Name: app_parameters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.app_parameters (
    id integer NOT NULL,
    key character varying(100) NOT NULL,
    value character varying(255) NOT NULL,
    type character varying(20) NOT NULL,
    category character varying(50) NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: app_parameters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.app_parameters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: app_parameters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.app_parameters_id_seq OWNED BY public.app_parameters.id;


--
-- Name: fields; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fields (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    project_id bigint NOT NULL,
    lease_type_id bigint NOT NULL,
    lease_type_percent double precision,
    lease_type_value double precision,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: lots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lots (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    field_id bigint NOT NULL,
    hectares double precision NOT NULL,
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
-- Name: projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.projects (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    customer_id bigint NOT NULL,
    campaign_id bigint NOT NULL,
    admin_cost numeric(12,2) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: base_admin_costs_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.base_admin_costs_view AS
 WITH project_total_hectares AS (
         SELECT p_1.id AS project_id,
            COALESCE(sum(l_1.hectares), (1)::double precision) AS total_hectares
           FROM ((public.projects p_1
             LEFT JOIN public.fields f_1 ON (((f_1.project_id = p_1.id) AND (f_1.deleted_at IS NULL))))
             LEFT JOIN public.lots l_1 ON (((l_1.field_id = f_1.id) AND (l_1.deleted_at IS NULL))))
          WHERE (p_1.deleted_at IS NULL)
          GROUP BY p_1.id
        )
 SELECT l.id AS lot_id,
    l.field_id,
    f.project_id,
    l.hectares,
    p.admin_cost,
    pth.total_hectares,
        CASE
            WHEN (l.hectares > (0)::double precision) THEN ((p.admin_cost)::double precision / pth.total_hectares)
            ELSE (0)::double precision
        END AS admin_cost_per_ha
   FROM (((public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
     JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
     JOIN project_total_hectares pth ON ((pth.project_id = p.id)))
  WHERE (l.deleted_at IS NULL);


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
-- Name: supplies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.supplies (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    name character varying(100) NOT NULL,
    price double precision NOT NULL,
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
-- Name: base_direct_costs_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.base_direct_costs_view AS
 WITH labor_costs AS (
         SELECT w.project_id,
            w.field_id,
            w.lot_id,
            sum((lb.price * w.effective_area)) AS labor_cost
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (lb.price IS NOT NULL))
          GROUP BY w.project_id, w.field_id, w.lot_id
        ), supply_costs AS (
         SELECT w.project_id,
            w.field_id,
            w.lot_id,
            sum((((wi.final_dose)::double precision * s.price) * (w.effective_area)::double precision)) AS supply_cost
           FROM ((public.workorders w
             JOIN public.workorder_items wi ON ((wi.workorder_id = w.id)))
             JOIN public.supplies s ON ((s.id = wi.supply_id)))
          WHERE ((w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric) AND (wi.final_dose > (0)::numeric) AND (s.price IS NOT NULL))
          GROUP BY w.project_id, w.field_id, w.lot_id
        )
 SELECT COALESCE(lc.project_id, sc.project_id) AS project_id,
    COALESCE(lc.field_id, sc.field_id) AS field_id,
    COALESCE(lc.lot_id, sc.lot_id) AS lot_id,
    COALESCE(lc.labor_cost, (0)::numeric) AS labor_cost,
    COALESCE(sc.supply_cost, (0)::double precision) AS supply_cost,
    ((COALESCE(lc.labor_cost, (0)::numeric))::double precision + COALESCE(sc.supply_cost, (0)::double precision)) AS direct_cost
   FROM (labor_costs lc
     FULL JOIN supply_costs sc ON (((lc.project_id = sc.project_id) AND (lc.field_id = sc.field_id) AND (lc.lot_id = sc.lot_id))));


--
-- Name: crop_commercializations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.crop_commercializations (
    id bigint NOT NULL,
    project_id bigint NOT NULL,
    crop_id bigint NOT NULL,
    board_price numeric(12,2) NOT NULL,
    freight_cost numeric(12,2) NOT NULL,
    commercial_cost double precision NOT NULL,
    net_price numeric(12,2) NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    deleted_at timestamp without time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
);


--
-- Name: base_income_net_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.base_income_net_view AS
 SELECT l.id AS lot_id,
    l.field_id,
    f.project_id,
    l.current_crop_id,
    l.hectares,
    l.tons,
    cc.net_price AS net_price_usd,
    (l.tons * cc.net_price) AS income_net_total,
        CASE
            WHEN (l.hectares > (0)::double precision) THEN (((l.tons * cc.net_price))::double precision / l.hectares)
            ELSE (0)::double precision
        END AS income_net_per_ha
   FROM ((public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.crop_commercializations cc ON (((cc.project_id = f.project_id) AND (cc.crop_id = l.current_crop_id) AND (cc.deleted_at IS NULL))))
  WHERE (l.deleted_at IS NULL);


--
-- Name: base_lease_calculations_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.base_lease_calculations_view AS
 SELECT l.id AS lot_id,
    l.field_id,
    f.project_id,
    l.hectares,
    f.lease_type_id,
    f.lease_type_percent,
    f.lease_type_value,
    bin.income_net_per_ha,
    (bdc.direct_cost / NULLIF(l.hectares, (0)::double precision)) AS cost_per_ha,
    bac.admin_cost_per_ha,
        CASE
            WHEN (f.lease_type_id = 1) THEN ((COALESCE(f.lease_type_percent, (0)::double precision) / (100.0)::double precision) * bin.income_net_per_ha)
            WHEN (f.lease_type_id = 2) THEN ((COALESCE(f.lease_type_percent, (0)::double precision) / (100.0)::double precision) * ((bin.income_net_per_ha - (bdc.direct_cost / NULLIF(l.hectares, (0)::double precision))) - bac.admin_cost_per_ha))
            WHEN (f.lease_type_id = 3) THEN COALESCE(f.lease_type_value, (0)::double precision)
            WHEN (f.lease_type_id = 4) THEN (COALESCE(f.lease_type_value, (0)::double precision) + ((COALESCE(f.lease_type_percent, (0)::double precision) / (100.0)::double precision) * bin.income_net_per_ha))
            ELSE (0)::double precision
        END AS rent_per_ha
   FROM ((((public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.base_income_net_view bin ON ((bin.lot_id = l.id)))
     LEFT JOIN public.base_direct_costs_view bdc ON ((bdc.lot_id = l.id)))
     LEFT JOIN public.base_admin_costs_view bac ON ((bac.lot_id = l.id)))
  WHERE (l.deleted_at IS NULL);


--
-- Name: base_active_total_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.base_active_total_view AS
 SELECT l.id AS lot_id,
    l.field_id,
    f.project_id,
    l.hectares,
    (bdc.direct_cost / NULLIF(l.hectares, (0)::double precision)) AS cost_per_ha,
    blc.rent_per_ha,
    bac.admin_cost_per_ha,
    (((bdc.direct_cost / NULLIF(l.hectares, (0)::double precision)) + blc.rent_per_ha) + bac.admin_cost_per_ha) AS active_total_per_ha
   FROM ((((public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.base_direct_costs_view bdc ON ((bdc.lot_id = l.id)))
     LEFT JOIN public.base_lease_calculations_view blc ON ((blc.lot_id = l.id)))
     LEFT JOIN public.base_admin_costs_view bac ON ((bac.lot_id = l.id)))
  WHERE (l.deleted_at IS NULL);


--
-- Name: base_operating_result_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.base_operating_result_view AS
 SELECT l.id AS lot_id,
    l.field_id,
    f.project_id,
    l.hectares,
    bin.income_net_per_ha,
    bat.active_total_per_ha,
    (bin.income_net_per_ha - bat.active_total_per_ha) AS operating_result_per_ha
   FROM (((public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.base_income_net_view bin ON ((bin.lot_id = l.id)))
     LEFT JOIN public.base_active_total_view bat ON ((bat.lot_id = l.id)))
  WHERE (l.deleted_at IS NULL);


--
-- Name: base_yield_calculations_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.base_yield_calculations_view AS
 SELECT l.id AS lot_id,
    l.field_id,
    f.project_id,
    l.current_crop_id,
    l.hectares,
    l.tons,
    COALESCE(((l.tons)::double precision / NULLIF(l.hectares, (0)::double precision)), (0)::double precision) AS yield_tn_per_ha
   FROM (public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
  WHERE (l.deleted_at IS NULL);


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
-- Name: dashboard_balance_management_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_balance_management_view AS
 WITH executed_costs AS (
         SELECT w.project_id,
            sum(
                CASE
                    WHEN (w.effective_area > (0)::numeric) THEN (lb.price * w.effective_area)
                    ELSE (0)::numeric
                END) AS labors_executed_usd,
            sum(
                CASE
                    WHEN ((wi.final_dose > (0)::numeric) AND (s.type_id <> 1)) THEN ((wi.total_used)::double precision * s.price)
                    ELSE (0)::double precision
                END) AS supplies_executed_usd,
            sum(
                CASE
                    WHEN ((wi.final_dose > (0)::numeric) AND (s.type_id = 1)) THEN ((wi.total_used)::double precision * s.price)
                    ELSE (0)::double precision
                END) AS seeds_executed_usd
           FROM (((public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
             LEFT JOIN public.workorder_items wi ON ((wi.workorder_id = w.id)))
             LEFT JOIN public.supplies s ON ((s.id = wi.supply_id)))
          GROUP BY w.project_id
        ), invested_costs AS (
         SELECT p_1.id AS project_id,
            COALESCE(ec_1.labors_executed_usd, (0)::numeric) AS labors_invested_usd,
            COALESCE(ec_1.supplies_executed_usd, (0)::double precision) AS supplies_invested_usd,
            COALESCE(ec_1.seeds_executed_usd, (0)::double precision) AS seeds_invested_usd
           FROM (public.projects p_1
             LEFT JOIN executed_costs ec_1 ON ((ec_1.project_id = p_1.id)))
        ), stock_costs AS (
         SELECT ic_1.project_id,
            (0)::numeric AS labors_stock_usd,
            GREATEST((0)::double precision, (ic_1.supplies_invested_usd - COALESCE(ec_1.supplies_executed_usd, (0)::double precision))) AS supplies_stock_usd,
            GREATEST((0)::double precision, (ic_1.seeds_invested_usd - COALESCE(ec_1.seeds_executed_usd, (0)::double precision))) AS seeds_stock_usd
           FROM (invested_costs ic_1
             LEFT JOIN executed_costs ec_1 ON ((ec_1.project_id = ic_1.project_id)))
        )
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(ec.seeds_executed_usd, (0)::double precision) AS seeds_executed_usd,
    COALESCE(ic.seeds_invested_usd, (0)::double precision) AS seeds_invested_usd,
    COALESCE(sc.seeds_stock_usd, (0)::double precision) AS seeds_stock_usd,
    COALESCE(ec.supplies_executed_usd, (0)::double precision) AS supplies_executed_usd,
    COALESCE(ic.supplies_invested_usd, (0)::double precision) AS supplies_invested_usd,
    COALESCE(sc.supplies_stock_usd, (0)::double precision) AS supplies_stock_usd,
    COALESCE(ec.labors_executed_usd, (0)::numeric) AS labors_executed_usd,
    COALESCE(ic.labors_invested_usd, (0)::numeric) AS labors_invested_usd,
    COALESCE(sc.labors_stock_usd, (0)::numeric) AS labors_stock_usd,
    ((COALESCE(ec.seeds_executed_usd, (0)::double precision) + COALESCE(ec.supplies_executed_usd, (0)::double precision)) + (COALESCE(ec.labors_executed_usd, (0)::numeric))::double precision) AS direct_costs_executed_usd,
    ((COALESCE(ic.seeds_invested_usd, (0)::double precision) + COALESCE(ic.supplies_invested_usd, (0)::double precision)) + (COALESCE(ic.labors_invested_usd, (0)::numeric))::double precision) AS direct_costs_invested_usd,
    ((COALESCE(sc.seeds_stock_usd, (0)::double precision) + COALESCE(sc.supplies_stock_usd, (0)::double precision)) + (COALESCE(sc.labors_stock_usd, (0)::numeric))::double precision) AS direct_costs_stock_usd,
    (0)::numeric AS lease_invested_usd,
    p.admin_cost AS structure_invested_usd,
    ((((COALESCE(ic.seeds_invested_usd, (0)::double precision) + COALESCE(ic.supplies_invested_usd, (0)::double precision)) + (COALESCE(ic.labors_invested_usd, (0)::numeric))::double precision) + (0)::double precision) + (p.admin_cost)::double precision) AS total_invested_usd
   FROM (((public.projects p
     LEFT JOIN executed_costs ec ON ((ec.project_id = p.id)))
     LEFT JOIN invested_costs ic ON ((ic.project_id = p.id)))
     LEFT JOIN stock_costs sc ON ((sc.project_id = p.id)));


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
-- Name: dashboard_contributions_progress_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_contributions_progress_view AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    f.id AS field_id,
    (0)::numeric(14,2) AS sowing_hectares,
    (0)::numeric(14,2) AS sowing_total_hectares,
    (0)::numeric(6,2) AS sowing_progress_pct,
    (0)::numeric(14,2) AS harvest_hectares,
    (0)::numeric(14,2) AS harvest_total_hectares,
    (0)::numeric(6,2) AS harvest_progress_pct,
    (0)::numeric(14,2) AS executed_costs_usd,
    (0)::numeric(14,2) AS executed_labors_usd,
    (0)::numeric(14,2) AS executed_supplies_usd,
    (0)::numeric(14,2) AS budget_cost_usd,
    (0)::numeric(14,2) AS budget_total_usd,
    (0)::numeric(6,2) AS costs_progress_pct,
    (0)::numeric(14,2) AS income_usd,
    (0)::numeric(14,2) AS operating_result_usd,
    (0)::numeric(6,2) AS operating_result_pct,
    (0)::numeric(14,2) AS operating_result_total_costs_usd,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS investor_percentage_pct,
    100.00::numeric(6,2) AS contributions_progress_pct,
    (0)::bigint AS crop_id,
    ''::text AS crop_name,
    (0)::numeric(14,2) AS crop_hectares,
    (0)::numeric(14,2) AS project_total_hectares,
    (0)::numeric(6,2) AS incidence_pct,
    (0)::numeric(14,2) AS crop_direct_costs_usd,
    (0)::numeric(14,2) AS cost_per_ha_usd,
    (0)::numeric(14,2) AS balance_executed_costs_usd,
    (0)::numeric(14,2) AS balance_budget_cost_usd,
    (0)::numeric(14,2) AS balance_operating_result_total_costs_usd,
    (0)::numeric(14,2) AS balance_operating_result_usd,
    (0)::numeric(6,2) AS balance_operating_result_pct,
    NULL::timestamp without time zone AS primera_orden_fecha,
    (0)::bigint AS primera_orden_id,
    NULL::timestamp without time zone AS ultima_orden_fecha,
    (0)::bigint AS ultima_orden_id,
    NULL::timestamp without time zone AS arqueo_stock_fecha,
    NULL::timestamp without time zone AS cierre_campana_fecha,
    'metric'::text AS row_kind
   FROM (((public.projects p
     JOIN public.fields f ON ((f.project_id = p.id)))
     JOIN public.project_investors pi ON ((pi.project_id = p.id)))
     JOIN public.investors i ON ((i.id = pi.investor_id)));


--
-- Name: dashboard_contributions_progress_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_contributions_progress_view_v2 AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS investor_percentage_pct,
    100.00 AS contributions_progress_pct
   FROM ((public.projects p
     JOIN public.project_investors pi ON (((pi.project_id = p.id) AND (pi.deleted_at IS NULL))))
     JOIN public.investors i ON (((i.id = pi.investor_id) AND (i.deleted_at IS NULL))))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id, pi.investor_id, i.name, pi.percentage;


--
-- Name: dashboard_costs_progress_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_costs_progress_view AS
 WITH labors_costs AS (
         SELECT w.project_id,
            sum((lb.price * w.effective_area)) AS executed_labors_usd
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE (w.effective_area > (0)::numeric)
          GROUP BY w.project_id
        ), supplies_costs AS (
         SELECT w.project_id,
            sum(((wi.total_used)::double precision * s.price)) AS executed_supplies_usd
           FROM ((public.workorders w
             JOIN public.workorder_items wi ON ((wi.workorder_id = w.id)))
             JOIN public.supplies s ON ((s.id = wi.supply_id)))
          WHERE (wi.final_dose > (0)::numeric)
          GROUP BY w.project_id
        )
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(lc.executed_labors_usd, (0)::numeric) AS executed_labors_usd,
    COALESCE(sc.executed_supplies_usd, (0)::double precision) AS executed_supplies_usd,
    ((COALESCE(lc.executed_labors_usd, (0)::numeric))::double precision + COALESCE(sc.executed_supplies_usd, (0)::double precision)) AS executed_costs_usd,
    p.admin_cost AS budget_cost_usd,
    (20000)::numeric AS budget_total_usd,
    LEAST(
        CASE
            WHEN (20000 > 0) THEN ((((COALESCE(lc.executed_labors_usd, (0)::numeric))::double precision + COALESCE(sc.executed_supplies_usd, (0)::double precision)) / (20000.0)::double precision) * (100)::double precision)
            ELSE (0)::double precision
        END, (100)::double precision) AS costs_progress_pct
   FROM ((public.projects p
     LEFT JOIN labors_costs lc ON ((lc.project_id = p.id)))
     LEFT JOIN supplies_costs sc ON ((sc.project_id = p.id)));


--
-- Name: dashboard_costs_progress_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_costs_progress_view_v2 AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(sum(bdc.direct_cost), (0)::double precision) AS executed_costs_usd,
    p.admin_cost AS budget_cost_usd,
        CASE
            WHEN (p.admin_cost > (0)::numeric) THEN ((COALESCE(sum(bdc.direct_cost), (0)::double precision) / (p.admin_cost)::double precision) * (100)::double precision)
            ELSE (0)::double precision
        END AS costs_progress_pct
   FROM (public.projects p
     LEFT JOIN public.base_direct_costs_view bdc ON ((bdc.project_id = p.id)))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost;


--
-- Name: dashboard_crop_cost_incidence_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_crop_cost_incidence_view AS
 WITH lot_costs AS (
         SELECT w.lot_id,
            ((sum((lb.price * w.effective_area)))::double precision + sum(((wi.total_used)::double precision * s.price))) AS direct_costs_usd
           FROM (((public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
             LEFT JOIN public.workorder_items wi ON ((wi.workorder_id = w.id)))
             LEFT JOIN public.supplies s ON ((s.id = wi.supply_id)))
          GROUP BY w.lot_id
        )
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    l.current_crop_id AS crop_id,
    c.name AS crop_name,
    sum(l.hectares) AS crop_hectares,
    sum(sum(l.hectares)) OVER (PARTITION BY p.id) AS project_total_hectares,
    ((((sum(l.hectares))::numeric)::double precision / NULLIF(sum(sum(l.hectares)) OVER (PARTITION BY p.id), (0)::double precision)) * (100)::double precision) AS incidence_pct,
    COALESCE(sum(lc.direct_costs_usd), (0)::double precision) AS crop_direct_costs_usd,
        CASE
            WHEN (sum(l.hectares) > (0)::double precision) THEN (((COALESCE(sum(lc.direct_costs_usd), (0)::double precision))::numeric)::double precision / sum(l.hectares))
            ELSE (0)::double precision
        END AS cost_per_ha_usd
   FROM ((((public.projects p
     JOIN public.fields f ON ((f.project_id = p.id)))
     JOIN public.lots l ON ((l.field_id = f.id)))
     JOIN public.crops c ON ((c.id = l.current_crop_id)))
     LEFT JOIN lot_costs lc ON ((lc.lot_id = l.id)))
  GROUP BY p.customer_id, p.id, p.campaign_id, l.current_crop_id, c.name;


--
-- Name: dashboard_crop_incidence_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_crop_incidence_view_v2 AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    l.current_crop_id,
    cc.name AS crop_name,
    sum(l.hectares) AS crop_hectares,
    ((sum(l.hectares) / NULLIF(sum(sum(l.hectares)) OVER (PARTITION BY p.id), (0)::double precision)) * (100)::double precision) AS crop_incidence_pct
   FROM (((public.projects p
     JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
     LEFT JOIN public.crops cc ON (((cc.id = l.current_crop_id) AND (cc.deleted_at IS NULL))))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id, l.current_crop_id, cc.name;


--
-- Name: dashboard_harvest_progress_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_harvest_progress_view AS
 WITH harvest_workorders AS (
         SELECT DISTINCT w.lot_id
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((lb.category_id = 13) AND (w.effective_area > (0)::numeric))
        ), harvest_lots AS (
         SELECT f.project_id,
            sum(
                CASE
                    WHEN (hw.lot_id IS NOT NULL) THEN l.hectares
                    ELSE (0)::double precision
                END) AS harvested_hectares,
            sum(l.hectares) AS total_hectares
           FROM (((public.projects p_1
             JOIN public.fields f ON ((f.project_id = p_1.id)))
             JOIN public.lots l ON ((l.field_id = f.id)))
             LEFT JOIN harvest_workorders hw ON ((hw.lot_id = l.id)))
          GROUP BY f.project_id
        )
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(hl.harvested_hectares, (0)::double precision) AS harvest_hectares,
    COALESCE(hl.total_hectares, (0)::double precision) AS harvest_total_hectares,
        CASE
            WHEN (COALESCE(hl.total_hectares, (0)::double precision) > (0)::double precision) THEN (((COALESCE(hl.harvested_hectares, (0)::double precision))::numeric / (hl.total_hectares)::numeric) * (100)::numeric)
            ELSE (0)::numeric
        END AS harvest_progress_pct
   FROM (public.projects p
     LEFT JOIN harvest_lots hl ON ((hl.project_id = p.id)));


--
-- Name: dashboard_harvest_progress_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_harvest_progress_view_v2 AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    sum(
        CASE
            WHEN ((l.tons IS NOT NULL) AND (l.tons > (0)::numeric)) THEN l.hectares
            ELSE (0)::double precision
        END) AS harvest_hectares,
    sum(l.hectares) AS harvest_total_hectares,
        CASE
            WHEN (sum(l.hectares) > (0)::double precision) THEN ((sum(
            CASE
                WHEN ((l.tons IS NOT NULL) AND (l.tons > (0)::numeric)) THEN l.hectares
                ELSE (0)::double precision
            END) / sum(l.hectares)) * (100)::double precision)
            ELSE (0)::double precision
        END AS harvest_progress_pct
   FROM ((public.projects p
     JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id;


--
-- Name: dashboard_management_balance_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_management_balance_view_v2 AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(sum(bin.income_net_total), (0)::numeric) AS income_usd,
    COALESCE(sum(bdc.direct_cost), (0)::double precision) AS costos_directos_ejecutados_usd,
    COALESCE(sum(bdc.direct_cost), (0)::double precision) AS costos_directos_invertidos_usd,
    COALESCE(sum((blc.rent_per_ha * l.hectares)), (0)::double precision) AS arriendo_invertidos_usd,
    COALESCE(sum((bac.admin_cost_per_ha * l.hectares)), (0)::double precision) AS estructura_invertidos_usd,
    COALESCE(sum((bor.operating_result_per_ha * l.hectares)), (0)::double precision) AS operating_result_usd,
        CASE
            WHEN (COALESCE(sum(bdc.direct_cost), (0)::double precision) > (0)::double precision) THEN ((COALESCE(sum((bor.operating_result_per_ha * l.hectares)), (0)::double precision) / COALESCE(sum(bdc.direct_cost), (0)::double precision)) * (100)::double precision)
            ELSE (0)::double precision
        END AS operating_result_pct
   FROM (((((((public.projects p
     LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
     LEFT JOIN public.base_income_net_view bin ON ((bin.lot_id = l.id)))
     LEFT JOIN public.base_direct_costs_view bdc ON ((bdc.lot_id = l.id)))
     LEFT JOIN public.base_lease_calculations_view blc ON ((blc.lot_id = l.id)))
     LEFT JOIN public.base_admin_costs_view bac ON ((bac.lot_id = l.id)))
     LEFT JOIN public.base_operating_result_view bor ON ((bor.lot_id = l.id)))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id;


--
-- Name: dashboard_operating_result_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_operating_result_view_v2 AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(sum((bor.operating_result_per_ha * l.hectares)), (0)::double precision) AS operating_result_usd,
    COALESCE(sum(bdc.direct_cost), (0)::double precision) AS operating_result_total_costs_usd,
        CASE
            WHEN (COALESCE(sum(bdc.direct_cost), (0)::double precision) > (0)::double precision) THEN ((COALESCE(sum((bor.operating_result_per_ha * l.hectares)), (0)::double precision) / COALESCE(sum(bdc.direct_cost), (0)::double precision)) * (100)::double precision)
            ELSE (0)::double precision
        END AS operating_result_pct
   FROM ((((public.projects p
     LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
     LEFT JOIN public.base_operating_result_view bor ON ((bor.lot_id = l.id)))
     LEFT JOIN public.base_direct_costs_view bdc ON ((bdc.lot_id = l.id)))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id;


--
-- Name: dashboard_operational_indicators_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_operational_indicators_view AS
 WITH workorder_costs AS (
         SELECT w.project_id,
            sum((lb.price * w.effective_area)) AS labors_cost_usd,
            sum(COALESCE(((wi.total_used)::double precision * s.price), (0)::double precision)) AS supplies_cost_usd
           FROM (((public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
             LEFT JOIN public.workorder_items wi ON ((wi.workorder_id = w.id)))
             LEFT JOIN public.supplies s ON ((s.id = wi.supply_id)))
          GROUP BY w.project_id
        ), supply_stocks AS (
         SELECT s.project_id,
            sum(
                CASE
                    WHEN (s.type_id = 1) THEN (s.price * (1000)::double precision)
                    ELSE (0)::double precision
                END) AS seeds_stock_usd,
            sum(
                CASE
                    WHEN (s.type_id = 3) THEN (s.price * (1000)::double precision)
                    ELSE (0)::double precision
                END) AS supplies_stock_usd
           FROM public.supplies s
          GROUP BY s.project_id
        ), workorder_dates AS (
         SELECT w.project_id,
            min(w.date) AS first_workorder_date,
            min(w.id) AS first_workorder_number,
            max(w.date) AS last_workorder_date,
            max(w.id) AS last_workorder_number
           FROM public.workorders w
          GROUP BY w.project_id
        ), lot_summary AS (
         SELECT f.project_id,
            sum(
                CASE
                    WHEN (l.sowing_date IS NOT NULL) THEN l.hectares
                    ELSE (0)::double precision
                END) AS sowing_hectares,
            sum(l.hectares) AS sowing_total_hectares,
            sum(
                CASE
                    WHEN (l.tons > (0)::numeric) THEN l.hectares
                    ELSE (0)::double precision
                END) AS harvest_hectares,
            sum(l.hectares) AS harvest_total_hectares
           FROM (public.fields f
             LEFT JOIN public.lots l ON ((l.field_id = f.id)))
          GROUP BY f.project_id
        )
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(ls.sowing_hectares, (0)::double precision) AS sowing_hectares,
    COALESCE(ls.sowing_total_hectares, (0)::double precision) AS sowing_total_hectares,
    COALESCE(ls.harvest_hectares, (0)::double precision) AS harvest_hectares,
    COALESCE(ls.harvest_total_hectares, (0)::double precision) AS harvest_total_hectares,
    wd.first_workorder_date,
    wd.first_workorder_number,
    wd.last_workorder_date,
    wd.last_workorder_number,
    NULL::date AS last_stock_count_date,
    NULL::date AS campaign_closing_date,
    COALESCE((wc.supplies_cost_usd * (0.5)::double precision), (0)::double precision) AS seeds_executed_usd,
    COALESCE((wc.supplies_cost_usd * (0.6)::double precision), (0)::double precision) AS seeds_invested_usd,
    COALESCE(ss.seeds_stock_usd, (0)::double precision) AS seeds_stock_usd,
    COALESCE((wc.supplies_cost_usd * (0.5)::double precision), (0)::double precision) AS supplies_executed_usd,
    COALESCE((wc.supplies_cost_usd * (0.4)::double precision), (0)::double precision) AS supplies_invested_usd,
    COALESCE(ss.supplies_stock_usd, (0)::double precision) AS supplies_stock_usd,
    COALESCE(wc.labors_cost_usd, (0)::numeric) AS labors_executed_usd,
    COALESCE((wc.labors_cost_usd * 1.2), (0)::numeric) AS labors_invested_usd,
    COALESCE((wc.labors_cost_usd * 0.3), (0)::numeric) AS labors_stock_usd
   FROM ((((public.projects p
     LEFT JOIN lot_summary ls ON ((ls.project_id = p.id)))
     LEFT JOIN workorder_costs wc ON ((wc.project_id = p.id)))
     LEFT JOIN supply_stocks ss ON ((ss.project_id = p.id)))
     LEFT JOIN workorder_dates wd ON ((wd.project_id = p.id)));


--
-- Name: dashboard_operational_indicators_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_operational_indicators_view_v2 AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    w_min.min_date AS start_date,
    w_max.max_date AS end_date,
    public.calculate_campaign_closing_date(w_max.max_date) AS campaign_closing_date
   FROM ((public.projects p
     LEFT JOIN ( SELECT workorders.project_id,
            min(workorders.date) AS min_date
           FROM public.workorders
          WHERE (workorders.deleted_at IS NULL)
          GROUP BY workorders.project_id) w_min ON ((w_min.project_id = p.id)))
     LEFT JOIN ( SELECT workorders.project_id,
            max(workorders.date) AS max_date
           FROM public.workorders
          WHERE (workorders.deleted_at IS NULL)
          GROUP BY workorders.project_id) w_max ON ((w_max.project_id = p.id)))
  WHERE (p.deleted_at IS NULL);


--
-- Name: dashboard_sowing_progress_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_sowing_progress_view AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(total.total_hectares, (0)::double precision) AS sowing_total_hectares,
    COALESCE(sown.sown_hectares, (0)::double precision) AS sowing_hectares,
        CASE
            WHEN (COALESCE(total.total_hectares, (0)::double precision) > (0)::double precision) THEN ((((COALESCE(sown.sown_hectares, (0)::double precision))::numeric)::double precision / total.total_hectares) * (100)::double precision)
            ELSE (0)::double precision
        END AS sowing_progress_pct
   FROM ((public.projects p
     LEFT JOIN ( SELECT f.project_id,
            sum(l.hectares) AS total_hectares
           FROM (public.fields f
             JOIN public.lots l ON ((l.field_id = f.id)))
          GROUP BY f.project_id) total ON ((total.project_id = p.id)))
     LEFT JOIN ( SELECT f.project_id,
            sum(l.hectares) AS sown_hectares
           FROM (public.fields f
             JOIN public.lots l ON ((l.field_id = f.id)))
          WHERE (l.sowing_date IS NOT NULL)
          GROUP BY f.project_id) sown ON ((sown.project_id = p.id)));


--
-- Name: dashboard_sowing_progress_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_sowing_progress_view_v2 AS
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    sum(
        CASE
            WHEN (l.sowing_date IS NOT NULL) THEN l.hectares
            ELSE (0)::double precision
        END) AS sowing_hectares,
    sum(l.hectares) AS sowing_total_hectares,
        CASE
            WHEN (sum(l.hectares) > (0)::double precision) THEN ((sum(
            CASE
                WHEN (l.sowing_date IS NOT NULL) THEN l.hectares
                ELSE (0)::double precision
            END) / sum(l.hectares)) * (100)::double precision)
            ELSE (0)::double precision
        END AS sowing_progress_pct
   FROM ((public.projects p
     JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id;


--
-- Name: dashboard_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dashboard_view AS
 WITH executed_labors_by_project AS (
         SELECT lb.project_id,
            sum(lb.price) AS labors_usd
           FROM public.labors lb
          WHERE ((lb.deleted_at IS NULL) AND (EXISTS ( SELECT 1
                   FROM public.workorders w
                  WHERE ((w.labor_id = lb.id) AND (w.effective_area > (0)::numeric) AND (w.deleted_at IS NULL)))))
          GROUP BY lb.project_id
        ), used_supplies_by_project AS (
         SELECT sp.project_id,
            sum(sp.price) AS supplies_usd
           FROM public.supplies sp
          WHERE ((sp.deleted_at IS NULL) AND (EXISTS ( SELECT 1
                   FROM public.workorder_items wi
                  WHERE ((wi.supply_id = sp.id) AND (wi.final_dose > (0)::numeric) AND (wi.deleted_at IS NULL)))))
          GROUP BY sp.project_id
        ), v_direct_costs_by_project AS (
         SELECT p.id AS project_id,
            (COALESCE(el.labors_usd, (0)::numeric))::numeric(14,2) AS labors_usd,
            (COALESCE(us.supplies_usd, (0)::double precision))::numeric(14,2) AS supplies_usd,
            (((COALESCE(el.labors_usd, (0)::numeric))::double precision + COALESCE(us.supplies_usd, (0)::double precision)))::numeric(14,2) AS direct_costs_usd
           FROM ((public.projects p
             LEFT JOIN executed_labors_by_project el ON ((el.project_id = p.id)))
             LEFT JOIN used_supplies_by_project us ON ((us.project_id = p.id)))
          WHERE (p.deleted_at IS NULL)
        ), v_income_by_field AS (
         SELECT f.project_id,
            f.id AS field_id,
            (COALESCE(sum((l.tons * (200)::numeric)), (0)::numeric))::numeric(14,2) AS income_usd
           FROM (public.fields f
             LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
          WHERE (f.deleted_at IS NULL)
          GROUP BY f.project_id, f.id
        ), levels AS (
         SELECT
                CASE
                    WHEN (GROUPING(p.customer_id) = 1) THEN NULL::bigint
                    ELSE p.customer_id
                END AS customer_id,
                CASE
                    WHEN (GROUPING(p.id) = 1) THEN NULL::bigint
                    ELSE p.id
                END AS project_id,
                CASE
                    WHEN (GROUPING(p.campaign_id) = 1) THEN NULL::bigint
                    ELSE p.campaign_id
                END AS campaign_id,
                CASE
                    WHEN (GROUPING(f.id) = 1) THEN NULL::bigint
                    ELSE f.id
                END AS field_id
           FROM (public.projects p
             LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
          WHERE (p.deleted_at IS NULL)
          GROUP BY GROUPING SETS ((p.customer_id, p.id, p.campaign_id, f.id), (p.customer_id, p.id, p.campaign_id), (p.customer_id, p.id), (p.customer_id), ())
        ), sowing AS (
         SELECT
                CASE
                    WHEN (GROUPING(p.customer_id) = 1) THEN NULL::bigint
                    ELSE p.customer_id
                END AS customer_id,
                CASE
                    WHEN (GROUPING(p.id) = 1) THEN NULL::bigint
                    ELSE p.id
                END AS project_id,
                CASE
                    WHEN (GROUPING(p.campaign_id) = 1) THEN NULL::bigint
                    ELSE p.campaign_id
                END AS campaign_id,
                CASE
                    WHEN (GROUPING(f.id) = 1) THEN NULL::bigint
                    ELSE f.id
                END AS field_id,
            (COALESCE(sum(
                CASE
                    WHEN (l.sowing_date IS NOT NULL) THEN l.hectares
                    ELSE (0)::double precision
                END), (0)::double precision))::numeric(14,2) AS sowed_area,
            (COALESCE(sum(l.hectares), (0)::double precision))::numeric(14,2) AS total_hectares
           FROM ((public.projects p
             LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
             LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
          WHERE (p.deleted_at IS NULL)
          GROUP BY GROUPING SETS ((p.customer_id, p.id, p.campaign_id, f.id), (p.customer_id, p.id, p.campaign_id), (p.customer_id, p.id), (p.customer_id), ())
        ), harvest AS (
         SELECT
                CASE
                    WHEN (GROUPING(p.customer_id) = 1) THEN NULL::bigint
                    ELSE p.customer_id
                END AS customer_id,
                CASE
                    WHEN (GROUPING(p.id) = 1) THEN NULL::bigint
                    ELSE p.id
                END AS project_id,
                CASE
                    WHEN (GROUPING(p.campaign_id) = 1) THEN NULL::bigint
                    ELSE p.campaign_id
                END AS campaign_id,
                CASE
                    WHEN (GROUPING(f.id) = 1) THEN NULL::bigint
                    ELSE f.id
                END AS field_id,
            (COALESCE(sum(
                CASE
                    WHEN ((l.tons IS NOT NULL) AND (l.tons > (0)::numeric)) THEN l.hectares
                    ELSE (0)::double precision
                END), (0)::double precision))::numeric(14,2) AS harvested_area,
            (COALESCE(sum(l.hectares), (0)::double precision))::numeric(14,2) AS total_hectares
           FROM ((public.projects p
             LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
             LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
          WHERE (p.deleted_at IS NULL)
          GROUP BY GROUPING SETS ((p.customer_id, p.id, p.campaign_id, f.id), (p.customer_id, p.id, p.campaign_id), (p.customer_id, p.id), (p.customer_id), ())
        ), costs_agg AS (
         SELECT
                CASE
                    WHEN (GROUPING(p.customer_id) = 1) THEN NULL::bigint
                    ELSE p.customer_id
                END AS customer_id,
                CASE
                    WHEN (GROUPING(p.id) = 1) THEN NULL::bigint
                    ELSE p.id
                END AS project_id,
                CASE
                    WHEN (GROUPING(p.campaign_id) = 1) THEN NULL::bigint
                    ELSE p.campaign_id
                END AS campaign_id,
            (COALESCE(sum(COALESCE(p.admin_cost, (0)::numeric)), (0)::numeric))::numeric(14,2) AS budget_cost_usd,
            (COALESCE(sum(COALESCE(dc.labors_usd, (0)::numeric)), (0)::numeric))::numeric(14,2) AS executed_labors_usd,
            (COALESCE(sum(COALESCE(dc.supplies_usd, (0)::numeric)), (0)::numeric))::numeric(14,2) AS executed_supplies_usd,
            (COALESCE(sum(COALESCE(dc.direct_costs_usd, (0)::numeric)), (0)::numeric))::numeric(14,2) AS executed_costs_usd
           FROM (public.projects p
             LEFT JOIN v_direct_costs_by_project dc ON ((dc.project_id = p.id)))
          WHERE (p.deleted_at IS NULL)
          GROUP BY GROUPING SETS ((p.customer_id, p.id, p.campaign_id), (p.customer_id, p.id), (p.customer_id), ())
        ), operating_result AS (
         WITH income_by_project AS (
                 SELECT f.project_id,
                    (COALESCE(sum(vf.income_usd), (0)::numeric))::numeric(14,2) AS income_usd
                   FROM (v_income_by_field vf
                     JOIN public.fields f ON (((f.id = vf.field_id) AND (f.deleted_at IS NULL))))
                  GROUP BY f.project_id
                )
         SELECT
                CASE
                    WHEN (GROUPING(p.customer_id) = 1) THEN NULL::bigint
                    ELSE p.customer_id
                END AS customer_id,
                CASE
                    WHEN (GROUPING(p.id) = 1) THEN NULL::bigint
                    ELSE p.id
                END AS project_id,
                CASE
                    WHEN (GROUPING(p.campaign_id) = 1) THEN NULL::bigint
                    ELSE p.campaign_id
                END AS campaign_id,
            (COALESCE(sum(COALESCE(ip.income_usd, (0)::numeric)), (0)::numeric))::numeric(14,2) AS income_usd,
            (COALESCE(sum(COALESCE(el.labors_usd, (0)::numeric)), (0)::numeric))::numeric(14,2) AS direct_labors_usd,
            (COALESCE(sum(COALESCE(us.supplies_usd, (0)::double precision)), (0)::double precision))::numeric(14,2) AS direct_supplies_usd,
            (((COALESCE(sum(COALESCE(el.labors_usd, (0)::numeric)), (0)::numeric))::double precision + COALESCE(sum(COALESCE(us.supplies_usd, (0)::double precision)), (0)::double precision)))::numeric(14,2) AS total_invested_usd
           FROM (((public.projects p
             LEFT JOIN income_by_project ip ON ((ip.project_id = p.id)))
             LEFT JOIN executed_labors_by_project el ON ((el.project_id = p.id)))
             LEFT JOIN used_supplies_by_project us ON ((us.project_id = p.id)))
          WHERE (p.deleted_at IS NULL)
          GROUP BY GROUPING SETS ((p.customer_id, p.id, p.campaign_id), (p.customer_id, p.id), (p.customer_id), ())
        )
 SELECT lvl.customer_id,
    lvl.project_id,
    lvl.campaign_id,
    lvl.field_id,
    (COALESCE(s.sowed_area, (0)::numeric))::numeric(14,2) AS sowing_hectares,
    (COALESCE(s.total_hectares, (0)::numeric))::numeric(14,2) AS sowing_total_hectares,
    (COALESCE(h.harvested_area, (0)::numeric))::numeric(14,2) AS harvest_hectares,
    (COALESCE(h.total_hectares, (0)::numeric))::numeric(14,2) AS harvest_total_hectares,
    (COALESCE(ca.executed_costs_usd, (0)::numeric))::numeric(14,2) AS executed_costs_usd,
    (COALESCE(ca.executed_labors_usd, (0)::numeric))::numeric(14,2) AS executed_labors_usd,
    (COALESCE(ca.executed_supplies_usd, (0)::numeric))::numeric(14,2) AS executed_supplies_usd,
    (COALESCE(ca.budget_cost_usd, (0)::numeric))::numeric(14,2) AS budget_cost_usd,
    (COALESCE(o.income_usd, (0)::numeric))::numeric(14,2) AS income_usd,
    ((COALESCE(o.income_usd, (0)::numeric) - COALESCE(o.total_invested_usd, (0)::numeric)))::numeric(14,2) AS operating_result_usd,
        CASE
            WHEN (COALESCE(o.total_invested_usd, (0)::numeric) > (0)::numeric) THEN round((((COALESCE(o.income_usd, (0)::numeric) - COALESCE(o.total_invested_usd, (0)::numeric)) / NULLIF(o.total_invested_usd, (0)::numeric)) * (100)::numeric), 2)
            ELSE (0)::numeric
        END AS operating_result_pct,
    ((COALESCE(ca.executed_costs_usd, (0)::numeric) + COALESCE(ca.budget_cost_usd, (0)::numeric)))::numeric(14,2) AS operating_result_total_costs_usd,
    (0)::numeric(14,2) AS contributions_progress_pct,
    'metric'::text AS row_kind
   FROM ((((levels lvl
     LEFT JOIN sowing s ON (((NOT (s.customer_id IS DISTINCT FROM lvl.customer_id)) AND (NOT (s.project_id IS DISTINCT FROM lvl.project_id)) AND (NOT (s.campaign_id IS DISTINCT FROM lvl.campaign_id)) AND (NOT (s.field_id IS DISTINCT FROM lvl.field_id)))))
     LEFT JOIN harvest h ON (((NOT (h.customer_id IS DISTINCT FROM lvl.customer_id)) AND (NOT (h.project_id IS DISTINCT FROM lvl.project_id)) AND (NOT (h.campaign_id IS DISTINCT FROM lvl.campaign_id)) AND (NOT (h.field_id IS DISTINCT FROM lvl.field_id)))))
     LEFT JOIN costs_agg ca ON (((NOT (ca.customer_id IS DISTINCT FROM lvl.customer_id)) AND (NOT (ca.project_id IS DISTINCT FROM lvl.project_id)) AND (NOT (ca.campaign_id IS DISTINCT FROM lvl.campaign_id)))))
     LEFT JOIN operating_result o ON (((NOT (o.customer_id IS DISTINCT FROM lvl.customer_id)) AND (NOT (o.project_id IS DISTINCT FROM lvl.project_id)) AND (NOT (o.campaign_id IS DISTINCT FROM lvl.campaign_id)))));


--
-- Name: engineering_principles_documentation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.engineering_principles_documentation (
    id integer NOT NULL,
    principle character varying(50) NOT NULL,
    description text NOT NULL,
    implementation text NOT NULL,
    migration_affected text[] NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: engineering_principles_documentation_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.engineering_principles_documentation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: engineering_principles_documentation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.engineering_principles_documentation_id_seq OWNED BY public.engineering_principles_documentation.id;


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
-- Name: fix_labors_list; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.fix_labors_list AS
 SELECT w.id AS workorder_id,
    w.number AS workorder_number,
    w.date,
    p.id AS project_id,
    f.id AS field_id,
    p.name AS project_name,
    f.name AS field_name,
    c.name AS crop_name,
    lb.name AS labor_name,
    lc.name AS category_name,
    w.contractor,
    w.effective_area AS surface_ha,
    lb.price AS cost_ha,
    lb.contractor_name,
    inv.name AS investor_name,
    i.id AS invoice_id,
    i.number AS invoice_number,
    i.company AS invoice_company,
    i.date AS invoice_date,
    i.status AS invoice_status
   FROM (((((((public.workorders w
     JOIN public.projects p ON ((w.project_id = p.id)))
     JOIN public.fields f ON ((w.field_id = f.id)))
     JOIN public.crops c ON ((w.crop_id = c.id)))
     JOIN public.labors lb ON ((w.labor_id = lb.id)))
     JOIN public.categories lc ON ((lb.category_id = lc.id)))
     JOIN public.investors inv ON ((w.investor_id = inv.id)))
     LEFT JOIN public.invoices i ON ((i.work_order_id = w.id)))
  WHERE ((w.deleted_at IS NULL) AND (p.deleted_at IS NULL));


--
-- Name: fix_lot_list; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.fix_lot_list AS
 SELECT f.project_id,
    l.field_id,
    p.name AS project_name,
    f.name AS field_name,
    l.id,
    l.name AS lot_name,
    l.variety,
        CASE
            WHEN (l.sowing_date IS NOT NULL) THEN l.hectares
            ELSE (0)::double precision
        END AS sowed_area,
    l.hectares,
    l.season,
    l.updated_at,
    l.tons,
    l.previous_crop_id,
    pc.name AS previous_crop,
    l.current_crop_id,
    cc_crop.name AS current_crop,
    bac.admin_cost_per_ha,
    byc.hectares AS harvested_area,
        CASE
            WHEN ((l.tons IS NOT NULL) AND (l.tons > (0)::numeric)) THEN CURRENT_DATE
            ELSE NULL::date
        END AS harvest_date,
    (bdc.direct_cost / NULLIF(l.hectares, (0)::double precision)) AS cost_usd_per_ha,
    byc.yield_tn_per_ha,
    bin.income_net_per_ha,
    blc.rent_per_ha,
    bat.active_total_per_ha,
    bor.operating_result_per_ha
   FROM (((((((((((public.lots l
     JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
     JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
     LEFT JOIN public.crops pc ON (((pc.id = l.previous_crop_id) AND (pc.deleted_at IS NULL))))
     LEFT JOIN public.crops cc_crop ON (((cc_crop.id = l.current_crop_id) AND (cc_crop.deleted_at IS NULL))))
     LEFT JOIN public.base_direct_costs_view bdc ON ((bdc.lot_id = l.id)))
     LEFT JOIN public.base_yield_calculations_view byc ON ((byc.lot_id = l.id)))
     LEFT JOIN public.base_income_net_view bin ON ((bin.lot_id = l.id)))
     LEFT JOIN public.base_admin_costs_view bac ON ((bac.lot_id = l.id)))
     LEFT JOIN public.base_lease_calculations_view blc ON ((blc.lot_id = l.id)))
     LEFT JOIN public.base_active_total_view bat ON ((bat.lot_id = l.id)))
     LEFT JOIN public.base_operating_result_view bor ON ((bor.lot_id = l.id)))
  WHERE (l.deleted_at IS NULL);


--
-- Name: fix_lots_metrics; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.fix_lots_metrics AS
 WITH lot_base AS (
         SELECT f.project_id,
            l.field_id,
            l.current_crop_id,
                CASE
                    WHEN (l.sowing_date IS NOT NULL) THEN l.hectares
                    ELSE (0)::double precision
                END AS seeded_area_lot,
                CASE
                    WHEN ((l.tons IS NOT NULL) AND (l.tons > (0)::numeric)) THEN l.hectares
                    ELSE (0)::double precision
                END AS harvested_area_lot,
            COALESCE(l.tons, (0)::numeric) AS tons_lot,
            COALESCE(bdc.direct_cost, (0)::double precision) AS direct_cost_lot
           FROM ((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             LEFT JOIN public.base_direct_costs_view bdc ON ((bdc.lot_id = l.id)))
          WHERE (l.deleted_at IS NULL)
        )
 SELECT project_id,
    field_id,
    current_crop_id,
    sum(seeded_area_lot) AS seeded_area,
    sum(harvested_area_lot) AS harvested_area,
        CASE
            WHEN (sum(harvested_area_lot) > (0)::double precision) THEN ((sum(tons_lot))::double precision / sum(harvested_area_lot))
            ELSE (0)::double precision
        END AS yield_tn_per_ha,
        CASE
            WHEN (sum(seeded_area_lot) > (0)::double precision) THEN (sum(direct_cost_lot) / sum(seeded_area_lot))
            ELSE (0)::double precision
        END AS cost_per_ha
   FROM lot_base b
  GROUP BY project_id, field_id, current_crop_id;


--
-- Name: flyway_schema_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.flyway_schema_history (
    installed_rank integer NOT NULL,
    version character varying(50),
    description character varying(200) NOT NULL,
    type character varying(20) NOT NULL,
    script character varying(1000) NOT NULL,
    checksum integer,
    installed_by character varying(100) NOT NULL,
    installed_on timestamp without time zone DEFAULT now() NOT NULL,
    execution_time integer NOT NULL,
    success boolean NOT NULL
);


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
-- Name: investor_contribution_data_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.investor_contribution_data_view AS
 SELECT p.id AS project_id,
    p.name AS project_name,
    p.customer_id,
    c.name AS customer_name,
    p.campaign_id,
    cam.name AS campaign_name,
    100.0 AS surface_total_ha,
    0.0 AS lease_fixed_usd,
    false AS lease_is_fixed,
    0.0 AS admin_per_ha_usd,
    COALESCE(p.admin_cost, (0)::numeric) AS admin_total_usd,
    ( SELECT COALESCE(jsonb_agg(jsonb_build_object('type', cat_costs.name, 'label', cat_costs.name, 'total_usd', cat_costs.total_cost, 'total_usd_ha',
                CASE
                    WHEN (100.0 > (0)::numeric) THEN (cat_costs.total_cost / (100.0)::double precision)
                    ELSE (0)::double precision
                END, 'investors', '[]'::jsonb, 'requires_manual_attribution', false)), '[]'::jsonb) AS "coalesce"
           FROM ( SELECT cat.name,
                    sum(((wi.total_used)::double precision * s.price)) AS total_cost
                   FROM (((public.workorders w2
                     JOIN public.workorder_items wi ON ((w2.id = wi.workorder_id)))
                     JOIN public.supplies s ON (((wi.supply_id = s.id) AND (s.deleted_at IS NULL))))
                     JOIN public.categories cat ON ((s.category_id = cat.id)))
                  WHERE ((w2.project_id = p.id) AND (w2.deleted_at IS NULL))
                  GROUP BY cat.id, cat.name) cat_costs) AS contributions_data,
    ( SELECT COALESCE(jsonb_agg(jsonb_build_object('investor_id', pi2.investor_id, 'investor_name', i2.name, 'agreed_share_pct', pi2.percentage, 'agreed_usd', (project_costs.total_project_cost * ((pi2.percentage / 100))::double precision), 'actual_usd', (project_costs.total_project_cost * ((pi2.percentage / 100))::double precision), 'adjustment_usd', 0)), '[]'::jsonb) AS "coalesce"
           FROM ((public.project_investors pi2
             JOIN public.investors i2 ON ((pi2.investor_id = i2.id)))
             CROSS JOIN ( SELECT COALESCE(sum(((wi.total_used)::double precision * s.price)), (0)::double precision) AS total_project_cost
                   FROM ((public.workorders w3
                     JOIN public.workorder_items wi ON ((w3.id = wi.workorder_id)))
                     JOIN public.supplies s ON (((wi.supply_id = s.id) AND (s.deleted_at IS NULL))))
                  WHERE ((w3.project_id = p.id) AND (w3.deleted_at IS NULL))) project_costs)
          WHERE (pi2.project_id = p.id)) AS comparison_data,
    jsonb_build_object('total_harvest_usd', COALESCE(( SELECT sum((cc.net_price * 100.0)) AS sum
           FROM public.crop_commercializations cc
          WHERE (cc.project_id = p.id)), (0)::numeric), 'total_harvest_usd_ha',
        CASE
            WHEN (100.0 > (0)::numeric) THEN (COALESCE(( SELECT sum((cc.net_price * 100.0)) AS sum
               FROM public.crop_commercializations cc
              WHERE (cc.project_id = p.id)), (0)::numeric) / 100.0)
            ELSE (0)::numeric
        END, 'investors', COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', pi2.investor_id, 'investor_name', i2.name, 'paid_usd', (COALESCE(( SELECT sum((cc.net_price * 100.0)) AS sum
                   FROM public.crop_commercializations cc
                  WHERE (cc.project_id = p.id)), (0)::numeric) * ((pi2.percentage / 100))::numeric), 'agreed_usd', (COALESCE(( SELECT sum((cc.net_price * 100.0)) AS sum
                   FROM public.crop_commercializations cc
                  WHERE (cc.project_id = p.id)), (0)::numeric) * ((pi2.percentage / 100))::numeric), 'adjustment_usd', 0)) AS jsonb_agg
           FROM (public.project_investors pi2
             JOIN public.investors i2 ON ((pi2.investor_id = i2.id)))
          WHERE (pi2.project_id = p.id)), '[]'::jsonb)) AS harvest_data
   FROM ((public.projects p
     JOIN public.customers c ON ((p.customer_id = c.id)))
     JOIN public.campaigns cam ON ((p.campaign_id = cam.id)))
  WHERE (p.deleted_at IS NULL);


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
-- Name: labor_cards_cube_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.labor_cards_cube_view AS
 SELECT w.project_id,
    w.field_id,
        CASE
            WHEN ((GROUPING(w.project_id) = 0) AND (GROUPING(w.field_id) = 0)) THEN 'project+field'::text
            WHEN ((GROUPING(w.project_id) = 0) AND (GROUPING(w.field_id) = 1)) THEN 'project'::text
            WHEN ((GROUPING(w.project_id) = 1) AND (GROUPING(w.field_id) = 0)) THEN 'field'::text
            ELSE 'global'::text
        END AS level,
    sum(w.effective_area) AS surface_ha,
    sum(lb.price) AS total_labor_cost,
        CASE
            WHEN (sum(w.effective_area) > (0)::numeric) THEN (sum((lb.price * w.effective_area)) / sum(w.effective_area))
            ELSE (0)::numeric
        END AS labor_cost_per_ha
   FROM (public.workorders w
     JOIN public.labors lb ON ((lb.id = w.labor_id)))
  WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric) AND (lb.price IS NOT NULL))
  GROUP BY GROUPING SETS ((w.project_id, w.field_id), (w.project_id), (w.field_id), ());


--
-- Name: labor_cards_cube_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.labor_cards_cube_view_v2 AS
 WITH labor_metrics AS (
         SELECT bdc.project_id,
            bdc.field_id,
            sum(w.effective_area) AS surface_ha,
            sum(bdc.labor_cost) AS total_labor_cost,
                CASE
                    WHEN (sum(w.effective_area) > (0)::numeric) THEN (sum(bdc.labor_cost) / sum(w.effective_area))
                    ELSE (0)::numeric
                END AS labor_cost_per_ha
           FROM (public.base_direct_costs_view bdc
             JOIN public.workorders w ON (((w.project_id = bdc.project_id) AND (w.field_id = bdc.field_id) AND (w.lot_id = bdc.lot_id))))
          WHERE ((w.deleted_at IS NULL) AND (w.effective_area > (0)::numeric))
          GROUP BY bdc.project_id, bdc.field_id
        )
 SELECT labor_metrics.project_id,
    labor_metrics.field_id,
    'project+field'::text AS level,
    labor_metrics.surface_ha,
    labor_metrics.total_labor_cost AS net_total_cost,
    labor_metrics.labor_cost_per_ha AS avg_cost_per_ha
   FROM labor_metrics
UNION ALL
 SELECT labor_metrics.project_id,
    NULL::bigint AS field_id,
    'project'::text AS level,
    sum(labor_metrics.surface_ha) AS surface_ha,
    sum(labor_metrics.total_labor_cost) AS net_total_cost,
        CASE
            WHEN (sum(labor_metrics.surface_ha) > (0)::numeric) THEN (sum(labor_metrics.total_labor_cost) / sum(labor_metrics.surface_ha))
            ELSE (0)::numeric
        END AS avg_cost_per_ha
   FROM labor_metrics
  GROUP BY labor_metrics.project_id
UNION ALL
 SELECT NULL::bigint AS project_id,
    NULL::bigint AS field_id,
    'global'::text AS level,
    sum(labor_metrics.surface_ha) AS surface_ha,
    sum(labor_metrics.total_labor_cost) AS net_total_cost,
        CASE
            WHEN (sum(labor_metrics.surface_ha) > (0)::numeric) THEN (sum(labor_metrics.total_labor_cost) / sum(labor_metrics.surface_ha))
            ELSE (0)::numeric
        END AS avg_cost_per_ha
   FROM labor_metrics;


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
-- Name: labor_metrics_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.labor_metrics_view AS
 WITH workorder_base AS (
         SELECT w.id AS workorder_id,
            w.project_id,
            w.field_id,
            w.effective_area,
            w.labor_id,
            lb.price AS labor_price_per_ha,
            (lb.price * w.effective_area) AS labor_cost_per_wo
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric) AND (lb.price IS NOT NULL))
        ), supply_consumption AS (
         SELECT w.id AS workorder_id,
            w.project_id,
            w.field_id,
            w.effective_area,
            COALESCE(wi.total_used, (wi.final_dose * w.effective_area), (0)::numeric) AS qty_used,
                CASE
                    WHEN (s.unit_id = 1) THEN 'LITERS'::text
                    WHEN (s.unit_id = 2) THEN 'KILOS'::text
                    ELSE 'OTHER'::text
                END AS unit_category,
            (COALESCE(s.price, (0)::double precision) * (COALESCE(wi.total_used, (wi.final_dose * w.effective_area), (0)::numeric))::double precision) AS supply_cost_per_wo
           FROM ((public.workorders w
             JOIN public.workorder_items wi ON (((wi.workorder_id = w.id) AND (wi.deleted_at IS NULL))))
             JOIN public.supplies s ON (((s.id = wi.supply_id) AND (s.deleted_at IS NULL))))
          WHERE ((w.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
        ), workorder_metrics AS (
         SELECT wb.workorder_id,
            wb.project_id,
            wb.field_id,
            wb.effective_area,
            wb.labor_cost_per_wo,
            sum(
                CASE
                    WHEN (sc.unit_category = 'LITERS'::text) THEN COALESCE(sc.qty_used, (0)::numeric)
                    ELSE (0)::numeric
                END) AS liters_used,
            sum(
                CASE
                    WHEN (sc.unit_category = 'KILOS'::text) THEN COALESCE(sc.qty_used, (0)::numeric)
                    ELSE (0)::numeric
                END) AS kilos_used,
            sum(COALESCE(sc.supply_cost_per_wo, (0)::double precision)) AS total_supplies_cost,
            ((wb.labor_cost_per_wo)::double precision + sum(COALESCE(sc.supply_cost_per_wo, (0)::double precision))) AS total_workorder_cost
           FROM (workorder_base wb
             LEFT JOIN supply_consumption sc ON ((sc.workorder_id = wb.workorder_id)))
          GROUP BY wb.workorder_id, wb.project_id, wb.field_id, wb.effective_area, wb.labor_cost_per_wo
        ), field_metrics AS (
         SELECT wm.project_id,
            wm.field_id,
            sum(wm.effective_area) AS surface_ha,
            sum(wm.liters_used) AS total_liters,
            sum(wm.kilos_used) AS total_kilos,
            sum(wm.labor_cost_per_wo) AS total_labor_cost,
            sum(wm.total_supplies_cost) AS total_supplies_cost,
            sum(wm.total_workorder_cost) AS net_total_cost,
            count(DISTINCT wm.workorder_id) AS total_workorders,
                CASE
                    WHEN (sum(wm.effective_area) > (0)::numeric) THEN (sum(wm.total_workorder_cost) / (sum(wm.effective_area))::double precision)
                    ELSE (0)::double precision
                END AS avg_cost_per_ha,
                CASE
                    WHEN (sum(wm.effective_area) > (0)::numeric) THEN (sum(wm.liters_used) / sum(wm.effective_area))
                    ELSE (0)::numeric
                END AS liters_per_ha,
                CASE
                    WHEN (sum(wm.effective_area) > (0)::numeric) THEN (sum(wm.kilos_used) / sum(wm.effective_area))
                    ELSE (0)::numeric
                END AS kilos_per_ha
           FROM workorder_metrics wm
          GROUP BY wm.project_id, wm.field_id
        )
 SELECT project_id,
    field_id,
    surface_ha,
    total_liters,
    total_kilos,
    net_total_cost,
    avg_cost_per_ha,
    total_labor_cost,
    total_supplies_cost,
    total_workorders,
    liters_per_ha,
    kilos_per_ha
   FROM field_metrics fm
  WHERE (surface_ha > (0)::numeric);


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
    id integer NOT NULL,
    lot_id bigint NOT NULL,
    sowing_date date,
    harvest_date date,
    sequence smallint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint,
    CONSTRAINT lot_dates_sequence_check CHECK (((sequence >= 1) AND (sequence <= 3)))
);


--
-- Name: lot_dates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lot_dates_id_seq
    AS integer
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
-- Name: lot_metrics_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.lot_metrics_view AS
 WITH labor_agg AS (
         SELECT w.lot_id,
            sum(w.effective_area) FILTER (WHERE (lb.category_id = 9)) AS seeded_area_lot,
            sum(w.effective_area) FILTER (WHERE (lb.category_id = 13)) AS harvested_area_lot,
            sum((COALESCE(lb.price, (0)::numeric) * w.effective_area)) AS labor_cost_lot
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
          GROUP BY w.lot_id
        ), supply_agg AS (
         SELECT w.lot_id,
            sum((((COALESCE(wi.final_dose, (0)::numeric))::double precision * COALESCE(s.price, (0)::double precision)) * (w.effective_area)::double precision)) AS supply_cost_lot
           FROM ((public.workorders w
             JOIN public.workorder_items wi ON (((wi.workorder_id = w.id) AND (wi.deleted_at IS NULL))))
             JOIN public.supplies s ON (((s.id = wi.supply_id) AND (s.deleted_at IS NULL))))
          WHERE ((w.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
          GROUP BY w.lot_id
        ), lot_base AS (
         SELECT f.project_id,
            l.field_id,
            l.current_crop_id,
            COALESCE(la.seeded_area_lot, (0)::numeric) AS seeded_area_lot,
            COALESCE(la.harvested_area_lot, (0)::numeric) AS harvested_area_lot,
            COALESCE(l.tons, (0)::numeric) AS tons_lot,
            ((COALESCE(la.labor_cost_lot, (0)::numeric))::double precision + COALESCE(sa.supply_cost_lot, (0)::double precision)) AS direct_cost_lot
           FROM (((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             LEFT JOIN labor_agg la ON ((la.lot_id = l.id)))
             LEFT JOIN supply_agg sa ON ((sa.lot_id = l.id)))
          WHERE (l.deleted_at IS NULL)
        )
 SELECT project_id,
    field_id,
    current_crop_id,
    sum(seeded_area_lot) AS seeded_area,
    sum(harvested_area_lot) AS harvested_area,
        CASE
            WHEN (sum(harvested_area_lot) > (0)::numeric) THEN (sum(tons_lot) / sum(harvested_area_lot))
            ELSE (0)::numeric
        END AS yield_tn_per_ha,
        CASE
            WHEN (sum(seeded_area_lot) > (0)::numeric) THEN (sum(direct_cost_lot) / (sum(seeded_area_lot))::double precision)
            ELSE (0)::double precision
        END AS cost_per_ha
   FROM lot_base b
  GROUP BY project_id, field_id, current_crop_id;


--
-- Name: lot_table_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.lot_table_view AS
 WITH sowing AS (
         SELECT w.lot_id,
            sum(w.effective_area) AS sowed_area
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (lb.category_id = 9) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
          GROUP BY w.lot_id
        ), harvest AS (
         SELECT w.lot_id,
            sum(w.effective_area) AS harvested_area,
            max(w.date) AS last_harvest_date
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (lb.category_id = 13) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
          GROUP BY w.lot_id
        ), direct_costs AS (
         SELECT w.lot_id,
            sum(COALESCE((lb.price * w.effective_area), (0)::numeric)) AS labor_cost,
            sum(COALESCE((((wi.final_dose)::double precision * s_1.price) * (w.effective_area)::double precision), (0)::double precision)) AS supply_cost
           FROM (((public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
             LEFT JOIN public.workorder_items wi ON (((w.id = wi.workorder_id) AND (wi.deleted_at IS NULL))))
             LEFT JOIN public.supplies s_1 ON (((s_1.id = wi.supply_id) AND (s_1.deleted_at IS NULL))))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
          GROUP BY w.lot_id
        ), income_net AS (
         SELECT l_1.id AS lot_id,
            (COALESCE(l_1.tons, (0)::numeric) * COALESCE(cc_1.net_price, (0)::numeric)) AS income_net_total
           FROM ((public.lots l_1
             JOIN public.fields f_1 ON (((f_1.id = l_1.field_id) AND (f_1.deleted_at IS NULL))))
             LEFT JOIN public.crop_commercializations cc_1 ON (((cc_1.project_id = f_1.project_id) AND (cc_1.crop_id = l_1.current_crop_id) AND (cc_1.deleted_at IS NULL))))
          WHERE ((l_1.deleted_at IS NULL) AND (l_1.tons IS NOT NULL) AND (l_1.tons > (0)::numeric))
        ), rent_calculation AS (
         SELECT l_1.id AS lot_id,
                CASE
                    WHEN (f_1.lease_type_id = 1) THEN (COALESCE(f_1.lease_type_value, (0)::double precision) * (COALESCE(h_1.harvested_area, (0)::numeric))::double precision)
                    WHEN (f_1.lease_type_id = 2) THEN ((COALESCE(f_1.lease_type_percent, (0)::double precision) / (100.0)::double precision) * (COALESCE(in_net_1.income_net_total, (0)::numeric))::double precision)
                    WHEN (f_1.lease_type_id = 3) THEN ((COALESCE(f_1.lease_type_value, (0)::double precision) * (COALESCE(h_1.harvested_area, (0)::numeric))::double precision) + ((COALESCE(f_1.lease_type_percent, (0)::double precision) / (100.0)::double precision) * (COALESCE(in_net_1.income_net_total, (0)::numeric))::double precision))
                    ELSE (0)::double precision
                END AS rent_total
           FROM (((public.lots l_1
             JOIN public.fields f_1 ON (((f_1.id = l_1.field_id) AND (f_1.deleted_at IS NULL))))
             LEFT JOIN harvest h_1 ON ((h_1.lot_id = l_1.id)))
             LEFT JOIN income_net in_net_1 ON ((in_net_1.lot_id = l_1.id)))
          WHERE (l_1.deleted_at IS NULL)
        ), admin_cost AS (
         SELECT l_1.id AS lot_id,
            ((COALESCE(p_1.admin_cost, (0)::numeric))::double precision * COALESCE(l_1.hectares, (0)::double precision)) AS admin_total
           FROM ((public.lots l_1
             JOIN public.fields f_1 ON (((f_1.id = l_1.field_id) AND (f_1.deleted_at IS NULL))))
             JOIN public.projects p_1 ON (((p_1.id = f_1.project_id) AND (p_1.deleted_at IS NULL))))
          WHERE ((l_1.deleted_at IS NULL) AND (l_1.hectares IS NOT NULL) AND (l_1.hectares > (0)::double precision))
        ), lot_dates AS (
         SELECT w.lot_id,
            min(
                CASE
                    WHEN (lb.category_id = 9) THEN w.date
                    ELSE NULL::date
                END) AS lot_sowing_date,
            max(
                CASE
                    WHEN (lb.category_id = 13) THEN w.date
                    ELSE NULL::date
                END) AS lot_harvest_date,
            count(DISTINCT w.id) AS sequence
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (lb.category_id = ANY (ARRAY[9, 13])) AND (w.date IS NOT NULL))
          GROUP BY w.lot_id
        )
 SELECT l.id,
    f.project_id,
    l.field_id,
    p.name AS project_name,
    f.name AS field_name,
    l.name AS lot_name,
    pc.name AS previous_crop,
    l.previous_crop_id,
    cc.name AS current_crop,
    l.current_crop_id,
    l.variety,
    l.hectares,
    COALESCE(s.sowed_area, (0)::numeric) AS sowed_area,
    l.season,
    COALESCE(l.tons, (0)::numeric) AS tons,
    COALESCE(h.harvested_area, (0)::numeric) AS harvested_area,
    h.last_harvest_date AS harvest_date,
    ((COALESCE(dc.labor_cost, (0)::numeric))::double precision + COALESCE(dc.supply_cost, (0)::double precision)) AS direct_cost_total,
        CASE
            WHEN (COALESCE(s.sowed_area, (0)::numeric) > (0)::numeric) THEN (((COALESCE(dc.labor_cost, (0)::numeric))::double precision + COALESCE(dc.supply_cost, (0)::double precision)) / (s.sowed_area)::double precision)
            ELSE (0)::double precision
        END AS cost_usd_per_ha,
    COALESCE(in_net.income_net_total, (0)::numeric) AS income_net_total,
        CASE
            WHEN (COALESCE(s.sowed_area, (0)::numeric) > (0)::numeric) THEN (COALESCE(in_net.income_net_total, (0)::numeric) / s.sowed_area)
            ELSE (0)::numeric
        END AS income_net_per_ha,
        CASE
            WHEN (COALESCE(h.harvested_area, (0)::numeric) > (0)::numeric) THEN (COALESCE(l.tons, (0)::numeric) / h.harvested_area)
            ELSE (0)::numeric
        END AS yield_tn_per_ha,
    COALESCE(rc.rent_total, (0)::double precision) AS rent_total,
        CASE
            WHEN (COALESCE(s.sowed_area, (0)::numeric) > (0)::numeric) THEN (COALESCE(rc.rent_total, (0)::double precision) / (s.sowed_area)::double precision)
            ELSE (0)::double precision
        END AS rent_per_ha,
    COALESCE(ac.admin_total, (0)::double precision) AS admin_total,
        CASE
            WHEN (COALESCE(s.sowed_area, (0)::numeric) > (0)::numeric) THEN (COALESCE(ac.admin_total, (0)::double precision) / (s.sowed_area)::double precision)
            ELSE (0)::double precision
        END AS admin_cost_per_ha,
    ((((COALESCE(dc.labor_cost, (0)::numeric))::double precision + COALESCE(dc.supply_cost, (0)::double precision)) + COALESCE(rc.rent_total, (0)::double precision)) + COALESCE(ac.admin_total, (0)::double precision)) AS active_total,
        CASE
            WHEN (COALESCE(s.sowed_area, (0)::numeric) > (0)::numeric) THEN (((((COALESCE(dc.labor_cost, (0)::numeric))::double precision + COALESCE(dc.supply_cost, (0)::double precision)) + COALESCE(rc.rent_total, (0)::double precision)) + COALESCE(ac.admin_total, (0)::double precision)) / (s.sowed_area)::double precision)
            ELSE (0)::double precision
        END AS active_total_per_ha,
    ((COALESCE(in_net.income_net_total, (0)::numeric))::double precision - ((((COALESCE(dc.labor_cost, (0)::numeric))::double precision + COALESCE(dc.supply_cost, (0)::double precision)) + COALESCE(rc.rent_total, (0)::double precision)) + COALESCE(ac.admin_total, (0)::double precision))) AS operating_result,
        CASE
            WHEN (COALESCE(s.sowed_area, (0)::numeric) > (0)::numeric) THEN (((COALESCE(in_net.income_net_total, (0)::numeric))::double precision - ((((COALESCE(dc.labor_cost, (0)::numeric))::double precision + COALESCE(dc.supply_cost, (0)::double precision)) + COALESCE(rc.rent_total, (0)::double precision)) + COALESCE(ac.admin_total, (0)::double precision))) / (s.sowed_area)::double precision)
            ELSE (0)::double precision
        END AS operating_result_per_ha,
    ld.lot_sowing_date,
    ld.lot_harvest_date,
    COALESCE(ld.sequence, (0)::bigint) AS sequence,
    l.updated_at
   FROM (((((((((((public.lots l
     JOIN public.fields f ON (((l.field_id = f.id) AND (f.deleted_at IS NULL))))
     JOIN public.projects p ON (((f.project_id = p.id) AND (p.deleted_at IS NULL))))
     LEFT JOIN public.crops pc ON (((l.previous_crop_id = pc.id) AND (pc.deleted_at IS NULL))))
     LEFT JOIN public.crops cc ON (((l.current_crop_id = cc.id) AND (cc.deleted_at IS NULL))))
     LEFT JOIN sowing s ON ((l.id = s.lot_id)))
     LEFT JOIN harvest h ON ((l.id = h.lot_id)))
     LEFT JOIN direct_costs dc ON ((l.id = dc.lot_id)))
     LEFT JOIN income_net in_net ON ((l.id = in_net.lot_id)))
     LEFT JOIN rent_calculation rc ON ((l.id = rc.lot_id)))
     LEFT JOIN admin_cost ac ON ((l.id = ac.lot_id)))
     LEFT JOIN lot_dates ld ON ((l.id = ld.lot_id)))
  WHERE ((l.deleted_at IS NULL) AND (l.hectares IS NOT NULL) AND (l.hectares > (0)::double precision));


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
-- Name: report_field_crop_metrics_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.report_field_crop_metrics_view_v2 AS
 WITH lot_crop_base AS (
         SELECT l.id AS lot_id,
            f.project_id,
            f.id AS field_id,
            f.name AS field_name,
            l.current_crop_id,
            c.name AS crop_name,
            l.hectares,
            l.tons,
            COALESCE(s.sowed_area, (0)::numeric) AS sowed_area,
            COALESCE(h.harvested_area, (0)::numeric) AS harvested_area
           FROM (((((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
             LEFT JOIN public.crops c ON (((c.id = l.current_crop_id) AND (c.deleted_at IS NULL))))
             LEFT JOIN ( SELECT w.lot_id,
                    sum(w.effective_area) AS sowed_area
                   FROM (public.workorders w
                     JOIN public.labors lb ON ((lb.id = w.labor_id)))
                  WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (lb.category_id = 9) AND (w.effective_area > (0)::numeric))
                  GROUP BY w.lot_id) s ON ((s.lot_id = l.id)))
             LEFT JOIN ( SELECT w.lot_id,
                    sum(w.effective_area) AS harvested_area
                   FROM (public.workorders w
                     JOIN public.labors lb ON ((lb.id = w.labor_id)))
                  WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (lb.category_id = 13) AND (w.effective_area > (0)::numeric))
                  GROUP BY w.lot_id) h ON ((h.lot_id = l.id)))
          WHERE ((l.deleted_at IS NULL) AND (l.hectares > (0)::double precision))
        ), lot_direct_costs AS (
         SELECT w.lot_id,
            sum(COALESCE((lb.price * w.effective_area), (0)::numeric)) AS labor_cost,
            sum(COALESCE((((wi.final_dose)::double precision * s.price) * (w.effective_area)::double precision), (0)::double precision)) AS supply_cost
           FROM (((public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
             LEFT JOIN public.workorder_items wi ON (((w.id = wi.workorder_id) AND (wi.deleted_at IS NULL))))
             LEFT JOIN public.supplies s ON (((s.id = wi.supply_id) AND (s.deleted_at IS NULL))))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL) AND (w.effective_area > (0)::numeric))
          GROUP BY w.lot_id
        ), crop_commercialization AS (
         SELECT cc_1.project_id,
            cc_1.crop_id,
            cc_1.board_price,
            cc_1.freight_cost,
            cc_1.commercial_cost,
            cc_1.net_price
           FROM public.crop_commercializations cc_1
          WHERE (cc_1.deleted_at IS NULL)
        ), lot_income AS (
         SELECT l.id AS lot_id,
            (COALESCE(l.tons, (0)::numeric) * COALESCE(cc_1.net_price, (0)::numeric)) AS income_net_total
           FROM ((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             LEFT JOIN crop_commercialization cc_1 ON (((cc_1.project_id = f.project_id) AND (cc_1.crop_id = l.current_crop_id))))
          WHERE (l.deleted_at IS NULL)
        ), lot_rent AS (
         SELECT l.id AS lot_id,
                CASE
                    WHEN (f.lease_type_id = 1) THEN (COALESCE(f.lease_type_value, (0)::double precision) * COALESCE(l.hectares, (0)::double precision))
                    WHEN (f.lease_type_id = 2) THEN ((COALESCE(f.lease_type_percent, (0)::double precision) / (100.0)::double precision) * (COALESCE(li_1.income_net_total, (0)::numeric))::double precision)
                    WHEN (f.lease_type_id = 3) THEN ((COALESCE(f.lease_type_value, (0)::double precision) * COALESCE(l.hectares, (0)::double precision)) + ((COALESCE(f.lease_type_percent, (0)::double precision) / (100.0)::double precision) * (COALESCE(li_1.income_net_total, (0)::numeric))::double precision))
                    ELSE (0)::double precision
                END AS rent_total
           FROM ((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             LEFT JOIN lot_income li_1 ON ((li_1.lot_id = l.id)))
          WHERE (l.deleted_at IS NULL)
        ), lot_admin_cost AS (
         SELECT l.id AS lot_id,
            ((COALESCE(p.admin_cost, (0)::numeric))::double precision * COALESCE(l.hectares, (0)::double precision)) AS admin_total
           FROM ((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
          WHERE ((l.deleted_at IS NULL) AND (l.hectares > (0)::double precision))
        )
 SELECT lcb.project_id,
    lcb.field_id,
    (lcb.field_name)::text AS field_name,
    lcb.current_crop_id,
    (lcb.crop_name)::text AS crop_name,
    (lcb.hectares)::text AS superficie_ha,
    (lcb.tons)::text AS produccion_tn,
    (lcb.sowed_area)::text AS area_sembrada_ha,
    (lcb.harvested_area)::text AS area_cosechada_ha,
        CASE
            WHEN (lcb.hectares > (0)::double precision) THEN (((lcb.tons)::double precision / lcb.hectares))::text
            ELSE '0'::text
        END AS rendimiento_tn_ha,
    (COALESCE(cc.board_price, (0)::numeric))::text AS precio_bruto_usd_tn,
    (COALESCE(cc.freight_cost, (0)::numeric))::text AS gasto_flete_usd_tn,
    (COALESCE(cc.commercial_cost, (0)::double precision))::text AS gasto_comercial_usd_tn,
    (COALESCE(cc.net_price, (0)::numeric))::text AS precio_neto_usd_tn,
    (COALESCE(li.income_net_total, (0)::numeric))::text AS ingreso_neto_usd,
        CASE
            WHEN (lcb.hectares > (0)::double precision) THEN (((COALESCE(li.income_net_total, (0)::numeric))::double precision / lcb.hectares))::text
            ELSE '0'::text
        END AS ingreso_neto_usd_ha,
    (COALESCE(ldc.labor_cost, (0)::numeric))::text AS costos_labores_usd,
    (COALESCE(ldc.supply_cost, (0)::double precision))::text AS costos_insumos_usd,
    (((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)))::text AS total_costos_directos_usd,
        CASE
            WHEN (lcb.hectares > (0)::double precision) THEN ((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) / lcb.hectares))::text
            ELSE '0'::text
        END AS costos_directos_usd_ha,
    (((COALESCE(li.income_net_total, (0)::numeric))::double precision - ((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision))))::text AS margen_bruto_usd,
        CASE
            WHEN (lcb.hectares > (0)::double precision) THEN ((((COALESCE(li.income_net_total, (0)::numeric))::double precision - ((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision))) / lcb.hectares))::text
            ELSE '0'::text
        END AS margen_bruto_usd_ha,
    (COALESCE(lr.rent_total, (0)::double precision))::text AS arriendo_usd,
        CASE
            WHEN (lcb.hectares > (0)::double precision) THEN ((COALESCE(lr.rent_total, (0)::double precision) / lcb.hectares))::text
            ELSE '0'::text
        END AS arriendo_usd_ha,
    (COALESCE(lac.admin_total, (0)::double precision))::text AS administracion_usd,
        CASE
            WHEN (lcb.hectares > (0)::double precision) THEN ((COALESCE(lac.admin_total, (0)::double precision) / lcb.hectares))::text
            ELSE '0'::text
        END AS administracion_usd_ha,
    (((COALESCE(li.income_net_total, (0)::numeric))::double precision - ((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) + COALESCE(lr.rent_total, (0)::double precision)) + COALESCE(lac.admin_total, (0)::double precision))))::text AS resultado_operativo_usd,
        CASE
            WHEN (lcb.hectares > (0)::double precision) THEN ((((COALESCE(li.income_net_total, (0)::numeric))::double precision - ((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) + COALESCE(lr.rent_total, (0)::double precision)) + COALESCE(lac.admin_total, (0)::double precision))) / lcb.hectares))::text
            ELSE '0'::text
        END AS resultado_operativo_usd_ha,
    (((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) + COALESCE(lr.rent_total, (0)::double precision)) + COALESCE(lac.admin_total, (0)::double precision)))::text AS total_invertido_usd,
        CASE
            WHEN (lcb.hectares > (0)::double precision) THEN ((((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) + COALESCE(lr.rent_total, (0)::double precision)) + COALESCE(lac.admin_total, (0)::double precision)) / lcb.hectares))::text
            ELSE '0'::text
        END AS total_invertido_usd_ha,
        CASE
            WHEN (((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) + COALESCE(lr.rent_total, (0)::double precision)) + COALESCE(lac.admin_total, (0)::double precision)) > (0)::double precision) THEN ((((COALESCE(li.income_net_total, (0)::numeric))::double precision - ((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) + COALESCE(lr.rent_total, (0)::double precision)) + COALESCE(lac.admin_total, (0)::double precision))) / ((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) + COALESCE(lr.rent_total, (0)::double precision)) + COALESCE(lac.admin_total, (0)::double precision))))::text
            ELSE '0'::text
        END AS renta_pct,
        CASE
            WHEN ((lcb.hectares > (0)::double precision) AND (lcb.tons > (0)::numeric)) THEN ((((((COALESCE(ldc.labor_cost, (0)::numeric))::double precision + COALESCE(ldc.supply_cost, (0)::double precision)) + COALESCE(lr.rent_total, (0)::double precision)) + COALESCE(lac.admin_total, (0)::double precision)) / ((lcb.tons)::double precision / lcb.hectares)))::text
            ELSE '0'::text
        END AS rinde_indiferencia_usd_tn
   FROM (((((lot_crop_base lcb
     LEFT JOIN lot_direct_costs ldc ON ((ldc.lot_id = lcb.lot_id)))
     LEFT JOIN crop_commercialization cc ON (((cc.project_id = lcb.project_id) AND (cc.crop_id = lcb.current_crop_id))))
     LEFT JOIN lot_income li ON ((li.lot_id = lcb.lot_id)))
     LEFT JOIN lot_rent lr ON ((lr.lot_id = lcb.lot_id)))
     LEFT JOIN lot_admin_cost lac ON ((lac.lot_id = lcb.lot_id)))
  WHERE (lcb.current_crop_id IS NOT NULL);


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
    refresh_tokens text[] DEFAULT ARRAY[]::text[],
    id_rol integer NOT NULL,
    is_verified boolean DEFAULT false,
    active boolean DEFAULT true,
    created_by integer NOT NULL,
    updated_by integer NOT NULL,
    deleted_by integer,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
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
-- Name: views_fixes; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.views_fixes AS
 SELECT 'fix_labors_list'::text AS fix_name,
    'Corrige duplicación de labores por múltiples meses de dólar promedio'::text AS description,
    'workorders'::text AS affected_table,
    'fix_labors_list_duplication'::text AS fix_type
UNION ALL
 SELECT 'placeholder_fix'::text AS fix_name,
    'Placeholder para futuros fixes de vistas'::text AS description,
    'various'::text AS affected_table,
    'placeholder'::text AS fix_type;


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
-- Name: workorder_list_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.workorder_list_view AS
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
    wi.final_dose AS dose,
    COALESCE(((wi.final_dose)::double precision * s.price), (0)::double precision) AS cost_per_ha,
    s.price AS unit_price,
    COALESCE((((wi.final_dose)::double precision * s.price) * (w.effective_area)::double precision), (0)::double precision) AS total_cost
   FROM ((((((((((public.workorders w
     JOIN public.projects p ON ((p.id = w.project_id)))
     JOIN public.fields f ON ((f.id = w.field_id)))
     JOIN public.lots l ON ((l.id = w.lot_id)))
     JOIN public.crops c ON ((c.id = w.crop_id)))
     JOIN public.labors lb ON ((lb.id = w.labor_id)))
     JOIN public.categories cat_lb ON ((cat_lb.id = lb.category_id)))
     LEFT JOIN public.workorder_items wi ON ((wi.workorder_id = w.id)))
     LEFT JOIN public.supplies s ON ((s.id = wi.supply_id)))
     LEFT JOIN public.types t ON ((t.id = s.type_id)))
     LEFT JOIN public.categories cat ON ((cat.id = s.category_id)))
  WHERE (w.deleted_at IS NULL);


--
-- Name: workorder_metrics_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.workorder_metrics_view AS
 SELECT w.project_id,
    w.field_id,
    p.customer_id,
    p.campaign_id,
    sum(w.effective_area) AS surface_ha,
    sum(COALESCE(wi.final_dose, (0)::numeric)) FILTER (WHERE (s.category_id = ANY (ARRAY[2, 4, 5, 6]))) AS liters,
    sum(COALESCE(wi.final_dose, (0)::numeric)) FILTER (WHERE (s.category_id = ANY (ARRAY[1, 3, 8]))) AS kilograms,
    ((( SELECT sum(labors.price) AS sum
           FROM public.labors
          WHERE (labors.id IN ( SELECT DISTINCT workorders.labor_id
                   FROM public.workorders
                  WHERE ((workorders.project_id = w.project_id) AND (workorders.field_id = w.field_id) AND (workorders.deleted_at IS NULL))))))::double precision + ( SELECT sum(supplies.price) AS sum
           FROM public.supplies
          WHERE (supplies.id IN ( SELECT DISTINCT wi2.supply_id
                   FROM (public.workorder_items wi2
                     JOIN public.workorders w2 ON ((wi2.workorder_id = w2.id)))
                  WHERE ((w2.project_id = w.project_id) AND (w2.field_id = w.field_id) AND (w2.deleted_at IS NULL) AND (wi2.deleted_at IS NULL)))))) AS direct_cost
   FROM (((public.workorders w
     JOIN public.projects p ON ((p.id = w.project_id)))
     LEFT JOIN public.workorder_items wi ON ((wi.workorder_id = w.id)))
     LEFT JOIN public.supplies s ON ((s.id = wi.supply_id)))
  WHERE ((w.deleted_at IS NULL) AND (p.deleted_at IS NULL) AND ((wi.deleted_at IS NULL) OR (wi.deleted_at IS NULL)) AND ((s.deleted_at IS NULL) OR (s.deleted_at IS NULL)) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric))
  GROUP BY w.project_id, w.field_id, p.customer_id, p.campaign_id;


--
-- Name: workorder_metrics_view_v2; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.workorder_metrics_view_v2 AS
 SELECT w.project_id,
    w.field_id,
    sum(w.effective_area) AS surface_ha,
    sum(
        CASE
            WHEN (s.unit_id = 1) THEN (wi.final_dose * w.effective_area)
            ELSE (0)::numeric
        END) AS liters,
    sum(
        CASE
            WHEN (s.unit_id = 2) THEN (wi.final_dose * w.effective_area)
            ELSE (0)::numeric
        END) AS kilograms,
    COALESCE(bdc.direct_cost, (0)::double precision) AS direct_cost
   FROM (((public.workorders w
     LEFT JOIN public.workorder_items wi ON ((wi.workorder_id = w.id)))
     LEFT JOIN public.supplies s ON ((s.id = wi.supply_id)))
     LEFT JOIN public.base_direct_costs_view bdc ON (((bdc.project_id = w.project_id) AND (bdc.field_id = w.field_id) AND (bdc.lot_id = w.lot_id))))
  WHERE (w.deleted_at IS NULL)
  GROUP BY w.project_id, w.field_id, bdc.direct_cost;


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
-- Name: workorder_metrics; Type: VIEW; Schema: report; Owner: -
--

CREATE VIEW report.workorder_metrics AS
 WITH base AS NOT MATERIALIZED (
         SELECT s.project_id,
            s.field_id,
            s.lot_id,
            s.surface_ha,
            p.dose_total
           FROM (calc_common.workorder_surface s
             JOIN calc_common.workorder_supply p USING (project_id, field_id, lot_id))
        )
 SELECT project_id,
    field_id,
    lot_id,
    surface_ha,
    dose_total,
    calc.norm_dose(dose_total, surface_ha) AS dose_per_ha
   FROM base b;


--
-- Name: workorder_metrics_mv; Type: MATERIALIZED VIEW; Schema: report; Owner: -
--

CREATE MATERIALIZED VIEW report.workorder_metrics_mv AS
 SELECT project_id,
    field_id,
    lot_id,
    surface_ha,
    dose_total,
    dose_per_ha
   FROM report.workorder_metrics
  WITH NO DATA;


--
-- Name: app_parameters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_parameters ALTER COLUMN id SET DEFAULT nextval('public.app_parameters_id_seq'::regclass);


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
-- Name: engineering_principles_documentation id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.engineering_principles_documentation ALTER COLUMN id SET DEFAULT nextval('public.engineering_principles_documentation_id_seq'::regclass);


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
-- Name: app_parameters app_parameters_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_parameters
    ADD CONSTRAINT app_parameters_key_key UNIQUE (key);


--
-- Name: app_parameters app_parameters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.app_parameters
    ADD CONSTRAINT app_parameters_pkey PRIMARY KEY (id);


--
-- Name: campaigns campaigns_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT campaigns_name_key UNIQUE (name);


--
-- Name: campaigns campaigns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.campaigns
    ADD CONSTRAINT campaigns_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: crop_commercializations crop_commercializations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT crop_commercializations_pkey PRIMARY KEY (id);


--
-- Name: crops crops_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT crops_name_key UNIQUE (name);


--
-- Name: crops crops_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crops
    ADD CONSTRAINT crops_pkey PRIMARY KEY (id);


--
-- Name: customers customers_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_name_key UNIQUE (name);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: engineering_principles_documentation engineering_principles_documentation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.engineering_principles_documentation
    ADD CONSTRAINT engineering_principles_documentation_pkey PRIMARY KEY (id);


--
-- Name: fields fields_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fields_pkey PRIMARY KEY (id);


--
-- Name: flyway_schema_history flyway_schema_history_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.flyway_schema_history
    ADD CONSTRAINT flyway_schema_history_pk PRIMARY KEY (installed_rank);


--
-- Name: fx_rates fx_rates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fx_rates
    ADD CONSTRAINT fx_rates_pkey PRIMARY KEY (id);


--
-- Name: investors investors_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT investors_name_key UNIQUE (name);


--
-- Name: investors investors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.investors
    ADD CONSTRAINT investors_pkey PRIMARY KEY (id);


--
-- Name: invoices invoices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);


--
-- Name: invoices invoices_work_order_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_work_order_id_key UNIQUE (work_order_id);


--
-- Name: labor_categories labor_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_categories
    ADD CONSTRAINT labor_categories_pkey PRIMARY KEY (id);


--
-- Name: labor_types labor_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_types
    ADD CONSTRAINT labor_types_name_key UNIQUE (name);


--
-- Name: labor_types labor_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_types
    ADD CONSTRAINT labor_types_pkey PRIMARY KEY (id);


--
-- Name: labors labors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT labors_pkey PRIMARY KEY (id);


--
-- Name: lease_types lease_types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT lease_types_name_key UNIQUE (name);


--
-- Name: lease_types lease_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lease_types
    ADD CONSTRAINT lease_types_pkey PRIMARY KEY (id);


--
-- Name: lot_dates lot_dates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT lot_dates_pkey PRIMARY KEY (id);


--
-- Name: lots lots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lots
    ADD CONSTRAINT lots_pkey PRIMARY KEY (id);


--
-- Name: managers managers_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT managers_name_key UNIQUE (name);


--
-- Name: managers managers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.managers
    ADD CONSTRAINT managers_pkey PRIMARY KEY (id);


--
-- Name: project_dollar_values project_dollar_values_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT project_dollar_values_pkey PRIMARY KEY (id);


--
-- Name: project_dollar_values project_dollar_values_project_id_year_month_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT project_dollar_values_project_id_year_month_key UNIQUE (project_id, year, month);


--
-- Name: project_investors project_investors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_investors
    ADD CONSTRAINT project_investors_pkey PRIMARY KEY (project_id, investor_id);


--
-- Name: project_managers project_managers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_managers
    ADD CONSTRAINT project_managers_pkey PRIMARY KEY (project_id, manager_id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: providers providers_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_name_key UNIQUE (name);


--
-- Name: providers providers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.providers
    ADD CONSTRAINT providers_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: stocks stocks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT stocks_pkey PRIMARY KEY (id);


--
-- Name: supplies supplies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT supplies_pkey PRIMARY KEY (id);


--
-- Name: supply_movements supply_movements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT supply_movements_pkey PRIMARY KEY (id);


--
-- Name: types types_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.types
    ADD CONSTRAINT types_name_key UNIQUE (name);


--
-- Name: types types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.types
    ADD CONSTRAINT types_pkey PRIMARY KEY (id);


--
-- Name: lot_dates unique_lot_dates; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT unique_lot_dates UNIQUE (lot_id, sequence);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: workorder_items workorder_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT workorder_items_pkey PRIMARY KEY (id);


--
-- Name: workorders workorders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_pkey PRIMARY KEY (id);


--
-- Name: flyway_schema_history_s_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX flyway_schema_history_s_idx ON public.flyway_schema_history USING btree (success);


--
-- Name: idx_app_parameters_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_parameters_category ON public.app_parameters USING btree (category);


--
-- Name: idx_app_parameters_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_app_parameters_key ON public.app_parameters USING btree (key);


--
-- Name: idx_crop_commercializations_crop_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crop_commercializations_crop_id ON public.crop_commercializations USING btree (crop_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_crop_commercializations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crop_commercializations_deleted_at ON public.crop_commercializations USING btree (deleted_at);


--
-- Name: idx_crop_commercializations_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_crop_commercializations_project_id ON public.crop_commercializations USING btree (project_id);


--
-- Name: INDEX idx_crop_commercializations_project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON INDEX public.idx_crop_commercializations_project_id IS 'Índice para optimizar cálculos de commercialization por project_id';


--
-- Name: idx_fields_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fields_project_id ON public.fields USING btree (project_id);


--
-- Name: idx_fx_rates_effective_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_fx_rates_effective_date ON public.fx_rates USING btree (effective_date);


--
-- Name: idx_fx_rates_unique_pair_date; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_fx_rates_unique_pair_date ON public.fx_rates USING btree (currency_pair, effective_date);


--
-- Name: idx_labor_grouping; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_grouping ON public.workorders USING btree (project_id, field_id, effective_area) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));


--
-- Name: idx_labor_labors_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_labors_notdel ON public.labors USING btree (id, price) WHERE (deleted_at IS NULL);


--
-- Name: idx_labor_supplies_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_supplies_notdel ON public.supplies USING btree (id, price) WHERE (deleted_at IS NULL);


--
-- Name: idx_labor_supplies_units_v2; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_supplies_units_v2 ON public.supplies USING btree (id, price, unit_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_labor_workorder_items_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_workorder_items_notdel ON public.workorder_items USING btree (workorder_id, supply_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_labor_workorder_items_supply; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_workorder_items_supply ON public.workorder_items USING btree (workorder_id, supply_id, final_dose) WHERE (deleted_at IS NULL);


--
-- Name: idx_labor_workorder_items_v2; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_workorder_items_v2 ON public.workorder_items USING btree (workorder_id, supply_id, total_used, final_dose) WHERE (deleted_at IS NULL);


--
-- Name: idx_labor_workorders_composite; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_workorders_composite ON public.workorders USING btree (project_id, field_id, labor_id, effective_area) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));


--
-- Name: idx_labor_workorders_metrics_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_workorders_metrics_notdel ON public.workorders USING btree (project_id, field_id, labor_id, effective_area) WHERE (deleted_at IS NULL);


--
-- Name: idx_labor_workorders_metrics_v2; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labor_workorders_metrics_v2 ON public.workorders USING btree (project_id, field_id, labor_id, effective_area) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));


--
-- Name: idx_labors_category_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labors_category_id ON public.labors USING btree (category_id) WHERE (deleted_at IS NULL);


--
-- Name: INDEX idx_labors_category_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON INDEX public.idx_labors_category_id IS 'Índice para optimizar cálculos de labors por category_id';


--
-- Name: idx_labors_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_labors_project_id ON public.labors USING btree (project_id) WHERE (deleted_at IS NULL);


--
-- Name: INDEX idx_labors_project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON INDEX public.idx_labors_project_id IS 'Índice para optimizar cálculos de labors por project_id';


--
-- Name: idx_lot_table_crop_commercializations_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_crop_commercializations_notdel ON public.crop_commercializations USING btree (project_id, crop_id, net_price) WHERE (deleted_at IS NULL);


--
-- Name: idx_lot_table_crops_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_crops_notdel ON public.crops USING btree (id, name) WHERE (deleted_at IS NULL);


--
-- Name: idx_lot_table_fields_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_fields_notdel ON public.fields USING btree (id, project_id, lease_type_id, lease_type_value, lease_type_percent) WHERE (deleted_at IS NULL);


--
-- Name: idx_lot_table_labors_harvest; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_labors_harvest ON public.labors USING btree (id, category_id) WHERE ((deleted_at IS NULL) AND (category_id = 2));


--
-- Name: idx_lot_table_labors_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_labors_notdel ON public.labors USING btree (id, category_id, price) WHERE (deleted_at IS NULL);


--
-- Name: idx_lot_table_labors_sowing; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_labors_sowing ON public.labors USING btree (id, category_id) WHERE ((deleted_at IS NULL) AND (category_id = 1));


--
-- Name: idx_lot_table_lots_composite; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_lots_composite ON public.lots USING btree (field_id, current_crop_id, previous_crop_id, tons, hectares) WHERE ((deleted_at IS NULL) AND (hectares > (0)::double precision));


--
-- Name: idx_lot_table_lots_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_lots_notdel ON public.lots USING btree (id, field_id, current_crop_id, previous_crop_id, tons, hectares) WHERE (deleted_at IS NULL);


--
-- Name: idx_lot_table_projects_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_projects_notdel ON public.projects USING btree (id, admin_cost) WHERE (deleted_at IS NULL);


--
-- Name: idx_lot_table_supplies_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_supplies_notdel ON public.supplies USING btree (id, price) WHERE (deleted_at IS NULL);


--
-- Name: idx_lot_table_workorder_items_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_workorder_items_notdel ON public.workorder_items USING btree (workorder_id, supply_id, final_dose) WHERE (deleted_at IS NULL);


--
-- Name: idx_lot_table_workorders_composite; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_workorders_composite ON public.workorders USING btree (lot_id, labor_id, effective_area, date) WHERE ((deleted_at IS NULL) AND (effective_area > (0)::numeric));


--
-- Name: idx_lot_table_workorders_notdel; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lot_table_workorders_notdel ON public.workorders USING btree (lot_id, effective_area, date) WHERE (deleted_at IS NULL);


--
-- Name: idx_lots_current_crop_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_current_crop_id ON public.lots USING btree (current_crop_id);


--
-- Name: idx_lots_field_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_field_id ON public.lots USING btree (field_id);


--
-- Name: INDEX idx_lots_field_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON INDEX public.idx_lots_field_id IS 'Índice para optimizar cálculos de lots por field_id';


--
-- Name: idx_lots_previous_crop_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_lots_previous_crop_id ON public.lots USING btree (previous_crop_id);


--
-- Name: idx_project_dollar_values_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_project_dollar_values_project_id ON public.project_dollar_values USING btree (project_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_projects_campaign_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_campaign_id ON public.projects USING btree (campaign_id);


--
-- Name: idx_projects_customer_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_customer_id ON public.projects USING btree (customer_id);


--
-- Name: idx_projects_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_projects_id ON public.projects USING btree (id) WHERE (deleted_at IS NULL);


--
-- Name: idx_workorder_items_supply_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorder_items_supply_id ON public.workorder_items USING btree (supply_id) WHERE (deleted_at IS NULL);


--
-- Name: INDEX idx_workorder_items_supply_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON INDEX public.idx_workorder_items_supply_id IS 'Índice para optimizar cálculos de workorder items por supply_id';


--
-- Name: idx_workorder_items_workorder_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_workorder_items_workorder_id ON public.workorder_items USING btree (workorder_id) WHERE (deleted_at IS NULL);


--
-- Name: INDEX idx_workorder_items_workorder_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON INDEX public.idx_workorder_items_workorder_id IS 'Índice para optimizar cálculos de workorder items por workorder_id';


--
-- Name: workorder_metrics_mv_pk; Type: INDEX; Schema: report; Owner: -
--

CREATE UNIQUE INDEX workorder_metrics_mv_pk ON report.workorder_metrics_mv USING btree (project_id, field_id, lot_id);


--
-- Name: users set_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER set_timestamp BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();


--
-- Name: categories categories_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.types(id);


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
-- Name: supplies fk_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: labors fk_category; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: crop_commercializations fk_crop; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT fk_crop FOREIGN KEY (crop_id) REFERENCES public.crops(id);


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
-- Name: stocks fk_investor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: supply_movements fk_investor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;


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
-- Name: fields fk_lease_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fields
    ADD CONSTRAINT fk_lease_type FOREIGN KEY (lease_type_id) REFERENCES public.lease_types(id);


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
-- Name: lot_dates fk_lot_dates_lot; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT fk_lot_dates_lot FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON DELETE CASCADE;


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
-- Name: project_dollar_values fk_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.project_dollar_values
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: labors fk_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labors
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: crop_commercializations fk_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.crop_commercializations
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id);


--
-- Name: stocks fk_project; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;


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
-- Name: supply_movements fk_provider; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_provider FOREIGN KEY (provider_id) REFERENCES public.providers(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: stocks fk_supply; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stocks
    ADD CONSTRAINT fk_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: supply_movements fk_supply; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supply_movements
    ADD CONSTRAINT fk_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: supplies fk_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.supplies
    ADD CONSTRAINT fk_type FOREIGN KEY (type_id) REFERENCES public.types(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: labor_categories labor_categories_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.labor_categories
    ADD CONSTRAINT labor_categories_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.labor_types(id);


--
-- Name: workorder_items workorder_items_supply_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT workorder_items_supply_id_fkey FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorder_items workorder_items_workorder_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorder_items
    ADD CONSTRAINT workorder_items_workorder_id_fkey FOREIGN KEY (workorder_id) REFERENCES public.workorders(id) ON UPDATE CASCADE ON DELETE CASCADE;


--
-- Name: workorders workorders_crop_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_crop_id_fkey FOREIGN KEY (crop_id) REFERENCES public.crops(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorders workorders_field_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_field_id_fkey FOREIGN KEY (field_id) REFERENCES public.fields(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorders workorders_labor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_labor_id_fkey FOREIGN KEY (labor_id) REFERENCES public.labors(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorders workorders_lot_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_lot_id_fkey FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- Name: workorders workorders_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.workorders
    ADD CONSTRAINT workorders_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;


--
-- PostgreSQL database dump complete
--
