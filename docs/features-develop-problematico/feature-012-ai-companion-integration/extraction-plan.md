# extraction-plan.md — feature-012 (BE)

- **repo**: ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE de extracción**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip vacío).
- **rama sugerida**: `pr/feature-012-ai-companion-integration-be`
- **merge**: BE-first (este repo va antes que el FE feature-012).

## PR title

`feat(be): cutover AI a Companion (axis) — cliente JWT + adapter + Nexus client (feature-012)`

## PR description (borrador)

> Reemplaza el cliente AI legacy (ponti-ai, `internal/ai/client.go`, `X-SERVICE-KEY`) por un cliente tipado contra **Companion** (`axis/companion`) con firma de JWT interno HS256 por request (`org_id`/`actor`/`scope`). Nuevo package `internal/axis` (clientes Companion + Nexus, jwt, errores, DTOs) y un `CompanionAdapter` que implementa `usecases.ClientPort`, así el handler/usecases no cambian de contrato hacia el FE.
>
> - El chat de Companion es síncrono → el adapter compone SSE sintético (`start`+`done`) para mantener el contrato del FE.
> - `project_id` NO se propaga a Companion (no tiene noción de project); se valida tenant↔project localmente cuando `AI_TENANT_SCOPE=true`.
> - Cliente **Nexus** presente pero NO cableado (MVP solo-chat; gating futuro).
> - Post-cutover `COMPANION_BASE_URL` + `COMPANION_INTERNAL_JWT_SECRET` son obligatorios (sin fallback dummy).
>
> Depende de: feature-005 (config Companion/Nexus + AITenantScope), feature-001 (`shared/authz`), feature-023 (`wire/ai_providers.go` + `wire_gen.go`). Coordina con FE feature-012.

## Dependencias previas (mergear o incluir ANTES)

1. **feature-005** (be-config-modularization): `cmd/config/companion.go` (structs `Companion`+`Nexus`), `Security.AITenantScope`, secciones en `cmd/config/loadconfig.go`. **BLOQUEANTE.**
2. **feature-001** (be-platform-tenancy-refactor): `internal/shared/authz.PrincipalFromContext`. **BLOQUEANTE.**
3. **feature-023** (be-wire-di): `wire/ai_providers.go` (reescrito) + `wire/wire_gen.go` (regenerado). **BLOQUEANTE** para compilar.

Si 005/001/023 no están mergeadas aún, este PR debe incluir esos hunks puntuales o coordinarse para ir después de ellas. Recomendado: ir **después** de 001+005+023.

## Pasos ordenados

1. Crear rama desde develop.
2. Traer **enteros** desde SOURCE: todo `internal/axis/**`, `internal/ai/companion_adapter.go`, `internal/ai/companion_adapter_test.go`, `internal/ai/handler.go`, `internal/ai/handler_test.go`, `internal/ai/usecases/usecases.go`, `internal/ai/usecases/usecases_test.go`, `wire/companion_providers.go`.
3. **Eliminar** `internal/ai/client.go` (cutover).
4. Verificar que las dependencias estén presentes en el árbol (ver checklist). Si faltan, traerlas de su feature.
5. **Reescribir `wire/ai_providers.go`** (NO traer el de develop): reemplazar `ProvideAIClient`+`config.AI` por `ProvideAIUseCases(companionClient)`→`ai.NewCompanionAdapter`, ampliar `ProvideAIHandler(...,repo,appCfg)` con `repo.Client()` + `appCfg.Security.AITenantScope`, y agregar `ProvideCompanionClient`+`ProvideNexusClient` al `AISet`.
6. **Regenerar `wire/wire_gen.go`** con `wire` (no editar a mano).
7. `go build ./...`, `go vet ./...`, `go test ./internal/ai/... ./internal/axis/...`.
8. `git diff --check` (whitespace/conflict markers).

## Archivos enteros vs parciales

- **Enteros (whole-file)**: `internal/axis/*.go` (8), `internal/ai/companion_adapter.go`, los 4 tests nuevos, `internal/ai/handler.go`, `internal/ai/usecases/usecases.go`, `wire/companion_providers.go`. El diff de handler.go y usecases.go es 100% de este feature → traer enteros es más seguro que parciales.
- **Manual / regenerado**: `wire/ai_providers.go` (reescribir), `wire/wire_gen.go` (regenerar con `wire`).
- **Borrado**: `internal/ai/client.go`.

## Migraciones / tests a incluir

