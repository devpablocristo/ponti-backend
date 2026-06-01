# file-list.md — feature-012 (BE)

Fuente: `cat /tmp/flists/be-012.txt` (16 paths). Diff fuente-de-verdad: `0972e565..777e5f6a`.
Leyenda extracción: `whole-file` = traer el archivo entero desde SOURCE | `partial-hunks` = solo algunos hunks (archivo compartido) | `manual-port` = reescribir a mano / regenerar | `do-not-extract-yet` = no traer todavía.

## Propios (núcleo del feature — traer enteros)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/axis/client.go | A | cliente HTTP | `CompanionClient` (Chat/List/Get) + firma JWT | whole-file | archivo nuevo, exclusivo del feature | bajo | alta |
| internal/axis/client_test.go | A | test | tests del CompanionClient (httptest + JWT) | whole-file | nuevo, exclusivo | bajo | alta |
| internal/axis/errors.go | A | errores | `ErrNotConfigured` + `mapHTTPError`→domainerr | whole-file | nuevo, exclusivo | bajo | alta |
| internal/axis/http.go | A | infra http | `newHTTPClient` timeouts LLM | whole-file | nuevo, exclusivo | bajo | alta |
| internal/axis/jwt.go | A | seguridad | `jwtSigner` HS256 + claims | whole-file | nuevo, exclusivo | bajo | alta |
| internal/axis/types.go | A | DTOs | DTOs Companion + `CallContext` | whole-file | nuevo, exclusivo | bajo | alta |
| internal/axis/nexus_client.go | A | cliente HTTP | `NexusClient` (gating, INACTIVO) | whole-file | nuevo; no cableado pero el package debe compilar entero | bajo | alta |
| internal/axis/nexus_types.go | A | DTOs | DTOs Nexus (allow/deny/approval) | whole-file | nuevo; compañero de nexus_client | bajo | alta |
| internal/ai/companion_adapter.go | A | adapter | implementa `ClientPort`; SSE sintético; fallback huérfano | whole-file | nuevo, exclusivo | medio | alta |
| internal/ai/companion_adapter_test.go | A | test | tests del adapter (httptest) | whole-file | nuevo, exclusivo | bajo | alta |
| internal/ai/handler_test.go | A | test | tests del handler (sqlite in-mem + authz ctx) | whole-file | nuevo; depende de `shared/authz` y `platform/security/go/contextkeys` | medio | alta |
| internal/ai/usecases/usecases_test.go | A | test | tests usecases (fakeClient con tenantID) | whole-file | nuevo, exclusivo | bajo | alta |
| internal/ai/client.go | D | cliente legacy | cliente ponti-ai a ELIMINAR | manual-port | borrado intencional (cutover); en develop AÚN existe | medio | alta |

## Compartidos (partial-hunks / coordinación con otra feature)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/ai/handler.go | M | handler HTTP | nueva firma `NewHandler(...,db,aiTenantScope)`; `extractIDs`→método; valida project↔tenant; `tenantID` en UseCasesPort | whole-file (recomendado) | el diff completo es de este feature; pero el archivo NO existe en su versión nueva en develop, depende de `shared/authz` | medio | alta |
| internal/ai/usecases/usecases.go | M | usecases | quita fallback dummy; agrega `tenantID` a `ClientPort` y métodos | whole-file (recomendado) | el diff completo es de este feature | bajo | alta |
| wire/companion_providers.go | A | wire providers | `ProvideCompanionClient/NexusClient/ConfigNexus` + `CompanionSet` | whole-file | nuevo y exclusivo, PERO depende de `config.Companion`/`config.Nexus` (feature-005) | medio | alta |

## Requeridos por dependencia (NO están en tu flist; verificar/portear desde su feature)

| path | dueño | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|
| cmd/config/companion.go | feature-005 | structs `Companion` + `Nexus` (envconfig) | do-not-extract-yet (lo trae 005) | sin esto `companion_providers.go` no compila | alto | alta |
| cmd/config/security.go (campo `AITenantScope`) | feature-005 | `AI_TENANT_SCOPE bool` | do-not-extract-yet (lo trae 005) | usado por `ProvideAIHandler` | alto | alta |
| cmd/config/loadconfig.go (campos `Companion`/`Nexus`/`Review`) | feature-005 | wiring del config root | do-not-extract-yet (lo trae 005) | sin esto `config.Config` no tiene las secciones | alto | media |
| internal/shared/authz/authz.go (`PrincipalFromContext`) | feature-001 | resolución de principal/tenant | do-not-extract-yet (lo trae 001) | usado en `handler.go` y `handler_test.go` | alto | alta |
| wire/ai_providers.go | feature-023 | provider set AI (en develop AÚN wirea `ai.Client` legacy) | manual-port (reescribir) | hay que reemplazar `ProvideAIClient`/`config.AI` por `CompanionAdapter` + agregar `CompanionClient`/`NexusClient` | alto | alta |
| wire/wire_gen.go | feature-023 | salida generada de wire | manual-port (regenerar con `wire`) | refleja la nueva firma de `ProvideAIHandler` | alto | media |

## Dudosos

| path | nota | extracción | confianza |
|---|---|---|---|
| internal/axis/nexus_*.go | el código existe pero NO está cableado a ningún flujo de runtime; entra solo para que el package compile entero | whole-file | media (sobre si conviene incluir lo no usado: SÍ, mantiene el package coherente) |
| cmd/config/companion.go (sección `Review`) | `loadconfig.go` referencia `Review Review` (Nexus Review/approvals) que NO está en tu flist; confirmar que viene con 005 | do-not-extract-yet | baja |

## NO traer todavía (DONE / fuera de este feature)

- Nada de tu flist está marcado DONE. Ningún path de feature-012 BE coincide con lo ya porteado (lot-metrics, tentative-prices, dependency-bumps, table-select-filters, reports-dark-mode).
- `go.mod` / `go.sum`: NO tocar en este PR — las libs necesarias (`platform/http/go`, `platform/errors/go`, `platform/security/go`, `golang-jwt/jwt/v5`) YA están en `develop` (`003a9b8f`). Verificado.
