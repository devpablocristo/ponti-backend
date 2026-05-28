-- 000231_consolidate_actor_archived_at.down.sql
--
-- Reverse the Actor lifecycle consolidation. Re-introduces `archived_at` and
-- copies `deleted_at` back into it. `deleted_at` stays on `actors` (it was
-- pre-existing); on `actor_roles` / `actor_aliases` it is dropped because
-- those tables didn't have it before 000231.

BEGIN;

-- ─────────────────────────── actors ───────────────────────────
ALTER TABLE public.actors
    ADD COLUMN IF NOT EXISTS archived_at timestamp with time zone;

UPDATE public.actors
   SET archived_at = deleted_at
 WHERE deleted_at IS NOT NULL
   AND archived_at IS NULL;

DROP INDEX IF EXISTS public.idx_actors_active;
CREATE INDEX IF NOT EXISTS idx_actors_archived_at
    ON public.actors (archived_at);

-- ──────────────────────── actor_roles ────────────────────────
ALTER TABLE public.actor_roles
    ADD COLUMN IF NOT EXISTS archived_at timestamp with time zone;

UPDATE public.actor_roles
   SET archived_at = deleted_at
 WHERE deleted_at IS NOT NULL
   AND archived_at IS NULL;

DROP INDEX IF EXISTS public.idx_actor_roles_actor_active;
DROP INDEX IF EXISTS public.idx_actor_roles_role;
ALTER TABLE public.actor_roles DROP COLUMN IF EXISTS deleted_at;

CREATE INDEX IF NOT EXISTS idx_actor_roles_role
    ON public.actor_roles (role)
    WHERE archived_at IS NULL;

-- ─────────────────────── actor_aliases ───────────────────────
ALTER TABLE public.actor_aliases
    ADD COLUMN IF NOT EXISTS archived_at timestamp with time zone;

UPDATE public.actor_aliases
   SET archived_at = deleted_at
 WHERE deleted_at IS NOT NULL
   AND archived_at IS NULL;

DROP INDEX IF EXISTS public.ux_actor_aliases_actor_alias;
ALTER TABLE public.actor_aliases DROP COLUMN IF EXISTS deleted_at;

CREATE UNIQUE INDEX IF NOT EXISTS ux_actor_aliases_actor_alias
    ON public.actor_aliases (actor_id, normalized_alias)
    WHERE archived_at IS NULL;

COMMIT;
