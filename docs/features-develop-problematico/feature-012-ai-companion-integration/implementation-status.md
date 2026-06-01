# implementation-status.md — feature-012 (BE)

## Estado general

- **Estado**: COMPLETA en SOURCE (`777e5f6a`), con tests a nivel de package.
- **% completitud (BE)**: ~95% (código + tests presentes; falta solo el wiring que vive en features 005/001/023 y la regeneración de `wire_gen.go`).
- **Estado en este repo (develop `003a9b8f`)**: NO presente. `internal/axis` ausente; `internal/ai/client.go` legacy aún existe; handler/usecases en forma vieja (sin `tenantID`); `wire/ai_providers.go` wirea el cliente legacy.
- **Estado en el otro repo (FE feature-012)**: desconocido desde aquí; según la nota existe (`pages/admin/ai`, BFF `ai.ts`, `managerChatStreamProxy`, `AIAssistant.tsx`). Coordinar.

## Tests

| test | cubre | confianza |
|---|---|---|
| `internal/axis/client_test.go` | `NewCompanionClient` (ErrNotConfigured), `Chat`, claims JWT firmados (httptest) | alta |
| `internal/ai/companion_adapter_test.go` | round-trip Chat, mapeo de respuesta, SSE sintético, fallback huérfano | alta |
| `internal/ai/handler_test.go` | `extractIDs` con principal del contexto, validación tenant-scope (sqlite in-mem), `AI_TENANT_SCOPE` | alta (requiere `shared/authz` + `contextkeys`) |
| `internal/ai/usecases/usecases_test.go` | passthrough con `tenantID`, clamp de limit | alta |

- No hay tests de `internal/axis/nexus_*` (cliente inactivo) → cobertura nula pero código no usado en runtime.
- Falta test de integración real contra Companion (esperable: requiere servicio).

## Pendientes

### BLOQUEANTE para mergear
- Presencia de `cmd/config/companion.go` (`config.Companion`/`config.Nexus`) y `Security.AITenantScope` (feature-005).
- Presencia de `internal/shared/authz.PrincipalFromContext` (feature-001).
- Reescritura de `wire/ai_providers.go` (Companion en vez de `ai.Client`) + regeneración de `wire/wire_gen.go` (feature-023).
- Eliminar `internal/ai/client.go` legacy.
- `go build ./...` + `go test ./internal/ai/... ./internal/axis/...` verdes.

### Mejora futura
- Streaming token-por-token real cuando Companion gane SSE (reescribir `DoStream` del adapter; no toca handler ni FE).
- Propagar/filtrar `project_id` a Companion vía metadata de task (hoy `_ = projectID`).
- Cablear cliente Nexus para gating de acciones sensibles (allow/deny/require_approval).

### Deuda aceptable
- `axis/nexus_*` presente pero no usado (preparación para gating). Compila, sin runtime path.
- Claims JWT duplican `scope` (string OAuth2) y `scopes` (array legacy) por compatibilidad con consumidores.
- `marshalCompanionChat` rellena campos legacy (`request_id`, `tokens_used:0`, `content_language:"es"`) con defaults porque Companion no los devuelve.

### Duda humana
- ¿La sección `Review` (Nexus Review/approvals) referenciada por `loadconfig.go` viene completa con feature-005? (no está en este flist).
- ¿El FE feature-012 ya está adaptado a la ausencia de streaming progresivo? Confirmar con el equipo FE.
- ¿El secret `COMPANION_INTERNAL_JWT_SECRET` está sincronizado en prod con el que valida Companion?

## Bugs / observaciones

- No se detectaron bugs en el diff leído. El fallback de `chat_id`/`task_id` huérfano (retry sin IDs en 404) está bien acotado (solo si traía identificador y el error es NOT_FOUND).
- Cambio de comportamiento operativo: AI deja de ser opcional. Sin `COMPANION_BASE_URL`/`SECRET` el binary NO arranca (antes había fallback dummy). Documentar en deploy.
