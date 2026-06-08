# Ponti IA Sobre Axis v1

## Decision

Ponti is the first real product surface connected to Axis.

- `product_surface`: `ponti`
- Axis `org_id`: Ponti `auth_tenants.id`
- `external_tenant_id`: initially the same Ponti `auth_tenants.id`
- Initial auth mode: server-to-server API key plus delegated headers
- Initial rollout: read-only
- Legacy fallback: `ponti-ai` remains available through `AI_PROVIDER=legacy`

## Runtime Config

Backend:

- `AI_PROVIDER=legacy|axis`
- `AI_AXIS_ENABLED=false|true`
- `AXIS_COMPANION_BASE_URL`
- `AXIS_COMPANION_API_KEY`
- `AXIS_PRODUCT_SURFACE=ponti`
- `PONTI_AXIS_API_KEY`

Frontend:

- `VITE_AI_PROVIDER=legacy|axis`

The frontend flag is diagnostic/rollout metadata only. Ponti web and mobile must not call Axis directly in production; Ponti backend is the product boundary.

## Identity Mapping

Ponti middleware resolves the effective tenant from the current authenticated request. When `AI_PROVIDER=axis` and `AI_AXIS_ENABLED=true`, backend calls Axis Companion with:

- `X-Org-ID: <auth_tenants.id>`
- `X-User-ID: <actor>`
- `X-On-Behalf-Of: <actor>`
- `X-Product-Surface: ponti`
- `X-Auth-Scopes: companion:tasks:write companion:connectors:execute ponti:insights:read ...`

Axis must already have an active installation for `org_id + ponti`.

Axis calls Ponti product endpoints with `Authorization: Bearer <PONTI_API_KEY>`. Ponti accepts that bearer only on product integration endpoints when it matches `PONTI_AXIS_API_KEY`. In Axis, store the same value as the installation secret referenced by `secret_ref` (default local ref: `env:PONTI_API_KEY`).

## Workspace Schema

Ponti workspace v1:

```json
{
  "customer_id": 1,
  "project_id": 10,
  "campaign_id": 3,
  "field_id": 25
}
```

Current Axis `POST /v1/chat` rejects unknown top-level fields, so Ponti does not send `workspace` as a top-level chat field yet. When present in the web request, Ponti forwards it inside `handoff.workspace`, which is accepted by Axis' current chat DTO and keeps the context shape available for the future contract.

## Published Capabilities

First cut publishes one manifest:

- `ponti.insights`

Tools:

- `ponti.insights.list`
- `ponti.insights.summary`
- `ponti.insights.explain`

All tools are:

- `mode=read`
- `side_effect=false`
- `risk_class=low`
- tenant-scoped by `auth_tenants.id`

The broader target capabilities remain planned, not published in this cut:

- `ponti.dashboard.summary`
- `ponti.stock.summary`
- `ponti.workorders.list`
- `ponti.lots.summary`
- `ponti.reports.summary`
- `ponti.data_integrity.summary`

Reason: the current Axis PontiConnector executes only the three insights operations. Publishing tools that Axis cannot execute would make the planner overpromise.

## Ponti Endpoints For Axis

- `GET /api/v1/capabilities`
- `GET /api/v1/insights`
- `GET /api/v1/insights/summary`
- `GET /api/v1/insights/:id/explain`

Every data endpoint requires tenant context and returns evidence:

- `source_ref`
- `captured_at`
- `tenant_scope`
- `workspace`

## Local Onboarding And Smoke

Required local env:

```bash
export AXIS_COMPANION_BASE_URL=http://localhost:8081
export AXIS_COMPANION_API_KEY=local-dev-companion-api-key
export PONTI_BASE_URL=http://localhost:8080
export PONTI_ORG_ID=<auth_tenants.id>
export PONTI_AXIS_API_KEY=local-dev-ponti-axis-api-key
```

Axis Companion must be running with:

```bash
PONTI_BASE_URL=$PONTI_BASE_URL
PONTI_API_KEY=$PONTI_AXIS_API_KEY
```

Register product + installation + refresh connector:

```bash
scripts/axis/onboard-ponti.sh
```

Run the read-only smoke:

```bash
scripts/axis/smoke-ponti-axis-readonly.sh
```

The smoke validates:

- Ponti publishes `ponti.insights`.
- Axis registers `ponti`.
- Axis has an active installation for `PONTI_ORG_ID + ponti`.
- Axis discovers the Ponti connector.
- Axis executes `ponti.insights.summary`.
- Axis executes `ponti.insights.list`.
- Execution evidence includes `product_surface=ponti`.

## Chat Flow

`POST /api/v1/ai/chat` keeps the legacy web contract.

When Axis is enabled, Ponti backend calls `POST /v1/chat` on Companion and adapts the response into:

- `request_id`
- `output_kind=chat_reply`
- `content_language`
- `chat_id`
- `reply`
- `blocks`
- `tokens_used`
- `tool_calls`
- `pending_confirmations`
- `routed_agent`
- `routing_source=axis`
- `axis_run_id`
- `axis_task_id`

Streaming is compatibility-first: `/api/v1/ai/chat/stream` synthesizes SSE events from the non-streaming Axis response until Axis exposes a stable streaming contract for product chat.

## Fallback

Fallback to legacy `ponti-ai` happens only for Axis configuration/network/server failures. Axis 4xx responses are returned closed because they usually indicate auth, tenant, installation, policy or contract problems that must not be hidden.

## Next Phases

1. Expand Axis/Ponti contract to accept workspace as a first-class field.
2. Add generic connector execution for dashboard, stock, workorders, lots, reports and integrity.
3. Add draft-only capabilities governed by Nexus.
4. Add web evidence rendering beyond plain text.
5. Add mobile after web read-only is stable.
6. Retire `ponti-ai` only after fallback rate is near zero and smokes/evals are green.
