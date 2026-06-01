# spec.md — feature-012 AI / Companion (axis) integration (BE)

- **id**: feature-012
- **slug**: ai-companion-integration
- **nombre**: AI / Companion (axis) integration
- **tipo**: feature
- **repo**: Backend Go — ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **merge**: BE-first
- **existe-en-BE**: SI (este paquete)
- **existe-en-FE**: SI (FULL-STACK, mismo feature-012 en el repo `fe=web`)
- **SOURCE de extracción**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (tip vacío / restore).
- **RANGO diff fuente-de-verdad**: `0972e565..777e5f6a`
- **rama destino**: `develop` (tip `003a9b8f`)

## Resumen

Cutover del proxy AI de Ponti: se reemplaza el cliente HTTP legacy hacia **ponti-ai** (FastAPI Python, `internal/ai/client.go` con `X-SERVICE-KEY`) por un cliente tipado contra **Companion** (servicio AI horizontal en el monorepo `axis/`), con firma de **JWT interno HS256** por request. Se agrega un paquete nuevo `internal/axis/` (clientes Companion + Nexus, JWT, errores, DTOs) y un **adapter** `internal/ai/companion_adapter.go` que implementa la interfaz `usecases.ClientPort` para que el handler HTTP no cambie de contrato hacia el FE.

## Objetivo

- Migrar el backend de Ponti del cliente AI legacy (servicio Python `ponti-ai`) al servicio horizontal **Companion** del stack `axis/`.
- Firmar JWTs internos cortos (HS256, secret compartido `COMPANION_INTERNAL_JWT_SECRET`) con claims `org_id` / `actor` / `scope` para que Companion sanitice headers anti-spoofing pero confíe en los claims firmados.
- Dejar preparado (sin activar gating) el cliente **Nexus** (`axis/nexus`, motor de decisiones allow/deny/require_approval) para futuro gating de acciones sensibles.
- Mantener el contrato HTTP hacia el FE intacto (mismas rutas `/ai/chat`, `/ai/chat/stream`, `/ai/chat/conversations[/:id]` y shapes de respuesta).

## Problema

- El cliente legacy `internal/ai/client.go` usaba `X-SERVICE-KEY` + `X-USER-ID` + `X-PROJECT-ID` contra `ponti-ai`; ese servicio ya no es el upstream.
- Companion **no tiene SSE real**: el chat es síncrono. El FE espera streaming (`/chat/stream` con evento `done`). El adapter resuelve esto componiendo un SSE sintético (`start` + `done`) sobre el response síncrono.
- Companion **no tiene noción de `project`**: hay que decidir qué hacer con `X-PROJECT-ID` (se valida tenant-scope localmente pero NO se propaga a Companion).
- Identidad/tenant: el legacy tomaba `X-USER-ID` del header; el nuevo handler resuelve el principal real del contexto (`authz.PrincipalFromContext`) y valida que el `project_id` pertenezca al tenant.

## Alcance en este repo (BE)

Paquete nuevo `internal/axis/`:
- `client.go` — `CompanionClient`: `Chat` (`POST /v1/chat`), `ListConversations` (`GET /v1/chat/conversations?limit=`), `GetConversation` (`GET /v1/chat/conversations/{id}`). Defaults de timeout/retry/TTL.
- `jwt.go` — `jwtSigner` HS256, claims `iss/aud/sub/iat/exp/org_id/actor/actor_id/scope/scopes`.
- `http.go` — `newHTTPClient` con timeouts ajustados a LLM.
- `errors.go` — `ErrNotConfigured` + `mapHTTPError` (status → `platform/errors/go/domainerr`).
- `types.go` — DTOs Companion (`ChatRequest/Response`, `Task`, `Message`, `Conversation*`, `CallContext`).
- `nexus_client.go` + `nexus_types.go` — `NexusClient` (`SubmitRequest`/`GetRequest`/`ReportResult`), DTOs allow/deny/require_approval + binding_hash. **Inactivo en runtime** (opcional, nil si `NEXUS_BASE_URL` vacío).

Paquete `internal/ai/`:
- **borra** `client.go` (cliente legacy ponti-ai).
- **agrega** `companion_adapter.go` — `CompanionAdapter` implementa `ClientPort`; rutea por `method+path`, traduce DTOs, compone SSE sintético, fallback de `chat_id`/`task_id` huérfano (retry sin IDs si Companion devuelve 404).
- **modifica** `handler.go` — `NewHandler` ahora recibe `*gorm.DB` + `aiTenantScope bool`; `extractIDs` pasa a método que resuelve principal del contexto y valida project↔tenant; firma de `UseCasesPort` agrega `tenantID`.
- **modifica** `usecases/usecases.go` — elimina el fallback dummy; `ClientPort` y métodos agregan `tenantID`.

Wire:
- **agrega** `wire/companion_providers.go` — `ProvideCompanionClient` (obligatorio), `ProvideNexusClient` (opcional/nil), `ProvideConfigNexus`, `CompanionSet`.

Tests nuevos: `internal/axis/client_test.go`, `internal/ai/companion_adapter_test.go`, `internal/ai/handler_test.go`, `internal/ai/usecases/usecases_test.go`.

## Alcance en el OTRO repo (FE — feature-012 fullstack)

Según la nota de la feature (no verificado en este repo): `pages/admin/ai`, BFF `ai.ts` + `managerChatStreamProxy`, componente `AIAssistant.tsx` que consume el evento SSE `done`. El contrato que el FE espera está embebido en el adapter (`marshalCompanionChat`/`marshalCompanionDetail` y la composición SSE `start`+`done`).

## Fuera de alcance

