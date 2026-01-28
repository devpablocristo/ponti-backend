-- ========================================
-- MIGRATION 000003 CORE TRIGGERS FUNCTIONS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Funciones de parámetros de aplicación
CREATE OR REPLACE FUNCTION public.get_app_parameter(p_key varchar)
RETURNS varchar AS $$
BEGIN
  RETURN (SELECT value FROM public.app_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_app_parameter_decimal(p_key varchar)
RETURNS decimal AS $$
BEGIN
  RETURN (SELECT value::decimal FROM public.app_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_app_parameter_integer(p_key varchar)
RETURNS integer AS $$
BEGIN
  RETURN (SELECT value::integer FROM public.app_parameters WHERE key = p_key);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_iva_percentage()
RETURNS decimal AS $$
BEGIN
  RETURN public.get_app_parameter_decimal('iva_percentage');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_campaign_closure_days()
RETURNS integer AS $$
BEGIN
  RETURN public.get_app_parameter_integer('campaign_closure_days');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.get_default_fx_rate()
RETURNS decimal AS $$
BEGIN
  RETURN public.get_app_parameter_decimal('default_fx_rate');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

CREATE OR REPLACE FUNCTION public.update_timestamp()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON public.users
FOR EACH ROW
EXECUTE FUNCTION public.update_timestamp();

COMMIT;
