# dependencies.md — feature-012 (BE)

## Depende-de (intra-repo)

| feature | fuerza | qué aporta | archivos/símbolos | verificado |
|---|---|---|---|---|
| 005 be-config-modularization | **fuerte** | structs de config Companion/Nexus + flag tenant-scope | `cmd/config/companion.go` (`config.Companion`, `config.Nexus`), `cmd/config/security.go` (`Security.AITenantScope`, env `AI_TENANT_SCOPE`), secciones en `cmd/config/loadconfig.go` | SI — ausentes en develop `003a9b8f` (`git show 003a9b8f:cmd/config/companion.go` → no existe) |
| 001 be-platform-tenancy-refactor | **fuerte** | resolución de principal/tenant del contexto | `internal/shared/authz.PrincipalFromContext` (usado en `handler.go` y `handler_test.go`) | SI — ausente en develop (`internal/shared/authz/authz.go` no existe en `003a9b8f`); presente en SOURCE; figura en `/tmp/flists/be-001.txt` |
| 023 be-wire-di | **fuerte (compilación)** | provider set AI + salida wire generada | `wire/ai_providers.go` (en develop wirea `ai.Client` legacy + `config.AI`), `wire/wire_gen.go` (firma de `ProvideAIHandler`) | SI — `wire/ai_providers.go` existe en develop en forma legacy; figura en `/tmp/flists/be-023.txt`. `companion_providers.go` solo en be-012 |

## Depende-de (libs ya presentes en develop — NO bloqueantes)

| lib | uso | verificado en `go.mod` de develop |
|---|---|---|
| `github.com/devpablocristo/platform/http/go/httpclient` | `Caller` en axis clients | SI (`platform/http/go v0.1.0`) |
| `github.com/devpablocristo/platform/errors/go/domainerr` | mapeo de errores | SI (`platform/errors/go v0.2.0`) |
| `github.com/devpablocristo/platform/security/go/contextkeys` | claves de contexto en `handler_test.go` | SI (`platform/security/go v0.2.2`) |
| `github.com/golang-jwt/jwt/v5` | firma HS256 | SI (`v5.3.1`) |
| `gorm.io/gorm` + `gorm.io/driver/sqlite` (test) | validación tenant-scope + handler_test | gorm SI; sqlite driver confirmar en go.sum |

## Bloquea-a (qué depende de este feature)

| consumidor | relación | nota |
|---|---|---|
| FE feature-012 | **fuerte cross-repo** | el FE consume el contrato `/ai/chat*` + SSE `done`; BE-first |
| feature-023 (wire) | acoplamiento mutuo | el `wire/ai_providers.go` final necesita los símbolos `axis.CompanionClient` + `ai.NewCompanionAdapter` de este feature; este feature necesita el set regenerado de 023. Coordinar en el mismo PR si van juntas |

## Fuertes / débiles / inciertas

- **Fuertes**: 005, 001, 023 (sin ellas no compila). FE feature-012 (sin él, contrato sin cliente actualizado).
- **Débiles**: cliente Nexus → no cableado; depende de `config.Nexus` solo para construirse opcional (nil si vacío). No bloquea runtime.
- **Inciertas**: sección `Review` en `loadconfig.go` (Nexus Review/approvals) referenciada por config root pero fuera de tu flist — confianza baja sobre si la trae 005 completa. Verificar.

## Archivos / tipos / config / migraciones / APIs compartidos

- **Compartidos (coordinación)**: `wire/ai_providers.go` (feature-023), `wire/wire_gen.go` (feature-023), `cmd/config/loadconfig.go` (feature-005).
- **Tipos compartidos**: `config.Companion`, `config.Nexus`, `config.Security.AITenantScope` (definidos en 005, consumidos aquí).
- **APIs upstream**: Companion `axis/companion/openapi.yaml` (no en este repo); Nexus `axis/nexus/openapi.yaml` (no en este repo). Los DTOs en `internal/axis/types.go` y `nexus_types.go` deben matchear esos contratos.
- **Migraciones**: ninguna.

## Recomendación de orden

1. feature-001 (authz) →
2. feature-005 (config) →
3. feature-023 (wire-di) →
4. **feature-012 BE** (este) →
5. feature-012 FE (cross-repo).

Si 023 y 012 se solapan en `wire/ai_providers.go`, mergearlas en orden estricto (023 primero con su forma legacy o coordinar el archivo final en el PR de 012). Lo más limpio: que 023 deje el `AISet` legacy y que 012 lo reescriba para Companion en su propio PR.
