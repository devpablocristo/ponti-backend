-- ========================================
-- MIGRATION 000015 CORE FUNCTIONS TRIGGERS (DOWN)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

DROP TRIGGER IF EXISTS trg_users_set_timestamp ON public.users;
DROP FUNCTION IF EXISTS public.update_timestamp();
DROP FUNCTION IF EXISTS public.get_default_fx_rate();
DROP FUNCTION IF EXISTS public.get_campaign_closure_days();
DROP FUNCTION IF EXISTS public.get_iva_percentage();
DROP FUNCTION IF EXISTS public.get_business_parameter_integer(varchar);
DROP FUNCTION IF EXISTS public.get_business_parameter_decimal(varchar);
DROP FUNCTION IF EXISTS public.get_business_parameter(varchar);

COMMIT;
