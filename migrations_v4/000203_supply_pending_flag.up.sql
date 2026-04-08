BEGIN;

ALTER TABLE public.supplies
    ADD COLUMN IF NOT EXISTS is_pending boolean NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_supplies_pending_notdel
    ON public.supplies USING btree (project_id, is_pending, name)
    WHERE deleted_at IS NULL;

COMMIT;
