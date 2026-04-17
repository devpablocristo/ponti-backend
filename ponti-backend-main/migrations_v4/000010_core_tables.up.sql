-- ========================================
-- MIGRATION 000010 CORE TABLES (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

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

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;
ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

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

CREATE SEQUENCE public.business_parameters_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.business_parameters_id_seq OWNED BY public.business_parameters.id;
ALTER TABLE ONLY public.business_parameters ALTER COLUMN id SET DEFAULT nextval('public.business_parameters_id_seq'::regclass);

CREATE TABLE public.fx_rates (
    id integer NOT NULL,
    currency_pair character varying(10) NOT NULL,
    rate numeric(10,4) NOT NULL,
    effective_date date NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE SEQUENCE public.fx_rates_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.fx_rates_id_seq OWNED BY public.fx_rates.id;
ALTER TABLE ONLY public.fx_rates ALTER COLUMN id SET DEFAULT nextval('public.fx_rates_id_seq'::regclass);

COMMIT;
