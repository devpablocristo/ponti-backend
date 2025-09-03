# Cálculos de Ponti - Documentación Detallada

## Índice de Cálculos

### 1. **Workorders (Órdenes de Trabajo)**
### 2. **Labors (Labores)**
### 3. **Lots (Lotes)**
### 4. **Project Rollups (Consolidados de Proyecto)**

---

## 1. WORKORDERS (ÓRDENES DE TRABAJO)

### 1.1 **Costo Total de Workorder - Solo Labor**
**Nombre Español:** Costo Total de Orden Solo con Labor  
**Nombre Inglés:** Workorder Total Cost - Labor Only  
**Vista:** `v_calc_workorders`

**¿Qué calcula?**  
El costo total de una orden de trabajo que solo incluye labor (sin insumos).

**¿Cómo calcula?**  
```
workorder_total_usd = labor_price_per_ha × effective_area
```

**Ejemplo de cálculo exitoso:**
- `labor_price_per_ha = 5 USD/ha`
- `effective_area = 15 ha`
- **Resultado:** `5 × 15 = 75 USD`

**Código SQL:**
```sql
SELECT 
    workorder_id,
    labor_price_per_ha,
    effective_area,
    workorder_total_usd
FROM v_calc_workorders 
WHERE supplies_total_usd = 0;
```

---

### 1.2 **Costo Total de Workorder - Labor + Insumos**
**Nombre Español:** Costo Total de Orden con Labor e Insumos  
**Nombre Inglés:** Workorder Total Cost - Labor + Supplies  
**Vista:** `v_calc_workorders`

**¿Qué calcula?**  
El costo total de una orden de trabajo que incluye tanto labor como insumos.

**¿Cómo calcula?**  
```
workorder_total_usd = supplies_total_usd + (labor_price_per_ha × effective_area)
```

**Ejemplo de cálculo exitoso:**
- `supplies_total_usd = 99.98 USD`
- `labor_price_per_ha = 5 USD/ha`
- `effective_area = 15 ha`
- **Resultado:** `99.98 + (5 × 15) = 99.98 + 75 = 174.98 USD`

**Código SQL:**
```sql
SELECT 
    workorder_id,
    supplies_total_usd,
    labor_price_per_ha,
    effective_area,
    workorder_total_usd
FROM v_calc_workorders 
WHERE supplies_total_usd > 0;
```

---

## 2. LABORS (LABORES)

### 2.1 **Total USD Neto por Labor**
**Nombre Español:** Total USD Neto por Labor  
**Nombre Inglés:** Total USD Net per Labor  
**Vista:** `v_calc_labors`

**¿Qué calcula?**  
El costo total en USD de una labor específica por el área efectiva trabajada.

**¿Cómo calcula?**  
```
total_usd_net = labor_price_per_ha × effective_area
```

**Ejemplo de cálculo exitoso:**
- `labor_price_per_ha = 8 USD/ha`
- `effective_area = 20 ha`
- **Resultado:** `8 × 20 = 160 USD`

**Código SQL:**
```sql
SELECT 
    labor_id,
    labor_price_per_ha,
    effective_area,
    total_usd_net
FROM v_calc_labors;
```

---

### 2.2 **IVA para Labores (10.5%)**
**Nombre Español:** IVA para Labores  
**Nombre Inglés:** VAT for Labors  
**Vista:** `v_calc_labors`

**¿Qué calcula?**  
El impuesto al valor agregado (IVA) aplicado a las labores, usando la tasa del 10.5%.

**¿Cómo calcula?**  
```
iva_amount = total_usd_net × 0.105
```

**Ejemplo de cálculo exitoso:**
- `total_usd_net = 160 USD`
- **Resultado:** `160 × 0.105 = 16.80 USD`

**Código SQL:**
```sql
SELECT 
    labor_id,
    total_usd_net,
    iva_amount,
    ROUND(iva_amount, 2) AS iva_rounded
FROM v_calc_labors 
WHERE iva_amount > 0;
```

