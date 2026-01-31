-- ========================================
-- MIGRATION 000040 CROPS TABLES (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

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

CREATE SEQUENCE public.crops_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.crops_id_seq OWNED BY public.crops.id;
ALTER TABLE ONLY public.crops ALTER COLUMN id SET DEFAULT nextval('public.crops_id_seq'::regclass);

COMMIT;
