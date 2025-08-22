# 🌱 LOT API - Documentación para Frontend

## 📋 **Información General**

**Entidad**: Lot (Lotes de cultivo)  
**Base URL**: `http://localhost:8080/api/v1/lots`  
**Descripción**: Gestión completa de lotes agrícolas con fechas de siembra, cosecha, métricas y control de versiones.

---

## 🔐 **Headers Requeridos**

```http
X-API-KEY: abc123secreta
X-USER-ID: 123
```

---

## 📊 **Estructura de Datos**

### **Lot Object**
```typescript
interface Lot {
  id?: number;
  name: string;                    // Requerido, 2-255 chars
  field_id: number;                // Requerido, > 0
  hectares: string;                // Requerido, decimal > 0, max 10000
  previous_crop_id: number;        // Requerido, > 0
  current_crop_id: number;         // Requerido, > 0
  season: string;                  // Requerido, formato: YYYY o YYYY-YYYY
  variety?: string;                // Opcional
  dates?: LotDate[];               // Opcional, máximo 3 fechas
  status?: string;                 // Opcional
  version?: number;                // Requerido para PUT
  created_at?: string;             // ISO timestamp
  updated_at?: string;             // ISO timestamp
  created_by?: number;             // User ID
  updated_by?: number;             // User ID
}

interface LotDate {
  sowing_date?: string;            // Formato: YYYY-MM-DD
  harvest_date?: string;           // Formato: YYYY-MM-DD
  sequence: number;                // Requerido, > 0
}
```

### **Lot Table Object** (Para listados)
```typescript
interface LotTable {
  id: number;
  project_id: number;
  field_id: number;
  project_name: string;
  field_name: string;
  lot_name: string;
  previous_crop: string;
  previous_crop_id: number;
  current_crop: string;
  current_crop_id: number;
  variety: string;
  sowed_area: string;
  season: string;
  tons: string;
  dates: LotDate[];
  admin_cost: string;
  updated_at: string;
  harvested_area: string;
  cost_usd_per_ha: string;
  yield_tn_per_ha: string;
  income_net_per_ha: string;
  rent_per_ha: string;
  active_total_per_ha: string;
  operating_result_per_ha: string;
}
```

### **Lot Metrics Object**
```typescript
interface LotMetrics {
  seeded_area: string;
  harvested_area: string;
  yield_tn_per_ha: string;
  cost_per_hectare: string;
}
```

---

## 🚀 **ENDPOINTS**

### 1. **GET /api/v1/lots** - Listar Lotes

**Descripción**: Obtiene lista paginada de lotes con filtros y métricas agregadas.

**Query Parameters**:
```typescript
interface ListLotsParams {
  project_id?: number;     // Requerido si no hay field_id
  field_id?: number;       // Requerido si no hay project_id
  crop_id?: number;        // Opcional, filtrar por cultivo
  page?: number;           // Default: 1
  page_size?: number;      // Default: 10, Max: 100
}
```

**Ejemplo de Request**:
```javascript
const response = await fetch('/api/v1/lots?project_id=1&field_id=1&page=1&page_size=10', {
  headers: {
    'X-API-KEY': 'abc123secreta',
    'X-USER-ID': '123'
  }
});
```

**Response Success (200)**:
```json
{
  "page_info": {
    "per_page": 10,
    "page": 1,
    "max_page": 1,
    "total": 3
  },
  "totals": {
    "sum_sowed_area": "25",
    "sum_cost": "2800"
  },
  "items": [
    {
      "id": 1,
      "project_id": 1,
      "field_id": 1,
      "project_name": "Project 1",
      "field_name": "Field 1",
      "lot_name": "Lot 1",
      "previous_crop": "Crop 1",
      "previous_crop_id": 1,
      "current_crop": "Crop 1",
      "current_crop_id": 1,
      "variety": "V01",
      "sowed_area": "25",
      "season": "2025",
      "tons": "0",
      "dates": [
        {
          "sowing_date": "2025-07-22",
          "harvest_date": "2025-11-19",
          "sequence": 1
        }
      ],
      "admin_cost": "200",
      "updated_at": "2025-08-21T15:01:44.115658-03:00",
      "harvested_area": "0",
      "cost_usd_per_ha": "2800",
      "yield_tn_per_ha": "0",
      "income_net_per_ha": "0",
      "rent_per_ha": "0",
      "active_total_per_ha": "3000",
      "operating_result_per_ha": "-3000"
    }
  ]
}
```