---

### 2.3 **Costo por Hectárea en Pesos Argentinos**
**Nombre Español:** Costo por Hectárea en Pesos Argentinos  
**Nombre Inglés:** Cost per Hectare in Argentine Pesos  
**Vista:** `v_calc_labors`

**¿Qué calcula?**  
El costo de la labor por hectárea convertido a pesos argentinos usando el tipo de cambio más reciente.

**¿Cómo calcula?**  
```
cost_ars_per_ha = labor_price_per_ha × usd_ars_rate
```

**Ejemplo de cálculo exitoso:**
- `labor_price_per_ha = 8 USD/ha`
- `usd_ars_rate = 850.50` (tipo de cambio actual)
- **Resultado:** `8 × 850.50 = 6,804 ARS/ha`

**Código SQL:**
```sql
SELECT 
    labor_id,
    labor_price_per_ha,
    usd_ars_rate,
    cost_ars_per_ha
FROM v_calc_labors 
WHERE usd_ars_rate > 1;
```

---

### 2.4 **Total en Pesos Argentinos por Orden**
**Nombre Español:** Total en Pesos Argentinos por Orden  
**Nombre Inglés:** Total in Argentine Pesos per Order  
**Vista:** `v_calc_labors`

**¿Qué calcula?**  
El costo total de la labor en pesos argentinos para toda el área trabajada.

**¿Cómo calcula?**  
```
total_ars = cost_ars_per_ha × effective_area
```

**Ejemplo de cálculo exitoso:**
- `cost_ars_per_ha = 6,804 ARS/ha`
- `effective_area = 20 ha`
- **Resultado:** `6,804 × 20 = 136,080 ARS`

**Código SQL:**
```sql
SELECT 
    workorder_id,
    cost_ars_per_ha,
    effective_area,
    total_ars
FROM v_calc_labors;
```

---

## 3. LOTS (LOTES)

### 3.1 **Rendimiento por Hectárea (ton/ha)**
**Nombre Español:** Rendimiento por Hectárea  
**Nombre Inglés:** Yield per Hectare  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
La producción de toneladas por hectárea de un lote específico.

**¿Cómo calcula?**  
```
yield_tonha = total_tons / lot_hectares
```

**Política de fuentes:**
1. **Primaria:** Suma de áreas cosechadas de workorders con `category_id = 13` (Harvest)
2. **Fallback:** `lots.tons` si no hay datos de cosecha

**Ejemplo de cálculo exitoso:**
- `total_tons = 150 ton`
- `lot_hectares = 25 ha`
- **Resultado:** `150 ÷ 25 = 6 ton/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    lot_hectares,
    yield_tonha,
    ROUND(yield_tonha, 2) AS yield_rounded
FROM v_calc_lots 
WHERE yield_tonha > 0;
```

---

### 3.2 **Precio Neto por Tonelada (USD)**
**Nombre Español:** Precio Neto por Tonelada  
**Nombre Inglés:** Net Price per Ton  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
El precio neto más reciente por tonelada para la combinación (field_id, crop_id).

**¿Cómo calcula?**  
```
net_price_usd = último_precio_de_crop_commercializations
```

**Política de selección:**
- Último precio por fecha para (field_id, crop_id)
- Fallback a (project_id, crop_id) si no hay precio específico

**Ejemplo de cálculo exitoso:**
- Último precio registrado: `net_price = 280 USD/ton`
- **Resultado:** `280 USD/ton`

**Código SQL:**
```sql
SELECT 
    lot_id,
    crop_id,
    net_price_usd
FROM v_calc_lots 
WHERE net_price_usd > 0;
```

---

### 3.3 **Ingreso Neto por Hectárea (USD)**
**Nombre Español:** Ingreso Neto por Hectárea  
**Nombre Inglés:** Net Income per Hectare  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
El ingreso neto esperado por hectárea basado en el rendimiento y precio.

