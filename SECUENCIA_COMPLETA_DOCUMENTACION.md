# 🚀 **DOCUMENTACIÓN COMPLETA - SECUENCIA DE POBLADO DEL SISTEMA PONTI API**

## 📋 **ÍNDICE**
1. [Datos Base que Deben Existir](#datos-base-que-deben-existir)
2. [Secuencia de Creación de Entidades](#secuencia-de-creación-de-entidades)
3. [Creación de Datos Operativos](#creación-de-datos-operativos)
4. [Verificación de Datos Cargados](#verificación-de-datos-cargados)
5. [Explicación de Campos Calculados](#explicación-de-campos-calculados)

---

## 🗃️ **DATOS BASE QUE DEBEN EXISTIR**

### **⚠️ IMPORTANTE: Estos datos NO tienen endpoints, debes insertarlos directamente en la base**

```sql
-- 1. USERS (Usuarios) - SIN ENDPOINT
INSERT INTO users (id, email, username, password, token_hash, id_rol, is_verified, active, created_by, updated_by) 
VALUES (123, 'admin@test.com', 'admin123', 'hashedpassword', 'token123', 1, true, true, 123, 123);

-- 2. CROPS (Cultivos) - SIN ENDPOINT
INSERT INTO crops (id, name, created_by, updated_by) VALUES
(1, 'Soja', 123, 123),
(2, 'Maíz', 123, 123),
(3, 'Trigo', 123, 123),
(4, 'Girasol', 123, 123),
(5, 'Sorgo', 123, 123);

-- 3. LEASE_TYPES (Tipos de Arriendo) - SIN ENDPOINT  
INSERT INTO lease_types (id, name, created_by, updated_by) VALUES
(1, 'ARRIENDO FIJO', 123, 123),
(2, '% INGRESO NETO', 123, 123),
(3, 'ARRIENDO FIJO + % INGRESO NETO', 123, 123);

-- 4. LABOR_TYPES (Tipos de Labor) - SIN ENDPOINT
INSERT INTO labor_types (id, name, created_by, updated_by) VALUES
(1, 'Semilla', 123, 123),
(2, 'Agroquímicos', 123, 123),
(3, 'Fertilizantes', 123, 123),
(4, 'Labores', 123, 123);

-- 5. LABOR_CATEGORIES (Categorías de Labor) - SIN ENDPOINT
INSERT INTO labor_categories (id, name, type_id, created_by, updated_by) VALUES
(1, 'Siembra', 4, 123, 123),      -- Categoría 1 = Siembra
(2, 'Cosecha', 4, 123, 123),      -- Categoría 2 = Cosecha
(3, 'Fertilización', 3, 123, 123),
(4, 'Aplicación de Herbicidas', 2, 123, 123);

-- 6. TYPES (Tipos de Insumo) - SIN ENDPOINT
INSERT INTO types (id, name, created_by, updated_by) VALUES
(1, 'Semilla', 123, 123),
(2, 'Agroquímicos', 123, 123),
(3, 'Fertilizantes', 123, 123),
(4, 'Labores', 123, 123);

-- 7. CATEGORIES (Categorías de Insumo) - SIN ENDPOINT
INSERT INTO categories (id, name, type_id, created_by, updated_by) VALUES
(1, 'Semilla', 1, 123, 123),
(2, 'Herbicidas', 2, 123, 123),
(3, 'Fertilizantes', 3, 123, 123),
(4, 'Siembra', 4, 123, 123);

-- 8. UNITS (Unidades) - SIN ENDPOINT
INSERT INTO units (id, name, created_by, updated_by) VALUES
(1, 'kg', 123, 123),
(2, 'lt', 123, 123),
(3, 'ha', 123, 123);

-- 9. CAMPAIGNS (Campañas) - SIN ENDPOINT
INSERT INTO campaigns (id, name, created_by, updated_by) VALUES
(1, '2024-2025', 123, 123);

-- 10. PROVIDERS (Proveedores) - SIN ENDPOINT
INSERT INTO providers (id, name, created_by, updated_by) VALUES
(1, 'Proveedor Semillas S.A.', 123, 123),
(2, 'Agroquímicos del Norte', 123, 123);
```

---

## 🚀 **SECUENCIA DE CREACIÓN DE ENTIDADES**

### **PASO 1: Crear Customer (Cliente)**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Inmobiliaria Buenos Aires S.A."
  }' \
  http://localhost:8080/api/v1/customers
```

**Respuesta esperada:**
```json
{
  "id": 1,
  "name": "Inmobiliaria Buenos Aires S.A."
}
```

### **PASO 2: Crear Manager (Gerente)**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "María López"
  }' \
  http://localhost:8080/api/v1/managers
```

**Respuesta esperada:**
```json
{
  "id": 1,
  "name": "María López"
}
```

### **PASO 3: Crear Segundo Manager**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Juan Pérez"
  }' \
  http://localhost:8080/api/v1/managers
```

**Respuesta esperada:**
```json
{
  "id": 2,
  "name": "Juan Pérez"
}
```

### **PASO 4: Crear Investor (Inversor)**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Fondo Capital Innovador"
  }' \
  http://localhost:8080/api/v1/investors
```

**Respuesta esperada:**
```json
{
  "id": 1,
  "name": "Fondo Capital Innovador"
}
```

### **PASO 5: Crear Segundo Investor**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Grupo Inversor del Sur"
  }' \
  http://localhost:8080/api/v1/investors
```

**Respuesta esperada:**
```json
{
  "id": 2,
  "name": "Grupo Inversor del Sur"
}
```

### **PASO 6: Crear Project Completo**
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

**Respuesta esperada:**
```json
{
  "id": 1,
  "name": "Construcción Torre Norte",
  "fields": [
    {
      "id": 1,
      "name": "Campo A",
      "lots": [
        {
          "id": 1,
          "name": "Parcela A1"
        },
        {
          "id": 2,
          "name": "Parcela A2"
        }
      ]
    },
    {
      "id": 2,
      "name": "Campo B",
      "lots": [
        {
          "id": 3,
          "name": "Parcela B1"
        }
      ]
    }
  ]
}
```

---

## 🔧 **CREACIÓN DE DATOS OPERATIVOS**

### **PASO 7: Crear Supply (Insumo) para Semilla**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Semilla Soja Premium",
    "project_id": 1,
    "type_id": 1,
    "category_id": 1,
    "unit_id": 1,
    "price": "150.00"
  }' \
  http://localhost:8080/api/v1/supplies
```

**¿Por qué estos campos?**
- `project_id`: 1 (tu proyecto "Construcción Torre Norte")
- `type_id`: 1 (Semilla - según labor_types)
- `category_id`: 1 (Semilla - según labor_categories) 
- `unit_id`: 1 (kg - unidad para semillas)
- `price`: "150.00" (precio por kg)

### **PASO 8: Crear Labor (Labor) para Siembra**
```bash
curl --location 'http://localhost:8080/api/v1/projects/1/labors' \
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

**¿Por qué estos campos?**
- `category_id`: 1 (Siembra - según labor_categories)
- `price`: Precio por hectárea de la labor

### **PASO 9: Crear Labor de Cosecha**
```bash
curl --location 'http://localhost:8080/api/v1/projects/1/labors' \
--header 'X-API-Key: abc123secreta' \
--header 'X-User-Id: 123' \
--header 'Content-Type: application/json' \
--data '{
  "labors": [
    {
      "name": "Cosecha de Soja",
      "price": 35.00,
      "category_id": 2,
      "contractor_name": "Cosechadora ABC",
      "description": "Cosecha mecanizada de soja"
    }
  ]
}'
```

**¿Por qué estos campos?**
- `category_id`: 2 (Cosecha - según labor_categories)
- `price`: Precio por hectárea de la cosecha

### **PASO 10: Crear Supply Movement (Movimiento de Insumo)**
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
        "project_destination_id": 1,
        "supply_id": 1,
        "investor_id": 1,
        "provider": {
          "id": 1,
          "name": "Proveedor Semillas S.A."
        }
      }
    ]
  }' \
  http://localhost:8080/api/v1/projects/1/supply-movements
```

**¿Por qué estos campos?**
- `supply_id`: 1 (el supply de semilla que creaste)
- `investor_id`: 1 (Fondo Capital Innovador)
- `project_destination_id`: 1 (tu proyecto)

### **PASO 11: Crear Workorder para Lote 1**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-001",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 1,
    "crop_id": 2,
    "labor_id": 1,
    "investor_id": 1,
    "date": "2025-01-15T00:00:00Z",
    "effective_area": "2.5",
    "contractor": "Juan Pérez",
    "observations": "Siembra de soja en parcela A1"
  }' \
  http://localhost:8080/api/v1/workorders
