# notes-for-future-agent.md — feature-012 (BE)

## Resumen corto

Cutover del proxy AI de Ponti: de cliente legacy ponti-ai (Python, `X-SERVICE-KEY`) a **Companion** (`axis/companion`) con JWT interno HS256 por request. Nuevo package `internal/axis` (clientes Companion + Nexus, jwt, errores, DTOs) + un `CompanionAdapter` que implementa `usecases.ClientPort` para no cambiar el contrato del handler hacia el FE. Companion es síncrono → SSE sintético (`start`+`done`). Cliente Nexus presente pero NO cableado (gating futuro).

## Qué está en BE y qué en FE

- **BE (este flist, 16 paths)**: package `internal/axis/**`, `internal/ai/companion_adapter*.go`, `internal/ai/handler*.go`, `internal/ai/usecases/usecases*.go`, `wire/companion_providers.go`, y el **borrado** de `internal/ai/client.go`.
- **FE (feature-012 en repo web)**: `pages/admin/ai`, BFF `ai.ts` + `managerChatStreamProxy`, `AIAssistant.tsx` (consume `done`). No verificado desde este repo.

## Archivos esenciales

- `internal/axis/client.go` — el cliente Companion (Chat/List/Get).
- `internal/axis/jwt.go` — la firma JWT (claims `org_id/actor/scope`); el punto más delicado para que Companion acepte.
- `internal/ai/companion_adapter.go` — TODA la traducción de contrato vive acá (SSE sintético, fallback huérfano, marshalers que rellenan campos legacy).
- `wire/companion_providers.go` — `ProvideCompanionClient` (obligatorio) / `ProvideNexusClient` (nil opcional).

## Archivos peligrosos / mezclados

- `wire/ai_providers.go` — **NO es de este flist** (es de feature-023). En `develop` AÚN wirea el `ai.Client` legacy + `config.AI`. Hay que **reescribirlo** para Companion. No traer ciegamente desde SOURCE ni dejar el de develop.
- `wire/wire_gen.go` — generado (feature-023). **Regenerar con `wire`**, nunca editar a mano.
- `cmd/config/companion.go` + `Security.AITenantScope` — **de feature-005**, ausentes en develop. Bloqueantes.
- `internal/shared/authz/authz.go` — **de feature-001**, ausente en develop. Bloqueante (`PrincipalFromContext`).

## Decisiones ya tomadas (en el código)

- Cutover total: se borra el cliente legacy, no hay fallback dummy. AI pasa a ser obligatorio (binary no arranca sin config).
- `project_id` NO viaja a Companion; solo se valida tenant↔project local cuando `AI_TENANT_SCOPE=true`.
- SSE sintético en vez de streaming real.
- JWT lleva `scope` (string) y `scopes` (array) por compatibilidad.
- Nexus se incluye pero queda inactivo (nil si `NEXUS_BASE_URL` vacío).

## Dudas abiertas

- ¿La sección `Review` de `loadconfig.go` (Nexus Review/approvals) viene completa con feature-005?
- ¿FE feature-012 ya tolera la ausencia de streaming progresivo?
- ¿Secret `COMPANION_INTERNAL_JWT_SECRET` sincronizado con Companion en prod?
- ¿`projects` en develop tiene `tenant_id`/`deleted_at` (depende de feature-010)?

## Comandos a mirar primero

```bash
cat /tmp/flists/be-012.txt
git -C <repo> diff 0972e565..777e5f6a -- internal/ai/handler.go internal/ai/usecases/usecases.go
git -C <repo> show 777e5f6a:internal/axis/client.go | head -160
git -C <repo> show 777e5f6a:wire/ai_providers.go      # forma NUEVA (referencia para reescribir)
git -C <repo> show 003a9b8f:wire/ai_providers.go      # forma LEGACY (lo que hay en develop)
git -C <repo> show 003a9b8f:cmd/config/companion.go   # → no existe (confirma dep 005)
git -C <repo> show 003a9b8f:internal/shared/authz/authz.go  # → no existe (confirma dep 001)
```

## Errores a evitar

- NO usar `develop-problematico` (tip vacío). Usar `develop-problematico~1` / SHA `777e5f6a`.
- NO dejar `internal/ai/client.go` (legacy) ni el `wire/ai_providers.go` de develop.
- NO editar `wire/wire_gen.go` a mano.
- NO tocar `go.mod`/`go.sum` (deps ya presentes en develop).
- NO mergear el FE antes que el BE.

## Camino más seguro

1. Asegurar mergeadas/incluidas feature-001 (authz), feature-005 (config), feature-023 (wire base).
2. Traer enteros `internal/axis/**`, `internal/ai/companion_adapter*.go`, `internal/ai/handler*.go`, `internal/ai/usecases/usecases*.go`, `wire/companion_providers.go`.
3. `git rm internal/ai/client.go`.
4. Reescribir `wire/ai_providers.go` (Companion) + regenerar `wire/wire_gen.go`.
5. `go build ./... && go test ./internal/ai/... ./internal/axis/...`.
6. PR BE-first; coordinar FE feature-012 después.

## Qué PR del otro repo va antes/después

- **Antes (este repo)**: feature-001, feature-005, feature-023.
- **Después (cross-repo)**: feature-012 FE (BE-first). El FE debe ir tras el merge de este BE.
