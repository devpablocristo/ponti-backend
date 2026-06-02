# notes-for-future-agent — feature-008 (BE)

## Resumen corto

FULL-STACK. En BE: endpoint `GET /api/v1/me/context` (bootstrap del FE: usuario + tenants + roles + permisos + tenant actual) y refactor del módulo `internal/admin` a capas (Handler delgado → `UseCases`/`UseCasesPort` → `Repository`/`RepositoryPort`), con autorización por permisos (`authz.HasPermission`) en vez de rol hardcodeado, y endpoints nuevos de invites/membership. SOURCE = `777e5f6a` (`develop-problematico~1`). Son 6 archivos del flist + parche de wire.

## Qué está en FE y qué en BE

- **BE (este paquete):** `/me/context`, refactor admin, invites/membership, DTOs `MeContext`. Tablas auth ya en develop salvo `tenant_invites`/`tenant_owner` (mig 000224, externa).
- **FE (otro repo, mismo feature-008):** TenantContext, navbar tenant switcher, general-entities-admin, login, BFF (`me.ts`, `authMiddleware`, `requestContext`). Consume `/me/context`.

## Archivos esenciales

- `internal/admin/me_context.go` — ruta + `GetMeContext` (lee `ctxkeys.Actor`, `ctxkeys.OrgID`).
- `internal/admin/usecases.go` — `GetMeContext` (la lógica real), DTOs `MeContext/MeUser/MeTenant`, tokens de invite.
- `internal/admin/repository.go` — queries `ListMembershipsForUser`, `ListPermissionsByRoleIDs`, invites, invariante de owner; DTOs `LocalUser/Tenant/UserMembership/TenantInvite/MeMembership/RolePermission`.
- `internal/admin/handler.go` — handler delgado + `requireAdminPermission` + registro de rutas.

## Archivos peligrosos / mezclados

- `wire/wire_gen.go` — GENERADO. No editarlo a mano; regenerar con la tool wire. Tocarlo mal rompe el build del binario completo.
- `wire/admin_providers.go` — compartido; parchear SÓLO la sección admin.
- `internal/admin/repository.go` y `handler.go` — mezclan varias intenciones (CRUD users/tenants + invites + membership + /me), pero TODO es feature-008. Traerlos enteros (no cherry-pick por hunks).

## Decisiones ya tomadas

- **Traer los 6 archivos enteros** (`git checkout 777e5f6a -- ...`): el delta develop→SOURCE es limpio y self-contained dentro del paquete admin.
- **Regenerar** `wire_gen.go` en vez de parchear.
- **No extraer** `internal/shared/authz` (es de feature-007) ni la migración `000224` (de tenancy/crudar-archive).
- `ctxkeys` no es bug: el paquete `platform/security/go/contextkeys` declara `package ctxkeys`.

## Dudas abiertas (para humano)

1. ¿Migración `000224_tenant_security_foundation` (tabla `tenant_invites` + rol `tenant_owner`) ya en develop por su feature dueña? Sin ella, invites/archive fallan en runtime; `/me/context` no la necesita.
2. ¿Seed de permisos `admin.*` y de `auth_role_permissions` presente en develop? Si no, `/me/context` devuelve `permissions: []` y la autorización admin bloquea.

## Qué comandos mirar primero

```sh
# prerrequisito authz (feature-007)
git -C /home/pablocristo/Proyectos/pablo/ponti/core ls-tree -r --name-only develop -- internal/shared/authz/
# estado del módulo admin en develop (debe ser el viejo: newRepo/requireAdmin/NewHandler(db,idp,...))
git -C /home/pablocristo/Proyectos/pablo/ponti/core show develop:internal/admin/handler.go | head -50
# delta real a aplicar
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff develop..777e5f6a --stat -- internal/admin/ wire/admin_providers.go wire/wire_gen.go
# tablas auth en develop
git -C /home/pablocristo/Proyectos/pablo/ponti/core grep -lE "tenant_invites|tenant_owner" develop -- 'migrations_v4/*.up.sql'   # esperado: vacío
```

## Errores a evitar

- Mergear FE antes que BE → `/me/context` 404.
- Portar handler.go/repository.go sin usecases.go → no compila (`UseCasesPort`/`RepositoryPort`).
- Editar `wire_gen.go` a mano y dejarlo desincronizado con `admin_providers.go`.
- "Arreglar" el import `contextkeys`/`ctxkeys` (no hace falta; es correcto).
- Intentar traer la migración 000224 dentro de este PR (no es de feature-008).

## Camino más seguro

1. Confirmar authz (feature-007) en develop.
2. Rama desde develop.
3. `git checkout 777e5f6a --` los 6 archivos del flist.
4. `git restore -p --source=777e5f6a -- wire/admin_providers.go` (sólo hunks admin).
5. Regenerar wire.
6. `go build ./... && go test ./internal/admin/...`.
7. PR BE; avisar al agente FE para mergear feature-008 FE después.

## Qué PR del otro repo va antes/después

- **Antes (mismo repo BE):** feature-007 actor-system (authz). Idealmente también la feature dueña de la migración 000224 (001/003/009) si se van a exponer invites.
- **Después (otro repo FE):** feature-008 FE (TenantContext + navbar switcher + BFF me.ts). Estricto BE-first.