**¿Cómo calcula?**  
```
net_income_per_ha = net_price_usd × yield_tonha
```

**Ejemplo de cálculo exitoso:**
- `net_price_usd = 280 USD/ton`
- `yield_tonha = 6 ton/ha`
- **Resultado:** `280 × 6 = 1,680 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    net_price_usd,
    yield_tonha,
    net_income_per_ha
FROM v_calc_lots 
WHERE net_income_per_ha > 0;
```

---

### 3.4 **Costo por Hectárea (USD)**
**Nombre Español:** Costo por Hectárea  
**Nombre Inglés:** Cost per Hectare  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
El costo total de todas las operaciones por hectárea del lote.

**¿Cómo calcula?**  
```
cost_per_ha = SUM(workorder_total_usd) / lot_hectares
```

**Ejemplo de cálculo exitoso:**
- `SUM(workorder_total_usd) = 2,500 USD`
- `lot_hectares = 25 ha`
- **Resultado:** `2,500 ÷ 25 = 100 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    lot_hectares,
    cost_per_ha
FROM v_calc_lots 
WHERE cost_per_ha > 0;
```

---

### 3.5 **Costo Administrativo por Hectárea (USD)**
**Nombre Español:** Costo Administrativo por Hectárea  
**Nombre Inglés:** Administrative Cost per Hectare  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
El costo administrativo fijo por hectárea del proyecto.

**¿Cómo calcula?**  
```
admin_cost_per_ha = projects.admin_cost
```

**Ejemplo de cálculo exitoso:**
- `admin_cost = 50 USD/ha`
- **Resultado:** `50 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    admin_cost_per_ha
FROM v_calc_lots;
```

---

### 3.6 **Arriendo por Hectárea - Modo Fixed (USD)**
**Nombre Español:** Arriendo por Hectárea - Modo Fijo  
**Nombre Inglés:** Lease per Hectare - Fixed Mode  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
El arriendo fijo por hectárea cuando `lease_type_id = 1`.

**¿Cómo calcula?**  
```
lease_per_ha = lease_type_value
```

**Ejemplo de cálculo exitoso:**
- `lease_type_value = 120 USD/ha`
- **Resultado:** `120 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    lease_type_id,
    lease_per_ha
FROM v_calc_lots 
WHERE lease_type_id = 1;
```

---

### 3.7 **Arriendo por Hectárea - % Ingreso Neto (USD)**
**Nombre Español:** Arriendo por Hectárea - % Ingreso Neto  
**Nombre Inglés:** Lease per Hectare - % Net Income  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
El arriendo calculado como porcentaje del ingreso neto cuando `lease_type_id = 2`.

**¿Cómo calcula?**  
```
lease_per_ha = lease_type_percent × net_income_per_ha
```

**Ejemplo de cálculo exitoso:**
- `lease_type_percent = 0.15` (15%)
- `net_income_per_ha = 1,680 USD/ha`
- **Resultado:** `0.15 × 1,680 = 252 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    lease_type_percent,
    net_income_per_ha,
    lease_per_ha
FROM v_calc_lots 
WHERE lease_type_id = 2;
```

---

### 3.8 **Arriendo por Hectárea - % Utilidad (USD)**
**Nombre Español:** Arriendo por Hectárea - % Utilidad  
**Nombre Inglés:** Lease per Hectare - % Utility  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
El arriendo calculado como porcentaje de la utilidad cuando `lease_type_id = 3`.

**¿Cómo calcula?**  
```
utility_per_ha = net_income_per_ha - cost_per_ha - admin_cost_per_ha
lease_per_ha = lease_type_percent × utility_per_ha
```

