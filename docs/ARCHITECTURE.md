## TLDR
- GORM solo para CRUD simple y relaciones básicas.
- SQL directo para reportes, agregaciones, vistas, CTEs y performance.
- Decisión documentada por módulo y por query crítica.

## Alcance
Este documento define cuándo usar GORM y cuándo usar SQL directo en este proyecto.

## Regla principal
- Usar **GORM** para operaciones CRUD simples.
- Usar **SQL directo** para consultas complejas o críticas.

## Usar GORM cuando
- CRUD básico por ID.
- Filtros simples sin agregaciones complejas.
- Transacciones simples de escritura.
- Relaciones básicas con `Preload` y sin explosión de joins.

## Usar SQL directo cuando
- Reportes, métricas, dashboards o vistas.
- Agregaciones con `GROUP BY`, `CTE`, subqueries complejas.
- Consultas que requieren índices específicos o tuning fino.
- Performance crítica o comportamiento exacto de SQL.

## Reglas operativas
- No mezclar GORM y SQL directo en la misma función sin necesidad.
- Si una query crítica usa `Raw()`, agregar comentario breve explicando por qué.
- Evitar `Preload` en cascada si no es imprescindible.

## Documentación mínima por módulo
Cada módulo debe indicar si:
- CRUD principal usa GORM.
- Reportes o métricas usan SQL directo.

## Ejemplo de decisión
- `internal/report/*`: SQL directo.
- `internal/customer/*`: GORM CRUD.

## AI (`InsightService` + `CopilotAgent`)
- Flujo: FE → BFF → Backend Go → Ponti AI.
- El FE no conoce claves; el Backend Go usa `X-SERVICE-KEY`.
- Ponti AI es READ-ONLY sobre dominio y solo escribe en `ai_*`.
- SQL en Ponti AI usa allowlist con `project_id` y `LIMIT` obligatorios.
- El backend expone proxy HTTP para:
  - `POST /api/v1/ai/insights/compute`
  - `GET /api/v1/ai/insights/summary`
  - `GET /api/v1/ai/insights/{entity_type}/{entity_id}`
  - `POST /api/v1/ai/insights/{insight_id}/actions`
  - `GET /api/v1/ai/copilot/insights/{insight_id}/explain`
  - `GET /api/v1/ai/copilot/insights/{insight_id}/why`
  - `GET /api/v1/ai/copilot/insights/{insight_id}/next-steps`

