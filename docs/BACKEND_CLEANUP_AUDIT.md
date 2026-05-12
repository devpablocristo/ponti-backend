# Backend Cleanup Audit

Fecha: 2026-05-12

Este documento es la base versionada para limpiar el backend de Ponti sin hacer un
refactor masivo ni cambiar contratos publicos. La regla de trabajo es avanzar por
etapas chicas, validar despues de cada etapa y no tocar calculos sensibles sin
golden master.

## Estado de linea base

Comandos verificados antes de iniciar la limpieza:

- `go test ./...`: verde
- `go build ./cmd/api ./cmd/migrate`: verde
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --timeout=5m`: verde

Inventario actual:

- 403 archivos Go
- 41 archivos de test
- 56 migraciones `up`
- 210 referencias a `authz.MaybeTenantScope`
- 166 usos de SQL raw/exec
- 318 menciones a `legacy`, `deprecated`, `TODO`, `FIXME` o `HACK`

Hotspots por tamano/riesgo:

- `internal/project/repository.go`
- `internal/supply/handler.go`
- `internal/data-integrity/usecases.go`
- `internal/labor/repository.go`
- `internal/actor/repository.go`
- `internal/work-order/repository.go`
- `internal/dashboard/handler/dto/dashboard.go`
- `internal/stock/repository.go`
- `internal/report/repository.go`

## Diagnostico arquitectonico

El backend es un monolito Go modular por dominio bajo `internal/`. La estructura
principal es Gin para HTTP, GORM para persistencia CRUD, SQL directo para
reportes/calculos, usecases por modulo, DTOs HTTP por handler y modelos de
persistencia por repository.

Lo que esta bien:

- Hay separacion general entre handler, usecase, repository, domain y DTOs.
- Los errores estan mayormente normalizados via `domainerr` y `sharedhandlers`.
- Los CRUDs principales ya tienen acciones explicitas `archive`, `restore` y
  `hard`.
- Hay base multi-tenant con `RequireTenant`, `TenantScope`,
  `MaybeTenantScope` y migraciones `000224`/`000225`.
- CI ejecuta lint, build, tests y govulncheck.
- Existen scripts de schema guardrails, tenant audit y golden master.

Problemas estructurales:

- `/me/context` esta implementado inline en `cmd/api/http_server.go`; deberia
  vivir en un modulo auth/admin con repository/usecase propio.
- `MaybeTenantScope` esta muy extendido en paths tenant-owned. Es correcto como
  transicion, pero no como estado final strict.
- Hay handlers grandes con demasiada logica operativa, especialmente
  `internal/supply/handler.go` y export/import handlers.
- Hay endpoints `DELETE /:id` legacy con semantica inconsistente: en algunas
  entidades archivan y en otras hacen hard delete.
- Varias interfaces triviales se repiten en todos los handlers
  (`GinEnginePort`, `ConfigAPIPort`, `MiddlewaresEnginePort`) y generan ruido.
- Hay SQL raw valido para reportes, pero no siempre queda documentado por que la
  query debe ser SQL directo.

## Matriz inicial por modulo

| Modulo | Tests directos | MaybeTenantScope | Raw/Exec | Riesgo inicial |
| --- | ---: | ---: | ---: | --- |
| actor | 0 | 0 | 39 | Alto: merge/deduplicacion y SQL manual |
| admin | 1 | 0 | 1 | Medio: auth/RBAC y contexto usuario |
| campaign | 1 | 13 | 0 | Medio: CRUD tenant-owned sin tests directos |
| customer | 1 | 10 | 1 | Medio: CRUD base, delete legacy hard |
| field | 1 | 11 | 0 | Medio: CRUD tenant-owned sin tests directos |
| investor | 1 | 11 | 0 | Medio: mirrors actors y relaciones |
| invoice | 1 | 5 | 7 | Medio: documentos y snapshots |
| labor | 2 | 15 | 8 | Alto: dos rutas semanticas y dependencias con OT |
| lot | 3 | 25 | 4 | Alto: fechas, hectareas y hard-delete con referencias |
| manager | 1 | 11 | 0 | Medio: mirrors actors y relaciones |
| project | 1 | 13 | 32 | Alto: entidad central y relaciones pivots |
| report | 0 | 2 | 5 | Alto: calculos sensibles |
| stock | 2 | 15 | 43 | Alto: stock real/calculado y movimientos |
| supply | 5 | 48 | 4 | Alto: handler grande, imports y movimientos |
| work-order | 3 | 18 | 5 | Alto: OT, splits, facturas y stock |
| work-order-draft | 1 | 8 | 0 | Medio: publicacion digital |

Los modulos de catalogo global/simple (`category`, `class-type`, `crop`,
`lease-type`, `dollar`, `business-parameters`, `commercialization`) no muestran
tenant scope ni SQL raw relevante en esta medicion, pero igual deben revisarse
por consistencia CRUD y permisos.

## Codigo muerto o sospechoso

No se elimino codigo muerto en esta etapa porque la linea base de lint esta
verde y varias piezas sospechosas pueden ser compatibilidad publica.

Candidatos a revisar antes de eliminar:

- `internal/platform/http/middlewares/gin/require_jwt.go`: middleware JWT
  legacy no referenciado por el wiring actual, mientras el auth real usa
  Identity Platform.
- `Middlewares.GetProtected()`: hoy devuelve slice vacio y aparece en muchas
  interfaces, pero no se usa como control real.
- `ParseUserID`: alias deprecated de `ParseActor`; conservar hasta verificar
  callers externos o limpiar todas las referencias internas.
- Funciones con `//nolint:unused` en mappers/reportes: revisar si son mapeos
  reservados para responses futuras o si quedaron de una version anterior.
