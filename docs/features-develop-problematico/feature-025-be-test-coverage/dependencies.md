# dependencies.md — feature-025 · BE test coverage sweep

## Resumen

feature-025 es **puramente derivada**: no produce código de runtime, solo prueba el código de otras
features. Por eso no bloquea a nadie y depende fuerte de varias. Es la clase "consumidor", no "productor".

## Depende-de

### Fuertes (bloqueantes para compilar)

| feature | por qué | evidencia (verificada) |
|---|---|---|
| **001 be-platform-tenancy-refactor** | los 23 `repository_tenant_test.go` inyectan `contextkeys.OrgID/Actor/Role/Scopes` y esperan que el repo filtre por tenant. | imports `platform/security/go/contextkeys`; producción de tenancy ausente en develop tip. |
| **009 crudar-archive-surface** | los 10 `repository_archived_refs_test.go` llaman a `assert<Entity>ReferencesActive(db, ...)` (funciones **no exportadas** del paquete de producción) y a `GetArchivedParameters`/`ListArchivedWorkOrders`. | `assertCustomerReferencesActive` existe en `internal/customer/repository.go` SOLO en 777e5f6a, NO en develop tip (003a9b8f). |
| **002 be-crudar-lifecycle-framework** | `work-order/handler_test.go` (M) renombra el stub `DeleteWorkOrderByID`→`HardDeleteWorkOrder`, agrega `ListArchivedWorkOrders` al stub y prueba rutas `/archive`,`/restore`,`/hard`. | `HardDeleteWorkOrder` existe en `internal/work-order/{handler,repository,usecases}.go` SOLO en 777e5f6a. |

### Débiles

| feature | por qué |
|---|---|
| **003 be-multitenant-db-hardening** | los tenant tests verifican `TENANT_STRICT_MODE=true` (`t.Setenv`); el comportamiento strict viene de 003 (apoyado en 001). |
| **004 shared-text-propername** | dominios (`domain.Customer`, etc.) pueden depender de tipos/normalizaciones de 004. Bajo impacto: si los structs compilan, alcanza. |
| **023 be-wire-di** | indirecto: `NewHandler(...)` / `NewRepository(...)` deben tener la firma que asume el test. Mientras 001/002/009 dejen esas firmas estables, no hay choque. |

### Inciertas

- Si 001/002/009 se mergearon en `develop` con **firmas distintas** a las del SOURCE 777e5f6a (p. ej. otro
  orden de parámetros en `assertSupplyMovementReferencesActive` o en `Routes()`), algún test no compilará.
  **Comando para chequear:** `git grep -n "func assertSupplyMovementReferencesActive" develop -- internal/supply/`
  y comparar la firma con la usada en el test. Confianza: media.

## Bloquea-a

Ninguna feature. feature-025 es hoja del grafo. (Mejora la confianza para mergear 001/002/009, pero no es
bloqueante de ellas.)

## Artefactos compartidos

- **Archivos compartidos del repo:** NINGUNO tocado por esta feature (no toca `wire/*`, `cmd/api/*`,
  `go.mod`, `go.sum`, `Makefile`, `internal/shared/**`).
- **Tipos compartidos referenciados (read-only):** `contextkeys.{Actor,OrgID,Role,Scopes}`,
  `domainerr.IsConflict`, dominios por módulo.
- **Config/env:** los tests setean `TENANT_STRICT_MODE` vía `t.Setenv` (scope local al test, no toca config global).
- **Migraciones:** ninguna; sqlite in-memory.
- **APIs:** ejercitan rutas existentes, no las definen.

## Recomendación de orden

```
001 + 003  ──►  (tenant tests)
009        ──►  (archived-refs tests + GetArchivedParameters handler tests)
002        ──►  (work-order handler_test M + rutas archive/restore/hard)
                         │
                         ▼
                   feature-025 (este PR, o sus 3 sub-PRs)
```

Regla práctica: **025 va siempre al final** de su tren de dependencias. Nunca antes de 001/002/009.