```

**¿Por qué estos campos?**
- `labor_id`: 1 (labor de siembra)
- `effective_area`: "2.5" (hectáreas del lote)
- `investor_id`: 1 (Fondo Capital Innovador)

### **PASO 12: Crear Workorder para Lote 2**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-002",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 2,
    "crop_id": 2,
    "labor_id": 1,
    "investor_id": 1,
    "date": "2025-01-20T00:00:00Z",
    "effective_area": "3.0",
    "contractor": "Juan Pérez",
    "observations": "Siembra de soja en parcela A2"
  }' \
  http://localhost:8080/api/v1/workorders
```

### **PASO 13: Crear Workorder de Cosecha para Lote 1**
```bash
curl -X POST -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-003",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 1,
    "crop_id": 2,
    "labor_id": 3,
    "investor_id": 1,
    "date": "2025-05-15T00:00:00Z",
    "effective_area": "2.5",
    "contractor": "Cosechadora ABC",
    "observations": "Cosecha de soja en parcela A1"
  }' \
  http://localhost:8080/api/v1/workorders
```

**¿Por qué estos campos?**
- `labor_id`: 3 (labor de cosecha)
- `effective_area`: "2.5" (área cosechada)

### **PASO 14: Crear Commercializations para el Proyecto**
```bash
curl --location 'http://localhost:8080/api/v1/projects/1/commercializations' \
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
      "net_price": "97.50"
    }
  ]
}'
```

