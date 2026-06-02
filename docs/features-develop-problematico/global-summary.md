# Global summary — descomposición de `develop-problematico` (Backend Go / ponti-backend)

- Repo: `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- Rango fuente: `0972e565..777e5f6a`
- SOURCE real de extracción: `develop-problematico~1` = **`777e5f6a`** (el pico). El tip de `develop-problematico` (`66f2b602` "restore: app a estado pre-new-cns3 + mantener tooling local actual") es un *restore* que vacía la rama; **nunca usar el tip como fuente**.
- Destino: `develop`.
- Fecha del análisis: 2026-05-30.

---

## 1. Qué cambió globalmente (magnitud)

Diff `0972e565..777e5f6a`:

- **427 archivos** cambiados — **+40.589 / -14.664** líneas.
- Desglose: **229 modificados, 169 añadidos, 29 borrados**.
- **101 archivos `_test.go`** tocados (añadidos o modificados): es una rama con mucha cobertura nueva.
- Migraciones SQL nuevas: **18 archivos** (9 pares up/down) en `migrations_v4/`, numeración **223–234**.

`develop-problematico` fue una rama de **integración** que apiló varias líneas de trabajo grandes: refactor de tenancy, framework de lifecycle/archive (CRUDAR), sistema de actores, identidad/me_context, projects, integración AI Companion, export CSV, modularización de config, y la migración en curso a los módulos `platform/*` (new-cns3). El BE aquí es la **base de casi todo**: el FE de los paquetes full-stack vive en el otro repo.

---

## 2. Grandes áreas tocadas (por volumen de archivos)

| Área | Archivos | Feature(s) principal(es) |
|---|---|---|
| `internal/shared` | 31 | 001 (tenancy), 002/009 (lifecycle+archive), 004 (text), 013 (csvexport), 023 |
| `internal/supply` | 29 | 001, 009, 013 + (residuo tenancy, ver §6) |
| `internal/work-order` | 16 | 001, 009, 013 |
| `internal/platform` | 16 | 021/new-cns3 (borrado de excelize), 005 |
| `internal/lot` | 16 | 001, 009, 013 |
| `internal/actor` | 14 | **007 (la feature grande)** |
| `internal/stock` / `labor` / `customer` / `report` | 13/13/11/10 | 001, 009, 013 |
| `internal/project` | 9 | **010** (scope/creator, rename, archived-refs) |
| `internal/axis` | 8 (todos nuevos) | **012** (Companion) + Nexus (ver §6, experimental) |
| `internal/data-integrity` | 8 | 018 (excluyendo tentative-prices, ya DONE #121) |
| `cmd/config` | 8 | **005** (modularización) |
| `wire/` | 16 | **023** (DI) |
| `.github/workflows` | 6 (M) | 020 |
| `scripts/` | ~18 | 019 (tooling local) |
| `docs/` + raíz | ~25 | 024 |

---

## 3. Cambios transversales (cortan todo el BE)

### feature-001 — tenancy refactor (`tenancy.Scope`)
Reemplazo de `MaybeTenantScope` por `tenancy.Scope` en ~23 repos. **Es la base de todo el BE**: prácticamente cada `internal/<dominio>/repository.go` quedó tocado. Sin cambio de contrato API (interno). Tests dedicados: **26 `repository_tenant_test.go`** nuevos (uno por dominio) + `usecases_tenant_test.go` en data-integrity.

### feature-002 + feature-009 — CRUDAR lifecycle + archive surface
- `internal/shared/lifecycle/` nuevo y completo: `lifecycle.go`, `policy.go`, `cascade.go`, `archive_cleanup.go`, `metrics.go` + tests (incl. `invariant_e2e_test.go`).
- **Cambio de CONTRATO API** en ~20 dominios: `DELETE /:id` → `POST /:id/archive` + `DELETE /:id/hard` + `GET /archived`. ~19 `handler.go` referencian archive. Su contraparte FE vive en feature-014 (pages) y feature-006 (ArchivedListPage) en el otro repo.
- Migraciones que lo fundan: 227, 228, 232, 233.

### feature-013 — export CSV (reemplaza Excel)
- Nuevo `internal/shared/csvexport/writer.go` + `csv-service.go` por dominio (labor, lot, stock, supply, work-order).
- **Borra todo el stack Excel**: `internal/{labor,lot,stock,supply,work-order}/excel*` + `internal/platform/files/excel/excelize/*` (bootstrap/config/service, ~616 líneas). Los endpoints export pasan de XLSX a CSV → **revisar consumo en el FE** antes de portar.

### Migración a `platform/*` (new-cns3) — transversal, NO está en la tabla como feature aislada
`go.mod` añade múltiples módulos `github.com/devpablocristo/platform/*`: `authn`, `databases/postgres`, `errors`, `http/gin`, `http`, `kernels/governance`, `notifications`, `security`, `validate`, `observability`, `persistence/gorm`. Esto es el reemplazo de `core/*` por `platform/*` que la memoria del repo documenta. Muchos cambios de import path aparentemente "menores" (p.ej. `internal/reviewproxy/client.go`, solo 6 líneas) son consecuencia de esta migración. **Riesgo de extracción alto**: cualquier paquete que dependa de un módulo `platform/*` arrastra el bump de `go.mod`/`go.sum`.

---

## 4. Migraciones encontradas (`migrations_v4/`, 223–234)

| # | Nombre | Feature |
|---|---|---|
| 223 | `actors_safe_migration` (+568 líneas up) | 007 |
| 224 | `tenant_security_foundation` | 003 (backfill→constraints) |
| 225 | `tenant_constraints_validation` | 003 |
| 226 | `customer_actor_master_link` | 007 |
| 227 | `crudar_archive_batches` | 002/009 |
| 228 | `crudar_remaining_archive_metadata` | 002/009 |
| 231 | `consolidate_actor_archived_at` | 007/009 |
| 232 | `document_archived_inclusion_in_reports` | 002/009 |
| 233 | `archived_invariant_triggers` | 002/009 |
| 234 | `actor_unique_normalized_name` | 004/007 |

Nota: **faltan 229 y 230** en la secuencia → confirmar si se consolidaron, se descartaron, o ya estaban en `develop`. `cmd/migrate/*` (main, migrate_gorm, migrate_sql + test) y `scripts/db/db_migrate_up.sh` también cambiaron.

**Riesgo**: 224/225 hacen backfill→constraints multitenant. Si hay datos stale en un entorno, los constraints fallan. 233 añade triggers de invariantes de `archived`. Aplicar en orden y validar contra `scripts/db/multi_tenant_golden_master.sql` / `tenant_isolation_audit.sql`.

---

## 5. Archivos críticos / compartidos / peligrosos

Tocados en esta rama y **mezclados entre features** (traer con `git restore -p`, no en bloque):

- `wire/wire.go`, `wire/wire_gen.go` — DI mezclado. `wire/actor_providers.go` (nuevo) → 007; `wire/companion_providers.go` (nuevo) → 012. También M: admin, ai, config, data_integrity, labor, lot, middleware, stock, supply, work_order providers.
- `cmd/api/main.go`, `cmd/api/http_server.go` — wiring de rutas + arranque, mezclados (007/008/009/012).
- `cmd/config/loadconfig.go` + nuevos `companion.go`, `reporting.go`, `security.go` (borra `ai.go`) — feature-005, funda 012 y 023.
- `go.mod` / `go.sum` — bumps `platform/*` + (ya DONE #124) go-jose v4.1.4 y x/net v0.55.0: **separar los bumps ya mergeados** de los `platform/*` nuevos.
- `Makefile` — feature-019.
- `internal/shared/{handlers,repository,models/base.go}/**` — base compartida (tenancy + lifecycle).

Mayor churn individual (referencia de tamaño): `internal/project/repository.go` (+2075), `internal/actor/repository.go` (+1158 nuevo), `internal/data-integrity/usecases.go` (-1095, gran limpieza), `internal/shared/lifecycle/archive_cleanup.go` (+818 nuevo), `internal/customer/repository.go`, `internal/labor/repository.go`, `internal/supply/repository_movement.go` (+562).

---

## 6. Cambios sospechosos / experimentales (revisar antes de portar)

1. **Integración Nexus en `internal/axis/` — NO está en la tabla de features.** La feature-012 documenta solo Companion, pero la fuente trae `internal/axis/nexus_client.go` y `nexus_types.go` (cliente que pide "decisiones" a un servicio Nexus con scopes `nexus:requests:*`, JWT HS256 con el mismo patrón que Companion). Parece **trabajo experimental/incompleto** apilado junto a Companion. **Recomendación**: aislarlo; no incluirlo en el paquete 012 salvo decisión explícita.

2. **Residuo del refactor de tenancy (feature-001 incompleto).** `MaybeTenantScope` todavía aparece en `internal/supply/repository_movement.go` (línea 409, en comentario que explica el caso de strings con `.`) y en `internal/shared/authz/authz_test.go`. El drop no fue 100% — `internal/supply` es justo el dominio de mayor churn (29 archivos). Verificar que el comportamiento de scope en movimientos de supply sea el esperado tras el port.

3. **`PONTI CAMBIOS PARA PODER CERRAR.pdf` borrado** y `scripts/db/schema.snapshot.sql` borrado (-6589 líneas) + `repair_stocks_investor_granularity.sql` borrado — limpieza de artefactos; confirmar que nada en CI/tooling dependa del snapshot.

4. **`internal/data-integrity/usecases.go` -1095 líneas**: gran reducción. Parte (tentative-prices) **ya está DONE #121 → excluir de feature-018**; verificar que lo que queda no re-introduzca lo ya mergeado.

5. **Borrado de JWT legacy**: `internal/shared/utils/jwt_tools.go`, `internal/shared/utils/strings.go`, `internal/platform/http/middlewares/gin/require_jwt.go` — feature-027 (cleanup). Confirmar que nada vivo los importe tras el port (el reemplazo viene de `platform/authn|security`).

---

## 7. Tests afectados

- 101 `_test.go` tocados. Bloques claros:
  - **26 `repository_tenant_test.go`** (validan 001/003) — feature-025, dependen de 001.
  - `repository_archived_refs_test.go` (project), `handler_test.go` varios, `repository_rename_test.go` (project) — validan 009/010.
  - `internal/shared/lifecycle/*_test.go` (incl. `invariant_e2e_test.go`) — corazón de 002/009.
  - `internal/axis/client_test.go`, `internal/admin/me_context_test.go`, `internal/actor/handler_test.go` — acompañan a 012/008/007.
- **Recomendación**: los tests de tenant/archive (feature-025) pueden ir como **follow-up** tras mergear 001/009, pero conviene traerlos junto a su módulo para no dejar BE-first sin red.

---

## 8. Configuración

- `cmd/config` modularizado (005): split en `auth/companion/reporting/security/service/http_server`, `.env.example` actualizado.
- `Dockerfile`, `docker-compose.yml` modificados (021) — `docker-compose` ahora referencia servicios `axis` (Companion/Nexus).
- `lefthook.yml` **nuevo** (022) — hooks locales, opcional.
- `.github/workflows/*` (020): ci-pr, deploy-{dev,staging,prod}, audit-service-alignment, reset-dev-db-from-prod modificados. **Pueden romper el deploy si se traen sin el resto** del BE.

---

## 9. Riesgos generales de extracción

1. **Acoplamiento por `platform/*` en `go.mod`**: casi cualquier paquete arrastra el bump. Decidir si la migración new-cns3 va como base única antes de cualquier feature, o si se mantiene `core/*` y se rebajan imports (mucho más costoso).
2. **Archivos mezclados** (wire, cmd/api, cmd/config, go.mod): usar `git restore -p` por hunk, agrupando cada hunk con su módulo. No traer wire/cmd en bloque.
3. **Orden de merge BE-first** crítico para full-stack: 001 → 002/003 → 004 → 007 → 008/009/010, antes de portar el FE en el otro repo.
4. **Migraciones backfill→constraint (224/225/233)**: ejecutar en entorno limpio/auditado; alto riesgo con datos stale.
5. **Cambios de contrato** (009 archive endpoints, 011 shape `project_id` minúscula, 013 XLSX→CSV): coordinar con FE o el dropdown/export se rompe.
6. **No incluir lo ya DONE**: tentative-prices (#121, excluir de 018), lot-metrics/total_tons (#117/#121/#124), dependency-bumps go-jose/x/net (#124, separar de 021).
7. **Numeración de migraciones** (faltan 229/230, y el repo usa el "set viejo" según la memoria de reset-local-db) → revisar contra `develop` antes de aplicar.

---

## 10. Recomendaciones de empaquetado (sugerencias; comandos git aquí son orientativos)

1. **Base BE primero**: feature-001 (tenancy) como primer PR — toca todo, mejor aislado y verde antes de lo demás. Cerrar el residuo `MaybeTenantScope` en supply.
2. **feature-002/009 juntos o encadenados** (lifecycle funda archive). Sugerir **PRs por entidad** para el archive surface (009) y para master-data (014, en el otro repo), igual que ya se hizo con lots/workorders (#104/#117).
3. **feature-005 antes de 012 y 023** (config funda ambos).
4. **feature-007 (actores)** como PR grande propio, con sus migraciones 223/226/231/234 y `wire/actor_providers.go`; luego 008 y 010 encima.
5. **feature-012**: traer Companion **sin** el cliente Nexus/axis-nexus salvo decisión explícita (ver §6.1).
6. **feature-013**: confirmar consumo FE de los export antes de borrar Excel.
7. **Tests (025)** y **docs (024)**: pueden seguir a sus módulos o ir como follow-up independiente.
8. **CI (020) y deploy (021)**: por repo y al final; no traerlos sueltos. Separar los bumps ya mergeados (#124).

---

## 11. Relación con el otro repo (FE = `ponti-fe` / web, BFF en `api/`)

Este BE es la mitad backend de varias features **full-stack**. El FE/BFF correspondiente vive en el otro repo y depende de que el BE vaya **first**:

- **007 actors**: BE `/api/v1/actors` (+migr 223/226/231/234) ↔ FE `useActors` + `master-data/actors` + BFF `api/src/routes/actors.ts`.
- **008 identity/me_context**: BE `internal/admin/me_context.go` (`/me` con array de tenants) ↔ FE `TenantContext` + Navbar switcher + BFF `me.ts`/`authMiddleware`/`requestContext`.
- **010 projects**: BE scope/creator + project-archive bridge ↔ FE `pages/admin/projects` + BFF `projects.ts`.
- **011 campaign DTO**: BE serializa `project_id`/`id`/`name` en minúscula ↔ FE campaigns. **Desync = dropdown de campañas vacío** (coordinar shape).
- **012 AI Companion**: BE `internal/axis` + `ai` adapter + `companion_providers` ↔ FE `pages/admin/ai` + BFF `ai.ts`/`managerChatStreamProxy`.
- **013 CSV / 009 archive / 018 data-integrity**: cambian contratos que el FE consume directamente (export, endpoints archive, data-integrity admin).

Las features **solo-FE** (006 design-system, 014 master-data pages, 015 dashboard, 016 access/notifications, 017 dollar/commerce, 026 fe-test-infra) **no están en este repo**; se documentan en su propio análisis del repo FE. La feature-006 (design-system) es la base FE equivalente a lo que 001 es en BE, y `router.tsx`/`main.tsx` allá son archivos mezclados como lo son `wire.go`/`cmd/api/main.go` acá.
