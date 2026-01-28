-- ========================================
-- MIGRATION 000003 CORE TRIGGERS FUNCTIONS (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TRIGGER IF EXISTS set_timestamp ON public.users;
DROP FUNCTION IF EXISTS public.update_timestamp();
DROP FUNCTION IF EXISTS public.get_default_fx_rate();
DROP FUNCTION IF EXISTS public.get_campaign_closure_days();
DROP FUNCTION IF EXISTS public.get_iva_percentage();
DROP FUNCTION IF EXISTS public.get_app_parameter_integer(varchar);
DROP FUNCTION IF EXISTS public.get_app_parameter_decimal(varchar);
DROP FUNCTION IF EXISTS public.get_app_parameter(varchar);

COMMIT;
