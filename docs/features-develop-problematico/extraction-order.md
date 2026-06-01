# Orden de extracción — descomposición de `develop-problematico`

**Fecha del análisis:** 2026-05-30
**Repo:** ponti-backend (`/home/pablocristo/Proyectos/pablo/ponti/core`)
**Rango fuente:** `0972e565..777e5f6a`
**SOURCE de extracción:** `develop-problematico~1` (SHA `777e5f6a81edf48bc9d7a9450ad77cf08780c9c0`) — el **pico** de la rama de integración.
**Destino:** `develop`.

> ⚠️ **NUNCA extraer desde el tip de `develop-problematico`** (SHA `66f2b602`). Ese último commit es un *restore* que vacía la rama. Toda extracción (`git restore --source=...`, `git checkout <sha> -- <path>`, `git cherry-pick`) debe partir de `777e5f6a`.

---

## 0. Contexto

`develop-problematico` fue una rama de integración donde se mezclaron varios trenes de trabajo (new-cns3 + projects + admin + tooling + docs). Su HEAD es un restore que la dejó vacía, así que la fuente real es `develop-problematico~1`. El objetivo de este documento es reconstruir un **orden de PRs hacia `develop`** que respete dependencias, minimice conflictos en archivos compartidos y permita validar incrementalmente.

Este repo es **solo backend (Go)**. Las features FE-only no se extraen acá; se listan abajo para coordinación cross-repo, pero su trabajo ocurre en el repo `fe`/web.

**Features con presencia BE en este repo** (tienen carpeta de paquete generada):
001, 002, 003, 004, 005, 007, 008, 009, 010, 011, 012, 013, 018, 019, 020, 021, 022, 023, 024, 025, 027.

**Features FE-only (NO se extraen acá):** 006, 014, 015, 016, 017, 026.

---

## 1. Criterio de ordenamiento

El orden sigue una progresión de **riesgo creciente y dependencia creciente**:

1. **Base / bajo riesgo** (001, 004, 005, 019, 027 + el FE 006 en su repo) — fundaciones que no cambian contratos API.
2. **Fundaciones BE** (002, 003) — lifecycle framework y hardening multitenant; fundan las features grandes.
3. **Contratos full-stack BE-first** (007 → 008 → 009 → 010) — exponen/cambian endpoints; el BE va primero y el FE consume después.
4. **Shapes sensibles** (011, y la cara FE de 009) — cambios de forma de payload que rompen si hay desync.
5. **FE consumidor** (014, 015–018) — en el repo FE, tras sus deps BE.
6. **Infra / tests / docs** (020–026) — acompañan o siguen a sus módulos; bajo riesgo de runtime, alto valor de cierre.

---

## 2. Grafo de dependencias (BE de este repo)

```
001 (tenancy refactor) ── base de TODO el BE
  ├─> 003 (multitenant db hardening)
  ├─> 007 (actor-system) ── también dep de 002,003,004,006
  ├─> 023 (wire/DI)
  └─> 027 (cleanup domain purity)

002 (crudar lifecycle) ──> 009 (archive surface)
                       └─> 007

004 (text/propername) ──> 007 (normalización nombres de actores)

005 (config modularization) ──> 012 (ai-companion) , 023 (wire)

007 (actor-system)  ── deps: 001,002,003,004,(006 FE)
  └─> 008 (identity tenant context)
        └─> 010 (projects)   [010 deps: 007,009]
  └─> 023 (wire/actor_providers)

009 (archive surface) ── deps: 002
  └─> 010 (projects)

012 (ai-companion) ── deps: 005 ──> 023 (wire/companion_providers)

023 (wire/DI) ── deps: 001,005,007,008,009,012  (se trae POR MÓDULO)

025 (be-test-coverage) ── deps: 001,002,009  (follow-up)
```

Features sin dependencias BE internas: **004, 005, 011, 013, 018(parcial), 019, 020, 021, 022, 024, 027**.

---

## 3. Orden recomendado — GLOBAL (cross-repo)

