-- In-app notifications + business insight candidates (patrón pymes/nexus).
-- Notificaciones persistentes por usuario dentro de un tenant (bandeja del asistente).

BEGIN;

-- =============================================================================
-- 1. in_app_notifications — bandeja visible por el usuario
-- =============================================================================
CREATE TABLE public.in_app_notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    title text NOT NULL,
    body text NOT NULL,
    kind text NOT NULL DEFAULT 'system',
    entity_type text NOT NULL DEFAULT '',
    entity_id text NOT NULL DEFAULT '',
    chat_context jsonb NOT NULL DEFAULT '{}'::jsonb,
    read_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_in_app_notifications_user_created
    ON public.in_app_notifications (user_id, created_at DESC);

CREATE INDEX idx_in_app_notifications_tenant_created
    ON public.in_app_notifications (tenant_id, created_at DESC);

-- =============================================================================
-- 2. business_insight_candidates — dedup + cooldown por fingerprint (tenant)
-- =============================================================================
CREATE TABLE public.business_insight_candidates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    kind text NOT NULL,
    event_type text NOT NULL,
    entity_type text NOT NULL,
    entity_id text NOT NULL DEFAULT '',
    fingerprint text NOT NULL,
    severity text NOT NULL DEFAULT 'info',
    status text NOT NULL DEFAULT 'new',
    title text NOT NULL,
    body text NOT NULL,
    evidence_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    occurrence_count integer NOT NULL DEFAULT 1,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    first_notified_at timestamptz,
    last_notified_at timestamptz,
    resolved_at timestamptz,
    last_actor text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT business_insight_candidates_fingerprint_uniq UNIQUE (tenant_id, fingerprint)
);

CREATE INDEX idx_business_insight_candidates_tenant_status
    ON public.business_insight_candidates (tenant_id, status, last_seen_at DESC);

CREATE INDEX idx_business_insight_candidates_tenant_entity
    ON public.business_insight_candidates (tenant_id, entity_type, entity_id);

COMMIT;
