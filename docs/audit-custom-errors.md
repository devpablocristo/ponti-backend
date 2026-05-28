# Audit — Custom errors (fmt.Errorf/errors.New) en Ponti backend

Read-only sweep de los archivos NO-infra que aún usan `fmt.Errorf`/`errors.New`. Por cada uso, propongo migración a `domainerr.*`.

Convenciones:
- **Validation**: input inválido del usuario / parámetro requerido faltante.
- **NotFound**: recurso solicitado no existe.
- **Conflict**: estado inconsistente / duplicado.
- **BusinessRule**: regla de negocio incumplida.
- **Internal**: error de infra (DB, JSON, etc) que se wrappa hacia arriba — el cliente ve "internal".
- **UpstreamError**: fallo de un servicio externo (review, ai, etc).
- **Unavailable**: servicio no disponible por configuración o estado.

---

## 1. `internal/business-parameters/usecases/domain/business_parameter.go`

| Línea | Snippet | Propuesta | Por qué |
|---|---|---|---|
| 26 | `fmt.Errorf("parameter %s is not of type decimal", ap.Key)` | `domainerr.Validation(fmt.Sprintf("parameter %s is not of type decimal", ap.Key))` | El caller solicita un tipo y el dato almacenado no matchea — input inválido desde el punto de vista del usecase consumer. |
| 38 | `fmt.Errorf("parameter %s is not of type integer", ap.Key)` | `domainerr.Validation(...)` | Idem. |
| 46 | `fmt.Errorf("parameter %s is not of type boolean", ap.Key)` | `domainerr.Validation(...)` | Idem. |

## 2. `internal/report/repository.go`

16 entries, todas wrappings de errores DB:

| Línea | Snippet | Propuesta |
|---|---|---|
| 45, 85, 547, 618, 697 | `fmt.Errorf("error obteniendo proyectos relacionados: %w", err)` | `domainerr.Internal(fmt.Sprintf("error obteniendo proyectos relacionados: %v", err))` |
| 67 | `fmt.Errorf("error al obtener métricas: %w", err)` | `domainerr.Internal(...)` |
| 110, 509 | `fmt.Errorf("error obteniendo columnas: %w", err)` | `domainerr.Internal(...)` |
| 399 | `fmt.Errorf("error querying supply categories: %w", err)` | `domainerr.Internal(...)` |
| 457 | `fmt.Errorf("error querying labor categories: %w", err)` | `domainerr.Internal(...)` |
| 500, 730 | `fmt.Errorf("error getting project information: %w", err)` | `domainerr.Internal(...)` |
| 515 | `fmt.Errorf("error getting metrics: %w", err)` | `domainerr.Internal(...)` |
| 586 | `fmt.Errorf("error consultando vista de aportes de inversores: %w", err)` | `domainerr.Internal(...)` |
| 600 | `fmt.Errorf("error convirtiendo modelo a domain: %w", err)` | `domainerr.Internal(...)` |
| 668 | `fmt.Errorf("error ejecutando query de resumen de resultados: %w", err)` | `domainerr.Internal(...)` |

**Trade-off**: `domainerr.Internal(...)` toma un `string`, no soporta `%w` para wrappear. Se pierde la cadena de unwrap. Mitigación: el error original puede agregarse al log (lo cual ya hacemos con el middleware) y al `domainerr.Error.Message` queda "error X: ...".

## 3. `internal/report/usecases/validators.go`

| Línea | Snippet | Propuesta |
|---|---|---|
| 22 | `fmt.Errorf("customer_id, project_id and campaign_id are required")` | `domainerr.Validation("customer_id, project_id and campaign_id are required")` |

## 4. `internal/report/usecases.go`

4 wrappings de repository:

| Línea | Snippet | Propuesta |
|---|---|---|
| 58 | `fmt.Errorf("error al obtener reporte de campo/cultivo: %w", err)` | `domainerr.Internal(...)` |
| 71 | `fmt.Errorf("error obteniendo reporte de aportes de inversores: %w", err)` | `domainerr.Internal(...)` |
| 89 | `fmt.Errorf("error obteniendo resumen de resultados: %w", err)` | `domainerr.Internal(...)` |
| 113 | `fmt.Errorf("error getting project information: %w", err)` | `domainerr.Internal(...)` |

**Alternativa**: si el repository ya devuelve `domainerr.Internal`, el usecase puede `return nil, err` directamente sin re-wrappear. Más limpio.