**Response Error (400)**:
```json
{
  "type": "BAD_REQUEST",
  "code": 400,
  "message": "Invalid request payload",
  "details": "VALIDATION_ERROR: project_id or field_id is required"
}
```

---

### 2. **GET /api/v1/lots/metrics** - Obtener Métricas

**Descripción**: Obtiene métricas agregadas de lotes para análisis.

**Query Parameters**:
```typescript
interface MetricsParams {
  project_id?: number;     // Requerido si no hay field_id
  field_id?: number;       // Requerido si no hay project_id
  crop_id?: number;        // Opcional
}
```

**Ejemplo de Request**:
```javascript
const response = await fetch('/api/v1/lots/metrics?project_id=1&field_id=1', {
  headers: {
    'X-API-KEY': 'abc123secreta',
    'X-USER-ID': '123'
  }
});
```

**Response Success (200)**:
```json
{
  "seeded_area": "25",
  "harvested_area": "0",
  "yield_tn_per_ha": "0",
  "cost_per_hectare": "2800"
}
```

---

### 3. **GET /api/v1/lots/:id** - Obtener Lote Específico

**Descripción**: Obtiene información detallada de un lote específico.

**Path Parameters**:
```typescript
interface GetLotParams {
  id: number;              // ID del lote
}
```

**Ejemplo de Request**:
```javascript
const response = await fetch('/api/v1/lots/1', {
  headers: {
    'X-API-KEY': 'abc123secreta',
    'X-USER-ID': '123'
  }
});
```

**Response Success (200)**:
```json
{
  "id": 1,
  "name": "Lot 1",
  "field_id": 1,
  "hectares": "5",
  "previous_crop_id": 1,
  "current_crop_id": 1,
  "season": "2025",
  "variety": "V01",
  "dates": null,
  "status": "",
  "version": 1,
  "created_at": "2025-08-21T15:01:44.115658-03:00",
  "updated_at": "2025-08-21T15:01:44.115658-03:00",
  "created_by": 2,
  "updated_by": 2
}
```

**Response Error (404)**:
```json
{
  "type": "NOT_FOUND",
  "code": 404,
  "message": "lot 999 not found",
  "details": "record not found"
}
```

---

### 4. **POST /api/v1/lots** - Crear Lote

**Descripción**: Crea un nuevo lote con validaciones completas.

**Request Body**:
```typescript
interface CreateLotRequest {
  name: string;                    // Requerido, 2-255 chars
  field_id: number;                // Requerido, > 0
  hectares: string;                // Requerido, decimal > 0, max 10000
  previous_crop_id: number;        // Requerido, > 0
  current_crop_id: number;         // Requerido, > 0
  season: string;                  // Requerido, formato: YYYY o YYYY-YYYY
  variety?: string;                // Opcional
  dates?: LotDate[];               // Opcional, máximo 3 fechas
}
```

**Ejemplo de Request**:
```javascript
const response = await fetch('/api/v1/lots', {
  method: 'POST',
  headers: {
    'X-API-KEY': 'abc123secreta',
    'X-USER-ID': '123',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: "Test Lot 1",
    field_id: 1,
    hectares: "10.5",
    previous_crop_id: 1,
    current_crop_id: 1,
    season: "2025",
    variety: "Test Variety",
    dates: [
      {
        sowing_date: "2025-01-15",
        harvest_date: "2025-06-15",
        sequence: 1
      }
    ]
  })
});
```

**Response Success (201)**:
```json
{
  "message": "Lot created successfully",
  "id": 11
}
```

**Response Error (400) - Validación**:
```json
{
  "errors": [
    {
      "field": "name",
      "message": "lot: must have at least 2 characters",
      "value": "A"
    },
    {
      "field": "hectares",
      "message": "hectares: must be greater than 0",
      "value": "0"
    },
    {
      "field": "season",
      "message": "season: invalid season format. Use format: YYYY or YYYY-YYYY",
      "value": "invalid-season"
    }
  ]
}
```

---

### 5. **PUT /api/v1/lots/:id** - Actualizar Lote

**Descripción**: Actualiza un lote existente con control de versiones optimista.

**Path Parameters**:
```typescript
interface UpdateLotParams {
  id: number;              // ID del lote
}
```

