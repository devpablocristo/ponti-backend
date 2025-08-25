# 🌱 SEED TOOL 2 - Generador de Datos de Prueba para LIST LOT

## 📋 Descripción

**SEED TOOL 2** es una herramienta especializada para generar datos de prueba coherentes y realistas específicamente para verificar el correcto funcionamiento del endpoint `list lot`. 

Esta herramienta está basada en la **secuencia corregida** y genera datos diferenciados con valores realistas para testing.

## 🎯 Objetivo Específico

Crear un conjunto de datos de prueba que permita:
- ✅ Verificar cálculos de costos directos
- ✅ Verificar cálculos de ingresos netos
- ✅ Verificar cálculos de arriendo (fijo vs porcentaje)
- ✅ Verificar cálculos de costos administrativos
- ✅ Verificar resultados operativos diferenciados entre lots

## 🏗️ Estructura de Datos Generados

### **Proyecto Principal**
- **Nombre:** "Construcción Torre Norte"
- **ID:** 1
- **Admin Cost:** $15,000
- **Customer:** Inmobiliaria Buenos Aires S.A.
- **Campaign:** Campaña Loteo 2025

### **Campos y Lotes**
| Campo | Lote | Hectáreas | Cultivo | Season | Arriendo |
|-------|------|-----------|---------|--------|----------|
| **Campo A** | Parcela A1 | 2.5 ha | Soja → Maíz | Invierno 2025 | Fijo: $100/ha |
| **Campo A** | Parcela A2 | 3.0 ha | Soja → Maíz | Verano 2025 | Fijo: $100/ha |
| **Campo B** | Parcela B1 | 1.2 ha | Soja → Maíz | Otoño 2025 | 15% ingresos |

### **Labores (Categorías Corregidas)**
- **Siembra Maíz:** $25.50/ha - Categoría 9 (Siembra)
- **Cosecha Maíz:** $45.00/ha - Categoría 13 (Cosecha)

### **Rendimiento y Comercialización**
- **Rendimiento:** 8 tn/ha para todos los lots
- **Precio Maíz:** $450/ton neto
- **Precio Soja:** $650/ton neto

## 🚀 Uso

### **1. Ejecutar Todos los Seeds**
```bash
# Desde el directorio raíz del proyecto
cd projects/ponti-api

# Ejecutar todos los seeds para LIST LOT
go run cmd/seed-tool2/main.go -all
```

### **2. Limpiar y Ejecutar Todo**
```bash
# Limpiar base de datos y ejecutar todos los seeds
go run cmd/seed-tool2/main.go -reset -all
```

### **3. Ejecutar Seed Específico**
```bash
# Ejecutar solo usuarios
go run cmd/seed-tool2/main.go -seed users

# Ejecutar solo campos
go run cmd/seed-tool2/main.go -seed fields

# Ejecutar solo lotes
go run cmd/seed-tool2/main.go -seed lots
```

### **4. Ver Ayuda**
```bash
go run cmd/seed-tool2/main.go -help
```

## 📁 Archivos SQL

### **Orden de Ejecución:**
1. **`00_reset.sql`** - Limpia base de datos
2. **`01_users.sql`** - Usuario demo (ID: 123)
3. **`02_types.sql`** - Tipos base del sistema
4. **`03_lease_types.sql`** - Tipos de arriendo
5. **`04_managers.sql`** - Managers del proyecto
6. **`05_investors.sql`** - Inversores
7. **`06_categories.sql`** - Categorías

8. **`07_customers.sql`** - Clientes
9. **`08_campaigns.sql`** - Campañas
10. **`09_crops.sql`** - Cultivos (Soja, Maíz)
11. **`10_projects.sql`** - Proyecto principal
12. **`11_project_managers.sql`** - Asociación managers
13. **`12_project_investors.sql`** - Asociación inversores
14. **`13_fields.sql`** - Campos (A y B)
15. **`14_lots.sql`** - Lotes (3 parcelas)
16. **`15_supplies.sql`** - Insumos
17. **`16_labors.sql`** - Labores (categorías corregidas)
18. **`17_workorders.sql`** - Órdenes de trabajo
19. **`18_tons.sql`** - Toneladas de los lotes
20. **`19_crop_commercializations.sql`** - Precios de comercialización
21. **`20_lot_dates.sql`** - Fechas de siembra/cosecha
22. **`21_verification.sql`** - Verificación final

