# Auditoría: specs de `develop-problematico` vs. estado real de `develop`

> **Fecha auditoría:** 2026-06-01 · **Auditor:** revisión asistida (33 validadores read-only + re-verificación manual de los hallazgos que cambian alcance).
> **Repos:** BE `core` (`/home/pablocristo/Proyectos/pablo/ponti/core`) · FE `web` (`/home/pablocristo/Proyectos/pablo/ponti/web`).
> **Destino BE:** `develop` tip `003a9b8f` (PR #124, 2026-05-29). **Fuente:** `develop-problematico~1` = `777e5f6a`. **Destino FE:** `web:develop` tip `8c25e88` (PR #120, 2026-05-29).
> **Naturaleza:** 100% solo lectura. Este documento NO modifica los specs; solo reporta. Los specs fueron escritos el **2026-05-30**.

---

## 0. Nota metodológica

Cada afirmación lleva evidencia (`git`/`grep`/`ls`). Los veredictos provienen de validadores automáticos **salvo los marcados “(re-verificado)”**, que confirmé con un comando independiente porque cambian el alcance de una feature o porque el propio validador se contradecía. Veredicto normalizado de cada feature (no el crudo del validador):

- **PENDIENTE · baseline OK** — la feature no está en develop y el spec describe bien el punto de partida.
- **PENDIENTE · re-enunciar baseline** — la feature no está, pero el spec asume un develop que ya cambió → el *diff real* difiere de lo que el spec narra.
- **PARCIAL** — develop ya tiene parte de la feature; el spec sobre-estima el trabajo restante.

---

## 1. Resumen ejecutivo

1. **Los specs son sólidos y coherentes entre sí, pero su baseline global quedó parcialmente desactualizado.** El análisis se fechó 2026-05-30 sobre `777e5f6a`; `develop` ya había avanzado (PR #124, 2026-05-29).
2. **La “migración core/\* → platform/\*” YA ESTÁ HECHA en `develop`** (0 imports `core/*`, 93 `.go` en `platform/*`, 9 módulos). El framing de varios docs (“decidir si platform es base única… riesgo de extracción alto”) está **STALE**: la decisión ya fue tomada. *(re-verificado)*
3. **PERO el refactor de tenancy (001) NO está hecho.** `develop` no tiene `MaybeTenantScope` **ni** `tenancy.Scope` (0 y 0). Scopea tenant de forma **ad-hoc e inconsistente** (algunos repos filtran `tenant_id`, otros no). El source centraliza con `tenancy.Scope`. → El spec de 001 debe **re-enunciarse**: el delta real no es “reemplazar MaybeTenantScope” (no existe) sino “**introducir** `tenancy.Scope` + 2 módulos `platform/{observability,persistence/gorm}`/go + refactorizar ~23 repos”. *(re-verificado)*
4. **Riesgo de numeración de migraciones CONFIRMADO y es el riesgo #1 de ejecución.** `develop` tiene 229/230 aplicadas pero le faltan 223–228 y 231–234. `golang-migrate` no aplica números menores a la versión actual → 224/225/227/228 quedarían “muertas”. Requiere decisión explícita (reset/reordenar vs renumerar). *(re-verificado parcial)*
5. **Dos features ya están parcialmente en `develop`** (el spec sobre-estima): **009** (ya tiene `Archive`/`Restore`/`ListArchived`; falta solo `HardDelete`+rutas`/hard`+`runXIDAction`) *(re-verificado)* y **019** (~87% del tooling ya presente) *(re-verificado)*.
6. **Cross-repo FE: los paths del `cross-repo-map` están parcialmente equivocados.** En `web:develop`: 007 FE no existe; 008 parcial (no `TenantContext` ni `/me`); 010 está en `database/projects` (no `pages/admin/projects`); 012 está en `ai-assistant` (no `pages/admin/ai`); 018 en `database/data-integrity`. Los DONE FE (#104/#105/#117/#121) **sí** están.
7. **Lo que el análisis acertó (verificado):** `777e5f6a` es el pico (padre del restore `66f2b602`); rango = 427 archivos; los DONE BE (lot-metrics, tentative-prices, reviewproxy, bumps go-jose/x-net) están en `develop`; 229/230 en develop y no en source.

---

## 2. Invariantes de fuente/rama + DONE — todos OK

| Afirmación | Veredicto | Evidencia |
|---|---|---|
| `777e5f6a` es el pico; tip `66f2b602` es un restore vacío | **OK** | `git log develop-problematico -3` → 66f2b602(restore) ← 777e5f6a |
| Rango `0972e565..777e5f6a` = 427 archivos | **OK** | `git diff --stat` → 427 files, +40589/-14664 |
| lot-metrics, tentative-prices, reviewproxy en `develop` | **OK** | `git show develop:<path>` existe para los 3 |
| bumps go-jose/v4 v4.1.4 + x/net v0.55.0 en `develop` | **OK** | `git show develop:go.mod` ✓ |
| 229/230 en `develop`, no en source | **OK** | `ls-tree develop` ✓ / `777e5f6a` ✗ |
| #104/#105/#117/#121 en `web:develop` | **OK** | PR #120 (commits 967bd55, 43e4a3e) |

→ Las exclusiones del spec (no re-portear DONE) son **correctas**.

---

## 3. Re-baseline: specs cuyo punto de partida ya no matchea `develop`

| Feature | Qué asume el spec | Realidad de `develop` | Diff real a aplicar |
|---|---|---|---|
| **001** tenancy-refactor | “core/\* → platform/\*” + “drop `MaybeTenantScope` → `tenancy.Scope` en ~23 repos” | core→platform **ya hecho**; **no hay** `MaybeTenantScope` **ni** `tenancy.Scope` (0/0); tenant scoping ad-hoc *(re-verificado)* | **Introducir** `tenancy.Scope` (no “reemplazar”): + módulos `platform/observability/go` y `platform/persistence/gorm/go`, + refactor de los repos al scoping centralizado |
| **021** build/deploy | “migrar build core→platform” | platform en Dockerfile/compose **ya migrado** (commits 9a5e465b/e7bb89d8/93f77883); falta quitar `CORE_REPO_DIR` mount + `GOWORK=off` del compose; `go.mod/go.sum` van atados al código de 001/002/013 | Solo el delta de compose + regenerar go.mod/sum cuando entren 001/002/013. Excluir bumps (#124 DONE) |
| **024** openapi-docs | docs “nuevos” + 3 modificados (README/docs.README/ARCHITECTURE) | `develop` **ya divergió**: `ARCHITECTURE.md` reescrito, renames de tooling (`core/*→platform/*`, `staging-db-2-local-db→reset-local-db-from-prod`) ya aplicados | Los 17 docs nuevos entran limpio; los **3 modificados conflictúan** (no son hunks “propios de 024”) → `git restore -p` selectivo |
| **025** test-coverage | “45 archivos (44 A + 1 M)” | el diff real son **52 archivos test**, con 7 módulos extra no listados (actor, ai, campaign, data-integrity, lease-type, project) | Corregir el file-list; depende de 001/002/009 (no compila sin ellos) |

> **Corrección a un validador:** el agente de consistencia-de-docs afirmó “001 ya está hecho en develop” (porque `MaybeTenantScope=0`). **Es incorrecto.** `develop` tampoco tiene `tenancy.Scope`: usa scoping ad-hoc. 001 sigue **pendiente**; lo único stale es su *narrativa*.

---

## 4. Gap de tenancy/platform (transversal — el más importante)

- `develop`: **0** `core/*`, 93 `.go` en `platform/*`, **9** módulos. Faltan exactamente **2** que el source agrega: `platform/observability/go` y `platform/persistence/gorm/go` (de este último sale `tenancy.Scope`). *(re-verificado)*
- `develop` **no** tiene `internal/shared/authz` *(re-verificado)* → prerequisito ausente para **007, 008, 012, 018**.
- **Modelo de tenancy divergente:** `develop` scopea tenant **ad-hoc** (p.ej. `businessinsights` filtra `tenant_id`; `category`/`campaign` **no** validan tenant). El source centraliza con `tenancy.Scope(ctx, …, 'tabla')` en **cada** método + migraciones 224/225 (`tenant_id` → NOT NULL). → portar el modelo del source es un cambio estructural que toca ~23 repos, no un find/replace.
- Matiz sobre “100% platform”: es cierto que **no queda `core/*`**, pero `develop` **no tiene todos los módulos platform** que el source necesita (le faltan 2). Ambas cosas conviven.

---

## 5. Migraciones: mapeo y riesgo de numeración (riesgo #1)

`develop`: 0001–0222, 0229, 0230. **Hueco:** 0223–0228 y 0231–0234 (en source, no en develop).

| Migr | Feature | Nota |
|---|---|---|
| 223 | 007 | actors base (aditivo) |
| 224, 225 | 003 | `tenant_id` + FK + NOT NULL (backfill→constraints) |
| 226 | 007 | customer-actor link (requiere 224) |
| 227, 228 | 002 | archive_batches / archive metadata |
| 229, 230 | **DONE** | lot-metrics / workorders (ya en develop) |
| 231–234 | 007/002/004 | actor archived/unique (requieren 224) |

**Riesgo (ALTO, confirmado):** con la DB en versión 230, `migrate up` **no** aplica 224/225/227/228 (números menores). Quedan migraciones muertas. Caminos posibles (de los specs): **A** = portar 223–228 en orden antes de pasar de 222 (reset) · **B** = renumerar 224/225 a >230 cuidando que 226/231/234 dependen de 224. **Gap en los docs:** describen bien el riesgo pero **no** dan el procedimiento concreto (no existe un `MIGRATION_MERGE_STRATEGY.md`).

---

## 6. Matriz por feature (21 BE)

Estado real en develop · baseline del spec · nota. (✔ = re-verificado por mí)

| ID | Feature | Estado en develop | Baseline | Nota clave |
|---|---|---|---|---|
| 001 | platform-tenancy | PENDIENTE | **re-enunciar** ✔ | core→platform hecho; sin MaybeTenantScope/tenancy.Scope; delta = +2 módulos + adoptar tenancy.Scope |
| 002 | crudar-lifecycle | PENDIENTE | OK | sin `internal/shared/lifecycle`; falta prometheus/observability/gorm en go.mod; **riesgo migr 227/228** |
| 003 | multitenant-db-hardening | PENDIENTE | OK + riesgo | **riesgo migr 224/225** (hueco); dep 001 |
| 004 | shared-text-propername | PENDIENTE | OK | autocontenido, limpio; entra temprano |
| 005 | config-modularization | PENDIENTE | OK ✔ | develop aún tiene `cmd/config/ai.go` (config vieja); *(el validador dijo STALE; corregido: baseline correcto)* |
| 007 | actor-system | PENDIENTE | OK | `internal/actor` ausente; prereqs (authz/text/2 módulos) también ausentes; migr 223/226/231/234 |
| 008 | identity-tenant-context | PENDIENTE | OK | sin `/me/context` ni capa UseCases; prereq `authz` (de 007) ausente |
| 009 | crudar-archive-surface | **PARCIAL** | OK ✔ | develop **ya** tiene Archive/Restore/ListArchived; falta `HardDelete`+`/hard`+`runXIDAction`+`/archived` → **scope menor** |
| 010 | projects | PENDIENTE | OK | repository 1478→2689 líneas; prereqs 004/007/008/009 + gorm ausentes |
| 011 | campaign-dto-projectid | PENDIENTE | OK | develop en estado buggy (serializa dominio, sin `project_id`); 2 hunks reales |
| 012 | ai-companion | PENDIENTE | OK | `internal/axis` ausente; prereqs 005/authz ausentes; Nexus dead-wired (excluir) |
| 013 | csv-export | PENDIENTE | OK ✔ | develop **aún** tiene `excelize v2.9.1` + 3 archivos excel; feature = quitar XLSX + CSV *(validador dijo FALSO; era nitpick de fraseo)* |
| 018 | data-integrity-admin | PENDIENTE | OK | develop tiene los 17-controles viejos; feature = reescritura 5-control; prereqs authz + 4 RAW + wire ausentes; excluir tentative-prices (DONE) |
| 019 | local-tooling-db-scripts | **PARCIAL** | OK ✔ | ~87% ya en develop (reset-db, actors-backfill, lint-tenant-leaks); falta `export-ai-conversations.sh` y `smoke-companion` (condicionado a 012) |
| 020 | ci-workflows | PENDIENTE | OK | 6 YAML en estado viejo (`AI_SERVICE_*`); estrategia sub-PR válida |
| 021 | build-and-deploy | PENDIENTE | **re-enunciar** | platform build ya migrado; resta compose mount/GOWORK + go.mod atado a código; excluir bumps |
| 022 | lefthook-git-hooks | PENDIENTE | OK | `lefthook.yml` ausente; riesgo nulo |
| 023 | wire-di | PENDIENTE | OK | wire providers ausentes; ~30% autocontenido (cmd/migrate slog, cmd/archive-cleanup); sumidero de 001/005/007/008/009/012/013/018 |
| 024 | openapi-and-docs | PENDIENTE | **re-enunciar** | docs nuevos OK; 3 modificados conflictúan (develop ya reescribió ARCHITECTURE.md + renames) |
| 025 | test-coverage | PENDIENTE | corregir conteo | 52 archivos (no 45), 7 módulos extra; depende de 001/002/009 |
| 027 | cleanup-domain-purity | PENDIENTE | OK | json/gorm tags del dominio aún en develop |

---

## 7. Cross-repo FE (`web:develop`)

| Afirmación del `cross-repo-map` | Veredicto | Realidad |
|---|---|---|
| 006 design-system = base FE presente | **PARCIAL / file-list FALSO** | solo 6 archivos shell (router/main/ProtectedLayout/Sidebar/client/index.css); los ~127 de núcleo (Button, Drawer, feedback, lib/*, crud…) **no** están en develop → revisar si migraron a paquetes npm `@devpablocristo/platform-ui-*` o siguen pendientes |
| 014 master-data = “212 archivos, 1 PR/entidad” | **PARCIAL / file-list FALSO** | carpeta `master-data/` **no existe** en develop; “212” es aproximado e incluye hooks+BFF; faltan entidades (actors, campaigns, commerce, dollar, projects) |
| 015 dashboard, 016 access/notif, 017 dollar/commerce, 026 test-infra | **OK** | paths existen en `web:develop` (no se auditó si el *refactor* específico está aplicado — solo presencia) |
| Caras FE de 007/008/010/012/018 | **PARCIAL — paths del doc equivocados** | 007 FE **no existe** (0/3); 008 parcial (navbar sí, `TenantContext`/`/me` **no**); 010 → `pages/admin/database/projects` (no `pages/admin/projects`); 012 → `ai-assistant` (no `pages/admin/ai`); 018 → `database/data-integrity` |
| DONE FE #104/#105/#117/#121 | **OK** | todos en `web:develop` (PR #120) |

---

## 8. Inconsistencias entre los docs orquestadores

- **`global-summary.md §9.1` y `dependency-map.md §9`** tratan la migración platform como “decisión pendiente / riesgo estratégico”. **STALE:** ya está resuelta (platform es la base de develop). *(re-verificado)*
- **`dependency-map.md §1`** describe 001 como “drop `MaybeTenantScope` → `tenancy.Scope`”. **STALE:** no hay `MaybeTenantScope` que dropear; el delta es *introducir* `tenancy.Scope`.
- **`validation-strategy.md §3 (001)`**: “grep `MaybeTenantScope` debe dar 0” → ya da 0 hoy; describe un test ya cumplido (asume 001 sin hacer).
- **Niveles de estado mezclados:** los docs no distinguen “feature como PR pendiente” (correcto) de “infraestructura platform como decisión abierta” (stale). Conviene una nota global: *“Análisis fechado 2026-05-30; develop (003a9b8f) ya está 100% sobre platform/\*. Las features 001–027 se extraen SOBRE esa base; no son la migración core→platform.”*
- **Consistente y correcto:** `cross-repo-map` (001 sin cambio de contrato), `shared-files` (wire_gen.go generado), descarte de `.claude/settings.json` (alineado en shared-files y discarded-or-questionable).

---

## 9. Correcciones sugeridas a los specs (sin aplicar — “solo reportar”)

**Alto impacto**
1. **001/global-summary/dependency-map/validation-strategy:** re-enunciar el baseline platform/tenancy (core→platform ya hecho; no hay MaybeTenantScope; delta = +2 módulos + adoptar `tenancy.Scope` sobre el scoping ad-hoc actual).
2. **009:** reducir el alcance — develop ya tiene Archive/Restore/ListArchived; documentar que solo falta `HardDelete`+`/hard`+`runXIDAction`+`/archived`.
3. **019:** marcar ~87% como ya presente en develop; el delta real son 1–2 archivos.
4. **Migraciones:** crear un `MIGRATION_MERGE_STRATEGY.md` con el procedimiento concreto (Camino A reset/reorden vs B renumerar) y verificación post-merge. Hoy los docs describen el riesgo pero no el “cómo”.
5. **Cross-repo FE:** corregir los paths reales (`database/projects`, `ai-assistant`, `database/data-integrity`) y aclarar que 006/014 están mayormente pendientes / posiblemente en paquetes npm.

**Medio / menor**
6. **021:** separar “platform build ya migrado” de “delta de compose (CORE_REPO_DIR/GOWORK)”.
7. **024:** marcar los 3 archivos modificados como conflictivos (develop ya divergió); los 17 nuevos entran limpio.
8. **025:** corregir el conteo (52 vs 45) y listar los 7 módulos extra.
9. **005/013:** corregir el fraseo de estado (develop tiene la versión vieja; baseline correcto) — no cambia el trabajo, sí la claridad.
10. **Nota de fecha global** en `index.md` aclarando el desfasaje 2026-05-30 (análisis) vs develop actual.

---

*Fin de la auditoría. Cero cambios a código o a los specs; este archivo es el único artefacto nuevo.*
