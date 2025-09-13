-- ========================================
-- MIGRATION 000001: CREATE V3 SCHEMA (UP)
-- ========================================
-- 
-- Purpose: Create all tables and basic functions for v3 system
-- Date: 2025-09-13
-- Author: System
-- 
-- Note: Code in English, comments in Spanish.

BEGIN;

-- Crear tipos ENUM
CREATE TYPE public.movement_type AS ENUM (
    'Stock',
    'Movimiento interno',
    'Remito oficial'
);

-- Crear tablas principales
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

CREATE TABLE public.fx_rates (
    id integer NOT NULL,
    currency_pair character varying(10) NOT NULL,
    rate numeric(10,4) NOT NULL,
    effective_date date NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);

-- Crear secuencias
CREATE SEQUENCE public.users_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.customers_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.campaigns_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.projects_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.lease_types_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.fields_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.crops_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.lots_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.types_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.categories_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.labor_types_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.labor_categories_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.labors_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.supplies_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.investors_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.workorders_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.workorder_items_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.crop_commercializations_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.managers_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.invoices_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.stocks_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.supply_movements_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.providers_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.project_dollar_values_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.app_parameters_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
CREATE SEQUENCE public.fx_rates_id_seq AS integer START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

-- Asignar secuencias a columnas
ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);
ALTER TABLE ONLY public.customers ALTER COLUMN id SET DEFAULT nextval('public.customers_id_seq'::regclass);
ALTER TABLE ONLY public.campaigns ALTER COLUMN id SET DEFAULT nextval('public.campaigns_id_seq'::regclass);
ALTER TABLE ONLY public.projects ALTER COLUMN id SET DEFAULT nextval('public.projects_id_seq'::regclass);
ALTER TABLE ONLY public.lease_types ALTER COLUMN id SET DEFAULT nextval('public.lease_types_id_seq'::regclass);
ALTER TABLE ONLY public.fields ALTER COLUMN id SET DEFAULT nextval('public.fields_id_seq'::regclass);
ALTER TABLE ONLY public.crops ALTER COLUMN id SET DEFAULT nextval('public.crops_id_seq'::regclass);
ALTER TABLE ONLY public.lots ALTER COLUMN id SET DEFAULT nextval('public.lots_id_seq'::regclass);
ALTER TABLE ONLY public.types ALTER COLUMN id SET DEFAULT nextval('public.types_id_seq'::regclass);
ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);
ALTER TABLE ONLY public.labor_types ALTER COLUMN id SET DEFAULT nextval('public.labor_types_id_seq'::regclass);
ALTER TABLE ONLY public.labor_categories ALTER COLUMN id SET DEFAULT nextval('public.labor_categories_id_seq'::regclass);
ALTER TABLE ONLY public.labors ALTER COLUMN id SET DEFAULT nextval('public.labors_id_seq'::regclass);
ALTER TABLE ONLY public.supplies ALTER COLUMN id SET DEFAULT nextval('public.supplies_id_seq'::regclass);
ALTER TABLE ONLY public.investors ALTER COLUMN id SET DEFAULT nextval('public.investors_id_seq'::regclass);
ALTER TABLE ONLY public.workorders ALTER COLUMN id SET DEFAULT nextval('public.workorders_id_seq'::regclass);
ALTER TABLE ONLY public.workorder_items ALTER COLUMN id SET DEFAULT nextval('public.workorder_items_id_seq'::regclass);
ALTER TABLE ONLY public.crop_commercializations ALTER COLUMN id SET DEFAULT nextval('public.crop_commercializations_id_seq'::regclass);
ALTER TABLE ONLY public.managers ALTER COLUMN id SET DEFAULT nextval('public.managers_id_seq'::regclass);
ALTER TABLE ONLY public.invoices ALTER COLUMN id SET DEFAULT nextval('public.invoices_id_seq'::regclass);
ALTER TABLE ONLY public.stocks ALTER COLUMN id SET DEFAULT nextval('public.stocks_id_seq'::regclass);
ALTER TABLE ONLY public.supply_movements ALTER COLUMN id SET DEFAULT nextval('public.supply_movements_id_seq'::regclass);
ALTER TABLE ONLY public.providers ALTER COLUMN id SET DEFAULT nextval('public.providers_id_seq'::regclass);
ALTER TABLE ONLY public.project_dollar_values ALTER COLUMN id SET DEFAULT nextval('public.project_dollar_values_id_seq'::regclass);
ALTER TABLE ONLY public.app_parameters ALTER COLUMN id SET DEFAULT nextval('public.app_parameters_id_seq'::regclass);
ALTER TABLE ONLY public.fx_rates ALTER COLUMN id SET DEFAULT nextval('public.fx_rates_id_seq'::regclass);

-- Crear funciones básicas necesarias
CREATE FUNCTION public.update_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$;

CREATE FUNCTION public.get_app_parameter(p_key character varying) RETURNS character varying
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value FROM app_parameters WHERE key = p_key);
END;
$$;

