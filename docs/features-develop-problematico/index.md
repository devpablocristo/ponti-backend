# Análisis de descomposición de `develop-problematico` — Índice global

## Resumen ejecutivo

`develop-problematico` fue una rama de integración que acumuló varias líneas de
trabajo en paralelo (new-cns3 + projects + admin + AI companion + tooling + docs)
sin una estrategia de merge incremental. Su **último commit es un `restore` que
la vacía**, por lo que el tip de la rama NO sirve como fuente. La **fuente real
de extracción es el pico de la rama**: `develop-problematico~1`.

Este documento es el índice maestro para descomponer ese pico en **27 features
globales** y portarlas, de forma ordenada y por dependencias, sobre la rama base
**`develop`**. Cada feature con cambios en este repo tiene su propia carpeta
`feature-XXX-slug/` con el detalle (file-list, plan de extracción, riesgos y
sugerencias de PR). Los comandos `git` que aparezcan en los docs son
**sugerencias**, no se ejecutan desde acá.

Estrategia general:
- **BE-first**: el backend funda el contrato (tenancy, lifecycle, archive surface,
  actors). El FE se portea después y contra esos contratos.
- **Por dependencias**: 001 (tenancy) es la base de todo el BE; 006 (design system)
  es la base de todo el FE. 007 (actor-system) es la feature full-stack grande.
