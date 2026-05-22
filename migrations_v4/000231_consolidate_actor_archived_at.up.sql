-- 000231_consolidate_actor_archived_at.up.sql
--
-- Consolidate Actor / ActorRole / ActorAlias lifecycle columns to a single
-- `deleted_at` column (GORM standard). Before this migration:
--
--   actors:        had BOTH `archived_at` (custom) and `deleted_at` (GORM)
--   actor_roles:   had ONLY `archived_at`
--   actor_aliases: had ONLY `archived_at`
--
-- This dual-column setup confused the source of truth ("which one means
-- archived?") and caused subtle bugs: queries on `archived_at IS NULL` could
-- diverge from queries on `deleted_at IS NULL`. After this migration the
-- single source of truth is `deleted_at`, matching every other CRUDAR entity
-- in the system.
--
-- Strategy: copy any non-NULL `archived_at` value into `deleted_at` (only
-- when `deleted_at` is still NULL to avoid clobbering), then drop the column
-- and its index. For `actor_roles` and `actor_aliases` we add `deleted_at`
-- first.

BEGIN;

-- ─────────────────────────── actors ───────────────────────────
-- Backfill: if archived_at was set but deleted_at wasn't, copy it over.
UPDATE public.actors
   SET deleted_at = archived_at
 WHERE archived_at IS NOT NULL
   AND deleted_at IS NULL;

DROP INDEX IF EXISTS public.idx_actors_archived_at;
ALTER TABLE public.actors DROP COLUMN IF EXISTS archived_at;

CREATE INDEX IF NOT EXISTS idx_actors_active
    ON public.actors (tenant_id)
    WHERE deleted_at IS NULL;

-- ──────────────────────── actor_roles ────────────────────────
ALTER TABLE public.actor_roles
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

UPDATE public.actor_roles
   SET deleted_at = archived_at
 WHERE archived_at IS NOT NULL
   AND deleted_at IS NULL;

DROP INDEX IF EXISTS public.idx_actor_roles_role;
ALTER TABLE public.actor_roles DROP COLUMN IF EXISTS archived_at;

CREATE INDEX IF NOT EXISTS idx_actor_roles_role
    ON public.actor_roles (role)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actor_roles_actor_active
    ON public.actor_roles (actor_id)
    WHERE deleted_at IS NULL;

-- ─────────────────────── actor_aliases ───────────────────────
ALTER TABLE public.actor_aliases
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone;

UPDATE public.actor_aliases
   SET deleted_at = archived_at
 WHERE archived_at IS NOT NULL
   AND deleted_at IS NULL;

DROP INDEX IF EXISTS public.ux_actor_aliases_actor_alias;
ALTER TABLE public.actor_aliases DROP COLUMN IF EXISTS archived_at;

CREATE UNIQUE INDEX IF NOT EXISTS ux_actor_aliases_actor_alias
    ON public.actor_aliases (actor_id, normalized_alias)
    WHERE deleted_at IS NULL;

COMMIT;
