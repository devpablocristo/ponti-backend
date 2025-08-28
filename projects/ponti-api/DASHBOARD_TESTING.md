# 🎯 Dashboard API - Guía de Pruebas con curl

Esta guía te ayudará a probar el endpoint del Dashboard que ha sido refactorizado para soportar filtros de arrays y retornar datos reales de la base de datos.

## 📋 Información del Endpoint

- **URL Base**: `http://localhost:8080/api/v1/dashboard`
- **Método**: `GET`
- **Versión**: `v1`

## 🚀 Preparación

### 1. Iniciar el Servidor

Antes de ejecutar las pruebas, asegúrate de que el servidor esté corriendo:

```bash
# Opción 1: Usar Makefile
make run

# Opción 2: Ejecutar directamente
go run cmd/api/main.go

# Opción 3: Usar Docker (si está configurado)
docker-compose up api
```

### 2. Verificar que el Servidor esté Activo

```bash
curl http://localhost:8080/health
```

Deberías recibir una respuesta `200 OK`.

## 🧪 Scripts de Pruebas

### Script Completo (Automático)

```bash
./test_dashboard_curl.sh
```

Este script ejecuta automáticamente todas las pruebas y muestra un resumen completo.

### Script Interactivo (Manual)

```bash
./test_dashboard_simple.sh
```

Este script te permite seleccionar qué pruebas ejecutar desde un menú interactivo.

## 📝 Pruebas Individuales con curl

### 1. Endpoint Básico (Sin Filtros)

```bash
curl -s "http://localhost:8080/api/v1/dashboard" | jq '.'
```

**Respuesta Esperada**: JSON con estructura completa del dashboard

### 2. Filtro por Customer IDs

```bash
# Un solo customer ID
curl -s "http://localhost:8080/api/v1/dashboard?customer_ids=1" | jq '.'

# Múltiples customer IDs
curl -s "http://localhost:8080/api/v1/dashboard?customer_ids=1,2,3" | jq '.'
```

**Respuesta Esperada**: Dashboard filtrado por los customers especificados

### 3. Filtro por Project IDs

```bash
# Un solo project ID
curl -s "http://localhost:8080/api/v1/dashboard?project_ids=10" | jq '.'

# Múltiples project IDs
curl -s "http://localhost:8080/api/v1/dashboard?project_ids=10,20,30" | jq '.'
```

**Respuesta Esperada**: Dashboard filtrado por los proyectos especificados

### 4. Filtro por Campaign IDs

```bash
# Un solo campaign ID
curl -s "http://localhost:8080/api/v1/dashboard?campaign_ids=100" | jq '.'

# Múltiples campaign IDs
curl -s "http://localhost:8080/api/v1/dashboard?campaign_ids=100,200,300" | jq '.'
```

**Respuesta Esperada**: Dashboard filtrado por las campañas especificadas

### 5. Filtro por Field IDs

```bash
# Un solo field ID
curl -s "http://localhost:8080/api/v1/dashboard?field_ids=1000" | jq '.'

# Múltiples field IDs
curl -s "http://localhost:8080/api/v1/dashboard?field_ids=1000,2000,3000" | jq '.'
```

**Respuesta Esperada**: Dashboard filtrado por los campos especificados

### 6. Combinación de Filtros

```bash
# Combinar customer_ids y project_ids
curl -s "http://localhost:8080/api/v1/dashboard?customer_ids=1,2&project_ids=10,20" | jq '.'

# Combinar todos los filtros
curl -s "http://localhost:8080/api/v1/dashboard?customer_ids=1&project_ids=10&campaign_ids=100&field_ids=1000" | jq '.'
```

**Respuesta Esperada**: Dashboard filtrado por todos los criterios especificados

## 🔍 Casos de Error

### 1. IDs Inválidos

```bash
# ID negativo
curl -s "http://localhost:8080/api/v1/dashboard?customer_ids=-1"

# ID cero
curl -s "http://localhost:8080/api/v1/dashboard?project_ids=0"

# ID como string
curl -s "http://localhost:8080/api/v1/dashboard?customer_ids=abc"
```

**Respuesta Esperada**: `400 Bad Request` con mensaje de error apropiado

### 2. Formato Malformado

