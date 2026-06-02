# implementation-status — feature-008 (BE)

## Estado global

**COMPLETA en el SOURCE (777e5f6a).** El módulo admin está refactorizado a capas, `/me/context` implementado, endpoints de invite/membership presentes, con tests de unidad y de handler. Compila contra `platform/*`. Confianza ALTA (código leído íntegramente: handler.go, repository.go, usecases.go y ambos tests).

- **% completitud (en SOURCE):** ~100% del código de la feature.
- **% completitud porteable hoy a develop:** ~95% (el 5% restante es dependencia externa: authz de feature-007 + migración 000224).

## Estado en este repo (BE)

- A tip `develop` (`003a9b8f`): el módulo admin está en su forma **previa** (handler con `newRepo(h.db)`, `requireAdmin` por rol, `NewHandler(db, idp, ...)`, sin `/me/context`, sin usecases, sin authz). Ya usa imports `platform/*` (la migración core→platform está mergeada).
- Lo que falta portar = exactamente los 6 archivos del flist + parche de wire.

## Estado en el otro repo (FE)

- Desconocido desde este paquete. Según NOTA de la feature, el FE feature-008 implementa TenantContext + navbar switcher + general-entities-admin + login + BFF (`me.ts`, `authMiddleware`, `requestContext`). Verificar con el agente FE. Cross-repo: BE-first.

## Tests

- `internal/admin/usecases_test.go` (306 líneas): fakes de `RepositoryPort` e `IDPClient`. Cubre `UsernameToEmail`, `hashInviteToken`, `newInviteToken`, `CreateUser` (happy + rechazo), `CreateInvite` (defaults), `GetMeContext` (requires actor + builds tenant list). Self-contained.
- `internal/admin/me_context_test.go` (175 líneas): sqlite in-memory con schema mínimo (`users`, `auth_tenants`, `auth_roles`, `auth_memberships`, `auth_permissions`, `auth_role_permissions`) + httptest. Cubre 200 con tenant/permisos + 401 sin actor. Usa `idp.NoopAdmin{}` y `NewRepository(db)`. Self-contained (no necesita la DB real ni la mig 000224).
- No hay tests para CreateInvite/AcceptInvite contra DB real ni para el invariante de owner (sólo el fake en usecases_test). Esa lógica de repository.go (transacciones, `requireAnotherActiveOwner`) queda sin cobertura automatizada → deuda.

## Pendientes / faltantes

### BLOQUEANTE para mergear (este repo)

- **`internal/shared/authz` debe estar en develop** (feature-007). Sin él no compila. Verificar antes del PR.
- **Regenerar `wire/wire_gen.go`** con la nueva cadena admin. Si queda desincronizado con `admin_providers.go`, el binario no linkea.

### Mejora futura

- Tests de integración para invites/owner-invariant contra DB real.
- Documentar/seed de permisos `admin.*` para que `authz.HasPermission` autorice (depende del seeding de auth, fuera del flist).

### Deuda aceptable

- `usecases.go` mezcla /me/context con CRUD de usuarios/tenants/invites en un solo archivo (cohesión del módulo admin; aceptable).
- Aliases de rol legacy (`viewer↔tenant_viewer`, etc.) en `RoleIDByName` — transitorio; remover cuando se consolide la nomenclatura de roles.

### Duda humana

- ¿La migración `000224_tenant_security_foundation` (tenant_invites + tenant_owner) llegará antes a develop por su feature dueña (001/003/009)? Si no, los endpoints de invites/archive estarán presentes pero romperán en runtime. Decidir si exponerlos ya o gatearlos.
- ¿El seed de `auth_permissions`/`auth_role_permissions` con permisos admin está en develop? Verificar para que `/me/context` no devuelva `permissions: []` siempre.
