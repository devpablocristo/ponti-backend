## **📋 SECUENCIA PARA CREAR DATOS DE PRUEBA - LIST LOT**

### **🎯 OBJETIVO:**
Crear datos de prueba coherentes que permitan verificar el correcto funcionamiento del endpoint `list lot` con valores diferenciados y realistas.

---

## **1. 🧑‍�� CREAR USUARIO DEMO**

```sql
INSERT INTO users (id, email, username, password, token_hash, refresh_tokens, id_rol, is_verified, active, created_by, updated_by, created_at, updated_at)
VALUES (
    123, 
    'demo@ponti.com', 
    'demo_user', 
    'demo_password', 
    'demo_token_hash', 
    '{}', 
    1, 
    TRUE, 
    TRUE, 
    1, 
    1, 
    NOW(), 
    NOW()
);
```

---

## **2. 🏗️ CREAR PROYECTO COMPLETO**

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/projects" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Construcción Torre Norte",
    "admin_cost": 15000,
    "customer": {
      "id": 4,
      "name": "Inmobiliaria Buenos Aires S.A.",
      "type": "tipo 1A"
    },
    "campaign": {
      "id": 1,
      "name": "Campaña Loteo 2025"
    },
    "managers": [
      {
        "id": 1,
        "name": "María López"
      },
      {
        "id": 1,
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
        "id": 1,
        "name": "Grupo Inversor del Sur",
        "percentage": 50
      }
    ],
    "fields": [
      {
        "name": "Campo A",
        "lease_type_id": 1,
        "lease_type_value": 100,
        "lots": [
          {
            "name": "Parcela A1",
            "hectares": 2.5,
            "previous_crop_id": 1,
            "current_crop_id": 2,
            "season": "Invierno 2025"
          },
          {
            "name": "Parcela A2",
            "hectares": 3.0,
            "previous_crop_id": 1,
            "current_crop_id": 2,
            "season": "Verano 2025"
          }
        ]
      },
      {
        "name": "Campo B",
        "lease_type_id": 2,
        "lease_type_percent": 15,
        "lots": [
          {
            "name": "Parcela B1", 
            "hectares": 1.2,
            "previous_crop_id": 1,
            "current_crop_id": 2,
            "season": "Otoño 2025"
          }
        ]
      }
    ]
  }'
```

**�� RESULTADO ESPERADO:**
- **Proyecto ID:** 1 - "Construcción Torre Norte"
- **Admin Cost:** $15,000
- **Campo A:** Arriendo fijo ($100/ha)
- **Campo B:** Arriendo por porcentaje (15% de ingresos)

---

## **3. 📦 CREAR INSUMOS**

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/supplies" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Semilla Maíz DK 72-10",
    "project_id": 1,
    "type_id": 1,
    "category_id": 1,
    "unit_id": 1,
    "price": 1.65
  }'
```

**�� RESULTADO ESPERADO:**
- **Supply ID:** 1 - "Semilla Maíz DK 72-10"
- **Precio:** $1.65 por unidad

---

## **4. �� CREAR LABORES (CATEGORÍAS CORREGIDAS)**

```bash
curl --location 'http://localhost:8080/api/v1/projects/1/labors' \
--header 'X-API-Key: abc123secreta' \
--header 'X-User-Id: 123' \
--header 'Content-Type: application/json' \
--data '{
    "labors": [
        {
            "name": "Siembra Maíz",
            "price": 25.50,
            "category_id": 9,  # ✅ CORRECTO: 9 para "Siembra"
            "contractor_name": "Juan Pérez",
            "description": "Siembra directa de maíz para calcular sowed_area"
        },
        {
            "name": "Cosecha Maíz",
            "price": 45.00,
            "category_id": 13, # ✅ CORRECTO: 13 para "Cosecha"
            "contractor_name": "María García",
            "description": "Cosecha mecánica de maíz para calcular harvested_area"
        }
    ]
}'
```

