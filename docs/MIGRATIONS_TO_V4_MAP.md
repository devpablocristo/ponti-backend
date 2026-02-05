TLDR: Mapeo legacy→v4 enfocado en datos en `public`. Las vistas y funciones viven en `v4_*` y se regeneran por migraciones; el único renombre de datos es `app_parameters`→`business_parameters`.

## Alcance
- La base **v4** es `v4-only` y se crea desde cero con las migraciones actuales.
- Los datos persistentes viven en `public`.
- Esquemas `v4_core`, `v4_ssot`, `v4_calc`, `v4_report` contienen funciones/vistas; no se migran como datos.

## Mapeo de esquemas lógicos
- Esquemas históricos de vistas (core/ssot/calc/report) → `v4_core`, `v4_ssot`, `v4_calc`, `v4_report`.
- Resultado: **no se copian datos** entre esquemas de vistas; se **recrean** por migraciones.

## Mapeo de tablas (public)
Nota: salvo donde se indique, el mapeo es **directo** (misma tabla y mismas columnas).

- `public.users` → `public.users` (sin cambios).
- `public.app_parameters` → `public.business_parameters` (**renombre**).
- `public.fx_rates` → `public.fx_rates` (sin cambios).
- `public.customers` → `public.customers` (sin cambios).
- `public.campaigns` → `public.campaigns` (sin cambios).
- `public.projects` → `public.projects` (sin cambios).
- `public.managers` → `public.managers` (sin cambios).
- `public.project_managers` → `public.project_managers` (sin cambios).
- `public.lease_types` → `public.lease_types` (sin cambios).
- `public.fields` → `public.fields` (sin cambios).
- `public.lots` → `public.lots` (sin cambios).
- `public.crops` → `public.crops` (sin cambios).
- `public.labor_types` → `public.labor_types` (sin cambios).
- `public.labor_categories` → `public.labor_categories` (sin cambios).
- `public.labors` → `public.labors` (sin cambios).
- `public.workorders` → `public.workorders` (sin cambios).
- `public.workorder_items` → `public.workorder_items` (sin cambios).
- `public.types` → `public.types` (sin cambios).
- `public.categories` → `public.categories` (sin cambios).
- `public.supplies` → `public.supplies` (sin cambios).
- `public.stocks` → `public.stocks` (sin cambios).
- `public.supply_movements` → `public.supply_movements` (sin cambios).
- `public.providers` → `public.providers` (sin cambios).
- `public.investors` → `public.investors` (sin cambios).
- `public.project_investors` → `public.project_investors` (sin cambios).
- `public.crop_commercializations` → `public.crop_commercializations` (sin cambios).
- `public.admin_cost_investors` → `public.admin_cost_investors` (sin cambios).
- `public.field_investors` → `public.field_investors` (sin cambios).
- `public.project_dollar_values` → `public.project_dollar_values` (sin cambios).
- `public.invoices` → `public.invoices` (sin cambios).

## Notas para volcado de datos
- **Renombre**: migrar `app_parameters` → `business_parameters` columna a columna.
- **Secuencias**: ajustar `*_id_seq` al `MAX(id)` luego del volcado.
- **FKs**: respetar orden de carga por dependencias (catálogos → entidades → relaciones).
- **Soft delete**: conservar `deleted_at` y `deleted_by` si existen.

## Orden sugerido de carga
1. Catálogos básicos: `users`, `business_parameters`, `fx_rates`, `customers`, `campaigns`, `managers`, `lease_types`, `crops`, `types`, `categories`, `providers`.
2. Núcleo: `projects`, `fields`, `lots`, `labor_types`, `labor_categories`, `labors`, `supplies`.
3. Movimientos/relaciones: `workorders`, `workorder_items`, `stocks`, `supply_movements`.
4. Inversores y comercialización: `investors`, `project_investors`, `field_investors`, `admin_cost_investors`, `crop_commercializations`, `project_dollar_values`, `invoices`.

## Transformación mínima necesaria
- `INSERT INTO public.business_parameters (...) SELECT ... FROM public.app_parameters`.
