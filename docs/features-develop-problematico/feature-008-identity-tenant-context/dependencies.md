# dependencies — feature-008 (BE)

## Depende de (upstream)

### Fuertes (bloqueantes)

1. **feature-007 actor-system → `internal/shared/authz`** (intra-repo, compilación).
   - `internal/admin/handler.go` importa `internal/shared/authz` y usa `authz.HasPermission`, `authz.PermissionAdminTenants`, `authz.PermissionAdminUsers`, `authz.PermissionAdminMemberships`.
   - Estado: a tip `003a9b8f` (develop) el paquete `internal/shared/authz` **NO existe** todavía → si no se mergeó feature-007, esto no compila.
   - Acción: mergear feature-007 antes. Verificar: `git ls-tree -r --name-only develop -- internal/shared/authz/`.

2. **Migración `000224_tenant_security_foundation`** (intra-repo, runtime). NO en el flist de 008.
   - Crea la tabla `tenant_invites` y el rol `tenant_owner` (más refuerzos de tenancy).
   - Consumida por: `Repository.CreateInvite/AcceptInvite/findInviteBy*` (tabla `tenant_invites`) y por el invariante de owner (`requireAnotherActiveOwner` busca `r.name = 'tenant_owner'`).
   - Estado en develop: **ausente** (develop tiene 55 migraciones v4; SOURCE 65).
   - Impacto: el binario compila igual (no hay dependencia de compilación con tablas). `/me/context` funciona sin esta migración. Pero invites + role-change/archive del owner fallan en runtime.
   - Dueña probable: feature 001 (be-platform-tenancy-refactor) / 003 (be-multitenant-db-hardening) / 009 (crudar-archive-surface). Coordinar.

### Débiles

3. **Archivos wire compartidos** (`wire/admin_providers.go`, `wire/wire_gen.go`): no son "otra feature", pero son archivos compartidos del DI que hay que editar/regenerar como parte de este PR. Riesgo de colisión si otra feature toca wire en paralelo.

### Inciertas

4. Tablas `auth_permissions` / `auth_role_permissions` deben estar **pobladas** (seed) con los permisos `admin.tenants/users/memberships` y `customers.read`, etc., y los roles deben tener permisos asociados, para que `/me/context` devuelva `permissions[]` no vacíos y los endpoints admin autoricen. El seeding vive fuera del flist 008 (probable feature-007/001/019). Confianza media; revisar mig 000180/000201 y data de seed.

## Bloquea a (downstream)

- **FE feature-008** (cross-repo): TenantContext, navbar tenant switcher, BFF `me.ts`/`authMiddleware`/`requestContext`, login. Consumen `GET /api/v1/me/context` y mandan `X-Tenant-ID` + `Authorization`. **BE-first**: este PR debe mergear antes que el FE.
- Cualquier feature BE que dependa del `Repository`/`UseCases` exportados del módulo admin (firma nueva). No detectado en este rango fuera de la propia 008.

## Tipos / config / migraciones / APIs compartidos

| recurso | compartido con | nota |
|---|---|---|
| `internal/shared/authz` | feature-007 (dueña) + casi todos los repos de dominio | 28 archivos lo importan a SOURCE |
| `internal/shared/handlers` (ParseActor, ParseOrgID, RespondOK/NoContent) | todo el repo | ya presentes en develop; no se tocan |
| `platform/security/go/contextkeys` (package `ctxkeys`) | todo el repo | provee `Actor`, `OrgID`, `Role` |
| `wire/admin_providers.go`, `wire/wire_gen.go` | DI global | editar/regenerar |
| `migrations_v4/000224` | tenancy/crudar-archive | provee `tenant_invites` + `tenant_owner` |
| API `GET /me/context` | FE feature-008 | contrato cross-repo |

## Recomendación de orden

1. feature-007 (authz) — BE — **antes** (bloquea compilación de 008).
2. (idealmente) migración 000224 de su feature dueña — BE — antes de exponer invites en prod (no bloquea `/me/context`).
3. **feature-008 BE (este PR)** — BE.
4. feature-008 FE — FE — **después** del BE.
