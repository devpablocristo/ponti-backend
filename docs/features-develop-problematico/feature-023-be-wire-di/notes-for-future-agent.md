# notes-for-future-agent.md — feature-023 · be-wire-di

## Resumen corto
023 es **costura**: el grafo de Google Wire (`wire/*`) + el bootstrap de los binarios (`cmd/api`, `cmd/migrate`, nuevo `cmd/archive-cleanup`). No tiene negocio propio. Su diff está intrínsecamente **MEZCLADO** con actor (007), companion (012), CSV export (013), data-integrity (018), lifecycle (002), config (005), identity/tenant (008). **No es un PR aislado: va al final y sus hunks de `wire/*` viajan con cada feature dueña.**

## Qué está en FE y en BE
- **FE**: nada. "Sin cambios FE" (anotarlo en cross-repo-map del FE).
- **BE**: todo. 22 archivos (ver file-list.md).

## Archivos esenciales
- `cmd/archive-cleanup/main.go` — binario nuevo, propio de 023 (consume `lifecycle.RunArchiveCleanup` de 002).
- `cmd/migrate/*` — solo migración a slog; **lo más portable** de toda la feature.
- `cmd/api/main.go` + `http_server.go` — observability/bootstrap (propio) + `ActorHandler.Routes()` (de 007).

## Archivos peligrosos / mezclados (NO copiar a ciegas)
- `wire/wire_gen.go` — **GENERADO**. Regenerar con `go run github.com/google/wire/cmd/wire ./wire`, no copiar a mano. Mezcla 007/012/013/018.
- `wire/wire.go` — `ActorSet`/`ActorHandler` son de 007.
- `wire/ai_providers.go`, `config_providers.go` — Companion/Nexus, de 012.
- `wire/middleware_providers.go` — `Environment`/`RequireTenantHeader`, de 008; quita `GetProtected()`.
- `wire/data_integrity_providers.go`, `admin_providers.go` — recableo de 018 (repos concretos, reorden de args).
- `wire/lot|supply|stock|work_order|labor_providers.go` — Excel→CSV, de 013.

## Archivos requeridos que NO están en este flist
- `wire/actor_providers.go` (be-007), `wire/companion_providers.go` (be-012): confirmado por `grep wire/ /tmp/flists/be-007.txt` y `be-012.txt`.
- `internal/shared/lifecycle/*` (be-002): provee `RunArchiveCleanup`/`RegisterMetrics`.

## Decisiones ya tomadas
- Cutover AI completo a Companion, **sin fallback**: si faltan `COMPANION_*` env, el binario no arranca (intencional).
- `ProvideNexusClient` se construye y descarta (`_`) hasta "ola 2".
- Exporters: Excel/XLSX eliminados, todo CSV.
- data-integrity recibe repos **concretos** (no ports) porque expone métodos privados (GetRawNetIncome, etc.).
- Logging unificado en `slog` JSON en todos los `cmd/*`.

## Dudas abiertas
- ¿El split admin Repository/UseCases es de 018 o refactor propio? Resolver con `git log -- internal/admin`.
- Orden exacto de args de `dataintegrity.NewUseCases` debe matchear 018 (dashboard, workorder, report, supply, project, lot).

## Comandos para mirar primero
```
cat /tmp/flists/be-023.txt
git -C <repo> diff 0972e565..777e5f6a -- wire/wire_gen.go | head -120
git -C <repo> show 777e5f6a:cmd/archive-cleanup/main.go | head -120
grep wire/ /tmp/flists/be-007.txt /tmp/flists/be-012.txt   # confirma overlap
grep -rn "GetProtected\|ProvideConfigAI\|NewExcelExporter" wire/ internal/
```

## Errores a evitar
- NO mergear 023 antes que 007/012/013/018/002 → build de develop roto.
- NO editar `wire_gen.go` a mano para "arreglar" un conflicto → regenerar con `wire`.
- NO traer `wire/actor_providers.go` ni `wire/companion_providers.go` desde aquí (pertenecen a 007/012).
- NO correr `archive-cleanup --apply` sin un `--dry-run` previo revisado.

## Camino más seguro
1. Portear primero 001→005→002/008/009→007/012/013/018 (cada una con sus hunks de `wire/*`).
2. Recién entonces traer lo propio de 023 (cmd/migrate, cmd/archive-cleanup, observability de cmd/api).
3. Regenerar `wire_gen.go`, `go build ./...`, `go test ./cmd/migrate/...`.

## Orden vs otro repo (FE)
- N/A. No hay PR de FE asociado. Es BE-only.
