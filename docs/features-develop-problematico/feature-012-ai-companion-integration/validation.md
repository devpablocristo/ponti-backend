# validation.md — feature-012 (BE)

## Checklist pre-PR (en este repo)

- [ ] `internal/axis/{client,client_test,errors,http,jwt,types,nexus_client,nexus_types}.go` presentes.
- [ ] `internal/ai/companion_adapter.go` + `companion_adapter_test.go` presentes.
- [ ] `internal/ai/client.go` **eliminado**.
- [ ] `internal/ai/handler.go` con firma `NewHandler(u, s, c, m, db, aiTenantScope)` y `extractIDs` como método (3 returns: userID, tenantID, projectID).
- [ ] `internal/ai/usecases/usecases.go` con `ClientPort` y métodos que incluyen `tenantID`; sin `dummyOrReal`/`isAIServiceNotConfigured`.
- [ ] `wire/companion_providers.go` presente.
- [ ] `wire/ai_providers.go` reescrito (Companion, no `ai.Client`).
- [ ] `wire/wire_gen.go` regenerado con `wire`.
- [ ] Dependencias presentes: `cmd/config/companion.go`, `Security.AITenantScope`, `internal/shared/authz.PrincipalFromContext`.
- [ ] `grep -rn "ai.NewClient\|ProvideAIClient\|config.AI\b\|X-SERVICE-KEY" internal/ai wire cmd/config` → SIN resultados.

## Tests sugeridos (BE)

```bash
go build ./...
go vet ./...
go test ./internal/ai/... ./internal/axis/... -count=1
# opcional, foco:
go test ./internal/ai -run TestCompanionAdapter -v
go test ./internal/axis -run TestNewCompanionClient -v
go test ./internal/ai -run TestExtract -v   # tenant-scope / principal
```

- Esperado: 4 archivos de test verdes. `handler_test.go` levanta sqlite in-memory y construye principal vía `platform/security/go/contextkeys`.

## Validación manual (API)

Con `COMPANION_BASE_URL` + `COMPANION_INTERNAL_JWT_SECRET` apuntando a un Companion real/stub:

1. `POST {APIBase}/ai/chat` con header `X-PROJECT-ID` válido y auth del tenant → 200 con `{chat_id, reply, ...}`.
2. `POST {APIBase}/ai/chat/stream` → respuesta `text/event-stream` con `event: start` y `event: done` (data con `reply`).
3. `GET {APIBase}/ai/chat/conversations?limit=10` → `{items:[...]}`.
4. `GET {APIBase}/ai/chat/conversations/:id` → `{id,title,messages:[{role,content,ts?}],created_at,updated_at}`.
5. Con `AI_TENANT_SCOPE=true` y un `project_id` de otro tenant → 403 `project is not available for this tenant`.
6. Sin `X-PROJECT-ID` → 400/validation.
7. Sin `COMPANION_BASE_URL`/`SECRET` al arrancar → el binary NO arranca (esperado).

## Casos borde

- `chat_id`/`task_id` huérfano (no existe en Companion) → debe reintentar sin IDs y crear conversación nueva (no 404 al usuario).
- Companion 401/403/404/409/5xx → mapeo a `domainerr` correcto (Unauthorized/Forbidden/NotFound/Conflict/Unavailable).
- `message` vacío → `message is required` (400).
- JWT expirado (TTL 5 min) → se firma uno nuevo por request, no debe verse en operación normal.

## Qué revisar en UI / API / DB / env

- **UI/API**: shapes de `chat`, `conversations`, `conversation detail` y los dos eventos SSE; el FE debe consumir `done`.
- **DB**: tabla `projects` con columnas `id`, `tenant_id`, `deleted_at` (la query CAST a TEXT debe funcionar en Postgres prod y sqlite test).
- **env**: `COMPANION_BASE_URL`, `COMPANION_INTERNAL_JWT_SECRET`, `COMPANION_INTERNAL_JWT_ISSUER` (default `ponti-backend`), `COMPANION_INTERNAL_JWT_AUDIENCE` (default `companion`), `COMPANION_INTERNAL_JWT_TTL_SEC` (300), `COMPANION_TIMEOUT_MS` (30000), `COMPANION_MAX_RETRIES` (2), `AI_TENANT_SCOPE` (false), y opcionales `NEXUS_*`.

## Qué validar en el OTRO repo (FE feature-012)

- BFF `ai.ts` / `managerChatStreamProxy` apuntan a `/ai/chat*` del BE.
- `AIAssistant.tsx` (u homólogo) procesa `event: done` y muestra `reply`; tolera la ausencia de tokens progresivos.
- El FE no asume campos que Companion no devuelve (o que el adapter rellena con defaults).

## Tests FE sugeridos (en el otro repo)

```bash
yarn test
yarn build
# e2e del flujo de chat si existe
```

## Señales de incompletitud / incompatibilidad

- `go build` falla por `undefined: authz.PrincipalFromContext` o `config.Companion` → falta dependencia previa.
- Wire compila pero el binary panica al arrancar → `ProvideCompanionClient` con config vacía (env no seteado).
- FE muestra "Asistente AI no configurado" (string del fallback dummy legacy) → quedó el usecases viejo, extracción incompleta.
- Respuestas de chat con `reply` vacío y `tokens_used` siempre 0 sin texto → revisar mapeo de Companion (no necesariamente bug: tokens_used es default).
