# implementation-status.md — feature-002 · be-crudar-lifecycle-framework

## Estado global

- **Estado**: COMPLETA como unidad de código (el paquete es autocontenido y con
  tests), pero **NO desplegable sola** sin (a) deps en go.mod y (b) resolución
  del orden de migraciones, y **sin valor funcional** sin la feature 009 que la
  consume.
- **% completitud** (de lo que ESTA feature debe aportar): ~95%. El 5% restante
  es la coordinación de deps/orden, que es responsabilidad de extracción, no del
  código en sí.

## Estado en este repo (BE) en 777e5f6a

- Paquete `internal/shared/lifecycle/`: 5 archivos de producción + 4 de test,
  todos `A` (creados). Coherentes entre sí.
- Migraciones 227/228/232/233 (up+down) presentes y idempotentes (227/228 usan
  `IF NOT EXISTS`/`to_regclass`; 232 solo COMMENT; 233 CREATE OR REPLACE FUNCTION
  + CREATE TRIGGER sin guarda de existencia de tabla).
- API pública del paquete usada por los consumidores (verificado en
  `internal/customer/repository.go`): `CreateArchiveBatch`, `CauseFromBatch`,
  `CauseFromRow`, `RunCascadeArchive`, `ArchiveUpdates`, `RestoreUpdates`,
  `RestoreScopedRows`, `ApplyCauseScope`, `ReadRowState`, `RequireAllActive`,
  `ActiveRef`, `Cause`. Todas existen en el paquete → API consistente.

## Estado en el otro repo (FE)

- N/A. Sin contraparte FE.

## Tests

- `cascade_test.go` (195 ln), `archive_cleanup_test.go` (263 ln),
  `invariant_e2e_test.go` (214 ln), `metrics_test.go` (67 ln).
- Todos usan `gorm.io/driver/sqlite` `:memory:` y seed propio. **No requieren
  Postgres ni fixtures externos**. Deberían correr en CI sin DB.
- No hay test que valide los triggers plpgsql de la migración 233 (los triggers
  son Postgres-only; sqlite no los reproduce). El invariante se testea a nivel Go
  en `invariant_e2e_test.go`, NO a nivel DB.

## Pendientes

### BLOQUEANTE-para-mergear

1. **Deps en go.mod/go.sum**: prometheus/client_golang, platform/observability/go,
   platform/persistence/gorm/go. Sin ellas no compila. (Ver dependencies.md.)
2. **Orden de migraciones 227/228 vs 229/230 ya en develop**: decidir renumerar
   o confirmar runner tolerante. (Ver risks.md.)

### Mejora-futura

- Test de integración Postgres para los triggers de 233 (hoy solo se valida en
  Go). Útil pero no bloqueante.
- Métrica `crudar_rejected_archived_ref_total` solo se activa si
  `RegisterMetrics` se llama en bootstrap (feature 009). Hasta entonces es no-op.

### Deuda-aceptable

- Migración 233 no usa guardas `to_regclass` como 227/228; asume que las tablas
  existen. En una DB al día está bien; en una DB parcial fallará ruidosamente
  (mejor que silenciosa).
- `archive_cleanup.go` (818 líneas) es grande; las reglas IA-1..IA-8 están
  hardcodeadas en paralelo al mapa `Policies` (duplican parte de la jerarquía).
  Posible deriva futura entre ambas fuentes.

### Duda-humana

- ¿El runner de migraciones del proyecto aplica estrictamente por número o por
  orden de archivo/permite out-of-order? Revisar `scripts/` y la memoria
  `reset-local-db-from-prod-old-migrations`. Determina si 227/228 son un problema
  real.
- ¿Las deps prometheus/observability/persistence vienen de feature 021 o hay que
  agregarlas aquí? Confirmar con el dueño de 021.

## Bugs conocidos

- Ninguno detectado en el código del paquete por lectura. El riesgo está en la
  INTEGRACIÓN (deps + orden migraciones + estado de schema), no en bugs de
  lógica.
