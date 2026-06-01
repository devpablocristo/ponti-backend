# Validation Strategy — Extracciones de `develop-problematico`

> **Documento global.** Estrategia para validar que cada extracción (feature-XXX) desde
> `develop-problematico~1` hacia `develop` está **completa y correcta**.
> No es un paquete por feature; es el manual de validación que aplica a todos.

- **Repo:** Backend Go `ponti-backend` — `github.com/devpablocristo/ponti-backend`
- **Path:** `/home/pablocristo/Proyectos/pablo/ponti/core`
- **Go:** `go 1.26.3`
- **SOURCE de extracción:** `develop-problematico~1` (SHA `777e5f6a`) — **NUNCA** el tip (`develop-problematico`, que es un restore que vacía la rama).
- **Rango fuente total:** `0972e565..777e5f6a`
- **Destino:** `develop`
- **Fecha del análisis:** 2026-05-30

Este repo es **solo backend (BE)**. Las features marcadas `en-este-repo=no`
(006, 014, 015, 016, 017, 026 y las partes FE de las full-stack) **no se validan acá**;
su validación FE (yarn / Playwright / knip) se documenta en el repo `web` (fe). Igual se
incluye una sección FE corta para las features full-stack, porque la **validación cross-repo**
es responsabilidad compartida y el BE-first arranca acá.

---

## 0. Convención de comandos (lo verificado en este repo)

Todos los comandos asumen `cwd = /home/pablocristo/Proyectos/pablo/ponti/core`.

| Acción | Comando canónico | Origen / nota |
|---|---|---|
| Compilar todo | `go build ./...` | estándar Go |
| Tests | `make test` → `go test ./...` | verificado en `Makefile` |
| Lint | `make lint` → `golangci-lint v2.11.4 run --timeout=5m` | verificado; corre vía `go run …@v2.11.4`, no requiere binario instalado |
| Vet (barato, sin lint) | `go vet ./...` | estándar Go |
| Migraciones up (local, Docker) | `make db-migrate-up` → `scripts/db/db_migrate_up.sh` | usa `migrate/migrate:v4.17.1` sobre `migrations_v4/` |
| Reset DB local | `make db-reset` → `scripts/db/db_reset.sh` | drop/recrea schema |
| Validación de schema/datos | `make db-validate` → `scripts/db/db_validate.sh` (+ `db_validate.sql`) | invariantes de DB |
| Snapshot + diff de schema | `make db-schema-snapshot` / `make db-schema-diff` | compara contra `scripts/db/schema.expected.sql` |
| **Pipeline DB completo** | `make db-verify` | = `db-reset → db-migrate-up → db-validate → db-schema-snapshot → db-schema-diff` |
| Backfill actors (007) | `make actors-backfill-sync` → `scripts/db/actors_backfill_sync.sql` | re-corre backfill sobre datos locales |
| Smoke de release | `make e2e-changes [BASE_URL=...]` → `scripts/smoke_release.sh` | smoke E2E contra API levantada |
| Levantar API local | `make run-api` (`go run ./cmd/api/`) o `make dev` (Docker + Air hot reload) | `dev` espera `GO_MODULES_TOKEN` para bajar `platform/*` |
| Stack full local | `make up-ponti-local` / `make down-ponti-local` | core + web + axis |

**`staticcheck` (feature-027):** el `Makefile` **no** tiene target `staticcheck` propio.
golangci-lint v2 ya incluye el analizador `staticcheck` entre sus linters, así que
`make lint` cubre 027. Si se necesita staticcheck standalone:
`go run honnef.co/go/tools/cmd/staticcheck@latest ./...` — **verificar con humano** si
quieren un target dedicado.

> **Recordatorio de entorno (memoria):** en este entorno `docker exec`/`docker run` se
> "tragan" el stdout. Para ver salida de contenedores (migrate, db, api) usá `docker logs <container>`,
> no el stdout directo del `exec`/`run`.

---

## 1. Niveles de validación (de barato a caro)

Aplicar **en orden**. No pasar al siguiente nivel si el anterior falla.

### Nivel 1 — Compila (gate mínimo, ~segundos)
```bash
go build ./...
```
- Si falla con `undefined:` / `cannot find package` → extracción **incompleta**
  (faltan archivos o un `restore -p` parcial dejó referencias colgando). Ver §4.
