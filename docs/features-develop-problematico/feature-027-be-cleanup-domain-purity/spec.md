# spec.md — feature-027 BE cleanup & domain purity

- **id**: feature-027
- **slug**: be-cleanup-domain-purity
- **nombre**: BE cleanup & domain purity
- **tipo**: cleanup
- **repo**: Backend Go (ponti-backend) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **merge**: BE independiente
- **existe-en-FE/BE**: Solo BE. En FE NO hay carpeta para esta feature (en el cross-repo-map del FE se menciona como "sin cambios FE").
- **fuente-de-verdad (diff)**: `0972e565..777e5f6a`
- **SOURCE REF de extracción**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **Rama destino**: `develop` (tip `003a9b8f`).

## Resumen

Limpieza de "pureza de dominio" en el módulo de reportes del backend: se eliminan los tags de serialización/persistencia (`json:"..."` y `gorm:"column:..."`) de los tipos del paquete de dominio `internal/report/usecases/domain/`. El dominio queda sin saber de transporte HTTP ni de columnas SQL; esas responsabilidades viven en las capas dedicadas DTO (`internal/report/handler/dto/`) y repositorio (`internal/report/repository/models/`).

> ALCANCE REAL DE ESTA EXTRACCIÓN (autoritativo según `/tmp/flists/be-027.txt`):
> Solo **2 archivos** modificados:
> - `internal/report/usecases/domain/field-crop.go` (M)
> - `internal/report/usecases/domain/summary-results.go` (M)
>
> La NOTA de la feature menciona además: "staticcheck + remove core/governance + borrar jwt utils legacy". **Esos ítems NO están en este flist** y por lo tanto NO se documentan ni extraen aquí (ver "Fuera de alcance"). Se incluyó honestidad sobre esa discrepancia.

## Objetivo

Que los tipos de dominio de reportes (`ReportFilter`, `ProjectInfo`, `FieldCropMetric`, `FieldCrop`, `FieldCropColumn`, `FieldCropRow`, `FieldCropValue`, `LaborMetric`, `SupplyMetric`, `SummaryResults`, `ProjectTotals`, `GeneralCrops`, `SummaryResultsResponse`, `SummaryResultsFilter`) sean POGOs puros: sin anotaciones de framework (json/gorm). La forma de respuesta HTTP la define la capa DTO; el mapeo a columnas SQL la define la capa repository/models.

## Problema (estado previo)

En el dominio los structs mezclaban tres responsabilidades:
1. Modelo de negocio (lo que el dominio debe ser).
2. Contrato HTTP (`json:"..."`).
3. Mapeo de persistencia (`gorm:"column:..."` en `ProjectInfo`).

Esto viola la separación de capas que el resto del módulo report ya adoptó (existe `handler/dto/` y `repository/models/` con sus propios tags). La limpieza alinea estos dos archivos con esa arquitectura.

## Alcance en este repo (BE)

- Quitar los tags `json:"..."` de TODOS los structs en los 2 archivos del flist.
- Quitar los tags `gorm:"column:..."` de `domain.ProjectInfo`.
- Reemplazar el comentario de `ProjectInfo` por uno que documenta dónde vive el mapping SQL: apunta a `internal/report/repository/models/project_info.go` y aclara que "el domain no conoce de persistencia".
- Se preservan comentarios funcionales que ya estaban (p.ej. `// TODO: Confirmar en vista v4`, `// "number" or "text"`, `// field_id-crop_id`).

No hay cambios de nombres de campos, ni de tipos, ni de firmas de funciones, ni endpoints. Es estrictamente remoción de tags + un comentario.

## Alcance en el otro repo (FE)

Ninguno. Solo-BE. El contrato JSON hacia el FE NO cambia (ver "Comportamiento esperado"). En el cross-repo-map del FE: "sin cambios FE".

## Fuera de alcance (de ESTA extracción)

Los siguientes ítems aparecen en la NOTA de la feature y/o en el rango de diff completo `0972e565..777e5f6a`, pero **NO están en `/tmp/flists/be-027.txt`** y por lo tanto NO se extraen aquí:

- `internal/shared/utils/jwt_tools.go` (D — borrar jwt utils legacy) — fuera de este flist.
- `internal/platform/http/middlewares/gin/require_jwt.go` (D) — fuera de este flist.
- `internal/axis/jwt.go` (A) — fuera de este flist.
- "remove core/governance" — no aparece como archivo en el flist; probablemente asignado a otra feature o ya removido.
- staticcheck / golangci config — no aparece en el flist.

Estos cambios pertenecen a otros flists (p.ej. 001 be-platform-tenancy-refactor, o features de identity/JWT). Si el humano quiere agruparlos bajo "027" debe extraerlos por separado y validar pertenencia.

## Comportamiento esperado

