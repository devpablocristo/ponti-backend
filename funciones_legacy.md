# Lista de Funciones Activas - Ponti Backend

**Fecha de generación:** $(date)  
**Base de datos:** ponti_api_db  
**Estado:** Migraciones v3 aplicadas (76-82)

## Resumen

- **Total de funciones:** 51
- **Funciones públicas (legacy):** 7
- **Funciones v3_calc (SSOT):** 44

---

## Funciones Públicas (Legacy)

Estas funciones son wrappers legacy que se mantienen para compatibilidad:

| Función | Tipo de Retorno | Descripción |
|---------|------------------|-------------|
| `calculate_campaign_closing_date` | date | Fecha de cierre de campaña |
| `calculate_cost_per_ha` | numeric | Costo por hectárea |
| `calculate_harvested_area` | numeric | Área cosechada |
| `calculate_labor_cost` | numeric | Costo de labor |
| `calculate_sowed_area` | numeric | Área sembrada |
| `calculate_supply_cost` | numeric | Costo de insumos |
| `calculate_yield` | numeric | Rendimiento |

---

## Funciones v3_calc (Single Source of Truth)

Estas son las funciones centralizadas del schema v3_calc que implementan el principio DRY:

### LAYER 1: Operaciones Matemáticas Seguras
| Función | Tipo de Retorno | Descripción |
|---------|------------------|-------------|
| `coalesce0` (double precision) | double precision | Convierte NULL a 0 (double) |
| `coalesce0` (numeric) | numeric | Convierte NULL a 0 (numeric) |
| `safe_div` | numeric | División segura (evita división por cero) |
| `safe_div_dp` | double precision | División segura (double precision) |
| `percentage` | numeric | Calcula porcentaje |
| `percentage_capped` | numeric | Calcula porcentaje con límite de 100% |

### LAYER 2: Conversiones por Hectárea
| Función | Tipo de Retorno | Descripción |
|---------|------------------|-------------|
| `per_ha` | numeric | Conversión por hectárea (numeric) |
| `per_ha_dp` | double precision | Conversión por hectárea (double precision) |

### LAYER 3: Cálculos del Dominio Agrícola
| Función | Tipo de Retorno | Descripción |
|---------|------------------|-------------|
| `dose_per_ha` | numeric | Dosis por hectárea |
| `seeded_area` | numeric | Área sembrada |
| `harvested_area` | numeric | Área cosechada |
| `yield_tn_per_ha_over_hectares` | numeric | Rendimiento tn/ha (base hectáreas) |
| `yield_tn_per_ha_over_harvested` | numeric | Rendimiento tn/ha (base área cosechada) |
| `labor_cost` | numeric | Costo de labor |
| `supply_cost` | numeric | Costo de insumos |
| `cost_per_ha` | numeric | Costo por hectárea |
| `income_net_total` | numeric | Ingreso neto total |
| `income_net_per_ha` | numeric | Ingreso neto por hectárea |
| `rent_per_ha` (integer) | double precision | Renta por hectárea (integer) |
| `rent_per_ha` (bigint) | double precision | Renta por hectárea (bigint) |
| `active_total_per_ha` | double precision | Activo total por hectárea |
| `operating_result_per_ha` | double precision | Resultado operativo por hectárea |
| `renta_pct` | double precision | Porcentaje de renta |
| `indifference_price_usd_tn` | double precision | Precio de indiferencia USD/tn |
| `units_per_ha` | numeric | Unidades por hectárea |
| `norm_dose` | numeric | Dosis normalizada |

### LAYER 4: Consultas a Nivel de Lote
| Función | Tipo de Retorno | Descripción |
|---------|------------------|-------------|
| `lot_hectares` | double precision | Hectáreas del lote |
| `lot_tons` | double precision | Toneladas del lote |
| `labor_cost_for_lot` | numeric | Costo de labor para lote |
| `supply_cost_for_lot` | double precision | Costo de insumos para lote |
| `direct_cost_for_lot` | double precision | Costo directo para lote |

### LAYER 5: Cálculos de Ingresos
| Función | Tipo de Retorno | Descripción |
|---------|------------------|-------------|
| `net_price_usd_for_lot` | numeric | Precio neto USD para lote |
| `income_net_total_for_lot` | numeric | Ingreso neto total para lote |
| `income_net_per_ha_for_lot` | double precision | Ingreso neto por hectárea para lote |

### LAYER 6: Agregaciones a Nivel de Proyecto
| Función | Tipo de Retorno | Descripción |
|---------|------------------|-------------|
| `total_hectares_for_project` | double precision | Total de hectáreas del proyecto |
| `admin_cost_per_ha_for_lot` | double precision | Costo de administración por hectárea para lote |
| `cost_per_ha_for_lot` | double precision | Costo por hectárea para lote |
| `rent_per_ha_for_lot` | double precision | Renta por hectárea para lote |
| `active_total_per_ha_for_lot` | double precision | Activo total por hectárea para lote |
| `operating_result_per_ha_for_lot` | double precision | Resultado operativo por hectárea para lote |
| `yield_tn_per_ha_for_lot` | double precision | Rendimiento tn/ha para lote |
| `seeded_area_for_lot` | numeric | Área sembrada para lote |
| `harvested_area_for_lot` | numeric | Área cosechada para lote |

### LAYER 7: Cálculos de Fechas de Campaña
| Función | Tipo de Retorno | Descripción |
|---------|------------------|-------------|
| `calculate_campaign_closing_date` | date | Fecha de cierre de campaña |

---

## Arquitectura v3

### Principios Aplicados

1. **Single Source of Truth (SSOT)**: Todas las funciones de cálculo están centralizadas en `v3_calc`
2. **DRY (Don't Repeat Yourself)**: No hay duplicación de lógica de cálculo
3. **Funciones Inmutables**: Funciones determinísticas y cacheables
4. **Operaciones Seguras**: Protección contra división por cero y valores NULL
5. **Separación por Capas**: Organización lógica de funciones por complejidad

### Uso en Vistas v3

Todas las vistas v3 (migraciones 78-82) usan **ÚNICAMENTE** funciones de `v3_calc`:

- `v3_workorder_metrics` y `v3_workorder_list`
- `v3_lot_metrics` y `v3_lot_list`
- `v3_labor_metrics` y `v3_labor_list`
- `v3_dashboard` (4 vistas: principal, contributions_progress, management_balance, crop_incidence)
- `v3_report_field_crop_metrics_view`

### Migraciones

- **76**: Elimina todo el esquema legacy
- **77**: Crea schema v3_calc con todas las funciones SSOT
- **78-82**: Crean vistas v3 que usan solo funciones v3_calc

---

## Notas Técnicas

- Las funciones públicas (legacy) se mantienen para compatibilidad
- Las funciones v3_calc son el futuro y deben usarse en todo código nuevo
- Todas las funciones v3_calc están optimizadas para PostgreSQL
- Los tipos de datos están optimizados (double precision para cálculos, numeric para precisión monetaria)
- Las funciones están documentadas con comentarios en español

---

*Documento generado automáticamente desde la base de datos ponti_api_db*