| # | Feature | Repo | merge | Por qué en esta posición |
|---|---------|------|-------|--------------------------|
| 1 | 001 be-platform-tenancy-refactor | BE | independiente | Base de todo el BE. Refactor interno SIN cambio de contrato API (drop `MaybeTenantScope` → `tenancy.Scope` en ~23 repos). Todo lo demás compila sobre esto. |
| 2 | 004 shared-text-propername | BE | independiente | Util chico, sin deps. Bloquea 007. Entra temprano para no ser cuello de botella. |
| 3 | 005 be-config-modularization | BE | independiente | `cmd/config` split + `.env.example`. Funda 012 y 023. Toca `cmd/config/loadconfig.go` (compartido) — mejor antes que los módulos que dependen de wire. |
| 4 | 006 fe-design-system | FE | independiente | Base de TODO el FE. No bloquea BE pero habilita 007(FE)/008(FE)/014–018. Arrancar en paralelo al tren BE. |
| 5 | 019 be-local-tooling-db-scripts | BE | independiente | Scripts + Makefile. Bajo riesgo, sin runtime. Útil tener los scripts (reset-local-db, data-audit, lint-tenant-leaks) ANTES de las migraciones de 002/003. |
| 6 | 027 be-cleanup-domain-purity | BE | independiente | dep 001. Cleanup (staticcheck, remove core/governance, borrar jwt legacy). Mejor sobre tenancy ya migrado, antes de apilar módulos. |
| 7 | 002 be-crudar-lifecycle-framework | BE | BE-first | `internal/shared/lifecycle` + migraciones 227/228/232/233. Funda 009. |
| 8 | 003 be-multitenant-db-hardening | BE | BE-first | dep 001. Migraciones 224/225 (backfill → constraints). **Riesgo si hay datos stale** → ver §6. |
| 9 | 007 actor-system | BE→FE | BE-first | deps 001,002,003,004,006. La feature grande full-stack. BE expone `/api/v1/actors` (+migr 223/226/231/234). FE (`useActors`, master-data/actors, BFF) DESPUÉS del BE. |
| 10 | 008 identity-tenant-context | BE→FE | BE-first | dep 007. BE `/me` con array de tenants; FE TenantContext + Navbar switcher + login + BFF `me.ts`/authMiddleware. |
| 11 | 009 crudar-archive-surface | BE→FE | BE-first | dep 002. **CONTRATO**: `DELETE /:id` → `POST /:id/archive` + `DELETE /:id/hard` + `GET /archived` en ~20 dominios (123 archivos). Cara FE vive en 014 (pages) y 006 (ArchivedListPage). Sugerir **PRs por-entidad**. |
| 12 | 010 projects | BE→FE | BE-first | deps 007,009. BE project-archive-bridge + scope/creator; FE pages/admin/projects + BFF projects.ts. |
| 13 | 011 campaign-dto-projectid | BE+FE | **coordinado** | Shape change: BE serializa `project_id`/`id`/`name` en minúscula; FE campaigns. Si desync → dropdown de campañas vacío. Mergear BE y FE **juntos** o en ventana corta. |
| 14 | 012 ai-companion-integration | BE→FE | BE-first | dep 005. BE `internal/axis` (cliente Companion + JWT) + ai adapter + `companion_providers`; FE pages/admin/ai + BFF ai.ts. |
| 15 | 013 be-csv-export | BE | BE-first | `internal/shared/csvexport` + csv-service por dominio; **borra excel**. Endpoints export cambian XLSX→CSV → revisar consumo FE antes de mergear. |
| 16 | 023 be-wire-di | BE | acompaña a su módulo | deps 001,005,007,008,009,012. `wire/` + `cmd/api`. NO es un PR único: traer cada provider CON su módulo (actor_providers→07, companion_providers→12). Ver §5. |
| 17 | 018 data-integrity-admin | BE+FE | coordinado | FE pages/admin/data-integrity + useDatabase; BE `internal/data-integrity`. **EXCLUIR tentative-prices** (ya DONE #121). |
| 18 | 014 fe-master-data-pages | FE | FE tras 007/009 | deps 006,007,009. FAMILIA (212 archivos). **1 PR por entidad** (customers/fields/lots/workorders/crops/investors/managers/labors/supplies/supply-movements/stock). lots/workorders parcialmente DONE (#104/#117). |
| 19 | 015 fe-dashboard-consolidation | FE | independiente | dep 006. pages/admin/dashboard + useDashboard. |
| 20 | 016 fe-access-notifications | FE | independiente | dep 006. pages/admin/access + notifications. |
| 21 | 017 fe-dollar-commerce-forms | FE | independiente | dep 006. pages/admin/dollar + commercialization. |
| 22 | 020 ci-workflows | BE+FE | por repo | `.github/workflows`. **Pueden romper deploy** si llegan sin el resto → mergear cuando el módulo correspondiente ya esté. |
| 23 | 021 build-and-deploy-config | BE+FE | por repo | BE Dockerfile/compose/go.mod-sum (**deps bumps go-jose/x/net YA DONE #124 → separar**); FE vite/tailwind/eslint/knip/tsconfig/lockfiles/generated client. |
| 24 | 022 lefthook-git-hooks | BE+FE | por repo | `lefthook.yml`. Local tooling, opcional. |
| 25 | 024 openapi-and-docs | BE+FE | independiente | Docs. BE openapi + CRUDAR/error-catalog + CLAUDE.md; FE docs/audit + RESPONSIVE_GUIDELINES. Sin runtime. |
| 26 | 025 be-test-coverage | BE | sigue a su módulo | deps 001,002,009. handler_test + repository_tenant_test + repository_archived_refs_test (45 archivos). **Follow-up**: validan 001/009; pueden ir después de cada módulo. |
| 27 | 026 fe-test-infra | FE | independiente | dep 006. ui/.vite-smoke + ui/e2e + api/test + api/src/mocks. |

---

## 4. Orden recomendado — SOLO ESTE REPO (BE)

Secuencia ejecutable de PRs hacia `develop` en ponti-backend, omitiendo lo FE:

```
Fase A — base / bajo riesgo (en paralelo entre sí):
  PR-001  001 tenancy-refactor        (primero; todo compila sobre esto)
  PR-004  004 text-propername         (paralelo a 001)
  PR-005  005 config-modularization   (paralelo; cuidado cmd/config/loadconfig.go)
  PR-019  019 local-tooling-db-scripts(paralelo; scripts antes de migraciones)
  PR-027  027 cleanup-domain-purity   (DESPUÉS de 001)

Fase B — fundaciones BE (BE-first, secuencial):
  PR-002  002 crudar-lifecycle        (funda 009)
  PR-003  003 multitenant-db-hardening(DESPUÉS de 001; riesgo datos stale)

Fase C — contratos full-stack BE-first (secuencial estricto):
  PR-007  007 actor-system            (deps 001,002,003,004)  → avisar a FE
  PR-008  008 identity-tenant-context (dep 007)               → avisar a FE
  PR-009  009 archive-surface         (dep 002; PRs POR ENTIDAD) → avisar a FE
  PR-010  010 projects                (deps 007,009)          → avisar a FE

Fase D — shape sensible / módulos independientes:
  PR-011  011 campaign-dto-projectid  (COORDINAR con FE; shape change)
  PR-012  012 ai-companion            (dep 005)
  PR-013  013 csv-export              (revisar consumo FE de exports)
  PR-018  018 data-integrity (BE)     (excluir tentative-prices #121)

Fase E — DI / infra / docs / tests (cierre):
  PR-023  023 wire-di                 (NO único; ver §5 — va POR módulo con C/D)
  PR-020  020 ci-workflows (BE)
  PR-021  021 build-deploy (BE)       (excluir deps bumps #124)
  PR-022  022 lefthook (BE)
  PR-024  024 openapi-and-docs (BE)
  PR-025  025 test-coverage (BE)      (follow-up de 001/002/009)
```

> **Nota sobre 023 (wire):** en la práctica no es un PR de fase E. `wire/wire.go`, `wire/wire_gen.go` y `cmd/api/main.go` son archivos **mezclados**. Cada vez que entra un módulo nuevo (007, 008, 009, 010, 012) hay que traer sus providers con `git restore -p` JUNTO al módulo, regenerar wire localmente y commitear. La "fase E / PR-023" representa el residuo de wiring que no pertenezca a ningún módulo concreto.

---

## 5. Archivos compartidos / peligrosos — no mezclar a ciegas

Estos archivos son tocados por varias features. **Nunca hacer un `git restore` masivo de ellos**; traer por hunks (`-p`) o reconstruir:

| Archivo | Lo tocan | Manejo |
|---------|----------|--------|
| `wire/wire.go`, `wire/wire_gen.go` | 007, 008, 009, 010, 012, 023 | Traer providers por módulo; **regenerar** wire_gen, no copiar. |
| `cmd/api/main.go`, `cmd/api/http_server.go` | 007, 008, 010, 012, 023 | `restore -p` por bloque de rutas/inicialización; verificar que arranca. |
| `cmd/config/loadconfig.go` | 005, 012, 023 | Base = 005. Companion config (012) se agrega encima. |
| `go.mod`, `go.sum` | 012 (axis/jwt), 021 | **deps bumps go-jose/x/net YA DONE #124** → no re-traer. Solo agregar deps nuevas reales (Companion). |
| `Makefile` | 019, 020 | Base = 019; targets de CI (020) encima. |
| `internal/shared/handlers/**` | 009, 013 | archive-surface (009) y csvexport (013) tocan handlers compartidos → secuenciar. |
| `internal/shared/models/base.go` | 001, 002 | tenancy (001) + lifecycle (002) → 001 primero. |
| `internal/shared/repository/**` | 001, 009 | tenancy.Scope (001) + archived refs (009) → 001 primero. |

**Migraciones (orden numérico = orden de aplicación):**
- 224/225 → feature 003 (hardening)
- 223/226/231/234 → feature 007 (actors)
- 227/228/232/233 → feature 002 (lifecycle)

Si se mergea 002 y 003 cerca, verificar que la numeración no colisione y que 224/225 (003) corran sobre el lifecycle de 002 si comparten tablas. Aplicar **002 antes que 007** (007 depende de lifecycle) y **003 antes que 007** (007 depende del hardening de tenancy).

---

## 6. Qué validar después de cada PR

- **001:** `go build ./...` + suite de repos; grep de que no quede `MaybeTenantScope`. Smoke de un endpoint tenant-scoped.
- **002:** migraciones 227/228/232/233 aplican limpio; lifecycle compila; un dominio archiva vía framework.
- **003:** correr backfill (224) en DB con datos reales ANTES de aplicar constraints (225). **Si hay filas stale sin tenant**, 225 falla → limpiar primero (usar scripts de 019: `data-audit`, `lint-tenant-leaks`).
- **004:** test unit del util propername.
- **005:** la app levanta con el nuevo `cmd/config`; `.env.example` cubre todas las claves; nada hardcodeado roto.
- **007:** `GET /api/v1/actors` responde; migraciones 223/226/231/234 ok; wire regenerado compila; **avisar a FE que el contrato está disponible**.
- **008:** `/me` devuelve array de tenants; switcher de tenant no rompe; coordinar con BFF (`me.ts`/authMiddleware) en FE.
- **009:** por cada entidad migrada — `POST /:id/archive`, `DELETE /:id/hard`, `GET /archived` responden; el viejo `DELETE /:id` ya no existe (o redirige). **Avisar a FE** (consume en 014/006).
- **010:** crear/scoping de projects; bridge project-archive-entidades; FE projects detrás.
- **011:** **coordinado** — desplegar BE y FE juntos; verificar que el dropdown de campañas se llena (`project_id`/`id`/`name` minúscula).
- **012:** cliente Companion conecta con JWT; `companion_providers` en wire; smoke `smoke-companion` (script de 019).
- **013:** endpoints export devuelven CSV (no XLSX); **confirmar que el FE ya no espera Excel** antes de mergear.
- **018:** páginas data-integrity contra `internal/data-integrity`; **NO incluir tentative-prices** (#121 ya en develop).
- **020:** los workflows corren en un PR de prueba sin romper deploy.
- **021:** build de imagen + compose levantan; `go.mod` sin re-bajar go-jose/x/net (ya en #124).
- **025:** los tests pasan y cubren 001/009.

---

## 7. Qué va en paralelo

- **Fase A entre sí:** 001 / 004 / 005 / 019 pueden abrirse en paralelo (027 espera a 001). En FE, 006 arranca en paralelo a toda la fase A.
- **Tras 006 (FE):** 015, 016, 017 son FE-independientes y pueden ir en paralelo entre sí (cada uno dep solo 006).
- **Docs (024)** y **tooling (022)** pueden avanzar en cualquier momento, sin bloquear.

## 8. Qué NO mezclar (PRs separados)

- **009 (archive-surface):** NO un PR monolítico de 123 archivos. **Un PR por entidad/dominio** (~20). Igual para su cara FE en 014.
- **014 (master-data):** NO un PR de 212 archivos. **Un PR por entidad** (customers, fields, lots, workorders, crops, investors, managers, labors, supplies, supply-movements, stock). lots/workorders ya parcialmente DONE.
- **023 (wire):** NO traer `wire_gen.go` de la rama vieja de un saque; regenerar por módulo.
- **021:** separar los deps bumps (go-jose/x/net, #124 — ya en develop) del resto del build config.
- **018:** separar tentative-prices (#121, ya DONE) del resto de data-integrity.

## 9. Qué postergar (follow-up)

- **025 (be-test-coverage):** sigue a sus módulos (001/002/009); no bloquea funcionalidad.
- **026 (fe-test-infra):** infra de tests FE, tras 006.
- **020/022 (CI/lefthook):** mergear al final por repo, cuando el código que validan ya esté en develop (si llegan antes, rompen deploy o fallan hooks sobre código ausente).
- **024 (docs):** independiente; idealmente al cierre para reflejar el estado final de contratos.

## 10. BE-first vs FE-first vs coordinado

- **BE-first (BE merge → FE consume):** 007, 008, 009, 010, 012, 013. El backend publica/cambia el contrato; el frontend lo consume después. Avisar a FE en cada uno.
- **Coordinado (mergear casi simultáneo):** 011 (shape change que rompe el dropdown si hay desync) y 018 (BE+FE de data-integrity).
- **FE independiente (no esperan BE):** 006, 015, 016, 017, 026 (solo dependen del design system 006).
- **Por repo (mismo nombre, dos lados):** 020, 021, 022 — cada repo mergea su parte; coordinar timing para no romper deploy.

---

## 11. Excluir de la extracción (ya en `develop` vía #104/#117/#121/#124)

- `table-select-filters` → FE #104.
- `reports-dark-mode` → FE #105 (la limpieza de json-tags del **dominio BE NO está porteada** → va en **027**).
- `lot-metrics` / `total_tons` → FE+BE #117/#121/#124.
- `tentative-prices` → FE+BE #121/#124 → **excluir de 018**.
- `dependency-bumps` (go-jose, x/net) → BE #124 → **excluir de 021**.

---

## 12. Recordatorios operativos

- Toda extracción parte de `777e5f6a` (`develop-problematico~1`), **nunca** del tip `66f2b602`.
- Los comandos `git` que aparezcan en docs de paquetes son **sugerencias**, no se ejecutan acá.
- Verificar `go build ./...` y migraciones tras cada PR de fase B/C.
- Regenerar `wire_gen.go` localmente; no copiar el de la rama vieja.
