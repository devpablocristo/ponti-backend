-- ========================================
-- MIGRATION FIELDS_LOTS TABLES (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

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

CREATE SEQUENCE public.lease_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.lease_types_id_seq OWNED BY public.lease_types.id;

ALTER TABLE ONLY public.lease_types ALTER COLUMN id SET DEFAULT nextval('public.lease_types_id_seq'::regclass);

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

CREATE SEQUENCE public.fields_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.fields_id_seq OWNED BY public.fields.id;

ALTER TABLE ONLY public.fields ALTER COLUMN id SET DEFAULT nextval('public.fields_id_seq'::regclass);

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


CREATE SEQUENCE public.lots_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.lots_id_seq OWNED BY public.lots.id;

ALTER TABLE ONLY public.lots ALTER COLUMN id SET DEFAULT nextval('public.lots_id_seq'::regclass);

CREATE SEQUENCE public.lot_dates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.lot_dates_id_seq OWNED BY public.lot_dates.id;

ALTER TABLE ONLY public.lot_dates ALTER COLUMN id SET DEFAULT nextval('public.lot_dates_id_seq'::regclass);


COMMIT;
