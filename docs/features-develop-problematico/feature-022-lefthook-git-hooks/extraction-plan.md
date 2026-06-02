# extraction-plan.md — feature-022 · lefthook-git-hooks (BE)

## Coordenadas
- **Repo:** `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`).
- **Rama base (destino):** `develop` (tip `003a9b8f`).
- **SOURCE de extracción:** `develop-problematico~1` (SHA **777e5f6a**). NUNCA usar `develop-problematico` (su tip es un restore/vacío).
- **Rango fuente-de-verdad (diff):** `0972e565..777e5f6a`.
- **Rama sugerida:** `pr/feature-022-lefthook-git-hooks-be`

## PR
- **Título:** `chore(tooling): agregar lefthook.yml con hooks pre-commit/pre-push (BE)`
- **Descripción (sugerida):**
  > Agrega `lefthook.yml` en la raíz del backend para validaciones locales opt-in:
  > - pre-commit (paralelo): `gofmt -l` sobre `.go` staged, `go vet ./...`, `golangci-lint run --fast --new-from-rev=HEAD`.
  > - pre-push: `go test ./... -count=1 -short`.
  >
  > Opt-in: requiere `lefthook install` por dev. Inerte para quien no lo instala (no rompe el flujo).
  > No toca código, deps, migraciones ni CI. Espejo del feature-022 en el repo FE.
  >
  > Pendiente conocido (no bloqueante): el comando `golangci-lint` asume binario de sistema con `--fast`; el repo estandariza golangci-lint **v2** pineado vía `go run ...@v2.11.4`. Considerar alinear en un follow-up.

## Archivos: enteros vs parciales
- **Enteros:** `lefthook.yml` (único archivo).
- **Parciales (partial-hunks):** ninguno.
- **Migraciones a incluir:** ninguna.
- **Tests a incluir:** ninguno (el hook invoca tests existentes; no agrega).

## Dependencias previas
- Ninguna feature debe mergearse antes. Es independiente.

## Coordinación con el otro repo (FE)
- **Orden:** indistinto (**coord cosmético**). No hay dependencia técnica entre los dos `lefthook.yml`.
- Recomendación: mergear ambos en la misma ventana para DX consistente, pero cada uno en su PR de repo.
- El PR FE de feature-022 va por separado en el repo web; no bloquea ni es bloqueado por este.

## Pasos ordenados (comandos git SUGERIDOS para un humano — el agente NO los ejecuta)
```bash
# 1) partir de develop limpio
git checkout develop
git pull --ff-only

# 2) crear rama de feature
git checkout -b pr/feature-022-lefthook-git-hooks-be

# 3) traer el archivo entero desde el SOURCE correcto (NUNCA develop-problematico tip)
git checkout develop-problematico~1 -- lefthook.yml
#   equivalente con SHA:
#   git checkout 777e5f6a -- lefthook.yml

# 4) verificar que es exactamente el archivo esperado
git diff --cached --stat            # debe listar SOLO lefthook.yml (nuevo)
git show :lefthook.yml | head -40   # contenido staged

# 5) sanity de whitespace/conflict markers
git diff --check

# 6) commit
git commit -m "chore(tooling): agregar lefthook.yml (hooks pre-commit/pre-push)"

# 7) push + PR
git push -u origin pr/feature-022-lefthook-git-hooks-be
```

## Qué NO traer
- NO traer cambios de `Makefile`, `go.mod`, `go.sum`, `.gitignore`, `.github/workflows/**` (no forman parte de esta feature; pertenecen a 019/020/021/024/etc.).
- NO crear `.golangci.yml` aquí (no existe en el SOURCE; sería invención).
- NO usar `develop-problematico` (tip vacío/restore).

## Qué podría romperse
- Nada en runtime ni en build: el archivo no se ejecuta salvo `lefthook install`.
- Para un dev que SÍ instala Lefthook con golangci-lint v2 de sistema, el `--fast` podría fallar el hook (ver risks.md). Esto afecta solo a ese dev localmente, no al merge ni a CI.

## Cómo detectar extracción incompleta
- `git -C <repo> show HEAD:lefthook.yml | diff - <(git -C <repo> show 777e5f6a:lefthook.yml)` → vacío.
- El archivo debe tener 38 líneas y los 4 comandos (gofmt, go-vet, golangci-lint, tests).

## Qué validar antes del PR
- `lefthook.yml` parsea: `lefthook validate` (si Lefthook está instalado) o `yamllint lefthook.yml`.
- `git diff --cached --stat` muestra únicamente `lefthook.yml`.

## Qué hacer después de mergear
- Anunciar al equipo: `brew install lefthook` (o `apt`) + `lefthook install`.
- (Follow-up opcional) abrir issue para alinear el comando `golangci-lint` con la versión v2 pineada del Makefile y/o agregar `.golangci.yml`.
- Coordinar con quien mergea el PR de feature-022 en FE para anunciar ambos juntos.