CREATE FUNCTION public.get_app_parameter_decimal(p_key character varying) RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value::DECIMAL FROM app_parameters WHERE key = p_key);
END;
$$;

CREATE FUNCTION public.get_app_parameter_integer(p_key character varying) RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN (SELECT value::INTEGER FROM app_parameters WHERE key = p_key);
END;
$$;

CREATE FUNCTION public.get_campaign_closure_days() RETURNS integer
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN get_app_parameter_integer('campaign_closure_days');
END;
$$;

CREATE FUNCTION public.get_default_fx_rate() RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN get_app_parameter_decimal('default_fx_rate');
END;
$$;

CREATE FUNCTION public.get_iva_percentage() RETURNS numeric
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
  RETURN get_app_parameter_decimal('iva_percentage');
END;
$$;

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

-- Crear PKs y constraints
ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users ADD CONSTRAINT users_username_key UNIQUE (username);
ALTER TABLE ONLY public.customers ADD CONSTRAINT customers_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.customers ADD CONSTRAINT customers_name_key UNIQUE (name);
ALTER TABLE ONLY public.campaigns ADD CONSTRAINT campaigns_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.campaigns ADD CONSTRAINT campaigns_name_key UNIQUE (name);
ALTER TABLE ONLY public.projects ADD CONSTRAINT projects_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.lease_types ADD CONSTRAINT lease_types_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.lease_types ADD CONSTRAINT lease_types_name_key UNIQUE (name);
ALTER TABLE ONLY public.fields ADD CONSTRAINT fields_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.crops ADD CONSTRAINT crops_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.crops ADD CONSTRAINT crops_name_key UNIQUE (name);
ALTER TABLE ONLY public.lots ADD CONSTRAINT lots_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.types ADD CONSTRAINT types_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.types ADD CONSTRAINT types_name_key UNIQUE (name);
ALTER TABLE ONLY public.categories ADD CONSTRAINT categories_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.labor_types ADD CONSTRAINT labor_types_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.labor_types ADD CONSTRAINT labor_types_name_key UNIQUE (name);
ALTER TABLE ONLY public.labor_categories ADD CONSTRAINT labor_categories_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.labors ADD CONSTRAINT labors_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.supplies ADD CONSTRAINT supplies_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.investors ADD CONSTRAINT investors_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.investors ADD CONSTRAINT investors_name_key UNIQUE (name);
ALTER TABLE ONLY public.project_investors ADD CONSTRAINT project_investors_pkey PRIMARY KEY (project_id, investor_id);
ALTER TABLE ONLY public.workorders ADD CONSTRAINT workorders_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.workorder_items ADD CONSTRAINT workorder_items_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.crop_commercializations ADD CONSTRAINT crop_commercializations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.managers ADD CONSTRAINT managers_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.managers ADD CONSTRAINT managers_name_key UNIQUE (name);
ALTER TABLE ONLY public.project_managers ADD CONSTRAINT project_managers_pkey PRIMARY KEY (project_id, manager_id);
ALTER TABLE ONLY public.invoices ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.invoices ADD CONSTRAINT invoices_work_order_id_key UNIQUE (work_order_id);
ALTER TABLE ONLY public.stocks ADD CONSTRAINT stocks_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.supply_movements ADD CONSTRAINT supply_movements_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.providers ADD CONSTRAINT providers_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.providers ADD CONSTRAINT providers_name_key UNIQUE (name);
ALTER TABLE ONLY public.project_dollar_values ADD CONSTRAINT project_dollar_values_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.project_dollar_values ADD CONSTRAINT project_dollar_values_project_id_year_month_key UNIQUE (project_id, year, month);
ALTER TABLE ONLY public.app_parameters ADD CONSTRAINT app_parameters_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.app_parameters ADD CONSTRAINT app_parameters_key_key UNIQUE (key);
ALTER TABLE ONLY public.fx_rates ADD CONSTRAINT fx_rates_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.schema_migrations ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);