- `connectWithConnectorIAMAuthN`: no usado, pero puede ser fallback de Cloud SQL;
  requiere decision operacional antes de eliminar.

Decision: documentar y testear primero; eliminar solo cuando `rg`, wiring,
tests y revision de contratos confirmen que no hay uso.

## Codigo duplicado e inconsistencias

Duplicacion detectada:

- Interfaces de handler repetidas por modulo para engine/config/middlewares.
- Parseo manual de IDs y parametros en handlers, aunque ya existen helpers en
  `internal/shared/handlers`.
- Patron CRUD repetido con pequenas diferencias de status codes y nombres.
- Logica de archive/restore/hard-delete repetida en repositories.
- Reglas de dependencia hard-delete repetidas por entidad.
- Export/import mezclado en handlers con parseo de archivos, response building
  y reglas de dominio.

Inconsistencias visibles:

- `DELETE /:id`:
  - `lot`: legacy hacia archive.
  - `customer`, `project`, `work-order`, `supply`: legacy hacia hard delete.
  - `supply movement`: delete operativo, no equivalente a archive.
  - `labor`: conviven rutas de proyecto y OT, con hard delete explicito solo en
    el grupo de OT.
- Algunos modulos usan `sharedhandlers.RespondNoContent`, otros responden JSON
  manual.
- Algunos listados devuelven envoltorios `{success,data}` y otros payloads
  directos.
- Algunos repositorios validan dependencias antes de hard delete; otros delegan
  mas en constraints o checks parciales.

## Riesgos criticos

- Cross-tenant accidental si un path tenant-owned queda con `MaybeTenantScope`
  sin tenant en contexto o con SQL raw sin filtro. Mitigacion: migrar modulo por
  modulo a `RequireTenant`/`TenantScope` y agregar tests IDOR.
- Cambios en `dashboard`, `report` o `data-integrity` pueden alterar calculos.
  Mitigacion: golden master antes/despues y diff numerico `0`.
- Hard delete legacy puede borrar datos donde la UI cree estar archivando.
  Mitigacion: mantener rutas por compatibilidad, pero documentar y mover UI a
  endpoints explicitos.
- Logs/middleware pueden exponer headers sensibles si se agrega debugging sin
  redaccion. Mitigacion: centralizar redaccion y tests de logging.

## Riesgos medios

- Handlers grandes son dificiles de testear y favorecen validaciones duplicadas.
- SQL raw sin comentario operacional complica mantenibilidad y tuning.
- Interfaces por paquete pueden ocultar dependencias reales y hacer mas dificil
  refactorizar.
- Migraciones acumuladas contienen reconciliaciones y compat layers; no deben
  compactarse sin estrategia de baseline.

## Quick wins seguros

- Alinear `make lint` con CI. Hecho: el target usa golangci-lint v2.11.4 via
  `go run` y no depende de un `.golangci.yml` inexistente.
- Crear este reporte de auditoria versionado. Hecho.
- Agregar tests de comportamiento para rutas legacy antes de modificar
  semantica.
