---
description: Recupera UNA feature de develop-problematico como spec definitivo en docs/specs/features/ (re-baselineado vs develop), sin escribir código de la app.
argument-hint: <id o slug de la feature, ej. 001  |  crudar-archive-surface>
allowed-tools: Bash(git*), Bash(ls*), Bash(find*), Read, Grep, Glob, Write, Edit, Agent
---

Sos el ejecutor del proceso de **recuperación controlada** de `develop-problematico`, feature por
feature, **a nivel SPEC** (NO se implementa código). Recuperás la feature indicada en `$ARGUMENTS`.

## Objetivo de esta corrida

Tomar UNA feature documentada en `docs/features-develop-problematico/`, analizarla contra el estado
REAL de `develop`, y **destilar un único spec definitivo** en `docs/specs/features/<slug>.md`. Luego
**borrar** su carpeta del backlog para trackear que ya fue tratada. **No** escribís código de la app,
**no** creás ramas, **no** corrés migraciones, **no** mergeás.

## Contexto fijo (verificado — NO re-derivar)

- Repo BE: `/home/pablocristo/Proyectos/pablo/ponti/core`. Destino: rama `develop` (tip `003a9b8f`).
- Fuente del trabajo: `777e5f6a` (= `develop-problematico~1`, el "pico"; el tip `66f2b602` es un restore vacío).
  Rango fuente: `0972e565..777e5f6a` (427 archivos). Toda extracción/lectura parte de `777e5f6a`, nunca del tip.
- `develop` YA está 100% en `platform/*` (0 imports `devpablocristo/core`, 9 módulos platform) y **no** tiene
  `MaybeTenantScope` ni `tenancy.Scope` (scoping ad-hoc). Re-baselinear SIEMPRE contra ESTE `develop`,
  no contra el baseline viejo de los docs (fechados 2026-05-30).
- **DONE en `develop` (excluir, no re-portear):** lot-metrics (#117/#121/#124), tentative-prices (#121),
  dependency-bumps go-jose/x-net (#124), `internal/reviewproxy/`.
- **Migraciones:** `develop` tiene 229/230; el source tiene 223–234. Riesgo: `golang-migrate` no aplica
  números menores a la versión actual (224/225/227/228 < 229/230 ya aplicadas).
- **Naming de specs (fijo):** archivo suelto, sin número, sin sufijo `-spec`: `docs/specs/features/<slug>.md`.
  `<slug>` = nombre de la carpeta de la feature **sin** el prefijo `feature-<NNN>-` **ni** un `be-` inicial
  (ej.: `feature-001-be-platform-tenancy-refactor` → `platform-tenancy-refactor`;
  `feature-009-crudar-archive-surface` → `crudar-archive-surface`).

## Pasos

1. **Resolver la feature.** Interpretá `$ARGUMENTS` como id (`001`) o slug. Encontrá la carpeta exacta:
   `ls -d docs/features-develop-problematico/feature-*<arg>* 2>/dev/null`. Si hay 0 o >1 coincidencias,
   listá las opciones y pedí precisión (no adivines). Si la carpeta ya no existe pero el spec ya está en
   `docs/specs/features/`, avisá que la feature ya fue tratada y frená.

2. **Leer todo lo de esa feature.** `spec.md`, `file-list.md`, `dependencies.md`, `risks.md`,
   `validation.md`, `extraction-plan.md`, `notes-for-future-agent.md`, `implementation-status.md`, y si
   existe `ANALYSIS-vs-develop.md` (re-baseline ya hecho — usalo como base). Consultá los orquestadores
   (`index.md`, `global-summary.md`, `dependency-map.md`, `shared-files.md`, `cross-repo-map.md`) según
   haga falta.

3. **Re-baselinear contra `develop` REAL (solo lectura).** Verificá afirmación por afirmación con
   `git -C <repo> show develop:<path>`, `git -C <repo> grep <pat> develop -- <glob>`,
   `git -C <repo> ls-tree -r --name-only develop -- <dir>`, y comparando con `777e5f6a`. Determiná:
   - **Estado en develop:** ¿la feature ya está (total/parcial)? ¿qué ya existe y qué falta?
   - **Diff REAL a aplicar** (no el que asume el spec viejo).
   - **Solape con DONE** (excluir lo ya porteado).
   - **Archivos:** whole-file vs partial-hunks (`git restore -p`), mezclados con otras features, qué excluir.
   - **Migraciones** (si aplica) y riesgo de numeración.
   - **Dependencias** (compile vs runtime) y qué desbloquea.
   - Para features grandes/entrelazadas, podés lanzar 2–4 subagentes **Explore** (solo lectura) en paralelo
     para inventario / modelo / build / deps, y sintetizar. Para features chicas, hacelo directo.

4. **Escribir el spec definitivo** en `docs/specs/features/<slug>.md`, con estas secciones (molde:
   `docs/specs/scripts/reset-local-db-from-prod.md`):
   1. **Propósito** (1 frase).
   2. **Estado vs `develop`** — el diff real re-baselineado: qué ya está, qué falta.
   3. **Alcance / archivos** — whole-file vs partial-hunks, compartidos/mezclados, qué excluir.
   4. **Migraciones** (si aplica) + riesgo de numeración.
   5. **Dependencias** — compile vs runtime; qué debe ir antes y qué desbloquea.
   6. **Plan de implementación** — pasos concretos, sin ejecutar.
   7. **Validación** — build / tests / migraciones / smoke.
   8. **Riesgos y decisiones pendientes.**
   Cada afirmación de hecho debe ser verificable (incluí el comando/evidencia clave cuando aporte).

5. **Tracking — borrar del backlog.** Una vez escrito el spec, `git rm -r
   docs/features-develop-problematico/feature-<NNN>-<...>/` (la carpeta de ESA feature). NO toques las
   carpetas de otras features ni los docs orquestadores. (Si el usuario prefiere conservar el backlog,
   que lo diga; por defecto se borra para reflejar "ya tratado".)

6. **Reportar:** ruta del spec creado, resumen del diff real (qué falta vs qué ya está), lo excluido (DONE),
   riesgos/decisiones, y confirmación del `git rm`. Recordá que NO se implementó código.

## Reglas duras

- **Solo lectura** sobre el código de la app; las únicas escrituras permitidas son: crear
  `docs/specs/features/<slug>.md` y borrar la carpeta tratada de `docs/features-develop-problematico/`.
- **No** crear ramas, **no** correr migraciones, **no** mergear, **no** modificar `internal/**`, `cmd/**`,
  `go.mod`, etc.
- Re-baselinear SIEMPRE contra `develop` actual; marcar como STALE lo que el spec viejo asuma de más.
- Si algo es ambiguo (slug, alcance, qué excluir), preguntar antes de escribir.
