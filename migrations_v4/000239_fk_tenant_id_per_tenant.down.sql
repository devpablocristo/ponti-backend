BEGIN;

ALTER TABLE public.customers           DROP CONSTRAINT IF EXISTS fk_customers_tenant;
ALTER TABLE public.campaigns           DROP CONSTRAINT IF EXISTS fk_campaigns_tenant;
ALTER TABLE public.projects            DROP CONSTRAINT IF EXISTS fk_projects_tenant;
ALTER TABLE public.managers            DROP CONSTRAINT IF EXISTS fk_managers_tenant;
ALTER TABLE public.investors           DROP CONSTRAINT IF EXISTS fk_investors_tenant;
ALTER TABLE public.providers           DROP CONSTRAINT IF EXISTS fk_providers_tenant;
ALTER TABLE public.crops               DROP CONSTRAINT IF EXISTS fk_crops_tenant;
ALTER TABLE public.categories          DROP CONSTRAINT IF EXISTS fk_categories_tenant;
ALTER TABLE public.types               DROP CONSTRAINT IF EXISTS fk_types_tenant;
ALTER TABLE public.lease_types         DROP CONSTRAINT IF EXISTS fk_lease_types_tenant;
ALTER TABLE public.business_parameters DROP CONSTRAINT IF EXISTS fk_business_parameters_tenant;

COMMIT;
