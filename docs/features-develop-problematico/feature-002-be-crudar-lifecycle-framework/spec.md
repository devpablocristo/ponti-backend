# spec.md — feature-002 · be-crudar-lifecycle-framework

- **id**: feature-002
- **nombre**: CRUDAR lifecycle framework
- **tipo**: refactor (infra de dominio compartida + migraciones)
- **repo**: Backend Go (ponti-backend) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **existe-en-FE/BE**: Solo BE. En FE NO hay carpeta para esta feature (mencionar en el cross-repo-map del FE como "sin cambios FE").
- **rango fuente-de-verdad**: `0972e565..777e5f6a`
- **SOURCE de extracción**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **rama destino**: `develop` (tip `003a9b8f`).

## Resumen

Introduce un paquete compartido `internal/shared/lifecycle` que centraliza el
modelo CRUDAR (Create/Read/Update/Delete/Archive/Restore) de todo el backend:
una tabla `archive_batches` que agrupa una operación de archivado, columnas de
"causa" (`archive_batch_id`, `archive_origin_entity`, `archive_origin_id`,
`archive_reason`) propagadas a ~30 tablas, un mapa declarativo de `Policies` por
entidad, helpers genéricos de cascade archive/restore, una barrera de invariante
("un row activo no puede referenciar a un padre archivado") con triple defensa
(Go, restore-check, triggers de Postgres), un job de limpieza
(`RunArchiveCleanup`) y una métrica Prometheus.

## Objetivo

Reemplazar la lógica de archivado/restore por-repositorio (cada repo tenía sus
propias funciones `assertXReferencesActive` / cascades ad-hoc) por un framework
declarativo único, consistente y testeable, que garantice el invariante
"archived = no existe" en toda la jerarquía customers → projects → fields →
lots → workorders/labors/supplies/...

## Problema que resuelve

- Cascadas de archivado inconsistentes por entidad.
- No había forma de **revertir exactamente** lo que una operación archivó
  (restaurar un project NO debe revivir un field archivado por separado): se
  resuelve con la "Cause" (batch + origen) que escopa el restore.
- No había barrera de DB: un INSERT/UPDATE crudo podía dejar un hijo activo
  bajo un padre archivado. Se agregan triggers `assert_parent_active`.
- No había observabilidad de cuántos writes se rechazan por referencia
  archivada (métrica `crudar_rejected_archived_ref_total`).

## Alcance en este repo (BE) — SOLO lo de mi flist

Paquete `internal/shared/lifecycle/` (9 archivos, todos `A`/creados):

- `lifecycle.go` — tipos `ArchiveBatch` (tabla `archive_batches`), `Cause`,
  `RowState`, `ActiveRef`; funciones `CreateArchiveBatch`, `RootCause`,
  `CauseFromBatch`, `CauseFromRow`, `ArchiveUpdates`, `RestoreUpdates`,
  `ApplyCauseScope`, `ReadRowState`, `IsArchived`, `RequireArchived`,
  `RequireActive`, `RequireAllActive`. Helpers `hasColumn`.
- `policy.go` — tipos `RelationPolicy` (`CASCADE_ARCHIVE`,
  `BLOCK_IF_ACTIVE_CHILDREN`, `NO_ACTION`, `RELATION_ONLY_DELETE`,
  `APPEND_ONLY_NO_DELETE`), `ParentRef`, `CascadeChild`, `Policy`, y el mapa
  global `Policies` con ~24 entidades (actors, customers, projects, fields,
  lots, campaigns, workorders, work_order_drafts, labors, supplies,
  supply_movements, stocks, managers, investors, providers, categories, crops,
  class_types/types, lease_types, business_parameters, crop_commercializations).
- `cascade.go` — `RunCascadeArchive`, `RunCascadeRestore`,
  `WouldOrphanActiveChildren`, `ListScopedIDs`, `ArchiveScopedRows`,
  `RestoreScopedRows`, `hasTable`.
