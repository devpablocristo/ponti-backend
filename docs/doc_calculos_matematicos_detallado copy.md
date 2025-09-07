# 🧮 Cálculos Matemáticos Ponti - Documentación Técnica Detallada

### **📋 Resumen Ejecutivo**
Este documento describe en detalle todos los cálculos matemáticos implementados en el sistema Ponti, desde el origen de los datos en las tablas base hasta los resultados finales en las vistas de cálculo. Cada fórmula está explicada paso a paso con ejemplos prácticos y referencias a las tablas de origen.

---

## 📑 **ÍNDICE DE CÁLCULOS**

### **1. CÁLCULOS DE COSTOS DIRECTOS (WORKORDERS)**
- 1.1 Fórmula de Costo Labor
- 1.2 Fórmula de Costo Supplies  
- 1.3 Fórmula de Costo Directo Total

### **2. CÁLCULOS DE LABORES (TABLA LABORS)**
- 2.1 Fórmula de Total USD Neto
- 2.2 Fórmula de IVA (10.5%)
- 2.3 Fórmula de Total USD Bruto
- 2.4 Fórmula de Conversión USD a ARS
- 2.5 Fórmula de Total U Neto (en pesos)

### **3. CÁLCULOS DE LOTES**
- 3.1 Fórmula de Rendimiento
- 3.2 Fórmula de Ingreso Neto Total
- 3.3 Fórmula de Ingreso Neto por Hectárea
- 3.4 Fórmula de Costo Administrativo por Hectárea

### **4. CÁLCULOS DE ARRIENDO (4 TIPOS)**
- 4.1 Tipo 1: % Ingreso Neto
- 4.2 Tipo 2: % Utilidad
- 4.3 Tipo 3: Arriendo Fijo
- 4.4 Tipo 4: Arriendo Mixto (Fijo + % Ingreso Neto)

### **5. CÁLCULOS DE ACTIVO TOTAL Y RESULTADO OPERATIVO**
- 5.1 Fórmula de Activo Total por Hectárea
- 5.2 Fórmula de Resultado Operativo por Hectárea

### **6. CÁLCULOS DE CONSOLIDACIÓN POR PROYECTO**
- 6.1 Fórmula de Costos Totales por Proyecto
- 6.2 Fórmula de Ingresos Totales por Proyecto
- 6.3 Fórmula de Resultado Operativo por Proyecto

### **7. EJEMPLOS PRÁCTICOS COMPLETOS**
- 7.1 Ejemplo: Proyecto de Soja (100 ha)
- 7.2 Ejemplo: Workorder de Siembra

---

## 💰 **1. CÁLCULOS DE COSTOS DIRECTOS (WORKORDERS)**

### **1.1 Fórmula de Costo Labor**

**Flujo de datos:**
```
labors.price → workorders.effective_area → COSTO_LABOR
```

**Fórmula:**
```
COSTO_LABOR = precio_labor_por_ha × área_efectiva_trabajada
```

**Ejemplo práctico:**
- Labor: Siembra de Maíz
- Precio: $50 USD/ha
- Área efectiva: 100 ha
- **Cálculo:** $50 × 100 = $5,000 USD

### **1.2 Fórmula de Costo Supplies**

**Flujo de datos:**
```
workorder_items.final_dose → supplies.price → workorders.effective_area → COSTO_SUPPLY
```

**Fórmula:**
```
COSTO_SUPPLY = dosis_final × precio_supply × área_efectiva
```

**Ejemplo práctico:**
- Supply: Fertilizante NPK
- Dosis final: 200 kg/ha
- Precio: $0.50 USD/kg
- Área efectiva: 100 ha
- **Cálculo:** 200 × $0.50 × 100 = $10,000 USD

### **1.3 Fórmula de Costo Directo Total**

**Flujo de datos:**
```
COSTO_LABOR + COSTO_SUPPLY → COSTO_DIRECTO_TOTAL
```

**Fórmula:**
```
COSTO_DIRECTO_TOTAL = costo_labor + costo_supplies
```

**Ejemplo práctico:**
- Costo labor: $5,000 USD
- Costo supplies: $10,000 USD
- **Total:** $5,000 + $10,000 = $15,000 USD

---

## 🏭 **2. CÁLCULOS DE LABORES (TABLA LABORS)**

