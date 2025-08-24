Campaing ya existe como base
1	"2024-2025"	"2025-08-24 15:54:39.927602+00"	"2025-08-24 15:54:39.927602+00"	


-- Crear usuario con ID 123
INSERT INTO users (id, email, username, password, token_hash, id_rol, is_verified, active, created_by, updated_by) 
VALUES (123, 'admin@test.com', 'admin123', 'hashedpassword', 'token123', 1, true, true, 123, 123);



## 🚀 **PASO 1: Crear Customer (Cliente)**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Inmobiliaria Buenos Aires S.A."
  }' \
  http://localhost:8080/api/v1/customers
```

## 🚀 **PASO 2: Crear Manager (Gerente)**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "María López"
  }' \
  http://localhost:8080/api/v1/managers
```

## 🚀 **PASO 3: Crear Segundo Manager**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Juan Pérez"
  }' \
  http://localhost:8080/api/v1/managers
```

## 🚀 **PASO 4: Crear Investor (Inversor)**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fondo Capital Innovador"
  }' \
  http://localhost:8080/api/v1/investors
```

## 🚀 **PASO 5: Crear Segundo Investor**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Grupo Inversor del Sur"
  }' \
  http://localhost:8080/api/v1/investors
```

## �� **Verificar IDs Obtenidos**

Después de cada paso, anota el ID que te devuelve:

- Customer ID: `?`
- Manager 1 ID: `?` (María López)
- Manager 2 ID: `?` (Juan Pérez)  
- Investor 1 ID: `?` (Fondo Capital Innovador)
- Investor 2 ID: `?` (Grupo Inversor del Sur)








## 🚀 **PASO 6: Crear Project Completo (Cuando tengas todos los IDs)**

```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Construcción Torre Norte",
    "admin_cost": "15000.00",
    "customer": {
      "id": 1,
      "name": "Inmobiliaria Buenos Aires S.A."
    },
    "campaign": {
      "id": 1,
      "name": "2024-2025"
    },
    "managers": [
      {
        "id": 1,
        "name": "María López"
      },
      {
        "id": 2,
        "name": "Juan Pérez"
      }
    ],
    "investors": [
      {
        "id": 1,
        "name": "Fondo Capital Innovador",
        "percentage": 50
      },
      {
        "id": 2,
        "name": "Grupo Inversor del Sur",
        "percentage": 50
      }
    ],
    "fields": [
      {
        "name": "Campo A",
        "lease_type_id": 1,
        "lease_type_percent": 25.0,
        "lease_type_value": 100.0,
        "lots": [
          {
            "name": "Parcela A1",
            "hectares": "2.5",
            "previous_crop_id": 1,
            "current_crop_id": 2,
            "season": "Invierno 2025"
          },
          {
            "name": "Parcela A2",
            "hectares": "3.0",
            "previous_crop_id": 1,
            "current_crop_id": 2,
            "season": "Verano 2025"
          }
        ]
      },
      {
        "name": "Campo B",
        "lease_type_id": 2,
        "lease_type_percent": 30.0,
        "lease_type_value": 150.0,
        "lots": [
          {
            "name": "Parcela B1", 
            "hectares": "1.2",
            "previous_crop_id": 1,
            "current_crop_id": 2,
            "season": "Otoño 2025"
          }
        ]
      }
    ]
  }' \
  http://localhost:8080/api/v1/projects
```

PROYECTO CREADO








































































## 🔧 **SECUENCIA COMPLETA CORREGIDA PARA LLENAR LOS CAMPOS**

### **PASO 0: Crear Datos Base (Sin Endpoints)**

```sql
-- 1. UNITS (Unidades) - SIN ENDPOINT
INSERT INTO units (id, name, created_by, updated_by) VALUES
(1, 'kg', 123, 123),
(2, 'lt', 123, 123),
(3, 'ha', 123, 123);



### **PASO 1: Crear Supply (Insumo) para Semilla - CORREGIDO**