**�� RESULTADO ESPERADO:**
- **Labor ID 1:** "Siembra Maíz" - $25.50/ha - Categoría 9 (Siembra)
- **Labor ID 2:** "Cosecha Maíz" - $45.00/ha - Categoría 13 (Cosecha)

---

## **5. �� CREAR WORKORDERS (FECHAS CORREGIDAS)**

### **�� Lote 1 - Parcela A1 (Invierno 2025):**

```bash
# Siembra - Junio (Invierno)
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/workorders" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-SOWING-A1-001",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 1,
    "crop_id": 2,
    "labor_id": 1,  # ✅ Labor con category_id = 9 (Siembra)
    "investor_id": 1,
    "date": "2025-06-15T00:00:00Z",  # ✅ CORREGIDO: Junio (invierno)
    "effective_area": 2.5
  }'

# Cosecha - Diciembre
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/workorders" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-HARVEST-A1-001",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 1,
    "crop_id": 2,
    "labor_id": 2,  # ✅ Labor con category_id = 13 (Cosecha)
    "investor_id": 1,
    "date": "2025-12-15T00:00:00Z",  # ✅ CORREGIDO: Diciembre
    "effective_area": 2.5
  }'
```

### **�� Lote 2 - Parcela A2 (Verano 2025):**

```bash
# Siembra - Diciembre (Verano)
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/workorders" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-SOWING-A2-001",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 2,
    "crop_id": 2,
    "labor_id": 1,  # ✅ Labor con category_id = 9 (Siembra)
    "investor_id": 1,
    "date": "2025-12-15T00:00:00Z",  # ✅ CORREGIDO: Diciembre (verano)
    "effective_area": 3.0
  }'

# Cosecha - Junio
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/workorders" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-HARVEST-A2-001",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 2,
    "crop_id": 2,
    "labor_id": 2,  # ✅ Labor con category_id = 13 (Cosecha)
    "investor_id": 1,
    "date": "2025-06-15T00:00:00Z",  # ✅ CORREGIDO: Junio
    "effective_area": 3.0
  }'
```

### **�� Lote 3 - Parcela B1 (Otoño 2025):**

```bash
# Siembra - Marzo (Otoño)
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/workorders" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-SOWING-B1-001",
    "project_id": 1,
    "field_id": 2,
    "lot_id": 3,
    "crop_id": 2,
    "labor_id": 1,  # ✅ Labor con category_id = 9 (Siembra)
    "investor_id": 1,
    "date": "2025-03-15T00:00:00Z",  # ✅ CORREGIDO: Marzo (otoño)
    "effective_area": 1.2
  }'

# Cosecha - Septiembre
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/workorders" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-HARVEST-B1-001",
    "project_id": 1,
    "field_id": 2,
    "lot_id": 3,
    "crop_id": 2,
    "labor_id": 2,  # ✅ Labor con category_id = 13 (Cosecha)
    "investor_id": 1,
    "date": "2025-09-15T00:00:00Z",  # ✅ CORREGIDO: Septiembre
    "effective_area": 1.2
  }'
```

---

## **6. ⚖️ ACTUALIZAR TONELADAS (RENDIMIENTO REALISTA)**

### **📊 VALORES DEMO REALISTAS:**

| Lote | Hectáreas | Rendimiento | Toneladas Totales |
|------|-----------|-------------|-------------------|
| **Parcela A1** | 2.5 ha | 8 tn/ha | **20 toneladas** |
| **Parcela A2** | 3.0 ha | 8 tn/ha | **24 toneladas** |
| **Parcela B1** | 1.2 ha | 8 tn/ha | **9.6 toneladas** |

### **�� COMANDOS CURL:**

