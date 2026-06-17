BEGIN;

-- T3 hardening (Pilar 1): FK tenant_id -> auth_tenants(id) ON DELETE RESTRICT en TODAS
-- las entidades per-tenant. Patrón transition-safe NOT VALID + VALIDATE (lock corto).
-- IMPORTANTE: el FK PERMITE NULL -> NO rompe los inserts con flag OFF (que todavía no
-- estampan tenant_id; el estampado es post-insert flag-gated). El NOT NULL se posterga
-- a un paso aparte que requiere estampado en INSERT siempre-on (ver plan, T3).
-- ON DELETE RESTRICT: no se puede hard-deletar un auth_tenant con datos (integridad).
-- Solo afecta DELETE FROM auth_tenants (HardDeleteTenant); los cascade de proyecto/hijas
-- NO tocan auth_tenants, así que quedan intactos.
-- fx_rates EXCLUIDO (referencia global, sin tenant_id).

ALTER TABLE public.customers           ADD CONSTRAINT fk_customers_tenant           FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.campaigns           ADD CONSTRAINT fk_campaigns_tenant           FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.projects            ADD CONSTRAINT fk_projects_tenant            FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.managers            ADD CONSTRAINT fk_managers_tenant            FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.investors           ADD CONSTRAINT fk_investors_tenant           FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.providers           ADD CONSTRAINT fk_providers_tenant           FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.crops               ADD CONSTRAINT fk_crops_tenant               FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.categories          ADD CONSTRAINT fk_categories_tenant          FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.types               ADD CONSTRAINT fk_types_tenant               FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.lease_types         ADD CONSTRAINT fk_lease_types_tenant         FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.business_parameters ADD CONSTRAINT fk_business_parameters_tenant FOREIGN KEY (tenant_id) REFERENCES public.auth_tenants(id) ON DELETE RESTRICT NOT VALID;

ALTER TABLE public.customers           VALIDATE CONSTRAINT fk_customers_tenant;
ALTER TABLE public.campaigns           VALIDATE CONSTRAINT fk_campaigns_tenant;
ALTER TABLE public.projects            VALIDATE CONSTRAINT fk_projects_tenant;
ALTER TABLE public.managers            VALIDATE CONSTRAINT fk_managers_tenant;
ALTER TABLE public.investors           VALIDATE CONSTRAINT fk_investors_tenant;
ALTER TABLE public.providers           VALIDATE CONSTRAINT fk_providers_tenant;
ALTER TABLE public.crops               VALIDATE CONSTRAINT fk_crops_tenant;
ALTER TABLE public.categories          VALIDATE CONSTRAINT fk_categories_tenant;
ALTER TABLE public.types               VALIDATE CONSTRAINT fk_types_tenant;
ALTER TABLE public.lease_types         VALIDATE CONSTRAINT fk_lease_types_tenant;
ALTER TABLE public.business_parameters VALIDATE CONSTRAINT fk_business_parameters_tenant;

COMMIT;