```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Semilla Soja Premium",
    "project_id": 3,
    "type_id": 1,
    "category_id": 1,
    "unit_id": 1,
    "price": "150.00"
  }' \
  http://localhost:8080/api/v1/supplies
```

**¿Por qué estos campos?**
- `project_id`: 3 (tu proyecto "Construcción Torre Norte")
- `type_id`: 1 (Semilla - según labor_types)
- `category_id`: 1 (Semilla - según labor_categories) 
- `unit_id`: 1 (kg - unidad para semillas)
- `price`: "150.00" (precio por kg)

### **PASO 2: Crear Labor (Labor) para Siembra - CORREGIDO**

```bash
curl --location 'http://localhost:8080/api/v1/projects/3/labors' \
--header 'X-API-Key: abc123secreta' \
--header 'X-User-Id: 123' \
--header 'Content-Type: application/json' \
--data '{
  "labors": [
    {
      "name": "Siembra de Soja",
      "price": 25.50,
      "category_id": 1,
      "contractor_name": "Juan Pérez",
      "description": "Siembra directa de soja"
    },
    {
      "name": "Aplicación de Herbicida",
      "price": 15.75,
      "category_id": 2,
      "contractor_name": "María García",
      "description": "Aplicación de glifosato"
    }
  ]
}'
```












































































Basándome en el código que has compartido, aquí está la secuencia completa de pasos para crear workorders y commercializations:

## �� **Secuencia Completa de Pasos**

### **PASO 1: Crear Workorder para Lote 1**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-001",
    "project_id": 3,
    "field_id": 1,
    "lot_id": 1,
    "crop_id": 2,
    "labor_id": 1,
    "investor_id": 2,
    "date": "2025-01-15T00:00:00Z",
    "effective_area": "2.5",
    "contractor": "Juan Pérez",
    "observations": "Siembra de soja en parcela A1"
  }' \
  http://localhost:8080/api/v1/workorders
```

### **PASO 2: Crear Workorder para Lote 2**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-002",
    "project_id": 3,
    "field_id": 1,
    "lot_id": 2,
    "crop_id": 2,
    "labor_id": 1,
    "investor_id": 2,
    "date": "2025-01-20T00:00:00Z",
    "effective_area": "3.0",
    "contractor": "Juan Pérez",
    "observations": "Siembra de soja en parcela A2"
  }' \
  http://localhost:8080/api/v1/workorders
```

### **PASO 3: Crear Commercializations para el Proyecto 3**
```bash
curl --location 'http://localhost:8080/api/v1/projects/3/commercializations' \
--header 'Content-Type: application/json' \
--header 'X-API-KEY: abc123secreta' \
--header 'X-USER-ID: 123' \
--data '{
  "values": [
    {
      "crop_id": 1,
      "crop_name": "Soja",
      "board_price": "150.00",
      "freight_cost": "18.00",
      "commercial_cost": "10.00",
      "net_price": "122.00"
    },
    {
      "crop_id": 2,
      "crop_name": "Maíz",
      "board_price": "120.50",
      "freight_cost": "15.00",
      "commercial_cost": "8.00",
      "net_price": "400.00"
    }
  ]
}'
```


















### **PASO 2: Crear Supply Movement (Movimiento de Insumo) - CORREGIDO**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "quantity": "50.0",
        "movement_type": "Stock",
        "movement_date": "2025-01-15T00:00:00Z",
        "reference_number": "REF-001",
        "project_destination_id": 3,
        "supply_id": 1,
        "investor_id": 2,
        "provider": {
          "id": 0,
          "name": "Proveedor Semillas S.A."
        }
      }
    ]
  }' \
  http://localhost:8080/api/v1/projects/3/supply-movements
