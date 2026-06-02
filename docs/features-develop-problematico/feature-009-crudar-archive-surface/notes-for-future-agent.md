# notes-for-future-agent.md — feature-009 · CRUDAR archive/restore/hard surface

## Resumen corto

Refactor BE-only del contrato de ciclo de vida en ~14 dominios CRUDAR. Cambia
`DELETE /:id` (ambiguo) por la cuádrupla `POST /:id/archive`, `POST /:id/restore`,
`DELETE /:id/hard`, `GET /archived`, y renombra `DeleteX → HardDeleteX` en
usecases/repos. Introduce el helper `run<Entity>IDAction`. Responde 204.

## Qué está en FE y en BE

- **BE**: todo (esta feature). En `develop-problematico~1` la superficie está completa.
- **FE**: NADA en 009. El consumo va en **FE-014** (master-data pages) y **FE-006**
  (ArchivedListPage). En el repo FE no hay carpeta → "sin cambios FE".

## Archivos esenciales (los que mejor explican la feature)

- `internal/customer/handler.go` — caso de referencia limpio (helper + rutas + HardDelete).
- `internal/customer/usecases.go` — el hunk más puro 009 (`DeleteCustomer→HardDeleteCustomer`).
- `internal/lot/handler.go` — el caso con cambio de semántica (DELETE pasó de archive a hard).
- `internal/supply/handler.go` — el más complejo (supplies + supply-movements + stock-movements + globales /archived).
- `internal/lot/handler_actions_test.go` — muestra el patrón de test (spy → 204).
- `docs/crudar-lifecycle.md` (en el árbol de 002/009) — tabla de qué recursos mantienen alias legacy.

## Archivos peligrosos / mezclados (NO portar enteros)

- TODOS los `handler.go`: traen `core→platform` (001) + `GetProtected` removal (008) + a veces `csvexport` (013). Usar `git restore -p`.
- `internal/lot/handler.go` / `internal/lot/usecases.go`: mezclan con lot-metrics (DONE #124) y csvexport.
- `internal/supply/mocks/mock_repository.go`: regenerar, no portar.
- `internal/supply/repository_movement.go`: 560 líneas, mayoría de 002/013.
- Todos los `repository/models/*.go`, `usecases/domain/*.go`, `dollar/*`, `commercialization/*`, `invoice/*`, `stock/*`, `report/*`, `dashboard/*`, `businessinsights/*`: **f009=0**, son de otras features (ver file-list.md sección E). NO traer.

## Decisiones ya tomadas

- 009 es Solo-BE; FE en 014/006.
- Extracción por hunks + PRs por entidad, NO archivo entero.
- `invoice` NO es CRUDAR (clave compuesta, conserva `DeleteInvoice`).
- `dollar`/`commercialization` solo perdieron `GetProtected()` → eso es feature-008.
- La implementación real (repository.go, base.go, shared/handlers) es de **feature-002** y no está en el flist de 009.

## Dudas abiertas

1. ¿`develop` está en `core/*` o `platform/*`? Define si se aceptan los hunks de import.
2. ¿`provider` expone CRUDAR completo o parcial? (f009=10, confianza media).
3. ¿`GetMetrics(LotListFilter)` ya está en develop por #124? Si sí, excluirlo del PR de lot.

## Comandos para mirar primero

```
R=/home/pablocristo/Proyectos/pablo/ponti/core
cat /tmp/flists/be-009.txt
# prerequisito 002:
git -C "$R" grep -n "HardDeleteCustomer" internal/customer/repository.go
git -C "$R" grep -n "DeletedAt gorm.DeletedAt" internal/shared/models/base.go
# diff de referencia:
git -C "$R" diff 0972e565..777e5f6a -- internal/customer/handler.go internal/customer/usecases.go
git -C "$R" diff 0972e565..777e5f6a -- internal/lot/handler.go internal/supply/handler.go
git -C "$R" show 777e5f6a:docs/crudar-lifecycle.md
# commits clave del rango:
git -C "$R" log --oneline 0972e565..777e5f6a | grep -i crudar
```

## Errores a evitar

- NO `git checkout develop-problematico~1 -- internal/<dom>/handler.go` (archivo entero) → arrastra 001/008/013.
- NO usar `develop-problematico` (tip restore/vacío). Usar SIEMPRE `develop-problematico~1` / `777e5f6a`.
- NO mergear 009 sin 002 (no compila).
- NO mergear FE-014/006 antes de 009 (404).
- NO portar `repository.go`/`base.go`/`shared/handlers` desde aquí: son de 002.

## Camino más seguro

1. Confirmar 002 en develop.
2. Decidir estado de 001 (platform).
3. PRs por entidad, empezando por customer (referencia), con `git restore -p` aceptando solo hunks `Archive/Restore/HardDelete/archived/runXIDAction`.
4. Regenerar mock de supply.
5. `go build ./...` + `go test` por paquete + `git diff --check`.

## PR del otro repo: orden

- ANTES: ninguno del FE.
- DESPUÉS: FE-014 (master-data pages) y FE-006 (ArchivedListPage).
- Intra-repo ANTES: feature-002 (obligatorio).
