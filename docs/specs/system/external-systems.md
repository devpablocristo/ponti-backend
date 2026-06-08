# External Systems Baseline Specification

Specification type: baseline current-state specification.

## Verified External Systems

### PostgreSQL / Cloud SQL

Purpose: primary persistence and reporting views/functions.

Direction: backend -> database.

Evidence:
- `cmd/config/db.go`
- `internal/platform/persistence/gorm/*`
- `migrations_v4/*`
- `.github/workflows/*`

### Google Identity Platform / JWKS

Purpose: JWT verification and identity context.

Direction: backend -> identity provider.

Evidence:
- `internal/platform/http/middlewares/gin/require_identity_platform_authz.go`
- `cmd/config/auth.go`

### Firebase / Identity Platform Admin

Purpose: admin user provisioning/reset behavior.

Direction: backend -> identity admin service.

Evidence:
- `internal/admin/idp/firebase_admin.go`
- `internal/admin/handler.go`

### Ponti AI Service

Purpose: AI chat, streaming chat, and conversation APIs proxied by backend.

Direction: backend -> AI service.

Verified proxied paths:
- `/v1/chat`
- `/v1/chat/stream`
- `/v1/chat/conversations`
- `/v1/chat/conversations/:conversation_id`

Evidence:
- `internal/ai/client.go`
- `internal/ai/usecases/usecases.go`
- `internal/ai/handler.go`

UNKNOWN:
- AI service implementation.
- AI service storage.
- AI service auth model beyond `X-SERVICE-KEY`, `X-USER-ID`, and `X-PROJECT-ID` headers sent by this backend.

### Review / Nexus Governance Service

Purpose: evaluate whether negative stock should create a business insight candidate.

Direction: backend -> governance/review service.

Evidence:
- `internal/reviewproxy/client.go`
- `internal/businessinsights/service.go`
- `cmd/api/http_server.go`
- `cmd/config/review.go`

UNKNOWN:
- Policy definitions.
- Governance service storage.
- Governance service deployment.

### GCP / Cloud Run / Artifact Registry / Secret Manager / Workload Identity

Purpose: deployment and runtime infrastructure as referenced by workflows.

Direction: CI/CD -> GCP.

Evidence:
- `.github/workflows/deploy-dev.yml`
- `.github/workflows/deploy-staging.yml`
- `.github/workflows/deploy-prod.yml`
- `.github/workflows/rollback-staging.yml`
- `.github/workflows/rollback-prod.yml`

UNKNOWN:
- Full infrastructure-as-code source.
- Complete network topology.
- Production scaling settings beyond workflow evidence.

### Frontend / BFF

Purpose: external consumer of backend REST API.

Evidence:
- Repository README and Makefile references indicate frontend/BFF/OpenAPI consumption.

UNKNOWN:
- Frontend source code.
- BFF source code.
- Client-side routing, authorization, and UX behavior.

## Not Verified

The current repository does not verify runtime use of:

- Kafka
- RabbitMQ
- Pub/Sub
- GraphQL server
- gRPC server
- Redis cache

Evidence:
- Route registration is HTTP/Gin based in `internal/*/handler.go`.
- No runtime queue/GraphQL/gRPC contracts were found in the audited code.
