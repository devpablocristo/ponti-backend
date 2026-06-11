-- ai_governance_requests: espejo local de requests gobernadas por Nexus (inbox de approvals + auditoria de ejecucion).

BEGIN;

CREATE TABLE public.ai_governance_requests (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    nexus_request_id text NOT NULL,
    action_type text NOT NULL,
    origin text NOT NULL,
    requester_id text,
    status text NOT NULL,
    decision text,
    risk_level text,
    binding_hash text,
    action_binding_json jsonb,
    params_json jsonb,
    payload_json jsonb,
    entity_type text,
    entity_id text,
    approval_id text,
    decided_by text,
    executed_at timestamptz,
    result_json jsonb,
    error_message text,
    axis_run_id text,
    axis_task_id text,
    ponti_request_id text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ai_governance_requests_nexus_request_uniq UNIQUE (tenant_id, nexus_request_id),
    CONSTRAINT ai_governance_requests_origin_check CHECK (origin IN ('agent','watcher'))
);

CREATE INDEX idx_ai_governance_requests_tenant_status
    ON public.ai_governance_requests (tenant_id, status);

CREATE INDEX idx_ai_governance_requests_nexus_request
    ON public.ai_governance_requests (nexus_request_id);

COMMIT;
