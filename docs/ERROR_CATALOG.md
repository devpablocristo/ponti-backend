# Catálogo de errores BE

## TLDR

- Errores del dominio se crean con `domainerr.<Kind>("mensaje en inglés")` del paquete `platform/errors/go/domainerr`.
- El handler los mapea a HTTP via `sharedhandlers.RespondError(c, err)` → `platform/http/gin/go` traduce kind → status code.
- El FE traduce el mensaje inglés a español via `translateBackendError(message)` (ver `ui/src/lib/translateBackendError.ts`).
- **Regla**: mensajes en inglés, lower-case, en presente, breves. No enumerar IDs ni datos del usuario (no leak).

## Kinds disponibles y mapping HTTP

| Kind | HTTP status | Cuándo usarlo |
|---|---|---|
| `domainerr.Validation(msg)` | 400 Bad Request | Input inválido (payload, query, params). |
| `domainerr.Unauthorized(msg)` | 401 Unauthorized | Falta auth context / JWT inválido. |
| `domainerr.Forbidden(msg)` | 403 Forbidden | Auth OK pero sin permiso / tenant ajeno. |
| `domainerr.NotFound(msg)` | 404 Not Found | Recurso inexistente (devolver `[]` para listas vacías, no NotFound). |
| `domainerr.Conflict(msg)` | 409 Conflict | Violación de invariante de dominio (entity archivada, owner único, etc.). |
| `domainerr.Unavailable(msg)` | 503 Service Unavailable | Dependencia externa caída (IDP, review service, etc.). |
| `domainerr.Internal(msg)` | 500 Internal Server Error | Bug, fallo de DB, situación no manejada. **Loguear stack, mensaje genérico al usuario.** |
| `domainerr.New(kind, msg)` | depende del kind | Constructor genérico — preferir las helpers tipadas. |
| `domainerr.Newf(...)` | depende | Constructor con formatting; mismo criterio. |

## Patrones de mensaje canónicos

### CRUDAR / lifecycle (regla "archived = no existe")
Familia `domainerr.Conflict`:
- **Crear/editar con FK archivada**: `"<entity> is archived"` → FE: "El <entidad> está archivado/a. Restauralo o elegí otro."
- **Restore con parent archivado**: `"cannot restore <child> while <parent> is archived; restore the <parent> first"` → FE: "No se puede restaurar el <child> hasta que se restaure el <parent>."
- **Archive de entity con references**: `"<entity> has historical or active references; archive it instead"` → FE: "El <entity> tiene referencias activas. Removelas antes de archivar."
- **Doble archive / restore inválido**: `"<entity> already archived"` / `"<entity> is not archived"` / `"<entity> not found or outdated"`.
- **Hard delete sin archive previo**: `"<entity> must be archived before hard delete"`.

### Auth/tenancy
Familia `domainerr.Forbidden`:
- `"authentication context required"` → FE: "Sesión expirada, ingresá nuevamente."
- `"local user not found"` → FE: "Tu usuario no está sincronizado con la organización."
- `"insufficient permissions"` / `"permission required"` → FE: "No tenés permisos para esta acción."
- `"tenant header required"` / `"tenant context required"` / `"tenant membership required"` → FE: "Seleccioná una organización."
- `"project is not available for this tenant"` → FE: "Este proyecto no pertenece a tu organización."

Familia `domainerr.Unauthorized` (sólo cuando no hay auth contexto):
- `"authentication context required"`.

### Validación de input
Familia `domainerr.Validation`:
- `"invalid request payload"` (binding JSON falla).
- `"<field> is required"` o `"<field> required"` (campo faltante).
- `"invalid <field>"` (formato/parseo falló).
- `"invalid actor_kind"` / `"invalid actor role"` (enum no permitido).
- `"role not found"`, `"identity user not found"` (lookups que el caller debería haber resuelto antes).
- `"invite is invalid or expired"`.

### Negocio específico (no CRUDAR)
- `"concurrent update conflict - <entity> was modified by another user"` → Conflict.
- `"tenant must keep at least one active owner"` → Validation.
- `"<entity> already exists in this <scope>"` → Conflict.
- `"<entity> with same <field> already exists in this <scope>"` → Conflict.

### Internal (loguear pero no exponer detalles)
Familia `domainerr.Internal`:
- `"error <action>: <root cause>"` — patrón para errores no esperados (DB, network).
- El FE muestra mensaje genérico ("Error inesperado, reintentá") sin filtrar el detalle a usuario.

## Reglas de redacción

1. **Inglés**, lower-case, presente, ≤80 chars.
2. **No incluir IDs ni emails del usuario** en el mensaje (no leak en logs/HTTP).
3. **No incluir stack traces** ni rutas internas en errores user-facing.
4. Cuando hay una acción correctiva conocida, agregarla con punto y coma:
   - ❌ `"lot is archived"` (FE no sabe qué hacer)
   - ✅ `"cannot restore lot while field is archived; restore the field first"`
5. Para `Conflict` por archived, usar **literalmente** `"<entity> is archived"` (la FE tiene pattern explícito).

## Cómo agregar un error nuevo

1. Elegir el kind por significado HTTP (no por conveniencia — un 404 nunca debería ser Conflict).
2. Escribir el mensaje siguiendo las reglas arriba.
3. Si el FE necesita traducirlo, agregar el pattern en `ui/src/lib/translateBackendError.ts`.
4. Si el pattern es de una familia nueva, agregar entrada en este doc.

## Cómo lo consume el FE

`ui/src/lib/translateBackendError.ts` tiene ~65 patterns mapeados a español. La función matchea contra `error.response.data.message` (o equivalente) y devuelve un string ES para mostrar en notificaciones. Si no matchea ningún pattern, cae a un mensaje genérico.

Cuando agregues un nuevo `domainerr.X("nuevo pattern")`, considerá:
- Si el pattern es **user-facing** (Validation, Conflict, Forbidden): probablemente quiere traducción FE.
- Si es **Internal**: no hace falta — el FE muestra genérico para todos los 5xx.

## Referencias

- Helpers BE: `github.com/devpablocristo/platform/errors/go/domainerr` (lib compartida).
- Mapper HTTP centralizado: `internal/shared/handlers/errors.go` → `platform/http/gin/go`.
- Catálogo FE: [`ui/src/lib/translateBackendError.ts`](../../ponti-frontend/ui/src/lib/translateBackendError.ts).
- Lifecycle CRUDAR: [docs/crudar-lifecycle.md](./crudar-lifecycle.md), [docs/archive-restore-policy.md](./archive-restore-policy.md).
