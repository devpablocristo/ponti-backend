-- ai_action_drafts: borradores accionables del agente (Ola B.1); staged en el preview y applied solo tras verificacion Nexus.

BEGIN;

CREATE TABLE public.ai_action_drafts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    draft_type text NOT NULL,
    status text NOT NULL DEFAULT 'staged',
    payload_json jsonb NOT NULL,
    nexus_request_id text,
    created_by text,
    applied_at timestamptz,
    applied_by text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ai_action_drafts_type_check CHECK (draft_type IN ('insight_resolution','stock_count','stock_adjustment')),
    CONSTRAINT ai_action_drafts_status_check CHECK (status IN ('staged','applied','discarded'))
);

CREATE INDEX idx_ai_action_drafts_tenant_status
    ON public.ai_action_drafts (tenant_id, status);

COMMIT;
