# risks.md — feature-024 · openapi-and-docs (BE)

Riesgo global: **BAJO**. Es documentación; no hay código, deps, migraciones ni config en el flist. El mayor riesgo real es de merge conflict en 3 archivos y de doc engañosa/desactualizada.

## Funcionales

- **Nulos.** Ningún cambio de runtime. Markdown/yaml/json estáticos.

## Técnicos

- **Conflicto de merge en `README.md` / `docs/README.md` / `docs/ARCHITECTURE.md`** (MEDIO-ALTO para README).
  - develop ya tiene contenido viejo y otras features (019/021/001) probablemente toquen las mismas líneas (renames de tooling).
  - *Mitigación*: usar `git restore -p --source=777e5f6a` y aceptar sólo hunks propios de 024 (secciones de arquitectura). Para README/docs.README, si 019/021 ya renombraron comandos, descartar esos hunks aquí.
- **Traer `docs/ARCHITECTURE.md` entero pisando un cambio de develop** (MEDIO).
  - *Mitigación*: `git diff develop:docs/ARCHITECTURE.md 777e5f6a:docs/ARCHITECTURE.md` antes de decidir entero vs parcial.
- **`git diff --check`** (BAJO): posibles whitespace warnings en `.md`. Trivial.

## Integración

- **Spec OpenAPI piloto (2 endpoints)** (MEDIO).
  - Si el FE corre `yarn codegen:openapi` esperando cobertura completa, generará `types.ts` parcial y podría asumir que faltan endpoints.
  - *Mitigación*: el propio `docs/OPENAPI.md` ya lo declara ("Anotado: 2 handlers piloto"). Comunicarlo al FE-024 en el PR.
- **Referencias a código que podría cambiar** (BAJO).
  - `audit-custom-errors.md` cita `internal/report/repository.go:45/85/547/...` y `business-parameters/.../business_parameter.go:26`. Esos line numbers pueden ya no matchear develop.
  - *Mitigación*: aceptar como auditoría fechada read-only; no se usa para automatizar nada.

## Cross-repo

- **Desincronización BE/FE del contrato** (MEDIO-BAJO).
  - El `swagger.yaml` congelado puede divergir de los DTOs reales (`IntegrityReportResponse`, `MeContext`, ...) si 018/008 evolucionan.
  - *Mitigación*: tratar el spec como snapshot; el contrato vivo se regenera con `make openapi`.
- **Mergear solo BE** (BAJO): el FE-024 trae su propia doc; no rompe nada que el BE vaya solo. El FE simplemente no tendrá `types.ts` regenerado hasta correr codegen.
- **Mergear solo FE** (BAJO): el FE-024 sin el BE no tendría `swagger.yaml` actualizado; usaría el viejo o fallaría el codegen. Por eso se recomienda BE-first para la parte OpenAPI.

## Datos / migración

- **Ninguno.** No hay migraciones en el flist.
- `MULTI_TENANT_100_EVIDENCE.md`/`BACKEND_CLEANUP_AUDIT.md` mencionan estados de DB (`schema_migrations=225`, conteos) — son evidencia histórica, NO instrucciones de migración. Riesgo: que alguien los lea como estado actual. *Mitigación*: las propias fechas (2026-05-12) los marcan como snapshot.

## Archivos compartidos

- `README.md`, `docs/README.md`, `docs/ARCHITECTURE.md` son compartidos con 001/019/021.
  - *Mitigación*: partial-hunks; coordinar dueño de los renames de tooling con 019/021.
- `CRUDAR_PLAN.md` es FE-céntrico dentro del repo BE; riesgo de confusión sobre dónde vive el "source of truth" del plan. *Mitigación*: nota en el PR.

## Extracción parcial

- **Riesgo de olvidar uno de los 17 nuevos** (BAJO).
  - *Mitigación*: tras el checkout, `git diff 777e5f6a -- <cada path>` debe ser vacío para los 17.
- **Riesgo de traer `docs/openapi/docs.go` por error** (BAJO): NO existe en SOURCE/flist; no debe aparecer. *Mitigación*: `ls docs/openapi/` debe mostrar exactamente 3 archivos.

## Resumen de riesgo por escenario de merge

- **Solo este repo (BE)**: seguro. Docs autocontenidos. FE queda sin codegen actualizado (cosmético).
- **Solo el otro repo (FE)**: el FE-024 puede mergear su doc; el codegen quedaría contra el swagger viejo. Preferir BE-first para OpenAPI.
- **Ambos coordinados**: ideal. BE publica swagger → FE corre codegen.