- Extraer `/me/context` fuera de `cmd/api/http_server.go` sin cambiar response.
- Crear helpers compartidos para legacy delete wrappers y parseo de ID donde ya
  se repite el mismo patron.

## Plan incremental

### Etapa 0 - Auditoria

- Mantener este documento actualizado.
- Completar matriz por modulo con handler/usecase/repo/dto/model/migrations/tests.
- Registrar rutas legacy y consumidores FE/BFF antes de remover o cambiar nada.

### Etapa 1 - Cleanup mecanico

- Eliminar imports, comentarios obsoletos y helpers internos muertos solo si
  lint/tests y wiring lo confirman.
- No eliminar endpoints ni DTOs publicos.
- Normalizar targets de Makefile y comandos de validacion.

### Etapa 2 - Consistencia CRUD

- Normalizar `Create`, `List`, `Get`, `Update`, `Archive`, `Restore`,
  `HardDelete` por entidad.
- Mantener rutas legacy pero redirigir internamente a metodos explicitos.
- Agregar tests para status code y efecto real de archive/hard-delete.

### Etapa 3 - Boundaries

- Extraer logica pesada de handlers hacia usecases/services locales.
- Mover `/me/context` a modulo auth/admin.
- Reducir interfaces triviales cuando no agregan inversion real ni facilitan
  tests.

### Etapa 4 - Datos y seguridad

- Migrar repos tenant-owned desde `MaybeTenantScope` a scope obligatorio.
- Revisar cada SQL raw con tenant, parametros y comentario de intencion.
- Agregar tests IDOR por modulo.

### Etapa 5 - Reportes y calculos

- Congelar golden master antes de tocar reportes.
- Refactorizar solo si el diff numerico sigue en `0`.

### Etapa 6 - Cierre

- Ejecutar test/build/lint/migraciones.
- Actualizar deuda restante con owner, riesgo y decision.
- Dejar CI verde.

## Cambios implementados en esta etapa

- `Makefile`: `make lint` ahora ejecuta el mismo golangci-lint que CI
  (`v2.11.4`) sin depender de un archivo de config inexistente.
- `/me/context`: la ruta salio del bootstrap `cmd/api/http_server.go` y ahora
  vive en `internal/admin`, preservando path, middleware y payload.
- `internal/admin`: agregado test de contrato para `/me/context` con usuario,
  tenant activo, rol y permisos. El test detecto y corrigio el mapeo GORM de
  `idp_sub` / `idp_email`.
- `internal/campaign`: agregado test de contrato para archive, restore y hard
  delete; se redujo duplicacion de parseo/respuesta en el handler con helper
  local. No se cambio el shape del listado legacy.
- `internal/customer`: agregado test de contrato para las acciones por ID:
  legacy `DELETE /:id`, hard delete explicito, archive y restore. Tambien se
  redujo duplicacion de parseo/respuesta en el handler con un helper local.
- `internal/customer/repository_harddelete_test.go`: actualizado el contrato
  documentado del test de integracion; el comportamiento actual bloquea hard
  delete si el cliente tiene proyectos, no borra en cascada.
- `internal/field`: agregado test de contrato para hard delete, archive y
  restore; se redujo duplicacion de parseo/respuesta en el handler con helper
  local.
- `internal/investor`: agregado test de contrato para hard delete, archive y
  restore; se redujo duplicacion de parseo/respuesta en el handler con helper
  local.
- `internal/manager`: agregado test de contrato para hard delete, archive y
  restore; se redujo duplicacion de parseo/respuesta en el handler con helper
  local equivalente al de `customer`.
- `docs/BACKEND_CLEANUP_AUDIT.md`: auditoria inicial, riesgos, quick wins y plan
  incremental versionado.

## Validacion ejecutada en esta etapa

- `make lint`: verde, 0 issues.
- `go test ./...`: verde.
- `go build ./cmd/api ./cmd/migrate`: verde.

## Deuda restante

- Falta profundizar la matriz modulo por modulo con rutas, DTOs, modelos,
  migraciones y consumidores FE/BFF.
- Falta decidir destino de middleware JWT legacy.
- Falta separar `/me/context` en usecase/repository propio; la primera extraccion
  ya saco la logica del bootstrap HTTP.
- Falta homogeneizar semantica de rutas `DELETE /:id`.
- Falta convertir `MaybeTenantScope` transicional a strict en todos los paths
  tenant-owned.
- Falta golden master antes de tocar reportes.
