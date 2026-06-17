BEGIN;

DO $$
DECLARE
	t      text;
	tables text[] := ARRAY[
		'customers','campaigns','projects',
		'managers','investors','providers',
		'crops','types','lease_types','business_parameters','categories'
	];
BEGIN
	FOREACH t IN ARRAY tables LOOP
		EXECUTE format('ALTER TABLE public.%I ALTER COLUMN tenant_id DROP NOT NULL', t);
		EXECUTE format('ALTER TABLE public.%I ALTER COLUMN tenant_id DROP DEFAULT', t);
	END LOOP;
END $$;

COMMIT;
