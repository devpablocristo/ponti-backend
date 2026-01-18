# Plan de Migración v4 - Estado y Saneamiento

## Resumen Ejecutivo

Migración de vistas de reportes desde schema `public` (v3_*) hacia schema `v4_report` usando el patrón **Strangler Fig**.

**Estado actual:** FASE 1 y 2 completadas, pendiente saneamiento DRY  
**Fecha:** Enero 2025  
**Última actualización:** Diagnóstico DRY + consolidación SSOT investor (000327)

---

## 1. Estado de Migración

### ✅ Vistas Migradas a v4_report (21 vistas)

| Grupo | Vistas | Migración |
|-------|--------|-----------|
| **Field Crop** | `field_crop_metrics`, `field_crop_cultivos`, `field_crop_labores`, `field_crop_insumos`, `field_crop_economicos`, `field_crop_rentabilidad` | 000301-000324 |
| **Lot** | `lot_metrics`, `lot_list` | 000301-000323 |
| **Labor** | `labor_metrics`, `labor_list` | 000301-000317 |
| **Workorder** | `workorder_metrics` | 000301 |
| **Summary** | `summary_results` | 000322 |
| **Dashboard** | `dashboard_metrics`, `dashboard_contributions_progress`, `dashboard_management_balance`, `dashboard_crop_incidence`, `dashboard_operational_indicators` | 000325 |
| **Investor** | `investor_project_base`, `investor_contribution_categories`, `investor_distributions`, `investor_contribution_data` | 000326 |

### ✅ Vistas SSOT en v4_calc (3 vistas)

| Vista | Propósito | Migración |
|-------|-----------|-----------|
| `lot_base_costs` | Costos base por lote | 000303 |
| `lot_base_income` | Ingresos base por lote | 000304 |
| `investor_real_contributions` | **SSOT de aportes reales por inversor** | 000327 |

### ❌ Vistas Legacy v3_* a Eliminar (22 vistas)

```
v3_dashboard_contributions_progress    v3_report_field_crop_cultivos
v3_dashboard_crop_incidence           v3_report_field_crop_economicos
v3_dashboard_management_balance       v3_report_field_crop_insumos
v3_dashboard_metrics                  v3_report_field_crop_labores
v3_dashboard_operational_indicators   v3_report_field_crop_metrics
v3_investor_contribution_data_view    v3_report_field_crop_rentabilidad
v3_labor_list                         v3_report_investor_contribution_categories
v3_labor_metrics                      v3_report_investor_distributions
v3_lot_list                          v3_report_investor_project_base
v3_lot_metrics                       v3_report_summary_results_view
v3_workorder_list                    v3_workorder_metrics
```

---

## 2. Diagnóstico DRY (Don't Repeat Yourself)

### Métricas Actuales

```
📊 ESTADO DRY - Enero 2025

Funciones SSOT totales:    162
Funciones DUPLICADAS:       62 (38%)  ← PROBLEMA
Vistas v4_report:           21 ✅
Vistas v3_* legacy:         22 ❌
Vistas SSOT en v4_calc:      3 ✅
```

### 🚨 Problema: 62 Funciones Duplicadas

Las mismas funciones existen en múltiples schemas:

| Función | Schemas donde existe |
|---------|---------------------|
| `rent_per_ha` | v3_calc (x2), v3_core_ssot (x2) |
| `rent_per_ha_for_lot` | v3_calc, v3_lot_ssot |
| `renta_pct` | v3_calc, v3_lot_ssot |
| `cost_per_ha_for_lot` | v3_calc, v3_lot_ssot |
| `direct_cost_for_lot` | v3_calc, v3_lot_ssot |
| `admin_cost_per_ha_for_lot` | v3_calc, v3_lot_ssot |
| `active_total_per_ha` | v3_calc, v3_core_ssot |
| `coalesce0` | v3_calc (x2), v3_core_ssot (x2) |
| `harvested_area` | v3_calc, v3_core_ssot |
| `dollar_average_for_month` | v3_calc, v3_core_ssot |
| `agrochemicals_invested_for_project_mb` | v3_calc, v3_dashboard_ssot |
| `direct_costs_total_for_project` | v3_calc, v3_dashboard_ssot |
| ... y 50 más | |

