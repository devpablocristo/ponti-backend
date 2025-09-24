--
-- PostgreSQL database dump
--

\restrict kArU3igAZCTaMLnvESuwVen1f6gQqevemM4DRAmiCFMErh8ugxPtcKj7MFe4LE4

-- Dumped from database version 17.6
-- Dumped by pg_dump version 17.6 (Ubuntu 17.6-1.pgdg24.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE ONLY public.workorders DROP CONSTRAINT workorders_project_id_fkey;
ALTER TABLE ONLY public.workorders DROP CONSTRAINT workorders_lot_id_fkey;
ALTER TABLE ONLY public.workorders DROP CONSTRAINT workorders_labor_id_fkey;
ALTER TABLE ONLY public.workorders DROP CONSTRAINT workorders_field_id_fkey;
ALTER TABLE ONLY public.workorders DROP CONSTRAINT workorders_crop_id_fkey;
ALTER TABLE ONLY public.workorder_items DROP CONSTRAINT workorder_items_workorder_id_fkey;
ALTER TABLE ONLY public.workorder_items DROP CONSTRAINT workorder_items_supply_id_fkey;
ALTER TABLE ONLY public.labor_categories DROP CONSTRAINT labor_categories_type_id_fkey;
ALTER TABLE ONLY public.supplies DROP CONSTRAINT fk_type;
ALTER TABLE ONLY public.supply_movements DROP CONSTRAINT fk_supply;
ALTER TABLE ONLY public.stocks DROP CONSTRAINT fk_supply;
ALTER TABLE ONLY public.supply_movements DROP CONSTRAINT fk_provider;
ALTER TABLE ONLY public.projects DROP CONSTRAINT fk_projects_updated_by;
ALTER TABLE ONLY public.projects DROP CONSTRAINT fk_projects_deleted_by;
ALTER TABLE ONLY public.projects DROP CONSTRAINT fk_projects_customer;
ALTER TABLE ONLY public.projects DROP CONSTRAINT fk_projects_created_by;
ALTER TABLE ONLY public.projects DROP CONSTRAINT fk_projects_campaign;
ALTER TABLE ONLY public.project_managers DROP CONSTRAINT fk_project_managers_updated_by;
ALTER TABLE ONLY public.project_managers DROP CONSTRAINT fk_project_managers_project;
ALTER TABLE ONLY public.project_managers DROP CONSTRAINT fk_project_managers_manager;
ALTER TABLE ONLY public.project_managers DROP CONSTRAINT fk_project_managers_deleted_by;
ALTER TABLE ONLY public.project_managers DROP CONSTRAINT fk_project_managers_created_by;
ALTER TABLE ONLY public.project_investors DROP CONSTRAINT fk_project_investors_updated_by;
ALTER TABLE ONLY public.project_investors DROP CONSTRAINT fk_project_investors_project;
ALTER TABLE ONLY public.project_investors DROP CONSTRAINT fk_project_investors_investor;
ALTER TABLE ONLY public.project_investors DROP CONSTRAINT fk_project_investors_deleted_by;
ALTER TABLE ONLY public.project_investors DROP CONSTRAINT fk_project_investors_created_by;
ALTER TABLE ONLY public.stocks DROP CONSTRAINT fk_project;
ALTER TABLE ONLY public.crop_commercializations DROP CONSTRAINT fk_project;
ALTER TABLE ONLY public.labors DROP CONSTRAINT fk_project;
ALTER TABLE ONLY public.project_dollar_values DROP CONSTRAINT fk_project;
ALTER TABLE ONLY public.managers DROP CONSTRAINT fk_managers_updated_by;
ALTER TABLE ONLY public.managers DROP CONSTRAINT fk_managers_deleted_by;
ALTER TABLE ONLY public.managers DROP CONSTRAINT fk_managers_created_by;
ALTER TABLE ONLY public.lots DROP CONSTRAINT fk_lots_updated_by;
ALTER TABLE ONLY public.lots DROP CONSTRAINT fk_lots_previous_crop;
ALTER TABLE ONLY public.lots DROP CONSTRAINT fk_lots_field;
ALTER TABLE ONLY public.lots DROP CONSTRAINT fk_lots_deleted_by;
ALTER TABLE ONLY public.lots DROP CONSTRAINT fk_lots_current_crop;
ALTER TABLE ONLY public.lots DROP CONSTRAINT fk_lots_created_by;
ALTER TABLE ONLY public.lot_dates DROP CONSTRAINT fk_lot_dates_lot;
ALTER TABLE ONLY public.lease_types DROP CONSTRAINT fk_lease_types_updated_by;
ALTER TABLE ONLY public.lease_types DROP CONSTRAINT fk_lease_types_deleted_by;
ALTER TABLE ONLY public.lease_types DROP CONSTRAINT fk_lease_types_created_by;
ALTER TABLE ONLY public.fields DROP CONSTRAINT fk_lease_type;
ALTER TABLE ONLY public.invoices DROP CONSTRAINT fk_invoices_work_order;
ALTER TABLE ONLY public.investors DROP CONSTRAINT fk_investors_updated_by;
ALTER TABLE ONLY public.investors DROP CONSTRAINT fk_investors_deleted_by;
ALTER TABLE ONLY public.investors DROP CONSTRAINT fk_investors_created_by;
ALTER TABLE ONLY public.supply_movements DROP CONSTRAINT fk_investor;
ALTER TABLE ONLY public.stocks DROP CONSTRAINT fk_investor;
ALTER TABLE ONLY public.fields DROP CONSTRAINT fk_fields_updated_by;
ALTER TABLE ONLY public.fields DROP CONSTRAINT fk_fields_project;
ALTER TABLE ONLY public.fields DROP CONSTRAINT fk_fields_deleted_by;
ALTER TABLE ONLY public.fields DROP CONSTRAINT fk_fields_created_by;
ALTER TABLE ONLY public.customers DROP CONSTRAINT fk_customers_updated_by;
ALTER TABLE ONLY public.customers DROP CONSTRAINT fk_customers_deleted_by;
ALTER TABLE ONLY public.customers DROP CONSTRAINT fk_customers_created_by;
ALTER TABLE ONLY public.crops DROP CONSTRAINT fk_crops_updated_by;
ALTER TABLE ONLY public.crops DROP CONSTRAINT fk_crops_deleted_by;
ALTER TABLE ONLY public.crops DROP CONSTRAINT fk_crops_created_by;
ALTER TABLE ONLY public.crop_commercializations DROP CONSTRAINT fk_crop;
ALTER TABLE ONLY public.labors DROP CONSTRAINT fk_category;
ALTER TABLE ONLY public.supplies DROP CONSTRAINT fk_category;
ALTER TABLE ONLY public.campaigns DROP CONSTRAINT fk_campaigns_updated_by;
ALTER TABLE ONLY public.campaigns DROP CONSTRAINT fk_campaigns_deleted_by;
ALTER TABLE ONLY public.campaigns DROP CONSTRAINT fk_campaigns_created_by;
ALTER TABLE ONLY public.categories DROP CONSTRAINT categories_type_id_fkey;
DROP TRIGGER set_timestamp ON public.users;
DROP INDEX public.idx_workorder_items_workorder_id;
DROP INDEX public.idx_workorder_items_supply_id;
DROP INDEX public.idx_users_deleted_at;
DROP INDEX public.idx_projects_id;
DROP INDEX public.idx_projects_customer_id;
DROP INDEX public.idx_projects_campaign_id;
DROP INDEX public.idx_project_dollar_values_project_id;
DROP INDEX public.idx_lots_previous_crop_id;
DROP INDEX public.idx_lots_field_id;
DROP INDEX public.idx_lots_current_crop_id;
DROP INDEX public.idx_lot_table_workorders_notdel;
DROP INDEX public.idx_lot_table_workorders_composite;
DROP INDEX public.idx_lot_table_workorder_items_notdel;
DROP INDEX public.idx_lot_table_supplies_notdel;
DROP INDEX public.idx_lot_table_projects_notdel;
DROP INDEX public.idx_lot_table_lots_notdel;
DROP INDEX public.idx_lot_table_lots_composite;
DROP INDEX public.idx_lot_table_labors_sowing;
DROP INDEX public.idx_lot_table_labors_notdel;
DROP INDEX public.idx_lot_table_labors_harvest;
DROP INDEX public.idx_lot_table_fields_notdel;
DROP INDEX public.idx_lot_table_crops_notdel;
DROP INDEX public.idx_lot_table_crop_commercializations_notdel;
DROP INDEX public.idx_labors_project_id;
DROP INDEX public.idx_labors_category_id;
DROP INDEX public.idx_labor_workorders_metrics_v2;
DROP INDEX public.idx_labor_workorders_metrics_notdel;
DROP INDEX public.idx_labor_workorders_composite;
DROP INDEX public.idx_labor_workorder_items_v2;
DROP INDEX public.idx_labor_workorder_items_supply;
DROP INDEX public.idx_labor_workorder_items_notdel;
DROP INDEX public.idx_labor_supplies_units_v2;
DROP INDEX public.idx_labor_supplies_notdel;
DROP INDEX public.idx_labor_labors_notdel;
DROP INDEX public.idx_labor_grouping;
DROP INDEX public.idx_fx_rates_unique_pair_date;
DROP INDEX public.idx_fx_rates_effective_date;
DROP INDEX public.idx_fields_project_id;
DROP INDEX public.idx_crop_commercializations_project_id;
DROP INDEX public.idx_crop_commercializations_deleted_at;
DROP INDEX public.idx_crop_commercializations_crop_id;
DROP INDEX public.idx_app_parameters_key;
DROP INDEX public.idx_app_parameters_category;
ALTER TABLE ONLY public.workorders DROP CONSTRAINT workorders_pkey;
ALTER TABLE ONLY public.workorder_items DROP CONSTRAINT workorder_items_pkey;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_pkey;
ALTER TABLE ONLY public.user_logins DROP CONSTRAINT user_logins_pkey;
ALTER TABLE ONLY public.lot_dates DROP CONSTRAINT unique_lot_dates;
ALTER TABLE ONLY public.users DROP CONSTRAINT uni_users_username;
ALTER TABLE ONLY public.types DROP CONSTRAINT types_pkey;
ALTER TABLE ONLY public.types DROP CONSTRAINT types_name_key;
ALTER TABLE ONLY public.supply_movements DROP CONSTRAINT supply_movements_pkey;
ALTER TABLE ONLY public.supplies DROP CONSTRAINT supplies_pkey;
ALTER TABLE ONLY public.stocks DROP CONSTRAINT stocks_pkey;
ALTER TABLE ONLY public.schema_migrations DROP CONSTRAINT schema_migrations_pkey;
ALTER TABLE ONLY public.providers DROP CONSTRAINT providers_pkey;
ALTER TABLE ONLY public.providers DROP CONSTRAINT providers_name_key;
ALTER TABLE ONLY public.projects DROP CONSTRAINT projects_pkey;
ALTER TABLE ONLY public.project_managers DROP CONSTRAINT project_managers_pkey;
ALTER TABLE ONLY public.project_investors DROP CONSTRAINT project_investors_pkey;
ALTER TABLE ONLY public.project_dollar_values DROP CONSTRAINT project_dollar_values_project_id_year_month_key;
ALTER TABLE ONLY public.project_dollar_values DROP CONSTRAINT project_dollar_values_pkey;
ALTER TABLE ONLY public.managers DROP CONSTRAINT managers_pkey;
ALTER TABLE ONLY public.managers DROP CONSTRAINT managers_name_key;
ALTER TABLE ONLY public.lots DROP CONSTRAINT lots_pkey;
ALTER TABLE ONLY public.lot_dates DROP CONSTRAINT lot_dates_pkey;
ALTER TABLE ONLY public.lease_types DROP CONSTRAINT lease_types_pkey;
ALTER TABLE ONLY public.lease_types DROP CONSTRAINT lease_types_name_key;
ALTER TABLE ONLY public.labors DROP CONSTRAINT labors_pkey;
ALTER TABLE ONLY public.labor_types DROP CONSTRAINT labor_types_pkey;
ALTER TABLE ONLY public.labor_types DROP CONSTRAINT labor_types_name_key;
ALTER TABLE ONLY public.labor_categories DROP CONSTRAINT labor_categories_pkey;
ALTER TABLE ONLY public.invoices DROP CONSTRAINT invoices_work_order_id_key;
ALTER TABLE ONLY public.invoices DROP CONSTRAINT invoices_pkey;
ALTER TABLE ONLY public.investors DROP CONSTRAINT investors_pkey;
ALTER TABLE ONLY public.investors DROP CONSTRAINT investors_name_key;
ALTER TABLE ONLY public.fx_rates DROP CONSTRAINT fx_rates_pkey;
ALTER TABLE ONLY public.fields DROP CONSTRAINT fields_pkey;
ALTER TABLE ONLY public.engineering_principles_documentation DROP CONSTRAINT engineering_principles_documentation_pkey;
ALTER TABLE ONLY public.customers DROP CONSTRAINT customers_pkey;
ALTER TABLE ONLY public.customers DROP CONSTRAINT customers_name_key;
ALTER TABLE ONLY public.crops DROP CONSTRAINT crops_pkey;
ALTER TABLE ONLY public.crops DROP CONSTRAINT crops_name_key;
ALTER TABLE ONLY public.crop_commercializations DROP CONSTRAINT crop_commercializations_pkey;
ALTER TABLE ONLY public.categories DROP CONSTRAINT categories_pkey;
ALTER TABLE ONLY public.campaigns DROP CONSTRAINT campaigns_pkey;
ALTER TABLE ONLY public.campaigns DROP CONSTRAINT campaigns_name_key;
ALTER TABLE ONLY public.app_parameters DROP CONSTRAINT app_parameters_pkey;
ALTER TABLE ONLY public.app_parameters DROP CONSTRAINT app_parameters_key_key;
ALTER TABLE public.workorders ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.workorder_items ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.users ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.user_logins ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.types ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.supply_movements ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.supplies ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.stocks ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.providers ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.projects ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.project_dollar_values ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.managers ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.lots ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.lot_dates ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.lease_types ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.labors ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.labor_types ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.labor_categories ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.invoices ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.investors ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.fx_rates ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.fields ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.engineering_principles_documentation ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.customers ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.crops ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.crop_commercializations ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.categories ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.campaigns ALTER COLUMN id DROP DEFAULT;
ALTER TABLE public.app_parameters ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE public.workorders_id_seq;
DROP SEQUENCE public.workorder_items_id_seq;
DROP VIEW public.v3_workorder_metrics;
DROP VIEW public.v3_workorder_list;
DROP VIEW public.v3_report_summary_results_view;
DROP VIEW public.v3_report_field_crop_metrics_view;
DROP VIEW public.v3_lot_metrics;
DROP VIEW public.v3_lot_list;
DROP VIEW public.v3_labor_metrics;
DROP VIEW public.v3_labor_list;
DROP VIEW public.v3_investor_contribution_data_view;
DROP TABLE public.workorder_items;
DROP VIEW public.v3_dashboard_management_balance;
DROP VIEW public.v3_dashboard_crop_incidence;
DROP VIEW public.v3_dashboard_contributions_progress;
DROP VIEW public.v3_dashboard;
DROP TABLE public.workorders;
DROP SEQUENCE public.users_id_seq;
DROP TABLE public.users;
DROP SEQUENCE public.user_logins_id_seq;
DROP TABLE public.user_logins;
DROP SEQUENCE public.types_id_seq;
DROP TABLE public.types;
DROP SEQUENCE public.supply_movements_id_seq;
DROP TABLE public.supply_movements;
DROP SEQUENCE public.supplies_id_seq;
DROP TABLE public.supplies;
DROP SEQUENCE public.stocks_id_seq;
DROP TABLE public.stocks;
DROP TABLE public.schema_migrations;
DROP SEQUENCE public.providers_id_seq;
DROP TABLE public.providers;
DROP SEQUENCE public.projects_id_seq;
DROP TABLE public.projects;
DROP TABLE public.project_managers;
DROP TABLE public.project_investors;
DROP SEQUENCE public.project_dollar_values_id_seq;
DROP TABLE public.project_dollar_values;
DROP SEQUENCE public.managers_id_seq;
DROP TABLE public.managers;
DROP SEQUENCE public.lots_id_seq;
DROP TABLE public.lots;
DROP SEQUENCE public.lot_dates_id_seq;
DROP TABLE public.lot_dates;
DROP SEQUENCE public.lease_types_id_seq;
DROP TABLE public.lease_types;
DROP SEQUENCE public.labors_id_seq;
DROP TABLE public.labors;
DROP SEQUENCE public.labor_types_id_seq;
DROP TABLE public.labor_types;
DROP SEQUENCE public.labor_categories_id_seq;
DROP TABLE public.labor_categories;
DROP SEQUENCE public.invoices_id_seq;
DROP TABLE public.invoices;
DROP SEQUENCE public.investors_id_seq;
DROP TABLE public.investors;
DROP SEQUENCE public.fx_rates_id_seq;
DROP TABLE public.fx_rates;
DROP SEQUENCE public.fields_id_seq;
DROP TABLE public.fields;
DROP SEQUENCE public.engineering_principles_documentation_id_seq;
DROP TABLE public.engineering_principles_documentation;
DROP SEQUENCE public.customers_id_seq;
DROP TABLE public.customers;
DROP SEQUENCE public.crops_id_seq;
DROP TABLE public.crops;
DROP SEQUENCE public.crop_commercializations_id_seq;
DROP TABLE public.crop_commercializations;
DROP SEQUENCE public.categories_id_seq;
DROP TABLE public.categories;
DROP SEQUENCE public.campaigns_id_seq;
DROP TABLE public.campaigns;
DROP SEQUENCE public.app_parameters_id_seq;
DROP TABLE public.app_parameters;
DROP FUNCTION v3_calc.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric);
DROP FUNCTION v3_calc.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric);
DROP FUNCTION v3_calc.yield_tn_per_ha_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.units_per_ha(units numeric, area numeric);
DROP FUNCTION v3_calc.total_invested_cost_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.total_hectares_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.total_costs_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.total_costs_for_crop(p_project_id bigint, p_crop_id bigint);
DROP FUNCTION v3_calc.total_budget_cost_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.supply_cost_received_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.supply_cost_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.supply_cost_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.supply_cost(final_dose double precision, supply_price numeric, effective_area numeric);
DROP FUNCTION v3_calc.stock_value_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.seeded_area_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.seeded_area(sowing_date date, hectares numeric);
DROP FUNCTION v3_calc.safe_div_dp(double precision, double precision);
DROP FUNCTION v3_calc.safe_div(numeric, numeric);
DROP FUNCTION v3_calc.renta_pct(operating_result_total_usd double precision, total_costs_usd double precision);
DROP FUNCTION v3_calc.rent_per_ha_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.rent_per_ha(lease_type_id bigint, lease_type_percent double precision, lease_type_value double precision, income_net_per_ha double precision, cost_per_ha double precision, admin_cost_per_ha double precision);
DROP FUNCTION v3_calc.rent_per_ha(lease_type_id integer, lease_type_percent double precision, lease_type_value double precision, income_net_per_ha double precision, cost_per_ha double precision, admin_cost_per_ha double precision);
DROP FUNCTION v3_calc.percentage_capped(numeric, numeric);
DROP FUNCTION v3_calc.percentage(numeric, numeric);
DROP FUNCTION v3_calc.per_ha_dp(double precision, double precision);
DROP FUNCTION v3_calc.per_ha(numeric, numeric);
DROP FUNCTION v3_calc.operating_result_total_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.operating_result_per_ha_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.operating_result_per_ha(income_net_per_ha double precision, active_total_per_ha double precision);
DROP FUNCTION v3_calc.norm_dose(dose numeric, area numeric);
DROP FUNCTION v3_calc.net_price_usd_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.lot_tons(p_lot_id bigint);
DROP FUNCTION v3_calc.lot_hectares(p_lot_id bigint);
DROP FUNCTION v3_calc.labor_cost_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.labor_cost(labor_price numeric, effective_area numeric);
DROP FUNCTION v3_calc.indifference_price_usd_tn(total_invested_per_ha double precision, yield_tn_per_ha double precision);
DROP FUNCTION v3_calc.income_net_total_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.income_net_total(tons numeric, net_price_usd numeric);
DROP FUNCTION v3_calc.income_net_per_ha_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.income_net_per_ha(income_net_total numeric, hectares numeric);
DROP FUNCTION v3_calc.harvested_area_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.harvested_area(tons numeric, hectares numeric);
DROP FUNCTION v3_calc.dose_per_ha(total_dose numeric, surface_ha numeric);
DROP FUNCTION v3_calc.dollar_average_for_month(p_project_id bigint, p_date date);
DROP FUNCTION v3_calc.direct_costs_invested_for_project(p_project_id bigint);
DROP FUNCTION v3_calc.direct_cost_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.cost_per_ha_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.cost_per_ha_for_crop(p_project_id bigint, p_crop_id bigint);
DROP FUNCTION v3_calc.cost_per_ha(total_cost numeric, hectares numeric);
DROP FUNCTION v3_calc.coalesce0(numeric);
DROP FUNCTION v3_calc.coalesce0(double precision);
DROP FUNCTION v3_calc.calculate_campaign_closing_date(end_date date);
DROP FUNCTION v3_calc.admin_cost_per_ha_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.active_total_per_ha_for_lot(p_lot_id bigint);
DROP FUNCTION v3_calc.active_total_per_ha(direct_cost_per_ha double precision, rent_per_ha double precision, admin_cost_per_ha double precision);
DROP FUNCTION public.update_timestamp();
DROP FUNCTION public.get_project_dollar_value(p_project_id bigint, p_month character varying);
DROP FUNCTION public.get_iva_percentage();
DROP FUNCTION public.get_default_fx_rate();
DROP FUNCTION public.get_campaign_closure_days();
DROP FUNCTION public.get_app_parameter_integer(p_key character varying);
DROP FUNCTION public.get_app_parameter_decimal(p_key character varying);
DROP FUNCTION public.get_app_parameter(p_key character varying);
DROP FUNCTION public.calculate_yield(p_tons numeric, p_hectares numeric);
DROP FUNCTION public.calculate_supply_cost(p_final_dose double precision, p_supply_price numeric, p_effective_area numeric);
DROP FUNCTION public.calculate_sowed_area(p_sowing_date date, p_hectares numeric);
DROP FUNCTION public.calculate_labor_cost(p_labor_price numeric, p_effective_area numeric);
DROP FUNCTION public.calculate_harvested_area(p_tons numeric, p_hectares numeric);
DROP FUNCTION public.calculate_cost_per_ha(p_total_cost numeric, p_hectares numeric);
DROP FUNCTION public.calculate_campaign_closing_date(end_date date);
DROP TYPE public.movement_type;
DROP SCHEMA v3_calc;
-- *not* dropping schema, since initdb creates it
--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