**Ejemplo de cálculo exitoso:**
- `net_income_per_ha = 1,680 USD/ha`
- `cost_per_ha = 100 USD/ha`
- `admin_cost_per_ha = 50 USD/ha`
- `lease_type_percent = 0.20` (20%)
- **Cálculo:** `utility_per_ha = 1,680 - 100 - 50 = 1,530 USD/ha`
- **Resultado:** `0.20 × 1,530 = 306 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    net_income_per_ha,
    cost_per_ha,
    admin_cost_per_ha,
    lease_type_percent,
    lease_per_ha
FROM v_calc_lots 
WHERE lease_type_id = 3;
```

---

### 3.9 **Arriendo por Hectárea - Modo Mixto (USD)**
**Nombre Español:** Arriendo por Hectárea - Modo Mixto  
**Nombre Inglés:** Lease per Hectare - Mixed Mode  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
El arriendo que combina un valor fijo más un porcentaje del ingreso neto cuando `lease_type_id = 4`.

**¿Cómo calcula?**  
```
lease_per_ha = lease_type_value + (lease_type_percent × net_income_per_ha)
```

**Ejemplo de cálculo exitoso:**
- `lease_type_value = 80 USD/ha`
- `lease_type_percent = 0.10` (10%)
- `net_income_per_ha = 1,680 USD/ha`
- **Cálculo:** `80 + (0.10 × 1,680) = 80 + 168 = 248 USD/ha`
- **Resultado:** `248 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    lease_type_value,
    lease_type_percent,
    net_income_per_ha,
    lease_per_ha
FROM v_calc_lots 
WHERE lease_type_id = 4;
```

---

### 3.10 **Total Activo por Hectárea (USD)**
**Nombre Español:** Total Activo por Hectárea  
**Nombre Inglés:** Active Total per Hectare  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
La suma total de todos los costos activos por hectárea del lote.

**¿Cómo calcula?**  
```
active_total_per_ha = cost_per_ha + lease_per_ha + admin_cost_per_ha
```

**Ejemplo de cálculo exitoso:**
- `cost_per_ha = 100 USD/ha`
- `lease_per_ha = 252 USD/ha`
- `admin_cost_per_ha = 50 USD/ha`
- **Resultado:** `100 + 252 + 50 = 402 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    cost_per_ha,
    lease_per_ha,
    admin_cost_per_ha,
    active_total_per_ha
FROM v_calc_lots;
```

---

### 3.11 **Resultado Operativo por Hectárea (USD)**
**Nombre Español:** Resultado Operativo por Hectárea  
**Nombre Inglés:** Operating Result per Hectare  
**Vista:** `v_calc_lots`

**¿Qué calcula?**  
La utilidad operativa por hectárea después de deducir todos los costos.

**¿Cómo calcula?**  
```
operating_result_per_ha = net_income_per_ha - active_total_per_ha
```

**Ejemplo de cálculo exitoso:**
- `net_income_per_ha = 1,680 USD/ha`
- `active_total_per_ha = 402 USD/ha`
- **Resultado:** `1,680 - 402 = 1,278 USD/ha`

**Código SQL:**
```sql
SELECT 
    lot_id,
    net_income_per_ha,
    active_total_per_ha,
    operating_result_per_ha
FROM v_calc_lots;
```

---

## 4. PROJECT ROLLUPS (CONSOLIDADOS DE PROYECTO)

### 4.1 **Costos por Proyecto - Cosecha (USD)**
**Nombre Español:** Costos de Cosecha por Proyecto  
**Nombre Inglés:** Harvest Costs per Project  
**Vista:** `v_calc_project_costs`

**¿Qué calcula?**  
La suma total de costos de cosecha (workorders con `category_id = 13`) por proyecto.

**¿Cómo calcula?**  
```
harvest_costs_usd = SUM(workorder_total_usd) WHERE category_id = 13
```

**Ejemplo de cálculo exitoso:**
- Proyecto con 3 workorders de cosecha: `[150, 200, 175] USD`
- **Resultado:** `150 + 200 + 175 = 525 USD`

