BEGIN;

-- Pilar 1 — cierre del gap NOT NULL en tenant_id (Modelo 1).
-- Estado verificado: 0 filas con tenant_id NULL en las 11 tablas per-tenant (el backfill
-- de migr 232/235/236 ya las dejó en 'default'). Acá: backfill defensivo (no-op), SET
-- DEFAULT al tenant 'default' (para que todo INSERT que omita tenant_id quede non-null sin
-- tocar el código — los modelos GORM no tienen campo TenantID, así que GORM omite la
-- columna y aplica el DEFAULT), y SET NOT NULL. Resuelve el uuid del tenant 'default'
-- dinámicamente (portable entre entornos).
--
-- NOTA Modelo 2: el estampado in-INSERT del tenant REAL por request (no el 'default') es un
-- paso aparte de la activación de Modelo 2 (≥2 tenants); hoy, con Modelo 1, todo es 'default'
-- y el DEFAULT + el dual-write existente bastan.

DO $$
DECLARE
	def    uuid;
	t      text;
	tables text[] := ARRAY[
		'customers','campaigns','projects',
		'managers','investors','providers',
		'crops','types','lease_types','business_parameters','categories'
	];
BEGIN
	SELECT id INTO def FROM public.auth_tenants WHERE name = 'default' LIMIT 1;
	IF def IS NULL THEN
		RAISE EXCEPTION 'no existe el tenant default; abortando NOT NULL de tenant_id';
	END IF;

	FOREACH t IN ARRAY tables LOOP
		EXECUTE format('UPDATE public.%I SET tenant_id = %L WHERE tenant_id IS NULL', t, def);
		EXECUTE format('ALTER TABLE public.%I ALTER COLUMN tenant_id SET DEFAULT %L', t, def);
		EXECUTE format('ALTER TABLE public.%I ALTER COLUMN tenant_id SET NOT NULL', t);
	END LOOP;
END $$;

COMMIT;
