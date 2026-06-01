# file-list.md — feature-002 · be-crudar-lifecycle-framework

Todos los paths salen de `/tmp/flists/be-002.txt`. Status: A=created.
Confianza = qué tan seguro estoy de que el archivo pertenece exclusivamente a
esta feature y de la recomendación de extracción.

## Propios (núcleo de la feature — extraer whole-file)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/shared/lifecycle/lifecycle.go | A | go/core | tipos `ArchiveBatch`/`Cause`/`RowState`/`ActiveRef` + helpers archive/restore/require | whole-file | archivo nuevo, no existe en develop | bajo (compila si están las deps) | alta |
| internal/shared/lifecycle/policy.go | A | go/data | mapa `Policies` declarativo por entidad (~24) | whole-file | nuevo | bajo | alta |
| internal/shared/lifecycle/cascade.go | A | go/core | `RunCascadeArchive`/`RunCascadeRestore`/`WouldOrphanActiveChildren` | whole-file | nuevo | bajo | alta |
| internal/shared/lifecycle/archive_cleanup.go | A | go/job | `RunArchiveCleanup` + reglas IA-1..IA-8 + reportes | whole-file | nuevo | bajo | alta |
| internal/shared/lifecycle/metrics.go | A | go/observability | `RegisterMetrics` + counter prometheus | whole-file | nuevo | medio (dep prometheus no está en develop) | alta |
| internal/shared/lifecycle/cascade_test.go | A | go/test | tests de cascade (sqlite in-memory) | whole-file | nuevo, no necesita DB | bajo | alta |
| internal/shared/lifecycle/archive_cleanup_test.go | A | go/test | tests del job de cleanup | whole-file | nuevo | bajo | alta |
| internal/shared/lifecycle/invariant_e2e_test.go | A | go/test | e2e del invariante (sqlite in-memory, seed customers/projects/fields/lots) | whole-file | nuevo, autocontenido | bajo | alta |
| internal/shared/lifecycle/metrics_test.go | A | go/test | test del counter | whole-file | nuevo | bajo | alta |

## Migraciones (extraer whole-file, pero ver riesgo de ORDEN)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| migrations_v4/000227_crudar_archive_batches.up.sql | A | sql/up | crea `archive_batches` + columnas causa/FK/índice en 14 tablas; actors.archived_at→deleted_at; índice parcial ux_customers_tenant_actor_id | whole-file | nuevo, idempotente | ALTO (queda bajo 229/230 ya en develop) | alta |
| migrations_v4/000227_crudar_archive_batches.down.sql | A | sql/down | revierte 227 (drop columnas/FK/índices/tabla; restaura ux index no-parcial) | whole-file | par de la up | alto | alta |
| migrations_v4/000228_crudar_remaining_archive_metadata.up.sql | A | sql/up | columnas causa/FK/índice en 18 tablas restantes | whole-file | nuevo, idempotente | ALTO (orden) | alta |
| migrations_v4/000228_crudar_remaining_archive_metadata.down.sql | A | sql/down | revierte 228 | whole-file | par | alto | alta |
| migrations_v4/000232_document_archived_inclusion_in_reports.up.sql | A | sql/up | SOLO COMMENT ON vistas/funciones report+ssot | whole-file | nuevo, no muta estructura | MEDIO (requiere que las vistas/fns existan; dependen de 196/otras migr.) | media |
| migrations_v4/000232_document_archived_inclusion_in_reports.down.sql | A | sql/down | COMMENT ... IS NULL | whole-file | par | bajo | alta |
| migrations_v4/000233_archived_invariant_triggers.up.sql | A | sql/up | función `assert_parent_active` + 14 triggers BEFORE INSERT/UPDATE | whole-file | nuevo | MEDIO-ALTO (requiere data limpia; las tablas deben tener deleted_at + las FK) | media |
| migrations_v4/000233_archived_invariant_triggers.down.sql | A | sql/down | drop triggers + función | whole-file | par | bajo | alta |

## Compartidos (partial-hunks) — NO están en mi flist pero MI código los toca

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| go.mod | M (otra feature) | deps | agrega `prometheus/client_golang`, `platform/observability/go`, `platform/persistence/gorm/go` | partial-hunks / coordinar con feat 021 | mi paquete NO compila sin estas deps; pero el archivo lo administra dependency-bumps | ALTO | alta |
| go.sum | M (otra feature) | deps | checksums de lo anterior | partial-hunks | idem | alto | alta |

## Requeridos por dependencia (de OTRAS features — NO traer aquí, solo coordinar)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| internal/customer/repository.go | M | go | consume `lifecycle.CreateArchiveBatch/RunCascadeArchive/...` | do-not-extract-yet | pertenece a feature 009 / customer | n/a | alta |
| internal/project/repository.go | M | go | consume lifecycle | do-not-extract-yet | feature 009/010 | n/a | alta |
| internal/work-order/repository.go | M | go | consume lifecycle | do-not-extract-yet | feature 009 | n/a | alta |
| internal/field|lot|labor|supply|... /repository.go (20 archivos) | A/M | go | consumen lifecycle | do-not-extract-yet | features 009 + por-entidad | n/a | alta |
| cmd/api/main.go | M | go/boot | `lifecycle.RegisterMetrics(metrics.Registry(),"ponti_backend")` | do-not-extract-yet | compartido (feat 023/005) | n/a | alta |
| cmd/archive-cleanup/main.go | A | go/cli | `lifecycle.RunArchiveCleanup(...)` | do-not-extract-yet | no existe en develop; feature 009/019 | n/a | alta |

## Dudosos

| path | status | tipo | nota |
|---|---|---|---|
| migrations_v4/000232 (COMMENT) | A | sql | Funcionalmente trivial; el riesgo es que las vistas/funciones comentadas existan en la DB destino. Si la migración 196 y las que crean `v4_report.workorder_list` / `v4_ssot.*` no están aplicadas, el COMMENT falla. Confianza media. |
| migrations_v4/000233 (triggers) | A | sql | Depende de que TODAS las tablas referenciadas tengan `deleted_at` y las FK existan en la DB destino. Si una tabla no existe el `CREATE TRIGGER` falla (no usa guardas `to_regclass` como 227/228). Confianza media. |

## NO traer todavía

- go.mod / go.sum: deben venir del PR de dependencias (feature 021) o agregarse
  explícitamente como prerequisito; no son "propios" de feature-002 aunque la
  bloqueen.
- Cualquier `internal/*/repository.go`, `cmd/*`: son de features 009/010/023.

## Nota sobre lo YA PORTEADO

- 000229/000230 (dashboard/lot-metrics, workorders_is_digital_origin) YA están
  en develop (#117/#121/#124). NO son míos, pero condicionan el ORDEN de mis
  227/228 (ver risks.md).
