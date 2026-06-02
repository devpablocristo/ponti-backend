# Feature-001 (be-platform-tenancy-refactor) — Análisis completo pre-implementación

> **Fecha:** 2026-06-01 · **Modo:** análisis, NO implementación. · **Solo lectura.**
> **Destino:** `develop` (`003a9b8f`). **Fuente:** `777e5f6a` (`develop-problematico~1`).
> Complementa (re-baseline) el `spec.md` de esta carpeta, que se fechó 2026-05-30 sobre un baseline desactualizado.

---

## 1. Qué ES realmente 001 (re-baseline)

El spec lo enuncia como “migración `core/*`→`platform/*` + reemplazo de `MaybeTenantScope`→`tenancy.Scope`”. **Eso está stale:** `develop` ya está 100% en `platform/*` (0 `core/*`) y **no tiene `MaybeTenantScope`** (no hay nada que reemplazar). Hoy `develop` scopea tenant **ad-hoc e inconsistente**.

**El 001 real sobre este `develop` =** *introducir* un modelo de tenancy centralizado y endurecer la capa HTTP:

1. **Nuevo paquete `internal/shared/authz`** (hoja, autocontenido): `Principal`, `PrincipalFromContext`, `TenantFromContext`, `RequireTenant`, `OptionalTenantOrStrict`, `TenantStrictModeEnabled`, `HasPermission`/`RequirePermission`, y constantes de permisos (14 entidades × 3 acciones + admin/api).
2. **Adoptar `tenancy.Scope(ctx, db, alias)`** (de `platform/persistence/gorm/go/tenancy`) en ~20 repos de dominio: envuelve el `db` al entrar a cada método de query y agrega `WHERE <alias>.tenant_id = ?` cuando hay tenant en contexto.
3. **`internal/shared/filters/workspace.go`** refactor: `ResolveProjectIDs` usa `tenancy.Scope` + nueva `ValidateProjectAccess` + semántica de strict-mode.
4. **Endurecer middlewares** (`internal/platform/http/middlewares/gin/`): validación issuer/audience, redacción de headers sensibles, `RejectUnsafeLocalAuthz` (bloquea auth local fuera de entornos local-like), y **nuevo middleware de observabilidad** (request-id, slog JSON, OTel, métricas RED).
5. **Borrar JWT legacy:** `require_jwt.go`, `internal/shared/utils/jwt_tools.go` (reemplazados por `platform/authn` JWKS).
6. **2 módulos nuevos en `go.mod`:** `platform/observability/go v0.2.1` y `platform/persistence/gorm/go v0.1.0` (+ indirectas prometheus/otel).

> **Sin cambio de contrato API** (refactor interno). No tiene cara FE.

---

## 2. El hallazgo clave: 001 se parte en CORE limpio + COLA entrelazada

El inventario real del diff `develop..777e5f6a` muestra **dos naturalezas muy distintas** dentro de 001:

### 2A. `001-core` — limpio, whole-file, mergeable de entrada (~29 archivos)
Archivos 001-puros (import-swap o nuevos), sin mezcla con otras features → **traer enteros**:
- `internal/shared/authz/{authz.go,authz_test.go}` (nuevo, 23 tests)
- `internal/shared/filters/{workspace.go,workspace_test.go}`
- `internal/shared/handlers/*` (auth, bind, errors, pagination, params_compat, query, responses, workspace_filters — swaps + logging)
- `internal/shared/{db/tx_context.go, repository/errors.go, repository/validation.go, types/errors_compat.go}` (swaps)
- `internal/platform/http/middlewares/gin/*` (8: `observability.go` NEW, `auth_hardening_test.go` NEW, + 6 modificados)
- `internal/platform/{config/godotenv/godotenv.go, persistence/gorm/repository.go, http/servers/gin/server.go}` (slog)
- **Borrados:** `middlewares/gin/require_jwt.go`, `internal/shared/utils/jwt_tools.go`

### 2B. `001-tail` — repos de dominio, partial-hunks (~18 archivos)
La adopción de `tenancy.Scope` por repo. Se subdivide por riesgo de mezcla:

| Riesgo | Repos (nº de `tenancy.Scope`) | Manejo |
|---|---|---|
| **Bajo** (casi solo import + scope) | class-type(1), commercialization(3), dollar(4), invoice(5), business-parameters(12), category(12), crop(12), lease-type(12), campaign(13), project(13), labor(17) | `git restore -p` simple; separable |
| **ALTO** (scope INTERLEAVED con 002/009 lifecycle y/o 007 actor-sync) | lot(28), supply(27), work-order(22), work-order-draft(16), stock(15), field(14), provider(13), manager(12), investor(11), customer(10) | `git restore -p` **obligatorio y difícil**; si los hunks no separan, **diferir a la feature dueña** (002/007/009) o re-aplicar `tenancy.Scope` a mano |
| **Especial** | `internal/shared/models/base.go` | SOLO el hunk de import `contextkeys`; el resto es 007/008 |

### 2C. EXCLUIR de 001 (no son 001 — el validador los marcó “delete” pero es *exclusión*)
`internal/{businessinsights,dashboard,report}/repository.go`: **0** `tenancy.Scope`; su diff es 027 (json-tags/`fmt.Errorf`→`domainerr`) y 013. **No traer en 001.**

---

## 3. Build / compilación

