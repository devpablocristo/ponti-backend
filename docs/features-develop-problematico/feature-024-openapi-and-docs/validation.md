# validation.md — feature-024 · openapi-and-docs (BE)

Como es docs, la validación es de integridad de archivos + render de markdown, NO de runtime. No hay `go test`/`go build` que aplique a este flist (no hay `.go`).

## Checklist pre-PR

- [ ] Los 17 archivos nuevos existen en la rama y son byte-idénticos al SOURCE:
  - `git -C <repo> diff 777e5f6a -- CLAUDE.md CRUDAR_PLAN.md docs/OPENAPI.md docs/ERROR_CATALOG.md docs/OBSERVABILITY.md docs/crudar-lifecycle.md docs/archive-restore-policy.md docs/entity-capabilities.md docs/customers-projects-lifecycle.md docs/DATA_INTEGRITY_CONTRACT.md docs/MULTI_TENANT_100_EVIDENCE.md docs/BACKEND_CLEANUP_AUDIT.md docs/audit-custom-errors.md docs/projects-archive-audit.md docs/openapi/openapi.yaml docs/openapi/swagger.yaml docs/openapi/swagger.json` → debe ser **vacío**.
- [ ] `ls docs/openapi/` muestra exactamente `openapi.yaml swagger.yaml swagger.json` (NO `docs.go`).
- [ ] `docs/ARCHITECTURE.md`: las secciones nuevas (layout por módulo, reglas duras, observabilidad, lifecycle, errores, seguridad) están presentes y NO se revirtió contenido que develop ya tenía.
- [ ] `README.md` / `docs/README.md`: si se incluyeron, los hunks NO entran en conflicto con 019/021 ni revierten renames ya aplicados en develop. Si hay duda, omitirlos del PR.
- [ ] `git -C <repo> diff --check` → sin errores de whitespace.
- [ ] `git -C <repo> status` → solo cambios en `docs/`, `CLAUDE.md`, `CRUDAR_PLAN.md`, `README.md`. NINGÚN `.go`, `Makefile`, `go.mod`, `go.sum`, migración.

## Validación manual

- [ ] Abrir cada `.md` nuevo y confirmar que renderiza (tablas, code fences cerrados). Especial atención a `ERROR_CATALOG.md`, `entity-capabilities.md`, `CRUDAR_PLAN.md` (muchas tablas).
- [ ] Validar YAML/JSON del spec:
  - `python3 -c "import yaml,sys; yaml.safe_load(open('docs/openapi/openapi.yaml')); yaml.safe_load(open('docs/openapi/swagger.yaml')); print('yaml ok')"`
  - `python3 -c "import json; json.load(open('docs/openapi/swagger.json')); print('json ok')"`
- [ ] Confirmar que los enlaces relativos dentro de los `.md` apuntan a archivos del paquete o a código existente en develop (ej. `docs/OBSERVABILITY.md`, `../CLAUDE.md`, `docs/crudar-lifecycle.md`). Enlace muerto = cosmético, no bloqueante.

## Tests sugeridos

- **BE**: ninguno aplica (no hay paquete Go en el flist). Opcional, como sanity general del repo tras el merge: `go build ./cmd/api` (debe seguir verde — pero 024 no lo toca, así que no debería cambiar nada).
- **FE (en el repo web, feature-024)**: tras publicar `swagger.yaml`, validar el pipeline de tipos:
  - `cd ../web/ui && yarn codegen:openapi` → debe generar `src/api/generated/types.ts` sin error (recordar: spec piloto, cobertura parcial esperada).
  - `yarn typecheck` / `yarn build` en FE para confirmar que los tipos generados compilan.

## Casos borde

- Spec OpenAPI con cobertura parcial: el FE codegen NO debe asumir endpoints faltantes como eliminados. Verificar que el FE-024 documenta que sólo hay 2 endpoints anotados.
- `swag` no instalado: `make openapi` fallaría, pero NO se necesita ejecutar para mergear docs (el yaml ya está versionado).
- Conflicto de merge en README: resolver a mano conservando los renames de develop.

## Qué revisar en UI / API / DB / env

- **UI**: nada en este repo. (FE-024 revisa su propia doc/audit.)
- **API**: ningún endpoint cambia. El spec sólo documenta.
- **DB**: ninguna migración. Ignorar los conteos históricos de los `*_EVIDENCE/_AUDIT.md`.
- **env**: ninguna variable nueva. `OBSERVABILITY.md` documenta `OTEL_EXPORTER` (ya existente), no la introduce.

## Qué validar en el otro repo (FE feature-024)

- Que el FE consume el `swagger.yaml` correcto (el de 024, no uno viejo).
- Que `yarn codegen:openapi` corre y los tipos generados se commitean.
- Que la doc FE (`docs/audit`, `RESPONSIVE_GUIDELINES`, `PR-92.md`) no duplica/contradice la doc BE.

## Señales de incompletitud / incompatibilidad

- `git diff 777e5f6a -- <path>` NO vacío en alguno de los 17 → extracción incompleta.
- Aparece `docs/openapi/docs.go` → se trajo algo que no corresponde.
- README/docs.README revirtieron `reset-local-db-from-prod`/`db-migrate-up`/`platform/*` a los nombres viejos → se pisó trabajo de 019/021.
- El PR incluye archivos `.go`/migraciones → fuera de alcance de 024.
