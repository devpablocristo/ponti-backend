Secuencia:

################################################
################################################

1. Usuario (manual)

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

5. Crear Ordenes

## **Lote 1 - Workorder de Siembra:**
 Siembra (workorder)

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
    "labor_id": 1,
    "investor_id": 1,
    "date": "2025-01-15T00:00:00Z",
    "effective_area": 2.5
}'
```

## **Lote 1 - Workorder de cosecha:**

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
    "labor_id": 2,
    "investor_id": 1,
    "date": "2025-06-15T00:00:00Z",
    "effective_area": 2.5
  }'
```

## **Lote 2 - Workorder de Siembra:**

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 1" -X POST "http://localhost:8080/api/v1/workorders" -H "Content-Type: application/json" -d '{"number": "WO-SOWING-A2-001", "project_id": 1, "field_id": 1, "lot_id": 2, "crop_id": 2, "labor_id": 1, "investor_id": 1, "date": "2025-02-15T00:00:00Z", "effective_area": 3.0}'
```
## **Lote 2 - Workorder de Cosecha:**

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 1" -X POST "http://localhost:8080/api/v1/workorders" -H "Content-Type: application/json" -d '{"number": "WO-HARVEST-A2-001", "project_id": 1, "field_id": 1, "lot_id": 2, "crop_id": 2, "labor_id": 2, "investor_id": 1, "date": "2025-07-15T00:00:00Z", "effective_area": 3.0}'
```
Ahora voy a crear workorders para el **Lote 3** (Parcela B1):

## **Lote 3 - Workorder de Siembra:**

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 1" -X POST "http://localhost:8080/api/v1/workorders" -H "Content-Type: application/json" -d '{"number": "WO-SOWING-B1-001", "project_id": 1, "field_id": 2, "lot_id": 3, "crop_id": 2, "labor_id": 1, "investor_id": 1, "date": "2025-03-15T00:00:00Z", "effective_area": 4.0}'
```
## **Lote 3 - Workorder de Cosecha:**

```bash
curl -H "X-API-KEY: abc123secreta" -H "X-USER-ID: 1" -X POST "http://localhost:8080/api/v1/workorders" -H "Content-Type: application/json" -d '{"number": "WO-HARVEST-B1-001", "project_id": 1, "field_id": 2, "lot_id": 3, "crop_id": 2, "labor_id": 2, "investor_id": 1, "date": "2025-08-15T00:00:00Z", "effective_area": 4.0}'
```

################################################
################################################

6. Toneladas (update lote)


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


7. Precio de venta (comercialización)

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


8. Fechas (opcional)
  

## 🚨 **PROBLEMA IDENTIFICADO:**

El error "parsing time" con "extra text" no viene del código Go del repositorio, sino de algún middleware de binding que está interfiriendo con el formato de fecha antes de que llegue a la validación de Go.

## �� **ANÁLISIS DEL CÓDIGO:**

Revisando el código del repositorio `lot`, veo que:

1. **El modelo `LotDates` está bien definido** con los tags de GORM correctos
2. **La función `UpdateLot` maneja correctamente las fechas** usando `clause.OnConflict` para upsert
3. **Los DTOs parsean correctamente las fechas** usando `time.Parse("2006-01-02", date.SowingDate)`
4. **El middleware `ValidateLotUpdate` está comentado** (líneas 375-395 en `validations.go`)

## 🚀 **SOLUCIÓN:**

El problema está en algún middleware de binding que está interfiriendo con el formato de fecha. Para solucionarlo:

### **Opción 1: Usar SQL directo (como se documentó)**
```sql
INSERT INTO lot_dates (lot_id, sowing_date, harvest_date, sequence, created_by, updated_by) VALUES 
(1, '2025-01-15 00:00:00', '2025-06-15 00:00:00', 1, 123, 123);
```

### **Opción 2: Revisar middlewares de binding**
El problema podría estar en:
- Middlewares de Gin que procesan el JSON antes de llegar a la validación
- Algún middleware de logging o parsing que modifica las fechas
- Configuración de timezone o formato de fecha en algún middleware

### **Opción 3: Habilitar validaciones comentadas**
En `validations.go` (líneas 375-395), las validaciones están comentadas. Habilitarlas podría ayudar a identificar el problema.


################################################
################################################