## 5. `internal/businessinsights/service.go`

| Línea | Snippet | Propuesta |
|---|---|---|
| 106 | `fmt.Errorf("review submit: %w", err)` | `domainerr.UpstreamError(fmt.Sprintf("review submit: %v", err))` — falla del servicio review (governance). |
| 139 | `fmt.Errorf("upsert candidate: %w", err)` | `domainerr.Internal(...)` — DB upsert. |

## 6. `internal/project/handler/dto/project.go`

| Línea | Snippet | Propuesta |
|---|---|---|
| 106 | `fmt.Errorf("lease_type_percent: %w", err)` | `domainerr.Validation(fmt.Sprintf("lease_type_percent: %v", err))` — parsing input usuario. |
| 110 | `fmt.Errorf("lease_type_value: %w", err)` | `domainerr.Validation(...)` |
| 162 | `fmt.Errorf("invalid decimal value")` | `domainerr.Validation("invalid decimal value")` |

## 7. `internal/report/repository/models/investor-contribution.go`

5 errores de deserialización JSON:

| Línea | Snippet | Propuesta |
|---|---|---|
| 136 | `fmt.Errorf("error deserializando investor_headers: %w (JSON: %s)", err, m.InvestorHeadersJSON)` | `domainerr.Internal(...)` |
| 145 | `fmt.Errorf("error deserializando general_project_data: %w", err)` | `domainerr.Internal(...)` |
| 160 | `fmt.Errorf("error deserializando contribution_categories: %w", err)` | `domainerr.Internal(...)` |
| 172 | `fmt.Errorf("error deserializando investor_contribution_comparison: %w", err)` | `domainerr.Internal(...)` |
| 182 | `fmt.Errorf("error deserializando harvest_settlement: %w", err)` | `domainerr.Internal(...)` |

## 8. `internal/data-integrity/usecases.go`

8 wrappings, todos errores de queries internas:

| Línea | Snippet | Propuesta |
|---|---|---|
| 115, 122, 129, 136, 142, 150, 160 | `fmt.Errorf("fetch X: %w", err)` | `domainerr.Internal(...)` |
| 199 | `setErr(fmt.Errorf("control %d failed: %w", controlNumber, err))` | `domainerr.Internal(...)` |

## 9. `internal/ai/client.go`

| Línea | Snippet | Propuesta |
|---|---|---|
| 47 | `fmt.Errorf("ai service url not configured")` | `domainerr.Unavailable("ai service url not configured")` — config missing, servicio no disponible. |
| 68 | `fmt.Errorf("ai service url not configured")` | `domainerr.Unavailable(...)` — idem. |

---

## Resumen del audit

| Archivo | Entradas | Validation | Internal | Upstream | Unavailable |
|---|---|---|---|---|---|
| business_parameter.go | 3 | 3 | — | — | — |
| report/repository.go | 16 | — | 16 | — | — |
| report/usecases/validators.go | 1 | 1 | — | — | — |
| report/usecases.go | 4 | — | 4* | — | — |
| businessinsights/service.go | 2 | — | 1 | 1 | — |
| project/handler/dto/project.go | 3 | 3 | — | — | — |
| investor-contribution.go | 5 | — | 5 | — | — |
| data-integrity/usecases.go | 8 | — | 8 | — | — |
| ai/client.go | 2 | — | — | — | 2 |
| **Total** | **44** | **7** | **34** | **1** | **2** |

\* `report/usecases.go` se puede simplificar: si el repo devuelve `domainerr.Internal`, el usecase devuelve `err` sin re-wrappear.

## Trade-off general

`domainerr.*` constructores aceptan `string`, no `error`. Para preservar la cadena de unwrap (`%w`), se requeriría agregar un constructor `domainerr.Wrap(kind, err)`. Sin esa adición, la migración mecánica:
- Pierde `errors.Is`/`errors.As` profundo.
- El `error` original se interpola con `%v` en el mensaje.
- El log central (que agregamos en Fase 1) emite el error original via `error_handling.go` antes de normalizar.

**Recomendación**: este trade-off es aceptable para los casos `Internal` (donde el unwrap es solo informativo). Para los `Validation` (3+1+3 = 7 casos), el original `err` (si lo hay) es secundario al mensaje.

---

## Próximo paso

Después de tu revisión file por file, aplico las migraciones en una Fase 3b. ¿Avanzo con todas, alguna excluyo, o querés ajustar el mapeo en algún punto específico?