- `archive_cleanup.go` — `RunArchiveCleanup(ctx, db, ArchiveCleanupOptions)`,
  reglas `archiveCleanupRules` (IA-1..IA-8+), tipos de reporte
  (`ArchiveCleanupReport`, `ArchiveCleanupAction`, `ArchiveCleanupCheck`),
  errores `ErrArchiveCleanup*`. Backfill de metadata legacy + remediación.
- `metrics.go` — `RegisterMetrics(registry, namespace)` +
  `observeRejectedArchivedRef`; counter `crudar_rejected_archived_ref_total`
  (default nil/no-op).
- Tests: `lifecycle_test`(no listado), `cascade_test.go` (195 ln),
  `archive_cleanup_test.go` (263 ln), `invariant_e2e_test.go` (214 ln, sqlite
  in-memory), `metrics_test.go` (67 ln). **Ninguno requiere DB real** (usan
  `gorm.io/driver/sqlite` `:memory:`).

Migraciones (4 pares up/down):

- `000227_crudar_archive_batches` — crea tabla `archive_batches` (+índices) y
  agrega columnas de causa + FK `NOT VALID` + índice de causa a 14 tablas
  "core" (actors, customers, projects, fields, lots, project_managers,
  project_investors, admin_cost_investors, workorders, labors, supply_movements,
  stocks, crop_commercializations, project_dollar_values). También migra
  `actors.archived_at` → `deleted_at` y recrea `ux_customers_tenant_actor_id`
  como índice **parcial** (`WHERE actor_id IS NOT NULL AND deleted_at IS NULL`).
- `000228_crudar_remaining_archive_metadata` — agrega las mismas 4 columnas +
  FK + índice a 18 tablas restantes (campaigns, supplies, managers, investors,
  providers, categories, crops, types, lease_types, business_parameters,
  field_investors, lot_dates, workorder_items, workorder_investor_splits,
  invoices, work_order_drafts, work_order_draft_items,
  work_order_draft_investor_splits).
- `000232_document_archived_inclusion_in_reports` — SOLO `COMMENT ON` sobre
  vistas/funciones (`v4_report.workorder_list`, `v4_report.labor_list`,
  `v4_ssot.labor_cost_for_lot`, `..._pre_harvest_for_lot`, `seeded_area_for_lot`,
  `supply_cost_for_lot`). No reescribe nada. Documenta la inclusión intencional
  de labors/supplies archivados (heredada de la migración 000196).
- `000233_archived_invariant_triggers` — función plpgsql
  `public.assert_parent_active(parent_table, parent_fk)` + 14 triggers
  `BEFORE INSERT OR UPDATE` en projects/fields/lots/workorders(×3)/
  work_order_drafts/labors/supplies/supply_movements/stocks/
  crop_commercializations(×2). Rechaza con SQLSTATE 23514 → Conflict en el BE.

## Alcance en el otro repo (FE)

Ninguno. Esta feature no tiene contraparte FE. (La superficie HTTP de
archive/restore que consume el FE es la feature **009 crudar-archive-surface**,
no esta.)

## Fuera de alcance (NO traer en esta feature)

- Las **20 modificaciones de `internal/*/repository.go`** que CONSUMEN el
  paquete (`internal/customer/repository.go`, `internal/project/repository.go`,
  `internal/work-order/repository.go`, etc.). Pertenecen a feature **009** y/o a
  las features por-entidad (010, 011, ...). Sin ellas, el paquete compila pero
  nadie lo llama.
- `cmd/api/main.go` (llama `lifecycle.RegisterMetrics(...)`) — compartido,
  feature 023/005.
- `cmd/archive-cleanup/main.go` (llama `lifecycle.RunArchiveCleanup(...)`) —
  no existe en develop; pertenece a feature 009/019.
- Migraciones vecinas del mismo rango que NO son mías: 223, 224, 225, 226
  (feature 001 tenancy), 231, 234 (feature 001/004 actor).
- Cambios en `go.mod`/`go.sum` (alta prob. de feature 021/dependency-bumps;
  ver risks.md — el paquete necesita prometheus + 2 módulos platform que NO
  están en develop).

## Comportamiento esperado

