# spec — platform-tenancy-refactor (feature-001)

> **Spec definitivo** re-baselineado contra `develop` real (tip `19b96dc4`, no el `003a9b8f` de los
> docs viejos). Fuente del trabajo: `777e5f6a` (= `develop-problematico~1`). **NO** se implementa código
> acá; esto es el contrato de qué traer y cómo.

- **id / slug:** feature-001 · `platform-tenancy-refactor`
- **tipo:** refactor de infraestructura transversal (solo-BE, sin cara FE, sin cambio de contrato API)
- **fuente:** `777e5f6a` · **destino:** `develop` (`19b96dc4`)
- **orden en el cluster tenants/users:** **1º** (base de 003, 007, 008 y del resto del BE)

---

## 1. Propósito

Introducir un modelo de **tenancy centralizado** en el BE (paquete `internal/shared/authz` + adopción de
`tenancy.Scope`) y endurecer la capa HTTP (auth + observabilidad), sin cambiar el contrato API.

## 2. Estado vs `develop` (diff real re-baselineado)

El `spec.md` viejo lo describía como "swap `core/*`→`platform/*` + reemplazo de `MaybeTenantScope`". **Eso
está STALE** — verificado contra `develop` actual:

- `git grep -c "devpablocristo/core/" develop -- 'internal/**/*.go'` → **0**. `develop` ya está 100% en
  `platform/*`. **No hay swap de imports que hacer.**
- `git grep "MaybeTenantScope\|TenantWhere" develop` → **0**. No existe el helper viejo; no hay nada que
  "reemplazar". Hoy `develop` scopea tenant ad-hoc e inconsistente.
- `git ls-tree develop -- internal/shared/authz/` → **vacío**. El paquete `authz` **no existe** en develop.
- `git grep "tenancy.Scope" develop -- 'internal/**/repository.go'` → **0**. Ningún repo adopta `tenancy.Scope`.
- `internal/shared/filters/workspace.go` **ya existe** y tiene `ResolveProjectIDs`, pero **no** tiene
  `ValidateProjectAccess` ni scoping por tenant (`func ValidateProjectAccess` → ausente en develop).
- JWT legacy `internal/platform/http/middlewares/gin/require_jwt.go` y `internal/shared/utils/jwt_tools.go`
  **siguen presentes** en develop → 001 los borra.

**Conclusión:** sobre este develop, 001 = **INTRODUCIR** (no migrar) el modelo de tenancy. Falta todo:
1. Paquete nuevo `internal/shared/authz` (`authz.go` + `authz_test.go`) — `Principal`, `PrincipalFromContext`,
   `TenantFromContext`, `RequireTenant`, `OptionalTenantOrStrict`, `TenantStrictModeEnabled`,
   `HasPermission`/`RequirePermission`, constantes `Permission*` (14 entidades × 3 acciones + admin/api).
2. Adopción de `tenancy.Scope(ctx, db, alias)` en ~20 repos de dominio.
3. `workspace.go`: `tenancy.Scope` en `ResolveProjectIDs`, nueva `ValidateProjectAccess`, scope en
   `ValidateFieldBelongsToProject`, semántica strict-mode (`[]int64{0}` centinela cuando 0 matches en modo tenant).
4. Hardening de middlewares gin (issuer/audience, redacción de headers, `RejectUnsafeLocalAuthz`, nuevo
   `observability.go` con request-id/slog JSON/OTel/RED metrics).
5. Borrado de JWT casero (`require_jwt.go`, `jwt_tools.go`) → reemplazado por JWKS de `platform/authn`.
6. `slog` estructurado en `godotenv.go`/`gorm/repository.go`/`server.go`.

## 3. Alcance / archivos

Flist autoritativa: `/tmp/flists/be-001.txt` (57 paths) y `file-list.md` de la carpeta de backlog (se borra
al cerrar este spec; el contenido relevante queda acá). Dos naturalezas:

### 3A. `001-core` — whole-file, mergeable de entrada (~29 archivos, bajo riesgo)
Traer enteros desde `777e5f6a` (`git checkout 777e5f6a -- <path>`):