### **2.1 Fórmula de Total USD Neto**

**Flujo de datos:**
```
labors.price → workorders.effective_area → TOTAL_USD_NETO
```

**Fórmula:**
```
TOTAL_USD_NETO = precio_labor × superficie
```

**Ejemplo práctico:**
- Labor: Pulverización
- Precio: $25 USD/ha
- Superficie: 200 ha
- **Cálculo:** $25 × 200 = $5,000 USD

### **2.2 Fórmula de IVA (10.5%)**

**Flujo de datos:**
```
TOTAL_USD_NETO → calc_values.iva_percentage → IVA_AMOUNT
```

**Fórmula:**
```
IVA_AMOUNT = total_usd_neto × 0.105
```

**Ejemplo práctico:**
- Total USD neto: $5,000 USD
- IVA (10.5%): $5,000 × 0.105 = $525 USD

### **2.3 Fórmula de Total USD Bruto**

**Flujo de datos:**
```
TOTAL_USD_NETO + IVA_AMOUNT → TOTAL_USD_BRUTO
```

**Fórmula:**
```
TOTAL_USD_BRUTO = total_usd_neto + iva_amount
```

**Ejemplo práctico:**
- Total USD neto: $5,000 USD
- IVA: $525 USD
- **Total bruto:** $5,000 + $525 = $5,525 USD

### **2.4 Fórmula de Conversión USD a ARS**

**Flujo de datos:**
```
labors.price → fx_rates.rate → COSTO_ARS_HA
```

**Fórmula:**
```
COSTO_ARS_HA = precio_labor × tasa_cambio_usd_ars
```

**Ejemplo práctico:**
- Precio labor: $25 USD/ha
- Tasa USD/ARS: 1,355
- **Cálculo:** $25 × 1,355 = 33,875 ARS/ha

### **2.5 Fórmula de Total U Neto (en pesos)**

**Flujo de datos:**
```
COSTO_ARS_HA → workorders.effective_area → TOTAL_U_NETO
```

**Fórmula:**
```
TOTAL_U_NETO = costo_ars_ha × superficie
```

**Ejemplo práctico:**
- Costo ARS/ha: 33,875 ARS/ha
- Superficie: 200 ha
- **Cálculo:** 33,875 × 200 = 6,775,000 ARS

---

## 🌾 **3. CÁLCULOS DE LOTES**

### **3.1 Fórmula de Rendimiento**

**Flujo de datos:**
```
lots.tons → lots.hectares → RENDIMIENTO
```

**Fórmula:**
```
RENDIMIENTO = toneladas_cosechadas ÷ hectáreas
```

**Ejemplo práctico:**
- Toneladas: 300 ton
- Hectáreas: 150 ha
- **Cálculo:** 300 ÷ 150 = 2 ton/ha

### **3.2 Fórmula de Ingreso Neto Total**

**Flujo de datos:**
```
lots.tons → crop_commercializations.net_price → INGRESO_NETO_TOTAL
```

**Fórmula:**
```
INGRESO_NETO_TOTAL = toneladas × precio_neto_por_tonelada
```

**Ejemplo práctico:**
- Toneladas: 300 ton
- Precio neto: $410 USD/ton (Soja)
- **Cálculo:** 300 × $410 = $123,000 USD

### **3.3 Fórmula de Ingreso Neto por Hectárea**

**Flujo de datos:**
```
INGRESO_NETO_TOTAL → lots.hectares → INGRESO_NETO_HA
```

**Fórmula:**
```
INGRESO_NETO_HA = ingreso_neto_total ÷ hectáreas
```

**Ejemplo práctico:**
- Ingreso neto total: $123,000 USD
- Hectáreas: 150 ha
- **Cálculo:** $123,000 ÷ 150 = $820 USD/ha

### **3.4 Fórmula de Costo Administrativo por Hectárea**

**Flujo de datos:**
```
projects.admin_cost → SUM(lots.hectares) → COSTO_ADMIN_HA
```

**Fórmula:**
```
COSTO_ADMIN_HA = costo_admin_total_proyecto ÷ total_hectáreas_proyecto
```

**Ejemplo práctico:**
- Costo admin total: $15,000 USD
- Total hectáreas proyecto: 500 ha
- **Cálculo:** $15,000 ÷ 500 = $30 USD/ha

