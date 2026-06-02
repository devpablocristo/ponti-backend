# feature-008 — Identity & tenant context (/me) — BE (ponti-backend / core)

- **ID:** feature-008
- **slug:** identity-tenant-context
- **Nombre:** Identity & tenant context (/me)
- **Tipo:** feature
- **Repo:** Backend Go — `ponti-backend` (path `/home/pablocristo/Proyectos/pablo/ponti/core`)
- **Existe en FE:** SÍ (FULL-STACK; mismo feature-008 en el repo FE)
- **Existe en BE:** SÍ (este paquete)
- **Merge:** BE-first, luego FE
- **SOURCE de extracción:** `develop-problematico~1` = SHA `777e5f6a` (Merge PR #120). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **Rango fuente-de-verdad (diff del orquestador):** `0972e565..777e5f6a`. Para el delta real a aplicar sobre destino, usar `003a9b8f..777e5f6a` (ver extraction-plan).
- **Rama destino:** `develop` (tip `003a9b8f`).

## Resumen

Endpoint de arranque del frontend `GET /api/v1/me/context` en el módulo `internal/admin`, más el refactor del módulo admin a una arquitectura por capas (Handler delgado → `UseCases` → `Repository` con port/interfaces). El handler pasa de control de acceso por rol (`requireAdmin` role=="admin") a control por permisos (`authz.HasPermission(..., authz.PermissionAdmin*)`). Se agregan endpoints de invitaciones y gestión de membership (crear invite, aceptar invite, cambiar rol, archivar), con invariante de "al menos un `tenant_owner` activo".

`/me/context` devuelve: el usuario local (mapeado por `idp_sub`), su lista de tenants con rol + permisos, y marca `is_current=true` el tenant indicado por `X-Tenant-ID` (resuelto vía `ctxkeys.OrgID`).

## Objetivo

Dar al FE un único endpoint de bootstrap que resuelva identidad + multitenancy: quién soy (usuario local detrás del JWT/idp_sub), en qué tenants tengo membership, con qué rol y qué permisos en cada uno, y cuál es el tenant activo. Habilita el `TenantContext` + navbar switcher del FE.

## Problema que resuelve

Antes el admin module mezclaba HTTP + DB en el handler (`newRepo(h.db)` dentro de cada handler), autorizaba sólo por rol "admin" hardcodeado, y no exponía un contrato de identidad/tenancy consolidado para el FE. El FE no tenía forma de saber sus tenants/roles/permisos en un solo request.

## Alcance en este repo (BE)

- **Nuevo endpoint:** `GET {APIBaseURL}/me/context` (no bajo `/admin`; se registra aparte con `registerMeContextRoute`, sólo middlewares de validación).
- **Refactor arquitectura admin:** introduce `UseCases` + `RepositoryPort` (interface) + `UseCasesPort` (interface consumida por el Handler). `Repository`/`NewRepository` exportados (antes `repo`/`newRepo` no exportados).
- **Autorización por permisos:** `requireAdmin` (role=="admin") → `requireAdminPermission(c, authz.PermissionAdmin*)`. Permisos usados: `admin.tenants`, `admin.users`, `admin.memberships`.
- **Nuevos endpoints admin:** `POST /admin/tenants/:tenant_id/invites`, `POST /admin/invites/accept`, `POST /admin/tenants/:tenant_id/memberships/:membership_id/role`, `POST /admin/tenants/:tenant_id/memberships/:membership_id/archive`.
- **Nuevos modelos/DTOs (en repository.go):** `LocalUser` (con `IDPEmail`), `Tenant`, `UserMembership`, `TenantInvite`, `MeMembership`, `RolePermission`. (en usecases.go): `MeContext`, `MeUser`, `MeTenant`, inputs/outputs de create/upsert/invite.
- **Lógica de tokens de invitación:** `newInviteToken` (32 bytes hex), `hashInviteToken` (sha256).
- **Aliases de rol legacy:** `RoleIDByName` soporta `viewer↔tenant_viewer`, `manager↔tenant_manager`, `admin↔tenant_admin`.
- **Wire:** `wire/admin_providers.go` cambia a `ProvideAdminRepository` + `ProvideAdminUseCases` + nuevo `ProvideAdminHandler(uc, ...)`; `wire/wire_gen.go` se regenera para esa cadena.
- **Tests:** `usecases_test.go` (fakes de repo+idp), `me_context_test.go` (sqlite in-memory + httptest).

## Alcance en el otro repo (FE)

Según la NOTA de la feature (no verificado en este paquete; coordinar con el agente FE): `TenantContext`, Navbar tenant switcher, `general-entities-admin`, pantalla de login, y BFF (`me.ts`, `authMiddleware`, `requestContext`). El FE consume `GET /api/v1/me/context` y manda `X-Tenant-ID` + `Authorization: Bearer`.

## Fuera de alcance (de ESTE paquete BE)

- Migraciones de las tablas auth (`auth_*`, `users.idp_email`, `tenant_invites`, rol `tenant_owner`): NO están en el flist de feature-008. Pertenecen a otras features (007/001/003/009). Ver dependencies.md.
- El paquete `internal/shared/authz`: dependencia de **feature-007 actor-system**, NO se extrae acá.
- La migración `core/* → platform/*`: ya está mergeada en `develop` (0 imports `core/` en `internal/` a tip `003a9b8f`); no se re-extrae.

## Comportamiento esperado

- `GET /me/context` sin actor en context → `401` (`authentication context required`).
- Con actor (idp_sub) pero sin usuario local → `403` (`local user not found`).
- Con usuario local: `200` con `{ user, current_tenant_id, tenants[] }`; cada tenant trae `role`, `permissions[]`, `is_current` (true si `tenant.id == ctxkeys.OrgID`).
- Endpoints admin: `403` (`insufficient permissions`) si falta el permiso correspondiente.
- Archivar/cambiar rol del último `tenant_owner` activo → `400` (`tenant must keep at least one active owner`).
- Aceptar invite inválido/expirado/ya aceptado → `400` (`invite is invalid or expired`).

## Estado en dp~1 (SHA 777e5f6a)

Código COMPLETO y coherente: compila contra `platform/*`, tiene tests de unidad (usecases) y de handler (me_context con sqlite). Wire regenerado. El uso de `ctxkeys.Actor`/`ctxkeys.OrgID` con import `platform/security/go/contextkeys` es CORRECTO: el paquete declara `package ctxkeys` (no es bug). Confianza alta (código leído íntegro).

## Criterios de aceptación

1. `go build ./...` y `go vet ./internal/admin/...` OK tras incluir `internal/shared/authz` (de feature-007).
2. `go test ./internal/admin/...` verde (4 tests usecases + 2 tests me_context).
3. `GET /api/v1/me/context` responde el shape `MeContext` documentado.
4. Wire compila: `ProvideAdminRepository`/`ProvideAdminUseCases`/`ProvideAdminHandler` resuelven; `wire_gen.go` usa la nueva cadena.
5. Los nuevos endpoints admin responden y respetan `authz` por permiso.

## Endpoints / Modelos / UI / DB / Tests afectados

- **Endpoints:** `GET /me/context`; `POST /admin/tenants/:tenant_id/invites`; `POST /admin/invites/accept`; `POST /admin/tenants/:tenant_id/memberships/:membership_id/role`; `POST /admin/tenants/:tenant_id/memberships/:membership_id/archive`. (existentes refactorizados: `GET/POST /admin/tenants`, `GET/POST /admin/users`, `POST /admin/memberships`).
- **Modelos/DTOs:** `MeContext`, `MeUser`, `MeTenant`, `LocalUser`, `Tenant`, `UserMembership`, `TenantInvite`, `MeMembership`, `RolePermission`, `CreateUserInput/Output`, `UpsertMembershipInput`, `CreateInviteInput/Output`.
- **UI:** N/A en BE (ver FE).
- **DB (consumida, no creada acá):** tablas `users` (col `idp_email`), `auth_tenants`, `auth_roles`, `auth_memberships`, `auth_permissions`, `auth_role_permissions` (existen en develop vía migraciones 000180/000201). `tenant_invites` y rol `tenant_owner` (NO existen en develop — migración 000224, otra feature).
- **Tests:** `internal/admin/usecases_test.go`, `internal/admin/me_context_test.go`.

## Dependencias

- **Intra-repo (fuerte):** feature-007 actor-system → aporta `internal/shared/authz` (importado por handler.go). Sin él, no compila.
- **Intra-repo (fuerte, runtime):** migración `000224_tenant_security_foundation` (tabla `tenant_invites`, rol `tenant_owner`) — NO en este flist; pertenece a la feature de tenant security (001/003) o crudar-archive (009). Sin ella, `/me/context` funciona, pero invites y el invariante de owner fallan en runtime.
- **Intra-repo (débil):** wire/* compartido (ver file-list).
- **Cross-repo:** FE feature-008 consume `/me/context`. BE-first.

## Riesgos

- **Funcional:** invites/owner-invariant fallan en runtime si no está la migración 000224.
- **Técnico:** wire_gen.go es archivo generado y compartido; un port parcial mal hecho rompe el build de todo el binario.
- **Cross-repo:** si FE mergea antes que BE, el switcher/`/me/context` 404ea.

## DECISIÓN recomendada

**Extraer tal cual, con prerrequisito.** Mergear feature-007 (authz) ANTES. El módulo admin se porta como archivos enteros (los 6 del flist) + parche parcial en `wire/admin_providers.go` y `wire/wire_gen.go`. Confirmar que la migración 000224 (tenant_invites + tenant_owner) llegue por su feature antes de usar invites en prod; `/me/context` no la necesita. No partir en subfeatures: el refactor del handler y `/me/context` están acoplados (mismo `UseCasesPort`).
