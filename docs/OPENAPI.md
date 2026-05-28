## TLDR

Pipeline contract-first BE → FE para evitar drift en tipos:

1. **BE** anota handlers con comentarios `@Summary @Tags @Param @Success @Router` (formato swaggo).
2. `make openapi` genera `docs/openapi/swagger.{yaml,json}` (Swagger 2.0).
3. **FE** corre `yarn codegen:openapi` que convierte a OpenAPI 3.0 y emite `src/api/generated/types.ts`.
4. Consumers FE importan de `@/api/generated` en lugar de redefinir types a mano.

Cuando cambia un handler anotado, basta correr ambos comandos para que los tipos FE queden alineados. CI puede correr ambos y fallar si hay diff sin commitear.

## Estado actual

Anotado: 2 handlers piloto.

- `GET /me/context` → `MeContext`, `MeUser`, `MeTenant` (bootstrap del FE).
- `POST /data-integrity/verify-costs/{projectId}` → `IntegrityReportResponse`, `IntegrityCheckDTO` (legacy anotation que ya existía).

Pendiente: anotar los ~48 handlers restantes en `internal/<module>/handler*.go`. Patrón canónico abajo.

## Setup local

### Backend — swag CLI

```bash
# instalar swag fuera del proyecto (no entra a go.mod)
go install github.com/swaggo/swag/cmd/swag@latest

# verificar
~/go/bin/swag --version  # debería ser >= 1.16

# generar spec
make openapi
```

Salida: `docs/openapi/swagger.yaml`, `docs/openapi/swagger.json`, `docs/openapi/docs.go`. Los 3 archivos se commitean — son el contrato versionado.

### Frontend — openapi-typescript

```bash
cd ../web/ui
yarn install  # ya incluye openapi-typescript@7 + swagger2openapi@7 como devDependencies

yarn codegen:openapi
```

Salida: `src/api/generated/types.ts`. Se commitea.

## Cómo anotar un endpoint

Agregar comentarios encima del método handler. Patrón mínimo:

```go
// CreateActor godoc
// @Summary      Crear un actor (cliente/inversor/proveedor/etc.)
// @Description  Crea un Actor maestro con sus roles. Valida invariantes via domain.Validate.
// @Tags         actor
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateActorRequest  true  "Datos del actor"
// @Success      201   {object}  dto.ActorResponse
// @Failure      400   {object}  map[string]string  "validation error"
// @Failure      403   {object}  map[string]string  "tenant required"
// @Failure      409   {object}  map[string]string  "conflict"
// @Security     BearerAuth
// @Router       /actors [post]
func (h *Handler) CreateActor(c *gin.Context) { ... }
```

Reglas:
- `@Tags` agrupa endpoints en la spec — un tag por módulo (`actor`, `lot`, `work-order`, etc.).
- `@Success` y `@Failure` referencian structs Go por nombre completo (incluyendo paquete). Para tipos `map[string]string` (error responses), el patrón anterior funciona.
- `@Router` es la **path desde `BasePath`** (no incluir `/api/v1`). El `@BasePath` global está en [cmd/api/main.go](../cmd/api/main.go).
- `@Security BearerAuth` indica que requiere JWT (la definición global está en main.go).
- Si el handler espera un header (`X-Tenant-ID`, etc.), agregar `@Param X-Tenant-ID header string true "tenant"`.

## Cómo el FE consume los tipos

`src/api/generated/index.ts` re-exporta tipos comunes con alias cortos:

```ts
import type { MeContext, MeTenant } from "@/api/generated";
```

Para tipos no aliased, importar el shape full:

```ts
import type { components } from "@/api/generated";
type CreateActorRequest = components["schemas"]["dto.CreateActorRequest"];
```

Para tipear request/response por endpoint:

```ts
import type { Paths } from "@/api/generated";
type MeContextResponse = Paths["/me/context"]["get"]["responses"]["200"]["content"]["application/json"];
```

## Migración progresiva

No hay big-bang. Cada módulo se migra cuando se toque:

1. Anotar el handler con comentarios swaggo.
2. Correr `make openapi && (cd ../web/ui && yarn codegen:openapi)`.
3. Reemplazar el tipo hand-written en FE por el generado.
4. PR.

Ejemplo de migración hecho en este sprint: `src/pages/login/context/TenantContext.shared.ts` cambió de `type Tenant = { ... }` hand-written a `type Tenant = Required<Pick<MeTenant, "id" | "name">> & Omit<MeTenant, ...>` usando el tipo generado.

## Decisiones técnicas

- **Por qué swag (Swagger 2.0) y no oapi-codegen (OpenAPI 3.0)**: swag es el estándar de facto para Gin, parsing simple, fewer dependencies en `go.mod`. La conversión 2.0→3.0 con `swagger2openapi` agrega un paso pero es transparente y rapidísimo.
- **Por qué openapi-typescript y no @types/openapi-fetch o RTK Query codegen**: openapi-typescript genera solo tipos (no runtime), respeta nuestro stack actual (axios) y no fuerza a adoptar React Query.
- **Required vs optional**: swag marca todos los fields como opcionales por defecto a menos que el struct Go tenga tag `binding:"required"`. El FE debe asumir opcional y validar runtime, o usar `Required<Pick<T, ...>>` para campos garantizados (como en `Tenant`).
- **Por qué se commitean los archivos generados**: docs/openapi/* son el contrato versionado entre BE y FE. Si solo viven en CI, un cambio de handler sin regenerar pasa el review. Si están commiteados, el diff es visible y el PR es la documentación.

## Verificación

```bash
# desde core/
make openapi
go build ./...

# desde web/ui/
yarn codegen:openapi
yarn tsc -b
```

CI debería correr ambos comandos y fallar si hay diff entre lo generado y lo committeado.