- **Migraciones**: ninguna. (Solo una query `SELECT 1 FROM projects ...` en runtime.)
- **Tests**: incluir los 4 archivos `*_test.go` listados. El `handler_test.go` usa `gorm.io/driver/sqlite` (driver de test) y `platform/security/go/contextkeys` — confirmar que el driver sqlite esté disponible (ya en go.sum si otros handler tests lo usan).

## Coordinación con el otro repo (FE feature-012)

- **Orden**: BE-first. Mergear este PR primero; el shape de respuesta es retrocompatible con el FE legacy salvo la ausencia de streaming token-por-token.
- El FE feature-012 (BFF `ai.ts` + `managerChatStreamProxy`, `AIAssistant.tsx`) debe seguir leyendo el evento SSE `done`. Coordinar para que el FE no asuma tokens progresivos.

## Comandos git SUGERIDOS (para un humano; este agente NO los ejecuta)

```bash
git checkout develop
git checkout -b pr/feature-012-ai-companion-integration-be

# Traer enteros desde SOURCE (NUNCA develop-problematico, usar el SHA o ~1)
git checkout develop-problematico~1 -- \
  internal/axis/client.go internal/axis/client_test.go internal/axis/errors.go \
  internal/axis/http.go internal/axis/jwt.go internal/axis/types.go \
  internal/axis/nexus_client.go internal/axis/nexus_types.go \
  internal/ai/companion_adapter.go internal/ai/companion_adapter_test.go \
  internal/ai/handler.go internal/ai/handler_test.go \
  internal/ai/usecases/usecases.go internal/ai/usecases/usecases_test.go \
  wire/companion_providers.go

# Eliminar el cliente legacy (cutover)
git rm internal/ai/client.go

# Reescribir a mano wire/ai_providers.go (ver paso 5) y regenerar wire:
#   cd wire && go run github.com/google/wire/cmd/wire   (o `wire` si está instalado)

# Para piezas dependientes que estén mezcladas con otra feature, usar parcial:
#   git restore -p --source=develop-problematico~1 -- cmd/config/loadconfig.go
#   git restore -p --source=develop-problematico~1 -- cmd/config/security.go

git diff --check
go build ./... && go vet ./... && go test ./internal/ai/... ./internal/axis/...
```

> Recordá: el agente solo SUGIERE estos comandos. Verificá cada `restore -p` hunk antes de aceptarlo.

## Qué NO traer

- `go.mod` / `go.sum` (las libs ya están en develop).
- El `wire/ai_providers.go` tal cual de develop (es legacy) ni el de SOURCE sin entender la base de feature-023.
- Nada de las features ya DONE.

## Qué podría romperse

- **Compilación**: si falta `internal/shared/authz`, `cmd/config/companion.go`, `Security.AITenantScope` o el `wire_gen.go` regenerado → no compila.
- **Runtime**: si el binary arranca sin `COMPANION_BASE_URL`/`COMPANION_INTERNAL_JWT_SECRET`, `ProvideCompanionClient` retorna error y el binary **no arranca** (intencional). Asegurar env/secret en deploy.
- **wire**: olvidar regenerar `wire_gen.go` deja la firma vieja de `ProvideAIHandler` → mismatch de tipos.

## Cómo detectar extracción incompleta

- `grep -rn "ai.Client\b\|ProvideAIClient\|config.AI\b\|X-SERVICE-KEY" internal/ai wire cmd/config` → si aparece, quedó código legacy.
- `grep -rn "extractIDs(c)" internal/ai` con la firma vieja (2 returns) → no se actualizó el handler.
- `go build ./...` falla por símbolos `authz`/`config.Companion` → falta dependencia previa.

## Qué validar antes del PR

- `go build ./... && go test ./internal/ai/... ./internal/axis/...` verde.
- `wire/wire_gen.go` regenerado (no editado a mano), `git diff` coherente con `ai_providers.go`.
- Env de ejemplo (`.env.example` si existe) documenta `COMPANION_BASE_URL`, `COMPANION_INTERNAL_JWT_SECRET`, `AI_TENANT_SCOPE`, `NEXUS_BASE_URL` (opcional).

## Qué hacer después de mergear

- Coordinar el merge del FE feature-012.
- Validar en staging un round-trip de chat real contra Companion (JWT aceptado, `done` recibido por el FE).
- Confirmar que el secret `COMPANION_INTERNAL_JWT_SECRET` esté sincronizado (Secret Manager GCP en prod) con el que valida Companion.