### Distribución de Funciones por Schema

| Schema | Funciones | Rol |
|--------|-----------|-----|
| `v3_calc` | 76 | ❌ Muchas son copias de otros schemas |
| `v3_core_ssot` | 30 | Funciones genéricas |
| `v3_lot_ssot` | 26 | Funciones de lote |
| `v3_dashboard_ssot` | 25 | Funciones de proyecto/dashboard |
| `v3_report_ssot` | 3 | Funciones de reporte |
| `v3_workorder_ssot` | 2 | Funciones de workorder |

---

## 3. Bugs Corregidos Durante Migración

| Bug | Descripción | Migración |
|-----|-------------|-----------|
| Bug del dólar | Labores multiplicaba precio × dólar dos veces | 000317 |
| Arriendo inconsistente | Mostraba valor fijo pero restaba total | 000320 |
| Total Activo | Usaba arriendo fijo en vez del configurado | 000321 |
| Summary duplica cálculos | Recalculaba en vez de agregar desde field_crop | 000322 |
| Renta % incorrecta | Numerador/denominador usaban arriendo diferente | 000322 |
| lot_list columnas faltantes | 000318 eliminó sowed_area_ha, harvested_area_ha | 000323 |
| Summary arriendo fijo | Mostraba 119,770 en vez de 161,773 | 000324 |
| Performance timeout | field_crop_metrics tardaba ~10 seg | 000324 |
| Stocks: filtro close_date | query.Where() no se asignaba | Código Go |
| Stocks: Rubro mostraba Tipo | Usaba Type.Name en vez de CategoryName | Código Go |
| Stocks: Consumed = 0 | Faltaba Preload de Category | Código Go |
| **Investor: Dashboard ≠ Informe** | Mismo cálculo, diferentes resultados | **000327** |

---

## 4. Arquitectura Actual

```
┌─────────────────────────────────────────────────────────────────┐
│                         CAPA GO                                  │
│  - Lee de vistas v4_report.* (feature flag REPORT_SCHEMA)       │
│  - NO hace cálculos, solo lee                                   │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                    SCHEMA: v4_report (21 vistas)                │
│  Vistas de reporte para consumo de la API                       │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                    SCHEMA: v4_calc (3 vistas SSOT)              │
│  - lot_base_costs                                               │
│  - lot_base_income                                              │
│  - investor_real_contributions  ← NUEVO 000327                  │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│              SCHEMAS: v3_*_ssot (162 funciones)                 │
│  - v3_lot_ssot.* (26 funciones)                                │
│  - v3_core_ssot.* (30 funciones)                               │
│  - v3_dashboard_ssot.* (25 funciones)                          │
│  - v3_calc.* (76 funciones) ← 62 DUPLICADAS                    │
└─────────────────────────────────────────────────────────────────┘
                              ↑
┌─────────────────────────────────────────────────────────────────┐
│                    SCHEMA: public                               │
│  - Tablas de datos (lots, fields, projects, etc.)              │
│  - 22 vistas v3_* legacy (A ELIMINAR)                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 5. Plan de Saneamiento DRY

### Fase A: Consolidar Funciones SSOT (Prioridad Alta)

**Objetivo:** Eliminar 62 funciones duplicadas, dejar UNA por cálculo.

| Paso | Acción | Resultado |
|------|--------|-----------|
| 1 | Inventariar qué funciones de v3_calc son copias | Lista de funciones a eliminar |
| 2 | Elegir schema canónico por dominio | v3_lot_ssot, v3_core_ssot, v3_dashboard_ssot |
| 3 | Actualizar vistas que usan v3_calc.* | Cambiar a schema canónico |
| 4 | DROP funciones duplicadas de v3_calc | ~62 funciones menos |
| 5 | Documentar cada función con COMMENT | Autodocumentación |

**Schemas canónicos propuestos:**
- `v3_lot_ssot` → Cálculos de lote (rent, cost, income, area)
- `v3_core_ssot` → Funciones genéricas (safe_div, percentage, coalesce0)
- `v3_dashboard_ssot` → Cálculos de proyecto (totals, aggregations)

### Fase B: Eliminar Vistas Legacy (Prioridad Media)

**Objetivo:** Limpiar schema `public` de las 22 vistas v3_*.

| Paso | Acción |
|------|--------|
| 1 | Verificar que REPORT_SCHEMA=v4_report en producción |
| 2 | Buscar referencias a v3_* en código Go |
| 3 | Quitar feature flag, hardcodear v4_report |
| 4 | DROP VIEW v3_* (22 vistas) |
| 5 | Eliminar código muerto en Go |

### Fase C: Estandarizar Naming (Prioridad Baja)

**Objetivo:** Todo en inglés snake_case.

| Actual (español) | Propuesto (inglés) |
|------------------|-------------------|
| `superficie_total` | `total_area_ha` |
| `rendimiento_tn_ha` | `yield_tons_ha` |
| `gastos_directos_usd` | `direct_costs_usd` |
| `arriendo_usd` | `rent_usd` |
| `resultado_operativo_usd` | `operating_result_usd` |

---

## 6. Migraciones Aplicadas

| Número | Descripción |
|--------|-------------|
| 000300-000316 | Creación inicial de schemas y vistas v4 |
| 000317 | Fix bug dólar en labor_list |
| 000318 | Fix arriendo mostrar TOTAL |
| 000320 | Recrear todas las vistas field_crop |
| 000321 | Fix Total Activo usa arriendo configurado |
| 000322 | summary_results agrega desde field_crop (SSOT) |
| 000323 | Fix lot_list columnas faltantes |
| 000324 | Reescribe field_crop_metrics (10x más rápido) |
| 000325 | Dashboard migrado a v4 (5 vistas) |
| 000326 | Investor migrado a v4 (4 vistas + fix aportes reales) |
| **000327** | **Consolida investor contributions en SSOT** |

---

## 7. Feature Flag

```bash
# En GCP Cloud Run
REPORT_SCHEMA=v4_report