- `go vet ./...` como complemento barato.

### Nivel 2 — Lint
```bash
make lint
```
- Es el linter de CI. Si CI usa otra versión, **verificar con humano** mirando
  `.github/workflows/` (feature-020). La versión local pineada es `v2.11.4`.

### Nivel 3 — Tests unitarios / de repo
```bash
make test            # go test ./...
```
- Para iterar sobre un dominio puntual: `go test ./internal/<dominio>/...`.
- feature-025 trae `handler_test`, `repository_tenant_test`,
  `repository_archived_refs_test` (45 archivos). Estos tests **son** parte de la
  validación de 001 y 009 — ver §3.

### Nivel 4 — DB / migraciones (las features con `migration`)
```bash
make db-verify       # reset + migrate up + validate + snapshot + diff
```
- Requiere Docker y la DB local (`docker compose … up -d ponti-db`).
- `db-schema-diff` compara contra `scripts/db/schema.expected.sql`: si una extracción
  agrega migraciones, **el `schema.expected.sql` también tiene que venir actualizado**
  o el diff dará falso-positivo. Ver §3 (feature-002/003).

### Nivel 5 — Runtime / endpoint (las features full-stack, BE-first)
```bash
make run-api         # o make dev (Docker)
# luego curl al endpoint nuevo — ver §5
```
- Único nivel que detecta DI faltante (wire) y rutas no registradas. Imprescindible
  para 007, 008, 010, 012, 018: que **compile no garantiza que la ruta exista**.

### Nivel 6 — Smoke / cross-repo
```bash
make e2e-changes [BASE_URL=http://localhost:PORT]
```
- Y, para full-stack, pegarle al endpoint nuevo **antes** de mergear el FE (§5).

---

## 2. Riesgos de integración transversales

1. **Archivos MEZCLADOS (shared/dangerous).** Estos archivos contienen líneas de
   varias features y se traen con `git restore -p` (hunk a hunk), no enteros:
   - `wire/wire.go`, `wire/wire_gen.go`, `cmd/api/main.go`, `cmd/api/http_server.go`
   - `cmd/config/loadconfig.go`, `internal/shared/handlers/**`,
     `internal/shared/models/base.go`, `internal/shared/repository/**`
   - `go.mod`, `go.sum`, `Makefile`
   - **Riesgo:** traer un hunk de más arrastra otra feature; traer uno de menos deja
     DI o rutas colgando (compila a veces, falla en runtime). Por eso §5 es obligatorio
     para features con wire.
   - `wire_gen.go` es **generado**: idealmente regenerar con `wire ./...` en vez de
     copiar hunks. **Verificar con humano** si el repo regenera wire en CI o lo commitea
     a mano (no hay target `wire` en el Makefile).

