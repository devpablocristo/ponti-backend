# Estructura de Migraciones Reorganizadas - Cálculos Ponti

## Resumen de Reorganización

Las migraciones para los cálculos Ponti han sido reorganizadas por entidad siguiendo las mejores prácticas del proyecto. Cada migración está claramente identificada con su entidad correspondiente.

## Estructura de Nomenclatura

### Formato de Archivos
```
db/migrations/{version}_{snake_case_title}.up.sql
db/migrations/{version}_{snake_case_title}.down.sql
```

### Convención de Nombres
- **Prefijo de versión:** Números secuenciales empezando en 000062
- **Entidad:** Nombre de la entidad en inglés (dollar, shared, workorder, labor, lot, project, dashboard)
- **Funcionalidad:** Descripción de lo que hace la migración
- **Sufijo:** `.up.sql` para aplicar, `.down.sql` para revertir

## Migraciones Implementadas

### 000062: DOLLAR - Tabla de Tipos de Cambio
**Archivos:**
- `000062_create_dollar_fx_rates_table.up.sql`
- `000062_create_dollar_fx_rates_table.down.sql`

**Funcionalidad:** Crear tabla `fx_rates` para conversiones USD/ARS
**Entidad:** dollar (Dólares)

**Contenido:**
- Tabla `fx_rates` con campos: `currency_pair`, `rate`, `effective_date`
- Índices únicos y de búsqueda por fecha
- Datos iniciales para USDARS

---

### 000063: SHARED - Índices de Soporte para Cálculos
**Archivos:**
- `000063_create_shared_calc_support_indexes.up.sql`
- `000063_create_shared_calc_support_indexes.down.sql`

**Funcionalidad:** Crear índices de soporte para optimizar cálculos
**Entidad:** shared (Compartido)

**Contenido:**
- Índices para `workorder_items`, `labors`, `lots`
- Índices para `crop_commercializations`, `projects`
- Todos los índices incluyen filtros `WHERE deleted_at IS NULL`

---

### 000064: WORKORDER - Vistas de Cálculo de Órdenes de Trabajo
**Archivos:**
- `000064_create_workorder_calc_views.up.sql`
- `000064_create_workorder_calc_views.down.sql`

**Funcionalidad:** Crear vistas para cálculos de workorders
**Entidad:** workorder (Órdenes de Trabajo)

**Contenido:**
- Vista `v_calc_workorders` con cálculos de labor + supplies
- Cálculo de costos totales por workorder
- Joins con `projects`, `fields`, `lots`, `crops`, `labors`, `categories`

---

### 000065: LABOR - Vistas de Cálculo de Labores
**Archivos:**
- `000065_create_labor_calc_views.up.sql`
- `000065_create_labor_calc_views.down.sql`

**Funcionalidad:** Crear vistas para cálculos de labors
**Entidad:** labor (Labores)

**Contenido:**
- Vista `v_calc_labors` con cálculos de IVA 10.5%
- Conversión USD a ARS usando tipos de cambio
- Joins con `projects` y `categories`

---

### 000066: LOT - Vistas de Cálculo de Lotes
**Archivos:**
- `000066_create_lot_calc_views.up.sql`
- `000066_create_lot_calc_views.down.sql`

**Funcionalidad:** Crear vistas para cálculos de lots
**Entidad:** lot (Lotes)

**Contenido:**
- Vista `v_calc_lots` con cálculos de rendimiento y economía
- Cálculo de ingresos netos por hectárea
- Joins con `fields`, `projects`, `crops`, `lease_types`

---

### 000067: PROJECT - Vistas de Consolidación de Proyectos
**Archivos:**
- `000067_create_project_rollup_views.up.sql`
- `000067_create_project_rollup_views.down.sql`

**Funcionalidad:** Crear vistas para consolidaciones de proyectos
**Entidad:** project (Proyectos)

**Contenido:**
- Vista `v_calc_project_costs` para consolidación de costos
- Vista `v_calc_project_economics` para consolidación de economía
- Agregaciones a nivel de proyecto

---

