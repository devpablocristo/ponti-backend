-- ========================================
-- MIGRATION 000234 ACTOR UNIQUE NORMALIZED NAME (DOWN)
-- ========================================

BEGIN;

DROP INDEX IF EXISTS public.ux_actors_tenant_normalized_name_active;

CREATE INDEX IF NOT EXISTS idx_actors_tenant_normalized_name
    ON public.actors (tenant_id, normalized_name)
    WHERE deleted_at IS NULL
      AND merged_into_actor_id IS NULL;

COMMIT;
