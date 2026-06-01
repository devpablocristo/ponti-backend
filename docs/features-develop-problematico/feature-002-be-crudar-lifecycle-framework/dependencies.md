# dependencies.md — feature-002 · be-crudar-lifecycle-framework

## Resumen direccional

- **Depende de**: el bump de dependencias Go (prometheus + 2 módulos platform).
- **Bloquea a**: feature **009** (crudar-archive-surface) y a todos los repos
  por-entidad que consumen el paquete.
- **Cross-repo**: ninguna (Solo-BE).

## Depende-de (upstream)

### Fuertes (bloqueantes — sin esto NO compila)

- **go.mod / go.sum con 3 deps nuevas**:
  - `github.com/prometheus/client_golang v1.23.2` (+ indirects:
    `prometheus/client_model v0.6.2`, `prometheus/common v0.66.1`,
    `prometheus/procfs v0.16.1`) — usado por `metrics.go`.
  - `github.com/devpablocristo/platform/observability/go v0.2.1` — el namespace
    de métricas se alinea con `observability.DefaultMetricsConfig` (mencionado en
    el doc-comment de `metrics.go`).
  - `github.com/devpablocristo/platform/persistence/gorm/go v0.1.0`.
  - Estado en develop tip (003a9b8f): **NINGUNA de las 3 está presente** en
    go.mod (verificado). `gorm.io/driver/sqlite v1.6.0` SÍ está (lo usan los
    tests). `gorm.io/gorm`, `google/uuid`, `platform/errors/go` también están.
  - Tipo: relación **intra-repo, fuerte, cierta**. Probablemente administrada
    por feature **021 build-and-deploy-config / dependency-bumps** — pero los
    bumps go-jose/x/net de #124 NO incluyen prometheus/observability/persistence;
    hay que agregarlas. Coordinar.

### Débiles

- `platform/errors/go` (`domainerr.Internal/Conflict`) — ya en develop. OK.

### Inciertas

- Estado del schema de DB destino para migraciones 232/233: depende de que
  migraciones previas (000196, las que crean `v4_report.*` y `v4_ssot.*`, y las
  columnas `deleted_at`/FK en cada tabla) estén aplicadas. En una DB al día con
  develop deberían estar; **verificar contra staging**. Incierta hasta validar.

## Bloquea-a (downstream)

### Fuertes

- **feature 009 (crudar-archive-surface)**: TODO el consumo del paquete vive
  ahí. 20 archivos `internal/*/repository.go` importan `internal/shared/lifecycle`
  (verificado con `git grep`): actor, business-parameters, campaign, category,
  class-type, commercialization, crop, customer, field, investor, labor,
  lease-type, lot, manager, project, provider, supply (+repository_movement),
  work-order-draft, work-order. Sin feature-002 mergeada, esos repos no
  compilan.
- `cmd/archive-cleanup/main.go` (feature 009/019) llama `RunArchiveCleanup`,
  `ArchiveCleanupOptions`, `ArchiveCleanupReport`, `ErrArchiveCleanup*`.
- `cmd/api/main.go` (feature 023/005) llama `lifecycle.RegisterMetrics(
  metrics.Registry(), "ponti_backend")`.

### Débiles

- Features por-entidad (010 projects, 011 campaign-dto, 018 data-integrity-admin)
  que toquen archive/restore se apoyan indirectamente en este framework vía 009.

## Archivos / tipos / config / migraciones / APIs compartidos

| recurso | compartido con | nota |
|---|---|---|
| go.mod / go.sum | feature 021 (deps), #124 (go-jose/x/net YA porteado) | tomar SOLO hunks de prometheus/observability/persistence; excluir go-jose/x/net |
| tabla `archive_batches` | feature 009 (la usa al archivar) | creada aquí (227) |
| columnas `archive_*` en ~32 tablas | feature 009, features por-entidad | creadas aquí (227/228) |
| triggers `assert_parent_active` | feature 009 (mapea SQLSTATE 23514→Conflict) | creados aquí (233) |
| `lifecycle.Policies` map | feature 009 + repos por-entidad | fuente de verdad de cascadas |
| `cmd/api/main.go` (RegisterMetrics) | feature 023/005 | NO traer aquí |
| migración 000196 (referenciada por 232) | features previas YA en develop | solo referencia textual |

## Migraciones — orden

- En develop YA están: 000229 (dashboard), 000230 (workorders_is_digital_origin),
  vía #117/#121/#124.
- Mis migraciones: 227, 228 (< 229) y 232, 233 (> 230).
- **227/228 quedan por DEBAJO de 229/230 ya aplicadas** → hazard de orden (ver
  risks.md). 232/233 quedan por encima, sin conflicto numérico.

## Recomendación de orden de merge

1. (Prerequisito) Deps Go en develop — feature 021 o bloque previo en este PR.
2. **feature-002** (este paquete + migraciones).
3. feature 009 (consumidores + endpoints + CLI + RegisterMetrics).
4. features por-entidad / 018 según necesiten.

No mergear 009 antes que 002 (no compilaría). No mergear 002 sin deps (no
compilaría).
