# risks.md — feature-022 · lefthook-git-hooks (BE)

## Riesgos funcionales
- **Muy bajo.** El archivo no afecta runtime, build ni datos. Es tooling de git local, opt-in (requiere `lefthook install`). Quien no instala Lefthook ni se entera.

## Riesgos técnicos
1. **golangci-lint: versión y flag (MEDIO).**
   - Hook: `golangci-lint run --fast --new-from-rev=HEAD` → asume binario de sistema.
   - Repo: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --timeout=5m` (Makefile) y `golangci/golangci-lint-action@v9` (CI) → golangci-lint **v2**.
   - En golangci-lint v2 el flag `--fast` fue **eliminado/renombrado** (existía en v1). Si el dev tiene v2 de sistema, el hook puede fallar al parsear `--fast`.
   - Si el dev tiene v1 de sistema, los resultados de lint diferirán de CI (v2) → falsa sensación de "limpio".
   - **Mitigación:** alinear el hook a `go run ...@v2.11.4 run --new-from-rev=HEAD` (follow-up no bloqueante). Documentar en el PR que es pendiente conocido.

2. **Sin `.golangci.yml` (BAJO).** No existe config en repo → hook y CI corren con defaults de su versión respectiva. Resultados pueden divergir. **Mitigación:** agregar `.golangci.yml` compartido (fuera de alcance de esta feature).

3. **`gofmt`/`go vet` de sistema vs toolchain pineado (BAJO).** `go vet ./...` y `gofmt` usan el `go` del PATH del dev; el repo declara `go 1.26.3`. Diferencias menores posibles. **Mitigación:** documentar versión de Go esperada.

## Riesgos de integración
- **Ninguno con CI:** el hook es local; CI (`ci-pr.yml`, feature-020) es independiente. No se pisan.
- **`pre-push` corre `go test ./... -short`:** si la suite es lenta o flaky (ver feature-025), puede molestar al push. Es `-short`, mitiga. Bypass disponible: `git push --no-verify`.

## Riesgos cross-repo (FE)
- **Muy bajo.** El `lefthook.yml` del FE es independiente. Ningún archivo, tipo o config compartido. Mergear uno sin el otro no rompe nada.

## Riesgos de datos / migración
- **Ninguno.** No hay DB, migraciones ni datos involucrados.

## Riesgos de archivos compartidos
- **Ninguno.** Esta feature no toca archivos compartidos (`wire/*`, `cmd/api/*`, `internal/shared/**`, `go.mod`, `Makefile`, `.gitignore`). Confirmado: único match de "lefthook" en el árbol es el propio archivo.

## Riesgos de extracción parcial
- **Ninguno.** Es un solo archivo nuevo; se trae entero. No hay hunks que separar. Si por error se trajera vacío o truncado, se detecta con el diff contra `777e5f6a:lefthook.yml` (ver validation.md).

## Riesgo de mergear SOLO este repo (BE) / SOLO el otro (FE)
- **Aceptable en ambos sentidos.** No hay acoplamiento técnico entre los dos `lefthook.yml`. Mergear solo el BE deja al FE sin hooks (y viceversa), lo cual es una inconsistencia de DX, no un fallo. Recomendado coordinar para evitar que un dev tenga hooks en un repo y no en el otro.

## Resumen de severidad
| Riesgo | Severidad | Bloquea merge |
|--------|-----------|---------------|
| golangci-lint v2 + `--fast` | Medio | No (opt-in) |
| Sin `.golangci.yml` | Bajo | No |
| Toolchain Go de sistema | Bajo | No |
| pre-push tests lentos/flaky | Bajo | No (bypass disponible) |
| Cross-repo desincronizado | Bajo | No |
