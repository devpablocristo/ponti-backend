-- ========================================
-- MIGRATION 000050 WORKORDERS LABORS TABLES (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

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

CREATE SEQUENCE public.labor_types_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.labor_types_id_seq OWNED BY public.labor_types.id;
ALTER TABLE ONLY public.labor_types ALTER COLUMN id SET DEFAULT nextval('public.labor_types_id_seq'::regclass);

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

CREATE SEQUENCE public.labor_categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.labor_categories_id_seq OWNED BY public.labor_categories.id;
ALTER TABLE ONLY public.labor_categories ALTER COLUMN id SET DEFAULT nextval('public.labor_categories_id_seq'::regclass);

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

CREATE SEQUENCE public.labors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.labors_id_seq OWNED BY public.labors.id;
ALTER TABLE ONLY public.labors ALTER COLUMN id SET DEFAULT nextval('public.labors_id_seq'::regclass);

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

CREATE SEQUENCE public.workorders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.workorders_id_seq OWNED BY public.workorders.id;
ALTER TABLE ONLY public.workorders ALTER COLUMN id SET DEFAULT nextval('public.workorders_id_seq'::regclass);

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

CREATE SEQUENCE public.workorder_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.workorder_items_id_seq OWNED BY public.workorder_items.id;
ALTER TABLE ONLY public.workorder_items ALTER COLUMN id SET DEFAULT nextval('public.workorder_items_id_seq'::regclass);

COMMIT;
