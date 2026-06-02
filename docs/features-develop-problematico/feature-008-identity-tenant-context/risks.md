# risks — feature-008 (BE)

## Funcionales

- **R-F1 — Invites/owner-invariant rompen en runtime sin migración 000224.** `Repository.CreateInvite/AcceptInvite` usan la tabla `tenant_invites`; `requireAnotherActiveOwner` filtra por rol `tenant_owner`. Ninguno existe en develop. `/me/context` NO se ve afectado.
  - *Mitigación:* coordinar que `000224_tenant_security_foundation` llegue por su feature dueña (001/003/009) antes de exponer invites; o gatear esos endpoints hasta entonces. Smoke `/me/context` por separado (no depende de 000224).

- **R-F2 — `/me/context` devuelve `permissions: []` si no hay seed de `auth_permissions`/`auth_role_permissions`.** La query funciona pero sin data los permisos quedan vacíos y el FE podría no mostrar/bloquear nada.
  - *Mitigación:* verificar seed de permisos en develop (mig 000180/000201 o feature 019/007). Test manual con un usuario que tenga rol con permisos.

- **R-F3 — Cambio de modelo de autorización (rol → permiso).** Antes `requireAdmin` exigía role=="admin"; ahora exige permisos `admin.tenants/users/memberships`. Usuarios "admin" sin esos permisos asignados pierden acceso.
  - *Mitigación:* confirmar que el rol admin/tenant_admin tenga esos permisos en seed antes de mergear.

## Técnicos

- **R-T1 — No compila sin `internal/shared/authz` (feature-007).** Dependencia dura.
  - *Mitigación:* prerrequisito verificado (`git ls-tree develop -- internal/shared/authz/`).

- **R-T2 — `ctxkeys` vs `contextkeys`.** `me_context.go`/`me_context_test.go` importan `.../security/go/contextkeys` y usan `ctxkeys.Actor`. NO es bug: el paquete declara `package ctxkeys`. (Verificado en `platform/security/go/contextkeys/contextkeys.go`.) No "arreglar" agregando alias.

- **R-T3 — Cambio de firma público del módulo admin.** `repo`/`newRepo`→`Repository`/`NewRepository`; `NewHandler(db, idp, ...)`→`NewHandler(uc, ...)`. Cualquier caller fuera de wire que use la firma vieja rompe.
  - *Mitigación:* `grep -rn "admin.NewHandler\|admin.newRepo\|admin\.NewRepository" --include=*.go` y traer enteros los 6 archivos para que el módulo quede consistente.

## Integración / cross-repo

- **R-I1 — FE mergeado antes que BE.** El navbar switcher / bootstrap llama `/api/v1/me/context` → 404.
  - *Mitigación:* BE-first estricto.

- **R-I2 — Mismatch de contrato `MeContext`.** Si el FE espera otro shape (campos `is_current`, `current_tenant_id`, `permissions[]`), romper el switcher.
  - *Mitigación:* fijar el JSON contract del `MeContext` (ver spec.md) y compartirlo con el agente FE. Snapshot del shape en el test `me_context_test.go`.

## Datos / migración

- **R-D1 — Sin 000224 no hay `tenant_invites`.** (= R-F1). No hay riesgo de pérdida de datos (sólo features inactivas).
- **R-D2 — `users.idp_email` debe existir** (lo usa `LocalUser`/queries). Verificado: presente en develop (mig 000180). Sin riesgo.

## Archivos compartidos

- **R-S1 — `wire/wire_gen.go` (generado) editado a mano o desincronizado con `admin_providers.go`.** Rompe el linkeo del binario completo.
  - *Mitigación:* regenerar con la tool wire; `git diff --check`; `go build ./...`.
- **R-S2 — `wire/admin_providers.go` con hunks ajenos.** Si el `git restore -p` arrastra cambios de otra feature.
  - *Mitigación:* aceptar sólo hunks de la cadena admin; revisar el diff resultante.

## Extracción parcial

- **R-X1 — Resabios del código viejo.** Si se trae handler.go nuevo pero queda `newRepo`/`requireAdmin`/`usernameToEmail` por mezcla → no compila o comportamiento inconsistente.
  - *Mitigación:* traer los 6 enteros; `grep -rn "newRepo\|requireAdmin\b\|usernameToEmail" internal/admin/` = 0.
- **R-X2 — Traer repository.go pero no usecases.go (o viceversa).** `UseCases` consume `RepositoryPort`; handler consume `UseCasesPort`. Parcializar rompe el build.
  - *Mitigación:* atomicidad: los 6 archivos juntos en el mismo PR.

## Riesgo de mergear solo este repo / solo el otro

- **Solo BE (este PR):** seguro. `/me/context` y el refactor admin quedan disponibles; sin FE no hay consumidor, pero no rompe nada existente. Los endpoints de invites quedan latentes (R-F1) hasta 000224.
- **Solo FE:** ROMPE. El FE feature-008 llama un endpoint inexistente → bootstrap/switcher fallan. No mergear FE sin BE.