**Request Body**:
```typescript
interface UpdateLotRequest {
  name?: string;                   // Opcional, 2-255 chars
  field_id?: number;               // Opcional, > 0
  hectares?: string;               // Opcional, decimal > 0, max 10000
  previous_crop_id?: number;       // Opcional, > 0
  current_crop_id?: number;        // Opcional, > 0
  season?: string;                 // Opcional, formato: YYYY o YYYY-YYYY
  variety?: string;                // Opcional
  dates?: LotDate[];               // Opcional
  version: number;                 // Requerido para control de concurrencia
}
```

**Ejemplo de Request**:
```javascript
const response = await fetch('/api/v1/lots/11', {
  method: 'PUT',
  headers: {
    'X-API-KEY': 'abc123secreta',
    'X-USER-ID': '123',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    name: "Updated Test Lot 1",
    field_id: 1,
    hectares: "12.0",
    previous_crop_id: 1,
    current_crop_id: 1,
    season: "2025",
    variety: "Updated Variety",
    version: 1
  })
});
```

**Response Success (204)**:
```
No Content
```

**Response Error (409) - Conflicto de Versión**:
```json
{
  "type": "CONFLICT",
  "code": 409,
  "message": "version conflict"
}
```

---

### 6. **PUT /api/v1/lots/:id/tons** - Actualizar Toneladas

**Descripción**: Actualiza las toneladas cosechadas de un lote específico.

**Path Parameters**:
```typescript
interface UpdateTonsParams {
  id: number;              // ID del lote
}
```

**Request Body**:
```typescript
interface UpdateTonsRequest {
  tons: string;            // Requerido, decimal >= 0, max 10000
}
```

**Ejemplo de Request**:
```javascript
const response = await fetch('/api/v1/lots/11/tons', {
  method: 'PUT',
  headers: {
    'X-API-KEY': 'abc123secreta',
    'X-USER-ID': '123',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    tons: "25.5"
  })
});
```

**Response Success (204)**:
```
No Content
```

**Response Error (400) - Valor Inválido**:
```json
{
  "error": "tons: must be greater than or equal to 0"
}
```

**Response Error (400) - Formato Inválido**:
```json
{
  "error": "Invalid tons format. Must be a valid decimal number."
}
```

---

### 7. **DELETE /api/v1/lots/:id** - Eliminar Lote

**Descripción**: Elimina un lote específico.

**Path Parameters**:
```typescript
interface DeleteLotParams {
  id: number;              // ID del lote
}
```

**Ejemplo de Request**:
```javascript
const response = await fetch('/api/v1/lots/13', {
  method: 'DELETE',
  headers: {
    'X-API-KEY': 'abc123secreta',
    'X-USER-ID': '123'
  }
});
```

**Response Success (204)**:
```
No Content
```

---

## 📋 **Reglas de Validación**

### **Nombre del Lote**
- **Requerido**: Sí
- **Longitud mínima**: 2 caracteres
- **Longitud máxima**: 255 caracteres
- **Caracteres permitidos**: Letras, números, espacios, guiones, apóstrofes, puntos, guiones bajos, paréntesis, corchetes, llaves, ampersands, acentos
- **Restricciones**: No espacios consecutivos

### **Hectáreas**
- **Requerido**: Sí
- **Valor mínimo**: > 0
- **Valor máximo**: 10,000
- **Formato**: Decimal

### **IDs de Cultivos**
- **Requerido**: Sí (previous_crop_id, current_crop_id)
- **Valor mínimo**: > 0

### **ID del Campo**
- **Requerido**: Sí
- **Valor mínimo**: > 0

### **Temporada**
- **Requerido**: Sí
- **Formato**: `YYYY` o `YYYY-YYYY`
- **Ejemplos válidos**: `2025`, `2025-2026`

### **Fechas**
- **Opcional**: Sí
- **Formato**: `YYYY-MM-DD`
- **Secuencia**: Debe ser > 0
- **Máximo**: 3 fechas por lote

### **Toneladas**
- **Valor mínimo**: >= 0
- **Valor máximo**: 10,000
- **Formato**: Decimal

---

## 🔄 **Flujos de Trabajo Recomendados**

### **Crear un Nuevo Lote**
```javascript
// 1. Validar datos en frontend
const validateLot = (data) => {
  const errors = {};
  
  if (!data.name || data.name.length < 2) {
    errors.name = 'El nombre debe tener al menos 2 caracteres';
  }
  
  if (!data.hectares || parseFloat(data.hectares) <= 0) {
    errors.hectares = 'Las hectáreas deben ser mayores a 0';
  }
  
  if (!data.season || !/^\d{4}(-\d{4})?$/.test(data.season)) {
    errors.season = 'Formato de temporada inválido (YYYY o YYYY-YYYY)';
  }
  
  return errors;
};

// 2. Enviar request
const createLot = async (lotData) => {
  try {
    const response = await fetch('/api/v1/lots', {
      method: 'POST',
      headers: {
        'X-API-KEY': 'abc123secreta',
        'X-USER-ID': '123',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(lotData)
    });
    
    if (response.ok) {
      const result = await response.json();
      return { success: true, id: result.id };
    } else {
      const error = await response.json();
      return { success: false, error };
    }
  } catch (error) {
    return { success: false, error: error.message };
  }
};
```