---

## 📊 **VERIFICACIÓN DE DATOS CARGADOS**

### **Verificar que todos los datos estén cargados:**

```bash
# 1. Verificar proyecto
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/projects/1"

# 2. Verificar campos
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/projects/1/fields"

# 3. Verificar lotes
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/lots?project_id=1&field_id=1"

# 4. Verificar supplies
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/projects/1/supplies"

# 5. Verificar labors
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/projects/1/labors"

# 6. Verificar workorders
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/workorders?project_id=1"
```

---

## 🎯 **EXPLICACIÓN DE CAMPOS CALCULADOS**

### **¿Por qué algunos campos están en 0?**

Los campos calculados están en 0 porque dependen de otros datos que aún no se han llenado:

#### **1. `harvested_area` = "0"**
- **¿Qué es?** Área efectivamente cosechada del lote
- **¿Por qué está en 0?** Porque no se han creado **workorders de cosecha** (labores de categoría 2)

#### **2. `yield_tn_per_ha` = "0"**
- **¿Qué es?** Rendimiento en toneladas por hectárea cosechada
- **¿Por qué está en 0?** Porque `harvested_area = 0` y `tons = 0`

#### **3. `income_net_per_ha` = "0"**
- **¿Qué es?** Ingreso neto por hectárea sembrada
- **¿Por qué está en 0?** Porque `tons = 0` (no hay cosecha registrada)

#### **4. `rent_per_ha` = "0"**
- **¿Qué es?** Costo de arriendo por hectárea
- **¿Por qué está en 0?** Porque depende del tipo de arriendo y los ingresos

### **Para llenar estos campos necesitas:**