-- *not* creating schema, since initdb creates it


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS '';


--
-- Name: v3_calc; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA v3_calc;


--
-- Name: movement_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.movement_type AS ENUM (
    'Stock',
    'Movimiento interno',
    'Remito oficial'
);


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


--
-- Name: active_total_per_ha(double precision, double precision, double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.active_total_per_ha(direct_cost_per_ha double precision, rent_per_ha double precision, admin_cost_per_ha double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(direct_cost_per_ha,0) + COALESCE(rent_per_ha,0) + COALESCE(admin_cost_per_ha,0)
$$;


--
-- Name: active_total_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.active_total_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.active_total_per_ha(
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.rent_per_ha_for_lot(p_lot_id),
           COALESCE(p.admin_cost, 0)::double precision
         )
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: FUNCTION active_total_per_ha_for_lot(p_lot_id bigint); Type: COMMENT; Schema: v3_calc; Owner: -
--

COMMENT ON FUNCTION v3_calc.active_total_per_ha_for_lot(p_lot_id bigint) IS 'Calcula el total activo por hectárea para un lote sumando: costo directo/ha + renta/ha + admin_cost del proyecto (sin prorrateo)';


--
-- Name: admin_cost_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.admin_cost_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(p.admin_cost, 0)::double precision
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  JOIN public.projects p ON p.id = f.project_id AND p.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: calculate_campaign_closing_date(date); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.calculate_campaign_closing_date(end_date date) RETURNS date
    LANGUAGE sql STABLE
    AS $$
  SELECT CASE 
    WHEN end_date IS NULL THEN NULL
    ELSE end_date + (get_campaign_closure_days() || ' days')::INTERVAL  -- Usar valor de app_parameters
  END::date
$$;


--
-- Name: coalesce0(double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.coalesce0(double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT COALESCE($1, 0)
$_$;


--
-- Name: coalesce0(numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.coalesce0(numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT COALESCE($1, 0)
$_$;


--
-- Name: cost_per_ha(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.cost_per_ha(total_cost numeric, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v3_calc.per_ha(total_cost, hectares)
$$;


--
-- Name: cost_per_ha_for_crop(bigint, bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.cost_per_ha_for_crop(p_project_id bigint, p_crop_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.per_ha_dp(
    v3_calc.total_costs_for_crop(p_project_id, p_crop_id),
    (SELECT COALESCE(SUM(l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  )
$$;


--
-- Name: cost_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.cost_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.safe_div_dp(
           COALESCE(v3_calc.direct_cost_for_lot(p_lot_id), 0)::double precision,
           v3_calc.lot_hectares(p_lot_id)
         )
$$;


--
-- Name: direct_cost_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.direct_cost_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(v3_calc.labor_cost_for_lot(p_lot_id), 0)::double precision
       + COALESCE(v3_calc.supply_cost_for_lot(p_lot_id), 0)
$$;


--
-- Name: direct_costs_invested_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.direct_costs_invested_for_project(p_project_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Labores invertidas (ejecutadas + no ejecutadas)
    (SELECT COALESCE(SUM(lb.price * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     JOIN public.labors lb ON lb.project_id = f.project_id AND lb.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Insumos invertidos (usados + no usados) - usar datos de stocks
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    +
    -- Insumos recibidos por movimientos internos
    v3_calc.supply_cost_received_for_project(p_project_id)
  , 0)::double precision
$$;


--
-- Name: dollar_average_for_month(bigint, date); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.dollar_average_for_month(p_project_id bigint, p_date date) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    (SELECT average_value 
     FROM public.project_dollar_values 
     WHERE project_id = p_project_id 
       AND month = TO_CHAR(p_date, 'YYYY-MM')  -- Convierte fecha a formato YYYY-MM
       AND deleted_at IS NULL
     LIMIT 1), 0  -- Si no hay datos para ese mes, retorna 0
  )::numeric
$$;


--
-- Name: dose_per_ha(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.dose_per_ha(total_dose numeric, surface_ha numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v3_calc.safe_div(total_dose, surface_ha)
$$;


--
-- Name: harvested_area(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.harvested_area(tons numeric, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN tons IS NOT NULL AND tons > 0 THEN COALESCE(hectares,0) ELSE 0 END
$$;


--
-- Name: harvested_area_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.harvested_area_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.harvested_area(
           v3_calc.lot_tons(p_lot_id)::numeric,
           v3_calc.lot_hectares(p_lot_id)::numeric
         )
$$;


--
-- Name: income_net_per_ha(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.income_net_per_ha(income_net_total numeric, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v3_calc.per_ha(income_net_total, hectares)
$$;


--
-- Name: income_net_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.income_net_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.safe_div_dp(
           COALESCE(v3_calc.income_net_total_for_lot(p_lot_id), 0)::double precision,
           v3_calc.lot_hectares(p_lot_id)
         )
$$;


--
-- Name: income_net_total(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.income_net_total(tons numeric, net_price_usd numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(tons,0) * COALESCE(net_price_usd,0)
$$;


--
-- Name: income_net_total_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.income_net_total_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(l.tons, 0)::numeric * COALESCE(v3_calc.net_price_usd_for_lot(l.id), 0)::numeric
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: indifference_price_usd_tn(double precision, double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.indifference_price_usd_tn(total_invested_per_ha double precision, yield_tn_per_ha double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v3_calc.per_ha_dp(total_invested_per_ha, yield_tn_per_ha)
$$;


--
-- Name: labor_cost(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.labor_cost(labor_price numeric, effective_area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(labor_price,0) * COALESCE(effective_area,0)
$$;


--
-- Name: labor_cost_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.labor_cost_for_lot(p_lot_id bigint) RETURNS numeric
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
-- Name: lot_hectares(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.lot_hectares(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(l.hectares, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: lot_tons(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.lot_tons(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(l.tons, 0)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: net_price_usd_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.net_price_usd_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(cc.net_price, 0)
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  LEFT JOIN public.crop_commercializations cc
    ON cc.project_id = f.project_id
   AND cc.crop_id    = l.current_crop_id
   AND cc.deleted_at IS NULL
  WHERE l.id = p_lot_id
    AND l.deleted_at IS NULL
  LIMIT 1
$$;


--
-- Name: norm_dose(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.norm_dose(dose numeric, area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN area > 0 THEN dose / area ELSE NULL END
$$;


--
-- Name: operating_result_per_ha(double precision, double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.operating_result_per_ha(income_net_per_ha double precision, active_total_per_ha double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(income_net_per_ha,0) - COALESCE(active_total_per_ha,0)
$$;


--
-- Name: operating_result_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.operating_result_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.operating_result_per_ha(
           v3_calc.income_net_per_ha_for_lot(p_lot_id),
           v3_calc.active_total_per_ha_for_lot(p_lot_id)
         )
$$;


--
-- Name: operating_result_total_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.operating_result_total_for_project(p_project_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Ingresos netos totales
    (SELECT COALESCE(SUM(v3_calc.income_net_total_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costos directos ejecutados
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    -
    -- Costo administrativo total
    (SELECT COALESCE(p.admin_cost, 0)::double precision
     FROM public.projects p
     WHERE p.id = p_project_id AND p.deleted_at IS NULL)
  , 0)::double precision
$$;


--
-- Name: per_ha(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.per_ha(numeric, numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v3_calc.safe_div($1, $2)
$_$;


--
-- Name: per_ha_dp(double precision, double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.per_ha_dp(double precision, double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v3_calc.safe_div_dp($1, $2)
$_$;


--
-- Name: percentage(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.percentage(numeric, numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT v3_calc.safe_div($1, $2) * 100
$_$;


--
-- Name: percentage_capped(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.percentage_capped(numeric, numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT LEAST(v3_calc.safe_div($1, $2) * 100, 100)
$_$;


--
-- Name: rent_per_ha(integer, double precision, double precision, double precision, double precision, double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.rent_per_ha(lease_type_id integer, lease_type_percent double precision, lease_type_value double precision, income_net_per_ha double precision, cost_per_ha double precision, admin_cost_per_ha double precision) RETURNS double precision
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
-- Name: rent_per_ha(bigint, double precision, double precision, double precision, double precision, double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.rent_per_ha(lease_type_id bigint, lease_type_percent double precision, lease_type_value double precision, income_net_per_ha double precision, cost_per_ha double precision, admin_cost_per_ha double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v3_calc.rent_per_ha(lease_type_id::integer, lease_type_percent, lease_type_value,
                          income_net_per_ha, cost_per_ha, admin_cost_per_ha)
$$;


--
-- Name: rent_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.rent_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.rent_per_ha(
           f.lease_type_id,
           f.lease_type_percent,
           f.lease_type_value,
           v3_calc.income_net_per_ha_for_lot(p_lot_id),
           v3_calc.cost_per_ha_for_lot(p_lot_id),
           v3_calc.admin_cost_per_ha_for_lot(p_lot_id)
         )
  FROM public.lots l
  JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: renta_pct(double precision, double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.renta_pct(operating_result_total_usd double precision, total_costs_usd double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN COALESCE(total_costs_usd,0) > 0
              THEN (COALESCE(operating_result_total_usd,0) / total_costs_usd) * 100
              ELSE 0 END
$$;


--
-- Name: safe_div(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.safe_div(numeric, numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$_$;


--
-- Name: safe_div_dp(double precision, double precision); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.safe_div_dp(double precision, double precision) RETURNS double precision
    LANGUAGE sql IMMUTABLE
    AS $_$
  SELECT CASE WHEN $2 IS NULL OR $2 = 0 THEN 0 ELSE $1 / $2 END
$_$;


--
-- Name: seeded_area(date, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.seeded_area(sowing_date date, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT CASE WHEN sowing_date IS NOT NULL THEN COALESCE(hectares,0) ELSE 0 END
$$;


--
-- Name: seeded_area_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.seeded_area_for_lot(p_lot_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.seeded_area(l.sowing_date, l.hectares::numeric)
  FROM public.lots l
  WHERE l.id = p_lot_id AND l.deleted_at IS NULL
$$;


--
-- Name: stock_value_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.stock_value_for_project(p_project_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Stock disponible = insumos comprados - insumos consumidos
    (SELECT COALESCE(SUM(s.price * st.initial_units), 0)::double precision
     FROM public.supplies s
     JOIN public.stocks st ON st.supply_id = s.id AND st.deleted_at IS NULL
     WHERE s.project_id = p_project_id 
       AND s.deleted_at IS NULL
       AND st.initial_units IS NOT NULL)
    -
    (SELECT COALESCE(SUM(wi.total_used * s.price), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
     JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
     WHERE w.project_id = p_project_id AND w.deleted_at IS NULL)
  , 0)::double precision
$$;


--
-- Name: supply_cost(double precision, numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.supply_cost(final_dose double precision, supply_price numeric, effective_area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT COALESCE(final_dose,0)::numeric * COALESCE(supply_price,0) * COALESCE(effective_area,0)
$$;


--
-- Name: supply_cost_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.supply_cost_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Costos por workorder_items (uso directo en workorders)
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.lot_id = p_lot_id)
    +
    -- Costos por movimientos internos de salida (insumos transferidos a otros proyectos)
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
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
  , 0)::double precision
$$;


--
-- Name: supply_cost_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.supply_cost_for_project(p_project_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Costos por workorder_items (uso directo en workorders)
    (SELECT COALESCE(SUM((wi.final_dose)::double precision * s.price * (w.effective_area)::double precision), 0)::double precision
     FROM public.workorders w
     JOIN public.workorder_items wi ON wi.workorder_id = w.id
     JOIN public.supplies s ON s.id = wi.supply_id
     WHERE w.deleted_at IS NULL
       AND w.effective_area > 0
       AND wi.final_dose > 0
       AND s.price IS NOT NULL
       AND w.project_id = p_project_id)
    +
    -- Costos por movimientos internos de salida (insumos transferidos a otros proyectos)
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno'
       AND sm.is_entry = false
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::double precision
$$;


--
-- Name: supply_cost_received_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.supply_cost_received_for_project(p_project_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Costos por movimientos internos de entrada (insumos recibidos de otros proyectos)
    (SELECT COALESCE(SUM(sm.quantity * s.price), 0)::double precision
     FROM public.supply_movements sm
     JOIN public.supplies s ON s.id = sm.supply_id
     WHERE sm.deleted_at IS NULL
       AND s.deleted_at IS NULL
       AND sm.movement_type = 'Movimiento interno entrada'
       AND sm.is_entry = true
       AND sm.project_id = p_project_id
       AND s.price IS NOT NULL
       AND sm.quantity > 0)
  , 0)::double precision
$$;


--
-- Name: total_budget_cost_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.total_budget_cost_for_project(p_project_id bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
  -- Por ahora retornamos un valor placeholder hasta que se implemente
  -- el sistema de presupuestos. En el futuro esto debería consultar
  -- una tabla de presupuestos o calcular basado en labores/insumos planificados
  SELECT COALESCE(p.admin_cost * 10, 0)::numeric  -- Placeholder: 10x admin_cost
  FROM public.projects p
  WHERE p.id = p_project_id AND p.deleted_at IS NULL
$$;


--
-- Name: total_costs_for_crop(bigint, bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.total_costs_for_crop(p_project_id bigint, p_crop_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados para el cultivo
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id 
       AND l.current_crop_id = p_crop_id 
       AND l.deleted_at IS NULL)
  , 0)::double precision
$$;


--
-- Name: total_costs_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.total_costs_for_project(p_project_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados para todo el proyecto
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::double precision
$$;


--
-- Name: total_hectares_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.total_hectares_for_project(p_project_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(SUM(l.hectares), 0)::double precision
  FROM public.fields f
  LEFT JOIN public.lots l ON l.field_id = f.id AND l.deleted_at IS NULL
  WHERE f.project_id = p_project_id AND f.deleted_at IS NULL
$$;


--
-- Name: total_invested_cost_for_project(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.total_invested_cost_for_project(p_project_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT COALESCE(
    -- Costos directos ejecutados
    (SELECT COALESCE(SUM(v3_calc.direct_cost_for_lot(l.id)), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Arriendo invertido
    (SELECT COALESCE(SUM(v3_calc.rent_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
    +
    -- Estructura invertida
    (SELECT COALESCE(SUM(v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares), 0)::double precision
     FROM public.lots l
     JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
     WHERE f.project_id = p_project_id AND l.deleted_at IS NULL)
  , 0)::double precision
$$;


--
-- Name: units_per_ha(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.units_per_ha(units numeric, area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v3_calc.per_ha(units, area)
$$;


--
-- Name: yield_tn_per_ha_for_lot(bigint); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.yield_tn_per_ha_for_lot(p_lot_id bigint) RETURNS double precision
    LANGUAGE sql STABLE
    AS $$
  SELECT v3_calc.per_ha_dp(
           v3_calc.lot_tons(p_lot_id),
           v3_calc.lot_hectares(p_lot_id)
         )
$$;


--
-- Name: yield_tn_per_ha_over_harvested(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.yield_tn_per_ha_over_harvested(tons numeric, harvested_area numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v3_calc.safe_div( COALESCE(tons,0), COALESCE(harvested_area,0) )
$$;


--
-- Name: yield_tn_per_ha_over_hectares(numeric, numeric); Type: FUNCTION; Schema: v3_calc; Owner: -
--

CREATE FUNCTION v3_calc.yield_tn_per_ha_over_hectares(tons numeric, hectares numeric) RETURNS numeric
    LANGUAGE sql IMMUTABLE
    AS $$
  SELECT v3_calc.safe_div( COALESCE(tons,0), COALESCE(hectares,0) )
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

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
    deleted_by bigint
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
-- Name: user_logins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_logins (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    login_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    ip_address text,
    device_info text,
    success boolean DEFAULT true,
    logout_at timestamp with time zone,
    session_duration bigint
);


--
-- Name: user_logins_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_logins_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_logins_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_logins_id_seq OWNED BY public.user_logins.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email text,
    username text NOT NULL,
    password text,
    token_hash text,
    refresh_tokens text[],
    id_rol bigint,
    is_verified boolean,
    active boolean,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint
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
-- Name: v3_dashboard; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_dashboard AS
 WITH lots_base AS (
         SELECT l.id AS lot_id,
            f.project_id,
            l.hectares,
            l.tons,
            l.sowing_date
           FROM (public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
          WHERE (l.deleted_at IS NULL)
        ), w_min AS (
         SELECT w.project_id,
            min(w.date) AS start_date,
            ( SELECT w2.id
                   FROM public.workorders w2
                  WHERE ((w2.project_id = w.project_id) AND (w2.date = min(w.date)) AND (w2.deleted_at IS NULL))
                 LIMIT 1) AS first_workorder_id
           FROM public.workorders w
          WHERE (w.deleted_at IS NULL)
          GROUP BY w.project_id
        ), w_max AS (
         SELECT w.project_id,
            max(w.date) AS end_date,
            ( SELECT w2.id
                   FROM public.workorders w2
                  WHERE ((w2.project_id = w.project_id) AND (w2.date = max(w.date)) AND (w2.deleted_at IS NULL))
                 LIMIT 1) AS last_workorder_id
           FROM public.workorders w
          WHERE (w.deleted_at IS NULL)
          GROUP BY w.project_id
        ), last_stock_count AS (
         SELECT stocks.project_id,
            max(stocks.close_date) AS last_stock_count_date
           FROM public.stocks
          WHERE (stocks.deleted_at IS NULL)
          GROUP BY stocks.project_id
        )
 SELECT p.customer_id,
    p.id AS project_id,
    p.campaign_id,
    COALESCE(sum(
        CASE
            WHEN (lb.sowing_date IS NOT NULL) THEN lb.hectares
            ELSE (0)::double precision
        END), (0)::double precision) AS sowing_hectares,
    COALESCE(sum(lb.hectares), (0)::double precision) AS sowing_total_hectares,
    v3_calc.percentage((COALESCE(sum(
        CASE
            WHEN (lb.sowing_date IS NOT NULL) THEN lb.hectares
            ELSE (0)::double precision
        END), (0)::double precision))::numeric, (COALESCE(sum(lb.hectares), (0)::double precision))::numeric) AS sowing_progress_pct,
    COALESCE(sum(
        CASE
            WHEN ((lb.tons IS NOT NULL) AND (lb.tons > (0)::numeric)) THEN lb.hectares
            ELSE (0)::double precision
        END), (0)::double precision) AS harvest_hectares,
    COALESCE(sum(lb.hectares), (0)::double precision) AS harvest_total_hectares,
    v3_calc.percentage((COALESCE(sum(
        CASE
            WHEN ((lb.tons IS NOT NULL) AND (lb.tons > (0)::numeric)) THEN lb.hectares
            ELSE (0)::double precision
        END), (0)::double precision))::numeric, (COALESCE(sum(lb.hectares), (0)::double precision))::numeric) AS harvest_progress_pct,
    COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision) AS executed_costs_usd,
    v3_calc.total_budget_cost_for_project(p.id) AS budget_cost_usd,
    v3_calc.percentage((COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision))::numeric, v3_calc.total_budget_cost_for_project(p.id)) AS costs_progress_pct,
    COALESCE(sum(v3_calc.income_net_total_for_lot(lb.lot_id)), (0)::numeric) AS income_usd,
    v3_calc.operating_result_total_for_project(p.id) AS operating_result_usd,
    v3_calc.total_invested_cost_for_project(p.id) AS operating_result_total_costs_usd,
    v3_calc.renta_pct(v3_calc.operating_result_total_for_project(p.id), v3_calc.total_costs_for_project(p.id)) AS operating_result_pct,
    w_min.start_date,
    w_max.end_date,
    v3_calc.calculate_campaign_closing_date(w_max.end_date) AS campaign_closing_date,
    w_min.first_workorder_id,
    w_max.last_workorder_id,
    lsc.last_stock_count_date
   FROM ((((public.projects p
     LEFT JOIN lots_base lb ON ((lb.project_id = p.id)))
     LEFT JOIN w_min ON ((w_min.project_id = p.id)))
     LEFT JOIN w_max ON ((w_max.project_id = p.id)))
     LEFT JOIN last_stock_count lsc ON ((lsc.project_id = p.id)))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.customer_id, p.id, p.campaign_id, p.admin_cost, w_min.start_date, w_max.end_date, w_min.first_workorder_id, w_max.last_workorder_id, lsc.last_stock_count_date;


--
-- Name: v3_dashboard_contributions_progress; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_dashboard_contributions_progress AS
 SELECT p.id AS project_id,
    pi.investor_id,
    i.name AS investor_name,
    pi.percentage AS investor_percentage_pct,
    (pi.percentage)::numeric AS contributions_progress_pct
   FROM ((public.projects p
     JOIN public.project_investors pi ON (((pi.project_id = p.id) AND (pi.deleted_at IS NULL))))
     JOIN public.investors i ON (((i.id = pi.investor_id) AND (i.deleted_at IS NULL))))
  WHERE (p.deleted_at IS NULL);


--
-- Name: v3_dashboard_crop_incidence; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_dashboard_crop_incidence AS
 WITH lot_base AS (
         SELECT l.id AS lot_id,
            f.project_id,
            l.current_crop_id,
            c.name AS crop_name,
            l.hectares
           FROM ((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             LEFT JOIN public.crops c ON (((c.id = l.current_crop_id) AND (c.deleted_at IS NULL))))
          WHERE ((l.deleted_at IS NULL) AND (l.hectares IS NOT NULL) AND (l.hectares > (0)::double precision))
        ), by_crop AS (
         SELECT lot_base.project_id,
            lot_base.current_crop_id,
            lot_base.crop_name,
            (sum(lot_base.hectares))::numeric AS crop_hectares,
            v3_calc.total_costs_for_crop(lot_base.project_id, lot_base.current_crop_id) AS crop_costs_usd
           FROM lot_base
          WHERE (lot_base.current_crop_id IS NOT NULL)
          GROUP BY lot_base.project_id, lot_base.current_crop_id, lot_base.crop_name
        ), total_by_project AS (
         SELECT by_crop.project_id,
            (sum(by_crop.crop_costs_usd))::numeric AS total_costs_usd
           FROM by_crop
          GROUP BY by_crop.project_id
        )
 SELECT bc.project_id,
    bc.current_crop_id,
    bc.crop_name,
    bc.crop_hectares,
    v3_calc.percentage((bc.crop_costs_usd)::numeric, t.total_costs_usd) AS crop_incidence_pct,
    (v3_calc.cost_per_ha_for_crop(bc.project_id, bc.current_crop_id))::numeric AS cost_per_ha_usd
   FROM (by_crop bc
     JOIN total_by_project t ON ((t.project_id = bc.project_id)));


--
-- Name: v3_dashboard_management_balance; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_dashboard_management_balance AS
 WITH lots_base AS (
         SELECT l.id AS lot_id,
            f.project_id,
            l.hectares
           FROM (public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
          WHERE (l.deleted_at IS NULL)
        )
 SELECT p.id AS project_id,
    COALESCE(sum(v3_calc.income_net_total_for_lot(lb.lot_id)), (0)::numeric) AS income_usd,
    COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision) AS costos_directos_ejecutados_usd,
    v3_calc.direct_costs_invested_for_project(p.id) AS costos_directos_invertidos_usd,
    COALESCE(sum((v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision) AS arriendo_invertidos_usd,
    COALESCE(sum((v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision) AS estructura_invertidos_usd,
    COALESCE(sum((v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision) AS operating_result_usd,
    v3_calc.renta_pct(COALESCE(sum((v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision), COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision)) AS operating_result_pct
   FROM (public.projects p
     LEFT JOIN lots_base lb ON ((lb.project_id = p.id)))
  WHERE (p.deleted_at IS NULL)
  GROUP BY p.id;


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
-- Name: v3_investor_contribution_data_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_investor_contribution_data_view AS
 WITH project_base AS (
         SELECT p.id AS project_id,
            p.name AS project_name,
            p.customer_id,
            c.name AS customer_name,
            p.campaign_id,
            cam.name AS campaign_name,
            (COALESCE(sum(l.hectares), (0)::double precision))::numeric AS surface_total_ha,
            (COALESCE(sum((v3_calc.rent_per_ha_for_lot(l.id) * l.hectares)), (0)::double precision))::numeric AS lease_fixed_usd,
                CASE
                    WHEN (COALESCE(sum((v3_calc.rent_per_ha_for_lot(l.id) * l.hectares)), (0)::double precision) > (0)::double precision) THEN true
                    ELSE false
                END AS lease_is_fixed,
            (COALESCE(sum((v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares)), (0)::double precision))::numeric AS admin_per_ha_usd,
            (COALESCE(sum((v3_calc.admin_cost_per_ha_for_lot(l.id) * l.hectares)), (0)::double precision))::numeric AS admin_total_usd
           FROM ((((public.projects p
             JOIN public.customers c ON (((p.customer_id = c.id) AND (c.deleted_at IS NULL))))
             JOIN public.campaigns cam ON (((p.campaign_id = cam.id) AND (cam.deleted_at IS NULL))))
             LEFT JOIN public.fields f ON (((f.project_id = p.id) AND (f.deleted_at IS NULL))))
             LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
          WHERE (p.deleted_at IS NULL)
          GROUP BY p.id, p.name, p.customer_id, c.name, p.campaign_id, cam.name
        ), contribution_categories AS (
         SELECT pb_1.project_id,
            (COALESCE(( SELECT sum(((wi.total_used)::double precision * s.price)) AS sum
                   FROM (((public.workorders w
                     JOIN public.workorder_items wi ON ((w.id = wi.workorder_id)))
                     JOIN public.supplies s ON (((wi.supply_id = s.id) AND (s.deleted_at IS NULL))))
                     JOIN public.categories cat ON ((s.category_id = cat.id)))
                  WHERE ((w.project_id = pb_1.project_id) AND (w.deleted_at IS NULL) AND (cat.type_id = 2) AND ((cat.name)::text = ANY ((ARRAY['Coadyuvantes'::character varying, 'Curasemillas'::character varying, 'Herbicidas'::character varying, 'Insecticidas'::character varying, 'Fungicidas'::character varying, 'Otros Insumos'::character varying])::text[])))), (0)::double precision))::numeric AS agrochemicals_total,
            (COALESCE(( SELECT sum(((wi.total_used)::double precision * s.price)) AS sum
                   FROM (((public.workorders w
                     JOIN public.workorder_items wi ON ((w.id = wi.workorder_id)))
                     JOIN public.supplies s ON (((wi.supply_id = s.id) AND (s.deleted_at IS NULL))))
                     JOIN public.categories cat ON ((s.category_id = cat.id)))
                  WHERE ((w.project_id = pb_1.project_id) AND (w.deleted_at IS NULL) AND ((cat.name)::text = 'Semilla'::text) AND (cat.type_id = 1))), (0)::double precision))::numeric AS seeds_total,
            COALESCE(( SELECT sum((l.price * w.effective_area)) AS sum
                   FROM ((public.workorders w
                     JOIN public.labors l ON (((w.labor_id = l.id) AND (l.deleted_at IS NULL))))
                     JOIN public.categories cat ON ((l.category_id = cat.id)))
                  WHERE ((w.project_id = pb_1.project_id) AND (w.deleted_at IS NULL) AND (cat.type_id = 4) AND ((cat.name)::text = ANY ((ARRAY['Pulverización'::character varying, 'Otras Labores'::character varying])::text[])))), (0)::numeric) AS general_labors_total,
            COALESCE(( SELECT sum((l.price * w.effective_area)) AS sum
                   FROM ((public.workorders w
                     JOIN public.labors l ON (((w.labor_id = l.id) AND (l.deleted_at IS NULL))))
                     JOIN public.categories cat ON ((l.category_id = cat.id)))
                  WHERE ((w.project_id = pb_1.project_id) AND (w.deleted_at IS NULL) AND ((cat.name)::text = 'Siembra'::text) AND (cat.type_id = 4))), (0)::numeric) AS sowing_total,
            COALESCE(( SELECT sum((l.price * w.effective_area)) AS sum
                   FROM ((public.workorders w
                     JOIN public.labors l ON (((w.labor_id = l.id) AND (l.deleted_at IS NULL))))
                     JOIN public.categories cat ON ((l.category_id = cat.id)))
                  WHERE ((w.project_id = pb_1.project_id) AND (w.deleted_at IS NULL) AND ((cat.name)::text = 'Riego'::text) AND (cat.type_id = 4))), (0)::numeric) AS irrigation_total,
            pb_1.lease_fixed_usd AS rent_total,
            pb_1.admin_total_usd AS administration_total
           FROM project_base pb_1
        ), investor_contributions AS (
         SELECT pb_1.project_id,
            pi.investor_id,
            i.name AS investor_name,
            pi.percentage,
            ((cc.agrochemicals_total * (pi.percentage)::numeric) / (100)::numeric) AS agrochemicals_contribution,
            ((cc.seeds_total * (pi.percentage)::numeric) / (100)::numeric) AS seeds_contribution,
            ((cc.general_labors_total * (pi.percentage)::numeric) / (100)::numeric) AS general_labors_contribution,
            ((cc.sowing_total * (pi.percentage)::numeric) / (100)::numeric) AS sowing_contribution,
            ((cc.irrigation_total * (pi.percentage)::numeric) / (100)::numeric) AS irrigation_contribution,
            0 AS rent_contribution,
            0 AS administration_contribution
           FROM (((project_base pb_1
             JOIN contribution_categories cc ON ((cc.project_id = pb_1.project_id)))
             JOIN public.project_investors pi ON ((pi.project_id = pb_1.project_id)))
             JOIN public.investors i ON ((pi.investor_id = i.id)))
        ), contributions_data AS (
         SELECT pb_1.project_id,
            ( SELECT COALESCE(jsonb_agg(jsonb_build_object('type', cat_data.type, 'label', cat_data.label, 'total_usd', cat_data.total_usd, 'total_usd_ha',
                        CASE
                            WHEN (pb_1.surface_total_ha > (0)::numeric) THEN (cat_data.total_usd / pb_1.surface_total_ha)
                            ELSE (0)::numeric
                        END, 'investors', cat_data.investors, 'requires_manual_attribution', cat_data.requires_manual_attribution)), '[]'::jsonb) AS "coalesce"
                   FROM ( SELECT 'agrochemicals'::text AS type,
                            'Agroquímicos'::text AS label,
                            cc.agrochemicals_total AS total_usd,
                            COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ic.investor_id, 'investor_name', ic.investor_name, 'amount_usd', ic.agrochemicals_contribution, 'share_pct',
CASE
 WHEN (cc.agrochemicals_total > (0)::numeric) THEN ((ic.agrochemicals_contribution / cc.agrochemicals_total) * (100)::numeric)
 ELSE (0)::numeric
END)) AS jsonb_agg
                                   FROM investor_contributions ic
                                  WHERE (ic.project_id = pb_1.project_id)), '[]'::jsonb) AS investors,
                            false AS requires_manual_attribution
                           FROM contribution_categories cc
                          WHERE (cc.project_id = pb_1.project_id)
                        UNION ALL
                         SELECT 'seeds'::text AS type,
                            'Semilla'::text AS label,
                            cc.seeds_total AS total_usd,
                            COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ic.investor_id, 'investor_name', ic.investor_name, 'amount_usd', ic.seeds_contribution, 'share_pct',
CASE
 WHEN (cc.seeds_total > (0)::numeric) THEN ((ic.seeds_contribution / cc.seeds_total) * (100)::numeric)
 ELSE (0)::numeric
END)) AS jsonb_agg
                                   FROM investor_contributions ic
                                  WHERE (ic.project_id = pb_1.project_id)), '[]'::jsonb) AS investors,
                            false AS requires_manual_attribution
                           FROM contribution_categories cc
                          WHERE (cc.project_id = pb_1.project_id)
                        UNION ALL
                         SELECT 'general_labors'::text AS type,
                            'Labores Generales'::text AS label,
                            cc.general_labors_total AS total_usd,
                            COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ic.investor_id, 'investor_name', ic.investor_name, 'amount_usd', ic.general_labors_contribution, 'share_pct',
CASE
 WHEN (cc.general_labors_total > (0)::numeric) THEN ((ic.general_labors_contribution / cc.general_labors_total) * (100)::numeric)
 ELSE (0)::numeric
END)) AS jsonb_agg
                                   FROM investor_contributions ic
                                  WHERE (ic.project_id = pb_1.project_id)), '[]'::jsonb) AS investors,
                            false AS requires_manual_attribution
                           FROM contribution_categories cc
                          WHERE (cc.project_id = pb_1.project_id)
                        UNION ALL
                         SELECT 'sowing'::text AS type,
                            'Siembra'::text AS label,
                            cc.sowing_total AS total_usd,
                            COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ic.investor_id, 'investor_name', ic.investor_name, 'amount_usd', ic.sowing_contribution, 'share_pct',
CASE
 WHEN (cc.sowing_total > (0)::numeric) THEN ((ic.sowing_contribution / cc.sowing_total) * (100)::numeric)
 ELSE (0)::numeric
END)) AS jsonb_agg
                                   FROM investor_contributions ic
                                  WHERE (ic.project_id = pb_1.project_id)), '[]'::jsonb) AS investors,
                            false AS requires_manual_attribution
                           FROM contribution_categories cc
                          WHERE (cc.project_id = pb_1.project_id)
                        UNION ALL
                         SELECT 'irrigation'::text AS type,
                            'Riego'::text AS label,
                            cc.irrigation_total AS total_usd,
                            COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', ic.investor_id, 'investor_name', ic.investor_name, 'amount_usd', ic.irrigation_contribution, 'share_pct',
CASE
 WHEN (cc.irrigation_total > (0)::numeric) THEN ((ic.irrigation_contribution / cc.irrigation_total) * (100)::numeric)
 ELSE (0)::numeric
END)) AS jsonb_agg
                                   FROM investor_contributions ic
                                  WHERE (ic.project_id = pb_1.project_id)), '[]'::jsonb) AS investors,
                            false AS requires_manual_attribution
                           FROM contribution_categories cc
                          WHERE (cc.project_id = pb_1.project_id)
                        UNION ALL
                         SELECT 'capitalizable_lease'::text AS type,
                            'Arriendo Capitalizable'::text AS label,
                            cc.rent_total AS total_usd,
                            '[]'::jsonb AS investors,
                            true AS requires_manual_attribution
                           FROM contribution_categories cc
                          WHERE ((cc.project_id = pb_1.project_id) AND (cc.rent_total > (0)::numeric))
                        UNION ALL
                         SELECT 'administration_structure'::text AS type,
                            'Administración y Estructura'::text AS label,
                            cc.administration_total AS total_usd,
                            '[]'::jsonb AS investors,
                            true AS requires_manual_attribution
                           FROM contribution_categories cc
                          WHERE ((cc.project_id = pb_1.project_id) AND (cc.administration_total > (0)::numeric))) cat_data) AS contributions_data
           FROM project_base pb_1
        ), project_totals AS (
         SELECT pb_1.project_id,
            ((((((cc.agrochemicals_total + cc.seeds_total) + cc.general_labors_total) + cc.sowing_total) + cc.irrigation_total) + cc.rent_total) + cc.administration_total) AS total_contributions
           FROM (project_base pb_1
             JOIN contribution_categories cc ON ((cc.project_id = pb_1.project_id)))
        ), comparison_data AS (
         SELECT pb_1.project_id,
            ( SELECT COALESCE(jsonb_agg(jsonb_build_object('investor_id', pi.investor_id, 'investor_name', i.name, 'agreed_share_pct', pi.percentage, 'agreed_usd', (pt.total_contributions * ((pi.percentage / 100))::numeric), 'actual_usd', COALESCE(( SELECT sum(((((((ic.agrochemicals_contribution + ic.seeds_contribution) + ic.general_labors_contribution) + ic.sowing_contribution) + ic.irrigation_contribution) + (ic.rent_contribution)::numeric) + (ic.administration_contribution)::numeric)) AS sum
                           FROM investor_contributions ic
                          WHERE ((ic.project_id = pb_1.project_id) AND (ic.investor_id = pi.investor_id))), (0)::numeric), 'adjustment_usd', (COALESCE(( SELECT sum(((((((ic.agrochemicals_contribution + ic.seeds_contribution) + ic.general_labors_contribution) + ic.sowing_contribution) + ic.irrigation_contribution) + (ic.rent_contribution)::numeric) + (ic.administration_contribution)::numeric)) AS sum
                           FROM investor_contributions ic
                          WHERE ((ic.project_id = pb_1.project_id) AND (ic.investor_id = pi.investor_id))), (0)::numeric) - (pt.total_contributions * ((pi.percentage / 100))::numeric)))), '[]'::jsonb) AS "coalesce"
                   FROM ((public.project_investors pi
                     JOIN public.investors i ON ((pi.investor_id = i.id)))
                     JOIN project_totals pt ON ((pt.project_id = pb_1.project_id)))
                  WHERE (pi.project_id = pb_1.project_id)) AS comparison_data
           FROM project_base pb_1
        ), harvest_data AS (
         SELECT pb_1.project_id,
            jsonb_build_object('total_harvest_usd', COALESCE(( SELECT sum(v3_calc.income_net_total_for_lot(l.id)) AS sum
                   FROM (public.lots l
                     JOIN public.fields f ON ((l.field_id = f.id)))
                  WHERE ((f.project_id = pb_1.project_id) AND (l.deleted_at IS NULL))), (0)::numeric), 'total_harvest_usd_ha',
                CASE
                    WHEN (pb_1.surface_total_ha > (0)::numeric) THEN (COALESCE(( SELECT sum(v3_calc.income_net_total_for_lot(l.id)) AS sum
                       FROM (public.lots l
                         JOIN public.fields f ON ((l.field_id = f.id)))
                      WHERE ((f.project_id = pb_1.project_id) AND (l.deleted_at IS NULL))), (0)::numeric) / pb_1.surface_total_ha)
                    ELSE (0)::numeric
                END, 'investors', COALESCE(( SELECT jsonb_agg(jsonb_build_object('investor_id', pi.investor_id, 'investor_name', i.name, 'paid_usd', 0, 'agreed_usd', (COALESCE(( SELECT sum(v3_calc.income_net_total_for_lot(l.id)) AS sum
                           FROM (public.lots l
                             JOIN public.fields f ON ((l.field_id = f.id)))
                          WHERE ((f.project_id = pb_1.project_id) AND (l.deleted_at IS NULL))), (0)::numeric) * ((pi.percentage / 100))::numeric), 'adjustment_usd', ((0)::numeric - (COALESCE(( SELECT sum(v3_calc.income_net_total_for_lot(l.id)) AS sum
                           FROM (public.lots l
                             JOIN public.fields f ON ((l.field_id = f.id)))
                          WHERE ((f.project_id = pb_1.project_id) AND (l.deleted_at IS NULL))), (0)::numeric) * ((pi.percentage / 100))::numeric)))) AS jsonb_agg
                   FROM (public.project_investors pi
                     JOIN public.investors i ON ((pi.investor_id = i.id)))
                  WHERE (pi.project_id = pb_1.project_id)), '[]'::jsonb)) AS harvest_data
           FROM project_base pb_1
        )
 SELECT pb.project_id,
    pb.project_name,
    pb.customer_id,
    pb.customer_name,
    pb.campaign_id,
    pb.campaign_name,
    pb.surface_total_ha,
    pb.lease_fixed_usd,
    pb.lease_is_fixed,
    pb.admin_per_ha_usd,
    pb.admin_total_usd,
    cd.contributions_data,
    compd.comparison_data,
    hd.harvest_data
   FROM (((project_base pb
     LEFT JOIN contributions_data cd ON ((cd.project_id = pb.project_id)))
     LEFT JOIN comparison_data compd ON ((compd.project_id = pb.project_id)))
     LEFT JOIN harvest_data hd ON ((hd.project_id = pb.project_id)));


--
-- Name: v3_labor_list; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_labor_list AS
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
    cat_lb.id AS labor_category_id,
    cat_lb.name AS labor_category_name,
    w.contractor,
    lb.contractor_name,
    w.effective_area AS surface_ha,
    lb.price AS cost_per_ha,
    v3_calc.labor_cost((lb.price)::numeric, (w.effective_area)::numeric) AS total_labor_cost,
    v3_calc.dollar_average_for_month(w.project_id, w.date) AS dollar_average_month,
    w.investor_id,
    i.name AS investor_name
   FROM (((((((public.workorders w
     JOIN public.projects p ON (((p.id = w.project_id) AND (p.deleted_at IS NULL))))
     JOIN public.fields f ON (((f.id = w.field_id) AND (f.deleted_at IS NULL))))
     LEFT JOIN public.lots l ON (((l.id = w.lot_id) AND (l.deleted_at IS NULL))))
     LEFT JOIN public.crops c ON (((c.id = w.crop_id) AND (c.deleted_at IS NULL))))
     JOIN public.labors lb ON (((lb.id = w.labor_id) AND (lb.deleted_at IS NULL))))
     LEFT JOIN public.categories cat_lb ON (((cat_lb.id = lb.category_id) AND (cat_lb.deleted_at IS NULL))))
     LEFT JOIN public.investors i ON (((i.id = w.investor_id) AND (i.deleted_at IS NULL))))
  WHERE ((w.deleted_at IS NULL) AND (w.effective_area IS NOT NULL) AND (w.effective_area > (0)::numeric) AND (lb.price IS NOT NULL));


--
-- Name: v3_labor_metrics; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_labor_metrics AS
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
            sum(v3_calc.labor_cost(wo.labor_price_per_ha, wo.effective_area)) AS total_labor_cost,
            min(wo.date) AS first_workorder_date,
            max(wo.date) AS last_workorder_date
           FROM wo
          GROUP BY wo.project_id, wo.field_id
        )
 SELECT project_id,
    field_id,
    surface_ha,
    total_labor_cost,
    v3_calc.cost_per_ha(total_labor_cost, surface_ha) AS avg_labor_cost_per_ha,
    total_workorders,
    first_workorder_date,
    last_workorder_date
   FROM agg a;


--
-- Name: v3_lot_list; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_lot_list AS
 WITH base AS (
         SELECT l.id AS lot_id,
            l.name AS lot_name,
            l.variety,
            l.season,
            l.hectares,
            l.tons,
            l.sowing_date,
            l.updated_at,
            f.id AS field_id,
            f.name AS field_name,
            f.project_id,
            p.name AS project_name,
            p.admin_cost AS project_admin_cost,
            l.previous_crop_id,
            pc.name AS previous_crop,
            l.current_crop_id,
            cc.name AS current_crop
           FROM ((((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
             LEFT JOIN public.crops pc ON (((pc.id = l.previous_crop_id) AND (pc.deleted_at IS NULL))))
             LEFT JOIN public.crops cc ON (((cc.id = l.current_crop_id) AND (cc.deleted_at IS NULL))))
          WHERE (l.deleted_at IS NULL)
        ), wo_dates AS (
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
                END) AS lot_harvest_date
           FROM (public.workorders w
             JOIN public.labors lb ON ((lb.id = w.labor_id)))
          WHERE ((w.deleted_at IS NULL) AND (lb.deleted_at IS NULL))
          GROUP BY w.lot_id
        )
 SELECT b.project_id,
    b.project_name,
    b.field_id,
    b.field_name,
    b.lot_id AS id,
    b.lot_name,
    b.variety,
    b.season,
    b.previous_crop_id,
    b.previous_crop,
    b.current_crop_id,
    b.current_crop,
    b.hectares,
    b.updated_at,
    v3_calc.seeded_area_for_lot(b.lot_id) AS sowed_area_ha,
    v3_calc.harvested_area_for_lot(b.lot_id) AS harvested_area_ha,
    v3_calc.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,
    v3_calc.cost_per_ha((COALESCE(v3_calc.direct_cost_for_lot(b.lot_id), (0)::double precision))::numeric, (b.hectares)::numeric) AS cost_usd_per_ha,
    (v3_calc.income_net_per_ha_for_lot(b.lot_id))::numeric AS income_net_per_ha_usd,
    (v3_calc.rent_per_ha_for_lot(b.lot_id))::numeric AS rent_per_ha_usd,
    COALESCE(b.project_admin_cost, (0)::numeric) AS admin_cost_per_ha_usd,
    (v3_calc.active_total_per_ha_for_lot(b.lot_id))::numeric AS active_total_per_ha_usd,
    (v3_calc.operating_result_per_ha_for_lot(b.lot_id))::numeric AS operating_result_per_ha_usd,
    v3_calc.income_net_total_for_lot(b.lot_id) AS income_net_total_usd,
    (COALESCE(v3_calc.direct_cost_for_lot(b.lot_id), (0)::double precision))::numeric AS direct_cost_total_usd,
    ((v3_calc.rent_per_ha_for_lot(b.lot_id) * b.hectares))::numeric AS rent_total_usd,
    COALESCE(b.project_admin_cost, (0)::numeric) AS admin_total_usd,
    ((v3_calc.active_total_per_ha_for_lot(b.lot_id) * b.hectares))::numeric AS active_total_usd,
    ((v3_calc.operating_result_per_ha_for_lot(b.lot_id) * b.hectares))::numeric AS operating_result_total_usd,
    wd.lot_sowing_date,
    wd.lot_harvest_date,
    b.tons,
    b.sowing_date AS raw_sowing_date
   FROM (base b
     LEFT JOIN wo_dates wd ON ((wd.lot_id = b.lot_id)));


--
-- Name: v3_lot_metrics; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_lot_metrics AS
 WITH lot_base AS (
         SELECT l.id AS lot_id,
            f.id AS field_id,
            f.project_id,
            l.hectares,
            COALESCE(sum(
                CASE
                    WHEN (lb.category_id = 9) THEN w.effective_area
                    ELSE (0)::numeric
                END), (0)::numeric) AS sowed_area_ha,
            COALESCE(sum(
                CASE
                    WHEN (lb.category_id = 13) THEN w.effective_area
                    ELSE (0)::numeric
                END), (0)::numeric) AS harvested_area_ha
           FROM (((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             LEFT JOIN public.workorders w ON (((w.lot_id = l.id) AND (w.deleted_at IS NULL))))
             LEFT JOIN public.labors lb ON (((lb.id = w.labor_id) AND (lb.deleted_at IS NULL))))
          WHERE (l.deleted_at IS NULL)
          GROUP BY l.id, f.id, f.project_id, l.hectares
        ), field_total_area AS (
         SELECT f.id AS field_id,
            (COALESCE(sum(l.hectares), (0)::double precision))::numeric AS total_hectares
           FROM (public.fields f
             LEFT JOIN public.lots l ON (((l.field_id = f.id) AND (l.deleted_at IS NULL))))
          WHERE (f.deleted_at IS NULL)
          GROUP BY f.id
        ), lot_worked_area AS (
         SELECT l.id AS lot_id,
            COALESCE(sum(w.effective_area), (0)::numeric) AS worked_hectares
           FROM (public.lots l
             LEFT JOIN public.workorders w ON (((w.lot_id = l.id) AND (w.deleted_at IS NULL))))
          WHERE (l.deleted_at IS NULL)
          GROUP BY l.id
        )
 SELECT b.project_id,
    b.field_id,
    b.lot_id,
    b.hectares,
    b.sowed_area_ha,
    b.harvested_area_ha,
    v3_calc.yield_tn_per_ha_for_lot(b.lot_id) AS yield_tn_per_ha,
    COALESCE(v3_calc.labor_cost_for_lot(b.lot_id), (0)::numeric) AS labor_cost_usd,
    (COALESCE(v3_calc.supply_cost_for_lot(b.lot_id), (0)::double precision))::numeric AS supplies_cost_usd,
    (COALESCE(v3_calc.direct_cost_for_lot(b.lot_id), (0)::double precision))::numeric AS direct_cost_usd,
    COALESCE(v3_calc.income_net_total_for_lot(b.lot_id), (0)::numeric) AS income_net_total_usd,
    (COALESCE(v3_calc.income_net_per_ha_for_lot(b.lot_id), (0)::double precision))::numeric AS income_net_per_ha_usd,
    COALESCE(p.admin_cost, (0)::numeric) AS admin_cost_per_ha_usd,
    (COALESCE(v3_calc.rent_per_ha_for_lot(b.lot_id), (0)::double precision))::numeric AS rent_per_ha_usd,
    (COALESCE(v3_calc.active_total_per_ha_for_lot(b.lot_id), (0)::double precision))::numeric AS active_total_per_ha_usd,
    (COALESCE(v3_calc.operating_result_per_ha_for_lot(b.lot_id), (0)::double precision))::numeric AS operating_result_per_ha_usd,
    COALESCE(p.admin_cost, (0)::numeric) AS admin_total_usd,
    ((COALESCE(v3_calc.rent_per_ha_for_lot(b.lot_id), (0)::double precision) * b.hectares))::numeric AS rent_total_usd,
    ((COALESCE(v3_calc.active_total_per_ha_for_lot(b.lot_id), (0)::double precision) * b.hectares))::numeric AS active_total_usd,
    ((COALESCE(v3_calc.operating_result_per_ha_for_lot(b.lot_id), (0)::double precision) * b.hectares))::numeric AS operating_result_total_usd,
    v3_calc.cost_per_ha((COALESCE(v3_calc.direct_cost_for_lot(b.lot_id), (0)::double precision))::numeric, COALESCE(fta.total_hectares, (0)::numeric)) AS direct_cost_per_ha_usd,
    COALESCE(fta.total_hectares, (0)::numeric) AS superficie_total
   FROM (((lot_base b
     LEFT JOIN field_total_area fta ON ((fta.field_id = b.field_id)))
     LEFT JOIN lot_worked_area lwa ON ((lwa.lot_id = b.lot_id)))
     LEFT JOIN public.projects p ON (((p.id = b.project_id) AND (p.deleted_at IS NULL))));


--
-- Name: v3_report_field_crop_metrics_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_report_field_crop_metrics_view AS
 WITH lot_base AS (
         SELECT l.id AS lot_id,
            f.project_id,
            f.id AS field_id,
            f.name AS field_name,
            l.current_crop_id,
            c.name AS crop_name,
            l.hectares,
            COALESCE(l.tons, (0)::numeric) AS tons,
            (v3_calc.seeded_area(l.sowing_date, (l.hectares)::numeric))::double precision AS seeded_area_ha,
            (v3_calc.harvested_area(l.tons, (l.hectares)::numeric))::double precision AS harvested_area_ha
           FROM (((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
             LEFT JOIN public.crops c ON (((c.id = l.current_crop_id) AND (c.deleted_at IS NULL))))
          WHERE ((l.deleted_at IS NULL) AND (l.hectares > (0)::double precision))
        )
 SELECT lb.project_id,
    lb.field_id,
    (lb.field_name)::text AS field_name,
    lb.current_crop_id,
    (lb.crop_name)::text AS crop_name,
    COALESCE(sum(v3_calc.income_net_total_for_lot(lb.lot_id)), (0)::numeric) AS income_usd,
    COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision) AS direct_costs_executed_usd,
    ((COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision) + COALESCE(sum((v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision)) + COALESCE(sum((v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision)) AS direct_costs_invested_usd,
    COALESCE(sum((v3_calc.rent_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision) AS rent_invested_usd,
    COALESCE(sum((v3_calc.admin_cost_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision) AS structure_invested_usd,
    COALESCE(sum((v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision) AS operating_result_usd,
    v3_calc.renta_pct(COALESCE(sum((v3_calc.operating_result_per_ha_for_lot(lb.lot_id) * lb.hectares)), (0)::double precision), COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision)) AS operating_result_pct
   FROM (lot_base lb
     LEFT JOIN public.crop_commercializations cc ON (((cc.project_id = lb.project_id) AND (cc.crop_id = lb.current_crop_id) AND (cc.deleted_at IS NULL))))
  WHERE (lb.current_crop_id IS NOT NULL)
  GROUP BY lb.project_id, lb.field_id, lb.field_name, lb.current_crop_id, lb.crop_name;


--
-- Name: v3_report_summary_results_view; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_report_summary_results_view AS
 WITH lot_base AS (
         SELECT l.id AS lot_id,
            f.project_id,
            l.current_crop_id,
            c.name AS crop_name,
            l.hectares,
            COALESCE(l.tons, (0)::numeric) AS tons,
            COALESCE(( SELECT sum(w.effective_area) AS sum
                   FROM ((public.workorders w
                     JOIN public.labors lab ON ((w.labor_id = lab.id)))
                     JOIN public.categories cat ON ((lab.category_id = cat.id)))
                  WHERE ((w.lot_id = l.id) AND (w.deleted_at IS NULL) AND ((cat.name)::text = 'Siembra'::text) AND (cat.type_id = 4))), (0)::numeric) AS seeded_area_ha
           FROM (((public.lots l
             JOIN public.fields f ON (((f.id = l.field_id) AND (f.deleted_at IS NULL))))
             JOIN public.projects p ON (((p.id = f.project_id) AND (p.deleted_at IS NULL))))
             LEFT JOIN public.crops c ON (((c.id = l.current_crop_id) AND (c.deleted_at IS NULL))))
          WHERE ((l.deleted_at IS NULL) AND (l.hectares > (0)::double precision))
        ), by_crop AS (
         SELECT lb.project_id,
            lb.current_crop_id,
            (lb.crop_name)::text AS crop_name,
            COALESCE(sum(lb.seeded_area_ha), (0)::numeric) AS surface_ha,
            COALESCE(sum((lb.seeded_area_ha * ( SELECT COALESCE(cc.net_price, (0)::numeric) AS "coalesce"
                   FROM public.crop_commercializations cc
                  WHERE ((cc.project_id = lb.project_id) AND (cc.crop_id = lb.current_crop_id) AND (cc.deleted_at IS NULL))
                 LIMIT 1))), (0)::numeric) AS net_income_usd,
            (COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision))::numeric AS direct_costs_usd,
            (COALESCE(sum(((lb.seeded_area_ha)::double precision * v3_calc.rent_per_ha_for_lot(lb.lot_id))), (0)::double precision))::numeric AS rent_usd,
            (COALESCE(sum(((lb.seeded_area_ha)::double precision * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id))), (0)::double precision))::numeric AS structure_usd,
            (((COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision))::numeric + (COALESCE(sum(((lb.seeded_area_ha)::double precision * v3_calc.rent_per_ha_for_lot(lb.lot_id))), (0)::double precision))::numeric) + (COALESCE(sum(((lb.seeded_area_ha)::double precision * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id))), (0)::double precision))::numeric) AS total_invested_usd,
            (COALESCE(sum((lb.seeded_area_ha * ( SELECT COALESCE(cc.net_price, (0)::numeric) AS "coalesce"
                   FROM public.crop_commercializations cc
                  WHERE ((cc.project_id = lb.project_id) AND (cc.crop_id = lb.current_crop_id) AND (cc.deleted_at IS NULL))
                 LIMIT 1))), (0)::numeric) - (((COALESCE(sum(v3_calc.direct_cost_for_lot(lb.lot_id)), (0)::double precision))::numeric + (COALESCE(sum(((lb.seeded_area_ha)::double precision * v3_calc.rent_per_ha_for_lot(lb.lot_id))), (0)::double precision))::numeric) + (COALESCE(sum(((lb.seeded_area_ha)::double precision * v3_calc.admin_cost_per_ha_for_lot(lb.lot_id))), (0)::double precision))::numeric)) AS operating_result_usd
           FROM lot_base lb
          WHERE (lb.current_crop_id IS NOT NULL)
          GROUP BY lb.project_id, lb.current_crop_id, lb.crop_name
        ), project_totals AS (
         SELECT by_crop.project_id,
            sum(by_crop.surface_ha) AS total_surface_ha,
            sum(by_crop.net_income_usd) AS total_net_income_usd,
            sum(by_crop.direct_costs_usd) AS total_direct_costs_usd,
            sum(by_crop.rent_usd) AS total_rent_usd,
            sum(by_crop.structure_usd) AS total_structure_usd,
            sum(by_crop.total_invested_usd) AS total_invested_usd,
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
    (v3_calc.renta_pct((bc.operating_result_usd)::double precision, (bc.total_invested_usd)::double precision))::numeric AS crop_return_pct,
    pt.total_surface_ha,
    pt.total_net_income_usd,
    pt.total_direct_costs_usd,
    pt.total_rent_usd,
    pt.total_structure_usd,
    pt.total_invested_usd AS total_invested_project_usd,
    pt.total_operating_result_usd,
    (v3_calc.renta_pct((pt.total_operating_result_usd)::double precision, (pt.total_invested_usd)::double precision))::numeric AS project_return_pct
   FROM (by_crop bc
     JOIN project_totals pt ON ((pt.project_id = bc.project_id)))
  ORDER BY bc.project_id, bc.current_crop_id;


--
-- Name: v3_workorder_list; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_workorder_list AS
 WITH workorder_surface AS (
         SELECT w_1.id,
            w_1.effective_area AS surface_ha
           FROM public.workorders w_1
          WHERE ((w_1.deleted_at IS NULL) AND (w_1.effective_area IS NOT NULL) AND (w_1.effective_area > (0)::numeric))
        )
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
    ws.surface_ha,
    (COALESCE(s.name, ''::character varying))::character varying(100) AS supply_name,
    (COALESCE(wi.total_used, (0)::numeric))::numeric(18,6) AS consumption,
    (COALESCE(cat.name, ''::character varying))::character varying(250) AS category_name,
    (COALESCE(wi.final_dose, (0)::numeric))::numeric(18,6) AS dose_per_ha,
    COALESCE(s.price, (0)::double precision) AS unit_price,
        CASE
            WHEN ((wi.final_dose IS NOT NULL) AND (s.price IS NOT NULL)) THEN v3_calc.cost_per_ha((((wi.final_dose)::double precision * s.price))::numeric, (1)::numeric)
            ELSE (0)::numeric
        END AS supply_cost_per_ha,
        CASE
            WHEN ((wi.final_dose IS NOT NULL) AND (s.price IS NOT NULL) AND (ws.surface_ha IS NOT NULL)) THEN v3_calc.supply_cost((wi.final_dose)::double precision, (s.price)::numeric, ws.surface_ha)
            ELSE (0)::numeric
        END AS supply_total_cost
   FROM (((((((((((public.workorders w
     JOIN workorder_surface ws ON ((ws.id = w.id)))
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
-- Name: v3_workorder_metrics; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v3_workorder_metrics AS
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
            sum(v3_calc.labor_cost(base.labor_price, base.effective_area)) AS labor_cost_usd
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
            sum(v3_calc.supply_cost((wi.final_dose)::double precision, (s.price)::numeric, b.effective_area)) AS supplies_cost_usd
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
    v3_calc.cost_per_ha((COALESCE(lc.labor_cost_usd, (0)::numeric) + COALESCE(sm.supplies_cost_usd, (0)::numeric)), COALESCE(sur.surface_ha, (0)::numeric)) AS avg_cost_per_ha_usd,
    v3_calc.per_ha(COALESCE(sm.liters, (0)::numeric), COALESCE(sur.surface_ha, (0)::numeric)) AS liters_per_ha,
    v3_calc.per_ha(COALESCE(sm.kilograms, (0)::numeric), COALESCE(sur.surface_ha, (0)::numeric)) AS kilograms_per_ha
   FROM ((surface sur
     FULL JOIN labor_costs lc USING (project_id, field_id, lot_id))
     FULL JOIN supply_metrics sm USING (project_id, field_id, lot_id));


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
-- Name: user_logins id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_logins ALTER COLUMN id SET DEFAULT nextval('public.user_logins_id_seq'::regclass);


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
-- Data for Name: app_parameters; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.app_parameters (id, key, value, type, category, description, created_at, updated_at) FROM stdin;
1	unit_liters	Lt	string	units	Unit of measurement: Liters	2025-09-19 19:57:26.186905+00	2025-09-19 19:57:26.186905+00
2	unit_kilos	Kg	string	units	Unit of measurement: Kilograms	2025-09-19 19:57:26.186905+00	2025-09-19 19:57:26.186905+00
3	unit_hectares	Ha	string	units	Unit of measurement: Hectares	2025-09-19 19:57:26.186905+00	2025-09-19 19:57:26.186905+00
4	iva_percentage	0.105	decimal	calculations	VAT percentage for labors (10.5%)	2025-09-19 19:57:26.186905+00	2025-09-19 19:57:26.186905+00
5	campaign_closure_days	30	integer	calculations	Days for campaign closure after end date	2025-09-19 19:57:26.186905+00	2025-09-19 19:57:26.186905+00
6	default_fx_rate	1.0000	decimal	calculations	Default exchange rate USD/USD	2025-09-19 19:57:26.186905+00	2025-09-19 19:57:26.186905+00
\.


--
-- Data for Name: campaigns; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.campaigns (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	2024-2025	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
2	2025-2026	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
3	2026-2027	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
\.


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.categories (id, name, type_id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Semilla	1	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
2	Coadyuvantes	2	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
3	Curasemillas	2	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
4	Herbicidas	2	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
5	Insecticidas	2	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
6	Fungicidas	2	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
7	Otros Insumos	2	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
8	Fertilizantes	3	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
9	Siembra	4	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
10	Pulverización	4	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
11	Otras Labores	4	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
12	Riego	4	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
13	Cosecha	4	2025-09-19 19:55:54.496455+00	2025-09-19 19:55:54.496455+00	\N	\N	\N	\N
\.


--
-- Data for Name: crop_commercializations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.crop_commercializations (id, project_id, crop_id, board_price, freight_cost, commercial_cost, net_price, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	2	12	30.00	20.00	10	7.00	2025-09-22 10:18:15.525608	2025-09-22 10:18:15.525608	\N	1	1	\N
2	5	1	250.00	50.00	2	195.00	2025-09-23 13:02:03.470024	2025-09-23 13:02:03.470024	\N	1	1	\N
3	5	2	175.00	50.00	2	121.50	2025-09-23 13:02:03.470024	2025-09-23 13:02:03.470024	\N	1	1	\N
4	4	11	150.00	50.00	0	100.00	2025-09-23 13:13:30.500299	2025-09-23 14:06:51.594225	\N	1	1	\N
5	4	14	200.00	50.00	3	144.00	2025-09-23 13:14:08.81681	2025-09-23 14:06:52.233288	\N	1	1	\N
6	4	2	220.00	50.00	2	165.60	2025-09-23 13:14:08.81681	2025-09-23 14:06:52.74525	\N	1	1	\N
\.


--
-- Data for Name: crops; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.crops (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Soja	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
2	Maíz	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
3	Maíz PPM	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
4	Maíz color	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
5	Poroto	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
6	Poroto negro	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
7	Poroto blanco	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
8	Poroto rojo	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
9	Poroto Mung	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
10	Poroto crawberry	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
11	Trigo	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
12	Girasol	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
13	Sorgo	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
14	Garbanzo	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
15	Sésamo	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
\.


--
-- Data for Name: customers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.customers (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Oscar	2025-09-19 19:59:32.707748+00	2025-09-19 19:59:32.707748+00	\N	1	1	\N
2	Oscar Nuevo	2025-09-22 12:30:50.843868+00	2025-09-22 12:30:50.843868+00	\N	1	1	\N
3	SOALEN SRL	2025-09-23 12:14:25.138867+00	2025-09-23 12:14:25.138867+00	\N	1	1	\N
4	gugu prueba	2025-09-23 12:53:32.828187+00	2025-09-23 12:53:32.828187+00	\N	1	1	\N
\.


--
-- Data for Name: engineering_principles_documentation; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.engineering_principles_documentation (id, principle, description, implementation, migration_affected, created_at) FROM stdin;
1	DRY	Don't Repeat Yourself - Eliminar duplicación de cálculos	Funciones encapsuladas para cálculos comunes	{000035,000037,000040,000042,000045,000046,000048,000049,000050,000052,000053,000054,000055,000056,000057,000058,000059,000060,000061}	2025-09-19 19:57:27.896623
2	SSOT	Single Source of Truth - Centralizar definiciones de cálculos	Vistas base reutilizables para cálculos comunes	{000035,000037,000040,000042,000045,000046,000048,000049,000050,000052,000053,000054,000055,000056,000057,000058,000059,000060,000061}	2025-09-19 19:57:27.896623
3	View Composition	Composición de vistas - Vistas derivadas que consumen vistas base	Vistas derivadas que reutilizan vistas base	{000035,000037,000040,000042,000045,000046,000048,000049,000050,000052,000053,000054,000055,000056,000057,000058,000059,000060,000061}	2025-09-19 19:57:27.896623
4	Encapsulation	Encapsulación de lógica de negocio - Funciones para reglas de negocio	Funciones PL/pgSQL para encapsular lógica de negocio	{000035,000037,000040,000042,000045,000046,000048,000049,000050,000052,000053,000054,000055,000056,000057,000058,000059,000060,000061}	2025-09-19 19:57:27.896623
\.


--
-- Data for Name: fields; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.fields (id, name, project_id, lease_type_id, lease_type_percent, lease_type_value, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	campo 1	1	2	33	0	2025-09-19 19:59:36.279009+00	2025-09-19 19:59:36.279009+00	\N	1	1	\N
2	Los nogales	2	1	10	0	2025-09-22 10:11:31.515335+00	2025-09-22 10:11:31.515335+00	\N	1	1	\N
3	Campo 7	3	1	40	0	2025-09-22 12:30:54.286877+00	2025-09-22 12:30:54.286877+00	\N	1	1	\N
4	GRANEROS	4	3	0	100	2025-09-23 12:14:28.358668+00	2025-09-23 12:50:38.996321+00	\N	1	1	\N
5	campo	5	4	10	50	2025-09-23 12:53:35.987741+00	2025-09-23 12:53:35.987741+00	\N	1	1	\N
\.


--
-- Data for Name: fx_rates; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.fx_rates (id, currency_pair, rate, effective_date, created_at, updated_at) FROM stdin;
1	USDARS	1.0000	2025-09-19	2025-09-19 19:57:17.685451+00	2025-09-19 19:57:17.685451+00
\.


--
-- Data for Name: investors; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.investors (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Oscar	2025-09-19 19:59:34.562813+00	2025-09-19 19:59:34.562813+00	\N	1	1	\N
2	Eugenia	2025-09-19 19:59:35.029188+00	2025-09-19 19:59:35.029188+00	\N	1	1	\N
3	May	2025-09-22 10:11:30.378548+00	2025-09-22 10:11:30.378548+00	\N	1	1	\N
4	Pablo	2025-09-22 12:30:52.755958+00	2025-09-22 12:30:52.755958+00	\N	1	1	\N
5	COTY	2025-09-23 12:14:26.537776+00	2025-09-23 12:14:26.537776+00	\N	1	1	\N
6	ALVARO	2025-09-23 12:14:26.958059+00	2025-09-23 12:14:26.958059+00	\N	1	1	\N
7	agro lajitas	2025-09-23 12:53:34.353811+00	2025-09-23 12:53:34.353811+00	\N	1	1	\N
8	joa	2025-09-23 12:53:34.730559+00	2025-09-23 12:53:34.730559+00	\N	1	1	\N
\.


--
-- Data for Name: invoices; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.invoices (id, work_order_id, number, company, date, status, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	1	11	las moritas	2025-09-23 00:00:00	Pagada	2025-09-22 10:45:54.374749+00	2025-09-22 10:45:54.374749+00	\N	1	1	\N
\.


--
-- Data for Name: labor_categories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.labor_categories (id, name, type_id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Semilla	1	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
2	Coadyuvantes	2	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
3	Curasemillas	2	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
4	Herbicidas	2	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
5	Insecticidas	2	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
6	Fungicidas	2	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
7	Otros Insumos	2	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
8	Fertilizantes	3	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
9	Siembra	4	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
10	Pulverización	4	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
11	Otras Labores	4	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
12	Riego	4	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
13	Cosecha	4	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
\.


--
-- Data for Name: labor_types; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.labor_types (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Semilla	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
2	Agroquímicos	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
3	Fertilizantes	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
4	Labores	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
\.


--
-- Data for Name: labors; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.labors (id, project_id, name, category_id, price, contractor_name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	2	nogales 30	13	12.00	Juan	2025-09-22 10:17:06.280411+00	2025-09-22 10:17:06.280411+00	\N	1	1	\N
2	2	Pulverizacion	10	30.00	Carlos	2025-09-22 10:21:24.670686+00	2025-09-22 10:21:24.670686+00	\N	1	1	\N
3	2	siembra x	9	123456.00	Ramon	2025-09-23 01:33:36.476054+00	2025-09-23 01:33:36.476054+00	\N	1	1	\N
4	2	Cosecha	13	198.00	Roque	2025-09-23 02:50:33.021057+00	2025-09-23 02:50:33.021057+00	\N	1	1	\N
5	4	SIEMBRA TRIGO	9	50.00	SANCHEZ	2025-09-23 12:18:50.545295+00	2025-09-23 12:18:50.545295+00	\N	1	1	\N
6	4	SIEMBRA GARBANZO	9	50.00	MUÑO	2025-09-23 12:18:51.153489+00	2025-09-23 12:18:51.153489+00	\N	1	1	\N
7	4	PULVERIZACION TERRESTRE	10	10.00	SANCHEZ	2025-09-23 12:18:51.610563+00	2025-09-23 12:18:51.610563+00	\N	1	1	\N
8	4	PULEVERIZACION TERRESTRE	10	10.00	MUÑO	2025-09-23 12:18:52.065998+00	2025-09-23 12:18:52.065998+00	\N	1	1	\N
9	5	pulv	10	5.00	pepe	2025-09-23 12:55:51.186185+00	2025-09-23 12:55:51.186185+00	\N	1	1	\N
10	5	cosecha	13	70.00	c	2025-09-23 12:55:51.748833+00	2025-09-23 12:55:51.748833+00	\N	1	1	\N
11	5	siembra	9	40.00	s	2025-09-23 12:55:52.170663+00	2025-09-23 12:55:52.170663+00	\N	1	1	\N
12	4	COSECHA TRIGO	13	100.00	EUGE	2025-09-23 13:03:44.57511+00	2025-09-23 13:03:44.57511+00	\N	1	1	\N
\.


--
-- Data for Name: lease_types; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.lease_types (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	% INGRESO NETO	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
2	% UTILIDAD	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
3	ARRIENDO FIJO	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
4	ARRIENDO FIJO + % INGRESO NETO	2025-09-19 19:55:35.814665+00	2025-09-19 19:55:35.814665+00	\N	\N	\N	\N
\.


--
-- Data for Name: lot_dates; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.lot_dates (id, lot_id, sowing_date, harvest_date, sequence, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	6	2025-05-01	\N	1	2025-09-23 12:16:05.597729+00	2025-09-23 12:16:12.387853+00	\N	1	1	\N
2	6	\N	\N	2	2025-09-23 12:16:05.879146+00	2025-09-23 12:16:12.387853+00	\N	1	1	\N
3	6	\N	\N	3	2025-09-23 12:16:06.018985+00	2025-09-23 12:16:12.387853+00	\N	1	1	\N
7	5	2025-05-15	\N	1	2025-09-23 12:16:48.813533+00	2025-09-23 12:16:48.503503+00	\N	1	1	\N
8	5	2025-05-20	\N	2	2025-09-23 12:16:48.967224+00	2025-09-23 12:16:48.503503+00	\N	1	1	\N
9	5	\N	\N	3	2025-09-23 12:16:49.120717+00	2025-09-23 12:16:48.503503+00	\N	1	1	\N
10	9	2025-02-01	2025-09-08	1	2025-09-23 13:00:33.065941+00	2025-09-23 13:00:32.685157+00	\N	1	1	\N
11	9	\N	\N	2	2025-09-23 13:00:33.320029+00	2025-09-23 13:00:32.685157+00	\N	1	1	\N
12	9	\N	\N	3	2025-09-23 13:00:33.447307+00	2025-09-23 13:00:32.685157+00	\N	1	1	\N
13	8	2025-02-01	2025-09-08	1	2025-09-23 13:00:52.856319+00	2025-09-23 13:00:52.482829+00	\N	1	1	\N
14	8	\N	\N	2	2025-09-23 13:00:53.10558+00	2025-09-23 13:00:52.482829+00	\N	1	1	\N
15	8	\N	\N	3	2025-09-23 13:00:53.22999+00	2025-09-23 13:00:52.482829+00	\N	1	1	\N
\.


--
-- Data for Name: lots; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.lots (id, name, field_id, hectares, previous_crop_id, current_crop_id, season, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, variety, sowing_date, tons) FROM stdin;
1	1	1	200	1	3	3	2025-09-19 19:59:37.052539+00	2025-09-19 19:59:37.052539+00	\N	1	1	\N		\N	0
2	nogal 1	2	200	14	12	2	2025-09-22 10:11:32.143459+00	2025-09-22 10:11:32.143459+00	\N	1	1	\N		\N	0
3	Lote 1	3	200	1	2	3	2025-09-22 12:30:54.924001+00	2025-09-22 12:30:54.924001+00	\N	1	1	\N		\N	0
4	Lote 2	3	300	5	1	3	2025-09-22 12:30:54.924001+00	2025-09-22 12:30:54.924001+00	\N	1	1	\N		\N	0
7	LOTE 3	4	100	1	2	3	2025-09-23 12:50:39.627657+00	2025-09-23 12:50:39.627657+00	\N	1	1	\N		\N	0
6	LOTE 2	4	100	1	14	2	2025-09-23 12:14:29.198484+00	2025-09-23 12:50:40.136313+00	\N	1	1	\N	KIARA	\N	0
9	2	5	25	2	2	4	2025-09-23 12:53:36.615539+00	2025-09-23 13:01:18.879376+00	\N	1	1	\N	hgb	\N	50
8	1	5	50	1	1	4	2025-09-23 12:53:36.615539+00	2025-09-23 13:01:24.892129+00	\N	1	1	\N	hgb	\N	234
5	LOTE 1	4	100	1	11	2	2025-09-23 12:14:29.198484+00	2025-09-23 13:16:33.983497+00	\N	1	1	\N	KLEIN	\N	400
\.


--
-- Data for Name: managers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.managers (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Osar	2025-09-19 19:59:33.63402+00	2025-09-19 19:59:33.63402+00	\N	1	1	\N
2	Pablo	2025-09-19 19:59:34.100441+00	2025-09-19 19:59:34.100441+00	\N	1	1	\N
3	May	2025-09-22 12:30:51.610515+00	2025-09-22 12:30:51.610515+00	\N	1	1	\N
4	EUGENIA	2025-09-23 12:14:25.977972+00	2025-09-23 12:14:25.977972+00	\N	1	1	\N
5	agro lajitas	2025-09-23 12:53:33.591759+00	2025-09-23 12:53:33.591759+00	\N	1	1	\N
6	joa	2025-09-23 12:53:33.973489+00	2025-09-23 12:53:33.973489+00	\N	1	1	\N
\.


--
-- Data for Name: project_dollar_values; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.project_dollar_values (id, project_id, year, month, start_value, end_value, average_value, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	2	2025	Enero	1100.00	1200.00	1150.00	2025-09-22 10:19:51.726022+00	2025-09-22 10:19:51.726022+00	\N	\N	\N	\N
2	2	2025	Febrero	1200.00	1300.00	1250.00	2025-09-22 10:19:52.346284+00	2025-09-22 10:19:52.346284+00	\N	\N	\N	\N
3	2	2025	Marzo	1000.00	1100.00	1050.00	2025-09-22 10:19:52.841902+00	2025-09-22 10:19:52.841902+00	\N	\N	\N	\N
4	2	2025	Abril	1250.00	1300.00	1275.00	2025-09-22 10:19:53.336658+00	2025-09-22 10:19:53.336658+00	\N	\N	\N	\N
5	2	2025	Mayo	1400.00	1200.00	1300.00	2025-09-22 10:19:53.831852+00	2025-09-22 10:19:53.831852+00	\N	\N	\N	\N
6	2	2025	Junio	1300.00	1000.00	1150.00	2025-09-22 10:19:54.327397+00	2025-09-22 10:19:54.327397+00	\N	\N	\N	\N
7	2	2025	Julio	1000.00	1200.00	1100.00	2025-09-22 10:19:54.823961+00	2025-09-22 10:19:54.823961+00	\N	\N	\N	\N
8	2	2025	Agosto	1100.00	1250.00	1175.00	2025-09-22 10:19:55.319031+00	2025-09-22 10:19:55.319031+00	\N	\N	\N	\N
9	2	2025	Septiembre	1400.00	1300.00	1350.00	2025-09-22 10:19:55.813811+00	2025-09-22 10:19:55.813811+00	\N	\N	\N	\N
10	2	2025	Octubre	1200.00	1200.00	1200.00	2025-09-22 10:19:56.30882+00	2025-09-22 10:19:56.30882+00	\N	\N	\N	\N
11	2	2025	Noviembre	1000.00	1200.00	1100.00	2025-09-22 10:19:56.803908+00	2025-09-22 10:19:56.803908+00	\N	\N	\N	\N
12	2	2025	Diciembre	1200.00	1100.00	1150.00	2025-09-22 10:19:57.29948+00	2025-09-22 10:19:57.29948+00	\N	\N	\N	\N
\.


--
-- Data for Name: project_investors; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.project_investors (project_id, investor_id, percentage, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	1	60	2025-09-19 19:59:35.95604+00	2025-09-19 19:59:35.95604+00	\N	1	1	\N
1	2	40	2025-09-19 19:59:35.95604+00	2025-09-19 19:59:35.95604+00	\N	1	1	\N
2	2	50	2025-09-22 10:11:31.263518+00	2025-09-22 10:11:31.263518+00	\N	1	1	\N
2	1	30	2025-09-22 10:11:31.263518+00	2025-09-22 10:11:31.263518+00	\N	1	1	\N
2	3	20	2025-09-22 10:11:31.263518+00	2025-09-22 10:11:31.263518+00	\N	1	1	\N
3	1	30	2025-09-22 12:30:54.031079+00	2025-09-22 12:30:54.031079+00	\N	1	1	\N
3	2	20	2025-09-22 12:30:54.031079+00	2025-09-22 12:30:54.031079+00	\N	1	1	\N
3	4	40	2025-09-22 12:30:54.031079+00	2025-09-22 12:30:54.031079+00	\N	1	1	\N
3	3	10	2025-09-22 12:30:54.031079+00	2025-09-22 12:30:54.031079+00	\N	1	1	\N
4	5	50	2025-09-23 12:14:28.077148+00	2025-09-23 12:14:28.077148+00	\N	1	1	\N
4	6	50	2025-09-23 12:14:28.077148+00	2025-09-23 12:14:28.077148+00	\N	1	1	\N
5	7	50	2025-09-23 12:53:35.73617+00	2025-09-23 12:53:35.73617+00	\N	1	1	\N
5	8	50	2025-09-23 12:53:35.73617+00	2025-09-23 12:53:35.73617+00	\N	1	1	\N
\.


--
-- Data for Name: project_managers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.project_managers (project_id, manager_id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	1	2025-09-19 19:59:32.320723+00	2025-09-19 19:59:32.320723+00	\N	\N	\N	\N
1	2	2025-09-19 19:59:32.320723+00	2025-09-19 19:59:32.320723+00	\N	\N	\N	\N
2	2	2025-09-22 10:11:28.920989+00	2025-09-22 10:11:28.920989+00	\N	\N	\N	\N
3	3	2025-09-22 12:30:50.52393+00	2025-09-22 12:30:50.52393+00	\N	\N	\N	\N
3	2	2025-09-22 12:30:50.52393+00	2025-09-22 12:30:50.52393+00	\N	\N	\N	\N
4	4	2025-09-23 12:14:24.781307+00	2025-09-23 12:14:24.781307+00	\N	\N	\N	\N
5	5	2025-09-23 12:53:32.509503+00	2025-09-23 12:53:32.509503+00	\N	\N	\N	\N
5	6	2025-09-23 12:53:32.509503+00	2025-09-23 12:53:32.509503+00	\N	\N	\N	\N
\.


--
-- Data for Name: projects; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.projects (id, name, customer_id, campaign_id, admin_cost, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Oscar 1	1	1	300.000	2025-09-19 19:59:35.645858+00	2025-09-19 19:59:35.645858+00	\N	1	1	\N
2	demo	1	1	45.000	2025-09-22 10:11:31.006865+00	2025-09-22 10:11:31.006865+00	\N	1	1	\N
3	Proyecto 2025 X	2	2	1200.000	2025-09-22 12:30:53.774727+00	2025-09-22 12:30:53.774727+00	\N	1	1	\N
4	PRUEBA PABLO	3	2	50.000	2025-09-23 12:14:27.796697+00	2025-09-23 12:50:38.742791+00	\N	1	1	\N
5	proy 1	4	1	35.000	2025-09-23 12:53:35.484026+00	2025-09-23 12:53:35.484026+00	\N	1	1	\N
\.


--
-- Data for Name: providers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.providers (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	prueba	2025-09-23 20:12:00.230141+00	2025-09-23 20:12:00.230141+00	\N	\N	\N	\N
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, dirty) FROM stdin;
85	f
\.


--
-- Data for Name: stocks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.stocks (id, project_id, supply_id, investor_id, close_date, real_stock_units, initial_units, year_period, month_period, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by, units_entered, units_consumed) FROM stdin;
1	5	9	7	\N	8.000	8.000	2025	9	2025-09-23 20:11:58.949103+00	2025-09-23 20:11:58.949103+00	\N	1	1	\N	0.000	0.000
\.


--
-- Data for Name: supplies; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.supplies (id, project_id, name, price, unit_id, category_id, type_id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	2	nogales 32	200	1	1	1	2025-09-22 10:17:38.490713+00	2025-09-22 10:17:38.490713+00	\N	1	\N	\N
2	2	Fertilizantes X	15	2	8	3	2025-09-22 10:21:58.971173+00	2025-09-22 10:21:58.971173+00	\N	1	\N	\N
3	2	curasemilla x	100	2	3	2	2025-09-22 10:29:20.963558+00	2025-09-22 10:29:20.963558+00	\N	1	\N	\N
4	4	SEMILLA GARBANZIO KIARA	1	2	1	1	2025-09-23 12:20:20.126077+00	2025-09-23 12:20:20.126077+00	\N	1	\N	\N
5	4	SEMILLA TRIGO	0.5	2	1	1	2025-09-23 12:20:20.126077+00	2025-09-23 12:20:20.126077+00	\N	1	\N	\N
6	4	GLIFOSATO 66%	5	1	4	2	2025-09-23 12:20:20.126077+00	2025-09-23 12:20:20.126077+00	\N	1	\N	\N
7	4	2,4 D AMINA	5	1	4	2	2025-09-23 12:20:20.126077+00	2025-09-23 12:20:20.126077+00	\N	1	\N	\N
8	4	CURASEMILLO	1	1	3	2	2025-09-23 12:20:20.126077+00	2025-09-23 12:20:20.126077+00	\N	1	\N	\N
9	5	24d	5	1	4	2	2025-09-23 12:55:21.936886+00	2025-09-23 12:55:21.936886+00	\N	1	\N	\N
10	5	atranex	15	2	4	2	2025-09-23 12:55:21.936886+00	2025-09-23 12:55:21.936886+00	\N	1	\N	\N
11	5	coady	10	1	2	2	2025-09-23 12:55:21.936886+00	2025-09-23 12:55:21.936886+00	\N	1	\N	\N
12	5	maiz	0.7	2	1	1	2025-09-23 12:55:21.936886+00	2025-09-23 12:55:21.936886+00	\N	1	\N	\N
13	5	soja	0.5	2	1	1	2025-09-23 12:55:21.936886+00	2025-09-23 12:55:21.936886+00	\N	1	\N	\N
\.


--
-- Data for Name: supply_movements; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.supply_movements (id, stock_id, quantity, movement_type, movement_date, reference_number, is_entry, project_id, project_destination_id, supply_id, investor_id, provider_id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	1	8.000	Stock	2025-09-23 00:00:00	111	t	5	0	9	7	1	2025-09-23 20:12:00.733781+00	2025-09-23 20:12:00.733781+00	\N	1	1	\N
\.


--
-- Data for Name: types; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.types (id, name, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	Semilla	2025-09-19 19:55:52.805042+00	2025-09-19 19:55:52.805042+00	\N	\N	\N	\N
2	Agroquímicos	2025-09-19 19:55:52.805042+00	2025-09-19 19:55:52.805042+00	\N	\N	\N	\N
3	Fertilizantes	2025-09-19 19:55:52.805042+00	2025-09-19 19:55:52.805042+00	\N	\N	\N	\N
4	Labores	2025-09-19 19:55:52.805042+00	2025-09-19 19:55:52.805042+00	\N	\N	\N	\N
\.


--
-- Data for Name: user_logins; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_logins (id, user_id, login_at, ip_address, device_info, success, logout_at, session_duration) FROM stdin;
1	1	2025-09-19 17:04:05.650749+00	2600:1900:0:2d01::1	axios/1.8.1	t	\N	\N
2	1	2025-09-19 17:04:14.702064+00	2600:1900:0:2d01::1	axios/1.8.1	t	\N	\N
3	1	2025-09-19 17:05:10.039719+00	2600:1900:0:2d01::1	axios/1.8.1	t	\N	\N
4	1	2025-09-22 12:23:44.049797+00	2600:1900:0:2d01::1300	axios/1.8.1	t	\N	\N
5	1	2025-09-22 12:27:29.603604+00	2600:1900:0:2d01::1300	axios/1.8.1	t	\N	\N
6	1	2025-09-22 12:51:48.303612+00	2600:1900:0:2d01::1300	axios/1.8.1	t	\N	\N
7	1	2025-09-22 16:42:43.517668+00	2600:1900:0:2d07::3701	axios/1.8.1	t	\N	\N
8	1	2025-09-22 16:42:57.838972+00	2600:1900:0:2d07::3701	axios/1.8.1	t	\N	\N
9	1	2025-09-22 16:43:02.019696+00	2600:1900:0:2d07::3701	axios/1.8.1	t	\N	\N
10	1	2025-09-23 12:50:54.124168+00	2600:1900:0:2d00::3701	axios/1.8.1	t	\N	\N
11	1	2025-09-23 19:52:51.48683+00	2600:1900:0:2d08::3701	axios/1.8.1	t	\N	\N
12	1	2025-09-23 21:06:56.760035+00	2600:1900:0:2d07::1301	axios/1.8.1	t	\N	\N
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, created_at, updated_at, deleted_at, email, username, password, token_hash, refresh_tokens, id_rol, is_verified, active, created_by, updated_by, deleted_by) FROM stdin;
1	2025-09-19 17:03:52.118685+00	2025-09-23 21:06:58.926756+00	\N	admin@example.com	soalenadmin25	$2a$14$0TVhmHrRYOQp9a.VJB81Nu3.iCdkJIoTpzSfxTt8qarKUmTXAgi.e	QrIOJizOGpIMImz	{eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzczODUzNDQ3LCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.Oiofv3Pxgh7BfheV94gpXIqTxeaG7K3QTmh-9j9a0e4,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzczODUzNDU2LCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.U9ehl92xb0wtDZq3PPqeghNJU7pMiZt0qmkTa_o89Lo,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzczODUzNTExLCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.pEuewds2vPLZB6Qswacb1F-jtCjZmfAMK0ntV_hR2jk,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzc0MDk1ODI1LCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.jsL-kLXf_qoyg5yzHKLlLteCxQuwW8KNg4SDSa2S1UA,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzc0MDk2MDUxLCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.bgGU3OVZVcPj6TUW-WUUpKGXwB3wKUvaClJT30zx1yE,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzc0MTExMzY0LCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.RngfucrCgdLf99bfQpYtBTv3pZmt6upU6DLdPnRnM0g,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzc0MTExMzc5LCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.o8Cr4waoGmzNs4qNIWvwjmXy5CuZkTdPEQkkiXHdB-g,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzc0MTExMzgzLCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.JR_bmA87sgdmzAuKh0n9bFwQNuuDL-owZBMlOZ4Ux6M,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzc0MTgzODU2LCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.Gbym-f3OakoZT99uBFIb3vcsa7VF64_I5LJg9mZgd5c,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzc0MjA5MTczLCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.1GvY0DCU_QLKF__ozN38nIeXyYUx79ztWBgSYyCSkGY,eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6MSwiSGFzaCI6IlFySU9KaXpPR3BJTUlteiIsIktleVR5cGUiOiJyZWZyZXNoIiwiZXhwIjoxNzc0MjEzNjE4LCJpc3MiOiJhdXRoLnNlcnZpY2UifQ.tVd7yt1GnKADhbtivRAMlji5HTUI7aI6_xHeSDcreHU}	1	t	t	1	1	\N
\.


--
-- Data for Name: workorder_items; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.workorder_items (id, workorder_id, supply_id, total_used, final_dose, created_at, updated_at, deleted_at) FROM stdin;
1	1	2	300.000000	1.000000	2025-09-22 10:25:12.479898+00	2025-09-22 10:25:12.479898+00	\N
2	2	2	34.000000	0.057000	2025-09-22 23:14:07.27978+00	2025-09-22 23:14:07.27978+00	\N
3	3	3	245.000000	1.000000	2025-09-23 01:43:30.89115+00	2025-09-23 01:43:30.89115+00	\N
4	4	5	50.000000	0.500000	2025-09-23 12:21:54.196546+00	2025-09-23 12:21:54.196546+00	\N
5	4	8	50.000000	0.500000	2025-09-23 12:21:54.196546+00	2025-09-23 12:21:54.196546+00	\N
6	5	6	100.000000	1.000000	2025-09-23 12:22:37.672933+00	2025-09-23 12:22:37.672933+00	\N
7	5	7	100.000000	1.000000	2025-09-23 12:22:37.672933+00	2025-09-23 12:22:37.672933+00	\N
8	6	4	100.000000	1.000000	2025-09-23 12:23:37.806698+00	2025-09-23 12:23:37.806698+00	\N
9	6	8	100.000000	1.000000	2025-09-23 12:23:37.806698+00	2025-09-23 12:23:37.806698+00	\N
10	7	6	100.000000	1.000000	2025-09-23 12:24:22.051076+00	2025-09-23 12:24:22.051076+00	\N
11	7	7	100.000000	1.000000	2025-09-23 12:24:22.051076+00	2025-09-23 12:24:22.051076+00	\N
15	9	9	50.000000	2.000000	2025-09-23 12:57:17.677034+00	2025-09-23 12:57:17.677034+00	\N
16	9	11	25.000000	1.000000	2025-09-23 12:57:17.677034+00	2025-09-23 12:57:17.677034+00	\N
17	9	10	10.000000	0.400000	2025-09-23 12:57:17.677034+00	2025-09-23 12:57:17.677034+00	\N
18	10	13	500.000000	10.000000	2025-09-23 12:57:51.990922+00	2025-09-23 12:57:51.990922+00	\N
20	11	12	500.000000	20.000000	2025-09-23 12:59:12.163202+00	2025-09-23 12:59:12.163202+00	\N
21	15	6	100.000000	1.000000	2025-09-23 13:19:00.755238+00	2025-09-23 13:19:00.755238+00	\N
22	8	9	50.000000	1.000000	2025-09-23 13:35:44.87978+00	2025-09-23 13:35:44.87978+00	\N
23	8	11	100.000000	2.000000	2025-09-23 13:35:44.87978+00	2025-09-23 13:35:44.87978+00	\N
24	8	10	25.000000	0.500000	2025-09-23 13:35:44.87978+00	2025-09-23 13:35:44.87978+00	\N
25	16	6	100.000000	1.000000	2025-09-23 13:40:35.625585+00	2025-09-23 13:40:35.625585+00	\N
\.


--
-- Data for Name: workorders; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.workorders (id, number, project_id, field_id, lot_id, crop_id, labor_id, contractor, observations, date, investor_id, effective_area, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by) FROM stdin;
1	1	2	2	2	12	2	Carlos		2025-09-22	3	300.000000	2025-09-22 10:25:12.54529+00	2025-09-22 10:25:12.54529+00	\N	1	1	\N
2	02	2	2	2	12	2	Carlos	ñañaña	2025-09-23	1	600.000000	2025-09-22 23:14:07.344621+00	2025-09-22 23:14:07.344621+00	\N	1	1	\N
3	4	2	2	2	12	3	Ramon	wsdsd	2025-09-23	1	245.000000	2025-09-23 01:43:30.956447+00	2025-09-23 01:43:30.956447+00	\N	1	1	\N
4	1	4	4	5	11	5	SANCHEZ		2025-05-01	6	100.000000	2025-09-23 12:21:54.273234+00	2025-09-23 12:21:54.273234+00	\N	1	1	\N
5	2	4	4	5	11	7	SANCHEZ		2025-05-05	5	100.000000	2025-09-23 12:22:37.749863+00	2025-09-23 12:22:37.749863+00	\N	1	1	\N
6	3	4	4	6	14	6	MUÑO		2025-05-05	5	100.000000	2025-09-23 12:23:37.868553+00	2025-09-23 12:23:37.868553+00	\N	1	1	\N
7	4	4	4	6	14	8	MUÑO		2025-05-05	6	100.000000	2025-09-23 12:24:22.113307+00	2025-09-23 12:24:22.113307+00	\N	1	1	\N
9	2	5	5	9	2	9	pepe		2025-01-01	8	25.000000	2025-09-23 12:57:17.738975+00	2025-09-23 12:57:17.738975+00	\N	1	1	\N
10	3	5	5	8	1	11	s		2025-02-02	7	50.000000	2025-09-23 12:57:52.06773+00	2025-09-23 12:57:52.06773+00	\N	1	1	\N
11	4	5	5	9	2	11	s		2025-02-01	7	25.000000	2025-09-23 12:58:19.837038+00	2025-09-23 12:59:12.662101+00	\N	1	1	\N
12	5	5	5	8	1	10	c		2025-09-08	7	50.000000	2025-09-23 12:59:34.810343+00	2025-09-23 12:59:34.810343+00	\N	1	1	\N
13	6	5	5	9	2	10	c		2025-09-08	8	25.000000	2025-09-23 12:59:58.005138+00	2025-09-23 12:59:58.005138+00	\N	1	1	\N
14	5	4	4	5	11	12	EUGE		2025-09-30	6	100.000000	2025-09-23 13:04:30.971422+00	2025-09-23 13:04:30.971422+00	\N	1	1	\N
15	6	4	4	7	2	8	MUÑO		2025-09-09	5	100.000000	2025-09-23 13:19:00.817428+00	2025-09-23 13:19:00.817428+00	2025-09-23 13:30:03.279906+00	1	1	\N
8	1	5	5	8	1	9	pepe		2025-01-01	7	50.000000	2025-09-23 12:56:40.012804+00	2025-09-23 13:35:45.882024+00	\N	1	1	\N
16	6	4	4	7	2	8	MUÑO		2025-09-30	5	100.000000	2025-09-23 13:40:35.689316+00	2025-09-23 13:40:35.689316+00	\N	1	1	\N
\.


--
-- Name: app_parameters_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.app_parameters_id_seq', 6, true);


--
-- Name: campaigns_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.campaigns_id_seq', 3, true);


--
-- Name: categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.categories_id_seq', 13, true);


--
-- Name: crop_commercializations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.crop_commercializations_id_seq', 6, true);


--
-- Name: crops_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.crops_id_seq', 15, true);


--
-- Name: customers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.customers_id_seq', 4, true);


--
-- Name: engineering_principles_documentation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.engineering_principles_documentation_id_seq', 4, true);


--
-- Name: fields_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.fields_id_seq', 5, true);


--
-- Name: fx_rates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.fx_rates_id_seq', 1, true);


--
-- Name: investors_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.investors_id_seq', 8, true);


--
-- Name: invoices_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.invoices_id_seq', 1, true);


--
-- Name: labor_categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.labor_categories_id_seq', 13, true);


--
-- Name: labor_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.labor_types_id_seq', 4, true);


--
-- Name: labors_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.labors_id_seq', 12, true);


--
-- Name: lease_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.lease_types_id_seq', 4, true);


--
-- Name: lot_dates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.lot_dates_id_seq', 15, true);


--
-- Name: lots_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.lots_id_seq', 9, true);


--
-- Name: managers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.managers_id_seq', 6, true);


--
-- Name: project_dollar_values_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.project_dollar_values_id_seq', 12, true);


--
-- Name: projects_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.projects_id_seq', 5, true);


--
-- Name: providers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.providers_id_seq', 1, true);


--
-- Name: stocks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.stocks_id_seq', 1, true);


--
-- Name: supplies_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.supplies_id_seq', 13, true);


--
-- Name: supply_movements_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.supply_movements_id_seq', 1, true);


--
-- Name: types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.types_id_seq', 4, true);


--
-- Name: user_logins_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.user_logins_id_seq', 12, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.users_id_seq', 1, true);


--
-- Name: workorder_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.workorder_items_id_seq', 25, true);


--
-- Name: workorders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.workorders_id_seq', 16, true);


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
-- Name: users uni_users_username; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT uni_users_username UNIQUE (username);


--
-- Name: lot_dates unique_lot_dates; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lot_dates
    ADD CONSTRAINT unique_lot_dates UNIQUE (lot_id, sequence);


--
-- Name: user_logins user_logins_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_logins
    ADD CONSTRAINT user_logins_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


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
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


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

\unrestrict kArU3igAZCTaMLnvESuwVen1f6gQqevemM4DRAmiCFMErh8ugxPtcKj7MFe4LE4

