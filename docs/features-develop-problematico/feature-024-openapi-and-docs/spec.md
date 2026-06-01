# spec.md — feature-024 · openapi-and-docs (BE)

- **id**: feature-024
- **nombre**: OpenAPI & docs
- **tipo**: docs
- **repo**: Backend Go (ponti-backend) — `/home/pablocristo/Proyectos/pablo/ponti/core`
- **existe en BE**: SÍ (este paquete)
- **existe en FE**: SÍ — FULL-STACK con el mismo `feature-024`. El FE aporta `docs/`, `docs/audit` (visual regression, posiblemente generado), `RESPONSIVE_GUIDELINES`, `PR-92.md`. Coordinar con el paquete FE.
- **SOURCE de extracción**: `develop-problematico~1` (SHA `777e5f6a`). NUNCA usar `develop-problematico` (tip = restore/vacío).
- **RANGO diff (fuente de verdad)**: `0972e565..777e5f6a`.
- **rama destino**: `develop` (tip `003a9b8f`).
- **DEPENDE DE**: ninguna feature (a nivel de contenido). Hay dependencias de *referencia* débiles (los docs apuntan a código/tooling de otras features) que NO bloquean el merge porque son documentación.

## Resumen

Paquete 100% documentación. Agrega 17 archivos `.md`/`.yaml`/`.json` nuevos y modifica 3 docs existentes. No toca código fuente, deps, migraciones ni config. Cubre dos bloques:

1. **OpenAPI contract-first** (`docs/OPENAPI.md` + `docs/openapi/{openapi.yaml,swagger.yaml,swagger.json}`): documenta el pipeline BE→FE (swag genera Swagger 2.0, FE corre `yarn codegen:openapi`). El spec generado es un **piloto** que sólo cubre 2 endpoints (`GET /me/context`, `POST /data-integrity/verify-costs/{projectId}` / `GET /data-integrity/costs-check`); los ~48 handlers restantes quedan sin anotar.
2. **Documentación de arquitectura/CRUDAR/errores/tenancy/observabilidad**: `CLAUDE.md` (root), `CRUDAR_PLAN.md` (root), y `docs/*.md` (ERROR_CATALOG, OBSERVABILITY, crudar-lifecycle, archive-restore-policy, entity-capabilities, customers-projects-lifecycle, DATA_INTEGRITY_CONTRACT, MULTI_TENANT_100_EVIDENCE, BACKEND_CLEANUP_AUDIT, audit-custom-errors, projects-archive-audit) + actualización de `README.md`, `docs/README.md`, `docs/ARCHITECTURE.md`.

## Objetivo

Dejar versionada la documentación de referencia del backend (guía de onboarding Claude, plan CRUDAR, catálogo de errores, contrato de observabilidad, políticas de lifecycle/archive, evidencia multi-tenant) y el contrato OpenAPI piloto, sin tocar comportamiento.

## Problema que resuelve

- Falta de documentación canónica del layout hexagonal, reglas CRUDAR ("archived = no existe"), catálogo de errores y observabilidad.
- Drift de tipos BE↔FE: se establece pipeline contract-first (swag → openapi → `codegen:openapi`).
- Auditorías históricas (cleanup, multi-tenant evidence, projects-archive-audit, audit-custom-errors) que estaban fuera del repo.

## Alcance en este repo (BE)

Archivos nuevos (A): `CLAUDE.md`, `CRUDAR_PLAN.md`, `docs/OPENAPI.md`, `docs/ERROR_CATALOG.md`, `docs/OBSERVABILITY.md`, `docs/crudar-lifecycle.md`, `docs/archive-restore-policy.md`, `docs/entity-capabilities.md`, `docs/customers-projects-lifecycle.md`, `docs/DATA_INTEGRITY_CONTRACT.md`, `docs/MULTI_TENANT_100_EVIDENCE.md`, `docs/BACKEND_CLEANUP_AUDIT.md`, `docs/audit-custom-errors.md`, `docs/projects-archive-audit.md`, `docs/openapi/openapi.yaml`, `docs/openapi/swagger.yaml`, `docs/openapi/swagger.json`.

Archivos modificados (M): `README.md`, `docs/README.md`, `docs/ARCHITECTURE.md`.

## Alcance en el otro repo (FE, mismo feature-024)

Documentación FE: `docs/`, `docs/audit` (visual regression, posible generado), `RESPONSIVE_GUIDELINES`, `PR-92.md`. El consumo del contrato OpenAPI (`yarn codegen:openapi` → `src/api/generated/types.ts`) vive en FE; este paquete BE sólo publica el `swagger.yaml` que el FE consume.

## Fuera de alcance

- Anotar los ~48 handlers restantes con comentarios swaggo (queda como pendiente declarado en `docs/OPENAPI.md`).
- El target `openapi` del `Makefile` y `migrations_v4/000233_*` (triggers) — YA existen en `develop`; NO forman parte de este flist (pertenecen a 020/021 y a la feature CRUDAR de código, respectivamente).
- `docs/openapi/docs.go` — el output `.go` de swag NO está en el flist ni en el SOURCE; NO generarlo aquí.
- Cualquier cambio de código de error/tenancy que los docs describen (eso es 002/003/027, no 024).