- **`001-core` compila solo sobre `develop`** SIEMPRE que `go.mod` tenga los 2 módulos. `authz` es paquete hoja (importa solo `platform/*` + stdlib, **cero** imports de lifecycle/actor). Verificado.
- **`go.mod` delta atribuible a 001:** `platform/observability/go v0.2.1`, `platform/persistence/gorm/go v0.1.0`, `prometheus/client_golang v1.23.2` (+ indirectas otel). **Ninguno está en `develop`** → **bloqueante de compilación**. #124 NO los agregó.
- **Riesgo en `001-tail`:** si se trae **entero** un repo high-churn (p.ej. `lot/repository.go`), arrastra `lifecycle.RootCause/ArchiveUpdates/RequireActive` (002/009) y `actorsync.*` (007) → **no compila** (esos paquetes no están en develop). Por eso es partial-hunks o se difiere.

---

## 4. Dependencias (compile vs runtime — distinción crítica)

- **Compile:** los 2 módulos `platform/*` en `go.mod` (duro).
- **Runtime (003):** `tenancy.Scope` agrega `WHERE tenant_id = ?`. Si las columnas `tenant_id` no existen (las agrega **003**, migr 224/225) **la query falla en runtime** (no en compile). **PERO** en *modo transición* (sin tenant en contexto) el scope es no-op → se preserva el comportamiento legacy. ⇒ **001 solo es seguro de deployar** (no crashea, no corrompe), pero **la multitenancy no se enforce** hasta que entren 003 (columnas) + 008 (tenant en contexto).
- **Runtime (008):** `TenantFromContext` lee `contextkeys.OrgID`; sin el middleware de identidad de 008, el tenant es siempre nil → scoping inactivo.
- **Orden funcional:** `001 (código) → 003 (columnas) → 008 (contexto)`. Cada capa es independiente y transition-safe; la **aislación** requiere las tres.
- **001 desbloquea:** 003, 007, 023, 025, 027 + todos los dominios (heredan authz/tenancy).
- **Migración:** 001 no tiene migr propias, pero su validación end-to-end depende de 224/225 (003), que chocan con la numeración (develop en 230). → la **estrategia de migración** (riesgo #1 de la auditoría) gatea la validación completa de 001.

---

## 5. Riesgos y decisiones que necesito de vos (antes de implementar)

1. **`go.mod`:** ¿agrego los 2 módulos (+indirectas) dentro del PR de 001, o en un PR previo? (No están en develop; sin ellos no compila.)
2. **Split 001-core vs 001-tail (LA decisión grande):**
   - **Opción A (recomendada):** mergear **001-core + repos low-churn** ahora (compila, transition-safe), y **diferir los ~9 repos high-churn** para que su `tenancy.Scope` viaje CON su feature dueña (002/007/009). 001 queda como “base de tenancy” sin tocar lo entrelazado.
   - **Opción B:** intentar 001 completo (los 20 repos) con `restore -p` repo por repo. Más fiel al spec, pero los 9 high-churn requieren separar hunks interleaved (difícil) o re-aplicar a mano → más riesgo y tiempo.
3. **Orden de merge:** confirmar `001 → 003 → 008`. 001 solo es seguro pero inerte hasta 003+008.
4. **Strict mode:** `TenantStrictModeEnabled()` — ¿default OFF durante la transición (recomendado) para no romper queries sin tenant?
5. **Estrategia de migración** (necesaria para validar 001 end-to-end): Camino A (reset/reordenar 223–228 antes de 229/230) vs B (renumerar). → mejor resolver antes de cerrar 001.

---

## 6. Plan de extracción propuesto (NO ejecutado — para tu aprobación)

> Asume Opción A del punto 5.2. Todo `git restore`/`checkout` parte de `777e5f6a`, nunca del tip.

1. **PR-001a (go.mod):** agregar `platform/observability/go`, `platform/persistence/gorm/go`, `prometheus/client_golang` (+ `go mod tidy`). Validar `go build ./...`.
2. **PR-001b (infra core):** whole-file de los ~29 de §2A (authz, workspace, handlers, middlewares, platform infra) + los 2 borrados JWT. Validar build + `go test ./internal/shared/authz ./internal/shared/filters ./internal/platform/http/middlewares/gin`.
3. **PR-001c (repos low-churn):** `git restore -p 777e5f6a -- <repo>` para los ~11 low-churn, aceptando SOLO hunks de import + `tenancy.Scope`. Validar build + smoke transition-mode.
4. **Diferidos:** los ~9 high-churn quedan documentados para entrar con 002/007/009 (o como 001-tail manual posterior).
5. **EXCLUIR siempre:** businessinsights/dashboard/report (son 027/013).

---

## 7. Plan de validación

- **Pre:** `go.mod` tiene los 2 módulos (si no, no compila).
- **Build:** `go build ./... && go vet ./...`. Cualquier `undefined: lifecycle.*` o `actorsync.*` ⇒ se arrastró código de 002/007 (partial-hunk falló).
- **Grep:** `git grep 'devpablocristo/core/' internal/` = 0 en paths de 001.
- **Tests:** `authz` (`TestTenantOwnerIsNotGlobalWildcard`, `TestTenantFromContextAllowsTransitionMode`), `filters` (`TestResolveProjectIDsScopesByTenant`, `ValidateProjectAccess`), middleware hardening (`TestValidateIdentityClaimsRejectsIssuerMismatch`, `TestRejectUnsafeLocalAuthzBlocksProductionEnv`).
- **Smoke (transition):** deploy 001 sin 003/008 → endpoints responden, scoping inactivo (devuelve todos los proyectos), confirma transition-mode. NO debe fallar.
- **Aislación end-to-end:** requiere 003 (columnas) + 008 (contexto) → diferida hasta esos PRs.

---

*Análisis solo lectura; cero cambios a código o specs. Único artefacto nuevo: este archivo.*
