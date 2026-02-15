# Contexto para Agente - Ponti Backend + Frontend

> Archivo de contexto para que otro agente continúe el trabajo. Actualizado al final de la sesión.

## TLDR

- **Proyecto**: Ponti - sistema agrícola (backend Go, frontend React, BFF Node/Express)
- **Stack local**: `make run-ponti-local` levanta backend, auth, frontend, AI
- **Error reciente resuelto**: 500 en `/api/supplies/20` - corregido en BFF y backend
- **Reglas críticas**: No hardcodear, no usar ROUND() en migraciones, cambios quirúrgicos

---

## Estructura del Proyecto

```
/home/pablo/Projects/Pablo/
├── ponti-backend/      # API Go (puerto 8080)
├── ponti-frontend/     # UI React + BFF Node (puertos 5173, 3000)
└── ponti-ai/           # Servicio AI (Python) - copilot + insights
```

### ponti-backend
- `cmd/api/` - Entrypoint, HTTP server
- `internal/` - Módulos por dominio (supply, lot, labor, etc.)
- `migrations_v4/` - SQL migrations (canónico)
- `pkg/` - Tipos compartidos, middlewares, utils
- `wire/` - Inyección de dependencias

### ponti-frontend
- `api/` - BFF (Express) - proxy entre UI y backend
- `ui/` - React + Vite - interfaz de usuario
- Vite proxy: `/api` → `localhost:3000/api` (BFF)

---

## Reglas del Proyecto (.cursorrules)

- **Idioma**: Comentarios en español, código en inglés, TODOs en inglés
- **Go**: Pointers para campos no-slice, `decimal.Decimal` para monetarios, `shareddomain.Base` al final de structs
- **Arquitectura**: Adapters usan DTOs, use cases con primitives/domain, repository con models
- **DB**: Priorizar SQL migrations, no modificar migraciones existentes, NUNCA ROUND() en migraciones
- **Tests**: Table-driven, gomock, ejecutar sin preguntar
- **Headers para curl**: `X-API-KEY: abc123secreta`, `X-User-Id: 123`

---

## Flujo de Requests

1. **Browser** → `GET /api/supplies/20` (Vite proxy a BFF)
2. **BFF** (ponti-frontend/api) → valida JWT, llama backend con headers
3. **Backend** (ponti-backend) → `GET /api/v1/supplies?project_id=20` con `X-API-KEY`, `X-User-Id`

El BFF inyecta `X-User-Id` desde `req.user.userID` (del token). El backend usa `RequireUserIDHeader` que pone el valor en `c.Request.Context()` para capas internas.

---

## Correcciones Aplicadas (Sesión Reciente)

### 1. Backend - CreateSuppliesBulk
**Archivo**: `internal/supply/handler.go`
**Problema**: Pasaba `c` (gin.Context) en lugar de `c.Request.Context()`, causando "user ID is not a string" al crear insumos en bulk.
**Fix**: `h.ucs.CreateSuppliesBulk(c.Request.Context(), supplies)`

### 2. BFF Supplies - Estructura de respuesta
**Archivo**: `ponti-frontend/api/src/routes/supplies.ts`
**Problema**: El backend devuelve `{ data: [...], page_info: {...} }` pero el BFF pasaba el objeto completo como `data`, y usaba `supplies.length` (objeto, no array).
**Fix**: Extraer `backendResp.data` y `backendResp.page_info`, devolver `{ success: true, data: { data: items, page_info } }`

### 3. BFF Supplies - URL correcta
**Problema**: GET "" llamaba `/supplies/${projectId}` que en backend es GetSupply (un insumo por ID), no ListSupplies.
**Fix**: Usar `/supplies?project_id=${projectId}` con URLSearchParams

### 4. BFF Supplies - X-User-Id como string
**Fix**: `"X-User-Id": String(userId)` en todos los handlers

### 5. BFF Lots (solución restaurada)
**Archivos**: `api/src/routes/lots.ts`, `api/src/clients/ApiClient.ts`, `ui/src/pages/admin/lots/Lots.tsx`
- ApiClient: extraer `message`/`details` del formato de error del backend
- Lots BFF: reenviar `customer_id`, `campaign_id`, usar `page_size` en lugar de `per_page`
- Lots.tsx: enviar `customer_id` y `campaign_id` en la query

---

## Archivos Clave

| Ruta | Descripción |
|------|-------------|
| `internal/supply/handler.go` | Handlers de supplies (ListSupplies, CreateSuppliesBulk, etc.) |
| `internal/supply/repository.go` | CreateSuppliesBulk usa `ConvertStringToID(ctx)` - requiere user ID en context |
| `internal/shared/models/base.go` | `ConvertStringToID` - error "user ID is not a string" si ctx no tiene string |
| `pkg/http/middlewares/gin/require_user_id_header.go` | Inyecta X-User-Id en `c.Request.Context()` |
| `ponti-frontend/api/src/routes/supplies.ts` | BFF supplies - proxy al backend |
| `ponti-frontend/api/src/routes/authMiddleware.ts` | Verifica JWT, setea req.user |
| `ponti-frontend/ui/src/hooks/useProducts/index.ts` | getSupplies(projectId) → GET /supplies/${projectId} |

---

## Cómo Ejecutar

```bash
# Stack completo (backend, frontend, AI)
cd ponti-backend && make run-ponti-local

# Solo backend
make run-api

# Probar supplies directamente
curl -s "http://localhost:8080/api/v1/supplies?project_id=20" \
  -H "X-API-KEY: abc123secreta" -H "X-User-Id: 123"
```

---

## Posibles Fuentes de Error 500

1. **User ID no en context**: Handlers que llaman use cases deben pasar `c.Request.Context()`, no `c`
2. **BFF estructura respuesta**: Backend devuelve `{ data, page_info }` - BFF debe mapear a `{ data: { data, page_info } }` para el frontend
3. **BFF URL incorrecta**: Backend tiene `GET /supplies` (list) y `GET /supplies/:supply_id` (uno) - no confundir
4. **Auth**: BFF requiere token válido (JWT de Identity Platform)

---

## Pendientes / Notas

- El error de AI insights (`/api/ai/insights/summa...` 500) puede ser independiente
- Supplies y Lots siguen patrones similares; al agregar customer_id/campaign_id en supplies, revisar lots como referencia
- Backend usa `page_size` para lots, `per_page` para supplies - verificar en cada módulo