- Archivar un padre (`RunCascadeArchive`) crea un `ArchiveBatch`, deriva una
  `Cause`, y propaga `deleted_at` + metadata de causa a todos los
  `CascadeTables` (pivots) y `ChildEntities` (recursivo) del `Policy`.
- Restaurar (`RunCascadeRestore`) revierte SOLO los rows con la misma `Cause`
  (vía `ApplyCauseScope`). Rows archivados por otra causa quedan archivados.
- Crear/actualizar un hijo valida que sus padres (FK en `Parents`) estén activos
  (`RequireAllActive`); si no, `Conflict("X is archived")` + métrica.
- A nivel DB, los triggers rechazan el estado imposible aunque el código se
  saltee la barrera Go.

## Estado en dp~1 (777e5f6a)

Paquete completo y autocontenido. Tests presentes (sqlite in-memory). La feature
es coherente como unidad. PERO: el paquete depende de 3 módulos que no están en
el `go.mod` de develop (ver dependencies/risks), y nadie lo invoca todavía si no
se traen también los consumidores (feature 009).

## Criterios de aceptación

1. `internal/shared/lifecycle/` existe en develop con los 9 archivos.
2. `go build ./internal/shared/lifecycle/...` compila (requiere prometheus +
   `platform/observability/go` + `platform/persistence/gorm/go` en go.mod).
3. `go test ./internal/shared/lifecycle/...` pasa (no necesita DB).
4. Migraciones 227/228/232/233 aplican sin error sobre un schema con las tablas
   esperadas (idempotentes: usan `IF NOT EXISTS` / `to_regclass`).
5. Down migrations revierten limpio.

## Endpoints / Modelos / UI / DB / Tests afectados

- **Endpoints**: ninguno propio (los endpoints archive/restore están en feature 009).
- **Modelos/tipos Go**: `ArchiveBatch`, `Cause`, `RowState`, `Policy`,
  `RelationPolicy`, `ParentRef`, `CascadeChild`, `ActiveRef`,
  `ArchiveCleanup*`.
- **UI**: ninguna.
- **DB**: tabla `archive_batches`; 4 columnas de causa + FK + índices en ~32
  tablas; 14 triggers + 1 función plpgsql; índice parcial
  `ux_customers_tenant_actor_id`; comentarios de catálogo.
- **Tests**: 4-5 archivos `*_test.go` con sqlite in-memory.

## Dependencias

- **Intra-repo**: depende de que `go.mod`/`go.sum` tengan prometheus +
  `platform/observability/go` + `platform/persistence/gorm/go`. Bloquea a
  feature **009** (crudar-archive-surface) y a los repos por-entidad que llaman
  al paquete.
- **Cross-repo**: ninguna. Solo-BE.

## Riesgos (resumen; detalle en risks.md)

- **Migración out-of-order**: develop YA tiene 000229/000230; mis 227/228 quedan
  por DEBAJO. golang-migrate estricto puede rechazar aplicar 227/228 en una DB
  que ya corrió 229/230. ALTO.
- **Dependencias faltantes en go.mod de develop**. ALTO.
- **Triggers 233 requieren data limpia** (la propia migración avisa: correr
  `scripts/data-audit/archived_invariants.sql` antes). MEDIO.
- **Paquete sin consumidores** si no se trae feature 009 → dead code que igual
  debe compilar. BAJO/MEDIO.

## DECISIÓN recomendada

**Partir en subfeatures / arreglar antes de extraer.** No extraer "tal cual"
solo:
1. Primero resolver el bump de dependencias (prometheus + 2 módulos platform) —
   coordinar con feature 021; sin esto NO compila.
2. Renumerar/coordinar migraciones para que 227/228 no queden por debajo de las
   ya-aplicadas 229/230 (o confirmar que el runner del proyecto tolera
   out-of-order; ver risks.md y memoria `reset-local-db-from-prod`).
3. El paquete Go puede extraerse como unidad limpia (whole-file) una vez (1) y
   (2) están resueltos; pero su valor real llega con feature 009 — coordinar el
   orden: este paquete + migraciones primero, luego 009 (consumidores).
