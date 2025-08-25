Secuencia:

################################################
################################################

1. Usuario (manual)

################################################
################################################


2. Crar Proyecto (crea todas las entidades)
{
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
}

#### **🏗️ PROYECTO:**
- **ID:** 1 - "Construcción Torre Norte"
- **Admin Cost:** $15,000
- **Campaign:** 2024-2025
- **Customer:** Inmobiliaria Buenos Aires S.A.
- **Managers:** María López
- **Investors:** Fondo Capital Innovador (50%)

#### **🏞️ CAMPOS Y LOTES:**

| Campo | Lote | Hectáreas | Cultivo | Season | Estado |
|-------|------|-----------|---------|--------|--------|
| **Campo A** | Parcela A1 | 2.5 ha | Soja → Maíz | Invierno 2025 | ✅ Creado |
| **Campo A** | Parcela A2 | 3.0 ha | Soja → Maíz | Verano 2025 | ✅ Creado |
| **Campo B** | Parcela B1 | 1.2 ha | Soja → Maíz | Otoño 2025 | ✅ Creado |

#### **�� DATOS BASE:**
- **Crops:** ✅ 10 cultivos disponibles
- **Lease Types:** ✅ 4 tipos de arriendo
- **Labors:** ✅ Categorías disponibles
- **Workorders:** ❌ 0 workorders existentes

#### **📅 ESTADO DE FECHAS:**
- **`dates`:** `null` (no hay fechas configuradas)
- **`variety`:** `""` (vacío)

################################################
################################################

3. Crear Insumos
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

################################################
################################################

4. Crear Labores
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
            "category_id": 1,
            "contractor_name": "Juan Pérez",
            "description": "Siembra directa de maíz para calcular sowed_area"
        },
        {
            "name": "Cosecha Maíz",
            "price": 45.00,
            "category_id": 7,
            "contractor_name": "María García",
            "description": "Cosecha mecánica de maíz para calcular harvested_area"
        }
    ]
}'
```

################################################
################################################

5. Crear Orden Siembra (workorder)

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/workorders" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-SOWING-A1-001",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 1,
    "crop_id": 2,
    "labor_id": 3,
    "investor_id": 1,
    "date": "2025-01-15T00:00:00Z",
    "effective_area": 2.5
}'
```

################################################
################################################

6. Crear Orden Cosecha (workorder)
```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/workorders" \
  -H "Content-Type: application/json" \
  -d '{
    "number": "WO-HARVEST-A1-001",
    "project_id": 1,
    "field_id": 1,
    "lot_id": 1,
    "crop_id": 2,
    "labor_id": 4,
    "investor_id": 1,
    "date": "2025-06-15T00:00:00Z",
    "effective_area": 2.5
  }'
```

################################################
################################################

7. Toneladas (update lote)


Valores demo realistas

| Lote | Hectáreas | Rendimiento | Toneladas Totales |
|------|-----------|-------------|-------------------|
| **Parcela A1** | 2.5 ha | 8 tn/ha | **20 toneladas** |
| **Parcela A2** | 3.0 ha | 8 tn/ha | **24 toneladas** |
| **Parcela B1** | 1.2 ha | 8 tn/ha | **9.6 toneladas** |

### **�� VALORES RECOMENDADOS:**

#### **Para Parcela A1 (2.5 ha):**
```bash
curl --location --request PUT 'http://localhost:8080/api/v1/lots/1/tons' \
--header 'X-API-KEY: abc123secreta' \
--header 'X-USER-ID: 123' \
--header 'Content-Type: application/json' \
--data '{
    "tons": "20.0"
}'
```

#### **Para Parcela A2 (3.0 ha):**
```bash
curl --location --request PUT 'http://localhost:8080/api/v1/lots/2/tons' \
--header 'X-API-KEY: abc123secreta' \
--header 'X-USER-ID: 123' \
--header 'Content-Type: application/json' \
--data '{
    "tons": "24.0"
}'
```