### **Actualizar un Lote**
```javascript
// 1. Obtener lote actual para obtener versión
const getLot = async (id) => {
  const response = await fetch(`/api/v1/lots/${id}`, {
    headers: {
      'X-API-KEY': 'abc123secreta',
      'X-USER-ID': '123'
    }
  });
  return await response.json();
};

// 2. Actualizar con control de versiones
const updateLot = async (id, updateData) => {
  try {
    // Obtener versión actual
    const currentLot = await getLot(id);
    
    const response = await fetch(`/api/v1/lots/${id}`, {
      method: 'PUT',
      headers: {
        'X-API-KEY': 'abc123secreta',
        'X-USER-ID': '123',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        ...updateData,
        version: currentLot.version
      })
    });
    
    if (response.status === 409) {
      // Conflicto de versión, reintentar
      return await updateLot(id, updateData);
    }
    
    return { success: response.ok };
  } catch (error) {
    return { success: false, error: error.message };
  }
};
```

### **Listar Lotes con Filtros**
```javascript
const useLots = () => {
  const [lots, setLots] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 10,
    total: 0
  });

  const fetchLots = async (filters = {}) => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        ...filters,
        page: pagination.page.toString(),
        page_size: pagination.pageSize.toString()
      });
      
      const response = await fetch(`/api/v1/lots?${params}`, {
        headers: {
          'X-API-KEY': 'abc123secreta',
          'X-USER-ID': '123'
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        setLots(data.items);
        setPagination(prev => ({
          ...prev,
          total: data.page_info.total,
          maxPage: data.page_info.max_page
        }));
        setError(null);
      } else {
        const errorData = await response.json();
        setError(errorData.message);
      }
    } catch (err) {
      setError('Error al cargar lotes');
    } finally {
      setLoading(false);
    }
  };

  return { lots, loading, error, pagination, fetchLots };
};
```

---

## 🎨 **Ejemplos de Componentes React**

### **Formulario de Creación de Lote**
```jsx
import React, { useState } from 'react';

const CreateLotForm = () => {
  const [formData, setFormData] = useState({
    name: '',
    field_id: '',
    hectares: '',
    previous_crop_id: '',
    current_crop_id: '',
    season: '',
    variety: '',
    dates: []
  });
  
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      const response = await fetch('/api/v1/lots', {
        method: 'POST',
        headers: {
          'X-API-KEY': 'abc123secreta',
          'X-USER-ID': '123',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
      });
      
      if (response.ok) {
        const result = await response.json();
        alert(`Lote creado exitosamente con ID: ${result.id}`);
        // Reset form or redirect
      } else {
        const errorData = await response.json();
        setErrors(errorData.errors || {});
      }
    } catch (error) {
      setErrors({ general: 'Error al crear el lote' });
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <div>
        <label>Nombre del Lote *</label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => setFormData({...formData, name: e.target.value})}
          required
          minLength={2}
          maxLength={255}
        />
        {errors.name && <span className="error">{errors.name}</span>}
      </div>
      
      <div>
        <label>Hectáreas *</label>
        <input
          type="number"
          step="0.01"
          min="0.01"
          max="10000"
          value={formData.hectares}
          onChange={(e) => setFormData({...formData, hectares: e.target.value})}
          required
        />
        {errors.hectares && <span className="error">{errors.hectares}</span>}
      </div>
      
      <div>
        <label>Temporada *</label>
        <input
          type="text"
          pattern="^\d{4}(-\d{4})?$"
          placeholder="2025 o 2025-2026"
          value={formData.season}
          onChange={(e) => setFormData({...formData, season: e.target.value})}
          required
        />
        {errors.season && <span className="error">{errors.season}</span>}
      </div>
      
      <button type="submit" disabled={loading}>
        {loading ? 'Creando...' : 'Crear Lote'}
      </button>
    </form>
  );
};
```