```bash
# Coma al final
curl -s "http://localhost:8080/api/v1/dashboard?customer_ids=1,"

# Coma al inicio
curl -s "http://localhost:8080/api/v1/dashboard?customer_ids=,1"
```

**Respuesta Esperada**: `400 Bad Request` o manejo gracioso del error

## 📊 Estructura de Respuesta Esperada

La respuesta del endpoint debe mantener la misma estructura JSON que antes, pero con datos reales:

```json
{
  "metrics": {
    "sowing": {
      "progress_pct": 75.5,
      "hectares": 80.0,
      "total_hectares": 100.0
    },
    "harvest": {
      "progress_pct": 60.0,
      "hectares": 60.0,
      "total_hectares": 100.0
    },
    "costs": {
      "progress_pct": 67.0,
      "executed": 10000.0,
      "budget": 15000.0
    },
    "operating_result": {
      "income_net": 25000.0,
      "total_costs": 15000.0
    }
  },
  "crop_incidence": {
    "crops": [...],
    "total": {
      "hectares": 100.0,
      "rotation_pct": 100.0,
      "cost_per_hectare": 100.0
    }
  },
  "management_balance": {
    "income_usd": 25000.0,
    "total_costs_usd": 15000.0,
    "operating_result_usd": 10000.0
  },
  "detailed_management_balance": {
    "balance": {
      "rows": [...]
    }
  }
}
```

## ✅ Criterios de Aceptación

### 1. Funcionalidad Básica
- [ ] Endpoint responde con status `200 OK`
- [ ] Respuesta es JSON válido
- [ ] Estructura JSON se mantiene igual que antes

### 2. Filtros de Arrays
- [ ] `customer_ids` filtra correctamente
- [ ] `project_ids` filtra correctamente
- [ ] `campaign_ids` filtra correctamente
- [ ] `field_ids` filtra correctamente
- [ ] Combinación de filtros funciona

### 3. Validación de Errores
- [ ] IDs negativos retornan `400 Bad Request`
- [ ] IDs cero retornan `400 Bad Request`
- [ ] IDs como strings retornan `400 Bad Request`
- [ ] Formato malformado se maneja graciosamente

### 4. Datos Reales
- [ ] `costs.executed > 0` cuando hay datos
- [ ] `sowing.hectares > 0` cuando hay datos
- [ ] `management_balance.operating_result_usd` calculado correctamente

## 🛠️ Herramientas Útiles

### jq (JSON Processor)

Para formatear las respuestas JSON de manera legible:

```bash
# Instalar en Ubuntu/Debian
sudo apt install jq

# Instalar en macOS
brew install jq

# Instalar en CentOS/RHEL
sudo yum install jq
```

### curl con Verbose

Para debugging detallado:

```bash
curl -v "http://localhost:8080/api/v1/dashboard?customer_ids=1"
```

### curl con Headers

Para simular diferentes tipos de cliente:

```bash
curl -H "Accept: application/json" \
     -H "User-Agent: Dashboard-Tester/1.0" \
     "http://localhost:8080/api/v1/dashboard"
```

## 🚨 Solución de Problemas

### Servidor No Responde

1. Verificar que el servidor esté corriendo:
   ```bash
   ps aux | grep "go run\|main.go"
   ```

2. Verificar el puerto:
   ```bash
   netstat -tlnp | grep :8080
   ```

3. Verificar logs del servidor

### Errores de Base de Datos

1. Verificar conexión a la base de datos
2. Verificar que las migraciones se hayan ejecutado
3. Verificar que la vista `dashboard_full_view` exista

### Errores de Validación

1. Verificar formato de los parámetros de query
2. Verificar que los IDs sean enteros positivos
3. Verificar que no haya espacios extra en los arrays

## 📚 Recursos Adicionales

- [Documentación de curl](https://curl.se/docs/)
- [jq Manual](https://stedolan.github.io/jq/manual/)
- [Go HTTP Testing](https://golang.org/pkg/net/http/httptest/)

## 🤝 Contribución

Si encuentras problemas o quieres agregar más pruebas:

1. Reporta el issue con detalles del error
2. Incluye la respuesta completa del servidor
3. Especifica el comando curl que causó el problema
4. Incluye información del entorno (OS, versión de Go, etc.)
