# Ponti IA Sobre Axis v1

## Decision

Ponti is the first real product surface connected to Axis.

- `product_surface`: `ponti`
- Axis `org_id`: Ponti `auth_tenants.id`
- `external_tenant_id`: initially the same Ponti `auth_tenants.id`
- Initial auth mode: server-to-server API key plus delegated headers
- Initial rollout: read-only plus governed preview actions
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

Cloud Run deploy defaults for v1:

- Ponti backend workflows default to `AI_PROVIDER=axis` and
  `AI_AXIS_ENABLED=true`, overridable per environment for rollback.
- Ponti backend mounts `AXIS_COMPANION_API_KEY` and `PONTI_AXIS_API_KEY` from
  Secret Manager.
- Axis Companion mounts `PONTI_API_KEY` from Secret Manager and receives
  `PONTI_BASE_URL`.
- This v1 remains API-key based. Do not mix it with the older internal
  JWT/JWKS cutover notes.

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

Read tools:

- `ponti.insights.list`
- `ponti.insights.summary`
- `ponti.insights.explain`

Read tools are:

- `mode=read`
- `side_effect=false`
- `risk_class=low`
- tenant-scoped by `auth_tenants.id`

Governed preview tools:

- `ponti.insight.resolve.prepare`
- `ponti.workorder.draft.prepare`
- `ponti.stock_adjustment.prepare`

Preview tools are:

- `mode=write`
- `side_effect=true`
- `risk_class=medium`
- `governance.requires_approval=true`
- `governance.action_type=agent.capability.invoke`
- preview-only: they prepare proposals but do not mutate Ponti data.

The broader target capabilities remain planned, not published in this cut:

- `ponti.dashboard.summary`
- `ponti.stock.summary`
- `ponti.workorders.list`
- `ponti.lots.summary`
- `ponti.reports.summary`
- `ponti.data_integrity.summary`

Reason: the current Axis PontiConnector executes the three insight reads and
the three governed preview actions. Publishing broader dashboard/stock/report
tools before the connector can execute them would make the planner overpromise.

Draft action contracts are published in the same `ponti.insights` manifest for
the first connector cut because Axis currently discovers that manifest. If
Ponti later publishes multiple manifests, Axis PontiConnector must expand
discovery before Ponti splits these tools into a separate package.

## Ponti Endpoints For Axis

- `GET /api/v1/capabilities`
- `GET /api/v1/insights`
- `GET /api/v1/insights/summary`
- `GET /api/v1/insights/:id/explain`
- `POST /api/v1/ai/actions/insight-resolve/prepare`
- `POST /api/v1/ai/actions/workorder-draft/prepare`
- `POST /api/v1/ai/actions/stock-adjustment/prepare`

Every data endpoint requires tenant context and returns evidence:

- `source_ref`
- `captured_at`
- `tenant_scope`
- `workspace`

Draft action endpoints return previews only. They validate the requested shape,
thread `tenant_scope`, `actor_id` and `workspace`, and return
`approval_required=true`, `nexus_action_type=agent.capability.invoke`,
`preview_only=true`, `write_performed=false` and `execution_allowed=false`.
They do not resolve insights, create work-order drafts, publish work orders or
apply stock adjustments.

## Operational Insights

Ponti is the source of business insight semantics. Axis only consumes the
published read capabilities.

Current insight producers:

- `ponti.stock.negative`: reactive insight from stock real updates.
- `ponti.data_integrity.critical`: reactive insight from
  `GET /api/v1/data-integrity/costs-check` when one or more controls return
  `ERROR`.
- `ponti.data_integrity.tentative_prices`: reactive insight from
  `GET /api/v1/data-integrity/tentative-prices` when one or more supplies have
  tentative/partial prices.
- `ponti.report.operating_result.negative`: reactive insight from
  `GET /api/v1/reports/summary-results` when the project total operating result
  is negative.

`ponti.data_integrity.critical` is scoped by tenant and project. Its evidence
includes:

- `project_id`
- `failed_checks`
- `total_checks`
- `failed_controls`
- `suggested_action=review_data_integrity`
- `source_ref=ponti.data_integrity.costs_check`

When the same project returns all integrity controls as `OK`, Ponti resolves the
open `ponti.data_integrity.critical` candidate automatically. This remains
read-only: it creates/updates/resolves insight candidates but does not execute
business actions.

`ponti.data_integrity.tentative_prices` is scoped by tenant and project when a
project workspace is provided. Its evidence includes:

