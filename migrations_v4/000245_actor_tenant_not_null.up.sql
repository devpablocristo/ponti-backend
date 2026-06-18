BEGIN;

-- Pilar 3 — Hardening: tenant_id NOT NULL en actors + actor_keys.
-- El resolver SIEMPRE llena un tenant concreto (OrgID del ctx, o 'default'); con NULL
-- el índice único parcial uq_actor_keys_active no deduplicaría (NULLs distintos) y el
-- scoping `tenant_id = ?` excluiría filas silenciosamente. Backfill defensivo antes del
-- NOT NULL (no-op si no hay filas, que es el caso con IDENTITY_GATE off).

UPDATE public.actors
   SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name = 'default' LIMIT 1)
 WHERE tenant_id IS NULL;

UPDATE public.actor_keys
   SET tenant_id = (SELECT id FROM public.auth_tenants WHERE name = 'default' LIMIT 1)
 WHERE tenant_id IS NULL;

ALTER TABLE public.actors     ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE public.actor_keys ALTER COLUMN tenant_id SET NOT NULL;

COMMIT;
