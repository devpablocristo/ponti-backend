# implementation-status.md — feature-009 · CRUDAR archive/restore/hard surface

## Estado global

- **Estado**: COMPLETA a nivel de superficie HTTP en `develop-problematico~1` (777e5f6a).
- **% completitud (en SOURCE)**: ~95% de la superficie de contrato. Lo que resta es decisión de extracción/coordinación, no código faltante.
- **% extraíble limpio a develop**: ~70% (el resto exige separar hunks de 001/008/013/027 con `restore -p`).

## Estado en este repo (BE)

| dominio | archive | restore | hard | /archived | tests 009 | confianza |
|---|---|---|---|---|---|---|
| customer | sí | sí | HardDeleteCustomer | sí (preexistía) | handler_delete_test.go (A) | alta |
| lot | sí | sí | HardDeleteLot (antes DELETE=archive) | ListArchivedLots (nuevo) | handler_actions_test.go, repository_crudar_test.go (A) | alta |
| supply | sí | sí | HardDeleteSupply | ListArchivedSupplies | usecases_delete_test, repository_delete_test (M) | media |
| supply-movement | sí | sí | HardDeleteSupplyMovement | ListArchived(+Global) | usecases_movement_test (M) | media |
| work-order | sí | sí | HardDeleteWorkOrder | ListArchivedWorkOrders | — | alta |
| work-order-draft | sí | sí | HardDeleteWorkOrderDraftByID | ListArchivedWorkOrderDrafts | usecases_test (M) | alta |
| field | sí | sí | HardDeleteField | ListArchivedFields | handler_actions_test.go (A) | alta |
| manager | sí | sí | HardDeleteManager | sí | handler_actions_test.go (A) | alta |
| investor | sí | sí | HardDeleteInvestor | sí | handler_actions_test.go (A) | alta |
| labor | sí | sí | HardDeleteLabor | ListArchivedLabors(+Global) | handler_update_labor_test (M) | media |
| business-parameters | sí | sí | HardDeleteParameter | GetArchivedParameters | — | alta |
| category | sí | sí | HardDeleteCategory | ListArchivedCategories | — | alta |
| class-type | sí | sí | HardDeleteClassType | ListArchivedClassTypes | usecases_test.go (A) | alta |
| crop | sí | sí | HardDeleteCrop | sí | — | alta |
| provider | parcial | parcial | rename | ? | — | media |

NO-CRUDAR confirmados (no aplican): `invoice` (conserva `DeleteInvoice`, clave compuesta), `dollar` y `commercialization` (solo se les quitó `GetProtected()`).

## Estado en el otro repo (FE)

Sin cambios en 009. El consumo es responsabilidad de FE-014 (pages) y FE-006 (ArchivedListPage). En el repo FE: **sin carpeta para 009** → marcar "sin cambios FE" en su cross-repo-map.

## Tests

- Creados (009): 7 archivos (`*_actions_test.go` x4, `handler_delete_test.go`, `lot/repository_crudar_test.go`, `class-type/usecases_test.go`).
- Modificados (009): tests de supply/customer/work-order-draft/labor para reflejar el rename y las nuevas rutas.
- Patrón de los handler-tests: spy del usecase que verifica que cada ruta invoca el método correcto y responde 204 (ej. `TestLotIDActionHandlersCallExplicitUseCases`).
- Commits de cobertura/lint en el rango: `f1eb01bf` (test lot CRUDAR), `656074b7`/`d629d547` (fix lint).

## Pendientes

### BLOQUEANTE para mergear
1. **feature-002 debe estar en develop** — sin los `repository.go` con `HardDeleteX/ArchiveX/RestoreX`, no compila.
2. **Separar hunks de 001/008/013** al extraer — si no, se arrastra platform/csvexport/GetProtected y el build de develop podría romper (imports a `platform/*` inexistentes).
3. **`supply/mocks/mock_repository.go`** debe coincidir con la interfaz tras extracción parcial (regenerar mock).

### Mejora-futura
- Mover archive/restore/hard de labor al grupo scoped (hay un comentario TODO en `labor/handler.go` indicando que sería más consistente).
- Deprecar alias legacy `DELETE /:id` una vez FE migrado.

### Deuda-aceptable
- Asimetría documentada de alias legacy por recurso (`docs/crudar-lifecycle.md`): unos exponen `DELETE /:id`→hard, otros no.
- `provider` con superficie CRUDAR parcial.

### Duda-humana
- ¿`develop` se queda en `core/*` o ya migró a `platform/*`? Define si los hunks de import se aceptan o rechazan.
- ¿`provider` debe exponer ciclo CRUDAR completo? Confirmar con diff.
- ¿`lot/usecases.go` — `GetMetrics(LotListFilter)` ya está en develop por lot-metrics (#124)? Si sí, ese hunk produciría conflicto/duplicado al extraer.

## Bugs conocidos

- Ninguno funcional detectado en la superficie. Riesgo principal = extracción contaminada (no bug del código fuente).
