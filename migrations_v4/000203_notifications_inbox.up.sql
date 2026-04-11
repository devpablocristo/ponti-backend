CREATE TABLE IF NOT EXISTS public.ponti_notification_candidates (
    id BIGSERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    project_id BIGINT NULL,
    candidate_key TEXT NOT NULL,
    kind TEXT NOT NULL,
    source TEXT NOT NULL,
    source_ref TEXT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    severity SMALLINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'new',
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ponti_notification_candidates_key
    ON public.ponti_notification_candidates (org_id, candidate_key);

CREATE INDEX IF NOT EXISTS idx_ponti_notification_candidates_project
    ON public.ponti_notification_candidates (org_id, project_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS public.ponti_notifications (
    id BIGSERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    project_id BIGINT NULL,
    recipient_actor TEXT NULL,
    kind TEXT NOT NULL,
    source TEXT NOT NULL,
    source_ref TEXT NULL,
    notification_key TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    severity SMALLINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'new',
    route_hint TEXT NOT NULL DEFAULT 'copilot',
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at TIMESTAMPTZ NULL,
    dismissed_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ponti_notifications_key
    ON public.ponti_notifications (org_id, notification_key);

CREATE INDEX IF NOT EXISTS idx_ponti_notifications_inbox
    ON public.ponti_notifications (org_id, project_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ponti_notifications_actor
    ON public.ponti_notifications (org_id, recipient_actor, status, created_at DESC);

CREATE TABLE IF NOT EXISTS public.ponti_notification_deliveries (
    id BIGSERIAL PRIMARY KEY,
    notification_id BIGINT NOT NULL REFERENCES public.ponti_notifications(id) ON DELETE CASCADE,
    channel TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    external_ref TEXT NULL,
    error_message TEXT NULL,
    attempted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ponti_notification_deliveries_notification
    ON public.ponti_notification_deliveries (notification_id, channel);

CREATE TABLE IF NOT EXISTS public.ponti_notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    org_id UUID NOT NULL,
    actor TEXT NOT NULL,
    notification_kind TEXT NOT NULL,
    channel TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    frequency TEXT NOT NULL DEFAULT 'immediate',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ponti_notification_preferences_key
    ON public.ponti_notification_preferences (org_id, actor, notification_kind, channel);
