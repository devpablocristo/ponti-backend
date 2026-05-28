-- ========================================
-- MIGRATION 000234 ACTOR UNIQUE NORMALIZED NAME (UP)
-- ========================================
-- Un actor activo es una identidad unica por tenant. El nombre normalizado no
-- puede repetirse aunque cambien tipo, roles, perfiles o datos de contacto.

BEGIN;

DROP INDEX IF EXISTS public.idx_actors_tenant_normalized_name;

CREATE UNIQUE INDEX IF NOT EXISTS ux_actors_tenant_normalized_name_active
    ON public.actors (tenant_id, normalized_name)
    WHERE deleted_at IS NULL
      AND merged_into_actor_id IS NULL;

COMMIT;
