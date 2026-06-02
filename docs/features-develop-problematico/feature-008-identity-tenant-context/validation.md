# validation — feature-008 (BE)

## Checklist pre-PR

- [ ] Prerrequisito authz: `git -C <repo> ls-tree -r --name-only develop -- internal/shared/authz/` lista `authz.go` (feature-007 mergeada).
- [ ] Los 6 archivos del flist traídos enteros desde `777e5f6a`.
- [ ] `wire/admin_providers.go` parcheado SÓLO en la sección admin (`ProvideAdminRepository`, `ProvideAdminUseCases`, nueva firma `ProvideAdminHandler(uc, ...)`, `AdminSet`).
- [ ] `wire/wire_gen.go` REGENERADO (no editado a mano); usa `adminRepository := ProvideAdminRepository(...)` → `adminUseCases := ProvideAdminUseCases(...)` → `adminHandler := ProvideAdminHandler(adminUseCases, ...)`.
- [ ] `grep -rn "newRepo\|requireAdmin\b\|usernameToEmail" internal/admin/` → 0 resultados.
- [ ] `grep -rn "devpablocristo/core/" internal/admin/` → 0 resultados.
- [ ] `git diff --check` limpio (sin whitespace/conflict markers).

## Comandos de build/test (BE)

```sh
cd /home/pablocristo/Proyectos/pablo/ponti/core
go build ./...
go vet ./internal/admin/... ./internal/shared/authz/...
go test ./internal/admin/...            # 4 tests usecases + 2 tests me_context
go test ./internal/shared/authz/...     # authz_test (de feature-007)
# regenerar/verificar wire:
go run github.com/google/wire/cmd/wire ./wire   # debe terminar sin error
```

Resultado esperado: build OK, `ok  .../internal/admin`, wire genera sin diffs sorpresa.

## Validación manual (smoke)

Con AUTH deshabilitado en local (NoopAdmin) o con un JWT válido + `X-Tenant-ID`:

```sh
# /me/context — requiere actor en context (lo inyecta el middleware desde el JWT)
curl -s -H "Authorization: Bearer <jwt>" -H "X-Tenant-ID: <tenant-uuid>" \
  http://localhost:8080/api/v1/me/context | jq

# esperado: { "user": {...}, "current_tenant_id": "<uuid>",
#             "tenants": [ { "id","name","role","permissions":[...],"is_current":true } ] }

# sin actor -> 401 ; usuario sin local user -> 403
```

Endpoints admin (requieren permiso correspondiente):

```sh
curl -s -X GET  .../api/v1/admin/tenants            # admin.tenants
curl -s -X POST .../api/v1/admin/tenants/<id>/invites -d '{"email":"x@y.com","role_name":"tenant_viewer"}'  # admin.memberships
# accept invite (sin permiso admin, requiere actor):
curl -s -X POST .../api/v1/admin/invites/accept -d '{"token":"<plain-token>"}'
```

> Nota: los endpoints de invites/archive sólo funcionan si la migración 000224 (tabla `tenant_invites`, rol `tenant_owner`) está aplicada en la DB local. Sin ella, devuelven 500. `/me/context` no la requiere.

## Casos borde a verificar

- `GET /me/context` sin `Authorization` → 401.
- actor presente pero `users` no tiene fila con ese `idp_sub` → 403.
- usuario con membership en N tenants → array `tenants` con N items; `is_current=true` sólo en el que matchea `X-Tenant-ID`.
- usuario sin permisos → `permissions: []` (no error).
- `RoleIDByName` con `"admin"` debe resolver vía alias `tenant_admin` y viceversa.
- archivar/cambiar rol del único `tenant_owner` activo → 400 "tenant must keep at least one active owner".
- aceptar invite expirado/ya aceptado/revocado → 400 "invite is invalid or expired".

## Qué revisar en UI/API/DB/env

- **API:** path exacto `{APIBaseURL}/me/context` (registrado fuera del grupo `/admin`, sólo con `GetValidation()` middlewares). Confirmar `APIBaseURL` (= `/api/v1`).
- **DB:** existencia de `users.idp_email`, `auth_tenants/roles/memberships/permissions/role_permissions` (presentes en develop) y, para invites, `tenant_invites` + rol `tenant_owner` (mig 000224, externa).
- **Seed:** que los roles tengan filas en `auth_role_permissions` con los permisos `admin.*` (si no, autorización admin falla y `/me/context` devuelve permisos vacíos).
- **env:** `cfg.Auth.Enabled` controla IDP real vs `NoopAdmin`.

## Qué validar en el otro repo (FE)

- El cliente FE llama `GET /api/v1/me/context` y consume `user`, `current_tenant_id`, `tenants[].{id,name,role,permissions,is_current}`.
- El FE manda `X-Tenant-ID` al cambiar de tenant en el navbar switcher.
- BFF `me.ts`/`authMiddleware`/`requestContext` propaga el `Authorization` y el tenant header.
- Mergear FE DESPUÉS del BE.

## Señales de incompletitud / incompatibilidad

- `undefined: authz` al compilar → falta feature-007.
- `wire_gen.go` con firma vieja `ProvideAdminHandler(repository, adminClient, ...)` → wire no regenerado.
- `/me/context` 404 → ruta no registrada (no se trajo `me_context.go` o `Routes()` viejo).
- `tenants[].permissions` siempre vacío → falta seed de permisos.
- 500 en invites/archive → falta migración 000224.
