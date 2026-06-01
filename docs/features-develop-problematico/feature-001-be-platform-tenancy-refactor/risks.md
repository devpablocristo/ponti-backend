# risks — feature-001 · Platform tenancy refactor (Fase 7)

## Funcionales

| riesgo | severidad | mitigación |
|---|---|---|
| Cambio de semántica en `ResolveProjectIDs`: en modo tenant, "sin filtros" = todos los proyectos del tenant; "0 matches" devuelve `[]int64{0}` (centinela) en vez de `nil`. Listados podrían cambiar de comportamiento. | alta | Mergear junto con (o después de) 003/008 para que el tenant esté presente; cubrir con `workspace_test.go`; smoke manual de listados por tenant. |
| `RejectUnsafeLocalAuthz` rechaza requests en entornos no local-like sin `cfg.Auth.Enabled`. Si una config de staging quedó sin Auth.Enabled, todo el tráfico cae a 401/403. | alta | Verificar `cfg.Auth.Enabled` y `cfg.Auth.Environment` en cada env antes de mergear; revisar `isLocalLikeEnvironment`. |
| Validación issuer/audience más estricta podría rechazar tokens válidos si la config de issuer/audience no coincide con Identity Platform. | media | Validar issuer/audience contra config real en staging. |

## Técnicos

| riesgo | severidad | mitigación |
|---|---|---|
| Extracción parcial mal cortada: un `repository.go` queda importando `lifecycle`/`actor` sin traer esos símbolos → no compila. | alta | `git restore -p` con split de hunks; `go build ./...` tras cada repo; si no compila standalone, dejar el repo para su feature dueña. |
| `platform/*` no resuelve en go.mod de develop → build roto. | alta | Confirmar go.mod/go.sum antes de empezar; bumps go-jose/x/net ya en #124. |
| Imports sin usar tras swap (goimports). | baja | `go vet` / `goimports -l`. |
| Borrado de `strings.go` con callers vivos. | media | `git grep IsNumeric\\|NormalizeString` antes de borrar; si hay callers, no borrar en 001. |

## Integración / cross-repo

| riesgo | severidad | mitigación |
|---|---|---|
| Cross-repo: ninguno (solo-BE). | n/a | Marcar FE como "sin cambios". |
| Otras features BE posteriores que tocan los mismos repos necesitan rebasar sobre 001. | media | Mergear 001 primero; rebasar 002/007/009 después. |

## Datos / migración

| riesgo | severidad | mitigación |
|---|---|---|
| No hay migraciones en esta flist; pero el scoping asume columna `tenant_id`. Si se mergea 001 sin 003, queries con alias quedan inocuas (modo transición) — no corrupción, pero el aislamiento no aplica. | media | Coordinar 001→003→008. |

## Archivos compartidos

| archivo | riesgo | mitigación |
|---|---|---|
| 20× `internal/*/repository.go` | mezclan 001/002/007/009 | partial-hunks; sección B del plan |
| `internal/shared/models/base.go` | compartido con 007/008 | traer solo hunk de import contextkeys |
| `internal/report/repository.go`, `dashboard/repository.go` | grueso del diff es 027/013, no 001 | posponer o solo import+ctx |
| `internal/platform/files/excel/excelize/*` | borrado de 013 | NO traer en 001 |
| go.mod/go.sum | resolución platform/* | no modificar en 001 si ya está en develop |

## Extracción parcial

- **Señal de incompletitud**: `go build` falla por `undefined: lifecycle.*` / `actorsync.*` → trajiste import sin símbolos, o cortaste de menos.
- **Señal de sobre-extracción**: el diff del PR incluye `RunCascadeArchive`, `EnsureCustomerFromActor`, `legacy_actor_map`, borrado excelize, o `fmt.Errorf`→domainerr en report → trajiste de más (002/007/009/013/027).

## Riesgo de mergear solo este repo / solo el otro

- **Solo este repo (BE)**: seguro y esperado — es solo-BE. El único riesgo real es mergear 001 sin 003/008 (scoping inocuo, sin aislamiento real, posible cambio de listados por la nueva semántica de `ResolveProjectIDs`).
- **Solo el otro repo (FE)**: N/A, FE no tiene cambios para esta feature.
