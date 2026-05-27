# Ponti BE — guía rápida

Backend Go (Gin + GORM + Postgres). 31 módulos en `internal/`, 230+ migraciones, ~110 archivos test.

## Stack y dependencias

- **Go 1.26+**, Gin, GORM 1.31, Postgres (pgvector para AI).
- Libs compartidas en `github.com/devpablocristo/platform/*` — fuente única. **NO usar `core/` ni `modules/`** (deprecados).
- Observabilidad: `slog` JSON + Prometheus RED metrics + OTel tracing (ver [docs/OBSERVABILITY.md](docs/OBSERVABILITY.md)).

## Arquitectura

Hexagonal pragmática. Layout canónico por módulo (`internal/<module>/`):

```
<module>/
├── handler.go          # HTTP routing + DTO marshal; recibe UseCasesPort
├── handler/dto/        # Request/response shapes con json: tags
├── usecases.go         # Orquestación, transacciones; recibe RepositoryPort
├── usecases/domain/    # Entities sin tags (json/gorm); métodos + invariantes
├── repository.go       # Implementación de RepositoryPort
└── repository/models/  # GORM models (con column: tags); ToDomain() mappers
```

**Reglas duras** (enforced en code review):
- Domain (`usecases/domain/*.go`) NO importa `gorm`, `gin`, `json`. Solo `time`, otros domain types, `domainerr`.
- Handler NO recibe `*gorm.DB`. Solo `UseCasesPort`.
- Transacciones se abren en usecase, no en handler ni repo independiente.
- Errores: dominio retorna `domainerr.*`; handler los mapea a HTTP via mapper centralizado.

## Convenciones críticas

### Lifecycle (CRUDAR)
- `archived = no existe`: si `deleted_at IS NOT NULL`, la entidad NO se puede usar, seleccionar ni referenciar.
- `assertXReferencesActive` antes de Create/Update por entidad (template: `internal/work-order/repository.go:585`).
- Restore valida parent activo (`lifecycle.RequireActive`).
- Helpers: `lifecycle.RunCascadeArchive`, `RunCascadeRestore`, `Policies` map.
- Counter custom: `crudar_rejected_archived_ref_total{table}` cuenta rechazos.
- Ver [docs/crudar-lifecycle.md](docs/crudar-lifecycle.md).

### Tenancy
- `authz.MaybeTenantScope(query, table)` en toda query tenant-scoped.
- `authz.TenantFromContext(ctx)` para leer tenant_id del request.
- 310 callers consistentes.

### Errores
- Catálogo completo en [docs/ERROR_CATALOG.md](docs/ERROR_CATALOG.md).
- Mensajes en inglés, lower-case, presente, breves.
- FE traduce con `translateBackendError` (65+ patterns).

### Tests
- Repos usan `gormtest` con DB SQLite/Postgres real.
- Usecases usan fakes de `RepositoryPort` (ver `internal/admin/usecases_test.go` como template).
- Cobertura por módulo: `go test ./internal/<x>/... -cover`.

## Cómo correr el proyecto

```bash
# Levantar stack completo (backend + db + bff + ui)
make up-ponti-local

# Solo backend
cd core && docker compose up -d

# Tests
go test ./...
go test ./... -race  # race detector
go vet ./...
golangci-lint run

# Observabilidad opcional (ver spans en stdout)
echo "OTEL_EXPORTER=stdout" >> .env
docker compose up -d --force-recreate ponti-api
```

## Endpoints clave

| Path | Para qué |
|---|---|
| `/api/v1/...` | API de negocio (~50 endpoints) |
| `/api/v1/ping`, `/api/v1/healthz`, `/api/v1/version` | Liveness/version |
| `/observability/metrics` | Prometheus (RED metrics + Go runtime + crudar counter) |
| `/api/v1/me/context` | Bootstrap FE: user + tenants + roles + permisos |

## Estado del refactor sistémico (rolling)

El [plan vivo](../../.claude/plans/...) tiene el estado por bloque. Highlights:
- **Bloque A** (críticos): observabilidad end-to-end ✅.
- **Bloque B** (altos): admin hex arch ✅, dominio puro ✅, CORS+rate-limit ✅, ESLint estricto ✅, skeletons FE ✅, coverage admin +10pp ✅.
- **Bloque C** (medios): domain methods piloto ✅, naming reducer FE ✅, catálogo errores ✅, docker-compose legacy mount removido ✅, lockfile cleanup ✅.

## Reglas de etiqueta para Claude

- **NO tocar `git`** salvo autorización explícita por turno. Listar comandos para que el usuario los ejecute.
- **NO usar `core/` o `modules/`**: deprecados. Solo `platform/`.
- Cambios en `platform/*`: bump VERSION + commit + tag + push, luego `go get` en ponti.
- Importadores (CSV/XLSX): nunca pisan registros — rechazar duplicados.
- Errores BE en inglés; FE traduce.
- "Trabajar sin parar a preguntar cuando se pueda" — pero pausar para decisiones arquitectónicas grandes.
