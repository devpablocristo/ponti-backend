# risks.md — feature-021 (BE)

## Riesgo MÁXIMO: portar `go.mod`/`go.sum` como si fueran "config"

- **Qué pasa:** el delta `0972e565..777e5f6a` agrega deps (`platform/observability/go`, `platform/persistence/gorm/go`, `prometheus/client_golang`, otel exporters) y quita otras (`excelize/v2`, `cloudsqlconn`, `xuri/*`, `richardlehane/*`, `tiendc/go-deepcopy`). Esas deps están atadas a CÓDIGO que en develop está incompleto (001/002/005) o todavía presente (013 usa excelize).
- **Consecuencia:** si se hace `git checkout 777e5f6a -- go.mod go.sum`, luego `go mod tidy` borra lo no usado y `go build` falla por lo que se quitó de más. Build roto, difícil de diagnosticar.
- **Mitigación:** NO incluir `go.mod`/`go.sum` en el PR de 021. Regenerarlos con `go mod tidy` SÓLO después de mergear las features de código. Gate de revisión: el PR no debe tocar esos 2 archivos.

## Riesgo de archivos compartidos

- `go.mod`/`go.sum` los toca casi cualquier feature de código → conflictos de merge garantizados si 021 los modifica. Mitigación: 021 los deja intactos.
- `.gitignore` lo comparte 019 (entrada `schema.snapshot.sql`) y el tooling local (`go.work`). Mitigación: portar SÓLO el hunk de `schema.snapshot.sql`; rechazar los hunks que borran `go.work`/`/api`/`scripts/db/*.env`.
- `Dockerfile` lo comparten 005/020 (prefetch observability/persistence). Mitigación: diferir ese hunk.

## Riesgo de extracción parcial

- Tomar `docker-compose.yml` whole-file desde dp~1 re-borraría `GOWORK=off` correctamente, pero también podría arrastrar diferencias de `env_file`/ports si el archivo divergió más. Mitigación: `git restore -p` hunk a hunk; `git diff --check` y `docker compose config` para validar.
- Tomar `.gitignore` whole-file revierte el tooling `go.work` de develop. Mitigación: hunk selectivo.

## Riesgos funcionales

- Quitar el mount `${CORE_REPO_DIR}` y `GOWORK=off`: si algún dev todavía depende de iterar contra un `platform/` local con `go.work`, su flujo cambia. Bajo impacto (la NOTA documenta la alternativa con `PLATFORM_REPO_DIR`). Mitigación: comunicar en el PR.

## Riesgos técnicos

- `go.sum` 679 líneas: imposible portar a mano sin error. Mitigación: nunca editar a mano; `go mod tidy`/`go mod download` lo regeneran.
- Bumps go-jose/x/net (#124) ya en develop: re-portarlos genera conflicto/regresión de versión. Mitigación: excluir explícitamente.

## Riesgos de integración / cross-repo

- FE-021 y BE-021 no comparten archivos. Riesgo de integración técnico: **nulo**. El único riesgo es de coordinación de mensajería de release (que se "anuncie" 021 completa cuando sólo un lado mergeó). Mitigación: trackear ambos PRs bajo el mismo feature-021.

## Riesgos de datos / migración

- Ninguno. 021 no toca DB ni migraciones. El ignore de `schema.snapshot.sql` sólo evita commitear un artefacto generado.

## Riesgo de mergear sólo un repo

- **Sólo BE-021:** seguro. No afecta al FE; build/compose del BE quedan consistentes.
- **Sólo FE-021:** seguro para el BE. No hay dependencia.
- **Conclusión:** ambos lados son independientes; mergear uno sin el otro no rompe nada.

## Señal temprana de problema

- Si el PR de 021 muestra `go.mod`/`go.sum` en `git diff --stat` → STOP, mezcla indebida.
- Si tras el port `git grep "CORE_REPO_DIR" docker-compose.yml` devuelve algo → port incompleto.
- Si `go.work` desapareció de `.gitignore` → se arrastró un hunk de más.