```
internal/shared/authz/authz.go                 (NUEVO)
internal/shared/authz/authz_test.go            (NUEVO)
internal/shared/filters/workspace.go           (M: scope + ValidateProjectAccess)
internal/shared/filters/workspace_test.go      (NUEVO)
internal/shared/db/tx_context.go
internal/shared/repository/errors.go
internal/shared/repository/validation.go
internal/shared/types/errors_compat.go
internal/shared/handlers/{auth,bind,errors,pagination,params_compat,query,responses,workspace_filters}.go
internal/platform/config/godotenv/godotenv.go              (slog)
internal/platform/persistence/gorm/repository.go           (slog + remueve Cloud SQL IAM no usado)
internal/platform/http/servers/gin/server.go               (slog)
internal/platform/http/middlewares/gin/observability.go          (NUEVO)
internal/platform/http/middlewares/gin/auth_hardening_test.go    (NUEVO)
internal/platform/http/middlewares/gin/{middleares,require_identity_platform_authz,
  require_credentials,local_dev_authz,error_handling,require_user_id_header,
  request_and_response_logger}.go
```
Borrados (whole-file): `internal/platform/http/middlewares/gin/require_jwt.go`,
`internal/shared/utils/jwt_tools.go`.

> Nota: el diff viejo asumía import-swap en estos archivos. Como develop YA está en `platform/*`, varios
> "M" se reducen a cambios reales menores (slog, hardening) — verificar `git diff develop 777e5f6a -- <path>`
> por archivo antes de copiar; donde el único cambio era el import, el archivo de develop ya está bien.

### 3B. `001-tail` — repos de dominio, partial-hunks (~18 archivos, riesgo alto)
Adopción de `tenancy.Scope`. Los `repository.go` **mezclan 001 con 002 (lifecycle), 007 (actor-sync), 009
(archive)**. Traer SOLO los hunks de `tenancy.Scope` (`git restore -p --source=777e5f6a -- <repo>`):

- **Bajo riesgo (separable simple):** class-type, commercialization, dollar, invoice, business-parameters,
  category, crop, lease-type, campaign, project, labor.
- **ALTO riesgo (scope interleaved con 002/007/009):** lot(28), supply(27), work-order(22),
  work-order-draft(16), stock(15), field(14), provider(13), manager(12), investor(11), customer(10). Si el
  hunk no separa limpio → **diferir a la feature dueña** (002/007/009) o re-aplicar `tenancy.Scope` a mano.
- **Especial:** `internal/shared/models/base.go` → solo el hunk del import `contextkeys` (el resto es 007/008).

### 3C. EXCLUIR de 001 (no son 001)
- `internal/{businessinsights,dashboard,report}/repository.go`: **0** `tenancy.Scope`; su diff es 027
  (json-tags / `fmt.Errorf`→`domainerr`) y 013. **No traer.**
- `internal/platform/files/excel/excelize/*` (3 borrados): son **feature-013 be-csv-export**.
- Cualquier hunk `lifecycle.*`, `actorsync.*`, `legacy_actor_map`, `EnsureCustomerFromActor`,
  `RunCascadeArchive`, `ArchiveUpdates`, `RestoreScopedRows` → 002/007/009.
- `internal/shared/utils/strings.go` (borrado): verificar callers (`godotenv.go` importa `pkgutils`) antes.

## 4. Migraciones

**Ninguna propia.** 001 es solo código. El scoping asume columnas `tenant_id` que las trae **003** (migr
224/225). Ver decisión transversal de renumeración en §8 (afecta a 003/007, no a 001).

## 5. Dependencias

- **Compile (DURO):** `go.mod` de develop **NO** tiene `platform/observability/go`,
  `platform/persistence/gorm/go` ni `prometheus/client_golang` (verificado: `git show develop:go.mod` → 0
  matches). Sin ellos, `001-core` **no compila**. Hay que agregarlos primero (PR-001a).
- **Runtime (DÉBIL→FUERTE):**
  - **003** aporta las columnas `tenant_id`. Sin 003, en modo transición el scope es no-op (no rompe build,
    no acota nada).
  - **008** setea el tenant en contexto (`contextkeys.OrgID`). Sin 008, `TenantFromContext` siempre nil →
    scoping inactivo permanente.