## 🔧 Correcciones Aplicadas

### **Categorías de Labores:**
- ✅ **Siembra:** `category_id = 9` (no 1)
- ✅ **Cosecha:** `category_id = 13` (no 7)

### **Fechas de Workorders:**
- ✅ **Invierno 2025:** Siembra Junio, Cosecha Diciembre
- ✅ **Verano 2025:** Siembra Diciembre, Cosecha Junio
- ✅ **Otoño 2025:** Siembra Marzo, Cosecha Septiembre

### **Precios de Comercialización:**
- ✅ **Maíz:** $450/ton neto (realista)
- ✅ **Soja:** $650/ton neto (realista)

### **Configuración de Arriendo:**
- ✅ **Campo A:** Arriendo fijo $100/ha
- ✅ **Campo B:** Arriendo 15% de ingresos

## 📊 Resultados Esperados

### **Datos Diferenciados:**
- **Parcela A1:** 2.5 ha, 20 tn, $9,000 ingresos
- **Parcela A2:** 3.0 ha, 24 tn, $10,800 ingresos  
- **Parcela B1:** 1.2 ha, 9.6 tn, $4,320 ingresos

### **Cálculos Coherentes:**
- **Costos directos** basados en workorders reales
- **Ingresos netos** basados en toneladas y precios reales
- **Arriendo diferenciado** por tipo de campo
- **Resultados operativos** realistas y diferenciados

## 🧪 Testing

### **Verificar Cálculos:**
```sql
-- Verificar que los cálculos sean coherentes
SELECT 
    id,
    lot_name,
    sowed_area,
    tons,
    income_net_total,
    direct_cost_total,
    rent_total,
    admin_total,
    active_total,
    operating_result_per_ha
FROM lot_table_view 
ORDER BY id;
```

### **Verificar Diferenciación:**
- ✅ Cada lote debe tener valores diferentes
- ✅ Los ingresos deben ser proporcionales a las toneladas
- ✅ Los costos deben variar según el área
- ✅ Los resultados operativos deben ser realistas

## 🔗 Probar API

### **Endpoint LIST LOT:**
```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/lots?project_id=1"
```

### **Resultado Esperado:**
```json
{
  "page_info": {
    "per_page": 10,
    "page": 1,
    "max_page": 1,
    "total": 3
  },
  "totals": {
    "sum_sowed_area": "6.7",
    "sum_cost": "211.5"
  },
  "items": [
    {
      "id": 1,
      "lot_name": "Parcela A1",
      "sowed_area": "2.5",
      "tons": "20.0",
      "income_net_per_ha": "3600",
      "cost_usd_per_ha": "70.5",
      "operating_result_per_ha": "3529.5"
    },
    // ... más lots con datos diferenciados
  ]
}
```

## 🎉 Beneficios

1. **Datos Realistas:** Valores coherentes con agricultura real
2. **Diferenciación Clara:** Cada lote tiene características únicas
3. **Cálculos Correctos:** Todos los campos calculados funcionan
4. **Testing Completo:** Cobertura de todos los escenarios
5. **Mantenimiento Fácil:** Archivos SQL organizados y comentados
6. **Enfoque Específico:** Solo para LIST LOT, sin ruido

## 🔄 Diferencias con SEED TOOL Original

| Aspecto | SEED TOOL Original | SEED TOOL 2 |
|---------|-------------------|-------------|
| **Propósito** | Datos generales | Solo LIST LOT |
| **Archivos** | 27+ archivos | 22 archivos |
| **Categorías** | Generales | Corregidas (9/13) |
| **Fechas** | Generales | Por estación |
| **Precios** | Generales | Realistas |
| **Enfoque** | Completo | Específico |

---

**¡Con SEED TOOL 2 podrás generar datos de prueba perfectos para verificar el funcionamiento de `list lot`!** 🚀