**Código SQL:**
```sql
SELECT 
    project_id,
    harvest_costs_usd
FROM v_calc_project_costs 
WHERE harvest_costs_usd > 0;
```

---

### 4.2 **Costos por Proyecto - Otros (USD)**
**Nombre Español:** Otros Costos por Proyecto  
**Nombre Inglés:** Other Costs per Project  
**Vista:** `v_calc_project_costs`

**¿Qué calcula?**  
La suma total de costos que no son de cosecha por proyecto.

**¿Cómo calcula?**  
```
other_costs_usd = SUM(workorder_total_usd) WHERE category_id != 13
```

**Ejemplo de cálculo exitoso:**
- Proyecto con 5 workorders de otros tipos: `[80, 120, 95, 150, 110] USD`
- **Resultado:** `80 + 120 + 95 + 150 + 110 = 555 USD`

**Código SQL:**
```sql
SELECT 
    project_id,
    other_costs_usd
FROM v_calc_project_costs 
WHERE other_costs_usd > 0;
```

---

### 4.3 **Costos Totales por Proyecto (USD)**
**Nombre Español:** Costos Totales por Proyecto  
**Nombre Inglés:** Total Costs per Project  
**Vista:** `v_calc_project_costs`

**¿Qué calcula?**  
La suma total de todos los costos (cosecha + otros) por proyecto.

**¿Cómo calcula?**  
```
total_costs_usd = harvest_costs_usd + other_costs_usd
```

**Ejemplo de cálculo exitoso:**
- `harvest_costs_usd = 525 USD`
- `other_costs_usd = 555 USD`
- **Resultado:** `525 + 555 = 1,080 USD`

**Código SQL:**
```sql
SELECT 
    project_id,
    harvest_costs_usd,
    other_costs_usd,
    total_costs_usd
FROM v_calc_project_costs;
```

---

### 4.4 **Ingreso Neto por Proyecto (USD)**
**Nombre Español:** Ingreso Neto por Proyecto  
**Nombre Inglés:** Net Income per Project  
**Vista:** `v_calc_project_economics`

**¿Qué calcula?**  
La suma total del ingreso neto de todos los lotes del proyecto.

**¿Cómo calcula?**  
```
net_income_usd = SUM(net_income_per_ha × lot_hectares)
```

**Ejemplo de cálculo exitoso:**
- Lote 1: `net_income_per_ha = 1,680 USD/ha`, `lot_hectares = 25 ha` → `42,000 USD`
- Lote 2: `net_income_per_ha = 1,200 USD/ha`, `lot_hectares = 30 ha` → `36,000 USD`
- **Resultado:** `42,000 + 36,000 = 78,000 USD`

**Código SQL:**
```sql
SELECT 
    project_id,
    net_income_usd
FROM v_calc_project_economics 
WHERE net_income_usd > 0;
```

---

### 4.5 **Total Activo por Proyecto (USD)**
**Nombre Español:** Total Activo por Proyecto  
**Nombre Inglés:** Active Total per Project  
**Vista:** `v_calc_project_economics`

**¿Qué calcula?**  
La suma total de todos los costos activos de todos los lotes del proyecto.

**¿Cómo calcula?**  
```
active_total_usd = SUM(active_total_per_ha × lot_hectares)
```

**Ejemplo de cálculo exitoso:**
- Lote 1: `active_total_per_ha = 402 USD/ha`, `lot_hectares = 25 ha` → `10,050 USD`
- Lote 2: `active_total_per_ha = 350 USD/ha`, `lot_hectares = 30 ha` → `10,500 USD`
- **Resultado:** `10,050 + 10,500 = 20,550 USD`

**Código SQL:**
```sql
SELECT 
    project_id,
    active_total_usd
FROM v_calc_project_economics 
WHERE active_total_usd > 0;
```

---

### 4.6 **Resultado Operativo por Proyecto (USD)**
**Nombre Español:** Resultado Operativo por Proyecto  
**Nombre Inglés:** Operating Result per Project  
**Vista:** `v_calc_project_economics`

