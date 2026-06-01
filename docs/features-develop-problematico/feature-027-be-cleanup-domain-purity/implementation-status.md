# implementation-status.md — feature-027 BE cleanup & domain purity

## Estado global

- **Estado**: COMPLETA en el SOURCE (`777e5f6a`), pero **NO PORTADA** a `develop`.
- **% completitud (del alcance del flist)**: 100% en el SOURCE. 0% en develop (pendiente de portar).

### Evidencia de "no portado"

- `develop:internal/report/usecases/domain/field-crop.go` → **79** ocurrencias de `json:`/`gorm:` (todavía con tags).
- `777e5f6a:...field-crop.go` → **0** ocurrencias.
- `777e5f6a:...summary-results.go` → **0** ocurrencias.

Esto confirma la NOTA: "la limpieza de json-tags del dominio BE NO está porteada -> va en 027".

## Estado en este repo (BE)

- SOURCE (dp~1): dominio limpio, capa DTO + mappers + `repository/models/project_info.go` presentes. Implementación coherente y completa.
- develop (destino): dominio aún con tags; DTO + mappers ya presentes; `repository/models/project_info.go` AUSENTE.

## Estado en el otro repo (FE)

- N/A. Solo-BE. Sin carpeta FE. "sin cambios FE".

## Tests

- No hay tests en el flist de 027.
- En el SOURCE existen `internal/report/handler_test.go` (A) y `internal/report/repository_tenant_test.go` (A), que ejercitan el módulo, pero pertenecen a OTRAS features (025 test-coverage / 001 tenancy). No se cuentan como cobertura de 027.
- Cobertura específica del cambio de tags: nula directa; se valida por compilación + igualdad de respuesta JSON + smoke de `getProjectInfo`.

## Pendientes

### BLOQUEANTE para mergear

- (Ninguno duro.) El único prerequisito fuerte —la capa DTO que serializa— ya está en develop. Por lo tanto se puede mergear de forma independiente.

### Mejora futura

- Portar `001` para que el comentario de `ProjectInfo` apunte a un `repository/models/project_info.go` existente.
- Correr `staticcheck` sobre `internal/report/...` (parte del espíritu "cleanup" de la NOTA) — fuera del flist, opcional.

### Deuda aceptable

- Comentario de `domain.ProjectInfo` referenciando `internal/report/repository/models/project_info.go` aún ausente en develop (si 027 va antes que 001). No afecta build ni runtime.

### Duda humana (a confirmar)

- ¿La NamingStrategy de GORM en develop es la default (snake_case)? El query de `getProjectInfo` ya aliasea en snake_case, así que aun con strategy custom debería mapear; confirmar config GORM si se quiere certeza total (ver risks.md / validation.md).
- ¿Los ítems "core/governance", "jwt utils legacy" y "staticcheck" de la NOTA se manejan en otras features? El flist de 027 NO los incluye; el humano debe decidir si los agrupa.

## Bugs conocidos

- Ninguno introducido por este cambio. Es remoción mecánica de tags.