# Comportamiento
REPORT_SCHEMA=""          → usa v3_* (legacy)
REPORT_SCHEMA="v4_report" → usa v4_report.* (actual)
```

---

## 8. Comandos Útiles

```bash
# Aplicar migraciones
make migrate-up

# Ver vistas en v4_report
psql -c "SELECT viewname FROM pg_views WHERE schemaname = 'v4_report';"

# Ver funciones duplicadas
psql -c "
SELECT p.proname, array_agg(n.nspname)
FROM pg_proc p
JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname IN ('v3_calc', 'v3_lot_ssot', 'v3_core_ssot')
GROUP BY p.proname
HAVING COUNT(DISTINCT n.nspname) > 1;
"

# Contar funciones por schema
psql -c "
SELECT n.nspname, COUNT(*)
FROM pg_proc p
JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname LIKE 'v3_%'
GROUP BY n.nspname;
"
```

---

## 9. Checklist de Calidad Final

### Base de Datos

- [x] Todas las vistas migradas a v4_report
- [x] Feature flag funcionando
- [x] Bugs heredados corregidos
- [ ] Funciones SSOT sin duplicados (62 pendientes)
- [ ] Vistas v3_* eliminadas (22 pendientes)
- [ ] Naming en inglés consistente

### Código Go

- [x] Usa feature flag para seleccionar schema
- [x] No hace cálculos, solo lee
- [ ] Referencias a v3_* eliminadas
- [ ] Feature flag removido (hardcoded v4)

---

## 10. Próximos Pasos

1. **Estabilizar** (1-2 semanas)
   - Usar la app en producción con v4
   - Detectar bugs si hay

2. **Fase A: Consolidar SSOT**
   - Eliminar 62 funciones duplicadas
   - Crear migración de consolidación

3. **Fase B: Eliminar Legacy**
   - DROP 22 vistas v3_*
   - Quitar feature flag

4. **Fase C: Naming** (opcional)
   - Renombrar columnas español → inglés
   - Actualizar código Go y frontend
