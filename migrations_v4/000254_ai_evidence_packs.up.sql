-- ai_evidence_packs: cache local de evidence packs firmados de Nexus por request gobernada.

BEGIN;

CREATE TABLE public.ai_evidence_packs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    nexus_request_id text NOT NULL,
    pack_json jsonb NOT NULL,
    signature_key_id text,
    signature text,
    retrieved_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ai_evidence_packs_nexus_request_uniq UNIQUE (tenant_id, nexus_request_id)
);

COMMIT;
