# dependencies.md — feature-024 · openapi-and-docs (BE)

## Depende de

**Ninguna feature a nivel de bloqueo.** Es documentación. Todas las cosas que los docs *referencian* ya existen en `develop` (`003a9b8f`), verificado:

| Lo que el doc referencia | Dónde vive | ¿En develop? | Feature dueña | Fuerza |
|---|---|---|---|---|
| `make openapi` target | `Makefile` | SÍ (verificado) | 020/021 (ci/build) | débil — sólo referencia |
| `migrations_v4/000233_archived_invariant_triggers` | `migrations_v4/` | SÍ (presente en SOURCE; triggers) | feature CRUDAR de código (002/009) | débil |
| `@Router` annotation | `internal/data-integrity/handler.go` | SÍ (develop y source) | 018/data-integrity | débil |
| mapper de errores | `internal/shared/handlers/errors.go` | (referenciado por ERROR_CATALOG/ARCHITECTURE) | 002/shared | débil |
| logger setup | `cmd/api/main.go` | SÍ | 005/020/023 | débil |
| `EnsureCustomerFromActor`, `legacy_sync.go` | `internal/actor/` | (referenciado por customers-projects-lifecycle.md) | 007/008 actor-system | débil |

Como son referencias en prosa/yaml y no imports/código, NINGUNA bloquea el merge de 024. Si alguna no existiera, el doc quedaría con un enlace muerto (cosmético), no roto.

## Bloquea a

- **FE feature-024 (parcial, débil)**: el FE consume `docs/openapi/swagger.yaml` con `yarn codegen:openapi` → `src/api/generated/types.ts`. Si el FE quiere regenerar tipos, conviene que el BE 024 esté mergeado primero. Pero el spec es piloto (2 endpoints), así que el "bloqueo" es marginal.
- Nada más. Ninguna feature de código depende de estos docs.

## Clasificación de dependencias

### Fuertes
- Ninguna.

### Débiles
- Cross-repo: BE-024 publica el contrato OpenAPI que FE-024 consume. Orden recomendado BE-first, pero no obligatorio.
- Intra-repo: los docs apuntan a tooling/código de 002/009/018/019/020/021/023; todo ya en develop.

### Inciertas
- Los hunks de tooling en `README.md`/`docs/README.md` (`staging-db-2-local-db`→`reset-local-db-from-prod`, `core/*`→`platform/*`, `migrate-up`→`db-migrate-up`) solapan con 019 (be-local-tooling-db-scripts), 021 (build-and-deploy-config) y 001 (be-platform-tenancy-refactor). **Incierto** quién debe ser dueño de esos hunks. Recomendación: que 019/021 sean dueños del rename real de comandos; 024 sólo debería traer prosa de doc. Decisión humana.

## Archivos / tipos / config / migraciones / APIs compartidos

- **Archivos compartidos (partial-hunks)**: `README.md`, `docs/README.md`, `docs/ARCHITECTURE.md`. Estos son docs que también tocarían 001/019/021. Ver `file-list.md` sección Compartidos.
- **Tipos compartidos**: el `swagger.yaml` define `IntegrityReportResponse`, `IntegrityCheckDTO`, `MeContext`, `MeUser`, `MeTenant` — generados desde DTOs de 018 (data-integrity) y 008 (identity/me). Si esos DTOs cambian, el spec quedaría desactualizado (pero 024 sólo congela el snapshot).
- **Config compartida**: ninguna (no toca `.env`, `go.mod`, `Makefile`).
- **Migraciones compartidas**: ninguna en el flist.
- **APIs**: documentadas, no definidas.

## Recomendación de orden

1. **024 puede mergear de forma INDEPENDIENTE** y en cualquier momento (es docs, sin deps fuertes).
2. Para los hunks de tooling en README/docs.README: mergear **019 y 021 primero**, y en 024 **descartar** esos hunks (o no tocar README/docs.README). Evita conflictos y doble fuente de verdad.
3. Cross-repo: BE-024 antes que el `codegen` del FE-024 si se quiere regenerar tipos; de lo contrario, orden libre.
