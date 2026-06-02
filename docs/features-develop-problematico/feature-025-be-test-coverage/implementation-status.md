# implementation-status.md — feature-025 · BE test coverage sweep

## Estado global

- **En el SOURCE (777e5f6a):** **COMPLETA**. Los 45 archivos existen, son coherentes y referencian
  símbolos de producción que también existen en ese commit.
- **En la rama destino develop (003a9b8f):** **NO PORTADA / BLOQUEADA**. Los símbolos de producción que
  estos tests prueban NO existen aún en develop (faltan 001/002/009). Por eso no se pueden mergear todavía.
- **% completitud (como artefacto a extraer):** ~95%. El contenido está listo; el 5% es el riesgo de
  desalineación de firmas si 001/002/009 aterrizan en develop con APIs distintas a las del SOURCE.

## Estado en este repo (BE)

- 44 tests nuevos + 1 modificado. Tres familias homogéneas y bien formadas.
- Todos white-box (`package <modulo>`), herméticos (sqlite `:memory:`), sin dependencia de Postgres/migraciones.
- Verificado en el SOURCE: `assertCustomerReferencesActive` (customer/repository.go),
  `HardDeleteWorkOrder` (work-order/{handler,repository,usecases}.go), `GetArchivedParameters`
  (business-parameters/{handler,usecases}.go), `ListAll`/`NewRepository` (business-parameters/repository.go).

## Estado en el otro repo (FE)

No aplica. Sin cambios FE.

## Tests

Es la feature en sí. Cobertura por familia:
- **Tenant isolation (23):** lista cross-tenant filtra, get/update/hard-delete cross-tenant fallan, strict
  mode sin tenant falla en list/create.
- **Archived refs (10):** all-active OK, rechazo con kind `Conflict` cuando referencia entidad archivada,
  nil-safe / zero-id ignorados. `supply` tiene dos: `repository_archived_refs_test.go` (supply) y
  `repository_movement_archived_refs_test.go` (supply movement, con tablas projects/supplies/investors/providers/actors).
- **Handlers (13+1):** status codes, parseo de body, propagación de actor a `CreatedBy`/`UpdatedBy`,
  rutas de ciclo de vida (archive/restore/hard) en work-order.

## Pendientes

- Mergear 001/002/003/009 antes de habilitar estos tests.
- Verificar alineación de firmas de producción develop vs SOURCE (ver `dependencies.md` → inciertas).

## Bugs

Ninguno detectado en los tests. Por diseño no introducen runtime nuevo.

## Clasificación de pendientes

### BLOQUEANTE-para-mergear
- 001/002/003/009 deben estar en `develop` antes de este PR. Si no, **rompe el build de CI** (errores
  `undefined`). Es el único bloqueante real, y es de **orden**, no de contenido.

### Mejora-futura
- Mover los tests white-box a `package <modulo>_test` (black-box) donde sea posible, para desacoplarlos de
  internals — pero los `assert*ReferencesActive` son no exportados, así que algunos DEBEN seguir white-box.
- Asegurar que CI ejecute `go test ./internal/...` (no solo `go build`).

### Deuda-aceptable
- sqlite in-memory no es 100% equivalente a Postgres (tipos, `deleted_at`); los tests son de lógica de
  filtrado/integridad, no de SQL específico de Postgres. Aceptable para unit tests.

### Duda-humana
- ¿Se mergean como 1 PR o 3 (enganchados a 001/009/002)? La nota de feature dice "pueden ir como follow-up"
  y "sigue a su módulo": sugiere 3 sub-PRs. Decisión del humano.
- ¿Las firmas de producción en develop tras 001/002/009 coinciden con las del SOURCE? Verificar antes de abrir PR.