---

## 🏠 **4. CÁLCULOS DE ARRIENDO (4 TIPOS)**

### **4.1 Tipo 1: % Ingreso Neto**

**Flujo de datos:**
```
fields.lease_type_percent → INGRESO_NETO_HA → ARRIENDO_HA
```

**Fórmula:**
```
ARRIENDO_HA = (porcentaje ÷ 100) × ingreso_neto_por_ha
```

**Ejemplo práctico:**
- Porcentaje: 30%
- Ingreso neto/ha: $820 USD/ha
- **Cálculo:** (30 ÷ 100) × $820 = $246 USD/ha

### **4.2 Tipo 2: % Utilidad**

**Flujo de datos:**
```
INGRESO_NETO_HA → COSTO_DIRECTO_HA → COSTO_ADMIN_HA → UTILIDAD_HA → fields.lease_type_percent → ARRIENDO_HA
```

**Fórmula:**
```
UTILIDAD_HA = ingreso_neto_ha - costo_directo_ha - costo_admin_ha
ARRIENDO_HA = (porcentaje ÷ 100) × utilidad_ha
```

**Ejemplo práctico:**
- Ingreso neto/ha: $820 USD/ha
- Costo directo/ha: $100 USD/ha
- Costo admin/ha: $30 USD/ha
- Porcentaje: 25%
- **Utilidad/ha:** $820 - $100 - $30 = $690 USD/ha
- **Arriendo/ha:** (25 ÷ 100) × $690 = $172.50 USD/ha

### **4.3 Tipo 3: Arriendo Fijo**

**Flujo de datos:**
```
fields.lease_type_value → ARRIENDO_HA
```

**Fórmula:**
```
ARRIENDO_HA = valor_fijo_por_ha
```

**Ejemplo práctico:**
- Valor fijo: $150 USD/ha
- **Arriendo/ha:** $150 USD/ha

### **4.4 Tipo 4: Arriendo Mixto (Fijo + % Ingreso Neto)**

**Flujo de datos:**
```
fields.lease_type_value → fields.lease_type_percent → INGRESO_NETO_HA → ARRIENDO_HA
```

**Fórmula:**
```
ARRIENDO_HA = valor_fijo + ((porcentaje ÷ 100) × ingreso_neto_ha)
```

**Ejemplo práctico:**
- Valor fijo: $100 USD/ha
- Porcentaje: 20%
- Ingreso neto/ha: $820 USD/ha
- **Cálculo:** $100 + ((20 ÷ 100) × $820) = $100 + $164 = $264 USD/ha

---

## 📊 **5. CÁLCULOS DE ACTIVO TOTAL Y RESULTADO OPERATIVO**

### **5.1 Fórmula de Activo Total por Hectárea**

**Flujo de datos:**
```
COSTO_DIRECTO_HA → ARRIENDO_HA → COSTO_ADMIN_HA → ACTIVO_TOTAL_HA
```

**Fórmula:**
```
ACTIVO_TOTAL_HA = costo_directo_ha + arriendo_ha + costo_admin_ha
```

**Ejemplo práctico:**
- Costo directo/ha: $100 USD/ha
- Arriendo/ha: $246 USD/ha
- Costo admin/ha: $30 USD/ha
- **Cálculo:** $100 + $246 + $30 = $376 USD/ha

### **5.2 Fórmula de Resultado Operativo por Hectárea**

**Flujo de datos:**
```
INGRESO_NETO_HA → ACTIVO_TOTAL_HA → RESULTADO_OPERATIVO_HA
```

**Fórmula:**
```
RESULTADO_OPERATIVO_HA = ingreso_neto_ha - activo_total_ha
```

**Ejemplo práctico:**
- Ingreso neto/ha: $820 USD/ha
- Activo total/ha: $376 USD/ha
- **Cálculo:** $820 - $376 = $444 USD/ha

---

## 🏢 **6. CÁLCULOS DE CONSOLIDACIÓN POR PROYECTO**

### **6.1 Fórmula de Costos Totales por Proyecto**

**Flujo de datos:**
```
Σ(COSTOS_DIRECTOS_LOTE) → Σ(ARRIENDOS_LOTE) → Σ(COSTOS_ADMIN_LOTE) → COSTOS_TOTALES
```

