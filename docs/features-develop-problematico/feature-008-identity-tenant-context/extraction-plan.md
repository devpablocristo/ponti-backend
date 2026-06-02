# extraction-plan — feature-008 (BE)

- **Repo:** ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **Rama base:** `develop` (tip `003a9b8f`)
- **SOURCE:** `develop-problematico~1` = SHA `777e5f6a` (NUNCA `develop-problematico` tip; está vacío/restore)
- **Rama sugerida:** `pr/feature-008-identity-tenant-context-be`
- **Orden cross-repo:** **BE-first**, luego FE.

## PR title

`feat(be): identity & tenant context — GET /me/context + admin module refactor (port #120)`

## PR description (sugerida)

> Porta el endpoint de bootstrap del FE `GET /api/v1/me/context` y el refactor del módulo `internal/admin` a capas (Handler → UseCases → Repository con ports). Migra la autorización admin de rol hardcodeado a permisos (`authz.HasPermission`). Agrega endpoints de invitaciones y gestión de membership (invite, accept, role, archive) con invariante de "al menos un tenant_owner activo".
>
> **Depende de:** feature-007 (paquete `internal/shared/authz`) ya mergeado en develop.
> **Nota DB:** `/me/context` usa tablas ya presentes (auth_*, users.idp_email). Los endpoints de invites + invariante de owner requieren la migración `000224_tenant_security_foundation` (tabla `tenant_invites`, rol `tenant_owner`), que llega por su propia feature; sin ella esos endpoints fallan en runtime pero el binario compila y `/me/context` funciona.
>
> Cross-repo: el FE feature-008 (TenantContext + navbar switcher + BFF me.ts) debe mergear DESPUÉS de este PR.

## PREREQUISITO (bloqueante de compilación)

`internal/shared/authz` debe existir en `develop` (lo aporta feature-007 actor-system). Verificar antes de empezar:

```sh
git -C /home/pablocristo/Proyectos/pablo/ponti/core ls-tree -r --name-only develop -- internal/shared/authz/
# debe listar authz.go y authz_test.go; si NO -> mergear feature-007 PRIMERO
```

## Pasos ordenados

1. **Confirmar prerrequisito** (authz presente en develop — comando arriba).
2. **Crear rama** desde develop.
3. **Traer los 6 archivos del flist enteros** desde SOURCE (son self-contained dentro del paquete admin).
4. **Parchear wire/admin_providers.go** (sección admin: nueva firma + 2 providers nuevos).
5. **Regenerar wire_gen.go** con la tool wire (no editar a mano).
6. **Compilar + test del paquete admin + wire.**
7. **Validar** (ver validation.md) y abrir PR.

## Archivos enteros vs parciales

- **Enteros (`git checkout SOURCE -- <path>`):**
  - `internal/admin/me_context.go`
  - `internal/admin/usecases.go`
  - `internal/admin/me_context_test.go`
  - `internal/admin/usecases_test.go`
  - `internal/admin/handler.go`
  - `internal/admin/repository.go`
- **Parcial:** `wire/admin_providers.go` (sólo sección admin).
- **Regenerado:** `wire/wire_gen.go`.

## Migraciones / tests a incluir

- **Migraciones:** NINGUNA en este PR (no están en el flist). Documentar dependencia de `000224` (otra feature).
- **Tests:** los 2 archivos `_test.go` del flist se traen enteros.

## Dependencias previas

- feature-007 actor-system (authz) — **mergeada en develop antes de este PR**.

## Coordinación con el otro repo

- **BE-first.** Mergear este PR de BE; luego el agente FE mergea feature-008 FE (TenantContext / navbar switcher / BFF `me.ts`). El FE depende del shape `MeContext` y del path `/api/v1/me/context`.

## Comandos git SUGERIDOS (para un humano; el agente NO los ejecuta)

```sh
cd /home/pablocristo/Proyectos/pablo/ponti/core

# 0) prerrequisito
git ls-tree -r --name-only develop -- internal/shared/authz/

# 1) rama
git checkout develop
git pull --ff-only
git checkout -b pr/feature-008-identity-tenant-context-be

# 2) archivos enteros del flist desde SOURCE (777e5f6a == develop-problematico~1)
git checkout 777e5f6a -- \
  internal/admin/me_context.go \
  internal/admin/usecases.go \
  internal/admin/me_context_test.go \
  internal/admin/usecases_test.go \
  internal/admin/handler.go \
  internal/admin/repository.go

# 3) wire providers: sólo la sección admin (revisar el hunk antes de aceptar)
git restore -p --source=777e5f6a -- wire/admin_providers.go
#   aceptar SÓLO los hunks de: ProvideAdminRepository, ProvideAdminUseCases,
#   nueva firma de ProvideAdminHandler y AdminSet. Rechazar cambios ajenos.

# 4) regenerar wire (preferido sobre parchear wire_gen.go a mano)
go run github.com/google/wire/cmd/wire ./wire    # o: cd wire && go generate ./...
#   alternativa si no regenera: git restore -p --source=777e5f6a -- wire/wire_gen.go
#   (aceptar SOLO el hunk admin: adminRepository/adminUseCases/adminHandler)

# 5) sanity de whitespace/conflictos
git diff --check

# 6) build + tests
go build ./...
go test ./internal/admin/... ./internal/shared/authz/...
```

## Qué NO traer

- `internal/shared/authz/**` (feature-007).
- `migrations_v4/000224_*` (feature de tenant security 001/003/009).
- Cualquier hunk de `wire/admin_providers.go` o `wire/wire_gen.go` que no sea de la cadena admin.
- `internal/admin/idp/**` (idéntico entre develop y SOURCE; no tocar).

## Qué podría romperse

- **Compilación:** si authz no está en develop → `internal/admin/handler.go` no compila (undefined: `authz`). Mitigación: prerrequisito.
- **Build del binario:** si `wire_gen.go` queda con la firma vieja `ProvideAdminHandler(repository, adminClient, ...)` mientras `admin_providers.go` ya tiene la nueva → mismatch de tipos. Mantener ambos en sincronía (regenerar).
- **Runtime invites/owner:** sin tabla `tenant_invites` / rol `tenant_owner` (mig 000224) → 500 en `/admin/.../invites` y `/admin/.../archive`. `/me/context` NO se ve afectado.

## Cómo detectar extracción incompleta

- `go build ./...` falla con `undefined: authz.*` → falta feature-007.
- `wire_gen.go` referencia `ProvideAdminHandler(repository, adminClient, ...)` (firma vieja) → no se regeneró wire.
- `grep -rn "newRepo\|requireAdmin\b\|usernameToEmail" internal/admin/` debe dar 0 (resabios del código viejo no porteado).
- `grep -rn "devpablocristo/core/" internal/admin/` debe dar 0.

## Qué validar antes del PR

Ver `validation.md`. Mínimo: `go build ./...`, `go test ./internal/admin/...`, `go vet ./internal/admin/...`, wire regenerado y compilando.

## Qué hacer después de mergear

1. Avisar al agente/PR FE feature-008 que el endpoint `/api/v1/me/context` ya está en develop.
2. Confirmar que la migración `000224` (tenant_invites + tenant_owner) esté planificada/mergeada antes de exponer los endpoints de invites en prod.
3. Smoke manual de `/me/context` (ver validation.md).
