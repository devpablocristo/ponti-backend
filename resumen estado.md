# Estado Actual Unificado

## 1. Decisiones de negocio confirmadas
- La numeración de work orders con sufijos como `.1`, `.2`, `.3` es **válida** y debe conservarse.
- Una work order siempre pertenece a **un campo específico**.
- `Todos los campos` es solo un **filtro agregado**, no un valor válido al crear una work order.
- El workspace mínimo válido de la app es:
  - `customer`
  - `project`
  - `campaign`
- `field` es opcional y, cuando no se elige, significa **`Todos los campos`**.

## 2. Hotfix vigente
- Rama backend: `dashboard-fix`
- Rama frontend: `dashboard-fix`
- Base: `main`

### Qué corrige
- La card `Última orden de trabajo` del dashboard ahora muestra la **última work order creada**, no la última por `date`.
- Eso aplica en los dos casos:
  - proyecto + campaña + todos los campos
  - proyecto + campaña + un campo específico
- El BFF del frontend ya no deja el dashboard stale por cache.
- La fecha de la card quedó formateada correctamente.

### Criterio funcional
- Con `field` vacío:
  - la card muestra la última WO creada para ese `customer + project + campaign`, considerando **todos los campos**.
- Con `field` elegido:
  - la card muestra la última WO creada para ese campo puntual.

## 3. Qué se descartó
- Se descartó la línea de trabajo de “normalizar work orders a enteros”.
- Se descartó también la transición con `legacy_number`.
- Los PRs de normalización final dejaron de tener sentido porque el negocio confirmó que la numeración con punto es correcta.

## 4. Estado actual de work orders
- La creación de WO sigue permitiendo numeraciones como `1861.1`, `1860.5`, etc.
- Eso es correcto según la definición actual del negocio.
- No hay que introducir ninguna ingeniería para convertir esos números a enteros.

## 5. Regla de visualización y consulta
- Dashboard:
  - usa **última creada**
- Listado de Órdenes de Trabajo (`/admin/work-orders`):
  - debe usar también **última creada**
  - ya no “número más alto”

## 6. Cambios locales actualmente hechos y pendientes de commit

### Backend
- [internal/shared/handlers/workspace_filters.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/shared/handlers/workspace_filters.go)
  - valida que `customer_id + project_id + campaign_id` sean obligatorios para dashboard/reportes
- [internal/dashboard/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/dashboard/handler.go)
  - aplica esa validación al dashboard
- [internal/report/handler.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/report/handler.go)
  - aplica esa validación a reportes
- [internal/report/usecases.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/report/usecases.go)
- [internal/report/usecases/validators.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/report/usecases/validators.go)
  - alinean el criterio de workspace mínimo válido
- [internal/work-order/repository.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/work-order/repository.go)
  - el listado de work orders quedó ordenado por **última creada** (`id desc`)
- [internal/work-order/repository_list_test.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/work-order/repository_list_test.go)
  - test de regresión del orden del listado

### Frontend
- [ui/src/hooks/useWorkspaceFilters.ts](/home/pablo/Projects/Pablo/ponti/ponti-frontend/ui/src/hooks/useWorkspaceFilters.ts)
  - `customer + project + campaign` pasan a ser obligatorios para que el workspace sea utilizable
  - `field` queda opcional
  - `field` vacío vuelve siempre a `Todos los campos`
- [ui/src/hooks/useDashboard/index.ts](/home/pablo/Projects/Pablo/ponti/ponti-frontend/ui/src/hooks/useDashboard/index.ts)
  - limpia estado si el workspace no es válido
- [ui/src/hooks/useReporting/index.ts](/home/pablo/Projects/Pablo/ponti/ponti-frontend/ui/src/hooks/useReporting/index.ts)
  - limpia estado si el workspace no es válido
- [ui/src/pages/admin/dashboard/Dashboard.tsx](/home/pablo/Projects/Pablo/ponti/ponti-frontend/ui/src/pages/admin/dashboard/Dashboard.tsx)
  - solo consulta dashboard con `customer + project + campaign`
- [ui/src/pages/admin/reports/SummaryResultsReport.tsx](/home/pablo/Projects/Pablo/ponti/ponti-frontend/ui/src/pages/admin/reports/SummaryResultsReport.tsx)
- [ui/src/pages/admin/reports/ByFieldOrCropReport.tsx](/home/pablo/Projects/Pablo/ponti/ponti-frontend/ui/src/pages/admin/reports/ByFieldOrCropReport.tsx)
- [ui/src/pages/admin/reports/InvestorContributionReport.tsx](/home/pablo/Projects/Pablo/ponti/ponti-frontend/ui/src/pages/admin/reports/InvestorContributionReport.tsx)
  - alineados a workspace mínimo válido
- [ui/src/pages/admin/database/data-integrity/Integrity.tsx](/home/pablo/Projects/Pablo/ponti/ponti-frontend/ui/src/pages/admin/database/data-integrity/Integrity.tsx)
  - también queda alineado al workspace mínimo
- [api/src/routes/reports.ts](/home/pablo/Projects/Pablo/ponti/ponti-frontend/api/src/routes/reports.ts)
  - reenvía `customer_id`

## 7. Qué ya quedó probado
- `go test ./...`
- `go test ./internal/work-order/...`
- `npm --prefix api run build`
- `npm --prefix ui run typecheck`
- `npm --prefix ui run build`

### Prueba funcional real del dashboard
- Se creó una WO nueva para `project_id=30`.
- El dashboard pasó a mostrar esa WO como `Última orden de trabajo`.
- Esto confirmó que la card ya no usa “última por fecha operativa”, sino “última creada”.

## 8. Qué sigue pendiente

### Pendiente funcional/técnico
- Corregir el swap artificial de arriendo en:
  - [internal/dashboard/handler/dto/dashboard.go](/home/pablo/Projects/Pablo/ponti/ponti-backend/internal/dashboard/handler/dto/dashboard.go)
- Agregar tests de regresión reales del dashboard para:
  - última WO creada por proyecto
  - última WO creada por campo
  - validación del workspace mínimo
  - balance de gestión sin swap artificial

### Pendiente de integración
- Los cambios de workspace mínimo válido y el cambio de orden del listado están hechos localmente, pero todavía **no están commiteados/pusheados**.

## 9. Resumen corto
- El hotfix correcto es: **dashboard usa última WO creada**.
- La numeración con punto **se mantiene**.
- `legacy_number` y la normalización a enteros quedan descartados.
- La app debe trabajar con:
  - `customer + project + campaign` obligatorios
  - `field` opcional con default `Todos los campos`.
