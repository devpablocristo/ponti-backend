# 🧮 Cálculos Matemáticos Ponti - Documentación Técnica Detallada

### **📋 Resumen Ejecutivo**
Este documento describe en detalle todos los cálculos matemáticos implementados en el sistema Ponti, desde el origen de los datos en las tablas base hasta los resultados finales en las vistas de cálculo. Cada fórmula está explicada paso a paso con ejemplos prácticos y referencias a las tablas de origen.

---

## 💰 **1. CÁLCULOS DE COSTOS DIRECTOS (WORKORDERS)**

### **Fórmula de Costo Labor**
```
COSTO_LABOR = precio_labor_por_ha × área_efectiva_trabajada
```

**Ejemplo práctico:**
- Labor: Siembra de Maíz
- Precio: $50 USD/ha
- Área efectiva: 100 ha
- **Cálculo:** $50 × 100 = $5,000 USD

### **Fórmula de Costo Supplies**
```
COSTO_SUPPLY = dosis_final × precio_supply × área_efectiva
```
**Ejemplo práctico:**
- Supply: Fertilizante NPK
- Dosis final: 200 kg/ha
- Precio: $0.50 USD/kg
- Área efectiva: 100 ha
- **Cálculo:** 200 × $0.50 × 100 = $10,000 USD

### **1.4 Fórmula de Costo Directo Total**
```
COSTO_DIRECTO_TOTAL = costo_labor + costo_supplies
```

**Ejemplo práctico:**
- Costo labor: $5,000 USD
- Costo supplies: $10,000 USD
- **Total:** $5,000 + $10,000 = $15,000 USD

---