#### **Para Parcela B1 (1.2 ha):**
```bash
curl --location --request PUT 'http://localhost:8080/api/v1/lots/3/tons' \
--header 'X-API-KEY: abc123secreta' \
--header 'X-USER-ID: 123' \
--header 'Content-Type: application/json' \
--data '{
    "tons": "9.6"
}'
```

################################################
################################################


8. Precio de venta (comercialización)

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 123" \
  -X POST "http://localhost:8080/api/v1/projects/1/commercializations" \
  -H "Content-Type: application/json" \
  -d '{
    "values": [
      {
        "crop_id": 2,
        "board_price": 200.00,
        "freight_cost": 10.00,
        "commercial_cost": 5.00
      },
      {
        "crop_id": 1,
        "board_price": 350.00,
        "freight_cost": 15.00,
        "commercial_cost": 8.00
      }
    ]
  }'
```

| Lote | Hectáreas | Toneladas | Precio/Ton | Ingreso Total | Ingreso/ha |
|------|-----------|-----------|------------|---------------|------------|
| **Parcela A1** | 2.5 ha | 20 tn | $200 | $4,000 | **$1,600/ha** |
| **Parcela A2** | 3.0 ha | 24 tn | $200 | $4,800 | **$1,600/ha** |
| **Parcela B1** | 1.2 ha | 9.6 tn | $200 | $1,920 | **$1,600/ha** |


################################################
################################################


9. Fechas (opcional)







¡Perfecto! Aquí tienes los curls correctos para crear las fechas de todos los lotes:

## 🎯 **CURLS CORRECTOS PARA FECHAS:**

### **✅ LOTE 1 - Parcela A1 (3 secuencias):**

```bash
curl -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -X PUT "http://localhost:8080/api/v1/lots/1" \
  -H "Content-Type: application/json" \
  -d '{
    "dates": [
      {
        "sowing_date": "2025-01-15T00:00:00Z",
        "harvest_date": "2025-06-15T00:00:00Z",
        "sequence": 1
      },
      {
        "sowing_date": "2025-09-15T00:00:00Z",
        "harvest_date": "2026-02-15T00:00:00Z",
        "sequence": 2
      },
      {
        "sowing_date": "2026-03-01T00:00:00Z",
        "harvest_date": "2026-08-01T00:00:00Z",
        "sequence": 3
      }
    ]
  }'
```

### **✅ LOTE 2 - Parcela A2 (3 secuencias):**

```bash
curl -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -X PUT "http://localhost:8080/api/v1/lots/2" \
  -H "Content-Type: application/json" \
  -d '{
    "dates": [
      {
        "sowing_date": "2025-02-01T00:00:00Z",
        "harvest_date": "2025-07-01T00:00:00Z",
        "sequence": 1
      },
      {
        "sowing_date": "2025-10-01T00:00:00Z",
        "harvest_date": "2026-03-01T00:00:00Z",
        "sequence": 2
      },
      {
        "sowing_date": "2026-04-01T00:00:00Z",
        "harvest_date": "2026-09-01T00:00:00Z",
        "sequence": 3
      }
    ]
  }'
```

### **✅ LOTE 3 - Parcela B1 (3 secuencias):**

```bash
curl -H "X-API-KEY: abc123secreta" \
  -H "X-USER-ID: 123" \
  -X PUT "http://localhost:8080/api/v1/lots/3" \
  -H "Content-Type: application/json" \
  -d '{
    "dates": [
      {
        "sowing_date": "2025-01-20T00:00:00Z",
        "harvest_date": "2025-06-20T00:00:00Z",
        "sequence": 1
      },
      {
        "sowing_date": "2025-09-20T00:00:00Z",
        "harvest_date": "2026-02-20T00:00:00Z",
        "sequence": 2
      },
      {
        "sowing_date": "2026-03-15T00:00:00Z",
        "harvest_date": "2026-08-15T00:00:00Z",
        "sequence": 3
      }
    ]
  }'
```


**¡Ejecuta estos curls en orden y tendrás fechas completas para todos los lotes!** 📅✨















################################################
################################################