2. **`go.mod` / `go.sum`.** Los bumps de `go-jose/v4` y `x/net` **YA están en `develop`**
   (#124). NO re-traerlos desde la fuente: re-introducir versiones viejas reintroduce
   las vulnerabilidades. Tras tocar `go.mod`: `go mod tidy` y `go build ./...`.
   Excluir estos bumps de feature-021.

3. **Orden de merge (deps de la tabla).** El BE se funda en este orden:
   `001` (tenancy.Scope) → `002` (lifecycle) → `003` (db hardening, depende de 001) →
   `004/005` (utils/config) → `007` (actors, depende de 001,002,003,004) →
   `008` (identity, depende de 007) → `009` (archive surface, depende de 002) →
   `010/012` (dependen de 007/009 y 005). `023` (wire) **acompaña a su módulo**, no va sola.
   Mergear fuera de orden = build roto por símbolos inexistentes.

4. **Migraciones: numeración y secuencia.** La fuente agrega `223..234` (actors 223/226/231/234,
   crudar 227/228/232/233, hardening 224/225). En `develop` HOY conviven `229/230` (dashboard,
   workorders_is_digital_origin) que **ya están**. Riesgo de **colisión / huecos de numeración**
   y de orden incorrecto al intercalar. Validar siempre con `make db-verify` sobre DB limpia.
   Backfill **antes** de constraints (003) — ver §3.

5. **Contrato de API que cambia (009, 011, 013).**
   - 009: `DELETE /:id` → `POST /:id/archive` + `DELETE /:id/hard` + `GET /archived`
     en ~20 dominios. Si se trae a medias, hay dominios con contrato nuevo y otros con el
     viejo → el FE rompe en los que quedaron a medias.
   - 011: `project_id/id/name` pasan a minúscula. Si BE y FE quedan desincronizados,
     **el dropdown de campañas queda vacío** (síntoma observable).
   - 013: endpoints de export pasan de XLSX a CSV → revisar consumo FE antes de mergear.

6. **Cross-repo BE-first.** Para 007/008/010/011/012/018: mergear BE primero, **verificar
   el endpoint con `curl`**, y recién ahí mergear el FE. Nunca al revés.

---

## 3. Validación específica por feature (las del BE)

> Solo features con `en-este-repo=sí`. Comando base = §1.

### feature-001 — be-platform-tenancy-refactor (refactor, base de todo el BE)
- Drop `MaybeTenantScope` → `tenancy.Scope` en ~23 repos. **Sin cambio de contrato API.**
- Validación: `go build ./...` + `make test` (incluye `repository_tenant_test` de 025).
- Verificar que **no quede ninguna** referencia residual:
  `grep -rn "MaybeTenantScope" internal/ cmd/ wire/` debe dar **0 resultados**.
- `scripts/lint-tenant-leaks.sh` (feature-019) detecta fugas de tenant: correrlo si está disponible.

### feature-002 — be-crudar-lifecycle-framework (refactor; funda 009)
- `internal/shared/lifecycle` + migraciones **227/228/232/233**.
- Validación: `go build` + `make db-verify` (las 4 migraciones deben aplicar limpio sobre DB reseteada).
- Si `db-schema-diff` falla: confirmar que `scripts/db/schema.expected.sql` se trajo actualizado.

### feature-003 — be-multitenant-db-hardening (migration; dep 001)
- Migraciones **224/225**: 224 = backfill de datos, 225 = constraints.
- **CRÍTICO — orden:** backfill (224) **antes** de constraints (225). Si hay datos stale,
  225 falla con violación de constraint. Validar contra una copia con datos reales:
  `make reset-local-db-from-prod` (trae data-only de PROD) y luego `make db-migrate-up`.
  Si rompe → hay filas que no cumplen el invariante → revisar 224.
- `scripts/db/hardening_post_restore.sql` y `tenant_isolation_audit.sql` ayudan a verificar aislamiento.
- `make db-validate` debe pasar después de aplicar 224/225.

### feature-004 — shared-text-propername (util; bloquea 007)
- Util chico. `go build` + test del paquete: `go test ./internal/shared/text/...`
  (**verificar path exacto** del paquete con `grep -rl ProperName internal/shared`).

### feature-005 — be-config-modularization (infra; funda 012, 023)
- `cmd/config` split + `.env.example`. `loadconfig.go` es archivo MEZCLADO.
- Validación: `go build` + `make run-api` debe **arrancar sin panic de config**.
  Comparar `.env.example` con tu `.env` local para detectar claves nuevas requeridas
  (companion/axis para 012). Si arranca pero loguea "missing config X" → falta una key en el split.

### feature-007 — actor-system (FULL-STACK; deps 001,002,003,004,006)
- BE: `/api/v1/actors` + migraciones **223/226/231/234**. La feature grande.
- DB: `make db-verify`; luego `make actors-backfill-sync` (backfill/sync sobre datos locales).
  Validar con `scripts/db/actors_golden_master.sql` si existe golden master.
- Runtime/DI (**imprescindible**): `make run-api` y comprobar la ruta:
  ```bash
  curl -i http://localhost:PORT/api/v1/actors -H "Authorization: Bearer <token>"
  ```
  Si 404 → ruta no registrada (falta hunk en `http_server.go`).
  Si 500 con "missing provider"/nil → DI incompleto (`wire/actor_providers.go` de 023).
- Cross-repo: este endpoint debe responder antes de mergear el FE (`useActors`, BFF `actors.ts`).

### feature-008 — identity-tenant-context (FULL-STACK; dep 007)
- BE: admin `me_context` → `/me` devuelve **array de tenants**.
- Runtime: `curl /api/v1/me` (o ruta admin) y verificar que `tenants` es un array.
- Cross-repo: el switcher de tenant del FE (Navbar) depende de este shape.

### feature-009 — crudar-archive-surface (refactor; dep 002; 123 archivos)
- Contrato: `DELETE /:id` → `POST /:id/archive` + `DELETE /:id/hard` + `GET /archived`
  en ~20 dominios. **Sugerir PRs por-entidad.**
- Validación de completitud por dominio: para cada entidad migrada, las 3 rutas nuevas
  deben existir y la vieja `DELETE /:id` (soft) ya no. `repository_archived_refs_test`
  (025) valida la integridad referencial del archivado.
- Runtime por dominio (smoke):
  ```bash
  curl -i -X POST   http://localhost:PORT/api/v1/<entidad>/<id>/archive
  curl -i           http://localhost:PORT/api/v1/<entidad>/archived
  curl -i -X DELETE http://localhost:PORT/api/v1/<entidad>/<id>/hard
  ```
- **Riesgo de a-medias:** si algunos dominios tienen contrato nuevo y otros el viejo, el FE
  (014/006) rompe sólo en los inconsistentes. Llevar una checklist de dominios cubiertos.

### feature-010 — projects (FULL-STACK; deps 007,009)
- BE: project-archive-entidades-bridge + scope/creator. Depende de que 007 y 009 ya estén.
- `go build` + `make test` + runtime de `/api/v1/projects` (CRUD + archive surface heredado de 009).

### feature-011 — campaign-dto-projectid (bugfix shape; merge coordinado)
- BE serializa `project_id/id/name` en **minúscula**.
- Validación: `curl` al endpoint de campaigns y verificar **literalmente** las claves del JSON
  (`grep` por `"project_id"`, `"id"`, `"name"` en minúscula). Síntoma de desync: dropdown de
  campañas vacío en el FE. Mergear coordinado con FE.

### feature-012 — ai-companion-integration (FULL-STACK; dep 005)
- BE: `internal/axis` (cliente Companion + JWT) + ai adapter + `companion_providers`.
- Config (de 005) debe traer las keys de Companion/axis en `.env.example`.
- `scripts/smoke_companion` / `make e2e-changes` para smoke. **Verificar con humano** el script
  exacto de smoke companion si `e2e-changes` no lo cubre.
- DI: `wire/ai_providers.go` + `companion_providers`. Runtime: que arranque sin "missing provider".

### feature-013 — be-csv-export (refactor)
- `internal/shared/csvexport` + csv-service por dominio; **borra excel**.
- Validación: `grep -rn "xlsx\|excelize" internal/` debe tender a 0 en los dominios migrados.
- Runtime: endpoint de export debe responder `Content-Type: text/csv`. **Revisar consumo FE**
  (el FE espera XLSX → coordinar).

### feature-018 — data-integrity-admin (FULL-STACK; merge coordinado)
- BE: `internal/data-integrity`. **Excluir tentative-prices** (ya DONE en #121).
- Verificar que la extracción **no** re-traiga código de tentative-prices que ya está en develop
  (colisión / duplicación). `git log`/`grep` por "tentative" en lo extraído.

### feature-019 — be-local-tooling-db-scripts (infra; bajo riesgo)
- `scripts/` + `Makefile`. Validación: que los scripts existan y sean ejecutables; `make -n <target>`
  para verificar que los targets resuelven. No requiere build.

### feature-020 — ci-workflows (infra)
- `.github/workflows/`. **No romper deploy.** Validación: revisar que el workflow referencia
  comandos que existen (`make lint`, `make test`, versión de Go `1.26.3`). Si CI pinea otra versión
  de golangci-lint que la local (`v2.11.4`) → alinear. Traer sólo con el resto, no aislado.

### feature-021 — build-and-deploy-config (config)
- BE: Dockerfile / compose / go.mod-sum. **Los bumps de deps YA están (#124) → separarlos.**
- Validación: `make build` (compose build) debe completar; imagen arranca con `make up` +
  `docker logs ponti-api` (no `docker exec` stdout).

### feature-022 — lefthook-git-hooks (config; opcional)
- `lefthook.yml`. Local, opcional. Validar instalación: `lefthook install` y que los hooks corran
  `make lint`/`make test`. Bajo riesgo si se omite.

### feature-023 — be-wire-di (infra; acompaña a su módulo)
- `wire/` + `cmd/api`. `actor_providers`→007, `companion_providers`→012. Archivos MEZCLADOS.
- **No mergear sola.** Validación: `go build ./...` (debe compilar el grafo wire) + **runtime**:
  arrancar la API y comprobar que las rutas de su módulo responden (un wire incompleto compila
  pero deja providers nil → 500 en runtime). Si el repo regenera wire: correr `wire ./...`
  y diffear `wire_gen.go` (**verificar con humano** si se commitea generado o se regenera en CI).

### feature-025 — be-test-coverage (tests; sigue a su módulo; deps 001,002,009)
- 45 archivos: `handler_test`, `repository_tenant_test`, `repository_archived_refs_test`.
- Validan 001 (tenant scope) y 009 (archived refs). Pueden ir como **follow-up** tras su módulo.
- Validación: `make test` debe pasarlos. Si un test referencia un símbolo/ruta que no existe →
  el módulo que validan (001/009) se extrajo incompleto.

### feature-027 — be-cleanup-domain-purity (cleanup; dep 001)
- Remoción de json-tags del dominio en reports + remove `core/governance` + borrar jwt utils legacy.
- Validación: `go build` + `make lint` (incluye staticcheck). `grep` de los símbolos removidos
  (governance / jwt legacy) debe dar 0. Cuidado: la limpieza FE de reports-dark-mode ya está (#105)
  pero **la parte BE de json-tags no** → es ésta.

---

## 4. Cómo detectar una extracción INCOMPLETA

Síntomas y diagnóstico:

| Síntoma | Causa probable | Diagnóstico |
|---|---|---|
| `undefined: X` / `cannot find package` en `go build` | falta un archivo o un hunk | `grep -rn "X" .` y ver qué feature lo define |
| **Compila pero falta una ruta** (404 en runtime) | `http_server.go` / handler no traído | `make run-api` + `curl` a la ruta; `grep` del path en `cmd/api/http_server.go` y handlers |
| **Compila pero panic/500 "nil provider"** | wire/DI incompleto (023) | levantar API y mirar `docker logs` / stack; revisar `wire/*_providers.go` del módulo |
| Migración no aplica / huecos de numeración | migración faltante o fuera de orden | `make db-verify` sobre DB limpia; revisar numeración 223..234 vs 229/230 ya presentes |
| `db-schema-diff` falla | `schema.expected.sql` desactualizado | confirmar que se trajo el `schema.expected.sql` de la feature |
| Constraint falla al migrar (003) | backfill 224 no corrió antes de 225 | correr sobre datos reales (`reset-local-db-from-prod`) |
| Tests de 025 fallan al referenciar símbolos | el módulo (001/009) que validan vino a medias | mirar qué símbolo falta el test |
| Referencia residual a algo que debía borrarse | hunk de cleanup no aplicado | `grep` del símbolo viejo (`MaybeTenantScope`, `xlsx`, `governance`, jwt legacy) → debe ser 0 |

**Regla de oro:** *compilar es el piso, no el techo.* Para features con rutas o DI
(007, 008, 009, 010, 011, 012, 018, 023) la validación **no está completa sin runtime + curl**.

---

## 5. Validación cross-repo (BE-first) — full-stack

Features full-stack en este repo: **007, 008, 010, 011, 012, 013(consumo), 018**.

Procedimiento obligatorio (BE primero, FE después):

1. Mergear/validar el BE en `develop` (niveles 1–5 de §1).
2. Levantar la API: `make run-api` (o `make up-ponti-local` para el stack completo core+web+axis).
3. **Pegarle al endpoint nuevo con `curl`** y verificar status + shape del JSON **antes** de tocar el FE:
   - 007 → `GET /api/v1/actors` responde 200 con lista.
   - 008 → `GET /me` devuelve `tenants` como **array**.
   - 010 → `GET /api/v1/projects` (+ archive surface).
   - 011 → claves `project_id/id/name` en **minúscula** (si no, dropdown de campañas vacío).
   - 012 → endpoints `ai`/companion responden; arranca sin "missing provider".
   - 013 → export responde `Content-Type: text/csv` (no XLSX).
   - 018 → `data-integrity` sin re-traer tentative-prices (ya en develop).
4. Recién con el endpoint verde, validar el FE en el **repo web** (yarn/Playwright/knip — §6).
5. Smoke conjunto: `make e2e-changes [BASE_URL=...]` y/o `make up-ponti-local`.

> Nunca mergear el FE primero: el FE contra un endpoint inexistente da errores que parecen
> bugs de FE pero son del contrato BE.

---

## 6. Lado FE (repo `web`, fuera de este repo) — referencia para cross-repo

Estos comandos corren en el repo **web** (fe), no acá. Se listan para cerrar el loop
cross-repo de las full-stack. **Verificar nombres exactos con `package.json` del repo web.**

```bash
yarn install
yarn build:ui          # build del front (ui)
yarn build:api         # build del BFF (api)
yarn test              # unit FE
# e2e Playwright:       (verificar script exacto en package.json — típicamente "test:e2e" o "e2e")
knip                   # detecta exports/deps muertos tras extracción FE
```

- Features puramente FE (`en-este-repo=no`): **006, 014, 015, 016, 017, 026** y las partes FE
  de 007/008/010/011/012/018. Su validación detallada va en el doc del repo web.
- `router.tsx` / `main.tsx` son archivos **MEZCLADOS** en el FE (base 006): mismo cuidado de
  hunks que en el BE.

---

## 7. Checklist de salida por extracción

Antes de proponer el merge a `develop`, confirmar:

- [ ] Fuente correcta: se extrajo desde `develop-problematico~1` (`777e5f6a`), **no** del tip.
- [ ] `go build ./...` limpio.
- [ ] `make lint` sin findings nuevos (golangci-lint v2.11.4).
- [ ] `make test` verde (incluidos los tests de 025 si aplica al módulo).
- [ ] Si toca DB: `make db-verify` limpio; backfill antes de constraints (003); `schema.expected.sql` actualizado.
- [ ] Si toca rutas/DI: API arranca y `curl` al endpoint nuevo responde el shape esperado.
- [ ] `go.mod/go.sum`: NO se reintrodujeron versiones viejas de `go-jose/v4` / `x/net` (#124 ya en develop); `go mod tidy` limpio.
- [ ] No se re-trajo nada ya-DONE: tentative-prices (#121, excluir de 018), lot-metrics/total_tons (#117/#121/#124), dependency-bumps (#124, excluir de 021).
- [ ] Sin referencias residuales a símbolos que debían borrarse (`MaybeTenantScope`, `xlsx`, `governance`, jwt legacy).
- [ ] Orden de deps respetado (001 → 002 → 003 → 004/005 → 007 → 008/009 → 010/012; 023 con su módulo).
- [ ] Full-stack: BE-first verificado con `curl` antes de mergear el FE (repo web).
- [ ] Comandos no-verificados marcados "verificar con humano" (staticcheck standalone, regeneración de wire, scripts FE, smoke companion).

---

## 8. Comandos a verificar con humano (no confirmados en este repo)

- **staticcheck standalone** (027): no hay target Makefile; golangci-lint v2 ya lo incluye.
  Alternativa probable: `go run honnef.co/go/tools/cmd/staticcheck@latest ./...`.
- **Regeneración de wire** (023): no hay target `wire` en el Makefile. ¿Se commitea `wire_gen.go`
  o se regenera en CI? Alternativa probable: `go run github.com/google/wire/cmd/wire@latest ./...`.
- **Smoke companion** (012): existe `make e2e-changes` (smoke_release.sh) y un `scripts/smoke_companion`
  mencionado en la tabla. Confirmar cuál cubre el flujo de IA.
- **Scripts FE exactos** (cross-repo): `build:ui`, `build:api`, script de e2e Playwright, knip →
  confirmar contra `package.json` del repo web.
- **Versión de golangci-lint en CI** (020): local pineada `v2.11.4`; confirmar que CI usa la misma.