**¿Qué calcula?**  
La utilidad operativa total del proyecto después de deducir todos los costos.

**¿Cómo calcula?**  
```
operating_result_usd = net_income_usd - active_total_usd
```

**Ejemplo de cálculo exitoso:**
- `net_income_usd = 78,000 USD`
- `active_total_usd = 20,550 USD`
- **Resultado:** `78,000 - 20,550 = 57,450 USD`

**Código SQL:**
```sql
SELECT 
    project_id,
    net_income_usd,
    active_total_usd,
    operating_result_usd
FROM v_calc_project_economics;
```

---

## 5. VISTAS HELPER (AUXILIARES)

### 5.1 **Totales de Cosecha por Lote**
**Nombre Español:** Totales de Cosecha por Lote  
**Nombre Inglés:** Harvest Totals per Lot  
**Vista:** `v_helper_harvests`

**¿Qué calcula?**  
El área cosechada y rendimiento por lote basado en workorders de cosecha.

**¿Cómo calcula?**  
```
harvested_area = SUM(workorders.effective_area) WHERE category_id = 13
yield_tonha = lots.tons / harvested_area
```

**Ejemplo de cálculo exitoso:**
- `harvested_area = 20 ha`
- `lots.tons = 120 ton`
- **Resultado:** `yield_tonha = 120 ÷ 20 = 6 ton/ha`

---

### 5.2 **Último Precio Neto por Campo y Cultivo**
**Nombre Español:** Último Precio Neto por Campo y Cultivo  
**Nombre Inglés:** Last Net Price per Field and Crop  
**Vista:** `v_helper_last_net_price`

**¿Qué calcula?**  
El precio neto más reciente para cada combinación (field_id, crop_id).

**¿Cómo calcula?**  
```
DISTINCT ON (field_id, crop_id) ORDER BY created_at DESC
```

**Ejemplo de cálculo exitoso:**
- Campo 1, Cultivo A: Último precio = `280 USD/ton`
- Campo 2, Cultivo B: Último precio = `320 USD/ton`

---

### 5.3 **Base Consolidada de Workorders**
**Nombre Español:** Base Consolidada de Órdenes de Trabajo  
**Nombre Inglés:** Consolidated Workorder Base  
**Vista:** `v_helper_workorder_base`

**¿Qué calcula?**  
Una vista consolidada que une workorders, labors y supplies con soft-delete filters.

**¿Cómo calcula?**  
```
JOIN workorders + labors + LEFT JOIN workorder_items + supplies
GROUP BY workorder_id
```

**Ejemplo de cálculo exitoso:**
- Workorder con labor de $8/ha y 15 ha de área
- Supplies totales: $150
- **Resultado:** Datos consolidados para cálculos posteriores

---

## 6. VERIFICACIÓN Y VALIDACIÓN

### 6.1 **Vista de Verificación General**
**Nombre Español:** Vista de Verificación de Cálculos  
**Nombre Inglés:** Calculation Verification View  
**Vista:** `v_calc_verification`

**¿Qué calcula?**  
Verificaciones automáticas de que todos los cálculos funcionen correctamente.

**¿Cómo calcula?**  
Ejecuta 8 verificaciones diferentes:
1. Labor-only workorder verification
2. Labor + supplies verification  
3. IVA calculation verification (10.5%)
4. ARS conversion verification
5. Yield calculation verification
6. Net price selection verification
7. Lease modes calculation verification
8. Project rollups verification

**Ejemplo de verificación exitosa:**
```
test_name: "Labor-only order verification"
test_result: "PASS: Found matching record"
```

---

## 7. TIPOS DE CAMBIO (FX RATES)

### 7.1 **Tabla de Tipos de Cambio**
**Nombre Español:** Tabla de Tipos de Cambio  
**Nombre Inglés:** Exchange Rates Table  
**Tabla:** `fx_rates`