```





¡Excelente pregunta! Déjame aclarar la relación entre estos campos:

## �� **Relación entre `tons`, `sowing_date` y `harvest_date`:**

### **1. `tons` (Toneladas)**
- **Campo manual** - Se ingresa manualmente después de la cosecha
- **Ubicación:** Tabla `lots` (campo directo del lote)
- **¿Cuándo se llena?** Después de cosechar y pesar el cultivo

### **2. `sowing_date` y `harvest_date` (Fechas)**
- **Campos manuales** - Se ingresan manualmente al programar/ejecutar
- **Ubicación:** Tabla `lot_dates` (tabla relacionada con `lots`)
- **¿Cuándo se llenan?** Al programar la siembra y cosecha

## 📊 **¿Están Relacionados?**

**SÍ, están relacionados lógicamente, pero NO en la base de datos:**

### **Relación Lógica:**
```
🌱 Siembra (sowing_date) → �� Crecimiento → 🚜 Cosecha (harvest_date) → ⚖️ Peso (tons)
```

### **Relación en la Base de Datos:**
- **`tons`** está en la tabla `lots`
- **`sowing_date` y `harvest_date`** están en la tabla `lot_dates`
- **Se relacionan por `lot_id`** (clave foránea)

## �� **Cómo se Relacionan en el Código:**

Según la vista `lot_table_view`, estos campos se usan juntos para cálculos:

### **Rendimiento por Hectárea:**
```sql
-- Rendimiento (toneladas por hectárea cosechada)
CASE 
  WHEN COALESCE(h.harvested_area, 0) > 0 
  THEN COALESCE(l.tons, 0) / h.harvested_area
  ELSE 0 
END AS yield_tn_per_ha
```

### **Ingreso Neto:**
```sql
-- Ingreso neto basado en toneladas cosechadas
COALESCE(l.tons, 0) * COALESCE(cc.net_price, 0) AS income_net_total
```

## �� **Flujo de Llenado Recomendado:**

### **Paso 1: Programar Siembra**
```bash
# Llenar sowing_date al programar
curl -X PUT -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "dates": [
      {
        "sowing_date": "2025-01-15",
        "sequence": 1
      }
    ]
  }' \
  http://localhost:8080/api/v1/lots/1
```

### **Paso 2: Programar Cosecha**
```bash
# Llenar harvest_date al programar
curl -X PUT -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "dates": [
      {
        "sowing_date": "2025-01-15",
        "harvest_date": "2025-05-15",
        "sequence": 1
      }
    ]
  }' \
  http://localhost:8080/api/v1/lots/1
```

### **Paso 3: Registrar Cosecha**
```bash
# Llenar tons después de cosechar
curl -X PUT -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "tons": "12.5"
  }' \
  http://localhost:8080/api/v1/lots/1
```

## 🎯 **Resumen de Relaciones:**

| Campo | Tabla | Cuándo se llena | Relación |
|-------|-------|------------------|----------|
| `sowing_date` | `lot_dates` | Al programar siembra | **Lógica** con `harvest_date` |
| `harvest_date` | `lot_dates` | Al programar cosecha | **Lógica** con `sowing_date` |
| `tons` | `lots` | Después de cosechar | **Lógica** con `harvest_date` |

## �� **Validaciones del Sistema:**

El sistema valida que:
- **`sowing_date`** debe ser anterior a **`harvest_date`**
- **`tons`** solo se puede llenar después de **`harvest_date`**
- **`sequence`** debe ser único por lote (1, 2, o 3)

## 💡 **Conclusión:**

**SÍ están relacionados lógicamente** (flujo agrícola), pero **NO están relacionados directamente en la base de datos**. El sistema los conecta a través de la lógica de negocio para calcular rendimientos, ingresos y métricas del proyecto.

La relación es: **Siembra → Crecimiento → Cosecha → Peso**, y cada campo se llena en su momento correspondiente del ciclo agrícola.


























Excelente pregunta! Estos campos están en 0 porque dependen de otros datos que aún no se han llenado. Déjame explicarte cada uno:

## 🌾 **¿Por qué están en 0?**

### **1. `harvested_area` = "0"**
**¿Qué es?** Área efectivamente cosechada del lote
**¿Por qué está en 0?** Porque no se han creado **workorders de cosecha** (labores de categoría 2)

**Según la vista SQL:**
```sql
harvest AS (
  SELECT 
    w.lot_id,
    SUM(w.effective_area) AS harvested_area,  -- Se suma de workorders
    MAX(w.date) AS last_harvest_date
  FROM workorders w
  JOIN labors lb ON lb.id = w.labor_id
  WHERE lb.category_id = 2  -- ID de "Categoría 2" (Cosecha)
    AND w.effective_area > 0
  GROUP BY w.lot_id
)
```

### **2. `yield_tn_per_ha` = "0"**
**¿Qué es?** Rendimiento en toneladas por hectárea cosechada
**¿Por qué está en 0?** Porque `harvested_area = 0` y `tons = 0`

**Fórmula:**
```sql
CASE 
  WHEN COALESCE(h.harvested_area, 0) > 0 
  THEN COALESCE(l.tons, 0) / h.harvested_area  -- tons / harvested_area
  ELSE 0 
