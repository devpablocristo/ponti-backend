# Cross-Repo Map — descomposición de `develop-problematico`

**Fecha del análisis:** 2026-05-30
**Repo de este documento:** Backend Go — `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
**Rango fuente:** `0972e565..777e5f6a`
**SOURCE de extracción:** `develop-problematico~1` (SHA `777e5f6a`, el *pico* de la rama de integración) — **NUNCA** el tip.
**Destino:** `develop`.

> **Contexto.** `develop-problematico` fue una rama de integración que acumuló varias líneas de trabajo
> (new-cns3 + projects + admin + ...). Su **último** commit (`66f2b602`, `restore: app a estado pre-new-cns3 + mantener tooling local actual`)
> es un *restore* que **vacía** la rama. Por eso la fuente real de extracción es `develop-problematico~1` (`777e5f6a`),
> el commit inmediatamente anterior al restore. Todo `git restore -s <SHA>` debe usar `777e5f6a`, no el tip.

> **Alcance de este doc.** Mapa cross-repo: por cada feature se indica qué porción vive en el **Backend (BE)** y cuál en el
> **Frontend (FE)** (repo separado: web + BFF/api), la dependencia entre repos, el orden de merge recomendado, el riesgo de
> desincronización y los contratos/migraciones/config implicados. Este documento es **conceptualmente idéntico** al que vive
> en el repo FE; cambia sólo el punto de vista de "qué vive aquí".

> **Convención de repos.** En la nomenclatura del proyecto: `fe` = web (UI React + BFF `api/`), `be` = core (este repo, Go).
> El `core/*` histórico fue deprecado por `platform/*` como parte de new-cns3.

---

## Tabla resumen

| ID | Nombre | FE | BE | Merge recomendado | Riesgo cross-repo | Notas |
|----|--------|----|----|-------------------|-------------------|-------|
| 001 | be-platform-tenancy-refactor | no | sí | BE independiente | Ninguno | Sin cambios FE. Refactor interno, sin cambio de contrato API. Base de todo el BE. |
| 002 | be-crudar-lifecycle-framework | no | sí | BE-first | Ninguno | Sin cambios FE. Funda 009. Migraciones 227/228/232/233. |
| 003 | be-multitenant-db-hardening | no | sí | BE-first (deps 001) | Bajo | Sin cambios FE. Migraciones 224/225. Riesgo de datos, no de contrato. |
| 004 | shared-text-propername | no | sí | BE independiente | Ninguno | Sin cambios FE. Util chico. Bloquea 007. |
| 005 | be-config-modularization | no | sí | BE independiente | Ninguno | Sin cambios FE. Funda 012 y 023. Cambia `.env.example`. |
| 006 | fe-design-system | sí | no | FE independiente | Ninguno | **Sin cambios BE.** No vive en este repo. Base de todo el FE. |
| 007 | actor-system | sí | sí | BE-first, luego FE | **Alto** | FULL-STACK. Nuevo contrato `/api/v1/actors`. Migr 223/226/231/234. |
| 008 | identity-tenant-context | sí | sí | BE-first, luego FE | **Alto** | FULL-STACK. `/me` cambia shape (array de tenants). |
| 009 | crudar-archive-surface | no | sí | BE-first (FE en 014/006) | **Alto** | Carpeta solo-BE, pero su contraparte FE vive en **014 + 006**. Cambio de contrato en ~20 dominios. |
| 010 | projects | sí | sí | BE-first, luego FE | Medio | FULL-STACK. Depende de 007 y 009. |
| 011 | campaign-dto-projectid | sí | sí | coordinado (shape change) | **Alto** | FULL-STACK. Cambio de casing en JSON; si desync, dropdown de campañas vacío. |
| 012 | ai-companion-integration | sí | sí | BE-first | Medio | FULL-STACK. Cliente Companion + JWT. Depende de 005. |
| 013 | be-csv-export | no | sí | BE-first | Medio | Solo-BE en código, **pero** cambia export XLSX -> CSV; revisar consumo FE. |
| 014 | fe-master-data-pages | sí | no | FE tras 007/009 | Medio | **Sin cambios BE.** No vive en este repo. Familia de 212 archivos; 1 PR por entidad. |
| 015 | fe-dashboard-consolidation | sí | no | FE independiente | Ninguno | **Sin cambios BE.** No vive en este repo. |
| 016 | fe-access-notifications | sí | no | FE independiente | Ninguno | **Sin cambios BE.** No vive en este repo. |
| 017 | fe-dollar-commerce-forms | sí | no | FE independiente | Ninguno | **Sin cambios BE.** No vive en este repo. |
| 018 | data-integrity-admin | sí | sí | coordinado | Medio | FULL-STACK. `internal/data-integrity` (BE) + pages/admin (FE). Excluir tentative-prices (DONE #121). |
| 019 | be-local-tooling-db-scripts | no | sí | BE independiente | Ninguno | Sin cambios FE. `scripts/` + `Makefile`. Bajo riesgo. |
| 020 | ci-workflows | sí | sí | por repo | Medio | `.github/workflows` en ambos repos. Puede romper deploy si se trae aislado. |
| 021 | build-and-deploy-config | sí | sí | por repo | Medio | BE Dockerfile/compose/go.mod-sum (bumps DONE #124, separar); FE vite/tailwind/etc. |
| 022 | lefthook-git-hooks | sí | sí | por repo | Ninguno | `lefthook.yml` en ambos. Local tooling, opcional. |
| 023 | be-wire-di | no | sí | acompaña a su módulo | Medio | Sin cambios FE. `wire/` + `cmd/api`. Archivos MEZCLADOS: traer con `restore -p`. |
| 024 | openapi-and-docs | sí | sí | independiente | Ninguno | Docs en ambos repos. Sin runtime. |
| 025 | be-test-coverage | no | sí | sigue a su módulo | Ninguno | Sin cambios FE. Tests (45 archivos). Validan 001/009. Follow-up. |
| 026 | fe-test-infra | sí | no | FE independiente | Ninguno | **Sin cambios BE.** No vive en este repo. |
| 027 | be-cleanup-domain-purity | no | sí | BE independiente | Bajo | Sin cambios FE. Limpieza (staticcheck, json-tags, governance, jwt legacy). |

**Leyenda de riesgo cross-repo.** *Alto* = cambio de contrato/shape que rompe al otro repo si se mergea desfasado.
*Medio* = cambio observable por el otro repo (export, dependencia de orden) pero degradación parcial, no caída total.
*Bajo* = riesgo de datos/entorno local, no de contrato. *Ninguno* = aislado por repo o solo-docs/tests.

---

## Detalle por feature

### feature-001 — be-platform-tenancy-refactor `[refactor]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** drop de `MaybeTenantScope` -> `tenancy.Scope` en ~23 repos. Es la base interna de todo el BE.
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna.
- **Orden de merge:** BE independiente; primero del set BE.
- **Riesgo de desincronización:** ninguno (refactor interno, **sin cambio de contrato API**).
- **Contratos API:** sin cambios. **Migraciones:** ninguna. **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-002 — be-crudar-lifecycle-framework `[refactor]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** `internal/shared/lifecycle` + migraciones **227/228/232/233**. Funda 009.
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna. **Deps internas BE:** -.
- **Orden de merge:** BE-first (antes de 009).
- **Riesgo de desincronización:** ninguno hacia FE.
- **Contratos API:** sin cambios directos. **Migraciones:** 227, 228, 232, 233. **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-003 — be-multitenant-db-hardening `[migration]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** migraciones **224/225** (backfill -> constraints).
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna. **Deps internas BE:** 001.
- **Orden de merge:** BE-first, después de 001.
- **Riesgo de desincronización:** bajo (riesgo de **datos stale**, no de contrato). Verificar backfill antes de aplicar constraints.
- **Contratos API:** sin cambios. **Migraciones:** 224 (backfill), 225 (constraints). **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-004 — shared-text-propername `[feature]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** util chico de texto (proper name).
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna.
- **Orden de merge:** BE independiente. Bloquea 007 (normalización de nombres de actores).
- **Riesgo de desincronización:** ninguno.
- **Contratos API:** sin cambios. **Migraciones:** ninguna. **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-005 — be-config-modularization `[infra]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** split de `cmd/config` + `.env.example`. Funda 012 y 023.
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna.
- **Orden de merge:** BE independiente; antes de 012 y 023.
- **Riesgo de desincronización:** ninguno hacia FE (riesgo de entorno BE si no se actualiza `.env`).
- **Contratos API:** sin cambios. **Migraciones:** ninguna. **Config/env:** **sí** — `cmd/config/loadconfig.go` + `.env.example` (nuevas claves). **Feature-flags:** ninguno.

### feature-006 — fe-design-system `[refactor]`
- **FE:** sí | **BE:** no — **sin cambios BE. No vive en este repo.**
- **Vive aquí (BE):** nada.
- **Vive en el otro repo (FE):** primitivos (feedback, button/drawer, input, modal, card, filters, `ArchivedListPage`), lib (format/theme/lifecycle), router/main shell. Base de todo el FE.
- **Dependencia entre repos:** ninguna.
- **Orden de merge:** FE independiente; primero del set FE.
- **Riesgo de desincronización:** ninguno cross-repo. **Aviso FE-interno:** `router.tsx` / `main.tsx` son archivos MEZCLADOS.
- **Contratos API:** N/A. **Migraciones:** N/A. **Config/env:** N/A. **Feature-flags:** ninguno.

### feature-007 — actor-system `[feature]` — FULL-STACK
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** expone `/api/v1/actors` + migraciones **223/226/231/234**.
- **Vive en el otro repo (FE):** `useActors`, `master-data/actors`, BFF `api/src/routes/actors.ts`.
- **Dependencia entre repos:** FE consume el contrato `/api/v1/actors` del BE. **Deps:** 001, 002, 003, 004, 006.
- **Orden de merge:** **BE-first, luego FE.** La feature grande.
- **Riesgo de desincronización:** **Alto** — si FE mergea antes que BE, las pages de actores quedan sin endpoint.
- **Contratos API:** **nuevo** `GET/POST/... /api/v1/actors`. **Migraciones:** 223, 226, 231, 234. **Config/env:** ninguna nueva. **Feature-flags:** ninguno.

### feature-008 — identity-tenant-context `[feature]` — FULL-STACK
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** admin `me_context` — `/me` devuelve **array de tenants**.
- **Vive en el otro repo (FE):** `TenantContext`, Navbar switcher, `general-entities-admin`, login; BFF `me.ts`, `authMiddleware`, `requestContext`.
- **Dependencia entre repos:** FE depende del nuevo shape de `/me`. **Deps:** 007.
- **Orden de merge:** **BE-first, luego FE.**
- **Riesgo de desincronización:** **Alto** — el shape de `/me` cambia; FE viejo contra BE nuevo (o viceversa) rompe el contexto de tenant y el login.
- **Contratos API:** `GET /me` cambia respuesta a array de tenants. **Migraciones:** ninguna nueva propia. **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-009 — crudar-archive-surface `[refactor]`
- **FE:** no (en *esta* carpeta) | **BE:** sí. **Su contraparte FE vive en 014 y 006.**
- **Vive aquí (BE):** cambio de contrato de archivado en ~20 dominios — **123 archivos** (handlers/usecases/repos).
- **Vive en el otro repo (FE):** la UI de archivado: páginas en **014** (`fe-master-data-pages`) y el componente `ArchivedListPage` en **006** (`fe-design-system`).
- **Dependencia entre repos:** FE de 014/006 consume los nuevos endpoints de archivado. **Deps internas BE:** 002.
- **Orden de merge:** **BE-first**; la contraparte FE entra con 014/006.
- **Riesgo de desincronización:** **Alto** — cambio de contrato amplio: el botón "borrar" del FE debe pasar de `DELETE` a `POST .../archive`. Si desync, los listados de archivados (FE) llaman endpoints inexistentes o el delete duro/blando no coincide.
- **Contratos API (en ~20 dominios):**
  - `DELETE /:id` -> **eliminado**, reemplazado por `POST /:id/archive`.
  - **nuevo** `DELETE /:id/hard` (borrado físico).
  - **nuevo** `GET /archived` (listado de archivados).
- **Migraciones:** las del framework de lifecycle (ver 002). **Config/env:** ninguna. **Feature-flags:** ninguno.
- **Sugerencia:** PRs **por entidad**, alineados con los PRs por-entidad del FE (014).

### feature-010 — projects `[feature]` — FULL-STACK
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** project-archive-entidades-bridge + scope/creator.
- **Vive en el otro repo (FE):** `pages/admin/projects` + BFF `projects.ts`.
- **Dependencia entre repos:** FE consume el contrato de projects del BE. **Deps:** 007, 009.
- **Orden de merge:** **BE-first, luego FE.**
- **Riesgo de desincronización:** Medio — depende de 007 y 009 ya estar en ambos repos.
- **Contratos API:** endpoints de projects (scope/creator). **Migraciones:** las que trae el bridge (verificar en el paquete BE 010). **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-011 — campaign-dto-projectid `[bugfix]` — FULL-STACK
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** serialización de `project_id` / `id` / `name` en **minúscula**.
- **Vive en el otro repo (FE):** consumo en `campaigns`.
- **Dependencia entre repos:** acopladísimo por el **casing del JSON**.
- **Orden de merge:** **coordinado** (shape change) — mergear BE y FE juntos o muy seguidos.
- **Riesgo de desincronización:** **Alto** — si desync, el **dropdown de campañas queda vacío** (FE espera un casing que BE no emite, o al revés).
- **Contratos API:** cambio de casing de campos en el DTO de campañas (`project_id`, `id`, `name`). **Migraciones:** ninguna. **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-012 — ai-companion-integration `[feature]` — FULL-STACK
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** `internal/axis` (cliente Companion + JWT) + adapter `ai` + `companion_providers`.
- **Vive en el otro repo (FE):** `pages/admin/ai` + BFF `ai.ts` / `managerChatStreamProxy`.
- **Dependencia entre repos:** FE proxea el chat stream al BE. **Deps internas BE:** 005.
- **Orden de merge:** **BE-first.**
- **Riesgo de desincronización:** Medio — el proxy FE necesita el endpoint/JWT del BE; degradación de la sección AI si desync.
- **Contratos API:** endpoints de companion/AI + auth JWT. **Migraciones:** ninguna nueva propia. **Config/env:** **sí** — credenciales/URL de Companion + JWT (asentado sobre 005). **Feature-flags:** posible toggle de provider en `companion_providers`.

### feature-013 — be-csv-export `[refactor]`
- **FE:** no — **sin cambios FE en código** | **BE:** sí.
- **Vive aquí (BE):** `internal/shared/csvexport` + csv-service por dominio; **borra excel**.
- **Vive en el otro repo (FE):** no hay código FE en el paquete, **pero** el FE consume los endpoints de export.
- **Dependencia entre repos:** los endpoints de export pasan de **XLSX a CSV**.
- **Orden de merge:** **BE-first.**
- **Riesgo de desincronización:** Medio — si el FE asume descarga `.xlsx` (content-type / extensión / parseo), romperá. **Revisar consumo FE de export** antes de mergear.
- **Contratos API:** endpoints `.../export` cambian formato de respuesta (XLSX -> CSV); cambia content-type y nombre de archivo. **Migraciones:** ninguna. **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-014 — fe-master-data-pages `[feature]`
- **FE:** sí | **BE:** no — **sin cambios BE. No vive en este repo.**
- **Vive aquí (BE):** nada.
- **Vive en el otro repo (FE):** **familia de 212 archivos** — `customers / fields / lots / workorders / crops / investors / managers / labors / supplies / supply-movements / stock` + hooks + BFF routes/utils.
- **Dependencia entre repos:** consume contratos de 007 (actors) y 009 (archive). **Deps:** 006, 007, 009.
- **Orden de merge:** **FE, tras 007 y 009** (que deben estar en BE).
- **Riesgo de desincronización:** Medio — depende de actor-system y archive-surface ya activos en BE.
- **Contratos API:** consume (no define) los de 007/009. **Migraciones:** N/A (FE). **Config/env:** N/A. **Feature-flags:** ninguno.
- **IMPRESCINDIBLE (lado FE):** agrupar la file-list **por entidad** y proponer **1 PR por entidad**. `lots`/`workorders` están parcialmente DONE (#104/#117).

### feature-015 — fe-dashboard-consolidation `[refactor]`
- **FE:** sí | **BE:** no — **sin cambios BE. No vive en este repo.**
- **Vive en el otro repo (FE):** `pages/admin/dashboard` + `useDashboard`. **Deps:** 006.
- **Orden de merge:** FE independiente. **Riesgo:** ninguno cross-repo.
- **Contratos API / Migraciones / Config / Flags:** N/A.

### feature-016 — fe-access-notifications `[refactor]`
- **FE:** sí | **BE:** no — **sin cambios BE. No vive en este repo.**
- **Vive en el otro repo (FE):** `pages/admin/access` + notifications. **Deps:** 006.
- **Orden de merge:** FE independiente. **Riesgo:** ninguno cross-repo.
- **Contratos API / Migraciones / Config / Flags:** N/A.

### feature-017 — fe-dollar-commerce-forms `[feature]`
- **FE:** sí | **BE:** no — **sin cambios BE. No vive en este repo.**
- **Vive en el otro repo (FE):** `pages/admin/dollar` + commercialization. **Deps:** 006.
- **Orden de merge:** FE independiente. **Riesgo:** ninguno cross-repo.
- **Contratos API / Migraciones / Config / Flags:** N/A.

### feature-018 — data-integrity-admin `[feature]` — FULL-STACK
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** `internal/data-integrity`.
- **Vive en el otro repo (FE):** `pages/admin/data-integrity` + `useDatabase`.
- **Dependencia entre repos:** FE consume los endpoints de data-integrity del BE.
- **Orden de merge:** **coordinado.**
- **Riesgo de desincronización:** Medio.
- **OJO — exclusión:** la parte **tentative-prices ya está DONE (#121)** -> **excluirla** de este paquete en ambos repos.
- **Contratos API:** endpoints de data-integrity (menos tentative-prices). **Migraciones:** verificar en paquete BE 018. **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-019 — be-local-tooling-db-scripts `[infra]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** `scripts/` (db, data-audit, lint-tenant-leaks, golden-master, smoke-companion, export-ai, reset-local-db-from-prod) + `Makefile`. Son los 18 sobrevivientes del 3-dot + más.
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna.
- **Orden de merge:** BE independiente. **Riesgo:** bajo (tooling local).
- **Contratos API:** ninguno. **Migraciones:** ninguna (gestiona, no define). **Config/env:** indirecta (los scripts leen `.env`). **Feature-flags:** ninguno.
- **Nota local (memoria):** `reset-local-db-from-prod` asumía numeración nueva de migraciones; aplicar los fixes para el set viejo.

### feature-020 — ci-workflows `[infra]`
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** `.github/workflows` del backend.
- **Vive en el otro repo (FE):** `.github/workflows` del frontend.
- **Dependencia entre repos:** independientes por repo, pero ambos parte de la misma feature lógica.
- **Orden de merge:** **por repo.**
- **Riesgo de desincronización:** Medio — pueden **romper el deploy** si se traen sin el resto de su feature (p.ej. workflow que asume artefactos/targets nuevos).
- **Contratos API:** ninguno. **Migraciones:** ninguna. **Config/env:** secrets/vars de CI. **Feature-flags:** ninguno.
- **Nota CI (memoria):** prefetch de módulos apunta a `platform` en `ci-pr` (core deprecado).

### feature-021 — build-and-deploy-config `[config]`
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** `Dockerfile` / `compose` / `go.mod` / `go.sum`.
- **Vive en el otro repo (FE):** `vite` / `tailwind` / `eslint` / `knip` / `tsconfig` / lockfiles / generated client.
- **Dependencia entre repos:** independientes por repo.
- **Orden de merge:** **por repo.**
- **Riesgo de desincronización:** Medio (build/deploy).
- **EXCLUSIÓN:** los **dependency bumps (go-jose, x/net) ya están DONE (#124)** -> **separar/excluir** del paquete BE 021.
- **Contratos API:** ninguno. **Migraciones:** ninguna. **Config/env:** build config. **Feature-flags:** ninguno.

### feature-022 — lefthook-git-hooks `[config]`
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** `lefthook.yml`.
- **Vive en el otro repo (FE):** `lefthook.yml`.
- **Dependencia entre repos:** ninguna.
- **Orden de merge:** **por repo.** Local tooling, **opcional.**
- **Riesgo de desincronización:** ninguno.
- **Contratos API / Migraciones / Config-runtime / Flags:** N/A.

### feature-023 — be-wire-di `[infra]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** `wire/` + `cmd/api`. `wire/actor_providers` -> 007, `companion_providers` -> 012.
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna. **Deps internas BE:** 001, 005, 007, 008, 009, 012.
- **Orden de merge:** **acompaña a su módulo** (cada provider entra con su feature).
- **Riesgo de desincronización:** Medio (interno BE — orden de DI). **Archivos MEZCLADOS:** `wire/wire.go`, `wire/wire_gen.go`, `cmd/api/main.go` -> traer con `git restore -p` (parcial) junto a cada módulo, **desde `777e5f6a`**.
- **Contratos API:** ninguno propio. **Migraciones:** ninguna. **Config/env:** ninguna. **Feature-flags:** ninguno.

### feature-024 — openapi-and-docs `[docs]`
- **FE:** sí | **BE:** sí.
- **Vive aquí (BE):** `docs/openapi` + CRUDAR / error-catalog / multi-tenant-evidence + `CLAUDE.md` / `CRUDAR_PLAN.md`.
- **Vive en el otro repo (FE):** `docs/` + `docs/audit` (visual regression, posiblemente generado) + `RESPONSIVE_GUIDELINES` + `PR-92.md`.
- **Dependencia entre repos:** ninguna (docs).
- **Orden de merge:** **independiente.** **Riesgo:** ninguno (sin runtime).
- **Contratos API:** documenta, no define. **Migraciones / Config / Flags:** N/A.

### feature-025 — be-test-coverage `[tests]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** `handler_test` + `repository_tenant_test` + `repository_archived_refs_test` (**45 archivos**). Validan 001 y 009.
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna. **Deps internas BE:** 001, 002, 009.
- **Orden de merge:** **sigue a su módulo** (puede ir como follow-up).
- **Riesgo de desincronización:** ninguno.
- **Contratos API / Migraciones / Config / Flags:** N/A.

### feature-026 — fe-test-infra `[tests]`
- **FE:** sí | **BE:** no — **sin cambios BE. No vive en este repo.**
- **Vive en el otro repo (FE):** `ui/.vite-smoke` + `ui/e2e` + `api/test` + `api/src/mocks`. **Deps:** 006.
- **Orden de merge:** FE independiente. **Riesgo:** ninguno cross-repo.
- **Contratos API / Migraciones / Config / Flags:** N/A.

### feature-027 — be-cleanup-domain-purity `[cleanup]`
- **FE:** no — **sin cambios FE.** | **BE:** sí.
- **Vive aquí (BE):** staticcheck + **report domain json-tag removal** + remove `core/governance` + borrar jwt utils legacy.
- **Vive en el otro repo (FE):** nada.
- **Dependencia entre repos:** ninguna. **Deps internas BE:** 001.
- **Orden de merge:** BE independiente.
- **Riesgo de desincronización:** Bajo — **atención:** la json-tag removal del dominio `report` afecta la serialización; coordinar con el FE de reports (la parte FE `reports-dark-mode` ya está DONE #105, pero la limpieza de json-tags BE **no** está porteada y va aquí). Verificar que el FE de reports no dependa de tags removidos.
- **Contratos API:** posible cambio de serialización en `report` (json-tags). **Migraciones:** ninguna. **Config/env:** ninguna. **Feature-flags:** ninguno.

---

## DONE — ya en `develop`, sin paquete (excluir de la descomposición)

| Trabajo | Repo(s) | PR(s) | Acción al descomponer |
|---------|---------|-------|------------------------|
| table-select-filters (filtros de tabla) | FE | #104 | Excluir; ya en develop. `lots`/`workorders` de 014 parcialmente cubiertos. |
| reports-dark-mode | FE | #105 | Excluir el FE. **OJO:** la limpieza de json-tags del dominio BE **NO** está porteada -> va en **027**. |
| lot-metrics / total_tons | FE + BE | #117 / #121 / #124 | Excluir; ya en develop. |
| tentative-prices | FE + BE | #121 / #124 | Excluir; **quitar de feature-018** (no re-portear). |
| dependency-bumps (go-jose, x/net) | BE | #124 | Excluir; **separar de feature-021** (no re-portear). |

---

## Notas operativas de extracción (sugerencias — NO ejecutar desde este doc)

> Todos los comandos `git` aquí son **sugerencias**. Este documento no ejecuta cambios de código.

- **SOURCE siempre `777e5f6a`** (`develop-problematico~1`), nunca el tip `66f2b602` (restore que vacía).
  - Extracción por carpeta sugerida: `git restore -s 777e5f6a -SW -- <ruta>` sobre una rama de feature derivada de `develop`.
  - Para archivos MEZCLADOS (ver lista abajo): `git restore -s 777e5f6a -p -- <archivo>` y elegir hunks por feature.
- **Orden global recomendado (BE):** 001 -> 002 -> 003/004/005 -> 007 -> 008 -> 009 -> 010 -> 011(coord) -> 012 -> 013 -> 018 -> resto (019/020/021/022/023/024/025/027), con 023 y 025 acompañando a sus módulos.
- **Sincronía cross-repo crítica (BE-first, luego FE):** 007, 008, 010, 012; **coordinado simultáneo:** 011 (casing), 018; **FE-tras-BE:** 009 (su FE en 014/006), 014.

### Archivos compartidos / peligrosos en este repo (BE)
`wire/wire.go`, `wire/wire_gen.go`, `cmd/api/main.go`, `cmd/api/http_server.go`,
`cmd/config/loadconfig.go`, `go.mod`, `go.sum`, `Makefile`,
`internal/shared/handlers/**`, `internal/shared/models/base.go`, `internal/shared/repository/**`.

> Estos archivos son tocados por varias features; **no** traerlos en bloque. Usar `restore -p` y asignar cada hunk
> a la feature que corresponde (notablemente 023 los concentra para el grafo de DI).