```bash
# Parcela A1 (2.5 ha × 8 tn/ha = 20 tn)
curl --location --request PUT 'http://localhost:8080/api/v1/lots/1/tons' \
--header 'X-API-KEY: abc123secreta' \
--header 'X-USER-ID: 123' \
--header 'Content-Type: application/json' \
--data '{"tons": "20.0"}'

# Parcela A2 (3.0 ha × 8 tn/ha = 24 tn)
curl --location --request PUT 'http://localhost:8080/api/v1/lots/2/tons' \
--header 'X-API-KEY: abc123secreta' \
--header 'X-USER-ID: 123' \
--header 'Content-Type: application/json' \
--data '{"tons": "24.0"}'

# Parcela B1 (1.2 ha × 8 tn/ha = 9.6 tn)
curl --location --request PUT 'http://localhost:8080/api/v1/lots/3/tons' \
--header 'X-API-KEY: abc123secreta' \
--header 'X-USER-ID: 123' \
--header 'Content-Type: application/json' \
--data '{"tons": "9.6"}'
```

---

## **7. 💰 CREAR PRECIOS DE COMERCIALIZACIÓN (REALISTAS)**

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/projects/1/commercializations" \
  -H "Content-Type: application/json" \
  -d '{
    "values": [
      {
        "crop_id": 2,  # Maíz
        "board_price": 600.00,  # ✅ CORREGIDO: $600/ton (realista)
        "freight_cost": 50.00,
        "commercial_cost": 100.00,
        "net_price": 450.00
      },
      {
        "crop_id": 1,  # Soja
        "board_price": 800.00,  # ✅ CORREGIDO: $800/ton (realista)
        "freight_cost": 50.00,
        "commercial_cost": 100.00,
        "net_price": 650.00
      }
    ]
  }'
```

**�� RESULTADO ESPERADO:**
| Lote | Hectáreas | Toneladas | Precio Neto/Ton | Ingreso Total | Ingreso/ha |
|------|-----------|-----------|-----------------|---------------|------------|
| **Parcela A1** | 2.5 ha | 20 tn | $450 | $9,000 | **$3,600/ha** |
| **Parcela A2** | 3.0 ha | 24 tn | $450 | $10,800 | **$3,600/ha** |
| **Parcela B1** | 1.2 ha | 9.6 tn | $450 | $4,320 | **$3,600/ha** |

---

## **8. 📅 CONFIGURAR FECHAS (OPCIONAL)**

```sql
-- Insertar fechas de siembra y cosecha
INSERT INTO lot_dates (lot_id, sowing_date, harvest_date, sequence, created_by, updated_by, created_at, updated_at) VALUES 
(1, '2025-06-15 00:00:00', '2025-12-15 00:00:00', 1, 123, 123, NOW(), NOW()),
(2, '2025-12-15 00:00:00', '2025-06-15 00:00:00', 1, 123, 123, NOW(), NOW()),
(3, '2025-03-15 00:00:00', '2025-09-15 00:00:00', 1, 123, 123, NOW(), NOW());
```

---

## **9. 🧪 VERIFICAR DATOS CREADOS**

### **📊 Verificar Workorders:**
```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/workorders?project_id=1"
```

### **📊 Verificar Lotes:**
```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/lots?project_id=1"
```

### **📊 Verificar Comercialización:**
```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/projects/1/commercializations"
```

---

## **10. �� VERIFICAR LIST LOT**

### **�� Endpoint Final:**
```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  "http://localhost:8080/api/v1/lots?project_id=1"
```

**�� RESULTADO ESPERADO:**
- **3 lots con datos diferenciados**
- **Ingresos realistas** basados en toneladas y precios
- **Costos directos** calculados correctamente
- **Arriendo diferenciado** (fijo vs porcentaje)
- **Resultados operativos** coherentes

---

## **✅ RESUMEN DE CORRECCIONES APLICADAS:**

1. **✅ Categorías de labores:** 9 (Siembra) y 13 (Cosecha)
2. **✅ Fechas de workorders:** Estaciones correctas
3. **✅ Precios de comercialización:** $450-650/ton (realistas)
4. **✅ Configuración de arriendo:** Fijo vs porcentaje
5. **✅ Rendimiento por hectárea:** 8 tn/ha (realista)
