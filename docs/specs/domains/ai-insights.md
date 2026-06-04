# AI And Business Insights Domain Baseline Specification

Specification type: baseline current-state domain specification.

## Purpose

Proxy AI chat/conversation APIs and manage tenant-scoped business insight candidates/read/resolve state.

## Boundaries

Owns:
- Backend AI proxy endpoints.
- Business insight candidate storage.
- Business insight read/unread and resolve/reopen behavior.
- Negative stock insight generation policy call path.

Does not own:
- Ponti AI service implementation.
- Review/Nexus policy definitions.
- Stock accounting.

## Owned Entities

- `business_insight_candidates`
- `business_insight_reads`

AI conversation persistence is UNKNOWN locally.

## Owned APIs

- `POST /api/v1/ai/chat`
- `POST /api/v1/ai/chat/stream`
- `GET /api/v1/ai/chat/conversations`
- `GET /api/v1/ai/chat/conversations/:conversation_id`
- `GET /api/v1/insights`
- `POST /api/v1/insights/:id/read`
- `DELETE /api/v1/insights/:id/read`
- `POST /api/v1/insights/:id/resolve`
- `DELETE /api/v1/insights/:id/resolve`

## Dependencies On Other Domains

- Platform, Identity, And Admin for actor/org context.
- Inventory And Stock for negative stock trigger.

## Inbound Dependencies

- Stock usecases call this domain through an optional notifier.
- Frontend/BFF likely consumes AI And Business Insights APIs, but external code is unavailable.

## Outbound Dependencies

- Ponti AI service.
- Review/Nexus governance service.
- PostgreSQL.

## Aggregate Roots

- `BusinessInsightCandidate`
- `InsightRead`

AI `Conversation` aggregate is external or UNKNOWN.

## Critical Business Rules

- AI proxy sends `X-SERVICE-KEY`, `X-USER-ID`, and `X-PROJECT-ID`.
- If AI service is not configured, chat returns a dummy response.
- If AI stream is not configured, stream emits `ai_not_configured`.
- Conversation list limit defaults to 50 and caps at 200.
- Negative stock notification only applies when quantity is less than 0.
- Review/Nexus must allow policy before a candidate is recorded.
- Negative stock candidates dedupe by bucketed fingerprint using a default 6 hour window.
- Stock returning to non-negative attempts auto-resolution.
- Read/unread is per user.
- Resolve/reopen is tenant-scoped.

## Tenant Isolation Requirements

- Business insight candidates are physically tenant-scoped by `tenant_id`.
- Insight operations require `OrgID` from request context.
- AI tenant propagation beyond user/project headers is UNKNOWN.

## Security Requirements

- Baseline auth applies.
- Insight operations require actor and org id.
- Ponti AI calls require configured service key for real external calls.

## Evidence

- `internal/ai/handler.go`
- `internal/ai/client.go`
- `internal/ai/usecases/usecases.go`
- `internal/businessinsights/handler.go`
- `internal/businessinsights/service.go`
- `internal/businessinsights/repository.go`
- `internal/stock/usecases.go`
- `internal/reviewproxy/client.go`
- `migrations_v4/000209_business_insight_candidates.up.sql`
- `migrations_v4/000210_business_insight_reads.up.sql`