- **Orden funcional:** `001 (código) → 003 (columnas) → 008 (contexto)`. Cada capa es transition-safe; el
  **aislamiento real** requiere las tres.
- **001 desbloquea:** 002, 003, 007, 008, 009, 023, 025, 027 (todos asumen `authz` + `tenancy.Scope`).
- **Cross-repo (FE):** ninguno. Marcar "sin cambios FE".

## 6. Plan de implementación (pasos, sin ejecutar acá)

1. **PR-001a (go.mod):** agregar `platform/observability/go`, `platform/persistence/gorm/go`,
   `prometheus/client_golang` (+ `go mod tidy`). Validar `go build ./...`.
2. **PR-001b (core):** whole-file de §3A + los 2 borrados JWT. Validar build + tests de
   `authz`/`filters`/`middlewares gin`.
3. **PR-001c (tail low-churn):** `git restore -p --source=777e5f6a` de los repos low-churn, aceptando SOLO
   hunks de `tenancy.Scope`. Validar build + smoke transition-mode.
4. **Diferidos:** los ~9 repos high-churn viajan con su feature dueña (002/007/009), o como 001-tail manual.
5. **EXCLUIR siempre:** businessinsights/dashboard/report + excelize.

## 7. Validación

- `git grep "devpablocristo/core/" internal/` → 0 (ya lo está; no debe reaparecer).
- `git grep "MaybeTenantScope\|TenantWhere\|RequireJWT" internal/` → 0 código activo.
- `go build ./... && go vet ./...`. Cualquier `undefined: lifecycle.*`/`actorsync.*` ⇒ se arrastró código
  de 002/007 (partial-hunk mal cortado).
- `go test ./internal/shared/authz/... ./internal/shared/filters/... ./internal/platform/http/middlewares/gin/...`
  — esperados: `TestTenantOwnerIsNotGlobalWildcard`, `TestSaaSSuperadminIsGlobalWildcard`,
  `TestTenantFromContextAllowsTransitionMode`, `TestResolveProjectIDsScopesByTenant`,
  `TestValidateIdentityClaimsRejects{Issuer,Audience}Mismatch`, `TestRejectUnsafeLocalAuthzBlocksProductionEnv`.
- **Smoke (transition):** deploy 001 sin 003/008 → endpoints responden, scoping inactivo (devuelve todos los
  proyectos). NO debe fallar.
- **Aislamiento end-to-end:** diferido hasta 003 (columnas) + 008 (contexto).

## 8. Riesgos y decisiones pendientes

- **`go.mod` (bloqueante):** sin los 3 módulos platform, 001-core no compila. → PR-001a primero.
- **Extracción parcial mal cortada** (riesgo alto): un `repository.go` que quede importando
  `lifecycle`/`actor` sin esos símbolos no compila. Mitigación: `git restore -p` con split de hunks; si no
  separa limpio, diferir el repo a su feature dueña.
- **Cambio de semántica de `ResolveProjectIDs`** (modo tenant = todos los del tenant; `[]int64{0}` cuando 0
  matches): puede alterar listados si se mergea sin 003/008. Mitigación: coordinar 001→003→008; cubrir con
  `workspace_test.go`.
- **`RejectUnsafeLocalAuthz`:** rechaza tráfico en entornos no local-like sin `cfg.Auth.Enabled`. Verificar
  `Auth.Enabled`/`Auth.Environment` por entorno antes de mergear.
- **Decisión transversal — renumeración de migraciones (afecta 003/007, no 001):** develop está en migr
  **231** (saltó de 222→229; tiene 229/230/231 propias de lot/excel). El source trae 223–228 y 231–234 →
  **colisión en 231** y `golang-migrate` no aplica números < 231. **Las migr de 003/007 deben renumerarse a
  ≥232.** Gatea la validación end-to-end de todo el cluster. (Detalle en los specs de 003 y 007.)
- **Decisiones abiertas para humano:** ¿strict-mode default OFF en transición? ¿`dashboard`/`report` entran
  solo import o se posponen (recomendado posponer)? ¿borrar `strings.go` (verificar callers de
  `IsNumeric`/`NormalizeString`)?