## Comportamiento esperado

Ninguno en runtime. Son docs estáticos. El único "comportamiento" es que `make openapi` (target ya presente en develop) regeneraría `docs/openapi/swagger.{yaml,json}` y el FE `yarn codegen:openapi` generaría tipos a partir del yaml.

## Estado en dp~1 (777e5f6a)

- Los 17 archivos nuevos existen y son legibles/coherentes.
- `docs/openapi/openapi.yaml` (OpenAPI 3.0) + `docs/openapi/swagger.{yaml,json}` (Swagger 2.0) existen pero son **piloto de 2 endpoints** (199/185/272 líneas respectivamente).
- `docs/OPENAPI.md` admite explícitamente: "Anotado: 2 handlers piloto. Pendiente: anotar los ~48 handlers restantes".
- `CRUDAR_PLAN.md` (810 líneas) es un plan FE con estado a 2026-05-19, mezcla BE/FE; es informativo, no normativo.
- Los 3 modificados (README, docs/README, ARCHITECTURE) ya existen en develop con contenido viejo.

## Criterios de aceptación

1. Los 17 archivos nuevos aparecen idénticos a `777e5f6a` en `develop` tras el merge.
2. Los 3 modificados quedan con las secciones nuevas (hex layout, observabilidad, lifecycle, errores en ARCHITECTURE; tooling rename en README/docs/README) **sin** revertir cambios ya aplicados en develop por otras features (riesgo de conflicto: ver `risks.md`).
3. `git diff --check` limpio (sin whitespace errors) — son `.md`, riesgo bajo.
4. No se introduce ningún `.go`, migración, ni cambio de `go.mod`/`Makefile`.

## Endpoints / modelos / UI / DB / tests afectados

- **Endpoints**: documentados (no creados). El spec OpenAPI referencia `GET /data-integrity/costs-check`, `POST /data-integrity/verify-costs/{projectId}`, `GET /me/context`. Los `.md` describen patrones `POST /v1/{resource}/{id}/archive|restore`, `GET /observability/metrics`.
- **Modelos/DTOs**: el spec referencia `IntegrityReportResponse`, `IntegrityCheckDTO`, `MeContext`, `MeUser`, `MeTenant` (definidos en código de otras features; aquí sólo el yaml generado).
- **UI**: ninguna (esto es BE). El FE de feature-024 maneja su propia doc.
- **DB**: ninguna migración en este flist.
- **Tests**: ninguno.

## Dependencias

- **Intra-repo (débiles, de referencia)**: los docs apuntan a `Makefile:openapi`, `migrations_v4/000233`, `internal/data-integrity/handler.go` (`@Router`), `internal/shared/handlers/errors.go`, `cmd/api/main.go`. Todos existen YA en develop, así que ningún doc queda "colgado". No bloquean.
- **Cross-repo (débil)**: `docs/OPENAPI.md` instruye correr `yarn codegen:openapi` en `web/ui`; ese script y `src/api/generated/types.ts` viven en FE feature-024. Coordinación recomendada pero no bloqueante para mergear el BE.

## Riesgos

- **Funcionales**: nulos (docs).
- **Técnicos**: conflictos de merge en `README.md`/`docs/README.md`/`docs/ARCHITECTURE.md` porque develop ya divergió (tooling rename `staging-db-2-local-db`→`reset-local-db-from-prod`, `core/*`→`platform/*`, `ponti-frontend`→`web`, `migrate-up`→`db-migrate-up`). Estos hunks pisan terreno de 019/021/001. Ver `risks.md` y `file-list.md` (sección Compartidos).
- **Doc desactualizada / engañosa**: `MULTI_TENANT_100_EVIDENCE.md` y `BACKEND_CLEANUP_AUDIT.md` citan conteos/fechas viejas (2026-05-12, `schema_migrations=225`) que ya no matchean develop. Aceptable como snapshot histórico, pero NO tratar como verdad presente.
- **OpenAPI incompleto**: el spec sólo cubre 2 endpoints; si el FE intenta `codegen` esperando cobertura total, generará tipos parciales. Documentar como piloto.

## DECISIÓN recomendada

**Extraer tal cual, con manejo cuidadoso de los 3 archivos modificados.**

- Los 17 nuevos: `whole-file` (copia directa desde `777e5f6a`). Seguro, sin conflicto posible (no existen en develop).
- Los 3 modificados (`README.md`, `docs/README.md`, `docs/ARCHITECTURE.md`): `partial-hunks` / `manual-port`. Traer SÓLO los hunks de contenido propios de 024 (secciones hex/observabilidad/lifecycle/errores en ARCHITECTURE; nada más en README/docs.README si ya está cubierto por 019/021). Revisar que no se revierta el rename de tooling que develop ya tiene. Ver `extraction-plan.md`.
- No partir en subfeatures: el bloque es pequeño y cohesivo. Único cuidado: el spec OpenAPI piloto debe quedar etiquetado como tal para no engañar al FE.
