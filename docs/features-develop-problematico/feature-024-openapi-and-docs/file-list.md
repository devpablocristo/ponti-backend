# file-list.md — feature-024 · openapi-and-docs (BE)

Flist autoritativo: `/tmp/flists/be-024.txt` (20 entradas). SOURCE = `777e5f6a`. Todos son documentación; ninguno es código/migración/config.

## Propios (whole-file, sin riesgo de conflicto — no existen en develop)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `CLAUDE.md` | A | doc (root) | Guía rápida onboarding Claude: stack, layout hex, reglas duras, CRUDAR, tenancy, errores | whole-file | nuevo, autocontenido; sólo enlaza otros docs del propio paquete | bajo | alta |
| `CRUDAR_PLAN.md` | A | doc (root) | Plan de homogeneización FE CRUDAR (810 líneas, estado 2026-05-19) | whole-file | nuevo; informativo, mezcla BE/FE pero no normativo | bajo (puede confundir por ser FE-centric) | alta |
| `docs/OPENAPI.md` | A | doc | Pipeline contract-first BE→FE + cómo anotar handlers | whole-file | nuevo; declara estado piloto (2 handlers) | bajo | alta |
| `docs/ERROR_CATALOG.md` | A | doc | Catálogo de kinds `domainerr.*` → HTTP + patrones de mensaje | whole-file | nuevo; describe código que ya existe (domainerr/sharedhandlers) | bajo | alta |
| `docs/OBSERVABILITY.md` | A | doc | slog JSON + Prometheus RED + OTel; env vars `OTEL_EXPORTER` | whole-file | nuevo; referencia `cmd/api/main.go`, ya en develop | bajo | alta |
| `docs/crudar-lifecycle.md` | A | doc | Definición CRUDAR, estados `deleted_at`, archive cause/batch | whole-file | nuevo; canónico del lifecycle | bajo | alta |
| `docs/archive-restore-policy.md` | A | doc | Política transaccional archive/restore/hard-delete ("restore by cause") | whole-file | nuevo | bajo | alta |
| `docs/entity-capabilities.md` | A | doc | Tabla CRUDAR por entidad (actors, customers, projects, fields, lots...) | whole-file | nuevo | bajo | alta |
| `docs/customers-projects-lifecycle.md` | A | doc | Relación Actor/Customer/Project + `EnsureCustomerFromActor` | whole-file | nuevo; referencia `internal/actor/master_link.go`, `legacy_sync.go` | bajo | alta |
| `docs/DATA_INTEGRITY_CONTRACT.md` | A | doc | Contrato `/data-integrity/costs-check`: estados OK/ERROR/WARNING/SKIPPED, STRONG/WEAK/FORMULA_ALIGNMENT | whole-file | nuevo | bajo | alta |
| `docs/openapi/openapi.yaml` | A | spec OpenAPI 3.0 | Spec convertido (piloto, ~2 endpoints, 199 líneas) | whole-file | generado; congelado al SOURCE | medio (incompleto, no engañar al FE) | alta |
| `docs/openapi/swagger.yaml` | A | spec Swagger 2.0 | Output `swag` (185 líneas) | whole-file | generado por `make openapi`; congelado | medio (incompleto) | alta |
| `docs/openapi/swagger.json` | A | spec Swagger 2.0 | Output `swag` JSON (272 líneas) | whole-file | generado; congelado | medio (incompleto) | alta |

## Snapshots históricos (whole-file, pero NO tratar como verdad presente)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `docs/MULTI_TENANT_100_EVIDENCE.md` | A | doc/auditoría | Evidencia multi-tenant local 2026-05-12 (`schema_migrations=225`, conteos) | whole-file | nuevo; útil como evidencia histórica | medio (datos viejos vs develop actual) | alta |
| `docs/BACKEND_CLEANUP_AUDIT.md` | A | doc/auditoría | Baseline de limpieza 2026-05-12 (403 archivos Go, 56 migraciones) | whole-file | nuevo; baseline histórico | medio (números desactualizados) | alta |
| `docs/audit-custom-errors.md` | A | doc/auditoría | Sweep read-only de `fmt.Errorf`/`errors.New` → propuesta `domainerr.*` | whole-file | nuevo; cita líneas exactas que pueden haber drifteado | medio (line numbers pueden no matchear) | alta |
| `docs/projects-archive-audit.md` | A | doc/auditoría | Veredicto "correcto con deuda" del restore customer/project, 2026-05-26 | whole-file | nuevo; documenta deuda BE de `RunCascadeRestore` | bajo | alta |

## Compartidos (partial-hunks / manual-port — YA existen en develop, riesgo de conflicto)

| path | status | tipo | rol | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `docs/ARCHITECTURE.md` | M | doc | Reescribe TLDR + agrega secciones: layout por módulo, reglas duras, observabilidad, lifecycle, errores, seguridad | partial-hunks | el grueso del diff es contenido nuevo propio de 024; SEGURO de traer. Verificar que develop no haya tocado ya las mismas líneas (TLDR) | medio | alta |
| `README.md` | M | doc | Renames de tooling: `frontend`→`web`, `core/*`→`platform/*`, `staging-db-2-local-db`→`reset-local-db-from-prod`, `ponti-frontend/api`→`web/api`, `.env.example` | partial-hunks / manual-port | estos renames son territorio de feat-019 (db scripts) / 021 (deploy) / 001 (platform). En develop el README AÚN dice `staging-db-2-local-db` y `ponti-frontend`. Decidir si estos hunks van en 024 o se delegan a 019/021 | ALTO (overlap de intención + conflicto) | media |
| `docs/README.md` | M | doc | 1 línea: `make migrate-up`→`make db-migrate-up` | manual-port | rename de target Makefile → pertenece a 019/021. En develop sigue `migrate-up` | medio | media |

## Requeridos por dependencia

Ninguno. Es docs; no requiere que se porte código previo. Las referencias internas (Makefile:openapi, migrations_v4/000233, internal/data-integrity/handler.go, internal/shared/handlers/errors.go, cmd/api/main.go) YA están en develop — verificado.

## Dudosos

- `README.md`, `docs/README.md`: los hunks de rename de tooling solapan con 019/021/001. Decisión humana: ¿van con 024 (porque el flist los marca M en este rango) o se dejan para 019/021? Recomendación: traer SOLO si no entran en conflicto con lo que 019/021 ya aplicó; si 019/021 ya renombró, hacer `git restore -p` y descartar esos hunks.
- `CRUDAR_PLAN.md`: es plan FE dentro del repo BE. ¿Pertenece a BE o debería estar en FE? El flist lo asigna a BE. Traer tal cual, pero saber que su contenido es FE-céntrico (no normativo para BE).

## NO traer todavía

- `docs/openapi/docs.go`: NO está en el flist ni en el SOURCE. Es el output `.go` de swag. NO crearlo manualmente; se regenera con `make openapi`.
- Cualquier cambio de código que los docs describen (domainerr, lifecycle helpers, triggers `000233`, handler annotations): NO es de 024 — son 002/003/009/027. Los docs sólo los referencian.

## Resumen de conteo

- Total flist: 20 paths.
- Whole-file seguros (nuevos): 17.
- Partial/manual (modificados, con conflicto posible): 3.
- Requeridos por dependencia: 0. Do-not-extract-yet: 0 (dentro del flist).