**¿Qué almacena?**  
Los tipos de cambio entre diferentes monedas con fechas de vigencia.

**Estructura:**
- `code`: Código de moneda (ej: 'USDARS')
- `rate`: Tasa de cambio
- `as_of_date`: Fecha de vigencia
- `created_at`, `updated_at`, `deleted_at`

**Ejemplo de uso:**
- `USDARS = 850.50` para conversiones USD → ARS
- Última tasa por `as_of_date` para cálculos en tiempo real

---

## 8. ÍNDICES DE SOPORTE

### 8.1 **Índices para Cálculos**
**Nombre Español:** Índices de Soporte para Cálculos  
**Nombre Inglés:** Support Indexes for Calculations  

**Índices creados:**
- `idx_workorders_labor_notdel`: workorders(labor_id) WHERE deleted_at IS NULL
- `idx_workorders_effarea_notdel`: workorders(effective_area) WHERE deleted_at IS NULL
- `idx_workorder_items_supply_notdel`: workorder_items(supply_id) WHERE deleted_at IS NULL
- `idx_labors_proj_notdel`: labors(project_id) WHERE deleted_at IS NULL
- `idx_supplies_proj_notdel`: supplies(project_id) WHERE deleted_at IS NULL
- `idx_workorders_lot_id_harvest_notdel`: workorders(lot_id) WHERE deleted_at IS NULL
- `idx_commercializations_f_c_date_notdel`: crop_commercializations(field_id, crop_id, created_at) WHERE deleted_at IS NULL

**¿Qué optimizan?**  
Las consultas de las vistas de cálculo, especialmente filtros por soft-delete y joins entre tablas.

---

## 9. RESUMEN DE FÓRMULAS PRINCIPALES

### **Workorders:**
- **Solo Labor:** `workorder_total_usd = labor_price_per_ha × effective_area`
- **Labor + Supplies:** `workorder_total_usd = supplies_total_usd + (labor_price_per_ha × effective_area)`

### **Labors:**
- **Total Neto:** `total_usd_net = labor_price_per_ha × effective_area`
- **IVA:** `iva_amount = total_usd_net × 0.105`
- **ARS por Ha:** `cost_ars_per_ha = labor_price_per_ha × usd_ars_rate`
- **Total ARS:** `total_ars = cost_ars_per_ha × effective_area`

### **Lots:**
- **Yield:** `yield_tonha = total_tons / lot_hectares`
- **Ingreso Neto:** `net_income_per_ha = net_price_usd × yield_tonha`
- **Costo por Ha:** `cost_per_ha = SUM(workorder_total_usd) / lot_hectares`
- **Total Activo:** `active_total_per_ha = cost_per_ha + lease_per_ha + admin_cost_per_ha`
- **Resultado Operativo:** `operating_result_per_ha = net_income_per_ha - active_total_per_ha`

### **Project Rollups:**
- **Costos Totales:** `total_costs_usd = harvest_costs_usd + other_costs_usd`
- **Ingreso Neto:** `net_income_usd = SUM(net_income_per_ha × lot_hectares)`
- **Resultado Operativo:** `operating_result_usd = net_income_usd - active_total_usd`

---

## 10. CONSIDERACIONES TÉCNICAS

### **Soft-Delete Filters:**
Todas las vistas incluyen `WHERE deleted_at IS NULL` para respetar la lógica de borrado lógico.

### **Null-Safety:**
- Uso de `COALESCE()` para valores por defecto
- Uso de `NULLIF(denominator, 0)` para evitar división por cero

### **Performance:**
- Índices parciales solo en registros activos
- CTEs optimizados para cálculos complejos
- Agregaciones eficientes con `GROUP BY`

### **Idempotencia:**
Todas las migraciones usan `CREATE OR REPLACE VIEW` y `CREATE INDEX IF NOT EXISTS` para ser idempotentes.

---

*Documento generado automáticamente - Cálculos de Ponti v1.0*
