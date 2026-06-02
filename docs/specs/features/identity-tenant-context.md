# spec — identity-tenant-context (feature-008)

> **Spec definitivo** re-baselineado contra `develop` real (tip `19b96dc4`). Fuente: `777e5f6a`
> (= `develop-problematico~1`, Merge PR #120). NO se implementa acá; es el contrato de qué traer y cómo.

- **id / slug:** feature-008 · `identity-tenant-context`
- **tipo:** feature FULL-STACK (este spec cubre el **BE**; el FE es el feature-008 del repo web)
- **fuente:** `777e5f6a` · **destino:** `develop` (`19b96dc4`)
- **orden en el cluster tenants/users:** **4º** (cierre: identidad + contexto de tenant para el FE)

---

## 1. Propósito

Exponer `GET /api/v1/me/context` (bootstrap del FE: quién soy + mis tenants/roles/permisos + tenant actual)
y refactorizar el módulo `internal/admin` a capas (Handler → UseCases → Repository con ports), con
autorización por permisos (`authz.HasPermission`) y endpoints de invites/membership.

## 2. Estado vs `develop` (diff real re-baselineado)

Verificado contra develop actual — la feature está al **0%**:
- `internal/admin/` existe en su forma **vieja**: `handler.go`, `repository.go`, `idp/*` (handler con
  `newRepo(h.db)`, `requireAdmin` por rol, sin `/me/context`).
- **Faltan** (a traer): `internal/admin/{me_context.go, usecases.go, me_context_test.go, usecases_test.go}`
  (`git ls-tree develop` → vacío para esos paths).
- `git grep "me/context\|registerMeContextRoute\|GetMeContext" develop` → **0**. El endpoint no existe.
- `handler.go`/`repository.go` **a reemplazar** por la versión refactorizada (diff develop→source limpio y
  self-contained dentro de `internal/admin`).

DB (verificado en develop):
- `users.idp_email`, `auth_tenants`, `auth_roles`, `auth_memberships`, `auth_permissions`,
  `auth_role_permissions` → **presentes** (migr 180/201). `/me/context` solo necesita estas.
- `tenant_invites` + rol `tenant_owner` → **ausentes** (los crea/siembra la migr **224 de feature-003**).

## 3. Alcance / archivos

Flist: `/tmp/flists/be-008.txt` (6 paths). Delta real = `develop..777e5f6a`.

### 3A. Propios — whole-file (`git checkout 777e5f6a -- <path>`)
```
internal/admin/me_context.go         (NUEVO: ruta registerMeContextRoute + GetMeContext)
internal/admin/usecases.go           (NUEVO: UseCases + DTOs MeContext/MeUser/MeTenant + invite tokens)
internal/admin/me_context_test.go    (NUEVO: sqlite in-memory + httptest)
internal/admin/usecases_test.go      (NUEVO: fakes repo+idp)
internal/admin/handler.go            (M→reemplazo: handler delgado + requireAdminPermission + rutas nuevas)
internal/admin/repository.go         (M→reemplazo: Repository/RepositoryPort + DTOs + queries membership/invite/permisos)
```
> `handler.go`/`repository.go` están como `M` pero el delta es limpio y todo pertenece a 008 → traerlos
> **enteros**, no cherry-pick por hunks.

### 3B. Cableado — partial-hunks / regenerar (FUERA del flist, requerido para compilar)
```
wire/admin_providers.go   (hunks: ProvideAdminRepository + ProvideAdminUseCases + nueva firma ProvideAdminHandler(uc,...) + AdminSet)  → git restore -p (solo sección admin)
wire/wire_gen.go          (cadena adminRepository→adminUseCases→adminHandler)  → REGENERAR (go run github.com/google/wire/cmd/wire ./wire)
```

### EXCLUIR
- `internal/shared/authz/**` → lo trae **001** (ver §5; el doc viejo lo atribuía a 007, es incorrecto).
- `migrations_v4/000224_*` → es **003**. No traer acá.
- `internal/admin/idp/**` → idéntico entre develop y source; no tocar.
- Cualquier hunk de wire ajeno a la cadena admin.

## 4. Migraciones

**Ninguna propia.** Consume tablas existentes (auth_*, `users.idp_email` de 180/201) y, para
invites/owner-invariant, `tenant_invites` + rol `tenant_owner` que vienen de la **224 (feature-003)**.
Sin la 224: el binario compila, `/me/context` funciona, pero `/admin/.../invites` y `/admin/.../archive`
fallan en runtime (500). No hay riesgo de numeración propio de 008.

## 5. Dependencias (re-baselineado — corrige el doc viejo)

- **Compile (DURO):** `internal/admin/handler.go` importa **solo** `internal/shared/authz` (verificado:
  `git show 777e5f6a:internal/admin/handler.go` → import de `internal/shared/authz`, **sin** `internal/actor`).
  Ese paquete lo introduce **feature-001** (está en su flist como whole-file). ⇒ **008 depende de 001, NO de
  007.** El doc de backlog decía "mergear 007 antes" — es incorrecto; el dep real es 001.
- **Runtime (DÉBIL):** migr **224 (003)** para invites + invariante de owner; seed de `auth_role_permissions`
  con permisos `admin.*` (lo siembra la propia 224) para que `/me/context` no devuelva `permissions: []`.
- **Cableado:** `wire/admin_providers.go` + regenerar `wire_gen.go`.
- **008 NO bloquea** otras features BE. **Cross-repo:** FE feature-008 (`TenantContext`, navbar switcher,
  BFF `me.ts`) consume `/me/context` → **BE-first**.

> Implicancia de orden: como 008 solo necesita 001 (compile) + 003 (runtime), **puede implementarse justo
> después de 001+003, sin esperar a 007**. Mantenerlo 4º es válido igual; pero no está gateado por actors.

## 6. Plan de implementación (pasos, sin ejecutar acá)

1. **Confirmar `internal/shared/authz` en develop** (lo aporta 001). Si no está → 001 primero.
2. Crear rama desde develop.
3. `git checkout 777e5f6a --` los 6 archivos del flist (enteros).
4. `git restore -p --source=777e5f6a -- wire/admin_providers.go` (solo hunks admin).
5. **Regenerar** `wire_gen.go` (`go run github.com/google/wire/cmd/wire ./wire`) — no editar a mano.
6. `go build ./... && go vet ./internal/admin/... && go test ./internal/admin/...`.
7. PR BE → avisar al FE feature-008 (BE-first). Confirmar que 224 (003) esté antes de exponer invites en prod.

## 7. Validación

```bash
go build ./...
go vet ./internal/admin/... ./internal/shared/authz/...
go test ./internal/admin/...        # 4 tests usecases + 2 tests me_context (sqlite in-memory)
go run github.com/google/wire/cmd/wire ./wire   # regenera sin error
```
- `grep -rn "newRepo\|requireAdmin\b\|usernameToEmail" internal/admin/` → 0 (resabios del código viejo).
- `grep -rn "devpablocristo/core/" internal/admin/` → 0.
- **Smoke `/me/context`:** con `Authorization: Bearer <jwt>` + `X-Tenant-ID: <uuid>` → `200`
  `{user, current_tenant_id, tenants[{id,name,role,permissions,is_current}]}`. Sin actor → `401`; actor sin
  usuario local → `403`.
- **Admin:** falta de permiso → `403`; archivar/cambiar rol del último `tenant_owner` activo → `400`
  ("tenant must keep at least one active owner"); invite inválido/expirado → `400`.

## 8. Riesgos y decisiones pendientes

- **Dep real = 001, no 007 (corrección):** el único bloqueante de compilación es `internal/shared/authz`
  (de 001). No esperar a 007 por 008.
- **`wire_gen.go` (ALTO):** generado y compartido; si queda desincronizado con `admin_providers.go`
  (firma vieja `ProvideAdminHandler(repository, adminClient, ...)`), el binario no linkea. **Regenerar**, no
  parchear a mano.
- **Runtime sin 224 (003) (MEDIO):** invites/owner-invariant → 500. Gatear esos endpoints o asegurar que la
  224 (renumerada, ver specs de 003/007) esté en develop antes de exponerlos.
- **Seed de permisos `admin.*` (MEDIO):** sin filas en `auth_role_permissions`, `/me/context` devuelve
  `permissions: []` y la autorización admin bloquea. Lo siembra la 224 (003) — verificar.
- **FE-first prohibido:** mergear el FE feature-008 antes del BE → `/me/context` 404. BE-first estricto.
- **Cobertura:** invites/owner-invariant contra DB real sin tests (solo fakes en `usecases_test`). Deuda.
- **Aliases de rol legacy** (`viewer↔tenant_viewer`, etc.) en `RoleIDByName`: transitorio; remover al
  consolidar nomenclatura.
