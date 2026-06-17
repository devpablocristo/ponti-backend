-- ai_decision_runs / ai_decision_cards: feed operativo IA con dedupe, auditoria y estados.

BEGIN;

CREATE TABLE public.ai_decision_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    workspace_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    requested_by text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'completed',
    routing_source text NOT NULL DEFAULT 'deterministic',
    axis_run_id text NOT NULL DEFAULT '',
    axis_task_id text NOT NULL DEFAULT '',
    degraded_reason text NOT NULL DEFAULT '',
    cards_created integer NOT NULL DEFAULT 0,
    cards_updated integer NOT NULL DEFAULT 0,
    cards_total integer NOT NULL DEFAULT 0,
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_decision_runs_tenant_created
    ON public.ai_decision_runs (tenant_id, created_at DESC);

CREATE TABLE public.ai_decision_cards (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    decision_run_id uuid REFERENCES public.ai_decision_runs(id) ON DELETE SET NULL,
    workspace_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    fingerprint text NOT NULL,
    domain text NOT NULL,
    route_hint text NOT NULL,
    severity text NOT NULL,
    bucket text NOT NULL,
    status text NOT NULL DEFAULT 'open',
    title text NOT NULL,
    summary text NOT NULL DEFAULT '',
    recommendation text NOT NULL DEFAULT '',
    impact_label text NOT NULL DEFAULT '',
    impact_value double precision,
    source text NOT NULL DEFAULT '',
    evidence_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    tools_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    action_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    axis_run_id text NOT NULL DEFAULT '',
    axis_task_id text NOT NULL DEFAULT '',
    occurrence_count integer NOT NULL DEFAULT 1,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    snooze_until timestamptz,
    status_changed_at timestamptz,
    last_actor text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ai_decision_cards_fingerprint_uniq UNIQUE (tenant_id, fingerprint),
    CONSTRAINT ai_decision_cards_status_check CHECK (status IN ('open','accepted','drafted','dismissed','snoozed','resolved'))
);

CREATE INDEX idx_ai_decision_cards_tenant_status_seen
    ON public.ai_decision_cards (tenant_id, status, last_seen_at DESC);

CREATE INDEX idx_ai_decision_cards_tenant_route
    ON public.ai_decision_cards (tenant_id, route_hint, bucket);

COMMIT;
