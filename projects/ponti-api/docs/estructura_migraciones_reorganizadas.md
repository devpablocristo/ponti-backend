# Estructura de Migraciones Reorganizadas - Cálculos de Ponti

## 📋 **Resumen de la Nueva Organización**

Las migraciones han sido reorganizadas por **entidad** y **funcionalidad**, siguiendo una estructura más lógica y mantenible.

---

## 🏗️ **Estructura de Migraciones (000062-000069)**

### **000062 - WORKORDERS (Órdenes de Trabajo)**
**Archivos:**
- `000062_create_workorder_calc_views.up.sql`
- `000062_create_workorder_calc_views.down.sql`

**Funcionalidad:** Cálculos de costos por orden de trabajo
**Vistas Creadas:**
- `v_calc_workorders` - Costos totales (labor-only y labor+supplies)

**Casos Implementados:**
- **Caso A:** `workorder_total_usd = labor_price_per_ha × effective_area`
- **Caso B:** `workorder_total_usd = supplies_total_usd + (labor_price_per_ha × effective_area)`

---

### **000063 - LABORS (Labores)**
**Archivos:**
- `000063_create_labor_calc_views.up.sql`
- `000063_create_labor_calc_views.down.sql`

**Funcionalidad:** Métricas de labor con IVA 10.5% y conversión ARS
**Vistas Creadas:**
- `v_calc_labors` - Métricas completas de labor

**Cálculos Implementados:**
- Total USD neto por labor
- IVA al 10.5% (reemplaza el viejo 21%)
- Conversión USD → ARS con tipos de cambio
- Total en pesos argentinos por orden

---

### **000064 - LOTS (Lotes)**
**Archivos:**
- `000064_create_lot_calc_views.up.sql`
- `000064_create_lot_calc_views.down.sql`

**Funcionalidad:** Rendimiento y economía por lote
**Vistas Creadas:**
- `v_calc_lots` - Métricas completas de lote

**Cálculos Implementados:**
- Yield (ton/ha) con fallback inteligente
- Precio neto por tonelada (último por fecha)
- Ingreso neto por hectárea
- Costo por hectárea
- Costo administrativo por hectárea
- **4 Modos de Arriendo:**
  - Fixed (constante por ha)
  - % Net income
  - % Utility (profit)
  - Mixed (fixed + % net income)
- Total activo por hectárea
- Resultado operativo por hectárea

---

### **000065 - PROJECT ROLLUPS (Consolidados de Proyecto)**
**Archivos:**
- `000065_create_project_rollup_views.up.sql`
- `000065_create_project_rollup_views.down.sql`

**Funcionalidad:** Consolidados de costos y economía por proyecto
**Vistas Creadas:**
- `v_calc_project_costs` - Consolidado de costos por proyecto
- `v_calc_project_economics` - Consolidado de economía por proyecto

**Cálculos Implementados:**
- Costos de cosecha por proyecto
- Otros costos por proyecto
- Costos totales por proyecto
- Ingreso neto por proyecto
- Total activo por proyecto
- Resultado operativo por proyecto

---

### **000066 - HELPER VIEWS (Vistas Auxiliares)**
**Archivos:**
- `000066_create_calc_helper_views.up.sql`
- `000066_create_calc_helper_views.down.sql`

**Funcionalidad:** Vistas helper compartidas para cálculos de todas las entidades
**Vistas Creadas:**
- `v_helper_harvests` - Totales de cosecha por lote
- `v_helper_last_net_price` - Último precio neto por campo y cultivo

**Propósito:** Proporcionar datos base para los cálculos de las entidades principales

---

### **000067 - FX RATES (Tipos de Cambio)**
**Archivos:**
- `000067_create_fx_rates_table.up.sql`
- `000067_create_fx_rates_table.down.sql`

**Funcionalidad:** Almacenar tasas de cambio para conversiones de moneda
**Tabla Creada:**
- `fx_rates` - Tipos de cambio con fechas de vigencia

**Características:**
- Soporte para múltiples monedas (USDARS, EURUSD, etc.)
- Fechas de vigencia para tasas históricas
- Soft-delete para auditoría

---

