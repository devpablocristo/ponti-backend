# risks.md — feature-009 · CRUDAR archive/restore/hard surface

## Funcionales

| riesgo | impacto | mitigación |
|---|---|---|
| Cambio de semántica de `DELETE /:id` en **lot**: antes archivaba (soft), ahora borra físico (hard). Un cliente viejo podría borrar datos pensando que archiva. | Alto (pérdida de datos) | Alias legacy documentado en `docs/crudar-lifecycle.md`; coordinar con FE-014/006 para usar `/archive`; comunicar al equipo. El hard delete suele exigir archivado previo + bloquear si hay hijos (protección de borrado accidental). |
| Endpoints `/archived` que no existían en algunos dominios → el FE podría asumir que existen antes de mergear BE. | Medio | BE-first estricto; no mergear FE-014/006 antes que 009. |
| `restore` de un hijo cuyo padre sigue archivado debe rechazar (invariante de integridad jerárquica). Si la implementación de 002 no lo cumple, 009 expone un endpoint roto. | Medio | Validar en repository.go de 002 (no es de 009); test `lot/repository_crudar_test.go` cubre parte. |

## Técnicos

| riesgo | impacto | mitigación |
|---|---|---|
| Diff de cada handler **mezclado** con `core→platform`, `csvexport`, `GetProtected` removal, json-tags. | Alto | EXTRAER por hunks (`git restore -p`), nunca archivo entero. Verificar con `git diff develop... \| grep -E "csvexport\|platform\|GetProtected"`. |
| `supply/mocks/mock_repository.go` desalineado con la interfaz si la extracción es parcial. | Alto (no compila) | Regenerar el mock con la herramienta del repo en vez de portar hunks. |
| `lot/usecases.go` y `lot/handler.go` mezclan CRUDAR con lot-metrics (DONE en #124): hunks de `GetMetrics(LotListFilter)` ya podrían estar en develop → conflicto o duplicado. | Medio | Excluir esos hunks; si ya están en develop, solo tomar Archive/Restore/ListArchivedLots. |

## Integración

| riesgo | impacto | mitigación |
|---|---|---|
| Si develop está en `core/*` y se traen hunks `platform/*`, no compila (import inexistente). | Alto | Decidir estado de 001 en develop ANTES; rechazar/aceptar hunks de import en consecuencia. |
| `shared/handlers` (RespondNoContent, ParsePaginationParams) ausentes en develop → helpers `runXIDAction` no compilan. | Alto | Confirmar 002/008 en develop. |

## Cross-repo

| riesgo | impacto | mitigación |
|---|---|---|
| Mergear FE-014/006 sin 009 en BE → 404 en archive/hard/archived. | Alto | BE-first; gate de release. |
| Mergear 009 sin FE → el FE viejo sigue llamando `DELETE /:id`. En dominios con alias legacy funciona (hace hard); en los que se quitó, da 404/405. | Medio | Mantener alias legacy donde haya clientes; documentar. |

## Datos / Migración

| riesgo | impacto | mitigación |
|---|---|---|
| 009 no trae migraciones, pero depende del esquema soft-delete de 002. Si 002 no aplicó las migraciones, `deleted_at` no existe → runtime error. | Alto | Verificar migraciones de 002 aplicadas en el entorno. |
| Hard delete físico irreversible. | Alto | Política "archivar antes de hard delete" + bloqueo por hijos (en repo de 002). |

## Archivos compartidos

- `internal/<dom>/handler.go`, `usecases.go`: compartidos con 001/008/013/027 → solo hunks 009.
- `internal/supply/repository_movement.go` (560 líneas): mayoría NO-009 → dejar para 002.
- `internal/shared/**` y `repository.go`: NO en mi flist → no tocar; pertenecen a 002.

## Extracción parcial

| riesgo | señal | mitigación |
|---|---|---|
| Quedan referencias a `DeleteX` viejo tras renombrar parcialmente → no compila. | `git grep DeleteCustomer\|DeleteLot...` devuelve matches en interfaces. | Renombrar handler + usecase + (repo de 002) coherentemente; build. |
| Falta `GET /archived` en un dominio porque su hunk no se trajo. | `git grep 'GET("/archived"'` no aparece para el dominio. | Checklist por dominio (tabla de implementation-status.md). |

## Riesgo de mergear SOLO este repo / SOLO el otro

- **Solo BE (009)**: seguro si hay alias legacy; el FE viejo sigue funcionando para hard delete. La UI de "archivados" no aparece hasta FE.
- **Solo FE (014/006)**: ROMPE — los endpoints no existen en BE. No hacer.