- Streaming token-por-token real (Companion no lo soporta; trade-off documentado en `companion_adapter.go`).
- Propagación de `project_id` a Companion (no tiene noción nativa de project).
- Gating de acciones vía Nexus (cliente presente pero **no cableado** a ningún flujo; MVP solo-chat).
- Migraciones de DB: ninguna en este feature (solo un `SELECT 1` de validación tenant↔project sobre la tabla existente `projects`).

## Comportamiento esperado

1. FE llama `POST {APIBase}/ai/chat` con `{message, chat_id?|task_id?, channel?}`.
2. Handler resuelve principal (`authz.PrincipalFromContext`), valida `X-PROJECT-ID` no vacío y (si `AI_TENANT_SCOPE=true`) que el project pertenezca al tenant.
3. UseCases → `CompanionAdapter.Do` → `CompanionClient.Chat` firmando JWT con `org_id=tenant`, `actor=principal.Actor`, scopes `companion:tasks:read|write`.
4. Si Companion responde 404 por `chat_id`/`task_id` huérfano → retry sin IDs (conversación nueva).
5. `/ai/chat/stream` → adapter ejecuta chat síncrono y emite SSE `start`+`done`.
6. Listado/detalle → mapeo de `agent_conversations` al shape del FE.

## Estado en dp~1 (SHA 777e5f6a)

Completo y con tests a nivel de paquete. Compila contra el árbol de `777e5f6a` (que ya tiene `internal/shared/authz`, `cmd/config/companion.go`, `Security.AITenantScope`, `wire/ai_providers.go` reescrito y `wire/wire_gen.go` regenerado). En `develop` (`003a9b8f`) esas piezas **NO existen todavía** → ver dependencias.

## Criterios de aceptación

- [ ] `internal/axis/` y `internal/ai/companion_adapter.go` presentes; `internal/ai/client.go` eliminado.
- [ ] `wire/companion_providers.go` presente; `wire/ai_providers.go` reescrito para `CompanionAdapter` (NO legacy `ai.Client`); `wire/wire_gen.go` regenerado con `wire`.
- [ ] `cmd/config/companion.go` (structs `Companion` + `Nexus`) y `Security.AITenantScope` presentes (vienen de feature-005).
- [ ] `internal/shared/authz.PrincipalFromContext` presente (viene de feature-001).
- [ ] `go build ./...` y `go test ./internal/ai/... ./internal/axis/...` verdes.
- [ ] FE feature-012 mergeado o coordinado (BE-first): el contrato `done`/shapes coincide.

## Endpoints / modelos / UI / DB / tests afectados

- **Endpoints (sin cambio de ruta)**: `POST /ai/chat`, `POST /ai/chat/stream`, `GET /ai/chat/conversations`, `GET /ai/chat/conversations/:conversation_id`.
- **Upstream Companion**: `POST /v1/chat`, `GET /v1/chat/conversations?limit=`, `GET /v1/chat/conversations/{id}`.
- **Upstream Nexus (no cableado)**: `POST /v1/requests`, `GET /v1/requests/{id}`, `POST /v1/requests/{id}/result`.
- **Modelos/DTOs**: `axis.ChatRequest/ChatResponse/ChatBlock/ChatToolCall/Task/Message/Conversation*/CallContext`; `axis.Nexus*`.
- **DB**: ninguna migración; un `SELECT 1 FROM projects WHERE id=? AND tenant_id=? AND deleted_at IS NULL` (validación tenant-scope).
- **Tests**: `internal/axis/client_test.go`, `internal/ai/companion_adapter_test.go`, `internal/ai/handler_test.go`, `internal/ai/usecases/usecases_test.go`.

## Dependencias

- **Intra-repo (fuertes)**: feature-005 (config: `cmd/config/companion.go`, `Security.AITenantScope`); feature-001 (`internal/shared/authz.PrincipalFromContext`); feature-023 (`wire/ai_providers.go` + `wire/wire_gen.go`).
- **Cross-repo**: feature-012 FE (contrato chat/SSE/conversations).
- **Libs (ya en go.mod de develop)**: `platform/http/go/httpclient`, `platform/errors/go/domainerr`, `platform/security/go/contextkeys`, `golang-jwt/jwt/v5`. Confianza ALTA (verificado en `go.mod` de `003a9b8f`).

## Riesgos

- **Funcional**: pérdida de streaming token-por-token (SSE sintético). FE debe seguir leyendo `done`.
- **Técnico**: `wire/ai_providers.go` y `wire/wire_gen.go` son compartidos con feature-023; extraer solo este feature deja el árbol sin compilar si esos no se actualizan a la vez.
- **Cross-repo**: mergear BE sin FE deja un contrato nuevo sin cliente actualizado (BE-first mitiga: el shape es retrocompatible con el FE legacy salvo streaming).
- **Operativo**: post-cutover `COMPANION_BASE_URL` + `COMPANION_INTERNAL_JWT_SECRET` pasan a ser **obligatorios** (el binary no arranca sin ellos). Antes el AI era opcional (fallback dummy).

## DECISIÓN recomendada

**Arreglar dependencias antes de extraer** (no extraer tal cual aislado). El feature es coherente y testeado, pero su compilación depende de feature-005, feature-001 y feature-023. Orden sugerido: mergear primero 005 (config) + 001 (authz) + 023 (wire-di), y en el PR de 012 traer `internal/axis/**`, `internal/ai/**` y `wire/companion_providers.go` enteros, reescribir `wire/ai_providers.go` y regenerar `wire/wire_gen.go`. No partir en subfeatures: el cutover es atómico (borrar legacy + agregar adapter deben ir juntos).
