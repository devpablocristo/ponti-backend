# Migración V4 - Estado Final

## TLDR
- Estado final v4 definido en `migrations`.
- Schemas v4: `v4_core`, `v4_ssot`, `v4_calc`, `v4_report`.
- Vistas v4_report finales listadas por dominio.
- Config vigente: `REPORT_SCHEMA=v4_report`.

---

## Alcance

- Solo estado final v4.
- Sin cambios de naming.
- Secciones divididas por partes coherentes (schemas, funciones, vistas, migraciones).

---

## Arquitectura final (v4)

```
┌─────────────────────────────────────────────────────────────┐
│                         CAPA GO                             │
│   Solo lectura de v4_report.*                               │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                   v4_report.*                               │
│   Vistas finales de reporte                                 │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    v4_calc.*                                │
│   Vistas de cálculo consolidadas                            │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    v4_ssot.*                                │
│   Funciones SSOT / wrappers                                 │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    v4_core.*                                │
│   Funciones matemáticas puras                               │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                    public.*                                 │
│   Tablas de datos                                            │
└─────────────────────────────────────────────────────────────┘
```

---

## Schemas v4 (estado final)

### `v4_core`
- Funciones matemáticas puras.
- Archivo base: `000301_create_v4_core_functions.up.sql`.

### `v4_ssot`
- Wrappers y funciones SSOT.
- Archivos base: `000302_create_v4_ssot_wrappers.up.sql`, `000308_add_v4_ssot_workorder_wrappers.up.sql`.

### `v4_calc`
- Vistas de cálculo consolidadas.
- Archivos base y consolidaciones: `000303` a `000348`.

### `v4_report`
- Vistas finales de reporte para Go.
- Archivos base: `000305` a `000316`, más extensiones `000322`, `000325`, `000326`.

---

## Vistas v4_report (finales, por dominio)

| Dominio | Vistas |
|---------|--------|
| Field Crop | `field_crop_metrics`, `field_crop_cultivos`, `field_crop_labores`, `field_crop_insumos`, `field_crop_economicos`, `field_crop_rentabilidad` |
| Lot | `lot_metrics`, `lot_list` |
| Labor | `labor_metrics`, `labor_list` |
| Workorder | `workorder_metrics` |
| Summary | `summary_results` |
| Dashboard | `dashboard_*` |
| Investor | `investor_*` |

---

## Migraciones v4 (agrupadas por sección)

### Schemas
- `000300_create_v4_schemas.up.sql`

### Core (funciones puras)
- `000301_create_v4_core_functions.up.sql`

### SSOT / Wrappers
- `000302_create_v4_ssot_wrappers.up.sql`
- `000308_add_v4_ssot_workorder_wrappers.up.sql`

### Calc (base)
- `000303_create_v4_calc_lot_base_costs.up.sql`
- `000304_create_v4_calc_lot_base_income.up.sql`

### Report (vistas finales)
- `000305_create_v4_report_lot_metrics.up.sql`
- `000306_create_v4_report_labor_list.up.sql`
- `000307_create_v4_report_lot_list.up.sql`
- `000309_create_v4_report_workorder_metrics.up.sql`
- `000310_create_v4_report_labor_metrics.up.sql`
- `000311_create_v4_report_field_crop_economicos.up.sql`
- `000312_create_v4_report_field_crop_rentabilidad.up.sql`
- `000313_create_v4_report_field_crop_cultivos.up.sql`
- `000314_create_v4_report_field_crop_labores.up.sql`
- `000315_create_v4_report_field_crop_insumos.up.sql`
- `000316_create_v4_report_field_crop_metrics.up.sql`
- `000322_create_v4_report_summary_results.up.sql`
- `000325_create_v4_dashboard_views.up.sql`
- `000326_create_v4_investor_views.up.sql`

### Fixes y consolidaciones (v4)
- `000317_fix_v4_labor_list_dollar_bug.up.sql`
- `000318_fix_v4_rent_show_total.up.sql`
- `000323_fix_v4_lot_list_missing_columns.up.sql`
- `000324_rewrite_field_crop_metrics_fast.up.sql`
- `000327_consolidate_investor_contributions_ssot.up.sql`
- `000330_fix_seeded_area_and_investor_per_ha_ssot.up.sql`
- `000331_fix_field_crop_per_ha_use_surface_total.up.sql`
- `000332_fix_dashboard_rent_fixed_ssot.up.sql`
- `000340_consolidate_field_crop_lot_base.up.sql`
- `000341_consolidate_field_crop_metrics_lot_base.up.sql`
- `000342_consolidate_field_crop_aggregated_costs.up.sql`
- `000343_consolidate_field_crop_supply_costs.up.sql`
- `000344_consolidate_field_crop_labor_costs.up.sql`
- `000345_consolidate_field_crop_metrics_costs.up.sql`
- `000346_consolidate_field_crop_metrics_aggregated.up.sql`
- `000347_consolidate_field_crop_totals.up.sql`
- `000348_consolidate_dashboard_supply_costs.up.sql`

---

## Migraciones baseline (migracions-2)

Estructura final por partes coherentes, sin cambiar naming:

1. `000001_create_schemas.(up|down).sql`
2. `000002_create_types_tables_indexes.(up|down).sql`
3. `000003_create_functions_ssot.(up|down).sql`
4. `000004_create_views_v3_v4.(up|down).sql` (contenido final v4)

---

## Configuración vigente

```
REPORT_SCHEMA=v4_report
```
