TLDR: Este documento es el Paso 1: inventario completo, diagnóstico y plan base para reconstruir migraciones sin romper compatibilidad.

# Paso 1 — Inventario y diagnóstico (migrations/)

## Alcance y supuestos
- Fuente de verdad: `migrations/` (no hay DB conectada).
- Objetivo: reconstrucción “perfecta” sin breaking changes en STG/PROD.
- Regla: no modificar migraciones existentes; el nuevo set irá en `migrations_v4/`.

## Orden actual de migraciones (as-is)
1. `000001_core_extensions` (up/down)
2. `000002_core_base_tables` (up/down)
3. `000003_core_triggers_functions` (up/down)
4. `000010_projects_tables` (up/down)
5. `000020_fields_lots_tables` (up/down)
6. `000030_crops_tables` (up/down)
7. `000040_workorders_labors_tables` (up/down)
8. `000050_supplies_inventory_tables` (up/down)
9. `000060_investors_commercialization_tables` (up/down)
10. `000070_constraints_fks_indexes` (up/down)
11. `000080_v4_schemas` (up/down)
12. `000090_v4_core_functions` (up/down)
13. `000100_v4_ssot_functions` (up/down)
14. `000110_v4_calc_views` (up/down)
15. `000120_v4_report_views` (up/down)
16. `000130_core_unaccent_extension` (up/down)
17. `000131_business_parameters_audit_columns` (up/down)
18. `000132_report_stock_consumed_by_supply` (up/down)

## Qué crea cada migración (resumen por archivo)

### 000001_core_extensions
- Extensiones: `unaccent`, `pg_trgm`.

### 000002_core_base_tables
- Tablas: `public.users`, `public.business_parameters`, `public.fx_rates`.
- Secuencias y defaults de IDs.
- Índice único en `fx_rates` por `(currency_pair, effective_date)`.

### 000003_core_triggers_functions
- Funciones `public.get_business_parameter*`, `public.get_iva_percentage`, `public.get_campaign_closure_days`, `public.get_default_fx_rate`.
- Trigger `set_timestamp` sobre `public.users` y función `public.update_timestamp()`.

### 000010_projects_tables
- Tablas: `public.customers`, `public.campaigns`, `public.projects`, `public.managers`, `public.project_managers`.
- Secuencias y defaults.
- `projects.admin_cost` (numeric(15,3)) y `projects.planned_cost` (numeric(12,2)).

### 000020_fields_lots_tables
- Tablas: `public.lease_types`, `public.fields`, `public.lots`, `public.lot_dates`.
- Secuencias y defaults.

### 000030_crops_tables
- Tabla: `public.crops` + secuencia.

### 000040_workorders_labors_tables
- Tablas: `public.labor_types`, `public.labor_categories`, `public.labors`, `public.workorders`, `public.workorder_items`.
- Secuencias y defaults.

### 000050_supplies_inventory_tables
- Tablas: `public.types`, `public.categories`, `public.supplies`, `public.stocks`, `public.supply_movements`, `public.providers`.
- Secuencias y defaults.
- Check `chk_supply_movements_movement_type`.

### 000060_investors_commercialization_tables
- Tablas: `public.investors`, `public.project_investors`, `public.crop_commercializations`, `public.admin_cost_investors`, `public.field_investors`, `public.project_dollar_values`, `public.invoices`.
- Secuencias y defaults.

### 000070_constraints_fks_indexes
- PKs y UNIQUEs para la mayoría de tablas.
- FKs cruzadas (usuarios/auditoría, relaciones de negocio).
- Índices de performance (muchos parciales por `deleted_at IS NULL`).

### 000080_v4_schemas
- Schemas: `v4_core`, `v4_ssot`, `v4_calc`, `v4_report`.

### 000090_v4_core_functions
- Funciones utilitarias (`coalesce0`, `safe_div`, `percentage`, `per_ha`, etc.).
- Cálculos de campaña, costos y rentas.
- Funciones de dólar mensual.
- Funciones de áreas y rendimiento.