END AS yield_tn_per_ha
```

### **3. `income_net_per_ha` = "0"**
**¿Qué es?** Ingreso neto por hectárea sembrada
**¿Por qué está en 0?** Porque `tons = 0` (no hay cosecha registrada)

**Fórmula:**
```sql
CASE 
  WHEN COALESCE(s.sowed_area, 0) > 0 
  THEN COALESCE(in_net.income_net_total, 0) / s.sowed_area
  ELSE 0 
END AS income_net_per_ha
```

**Donde `income_net_total` es:**
```sql
COALESCE(l.tons, 0) * COALESCE(cc.net_price, 0) AS income_net_total
```

### **4. `rent_per_ha` = "0"**
**¿Qué es?** Costo de arriendo por hectárea
**¿Por qué está en 0?** Porque depende del tipo de arriendo y los ingresos

**Fórmula:**
```sql
CASE 
  WHEN f.lease_type_id = 1 THEN -- Fixed amount
    COALESCE(f.lease_type_value, 0) * COALESCE(h.harvested_area, 0)
  WHEN f.lease_type_id = 2 THEN -- Percentage
    (COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0)
  WHEN f.lease_type_id = 3 THEN -- Both
    (COALESCE(f.lease_type_value, 0) * COALESCE(h.harvested_area, 0)) + 
    ((COALESCE(f.lease_type_percent, 0) / 100.0) * COALESCE(in_net.income_net_total, 0))
  ELSE 0
END AS rent_total
```

## 🔧 **Para llenar estos campos necesitas:**

### **Paso 1: Crear Labor de Cosecha**
```bash
curl --location 'http://localhost:8080/api/v1/projects/3/labors' \
--header 'X-API-Key: abc123secreta' \
--header 'X-User-Id: 123' \
--header 'Content-Type: application/json' \
--data '{
  "labors": [
    {
      "name": "Cosecha de Soja",
      "price": 35.00,
      "category_id": 2,  -- IMPORTANTE: Categoría 2 = Cosecha
      "contractor_name": "Cosechadora ABC",
      "description": "Cosecha mecanizada de soja"
    }
  ]
}'
```

### **Paso 2: Crear Workorder de Cosecha**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-003",
    "project_id": 3,
    "field_id": 1,
    "lot_id": 1,
    "crop_id": 2,
    "labor_id": 2,  -- ID del labor de cosecha que creaste
    "investor_id": 2,
    "date": "2025-05-15T00:00:00Z",
    "effective_area": "2.5",  -- Área cosechada
    "contractor": "Cosechadora ABC",
    "observations": "Cosecha de soja en parcela A1"
  }' \
  http://localhost:8080/api/v1/workorders
```

### **Paso 3: Registrar Toneladas Cosechadas**
```bash
curl -X PUT -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "tons": "12.5"
  }' \
  http://localhost:8080/api/v1/lots/1
```

## �� **Después de estos pasos:**

