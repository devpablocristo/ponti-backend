# file-list.md — feature-022 · lefthook-git-hooks (BE)

Fuente de la lista: `/tmp/flists/be-022.txt` (1 entrada). Verificada contra
`git diff 0972e565..777e5f6a -- lefthook.yml` (new file mode, +38 líneas) y
`git cat-file -e develop:lefthook.yml` (no existe en develop).

## Propios (de esta feature)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|------|--------|------|-------------------|------------|--------|--------|-----------|
| `lefthook.yml` | A | config YAML (raíz) | Define hooks `pre-commit` (gofmt/go-vet/golangci-lint) y `pre-push` (go test -short). Es la feature completa. | **whole-file** | Archivo nuevo, autónomo, sin hunks mezclados con otras intenciones. No existe en `develop`. | bajo (es opt-in; inerte sin `lefthook install`) | alta |

## Compartidos (partial-hunks)
_Ninguno._ Esta feature no toca archivos compartidos. Confirmado: el único match de "lefthook" en todo el árbol `777e5f6a` es el propio `lefthook.yml` (`git grep -l -i lefthook 777e5f6a`). No hay hunks en `Makefile`, `go.mod`, `go.sum`, `.gitignore`, `wire/*`, `cmd/api/*`, `internal/shared/**`.

## Requeridos por dependencia
_Ninguno._ La feature no depende de otra feature para mergear.

## Dudosos
_Ninguno._

## NO traer todavía
_Nada de la lista._ La feature es un solo archivo y se trae entero.

## Notas de archivos referenciados pero NO incluidos (contexto, NO extraer aquí)
- `.golangci.yml` / `.golangci.yaml` — **no existe** en `777e5f6a` ni en `develop`. El hook `golangci-lint` correrá con la config por defecto. No forma parte de esta feature.
- `Makefile` — ya tiene target `lint` (`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run --timeout=5m`) y `test` (`go test ./...`). NO se modifica en esta feature; es contexto para entender el desalineamiento de versión de golangci-lint.
- `.github/workflows/ci-pr.yml` — corre lint vía `golangci/golangci-lint-action@v9` y tests con coverage (feature-020). Independiente del hook local; NO se toca.