- `project_id`
- `customer_id`
- `campaign_id`
- `field_id`
- `count`
- `sample_items`
- `suggested_action=review_tentative_prices`
- `source_ref=ponti.data_integrity.tentative_prices`

When the same project no longer has tentative prices, Ponti resolves the open
`ponti.data_integrity.tentative_prices` candidate automatically.

`ponti.report.operating_result.negative` is also scoped by tenant and project.
Its evidence includes:

- `project_id`
- `customer_id`
- `campaign_id`
- `total_operating_result_usd`
- `project_return_pct`
- `total_invested_project_usd`
- `negative_crops`
- `suggested_action=review_summary_results`
- `source_ref=ponti.reports.summary_results`

When the project total operating result is zero or positive again, Ponti
resolves the open `ponti.report.operating_result.negative` candidate
automatically.

## Local Onboarding And Read-Only Smoke

Required local env:

```bash
export AXIS_COMPANION_BASE_URL=http://localhost:18085
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

## Local Chat Smoke

Ponti API must run with Axis provider enabled:

```bash
AI_PROVIDER=axis
AI_AXIS_ENABLED=true
AXIS_COMPANION_BASE_URL=http://host.docker.internal:18085
AXIS_COMPANION_API_KEY=<companion-api-key-for-PONTI_ORG_ID>
AXIS_PRODUCT_SURFACE=ponti
```

Run:

```bash
export PONTI_BASE_URL=http://localhost:8080
export PONTI_API_KEY=<ponti-backend-x-api-key>
export PONTI_ORG_ID=<auth_tenants.id>
export PONTI_PROJECT_ID=<project id owned by tenant>
scripts/axis/smoke-ponti-axis-chat.sh
```

The chat smoke validates:

- `POST /api/v1/ai/chat` returns the legacy web shape.
- `routing_source=axis`.
- `reply` is non-empty.
- optional technical ids such as `axis_run_id` and `axis_task_id` remain
  compatible with web clients.
- `/api/v1/ai/chat/stream` emits compatible SSE events: `start`, `text`, `done`.

## Local Full Smoke

Run every Ponti/Axis local smoke in the supported order:

```bash
make smoke-axis-all
```

This target executes:

1. onboarding;
2. read-only capability execution;
3. draft action conformance/governance checks;
4. direct preview endpoints;
5. Nexus-approved draft preview execution through Axis;
6. chat through Axis.

The Nexus-approved smoke is local and idempotent. It creates or updates the
`agent.capability.invoke` action type and a Ponti policy if needed, approves one
request, executes `ponti.workorder.draft.prepare` through Axis, validates that
Ponti still returns preview-only output, and reports the Nexus request as
`executed`.

## Local Draft Preview Smoke

Ponti API must be running. In normal local development, run the smoke with the
local API key:

```bash
export PONTI_BASE_URL=http://localhost:8080
export PONTI_ORG_ID=<auth_tenants.id>
export PONTI_AUTH_MODE=local
export PONTI_API_KEY=<ponti-backend-x-api-key>
export PONTI_PROJECT_ID=<project id owned by tenant>
export PONTI_FIELD_ID=<optional field id>
export PONTI_CAMPAIGN_ID=<optional campaign id>
export PONTI_SUPPLY_ID=<optional supply id>
scripts/axis/smoke-ponti-axis-draft-previews.sh
```

To validate the exact Axis product integration auth path, run the same smoke
with bearer auth and make sure `PONTI_AXIS_API_KEY` matches the backend env:

```bash
export PONTI_BASE_URL=http://localhost:8080
export PONTI_ORG_ID=<auth_tenants.id>
export PONTI_AUTH_MODE=axis
export PONTI_AXIS_API_KEY=<ponti-axis-product-key>
export PONTI_PROJECT_ID=<project id owned by tenant>
export PONTI_FIELD_ID=<optional field id>
export PONTI_CAMPAIGN_ID=<optional campaign id>
export PONTI_SUPPLY_ID=<optional supply id>
scripts/axis/smoke-ponti-axis-draft-previews.sh
```

The draft preview smoke validates:

- all three prepare endpoints return `status=preview`;
- `approval_required=true`;
- `nexus_action_type=agent.capability.invoke`;
- `preview_only=true`;
- `write_performed=false`;
- `execution_allowed=false`;
- tenant/workspace evidence is present;
- invalid zero-delta stock adjustment fails with `400`.

Axis connector execution of these preview tools is guarded by Nexus. A direct
connector execution without `nexus_request_id` is expected to fail before Axis
calls Ponti. With an approved Nexus request and an `idempotency_key`, Axis may
call the preview endpoint, which still returns `write_performed=false`.

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

## Web Read-Only UI

`AIAssistant` remains the single web surface. It supports both legacy and Axis
responses and preserves:

- `blocks`
- `tool_calls`
- `pending_confirmations`
- `routing_source`
- `axis_run_id`
- `axis_task_id`

When Axis sends non-text blocks or tool metadata, the UI renders a compact
evidence section under the assistant message. `VITE_AI_PROVIDER=legacy|axis`
controls the visible provider mode only; it does not make the browser call Axis
directly.

## Fallback

Fallback to legacy `ponti-ai` happens only for Axis configuration/network/server failures. Axis 4xx responses are returned closed because they usually indicate auth, tenant, installation, policy or contract problems that must not be hidden.

## Next Phases

1. Expand Axis/Ponti contract to accept workspace as a first-class field.
2. Add generic connector execution for dashboard, stock, workorders, lots, reports and integrity.
3. Add more Ponti insight producers: low stock and overdue operational work once thresholds are defined.
4. Add mobile after web read-only is stable.
5. Retire `ponti-ai` only after fallback rate is near zero and smokes/evals are green.

## Backlog Despues

Contrato Axis/Ponti:

- Make `workspace` a first-class field in Axis chat/runtime instead of sending
  it only inside `handoff.workspace`.
- Replace the initial API-key delegated headers with internal JWT/JWKS when
  Axis production auth is ready.
- Keep the `ponti-golden` eval pack expanded as new capabilities are published.

Capabilities:

- Publish and execute `ponti.dashboard.summary`.
- Publish and execute `ponti.stock.summary`.
- Publish and execute `ponti.workorders.list`.
- Publish and execute `ponti.lots.summary`.
- Publish and execute `ponti.reports.summary`.
- Publish and execute `ponti.data_integrity.summary`.
- Keep capability manifests aligned with what Axis can actually execute; do not
  publish planner-facing tools before the connector path exists.
- Extend Axis PontiConnector discovery beyond `ponti.insights` if Ponti starts
  publishing multiple manifests.
- Map product manifest governance metadata into Axis canonical fields including
  Nexus action type, idempotency and postconditions for write conformance.

Insights:

- Define the business threshold for low stock before implementing
  `ponti.stock.low`.
- Define what "overdue operational work" means before implementing overdue
  work-order/labor insights.
- Add margin anomaly insights only when the comparison baseline is explicit
  enough: budget, historical campaign, planned cost or accepted threshold.
- Add feedback states for AI usefulness: useful, not useful, false positive,
  resolved by action.

Draft actions:

- Add an end-to-end smoke that creates/approves the Nexus request and then
  executes a Ponti preview capability through Axis.
- Require `approval_required=true`, `side_effect_type=write` and
  `nexus_action_type=agent.capability.invoke` in every draft/write manifest.
- Convert preview-only endpoints into real draft creation only after Nexus
  approval is present and replay/idempotency are handled.
- Do not let chat execute sensitive writes directly.

Web:

- Add richer evidence rendering for source references, captured dates and
  workspace provenance.
- Add UI for pending confirmations once Nexus action requests are returned by
  Axis chat.
- Add fallback visibility and metrics in the UI so operators can see when Axis
  degraded to legacy.

Mobile:

- Add mobile AI only after web read-only is stable.
- Mobile must call Ponti backend, not Axis directly.
- Start with text/voice query and OT explanation.
- Keep photo/document ingestion as attached evidence first, not autonomous
  interpretation.

Operations:

- Add smokes for chat fallback, connector refresh and tenant leakage.
- Add an end-to-end smoke for each published Ponti capability.
- Add observability correlation between Ponti request id, Axis run/task id and
  Nexus request id.
- Track fallback rate, tool success rate, latency and cost per tenant/product.

Legacy:

- Keep `ponti-ai` until Axis covers current web conversations and fallback rate
  is near zero.
- Remove legacy only after web and mobile no longer depend on legacy-only
  response fields or endpoints.

External intelligence:

- Add weather, satellite/NDVI, commodity prices, dollar and input price feeds
  only after the internal read/draft loop is stable.
- Treat external data as evidence with freshness and source metadata, not as
  unqualified model context.
