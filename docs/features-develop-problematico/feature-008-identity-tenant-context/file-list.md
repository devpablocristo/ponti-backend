# file-list — feature-008 (BE)

Flist autoritativo: `/tmp/flists/be-008.txt` (6 entradas). Delta real a aplicar = `003a9b8f..777e5f6a`.

Leyenda extracción: `whole-file` = traer el archivo completo desde SOURCE; `partial-hunks` = sólo algunos hunks (archivo compartido); `manual-port` = aplicar a mano; `do-not-extract-yet` = no traer en este paquete.

## Propios (de feature-008, traer del SOURCE)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/admin/me_context.go` | A | handler (route + GetMeContext) | endpoint `/me/context` | whole-file | archivo nuevo, exclusivo de la feature | bajo | alta |
| `internal/admin/usecases.go` | A | capa UseCases + DTOs `MeContext/MeUser/MeTenant` + invite tokens | orquestación admin + /me | whole-file | archivo nuevo | bajo | alta |
| `internal/admin/me_context_test.go` | A | test (sqlite in-memory + httptest) | cobertura /me/context | whole-file | test nuevo, sólo depende de admin+idp.NoopAdmin | bajo | alta |
| `internal/admin/usecases_test.go` | A | test (fakes repo+idp) | cobertura usecases | whole-file | test nuevo, self-contained | bajo | alta |
| `internal/admin/handler.go` | M | handler delgado | refactor a UseCasesPort + authz + nuevos endpoints | whole-file | el delta develop→source es limpio (no core/ noise); reemplazo completo recomendado | medio | alta |
| `internal/admin/repository.go` | M | repository + modelos | DTOs + queries de membership/invite/permisos | whole-file | el delta develop→source es limpio | medio | alta |

> Nota: aunque `handler.go` y `repository.go` están como `M`, el diff develop-tip→SOURCE es coherente y self-contained dentro del paquete admin. Recomiendo traerlos enteros (`git checkout 777e5f6a -- internal/admin/handler.go internal/admin/repository.go`) en lugar de cherry-pick por hunks. Mezclan varias intenciones (ver abajo) pero todas pertenecen a feature-008.

## Compartidos (partial-hunks — NO en el flist, pero REQUERIDOS para compilar)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `wire/admin_providers.go` | M | wire (DI providers) | cambia firma de ProvideAdminHandler + agrega ProvideAdminRepository/UseCases | partial-hunks | archivo compartido del DI; sólo la sección admin cambia | medio | alta |
| `wire/wire_gen.go` | M | wire generado | cablea adminRepository→adminUseCases→adminHandler | manual-port (regenerar `go generate ./wire` o `wire ./wire`) | generado; mejor regenerar que parchear a mano; tocarlo mal rompe todo el build | alto | alta |

## Requeridos por dependencia (NO traer en este paquete; vienen por otra feature)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `internal/shared/authz/authz.go` | A | paquete authz | `HasPermission`, constantes `PermissionAdmin*` | do-not-extract-yet | pertenece a **feature-007 actor-system**; importado por handler.go | alto (bloqueante de compilación) | alta |
| `internal/shared/authz/authz_test.go` | A | test authz | — | do-not-extract-yet | idem feature-007 | bajo | alta |
| `migrations_v4/000224_tenant_security_foundation.{up,down}.sql` | A | migración | crea `tenant_invites`, rol `tenant_owner`, refuerza tenancy | do-not-extract-yet | pertenece a feature de tenant security (001/003) o crudar-archive (009); NO en flist 008 | alto (bloqueante runtime de invites/owner-invariant) | media |

## Dudosos

| path | status | tipo | nota | extracción | confianza |
|---|---|---|---|---|---|
| `internal/admin/idp/*` | (sin cambios) | idp client | interfaz `AdminClient` y `NoopAdmin` IDÉNTICOS entre develop y SOURCE (verificado) | do-not-extract-yet | alta |

## NO traer todavía

- Nada adicional propio de 008. La feature son exactamente los 6 archivos del flist + parche de wire. El resto (authz, migración 000224) llega por sus dueñas.

## Verificación de tablas DB (a tip `003a9b8f` develop)

| tabla / columna | ¿existe en develop? | origen |
|---|---|---|
| `users.idp_email` | SÍ | mig 000180 |
| `auth_tenants`, `auth_roles`, `auth_memberships` | SÍ | mig 000180/000201 |
| `auth_permissions`, `auth_role_permissions` | SÍ | mig 000180/000201 |
| `tenant_invites` | **NO** | mig 000224 (otra feature) |
| rol `tenant_owner` (en seed/data) | **NO** | mig 000224 (otra feature) |