-- Asignar ownership de secuencias
ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;
ALTER SEQUENCE public.customers_id_seq OWNED BY public.customers.id;
ALTER SEQUENCE public.campaigns_id_seq OWNED BY public.campaigns.id;
ALTER SEQUENCE public.projects_id_seq OWNED BY public.projects.id;
ALTER SEQUENCE public.lease_types_id_seq OWNED BY public.lease_types.id;
ALTER SEQUENCE public.fields_id_seq OWNED BY public.fields.id;
ALTER SEQUENCE public.crops_id_seq OWNED BY public.crops.id;
ALTER SEQUENCE public.lots_id_seq OWNED BY public.lots.id;
ALTER SEQUENCE public.types_id_seq OWNED BY public.types.id;
ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;
ALTER SEQUENCE public.labor_types_id_seq OWNED BY public.labor_types.id;
ALTER SEQUENCE public.labor_categories_id_seq OWNED BY public.labor_categories.id;
ALTER SEQUENCE public.labors_id_seq OWNED BY public.labors.id;
ALTER SEQUENCE public.supplies_id_seq OWNED BY public.supplies.id;
ALTER SEQUENCE public.investors_id_seq OWNED BY public.investors.id;
ALTER SEQUENCE public.workorders_id_seq OWNED BY public.workorders.id;
ALTER SEQUENCE public.workorder_items_id_seq OWNED BY public.workorder_items.id;
ALTER SEQUENCE public.crop_commercializations_id_seq OWNED BY public.crop_commercializations.id;
ALTER SEQUENCE public.managers_id_seq OWNED BY public.managers.id;
ALTER SEQUENCE public.invoices_id_seq OWNED BY public.invoices.id;
ALTER SEQUENCE public.stocks_id_seq OWNED BY public.stocks.id;
ALTER SEQUENCE public.supply_movements_id_seq OWNED BY public.supply_movements.id;
ALTER SEQUENCE public.providers_id_seq OWNED BY public.providers.id;
ALTER SEQUENCE public.project_dollar_values_id_seq OWNED BY public.project_dollar_values.id;
ALTER SEQUENCE public.app_parameters_id_seq OWNED BY public.app_parameters.id;
ALTER SEQUENCE public.fx_rates_id_seq OWNED BY public.fx_rates.id;

-- Crear FKs
ALTER TABLE ONLY public.categories ADD CONSTRAINT categories_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.types(id);
ALTER TABLE ONLY public.fields ADD CONSTRAINT fk_fields_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.fields ADD CONSTRAINT fk_lease_type FOREIGN KEY (lease_type_id) REFERENCES public.lease_types(id);
ALTER TABLE ONLY public.lots ADD CONSTRAINT fk_lots_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.lots ADD CONSTRAINT fk_lots_current_crop FOREIGN KEY (current_crop_id) REFERENCES public.crops(id);
ALTER TABLE ONLY public.lots ADD CONSTRAINT fk_lots_previous_crop FOREIGN KEY (previous_crop_id) REFERENCES public.crops(id);
ALTER TABLE ONLY public.labors ADD CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.labors ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.supplies ADD CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.supplies ADD CONSTRAINT fk_type FOREIGN KEY (type_id) REFERENCES public.types(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.project_investors ADD CONSTRAINT fk_project_investors_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.project_investors ADD CONSTRAINT fk_project_investors_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id);
ALTER TABLE ONLY public.projects ADD CONSTRAINT fk_projects_customer FOREIGN KEY (customer_id) REFERENCES public.customers(id);
ALTER TABLE ONLY public.projects ADD CONSTRAINT fk_projects_campaign FOREIGN KEY (campaign_id) REFERENCES public.campaigns(id);
ALTER TABLE ONLY public.workorders ADD CONSTRAINT workorders_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorders ADD CONSTRAINT workorders_field_id_fkey FOREIGN KEY (field_id) REFERENCES public.fields(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorders ADD CONSTRAINT workorders_lot_id_fkey FOREIGN KEY (lot_id) REFERENCES public.lots(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorders ADD CONSTRAINT workorders_crop_id_fkey FOREIGN KEY (crop_id) REFERENCES public.crops(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorders ADD CONSTRAINT workorders_labor_id_fkey FOREIGN KEY (labor_id) REFERENCES public.labors(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.workorder_items ADD CONSTRAINT workorder_items_workorder_id_fkey FOREIGN KEY (workorder_id) REFERENCES public.workorders(id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE ONLY public.workorder_items ADD CONSTRAINT workorder_items_supply_id_fkey FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.crop_commercializations ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id);
ALTER TABLE ONLY public.crop_commercializations ADD CONSTRAINT fk_crop FOREIGN KEY (crop_id) REFERENCES public.crops(id);
ALTER TABLE ONLY public.project_managers ADD CONSTRAINT fk_project_managers_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.project_managers ADD CONSTRAINT fk_project_managers_manager FOREIGN KEY (manager_id) REFERENCES public.managers(id);
ALTER TABLE ONLY public.invoices ADD CONSTRAINT fk_invoices_work_order FOREIGN KEY (work_order_id) REFERENCES public.workorders(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.stocks ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.stocks ADD CONSTRAINT fk_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.stocks ADD CONSTRAINT fk_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.supply_movements ADD CONSTRAINT fk_supply FOREIGN KEY (supply_id) REFERENCES public.supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.supply_movements ADD CONSTRAINT fk_investor FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.supply_movements ADD CONSTRAINT fk_provider FOREIGN KEY (provider_id) REFERENCES public.providers(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.project_dollar_values ADD CONSTRAINT fk_project FOREIGN KEY (project_id) REFERENCES public.projects(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE ONLY public.labor_categories ADD CONSTRAINT labor_categories_type_id_fkey FOREIGN KEY (type_id) REFERENCES public.labor_types(id);

-- Crear trigger
CREATE TRIGGER set_timestamp BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_timestamp();

COMMIT;
