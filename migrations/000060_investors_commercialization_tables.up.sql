-- ========================================
-- MIGRATION INVESTORS_COMMERCIALIZATION TABLES (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

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

CREATE SEQUENCE public.investors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.investors_id_seq OWNED BY public.investors.id;

ALTER TABLE ONLY public.investors ALTER COLUMN id SET DEFAULT nextval('public.investors_id_seq'::regclass);

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

CREATE SEQUENCE public.crop_commercializations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.crop_commercializations_id_seq OWNED BY public.crop_commercializations.id;

ALTER TABLE ONLY public.crop_commercializations ALTER COLUMN id SET DEFAULT nextval('public.crop_commercializations_id_seq'::regclass);


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

CREATE SEQUENCE public.project_dollar_values_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.project_dollar_values_id_seq OWNED BY public.project_dollar_values.id;

ALTER TABLE ONLY public.project_dollar_values ALTER COLUMN id SET DEFAULT nextval('public.project_dollar_values_id_seq'::regclass);

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

CREATE SEQUENCE public.invoices_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.invoices_id_seq OWNED BY public.invoices.id;

ALTER TABLE ONLY public.invoices ALTER COLUMN id SET DEFAULT nextval('public.invoices_id_seq'::regclass);

COMMIT;