### **Tabla de Lotes con Paginación**
```jsx
import React, { useEffect, useState } from 'react';

const LotsTable = () => {
  const [lots, setLots] = useState([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 10,
    total: 0,
    maxPage: 1
  });
  const [filters, setFilters] = useState({
    project_id: '',
    field_id: '',
    crop_id: ''
  });

  const fetchLots = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        ...filters,
        page: pagination.page.toString(),
        page_size: pagination.pageSize.toString()
      });
      
      const response = await fetch(`/api/v1/lots?${params}`, {
        headers: {
          'X-API-KEY': 'abc123secreta',
          'X-USER-ID': '123'
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        setLots(data.items);
        setPagination(prev => ({
          ...prev,
          total: data.page_info.total,
          maxPage: data.page_info.max_page
        }));
      }
    } catch (error) {
      console.error('Error fetching lots:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLots();
  }, [pagination.page, pagination.pageSize, filters]);

  return (
    <div>
      <div className="filters">
        <input
          type="number"
          placeholder="Project ID"
          value={filters.project_id}
          onChange={(e) => setFilters({...filters, project_id: e.target.value})}
        />
        <input
          type="number"
          placeholder="Field ID"
          value={filters.field_id}
          onChange={(e) => setFilters({...filters, field_id: e.target.value})}
        />
      </div>
      
      {loading ? (
        <div>Loading...</div>
      ) : (
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Nombre</th>
              <th>Campo</th>
              <th>Hectáreas</th>
              <th>Temporada</th>
              <th>Toneladas</th>
              <th>Acciones</th>
            </tr>
          </thead>
          <tbody>
            {lots.map(lot => (
              <tr key={lot.id}>
                <td>{lot.id}</td>
                <td>{lot.lot_name}</td>
                <td>{lot.field_name}</td>
                <td>{lot.sowed_area}</td>
                <td>{lot.season}</td>
                <td>{lot.tons}</td>
                <td>
                  <button onClick={() => handleEdit(lot.id)}>Editar</button>
                  <button onClick={() => handleDelete(lot.id)}>Eliminar</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      
      <div className="pagination">
        <button
          disabled={pagination.page === 1}
          onClick={() => setPagination(prev => ({...prev, page: prev.page - 1}))}
        >
          Anterior
        </button>
        <span>Página {pagination.page} de {pagination.maxPage}</span>
        <button
          disabled={pagination.page === pagination.maxPage}
          onClick={() => setPagination(prev => ({...prev, page: prev.page + 1}))}
        >
          Siguiente
        </button>
      </div>
    </div>
  );
};
```

---

## ⚠️ **Manejo de Errores**

### **Códigos de Error Comunes**
- **400**: Datos inválidos o validación fallida
- **404**: Lote no encontrado
- **409**: Conflicto de versión (actualizar y reintentar)
- **500**: Error interno del servidor

### **Estructura de Errores de Validación**
```json
{
  "errors": [
    {
      "field": "nombre_del_campo",
      "message": "Descripción del error",
      "value": "Valor que causó el error"
    }
  ]
}
```

### **Función de Manejo de Errores**
```javascript
const handleApiError = (error, setErrors) => {
  if (error.errors && Array.isArray(error.errors)) {
    const fieldErrors = {};
    error.errors.forEach(err => {
      fieldErrors[err.field] = err.message;
    });
    setErrors(fieldErrors);
  } else if (error.message) {
    setErrors({ general: error.message });
  } else {
    setErrors({ general: 'Error desconocido' });
  }
};
```

---

## 🚀 **Mejores Prácticas**

### **1. Control de Versiones**
- Siempre obtener la versión actual antes de actualizar
- Implementar reintento automático en caso de conflicto
- Mostrar mensaje al usuario si hay conflictos

### **2. Validación Frontend**
- Validar datos antes de enviar al servidor
- Mostrar errores de validación en tiempo real
- Usar patrones de regex para formatos específicos

### **3. Manejo de Estados**
- Mostrar loading states durante requests
- Deshabilitar botones durante operaciones
- Proporcionar feedback visual de éxito/error

### **4. Paginación**
- Implementar paginación del lado del cliente
- Mantener filtros activos entre páginas
- Mostrar información de paginación

### **5. Optimización**
- Cachear datos de lotes frecuentemente accedidos
- Implementar debounce en filtros de búsqueda
- Usar React.memo para componentes de lista

Esta documentación proporciona toda la información necesaria para implementar la integración completa con la API de lotes en el frontend, incluyendo ejemplos prácticos, manejo de errores y mejores prácticas.
