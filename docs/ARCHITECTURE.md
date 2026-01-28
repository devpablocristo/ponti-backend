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

