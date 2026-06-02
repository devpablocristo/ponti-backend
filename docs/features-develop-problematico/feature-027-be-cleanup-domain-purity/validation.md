# validation.md — feature-027 BE cleanup & domain purity

## Checklist pre-PR

- [ ] El diff del PR contiene EXACTAMENTE 2 archivos:
  - `internal/report/usecases/domain/field-crop.go`
  - `internal/report/usecases/domain/summary-results.go`
- [ ] Ningún tag residual en el dominio:
  ```bash
  grep -rn 'json:"' internal/report/usecases/domain/field-crop.go internal/report/usecases/domain/summary-results.go   # esperado: 0
  grep -rn 'gorm:"' internal/report/usecases/domain/                                                                   # esperado: 0
  ```
- [ ] Comentario de `ProjectInfo` actualizado (apunta a `internal/report/repository/models/project_info.go` y dice "el domain no conoce de persistencia").
- [ ] `git diff --check` sin errores de whitespace.

## Build / vet / test (BE)

```bash
go build ./...
go vet ./internal/report/...
go test ./internal/report/...
# si existe la herramienta:
staticcheck ./internal/report/...
```

- Esperado: compila, vet sin nuevos hallazgos, tests del módulo verdes (los tests presentes en develop NO dependen de los tags del dominio).

## Validación manual / runtime

1. **Contrato JSON sin cambios** (lo más importante):
   - Levantar el server y golpear los endpoints de report (field-crop, summary-results, investor-contribution).
   - Comparar el payload de respuesta contra develop pre-cambio. Debe ser idéntico (mismos nombres snake_case: `project_id`, `customer_name`, `net_income_usd`, etc.), porque los nombres los aporta la capa DTO (`internal/report/handler/dto/*.go`).
2. **getProjectInfo** (tipo `ProjectInfo`, el que tenía `gorm:column:`):
   - Ejecutar un reporte que dispare `getProjectInfo` y confirmar que `project_name`, `customer_name`, `campaign_name` vienen poblados (no vacíos). Si vinieran vacíos → GORM no mapeó (revisar NamingStrategy).

## Verificación de que NO se serializa el dominio directo

```bash
grep -rn 'c.JSON' internal/report/
# confirmar que cada c.JSON recibe un DTO (dto.*), no un domain.* crudo.
grep -rn 'BuildFieldCropResponse\|FromDomainSummaryResults\|FromDomainInvestorReport' internal/report/handler.go
```

## Casos borde

- Reporte sin datos: `getProjectInfo` devuelve `&domain.ProjectInfo{}` (early return cuando no hay proyectos) — sigue válido tras quitar tags.
- Campos `decimal.Decimal`: el DTO ya define su propia representación (p.ej. `Number string` en `NumberValue`); quitar el json del dominio no afecta cómo el DTO serializa decimales.

## Qué revisar en UI / API / DB / env

- **UI**: nada (sin cambios FE).
- **API**: forma de respuesta idéntica; revisar solo que no cambien nombres de campo.
- **DB**: nada; sin migraciones.
- **Env**: nada.

## Qué validar en el otro repo (FE)

- Nada que portar. Validación: que el contrato JSON de los endpoints de report siga igual (responsabilidad del BE). Si el FE tiene tests de integración contra estos endpoints, que sigan verdes — pero no requieren cambios.

## Señales de incompletitud / incompatibilidad

- `grep 'json:"'` o `grep 'gorm:"'` en los 2 archivos devuelve > 0 → extracción incompleta.
- El PR toca más de 2 archivos → se arrastró algo de otra feature (revisar y quitar).
- Respuesta JSON con claves en CamelCase → algún path serializa el dominio directo (bug a investigar antes de mergear).
- `getProjectInfo` devuelve campos vacíos → problema de NamingStrategy GORM (revisar config).
