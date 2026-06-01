# implementation-status.md — feature-022 · lefthook-git-hooks (BE)

## Estado general
- **Estado:** completa (como config). El archivo existe, parsea y cubre el flujo previsto.
- **% completitud:** ~95%. El 5% faltante es el desalineamiento del comando `golangci-lint` con la versión v2 pineada del repo (calidad, no funcionalidad).

## Estado en este repo (BE)
- `lefthook.yml` presente en `777e5f6a`, 38 líneas, status `A`.
- **Ausente en `develop`** (tip `003a9b8f`) → pendiente de portar.
- Contenido:
  - `pre-commit` (paralelo): `gofmt`, `go-vet` (`go vet ./...`), `golangci-lint` (`run --fast --new-from-rev=HEAD`).
  - `pre-push`: `go test ./... -count=1 -short`.
  - Comentarios de instalación y bypass incluidos.
- NO hay wiring automático (Makefile / go.mod / tools.go no instalan ni ejecutan Lefthook). Es opt-in puro.

## Estado en el otro repo (FE)
- Existe `lefthook.yml` espejo (feature-022 FE). Estado a confirmar en el paquete del repo web. Independiente del BE.

## Tests
- No agrega tests propios.
- El hook `pre-push` invoca `go test ./... -short` (depende de feature-025 para tener cobertura real).

## Pendientes
### BLOQUEANTE para mergear
- **Ninguno.** El archivo es inerte sin `lefthook install`; mergearlo no puede romper build, runtime, CI ni a otros devs.

### Mejora futura (no bloqueante)
- Alinear el comando del hook con golangci-lint v2: el Makefile usa `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4`, pero el hook llama al binario de sistema `golangci-lint` con `--fast`. En v2, `--fast` fue eliminado/renombrado → posible fallo del hook. Propuesta: `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --new-from-rev=HEAD`.
- (Opcional) agregar `.golangci.yml` para que hook y CI compartan config (hoy ambos corren con defaults distintos según versión).
- (Opcional) agregar un target `make hooks` o script de bootstrap que corra `lefthook install`, para reducir fricción.

### Deuda aceptable
- El hook depende de que cada dev instale Lefthook manualmente. Aceptable para tooling opt-in.

### Duda humana
- ¿El equipo quiere golangci-lint del sistema (rápido, requiere instalar) o el `go run` pineado (reproducible, más lento)? Decisión de DX que conviene unificar entre hook, Makefile y CI.
- ¿Se mergean los `lefthook.yml` de BE y FE juntos? (cosmético).

## Bugs observados
- **Potencial (medio):** `golangci-lint --fast` puede fallar con golangci-lint v2 instalado de sistema. Solo afecta a devs que instalaron el hook; no afecta el merge ni CI.
