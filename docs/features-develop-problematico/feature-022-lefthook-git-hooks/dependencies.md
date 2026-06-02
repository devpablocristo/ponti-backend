# dependencies.md — feature-022 · lefthook-git-hooks (BE)

## Resumen
Feature autónoma de un solo archivo (`lefthook.yml`). **No depende de** ninguna otra feature para mergear, y **no bloquea** a ninguna.

## Depende de
- **Ninguna (dura).** Declarado en el brief: "DEPENDE DE: ninguna".

### Dependencias débiles / funcionales (no bloquean el merge, afectan el uso)
- **golangci-lint** (herramienta externa). El hook `golangci-lint run --fast --new-from-rev=HEAD` necesita el binario instalado. El repo, sin embargo, usa golangci-lint **v2** pineado vía `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4` en el `Makefile`. → desalineamiento débil (ver risks.md). No bloquea el merge porque el hook es opt-in.
- **feature-025 (be-test-coverage):** el hook `pre-push` corre `go test ./... -short`. Su utilidad depende de que existan tests; si el repo tiene pocos/ningún test, el hook simplemente pasa rápido. Relación informativa, no bloqueante.
- **feature-020 (ci-workflows):** CI corre lint con `golangci/golangci-lint-action@v9` y tests con coverage en `ci-pr.yml`. Es la red de seguridad "real"; el hook local es complementario. No hay dependencia de orden.

## Bloquea a
- **Ninguna.** Ningún otro paquete necesita `lefthook.yml` para funcionar.

## Cross-repo
- **FE feature-022 (lefthook-git-hooks):** existe un `lefthook.yml` espejo en el repo web. **Dependencia: ninguna (técnica).** Coordinación cosmética/DX para mergear en la misma ventana. Orden indistinto.

## Archivos / tipos / config / migraciones / APIs compartidos
- **Compartidos:** ninguno. Confirmado con `git grep -l -i lefthook 777e5f6a` → solo `lefthook.yml`.
- No toca `wire/*`, `cmd/api/*`, `internal/shared/**`, `go.mod`/`go.sum`, `Makefile`, `.gitignore`.

## Recomendación de orden
- Mergeable de forma **independiente y en cualquier momento**.
- Sugerencia DX: mergear junto al PR de feature-022 del FE (no obligatorio).
- No requiere que 019/020/021/024/025 estén mergeadas antes.
