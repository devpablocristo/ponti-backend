# extraction-plan.md — feature-027 BE cleanup & domain purity

- **repo**: ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base**: `develop` (tip `003a9b8f`)
- **SOURCE**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA `develop-problematico` (tip = restore/vacío).
- **rama sugerida**: `pr/feature-027-be-cleanup-domain-purity-be`
- **merge**: BE independiente. Solo-BE, sin coordinación con FE.

## PR title

`refactor(report): domain purity — quitar json/gorm tags del dominio de reportes`

## PR description (sugerida)

```
Quita los tags de serialización/persistencia (`json:"..."` y `gorm:"column:..."`)
de los tipos del paquete de dominio internal/report/usecases/domain/.

El dominio queda como POGO puro. El contrato HTTP lo define la capa DTO
(internal/report/handler/dto/), que ya existe en develop y por la que el
handler serializa (dto.BuildFieldCropResponse / dto.FromDomainSummaryResults
antes de c.JSON). El mapeo SQL lo define internal/report/repository/models/.

Archivos:
- internal/report/usecases/domain/field-crop.go
- internal/report/usecases/domain/summary-results.go

Sin cambios de API: la salida JSON es idéntica (la dan los DTO).
Sin migraciones. Sin cambios FE.

Verificado:
- handler ya serializa vía DTO (no serializa el dominio directo).
- getProjectInfo: el query aliasea columnas en snake_case; GORM mapea por
  default sin necesidad de gorm tags. Por eso quitarlos no rompe el scan.
```

## Pasos ordenados

1. Posicionarse en develop limpio y crear la rama.
2. Traer los 2 archivos ENTEROS desde el SOURCE (no hay hunks mixtos).
3. Compilar y vet.
4. Verificar que la salida JSON no cambió (DTO intacto).
5. Verificar `getProjectInfo` (smoke del endpoint que usa `ProjectInfo`).
6. Abrir PR.

## Archivos enteros vs parciales

- **Enteros (whole-file)**: ambos. El diff es puramente remoción de tags + 1 comentario; no hay riesgo de arrastrar otra intención.
- **Parciales**: ninguno.

## Migraciones / tests a incluir

- Migraciones: ninguna.
- Tests: ninguno en el flist. NO arrastrar `internal/report/handler_test.go` ni `internal/report/repository_tenant_test.go` (pertenecen a 025/001).

## Dependencias previas

- **Requerido en develop (ya presente)**: capa DTO + mappers + `handler.go` que serializa vía DTO. Confirmado presente en develop, así que esta extracción es auto-suficiente para compilar y para el contrato HTTP.
- **Recomendado antes (no bloqueante)**: feature `001-be-platform-tenancy-refactor`, que crea `internal/report/repository/models/project_info.go`. Si 001 va primero, el comentario nuevo de `ProjectInfo` apunta a un archivo real. Si 027 va primero, todo funciona igual; solo queda un comentario referenciando un archivo aún ausente (deuda cosmética).

## Coordinación con el otro repo

Ninguna. BE-first/FE-first no aplica: no hay cambios FE. En el cross-repo-map del FE: "sin cambios FE".

## Comandos git SUGERIDOS (para un humano — NO ejecutados por el agente)

```bash
# 0) partir de develop limpio
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout develop
git -C /home/pablocristo/Proyectos/pablo/ponti/core pull --ff-only
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout -b pr/feature-027-be-cleanup-domain-purity-be

# 1) traer los 2 archivos enteros desde el SOURCE (777e5f6a == develop-problematico~1)
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout develop-problematico~1 -- \
  internal/report/usecases/domain/field-crop.go \
  internal/report/usecases/domain/summary-results.go

# 2) revisar el diff resultante (debe ser SOLO remoción de tags + 1 comentario)
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff --staged
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff --check   # sin whitespace errors

# 3) compilar / vet / test del paquete
go build ./...
go vet ./internal/report/...
go test ./internal/report/...

# (si por algún motivo quisieras seleccionar hunks — aquí NO hace falta)
# git restore -p --source=develop-problematico~1 -- internal/report/usecases/domain/field-crop.go
```

## Qué NO traer

- Tests del módulo report (`handler_test.go`, `repository_tenant_test.go`) — otras features.
- `repository/models/project_info.go` — feature 001.
- Cualquier cosa de JWT/governance/staticcheck — fuera de este flist.

## Qué podría romperse

- Si en develop existiera otro punto que serialice `domain.*` de report directamente (sin DTO) → la respuesta perdería los nombres snake_case. Mitigación: grep (ver validation.md). Riesgo evaluado bajo: el handler ya usa DTO.
- Si GORM en develop usa NamingStrategy no-snake → `getProjectInfo` devolvería ceros. Riesgo bajo: el query ya aliasea en snake_case.

## Cómo detectar extracción incompleta

- `grep -rn 'json:"' internal/report/usecases/domain/field-crop.go internal/report/usecases/domain/summary-results.go` debe dar **0** resultados.
- `grep -rn 'gorm:"' internal/report/usecases/domain/` debe dar **0** resultados.
- Si quedaran tags, la extracción fue parcial.

## Qué validar antes del PR

- `go build ./...` y `go vet ./internal/report/...` OK.
- Diff de respuesta JSON nula (comparar contra develop pre-cambio en un endpoint de report).
- El diff del PR contiene EXACTAMENTE 2 archivos.

## Qué hacer después de mergear

- Si aún no se mergeó 001, dejar anotado que el comentario de `ProjectInfo` referencia `repository/models/project_info.go` que llegará con 001.
- Opcional: correr `staticcheck ./internal/report/...` para confirmar que no quedaron campos exportados sin uso por el cambio (no debería).