- **`harvested_area`** = "2.5" (área cosechada del workorder)
- **`yield_tn_per_ha`** = "5.0" (12.5 tons / 2.5 ha)
- **`income_net_per_ha`** = "610.00" (12.5 tons × $122 net_price / 2.5 ha)
- **`rent_per_ha`** = Se calcula según el tipo de arriendo del campo

## 🎯 **Resumen:**

Estos campos están en 0 porque:
1. **No hay workorders de cosecha** (categoría 2)
2. **No hay toneladas registradas** (`tons = 0`)
3. **No hay ingresos calculados** (dependen de `tons`)
4. **No hay arriendo calculado** (depende de ingresos y tipo de arriendo)

**La secuencia correcta es:** Siembra → Cosecha → Registro de peso → Cálculos automáticos.












































**¿Qué hace?** Obtiene fechas de siembra/cosecha y cuenta workorders

## 🧮 **Cálculo de Cada Columna del SELECT Principal:**

### **�� CAMPOS BÁSICOS:**
- **`sowed_area`**: `COALESCE(s.sowed_area, 0)` - Área sembrada del CTE sowing
- **`tons`**: `COALESCE(l.tons, 0)` - Toneladas cosechadas del lote
- **`harvested_area`**: `COALESCE(h.harvested_area, 0)` - Área cosechada del CTE harvest

### **💰 COSTOS DIRECTOS:**
- **`direct_cost_total`**: `COALESCE(dc.labor_cost, 0) + COALESCE(dc.supply_cost, 0)`
- **`cost_usd_per_ha`**: `direct_cost_total / sowed_area` (si sowed_area > 0)

### **💵 INGRESOS:**
- **`income_net_total`**: `COALESCE(in_net.income_net_total, 0)` - Del CTE income_net
- **`income_net_per_ha`**: `income_net_total / sowed_area` (si sowed_area > 0)

### **🌾 RENDIMIENTO:**
- **`yield_tn_per_ha`**: `tons / harvested_area` (si harvested_area > 0)

### **🏠 ARRIENDO:**
- **`rent_total`**: `COALESCE(rc.rent_total, 0)` - Del CTE rent_calculation
- **`rent_per_ha`**: `rent_total / sowed_area` (si sowed_area > 0)

### **�� COSTOS ADMINISTRATIVOS:**
- **`admin_total`**: `COALESCE(ac.admin_total, 0)` - Del CTE admin_cost
- **`admin_cost_per_ha`**: `admin_total / sowed_area` (si sowed_area > 0)

### **📈 TOTALES ACTIVOS:**
- **`active_total`**: `direct_cost_total + rent_total + admin_total`
- **`active_total_per_ha`**: `active_total / sowed_area` (si sowed_area > 0)

### **�� RESULTADO OPERATIVO:**
- **`operating_result`**: `income_net_total - active_total`
- **`operating_result_per_ha`**: `operating_result / sowed_area` (si sowed_area > 0)

## 🔍 **Flujo de Cálculos:**

```
1. 🌱 Siembra → sowed_area → cost_usd_per_ha
2. 🚜 Cosecha → harvested_area → yield_tn_per_ha  
3. ⚖️ Peso → tons → income_net_total → income_net_per_ha
4. 🏠 Arriendo → rent_total → rent_per_ha
5. 📋 Admin → admin_total → admin_cost_per_ha
6. 📊 Totales → active_total → active_total_per_ha
7. 🎯 Resultado → operating_result → operating_result_per_ha
```

## �� **Ventajas de esta Vista:**

1. **Cálculos automáticos** - No hay que calcular manualmente
2. **Consistencia** - Todos los cálculos usan la misma lógica
3. **Performance** - Índices optimizados para GCP
4. **Flexibilidad** - Maneja diferentes tipos de arriendo
5. **Auditoría** - Traza todos los costos e ingresos

Esta vista es esencialmente un **motor de cálculo financiero** que convierte datos operativos (siembra, cosecha, costos) en métricas financieras (ingresos, rentabilidad, costos por hectárea).