### 000100_v4_ssot_functions
- Funciones SSOT para métricas por lote/proyecto.
- Funciones para dashboards, costos, rentas, y aportes de inversores.

### 000110_v4_calc_views
- Vistas de cálculo (`workorder_metrics`, `lot_base_costs`, `field_crop_*`, `investor_*`).

### 000120_v4_report_views
- Vistas de reportes (`lot_metrics`, `labor_list`, `field_crop_*`, dashboards, investor_*).
- Wrapper `v4_report.workorder_metrics`.

### 000130_core_unaccent_extension
- Extensión `unaccent` (duplicada respecto a `000001`).

### 000131_business_parameters_audit_columns
- Agrega columnas de auditoría faltantes en `public.business_parameters`.

### 000132_report_stock_consumed_by_supply
- Vista `v4_report.stock_consumed_by_supply`.

## Objetos por schema (inventario exhaustivo)

### Schema `public`
**Tablas core**
- `users`
- `business_parameters`
- `fx_rates`

**Tablas operativas**
- `customers`
- `campaigns`
- `projects`
- `managers`
- `project_managers`
- `lease_types`
- `fields`
- `lots`
- `lot_dates`
- `crops`
- `labor_types`
- `labor_categories`
- `labors`
- `workorders`
- `workorder_items`
- `types`
- `categories`
- `supplies`
- `stocks`
- `supply_movements`
- `providers`
- `investors`
- `project_investors`
- `crop_commercializations`
- `admin_cost_investors`
- `field_investors`
- `project_dollar_values`
- `invoices`

**Funciones**
- `public.get_business_parameter(varchar)`
- `public.get_business_parameter_decimal(varchar)`
- `public.get_business_parameter_integer(varchar)`
- `public.get_iva_percentage()`
- `public.get_campaign_closure_days()`
- `public.get_default_fx_rate()`
- `public.update_timestamp()`

**Triggers**
- `set_timestamp` en `public.users`.

### Schema `v4_core`
Funciones utilitarias de cálculo y normalización:
- `coalesce0(numeric)`
- `coalesce0(double precision)`
- `safe_div(numeric, numeric)`
- `safe_div_dp(double precision, double precision)`
- `percentage(numeric, numeric)`
- `percentage_capped(numeric, numeric)`
- `percentage_rounded(numeric, numeric)`
- `per_ha(numeric, numeric)`
- `per_ha_dp(double precision, double precision)`
- `per_ha(double precision, numeric)`
- `per_ha(numeric, double precision)`
- `units_per_ha(numeric, numeric)`
- `dose_per_ha(numeric, numeric)`
- `norm_dose(numeric, numeric)`
- `calculate_campaign_closing_date(date)`
- `get_project_dollar_value(bigint, varchar)`
- `dollar_average_for_month(bigint, date)`
- `seeded_area(date, numeric)`
- `harvested_area(numeric, numeric)`
- `yield_tn_per_ha_over_hectares(numeric, numeric)`
- `yield_tn_per_ha_over_harvested(numeric, numeric)`
- `labor_cost(numeric, numeric)`
- `supply_cost(numeric, numeric, numeric)`
- `cost_per_ha(numeric, numeric)`
- `income_net_total(numeric, numeric)`
- `income_net_per_ha(numeric, numeric)`
- `rent_per_ha(integer, numeric, numeric, numeric, numeric, numeric)`
- `rent_per_ha(bigint, numeric, numeric, numeric, numeric, numeric)`
- `calculate_rent_per_ha(numeric)`
- `active_total_per_ha(numeric, numeric, numeric)`
- `operating_result_per_ha(numeric, numeric)`
- `indifference_price_usd_tn(numeric, numeric)`

### Schema `v4_ssot`
Funciones SSOT (resumen):
- Métricas por lote: hectares, tons, costos, insumos, áreas, rentas, ingresos.
- Métricas por proyecto: costos directos, costos totales, rentas, admin, fechas, stocks.
- Funciones de contribuciones y aportes por inversor.