**Fórmula:**
```
COSTOS_TOTALES = Σ(costos_directos_lote) + Σ(arriendos_lote) + Σ(costos_admin_lote)
```

**Ejemplo práctico:**
- Proyecto con 3 lotes:
  - Lote 1: $15,000 USD (directos) + $24,600 USD (arriendo) + $3,000 USD (admin) = $42,600 USD
  - Lote 2: $12,000 USD (directos) + $14,880 USD (arriendo) + $2,400 USD (admin) = $29,280 USD
  - Lote 3: $18,000 USD (directos) + $31,584 USD (arriendo) + $3,600 USD (admin) = $53,184 USD
- **Total proyecto:** $42,600 + $29,280 + $53,184 = $125,064 USD

### **6.2 Fórmula de Ingresos Totales por Proyecto**

**Flujo de datos:**
```
Σ(INGRESOS_NETOS_LOTE) → INGRESOS_TOTALES
```

**Fórmula:**
```
INGRESOS_TOTALES = Σ(ingresos_netos_lote)
```

**Ejemplo práctico:**
- Proyecto con 3 lotes:
  - Lote 1: $82,000 USD
  - Lote 2: $49,600 USD
  - Lote 3: $49,200 USD
- **Total ingresos:** $82,000 + $49,600 + $49,200 = $180,800 USD

### **6.3 Fórmula de Resultado Operativo por Proyecto**

**Flujo de datos:**
```
INGRESOS_TOTALES → COSTOS_TOTALES → RESULTADO_OPERATIVO
```

**Fórmula:**
```
RESULTADO_OPERATIVO = ingresos_totales - costos_totales
```

**Ejemplo práctico:**
- Ingresos totales: $180,800 USD
- Costos totales: $125,064 USD
- **Resultado operativo:** $180,800 - $125,064 = $55,736 USD

---

## 🧪 **7. EJEMPLOS PRÁCTICOS COMPLETOS**

### **7.1 Ejemplo: Proyecto de Soja (100 ha)**

**Datos de entrada:**
- Lote: 100 ha de Soja
- Rendimiento: 3.5 ton/ha
- Precio neto: $410 USD/ton
- Costo directo: $120 USD/ha
- Arriendo: 30% ingreso neto
- Costo admin: $25 USD/ha

**Cálculos paso a paso:**

1. **Ingreso neto total:**
   - Toneladas: 100 ha × 3.5 ton/ha = 350 ton
   - Ingreso: 350 ton × $410/ton = $143,500 USD

2. **Ingreso neto por hectárea:**
   - $143,500 ÷ 100 ha = $1,435 USD/ha

3. **Arriendo (30% ingreso neto):**
   - $1,435/ha × 0.30 = $430.50 USD/ha
   - Total arriendo: $430.50 × 100 ha = $43,050 USD

4. **Costo administrativo:**
   - $25/ha × 100 ha = $2,500 USD

5. **Activo total:**
   - Costo directo: $120/ha × 100 ha = $12,000 USD
   - Arriendo: $43,050 USD
   - Costo admin: $2,500 USD
   - **Total activo:** $57,550 USD

6. **Resultado operativo:**
   - Ingreso neto: $143,500 USD
   - Activo total: $57,550 USD
   - **Resultado:** $143,500 - $57,550 = $85,950 USD

### **7.2 Ejemplo: Workorder de Siembra**

**Datos de entrada:**
- Labor: Siembra de Maíz
- Precio labor: $45 USD/ha
- Área efectiva: 80 ha
- Supply: Semilla de Maíz
- Dosis: 25 kg/ha
- Precio semilla: $2.50 USD/kg

**Cálculos paso a paso:**

1. **Costo labor:**
   - $45/ha × 80 ha = $3,600 USD

2. **Costo supply:**
   - 25 kg/ha × $2.50/kg × 80 ha = $5,000 USD

3. **Costo directo total:**
   - $3,600 + $5,000 = $8,600 USD

4. **IVA labor (10.5%):**
   - $3,600 × 0.105 = $378 USD

5. **Total labor con IVA:**
   - $3,600 + $378 = $3,978 USD

---

*Documento Técnico - Cálculos Matemáticos Ponti v2.0 - Implementación Completa y Verificada*