### 000068: DASHBOARD - Vista de Verificación de Cálculos
**Archivos:**
- `000068_create_dashboard_verification_view.up.sql`
- `000068_create_dashboard_verification_view.down.sql`

**Funcionalidad:** Crear vista de verificación para validar cálculos
**Entidad:** dashboard (Dashboard)

**Contenido:**
- Vista `v_calc_verification` con 8 verificaciones automáticas
- Sanity checks para todos los cálculos implementados
- Validación de fórmulas y resultados

---

## Dependencias y Orden de Ejecución

### Dependencias Directas
```
000062 (fx_rates) ← 000065 (labors - conversión ARS)
000063 (índices) ← 000064, 000065, 000066, 000067 (todas las vistas)
000064 (workorders) ← 000063 (índices)
000065 (labors) ← 000062 (fx_rates), 000063 (índices)
000066 (lots) ← 000063 (índices)
000067 (projects) ← 000064, 000066 (workorders, lots)
000068 (verificación) ← 000064, 000065, 000066, 000067 (todas las vistas)
```

### Orden de Ejecución Recomendado
1. **000062** - Crear tabla fx_rates
2. **000063** - Crear índices de soporte
3. **000064** - Crear vista de workorders
4. **000065** - Crear vista de labors
5. **000066** - Crear vista de lots
6. **000067** - Crear vistas de proyectos
7. **000068** - Crear vista de verificación

---

## Ventajas de la Reorganización

### 1. **Claridad de Entidad**
- Cada migración está claramente asociada a una entidad específica
- Fácil identificación de qué funcionalidad afecta cada migración
- Nomenclatura consistente y predecible

### 2. **Mantenibilidad**
- Separación clara de responsabilidades
- Fácil localización de migraciones por funcionalidad
- Rollback granular por entidad si es necesario

### 3. **Trazabilidad**
- Historial claro de cambios por entidad
- Fácil identificación de dependencias
- Documentación automática de la evolución del sistema

### 4. **Escalabilidad**
- Estructura preparada para futuras migraciones
- Patrón consistente para nuevas entidades
- Fácil integración con CI/CD

---

## Comandos de Gestión

### Verificar Estado Actual
```bash
# Ver versión actual
docker run --rm --network ponti-api_app-network -v $(pwd)/migrations:/migrations migrate/migrate:latest -path=/migrations -database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" version

# Ver historial de migraciones
docker run --rm --network ponti-api_app-network -v $(pwd)/migrations:/migrations migrate/migrate:latest -path=/migrations -database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" history
```

### Aplicar Migraciones
```bash
# Aplicar todas las migraciones pendientes
docker run --rm --network ponti-api_app-network -v $(pwd)/migrations:/migrations migrate/migrate:latest -path=/migrations -database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" up

# Aplicar migración específica
docker run --rm --network ponti-api_app-network -v $(pwd)/migrations:/migrations migrate/migrate:latest -path=/migrations -database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" up 1
```

### Revertir Migraciones
```bash
# Revertir última migración
docker run --rm --network ponti-api_app-network -v $(pwd)/migrations:/migrations migrate/migrate:latest -path=/migrations -database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" down 1

# Forzar versión específica (útil para resolver estados "dirty")
docker run --rm --network ponti-api_app-network -v $(pwd)/migrations:/migrations migrate/migrate:latest -path=/migrations -database "postgres://admin:admin@ponti-db:5432/ponti_api_db?sslmode=disable" force 67
```

---

## Consideraciones Técnicas

### Idempotencia
- Todas las migraciones usan `CREATE OR REPLACE VIEW`
- Índices usan `CREATE INDEX IF NOT EXISTS`
- Tablas usan `CREATE TABLE IF NOT EXISTS`

### Soft-Delete
- Todas las vistas respetan `WHERE deleted_at IS NULL`
- Índices parciales solo en registros activos
- Consistencia con la lógica de negocio existente

### Performance
- Índices optimizados para las consultas de cálculo
- Vistas materializadas donde sea apropiado
- Agregaciones eficientes con `GROUP BY`

---

*Documento de Estructura de Migraciones - Cálculos Ponti v1.0*
