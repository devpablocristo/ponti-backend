# risks.md — feature-002 · be-crudar-lifecycle-framework

## Riesgos técnicos

### R1 — Dependencias Go faltantes en develop (ALTO, bloqueante)
`metrics.go` importa `github.com/prometheus/client_golang/prometheus`; el go.mod
de 777e5f6a también suma `platform/observability/go v0.2.1` y
`platform/persistence/gorm/go v0.1.0`. **Ninguna está en el go.mod de develop
(003a9b8f)** — verificado. `go build` fallará.
- Mitigación: agregar las 3 deps (+ indirects prometheus client_model/common/
  procfs) en go.mod/go.sum ANTES o EN este PR, tomando solo esos hunks del rango
  (NO los bumps go-jose/x/net, ya en #124). `go mod tidy` y `go mod verify`.

### R2 — Orden de migraciones 227/228 por debajo de 229/230 (ALTO)
Develop ya tiene 000229 y 000230 (lot-metrics/dashboard #117/#121/#124). Mis
migraciones 000227/000228 tienen números MENORES. golang-migrate aplica por
número estricto: una DB ya en versión 230 NO acepta volver a aplicar 227/228.
- Mitigación: (a) confirmar el comportamiento del runner del proyecto (revisar
  `scripts/`, memoria `reset-local-db-from-prod-old-migrations` que ya documenta
  que el set viejo no respeta numeración nueva); (b) si es estricto, renumerar
  227/228/232/233 a > 234 manteniendo el ORDEN relativo (227<228<232<233) y
  actualizar cualquier referencia textual; (c) NO renumerar a ciegas — verificar
  que feature 009 no dependa de los números exactos.

## Riesgos de datos / migración

### R3 — Migración 233 requiere data limpia (MEDIO-ALTO)
Los triggers `assert_parent_active` no fallan al crearse, pero el primer
UPDATE sobre un row que YA viola el invariante (hijo activo bajo padre
archivado) explotará. La propia migración advierte: correr antes
`scripts/data-audit/archived_invariants.sql` en staging.
- Mitigación: ejecutar el data-audit y el `RunArchiveCleanup` (feature 009) en
  staging/prod ANTES de aplicar 233. Si no se trae 009 todavía, 233 puede
  aplicarse igual (los triggers solo afectan writes futuros), pero dejará writes
  legítimos rotos si la data está sucia. Considerar mergear 233 junto con o
  después de la limpieza de 009.

### R4 — Migración 233 sin guarda de existencia de tabla (MEDIO)
A diferencia de 227/228 (`to_regclass`), 233 hace `CREATE TRIGGER ... ON
public.X` directo. Si una tabla (p.ej. `crop_commercializations`, `stocks`,
`work_order_drafts`) no existe en la DB destino, la migración falla.
- Mitigación: confirmar que todas las tablas existen en develop antes de aplicar.
  En una DB al día deberían existir.

### R5 — Migración 232 COMMENT sobre objetos inexistentes (MEDIO)
232 hace `COMMENT ON VIEW v4_report.workorder_list`, `... labor_list`,
`COMMENT ON FUNCTION v4_ssot.labor_cost_for_lot(bigint)`, etc. Si alguna vista/
función no existe (porque la migración que la crea no está aplicada), falla.
- Mitigación: verificar que `v4_report.*` y `v4_ssot.*` existan en la DB destino.
  Es la migración de menor impacto (solo metadata), fácil de re-correr.

### R6 — Cambio de unicidad customers + actors.archived_at→deleted_at (MEDIO)
227 DROP+CREATE `ux_customers_tenant_actor_id` como índice PARCIAL
(`WHERE actor_id IS NOT NULL AND deleted_at IS NULL`) y hace
`UPDATE actors SET deleted_at = archived_at WHERE archived_at IS NOT NULL AND
deleted_at IS NULL`. Esto cambia semántica: a partir de aquí `deleted_at` es la
fuente de verdad de archivado de actors.
- Mitigación: validar que feature 001 (tenancy) / 004 (propername) y los repos
  de actor (feature 007/009) ya tratan `deleted_at` como flag de archivado, no
  `archived_at`. Si algún código lee `archived_at` directamente, romperá la
  semántica. El down de 227 NO recompone `archived_at` desde `deleted_at` (solo
  recrea el índice no-parcial) → pérdida de información en rollback.

## Riesgos funcionales

### R7 — Paquete sin consumidores = dead code temporal (BAJO/MEDIO)
Si se mergea 002 sin 009, el paquete compila pero nadie lo usa. Las migraciones
agregan columnas/triggers que aún no se llenan por código. Funcionalmente inerte
hasta 009. Riesgo: alguien podría borrarlo creyéndolo muerto.
- Mitigación: documentar en el PR que 009 lo consume; mergear 009 pronto.

### R8 — Doble fuente de jerarquía (BAJO)
`policy.go` (`Policies`) y `archive_cleanup.go` (`archiveCleanupRules` IA-1..IA-8)
describen la misma jerarquía padre→hijo por separado. Pueden divergir.
- Mitigación: al agregar entidades, tocar ambas. Test futuro que cruce ambas.

## Riesgos de integración / cross-repo

### R9 — Cross-repo (NINGUNO)
Solo-BE. No hay PR de FE que sincronizar. Marcar FE "sin cambios".

## Riesgos de archivos compartidos

### R10 — go.mod/go.sum (ALTO de coordinación)
Tomar el diff entero de go.mod del rango arrastraría cambios de OTRAS features
(tenancy, actor-system) y duplicaría/colisionaría con #124. Tomar solo los
hunks de las 3 deps de este paquete.

## Riesgo de extracción parcial

### R11 — Mezclar consumidores por error
Si al hacer `git checkout` se incluyen `internal/*/repository.go` o `cmd/*`, se
arrastra alcance de feature 009 y el build romperá por símbolos de OTRAS
features (actor-system, identity-tenant). Señal de scope-creep.
- Mitigación: limitar el checkout a `internal/shared/lifecycle/` y a las 8
  migraciones nombradas. Revisar `git status` (debe haber 17 paths + opcional
  go.mod/go.sum).

## Riesgo de mergear SOLO este repo

- Mergear solo BE: OK funcionalmente (no hay FE). El riesgo es interno: deps +
  orden migraciones + paquete sin consumidores. No hay incompatibilidad con FE.

## Riesgo de mergear SOLO el otro repo

- N/A (no hay FE).
