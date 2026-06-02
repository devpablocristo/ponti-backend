# risks.md — feature-023 · be-wire-di

## Riesgos funcionales
- **Companion sin fallback**: `ProvideCompanionClient` falla si faltan `COMPANION_BASE_URL` / `COMPANION_INTERNAL_JWT_SECRET`, y `wire.Initialize()` devuelve error → **el binario no arranca**. Intencional (cutover hecho). *Mitigación*: documentar estas env como obligatorias en deploy (021) y validarlas en el smoke test de arranque.
- **Rate-limit**: `HTTP_RATE_LIMIT_PER_MINUTE` mal calibrado rechaza tráfico legítimo. Default `0` = desactivado. *Mitigación*: dejar en 0 hasta calibrar; medir con `/observability/metrics`.
- **CORS**: `CORSOriginList()` desde `CORS_ORIGINS`; si queda vacío en prod, el FE puede recibir errores CORS. *Mitigación*: setear `CORS_ORIGINS` por entorno.

## Riesgos técnicos
- **`wire_gen.go` es generado**: editarlo a mano o copiarlo desfasado de los `*_providers.go` produce builds rotos sutiles (un provider de más/menos, orden de args). *Mitigación*: regenerar con `go run github.com/google/wire/cmd/wire ./wire` tras tener todos los providers; nunca resolver conflictos a mano sin regenerar.
- **Reorden de args en data-integrity**: `ProvideDataIntegrityUseCases` cambió el orden (dashboard, workorder, report, supply, project, lot) y pasa repos **concretos** (`*lot.Repository`, `*report.ReportRepository`, `*supply.Repository`, `*project.Repository`) en vez de ports. Si la firma de `dataintegrity.NewUseCases` (018) no coincide exacto, compila mal o cablea repos cruzados. *Mitigación*: diffear contra `internal/data-integrity/usecases.go` de 777e5f6a.
- **`MiddlewaresEnginePort.GetProtected()` eliminado**: cualquier consumidor restante rompe. *Mitigación*: `grep -rn GetProtected internal/ cmd/ wire/` antes de portear.

## Riesgos de integración
- **Orden de merge**: si 023 entra antes que 007/012/013/018, develop queda con build roto (`undefined: ActorSet`, `axis.CompanionClient`, `NewCSVExporter`, `SupplyRepositoryPort`). *Mitigación*: 023 último; cada feature trae sus propios hunks de `wire/*`.
- **Exporters CSV (013)**: si 013 no está, `lot/supply/stock/work-order/labor.NewCSVExporter` no existen y los providers no compilan. *Mitigación*: portear 013 antes.

## Riesgos cross-repo
- Ninguno: feature solo-BE. Riesgo de mergear "solo el otro repo": N/A (no hay cambios FE). Mergear solo este repo es lo esperado.

## Riesgos de datos / migración
- `cmd/archive-cleanup --apply` **muta datos** (borra/archiva). *Mitigación*: default es dry-run; el flag `--apply` es explícito y `--apply`+`--dry-run` se rechazan (exit 2). Probar siempre `--dry-run` primero y revisar la sección "Manual Review/Blockers" del reporte.
- `cmd/migrate` usa `pg_advisory_lock` por nombre de DB para evitar corridas concurrentes; sin cambios de comportamiento, solo logging. Bajo riesgo.

## Riesgos de archivos compartidos
- `wire.go` / `wire_gen.go` / `cmd/api/main.go` / `http_server.go` son tocados por varias features simultáneamente → **conflictos de merge probables**. *Mitigación*: `restore -p` por hunk + regeneración de wire; revisar cada hunk contra la feature dueña (ver file-list.md).

## Riesgos de extracción parcial
- Traer parte de los hunks (p.ej. observability sí, pero olvidar `reporting.read_mode` o `ActorHandler`) deja referencias colgadas o features a medias. *Señal*: `undefined:`/`unused import` en `go build`. *Mitigación*: `go build ./... && go vet ./...` tras cada feature; checklist de validation.md.

## Resumen de riesgo de merge selectivo
- **Solo este repo (BE)**: correcto y suficiente. Único cuidado: que las dependencias BE (007/012/013/018/002/005/008) ya estén.
- **Solo el otro repo (FE)**: no aplica.