- **Excluir lo ya DONE**: lot-metrics/total_tons (#117/#121/#124), tentative-prices
  (#121/#124, sacar de 018), dependency-bumps go-jose/x/net (#124, sacar de 021),
  table-select-filters (#104), reports-dark-mode FE (#105).

## Datos del análisis

| Campo | Valor |
|---|---|
| Repo | Backend Go — `ponti-backend` (`core`) |
| Path | `/home/pablocristo/Proyectos/pablo/ponti/core` |
| Rama base (destino) | `develop` |
| Rama problemática | `develop-problematico` (tip = restore que la vacía; **NO usar**) |
| Fuente real de extracción | `develop-problematico~1` — SHA `777e5f6a` |
| Rango fuente | `0972e565..777e5f6a` |
| Fecha del análisis | 2026-05-30 |
| Archivos cambiados en este repo (BE) | **427** |
| Features globales (full-stack) | **27** |
| Features con cambios en **este repo (BE)** | **21** (001-005, 007-013, 018-025, 027) |
| Features sin cambios en este repo (solo FE) | 6 (006, 014, 015, 016, 017, 026) |

> Las features marcadas `en-este-repo=no` (FE puro) se documentan en el repo FE
> (`fe`/web). Acá se listan en la tabla para mantener la numeración global y las
> dependencias coherentes, pero **no tienen carpeta `feature-XXX/` en este repo**.

## Tabla de features

Leyenda de columnas:
- **Tamaño**: peso relativo *en este repo* (BE). `—` = sin cambios BE.
- **Listo p/extraer**: `sí` / `parcial` / `no` (depende de prerequisitos o de
  separar lo ya DONE).
- **Orden**: orden recomendado de extracción/merge (BE-first y por deps).

| ID | Nombre | Tipo | FE | BE | Estado | Tamaño (este repo) | Riesgo | Deps | Listo p/extraer | Orden | Carpeta |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 001 | be-platform-tenancy-refactor | refactor | no | sí | pendiente | grande (~23 repos) | medio | — | sí | 1 | [feature-001-be-platform-tenancy-refactor/](feature-001-be-platform-tenancy-refactor/) |
| 002 | be-crudar-lifecycle-framework | refactor | no | sí | pendiente | medio (migr 227/228/232/233) | medio | — | sí | 2 | [feature-002-be-crudar-lifecycle-framework/](feature-002-be-crudar-lifecycle-framework/) |
| 003 | be-multitenant-db-hardening | migration | no | sí | pendiente | chico (migr 224/225) | **alto** (datos stale) | 001 | sí | 3 | [feature-003-be-multitenant-db-hardening/](feature-003-be-multitenant-db-hardening/) |
| 004 | shared-text-propername | feature | no | sí | pendiente | chico (util) | bajo | — | sí | 4 | [feature-004-shared-text-propername/](feature-004-shared-text-propername/) |
| 005 | be-config-modularization | infra | no | sí | pendiente | chico (cmd/config + .env) | bajo | — | sí | 5 | [feature-005-be-config-modularization/](feature-005-be-config-modularization/) |
| 006 | fe-design-system | refactor | sí | no | pendiente | — (FE) | medio | — | sí (en FE) | (FE base) | — (repo FE) |
| 007 | actor-system | feature | sí | sí | pendiente | **grande** (migr 223/226/231/234) | alto | 001,002,003,004,006 | sí (BE-first) | 6 | [feature-007-actor-system/](feature-007-actor-system/) |
| 008 | identity-tenant-context | feature | sí | sí | pendiente | medio (admin me_context) | medio | 007 | sí (BE-first) | 7 | [feature-008-identity-tenant-context/](feature-008-identity-tenant-context/) |
| 009 | crudar-archive-surface | refactor | no | sí | pendiente | **grande** (123 archivos, ~20 dominios) | alto (cambio de contrato) | 002 | sí (PRs por-entidad) | 8 | [feature-009-crudar-archive-surface/](feature-009-crudar-archive-surface/) |
| 010 | projects | feature | sí | sí | pendiente | medio | medio | 007,009 | sí (BE-first) | 9 | [feature-010-projects/](feature-010-projects/) |
| 011 | campaign-dto-projectid | bugfix | sí | sí | pendiente | chico (shape change) | medio (desync FE/BE) | — | sí (coordinado) | 10 | [feature-011-campaign-dto-projectid/](feature-011-campaign-dto-projectid/) |
| 012 | ai-companion-integration | feature | sí | sí | pendiente | medio (internal/axis) | medio | 005 | sí (BE-first) | 11 | [feature-012-ai-companion-integration/](feature-012-ai-companion-integration/) |
| 013 | be-csv-export | refactor | no | sí | pendiente | medio (csvexport) | medio (XLSX→CSV, consumo FE) | — | sí (BE-first) | 12 | [feature-013-be-csv-export/](feature-013-be-csv-export/) |
| 014 | fe-master-data-pages | feature | sí | no | parcial (lots/workorders #104/#117) | — (FE, 212 archivos) | medio | 006,007,009 | parcial (en FE, 1 PR/entidad) | (FE) | — (repo FE) |
| 015 | fe-dashboard-consolidation | refactor | sí | no | pendiente | — (FE) | bajo | 006 | sí (en FE) | (FE) | — (repo FE) |
| 016 | fe-access-notifications | refactor | sí | no | pendiente | — (FE) | bajo | 006 | sí (en FE) | (FE) | — (repo FE) |
| 017 | fe-dollar-commerce-forms | feature | sí | no | pendiente | — (FE) | bajo | 006 | sí (en FE) | (FE) | — (repo FE) |
| 018 | data-integrity-admin | feature | sí | sí | parcial (tentative-prices DONE #121) | chico (internal/data-integrity) | medio | — | parcial (excluir tentative-prices) | 13 | [feature-018-data-integrity-admin/](feature-018-data-integrity-admin/) |
| 019 | be-local-tooling-db-scripts | infra | no | sí | pendiente | medio (scripts + Makefile) | bajo | — | sí | 14 | [feature-019-be-local-tooling-db-scripts/](feature-019-be-local-tooling-db-scripts/) |
| 020 | ci-workflows | infra | sí | sí | pendiente | chico (.github/workflows) | medio (deploy) | — | sí (por repo) | 15 | [feature-020-ci-workflows/](feature-020-ci-workflows/) |
| 021 | build-and-deploy-config | config | sí | sí | parcial (deps bumps DONE #124) | chico (Dockerfile/compose/go.mod) | medio | — | parcial (excluir go-jose/x/net) | 16 | [feature-021-build-and-deploy-config/](feature-021-build-and-deploy-config/) |
| 022 | lefthook-git-hooks | config | sí | sí | pendiente | chico (lefthook.yml) | bajo | — | sí (opcional, por repo) | 17 | [feature-022-lefthook-git-hooks/](feature-022-lefthook-git-hooks/) |
| 023 | be-wire-di | infra | no | sí | pendiente | medio (wire/ + cmd/api, MEZCLADOS) | alto (archivos mezclados) | 001,005,007,008,009,012 | sí (acompaña a su módulo, `restore -p`) | 18 | [feature-023-be-wire-di/](feature-023-be-wire-di/) |
| 024 | openapi-and-docs | docs | sí | sí | pendiente | medio (docs/openapi + CRUDAR) | bajo | — | sí (independiente) | 19 | [feature-024-openapi-and-docs/](feature-024-openapi-and-docs/) |
| 025 | be-test-coverage | tests | no | sí | pendiente | medio (45 archivos test) | bajo | 001,002,009 | sí (follow-up de su módulo) | 20 | [feature-025-be-test-coverage/](feature-025-be-test-coverage/) |
| 026 | fe-test-infra | tests | sí | no | pendiente | — (FE) | bajo | 006 | sí (en FE) | (FE) | — (repo FE) |
| 027 | be-cleanup-domain-purity | cleanup | no | sí | pendiente | medio (staticcheck + json-tag removal) | bajo | 001 | sí | 21 | [feature-027-be-cleanup-domain-purity/](feature-027-be-cleanup-domain-purity/) |

### Items DONE (ya en `develop`, sin paquete — excluir de la extracción)

| Item | Dónde aterrizó | Nota |
|---|---|---|
| table-select-filters | FE #104 | filtros de tabla |
| reports-dark-mode | FE #105 | la limpieza de json-tags del dominio **BE NO** está porteada → va en 027 |
| lot-metrics / total_tons | FE+BE #117/#121/#124 | — |
| tentative-prices | FE+BE #121/#124 | **excluir de 018** |
| dependency-bumps (go-jose, x/net) | BE #124 | **excluir de 021** |

## Archivos compartidos / peligrosos en este repo

Tratar con cuidado (no traer enteros; `git restore -p` junto a cada módulo):
`wire/wire.go`, `wire/wire_gen.go`, `cmd/api/main.go`, `cmd/api/http_server.go`,
`cmd/config/loadconfig.go`, `go.mod`, `go.sum`, `Makefile`,
`internal/shared/handlers/**`, `internal/shared/models/base.go`,
`internal/shared/repository/**`.

## Checklist de progreso por feature

> Marcar a medida que avanza la extracción. `dp` = `develop-problematico`.
> Las features FE-only (006, 014, 015, 016, 017, 026) se trackean en el repo FE.

### Backend (este repo)

- **feature-001 be-platform-tenancy-refactor** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-002 be-crudar-lifecycle-framework** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-003 be-multitenant-db-hardening** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-004 shared-text-propername** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-005 be-config-modularization** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-007 actor-system** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-008 identity-tenant-context** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-009 crudar-archive-surface** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-010 projects** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-011 campaign-dto-projectid** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-012 ai-companion-integration** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-013 be-csv-export** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-018 data-integrity-admin** (excluir tentative-prices #121) — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-019 be-local-tooling-db-scripts** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-020 ci-workflows** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-021 build-and-deploy-config** (excluir go-jose/x/net #124) — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-022 lefthook-git-hooks** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-023 be-wire-di** (archivos MEZCLADOS, `restore -p`) — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-024 openapi-and-docs** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-025 be-test-coverage** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-027 be-cleanup-domain-purity** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada

### Frontend (repo FE — sin carpeta en este repo)

- **feature-006 fe-design-system** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-014 fe-master-data-pages** (1 PR por entidad; lots/workorders parcial #104/#117) — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-015 fe-dashboard-consolidation** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-016 fe-access-notifications** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-017 fe-dollar-commerce-forms** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
- **feature-026 fe-test-infra** — [ ] documentada · [ ] validada · [ ] extraída · [ ] PR creado · [ ] PR mergeado · [ ] dp actualizada/descartada