- La salida JSON de los endpoints de reporte NO cambia. El handler (`internal/report/handler.go`, líneas ~167/174/181) NO serializa el dominio directamente: mapea a DTO con `dto.BuildFieldCropResponse(report)`, `dto.FromDomainInvestorReport(report)`, `dto.FromDomainSummaryResults(report)` y recién ahí hace `c.JSON(...)`. Los `json:"..."` viven en `handler/dto/*.go` (verificado: `ReportTableResponse`, `ReportTableColumn`, etc. mantienen sus tags). Por eso quitar los json del dominio es inocuo para el wire format.
- El scan GORM de `ProjectInfo` sigue funcionando: en `develop` el query de `getProjectInfo` aliasea columnas como `project_id`, `project_name`, `customer_id`, ... (snake_case) y GORM por NamingStrategy default snake-casea `ProjectID`→`project_id`, etc. Los `gorm:"column:..."` eran redundantes con ese default, así que removerlos no rompe el mapeo.

## Estado en dp~1 (777e5f6a)

- Ambos archivos ya están limpios (0 tags `json:`/`gorm:`). Verificado por conteo: `grep -c 'json:|gorm:'` = 0 en ambos.
- El módulo report en dp~1 ya tiene capa DTO + mappers + `repository/models/project_info.go` con su `ToDomain()`.

## Criterios de aceptación

1. `internal/report/usecases/domain/field-crop.go` y `summary-results.go` quedan sin tags `json:`/`gorm:`.
2. `go build ./...` compila.
3. `go vet ./internal/report/...` sin errores nuevos.
4. La respuesta JSON de los endpoints de reporte es byte-equivalente a la previa (mismos nombres de campo) — porque el contrato lo da el DTO.
5. `getProjectInfo` devuelve datos correctos (no campos en cero por desmapeo de columnas).

## Endpoints / modelos / UI / DB / tests afectados

- **Endpoints**: los de report servidos por `internal/report/handler.go` (field-crop, summary-results, investor-contribution). NO cambian de path/método; solo se confirma que su salida pasa por DTO.
- **Modelos (dominio, los que se limpian)**: `ReportFilter`, `ProjectInfo`, `FieldCropMetric`, `FieldCrop`, `FieldCropColumn`, `FieldCropRow`, `FieldCropValue`, `LaborMetric`, `SupplyMetric` (en field-crop.go); `SummaryResults`, `ProjectTotals`, `GeneralCrops`, `SummaryResultsResponse`, `SummaryResultsFilter` (en summary-results.go).
- **UI**: ninguna.
- **DB**: ninguna migración. Solo afecta el mapeo en memoria de `ProjectInfo` (sigue OK por default GORM).
- **Tests**: no hay tests en el flist. Existen `internal/report/handler_test.go` y `internal/report/repository_tenant_test.go` (creados en el rango, pero NO en este flist) que cubren parcialmente el módulo; pertenecen a otras features (025 test-coverage / 001 tenancy).

## Dependencias

- **Intra-repo (fuerte)**: la capa DTO (`internal/report/handler/dto/`) y los mappers (`internal/report/usecases/mappers/`) deben existir en `develop`. **Verificado: ya existen en develop** y el handler ya serializa vía DTO. Sin esto, quitar los json rompería el wire format.
- **Intra-repo (débil)**: `internal/report/repository/models/project_info.go` (el archivo al que apunta el nuevo comentario) **NO existe en develop**. No es bloqueante para compilar ni para el runtime (ver arriba), pero el comentario quedaría apuntando a un archivo inexistente hasta que se porte 001/refactor de tenancy o la creación de ese models.
- **Cross-repo**: ninguna. Solo-BE.
- **Feature dependency declarada**: DEPENDE DE `001-be-platform-tenancy-refactor` (esa feature toca `internal/report/repository.go` y crea `repository/models/project_info.go`).

## Riesgos

- **Funcional (bajo)**: que algún consumidor serialice el dominio directamente (sin DTO) en otra parte del código. Mitigación: grep `c.JSON(` con structs `domain.` (ver validation.md).
- **Técnico (bajo)**: que GORM en `develop` use una NamingStrategy custom que NO snake-casee, rompiendo `ProjectInfo`. Mitigación: revisar config GORM; el query ya aliasea en snake_case, por lo que el riesgo es marginal.
- **Comentario colgado (cosmético)**: el comentario nuevo de `ProjectInfo` referencia `repository/models/project_info.go`, ausente en develop hasta portar 001.

## DECISIÓN recomendada

**EXTRAER TAL CUAL** (whole-file de los 2 archivos), con una nota:
- Es de bajo riesgo y mecánico.
- Idealmente se ordena DESPUÉS o JUNTO con 001 (be-platform-tenancy-refactor) para que el comentario de `ProjectInfo` apunte a un archivo que exista (`repository/models/project_info.go`). Si se mergea antes, el código compila y funciona igual; solo queda un comentario que referencia un archivo aún no presente — deuda cosmética aceptable.