### Schema `v4_calc`
Vistas:
- `workorder_metrics`
- `workorder_metrics_raw`
- `lot_base_costs`
- `lot_base_income`
- `field_crop_lot_base`
- `field_crop_supply_costs_by_lot`
- `field_crop_labor_costs_by_lot`
- `field_crop_aggregated`
- `field_crop_metrics_lot_base`
- `field_crop_metrics_aggregated`
- `dashboard_fertilizers_invested_by_project`
- `dashboard_supply_costs_by_project`
- `investor_contribution_categories`
- `investor_real_contributions`

### Schema `v4_report`
Vistas:
- `lot_metrics`
- `lot_list`
- `labor_list`
- `labor_metrics`
- `field_crop_cultivos`
- `field_crop_economicos`
- `field_crop_insumos`
- `field_crop_labores`
- `field_crop_metrics`
- `field_crop_rentabilidad`
- `summary_results`
- `dashboard_management_balance`
- `dashboard_metrics`
- `dashboard_crop_incidence`
- `dashboard_operational_indicators`
- `investor_project_base`
- `investor_contribution_categories`
- `investor_distributions`
- `investor_contribution_data`
- `dashboard_contributions_progress`
- `workorder_list`
- `workorder_metrics` (wrapper)
- `stock_consumed_by_supply`

## Dependencias (grafo textual)

### Base
- Tablas `public.*` son base de todo.
- Constraints/FKs/índices dependen de las tablas.

### v4_core
- Depende de: `public.business_parameters`, `public.project_dollar_values`, `public.projects`.
- Es dependencia de `v4_ssot`, `v4_calc`, `v4_report`.

### v4_ssot
- Depende de: `public.*` (workorders, workorder_items, supplies, lots, fields, projects, crops, investors, supply_movements, stocks, categories).
- Depende de `v4_core` (safe_div, per_ha, rent_per_ha, etc.).
- Es dependencia de `v4_calc` y `v4_report`.

### v4_calc
- Depende de `public.*` y `v4_ssot` y `v4_core`.

### v4_report
- Depende de `v4_calc`, `v4_ssot`, `v4_core` y tablas `public.*`.

## Diagnóstico de problemas (con severidad)

### Alta
- **Inconsistencia de nombres**: `sowed_area_ha`, `seeded_area_ha`, `sown_area_ha` para la misma métrica en `v4_calc` y `v4_report`. Riesgo de consumo inconsistente.

### Media
- **UNIQUE duplicado** en `users.username` (dos constraints distintos).
- **Duplicación de extensión** `unaccent` (`000001` y `000130`).

### Media/Baja
- `public.update_timestamp` solo aplica a `users`, el resto de tablas no tiene trigger de update automático.
- Down migrations eliminan objetos en orden amplio; no hay garantía de reversibilidad completa por dependencias de vistas si se reordenan.

### Observaciones de coherencia
- Varios índices parciales por `deleted_at IS NULL` (bien), pero no está sistematizado en naming.
- Tipos numéricos mezclan `numeric(12,2)` y `numeric(18,6)` para costos/cantidades; conviene estandarizar según dominio.

## Estándares propuestos (base para Paso 2)

### Naming
- Constraints: `pk_<table>`, `fk_<table>_<ref>`, `uq_<table>_<cols>`, `idx_<table>_<cols>[_where]`.
- Índices parciales `..._notdel`.
- Métricas de área sembrada: **canonical** `seeded_area_ha`.

### Auditoría
- `created_at`, `updated_at`, `deleted_at` con TZ.
- `created_by`, `updated_by`, `deleted_by` (FK a `users` cuando aplique).

### Soft delete
- Estándar de `deleted_at IS NULL` en índices y vistas.

### Extensiones
- Extensiones sólo en una migración dedicada (idempotente).

### Compatibilidad
- Para cambios de naming: vistas wrapper y/o columnas alias (sin romper consumidores).
- Prohibido renombrar/borrar sin capa de compatibilidad.

