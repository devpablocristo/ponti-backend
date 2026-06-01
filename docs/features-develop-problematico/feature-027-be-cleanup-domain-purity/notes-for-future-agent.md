# notes-for-future-agent.md — feature-027 BE cleanup & domain purity

## Resumen corto

Feature de cleanup de "pureza de dominio" en el módulo report del BE. El flist real (`/tmp/flists/be-027.txt`) tiene SOLO 2 archivos:
- `internal/report/usecases/domain/field-crop.go`
- `internal/report/usecases/domain/summary-results.go`

El cambio es 100% mecánico: quitar tags `json:"..."` (y los `gorm:"column:..."` de `ProjectInfo`) de los structs de dominio, dejándolos como POGOs. Más un comentario en `ProjectInfo` que apunta a `repository/models/project_info.go`. Extracción **whole-file** de ambos, **bajo riesgo**.

## Qué está en FE y en BE

- **FE**: NADA. Solo-BE. En el cross-repo-map del FE: "sin cambios FE".
- **BE**: los 2 archivos de dominio. El contrato JSON NO cambia.

## Archivos esenciales / peligrosos / mezclados

- **Esenciales (los que extraés)**: los 2 archivos del flist. Nada más.
- **Peligrosos**: ninguno realmente. No hay archivos compartidos (no se toca `wire/`, `cmd/`, `go.mod/sum`, `Makefile`, `internal/shared/**`).
- **Mezclados / a NO arrastrar**: en el rango de diff completo aparecen muchos cambios de otras features (DTO de report, mappers, `repository.go` de tenancy, tests, borrado de JWT en `internal/shared/utils/jwt_tools.go`, `internal/platform/http/middlewares/gin/require_jwt.go`, `internal/axis/jwt.go`). NADA de eso está en este flist. No lo traigas bajo "027".

## Decisiones ya tomadas

1. Extracción whole-file de los 2 archivos (el diff no se mezcla con otra intención).
2. Confirmado que quitar los `json:` del dominio es SEGURO: el handler ya serializa vía DTO (`dto.BuildFieldCropResponse`, `dto.FromDomainSummaryResults`) antes de `c.JSON`. La capa DTO (con sus json tags) YA está en develop.
3. Confirmado que quitar los `gorm:column:` de `ProjectInfo` es seguro: el query de `getProjectInfo` aliasea columnas en snake_case y GORM default snake-casea los campos → mapea igual sin tags.
4. La NOTA de la feature (staticcheck / core/governance / jwt utils legacy) excede el flist; esos ítems se documentan como "fuera de alcance / otra feature".

## Dudas abiertas

- ¿La NamingStrategy de GORM en develop es la default? (alta confianza que sí; el query aliasea en snake_case de todos modos). Confirmar en la config de la conexión GORM si querés certeza absoluta.
- ¿Dónde viven realmente "core/governance" y "staticcheck" del enunciado? No están en `be-027.txt`. Probablemente en 001 o en una feature de CI/config.

## Qué comandos mirar primero

```bash
# el flist autoritativo
cat /tmp/flists/be-027.txt

# el diff real (debe ser solo remoción de tags + 1 comentario)
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff 0972e565..777e5f6a -- \
  internal/report/usecases/domain/field-crop.go \
  internal/report/usecases/domain/summary-results.go

# confirmar dependencia satisfecha en develop (DTO + handler serializa vía DTO)
git -C /home/pablocristo/Proyectos/pablo/ponti/core show develop:internal/report/handler.go | grep -nE 'dto\.|c\.JSON'

# confirmar que el dominio en develop AÚN tiene tags (no portado)
git -C /home/pablocristo/Proyectos/pablo/ponti/core show develop:internal/report/usecases/domain/field-crop.go | grep -c 'json:\|gorm:'   # ~79
```

## Errores a evitar

- NO usar `develop-problematico` como SOURCE (tip = restore/vacío). Usar `develop-problematico~1` (`777e5f6a`).
- NO traer tests ni `repository/models/project_info.go` con un glob amplio: solo los 2 paths.
- NO ejecutar comandos que muten git. Los comandos del plan son sugerencias para un humano.
- NO asumir que la NOTA == el flist. El flist manda; la NOTA es más amplia.

## Camino más seguro

`git checkout develop` → branch nueva → `git checkout develop-problematico~1 -- <los 2 paths>` → `go build ./... && go vet ./internal/report/...` → comparar payload de un endpoint de report → PR.

## Qué PR del otro repo va antes / después

- Ninguno del FE (no hay cambios FE).
- En el BE, idealmente DESPUÉS de `001-be-platform-tenancy-refactor` (que crea `repository/models/project_info.go`), pero no es bloqueante. Si va antes, queda solo un comentario referenciando un archivo aún ausente — deuda cosmética.
