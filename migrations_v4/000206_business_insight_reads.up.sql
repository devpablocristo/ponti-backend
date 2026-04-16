-- business_insight_reads: tracking de leida/no-leida por usuario.
-- 1 fila = "este usuario marco esta notificacion como leida".
-- Si una notificacion se reabre (status pasa de resolved a notified), se borran
-- todas sus rows aca para que vuelva a aparecer no-leida para todos.

BEGIN;

CREATE TABLE public.business_insight_reads (
    insight_id uuid NOT NULL REFERENCES public.business_insight_candidates(id) ON DELETE CASCADE,
    user_id text NOT NULL,
    read_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (insight_id, user_id)
);

CREATE INDEX idx_business_insight_reads_user ON public.business_insight_reads (user_id);

COMMIT;