## Plan inmediato (para Paso 2)
1. Diseñar nuevo set de migraciones por dominio.
2. Definir orden estable (extensiones → tablas → constraints → schemas → funciones → vistas).
3. Definir capa de compatibilidad para `sowed/sown` → `seeded`.
4. Preparar scripts de verificación (db reset, migrate, validate, snapshot, diff).

---

# Paso 2 — Diseño del nuevo set “perfecto”

## Estructura propuesta (nuevo directorio `migrations_v4/`)
Orden estable:
1. Extensiones
2. Tablas core
3. Tablas de dominio
4. Constraints/FKs/índices
5. Schemas v4
6. Funciones v4_core
7. Funciones v4_ssot
8. Vistas v4_calc
9. Vistas v4_report
10. Compatibilidad (wrappers/aliases)

## Índice de migraciones nuevas (nombres exactos)
1. `000001_extensions.up.sql` / `000001_extensions.down.sql`
   - `unaccent`, `pg_trgm` (idempotentes).

2. `000010_core_tables.up.sql` / `000010_core_tables.down.sql`
   - `public.users`, `public.business_parameters`, `public.fx_rates` + secuencias + defaults.

3. `000015_core_functions_triggers.up.sql` / `000015_core_functions_triggers.down.sql`
   - Funciones core de parámetros + trigger `updated_at` para `users`.

4. `000020_projects_tables.up.sql` / `000020_projects_tables.down.sql`
   - `customers`, `campaigns`, `projects`, `managers`, `project_managers` + secuencias.

5. `000030_fields_lots_tables.up.sql` / `000030_fields_lots_tables.down.sql`
   - `lease_types`, `fields`, `lots`, `lot_dates` + secuencias.

6. `000040_crops_tables.up.sql` / `000040_crops_tables.down.sql`
   - `crops` + secuencia.

7. `000050_workorders_labors_tables.up.sql` / `000050_workorders_labors_tables.down.sql`
   - `labor_types`, `labor_categories`, `labors`, `workorders`, `workorder_items` + secuencias.

8. `000060_supplies_inventory_tables.up.sql` / `000060_supplies_inventory_tables.down.sql`
   - `types`, `categories`, `supplies`, `stocks`, `supply_movements`, `providers` + secuencias + checks.

9. `000070_investors_commercialization_tables.up.sql` / `000070_investors_commercialization_tables.down.sql`
   - `investors`, `project_investors`, `crop_commercializations`,
     `admin_cost_investors`, `field_investors`, `project_dollar_values`,
     `invoices` + secuencias.

10. `000080_constraints_fks_indexes.up.sql` / `000080_constraints_fks_indexes.down.sql`
   - PK/UNIQUE/FK/INDEX con naming consistente y sin redundancias.

11. `000090_v4_schemas.up.sql` / `000090_v4_schemas.down.sql`
    - `v4_core`, `v4_ssot`, `v4_calc`, `v4_report`.

12. `000100_v4_core_functions.up.sql` / `000100_v4_core_functions.down.sql`
    - Todas las funciones utilitarias y de cálculo.

13. `000110_v4_ssot_functions.up.sql` / `000110_v4_ssot_functions.down.sql`
    - Funciones SSOT.

14. `000120_v4_calc_views.up.sql` / `000120_v4_calc_views.down.sql`
    - Vistas de cálculo.

15. `000130_v4_report_views.up.sql` / `000130_v4_report_views.down.sql`
    - Vistas de reporte.

16. `000140_compat_seeded_area_aliases.up.sql` / `000140_compat_seeded_area_aliases.down.sql`
    - Compatibilidad para `sowed_area_ha` y `sown_area_ha`.
    - Canonical: `seeded_area_ha`.
    - Implementación prevista: vistas wrapper y/o columnas alias.

17. `000000_baseline_schema.up.sql` / `000000_baseline_schema.down.sql`
    - Baseline del schema actual (snapshot) para adopción en STG/PROD.