### **000068 - SUPPORT INDEXES (Índices de Soporte)**
**Archivos:**
- `000068_create_calc_support_indexes.up.sql`
- `000068_create_calc_support_indexes.down.sql`

**Funcionalidad:** Índices para optimizar consultas de cálculos
**Índices Creados:**
- `idx_workorders_labor_notdel` - workorders(labor_id) WHERE deleted_at IS NULL
- `idx_workorders_effarea_notdel` - workorders(effective_area) WHERE deleted_at IS NULL
- `idx_workorder_items_supply_notdel` - workorder_items(supply_id) WHERE deleted_at IS NULL
- `idx_labors_proj_notdel` - labors(project_id) WHERE deleted_at IS NULL
- `idx_supplies_proj_notdel` - supplies(project_id) WHERE deleted_at IS NULL
- `idx_workorders_lot_id_harvest_notdel` - workorders(lot_id) WHERE deleted_at IS NULL
- `idx_commercializations_f_c_date_notdel` - crop_commercializations(field_id, crop_id, created_at) WHERE deleted_at IS NULL

---

### **000069 - VERIFICATION (Verificación)**
**Archivos:**
- `000069_create_calc_verification_view.up.sql`
- `000069_create_calc_verification_view.down.sql`

**Funcionalidad:** Verificar que todos los cálculos funcionen correctamente
**Vista Creada:**
- `v_calc_verification` - 8 verificaciones automáticas

**Verificaciones Implementadas:**
1. Labor-only workorder verification
2. Labor + supplies verification
3. IVA calculation verification (10.5%)
4. ARS conversion verification
5. Yield calculation verification
6. Net price selection verification
7. Lease modes calculation verification
8. Project rollups verification

---

## 🔄 **Orden de Ejecución Recomendado**

### **Fase 1: Infraestructura Base**
1. **000067** - FX Rates (tabla de tipos de cambio)
2. **000068** - Support Indexes (índices de soporte)

### **Fase 2: Vistas Helper**
3. **000066** - Helper Views (vistas auxiliares compartidas)

### **Fase 3: Entidades Principales**
4. **000062** - Workorders (cálculos de órdenes)
5. **000063** - Labors (métricas de labor)
6. **000064** - Lots (rendimiento y economía)
7. **000065** - Project Rollups (consolidados)

### **Fase 4: Verificación**
8. **000069** - Verification (validación de cálculos)

---

## 📊 **Dependencias entre Migraciones**

```
000067 (FX Rates) ← 000063 (Labors - conversión ARS)
000068 (Indexes) ← Todas las entidades (performance)
000066 (Helpers) ← 000064 (Lots - yield y precios)
000062 (Workorders) ← 000064 (Lots - costos)
000063 (Labors) ← 000067 (FX Rates)
000064 (Lots) ← 000062 (Workorders) + 000066 (Helpers)
000065 (Project Rollups) ← 000062 (Workorders) + 000064 (Lots)
000069 (Verification) ← Todas las vistas de cálculo
```

---

## 🎯 **Beneficios de la Nueva Estructura**

### **✅ Organización Clara**
- **Por Entidad:** Cada migración se enfoca en una entidad específica
- **Por Funcionalidad:** Separación clara entre métricas, tablas y listas
- **Dependencias Explícitas:** Orden lógico de ejecución

### **✅ Mantenibilidad**
- **Fácil Debugging:** Problemas aislados por entidad
- **Rollback Selectivo:** Revertir funcionalidad específica
- **Testing Incremental:** Probar entidad por entidad

### **✅ Escalabilidad**
- **Nuevas Entidades:** Agregar migraciones 000070+ siguiendo el patrón
- **Modificaciones:** Cambiar solo la migración de la entidad afectada
- **Documentación:** Cada entidad tiene su propia documentación

---

## 🚀 **Próximos Pasos**

1. **Ejecutar Migraciones:** Seguir el orden recomendado
2. **Verificar Funcionalidad:** Usar `v_calc_verification`
3. **Testing por Entidad:** Validar cada entidad individualmente
4. **Performance Testing:** Verificar rendimiento con índices

---

*Documento generado automáticamente - Estructura de Migraciones Reorganizadas v1.0*
