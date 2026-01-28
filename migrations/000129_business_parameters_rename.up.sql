-- ============================================================
-- MIGRATION 000129 BUSINESS PARAMETERS RENAME (UP)
-- ============================================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Renombrar tabla y secuencia (si vienen del nombre anterior)
DO $$
BEGIN
  IF to_regclass('public.app_parameters') IS NOT NULL THEN
    ALTER TABLE public.app_parameters RENAME TO business_parameters;
  END IF;
END $$;

DO $$
BEGIN
  IF to_regclass('public.app_parameters_id_seq') IS NOT NULL THEN
    ALTER SEQUENCE public.app_parameters_id_seq RENAME TO business_parameters_id_seq;
  END IF;
END $$;

ALTER SEQUENCE IF EXISTS public.business_parameters_id_seq OWNED BY public.business_parameters.id;
ALTER TABLE IF EXISTS public.business_parameters
  ALTER COLUMN id SET DEFAULT nextval('public.business_parameters_id_seq'::regclass);

-- Renombrar constraints e índices si existen con nombre anterior
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'app_parameters_pkey'
      AND conrelid = 'public.business_parameters'::regclass
  ) THEN
    ALTER TABLE public.business_parameters
      RENAME CONSTRAINT app_parameters_pkey TO business_parameters_pkey;
  END IF;
END $$;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'app_parameters_key_key'
      AND conrelid = 'public.business_parameters'::regclass
  ) THEN
    ALTER TABLE public.business_parameters
      RENAME CONSTRAINT app_parameters_key_key TO business_parameters_key_key;
  END IF;
END $$;

DO $$
BEGIN
  IF to_regclass('public.idx_app_parameters_key') IS NOT NULL THEN
    ALTER INDEX public.idx_app_parameters_key RENAME TO idx_business_parameters_key;
  END IF;
END $$;

DO $$
BEGIN
  IF to_regclass('public.idx_app_parameters_category') IS NOT NULL THEN
    ALTER INDEX public.idx_app_parameters_category RENAME TO idx_business_parameters_category;
  END IF;
END $$;

-- Reemplazar funciones de parámetros
DROP FUNCTION IF EXISTS public.get_app_parameter(varchar);
DROP FUNCTION IF EXISTS public.get_app_parameter_decimal(varchar);
DROP FUNCTION IF EXISTS public.get_app_parameter_integer(varchar);

CREATE OR REPLACE FUNCTION public.get_business_parameter(p_key varchar)
RETURNS varchar AS $$
BEGIN
  RETURN (SELECT value FROM public.business_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_business_parameter_decimal(p_key varchar)
RETURNS decimal AS $$
BEGIN
  RETURN (SELECT value::decimal FROM public.business_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_business_parameter_integer(p_key varchar)
RETURNS integer AS $$
BEGIN
  RETURN (SELECT value::integer FROM public.business_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Actualizar funciones de negocio para usar business_parameters
CREATE OR REPLACE FUNCTION public.get_iva_percentage()
RETURNS decimal AS $$
BEGIN
  RETURN public.get_business_parameter_decimal('iva_percentage');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_campaign_closure_days()
RETURNS integer AS $$
BEGIN
  RETURN public.get_business_parameter_integer('campaign_closure_days');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_default_fx_rate()
RETURNS decimal AS $$
BEGIN
  RETURN public.get_business_parameter_decimal('default_fx_rate');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

COMMIT;