#### **Paso 1: Registrar Toneladas Cosechadas**
```bash
curl -X PUT -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -H "Content-Type: application/json" \
  -d '{
    "tons": "12.5"
  }' \
  http://localhost:8080/api/v1/lots/1
```

#### **Paso 2: Llenar Fechas de Siembra/Cosecha**
```bash
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

### **Después de estos pasos:**

- **`harvested_area`** = "2.5" (área cosechada del workorder)
- **`yield_tn_per_ha`** = "5.0" (12.5 tons / 2.5 ha)
- **`income_net_per_ha`** = "610.00" (12.5 tons × $122 net_price / 2.5 ha)
- **`rent_per_ha`** = Se calcula según el tipo de arriendo del campo

---

## 🔍 **VISTA `lot_table_view` - Cómo Funciona**

### **CTEs (Common Table Expressions) utilizados:**

#### **1. `sowing` - Área Sembrada**
```sql
-- Suma workorders de categoría 1 (Siembra)
SELECT w.lot_id, SUM(w.effective_area) AS sowed_area
FROM workorders w
JOIN labors lb ON lb.id = w.labor_id
WHERE lb.category_id = 1  -- Categoría 1 = Siembra
```

#### **2. `harvest` - Área Cosechada**
```sql
-- Suma workorders de categoría 2 (Cosecha)
SELECT w.lot_id, SUM(w.effective_area) AS harvested_area
FROM workorders w
JOIN labors lb ON lb.id = w.labor_id
WHERE lb.category_id = 2  -- Categoría 2 = Cosecha
```

#### **3. `direct_costs` - Costos Directos**
```sql
-- Suma costos de labor + insumos
SELECT w.lot_id,
  SUM(lb.price * w.effective_area) AS labor_cost,
  SUM(wi.final_dose * s.price * w.effective_area) AS supply_cost
FROM workorders w
JOIN labors lb ON lb.id = w.labor_id
LEFT JOIN workorder_items wi ON w.id = wi.workorder_id
LEFT JOIN supplies s ON s.id = wi.supply_id
```

#### **4. `income_net` - Ingreso Neto**
```sql
-- Calcula ingresos basados en toneladas y precio neto
SELECT l.id AS lot_id,
  COALESCE(l.tons, 0) * COALESCE(cc.net_price, 0) AS income_net_total
FROM lots l
LEFT JOIN crop_commercializations cc ON cc.project_id = f.project_id 
  AND cc.crop_id = l.current_crop_id
```

#### **5. `rent_calculation` - Cálculo de Arriendo**
```sql
-- Calcula arriendo según tipo (fijo, porcentaje, o ambos)
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

---

## 📈 **FLUJO COMPLETO DEL SISTEMA**

```
1. 🌱 DATOS BASE → crops, lease_types, labor_types, etc.
2. 👥 ENTIDADES → customers, managers, investors
3. 🏗️ PROYECTO → project + fields + lots
4. 📦 INSUMOS → supplies + supply_movements
5. 🚜 LABORES → labors (siembra, cosecha, etc.)
6. 📋 WORKORDERS → ejecución de labores
7. ⚖️ COSECHA → registro de tons + fechas
8. 💰 COMERCIALIZACIÓN → precios de venta
9. 📊 CÁLCULOS → vista lot_table_view
```

---

## ✅ **RESUMEN FINAL**

**¡SÍ puedes construir tu tabla de lotes completa!** Sigue esta secuencia:

1. **Datos base** → Insertar manualmente en SQL
2. **Entidades** → Crear via API (customers, managers, investors)
3. **Proyecto** → Crear proyecto completo con campos y lotes
4. **Operaciones** → Crear supplies, labors, workorders
5. **Ejecución** → Registrar cosecha y toneladas
6. **Resultado** → Vista `lot_table_view` con métricas completas

La tabla se poblará inicialmente con datos básicos y se enriquecerá automáticamente con métricas cuando agregues workorders, supplies y commercializations.

**La vista `lot_table_view` es tu motor de cálculo financiero** que convierte datos operativos en métricas financieras automáticamente.
