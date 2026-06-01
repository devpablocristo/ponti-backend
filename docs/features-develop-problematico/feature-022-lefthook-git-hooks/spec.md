# spec.md — feature-022 · lefthook-git-hooks (Backend Go / ponti-backend)

## Identificación
- **ID:** feature-022
- **Slug:** lefthook-git-hooks
- **Nombre:** Lefthook git hooks
- **Tipo:** config (tooling local de desarrollo)
- **Repo:** Backend Go — `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **Existe en BE:** Sí (este paquete).
- **Existe en FE:** Sí — FULL-STACK. El mismo feature-022 existe en el repo web con su propio `lefthook.yml` (hooks de JS/TS: probablemente eslint/prettier/tsc/test). Coordinar con el paquete del otro repo.
- **Merge:** por repo (cada repo mergea su propio `lefthook.yml`; no hay artefacto compartido entre repos).

## Resumen
Se agrega un único archivo nuevo `lefthook.yml` en la raíz del repo BE que define git hooks gestionados por [Lefthook](https://github.com/evilmartians/lefthook):
- **pre-commit** (en paralelo): `gofmt -l` sobre archivos `.go` staged, `go vet ./...`, y `golangci-lint run --fast --new-from-rev=HEAD`.
- **pre-push:** `go test ./... -count=1 -short`.

El archivo incluye comentarios de instalación (`brew install lefthook` / `sudo apt install lefthook` + `lefthook install`), cómo correrlo a mano (`lefthook run pre-commit`) y cómo hacer bypass (`git commit --no-verify`).

## Objetivo
Dar a los devs del backend un set de validaciones rápidas locales antes de commitear/pushear (formato, vet, lint incremental, tests cortos), para reducir ida y vuelta con CI.

## Problema que resuelve
Sin hooks locales, errores de formato/vet/lint solo se detectan en CI (workflow `ci-pr.yml`, ver feature-020), gastando ciclos de pipeline. Lefthook acerca esas validaciones al momento del commit/push.

## Alcance en este repo (BE)
- **Único archivo:** `lefthook.yml` (raíz). Status `A` (creado). 38 líneas.
- NO modifica `Makefile`, `go.mod`, `.gitignore`, ni los workflows de `.github/workflows/`.
- NO agrega `.golangci.yml` (la config de golangci-lint no existe en el repo, ni en `777e5f6a` ni en `develop`).

## Alcance en el otro repo (FE)
- El FE tiene su propio `lefthook.yml` con comandos JS/TS. No comparte contenido con el BE.
- Ambos son independientes: mergear uno no requiere ni rompe al otro. Es razonable mergearlos juntos por consistencia de DX, pero no es obligatorio.

## Fuera de alcance
- Instalación automática de Lefthook (no hay hook de bootstrap en Makefile ni en `go.mod`/`tools.go`).
- Config de golangci-lint (`.golangci.yml`) — no se agrega.
- Cambios en CI (`ci-pr.yml` ya corre golangci-lint vía `golangci/golangci-lint-action@v9`; ver feature-020).

## Comportamiento esperado
- Con Lefthook instalado y `lefthook install` ejecutado: al hacer `git commit` con `.go` staged, corren en paralelo gofmt/vet/lint; al `git push`, corren tests cortos.
- Sin Lefthook instalado / sin `lefthook install`: el archivo es inerte (git no lo ejecuta). No rompe nada para quien no lo usa.

## Estado en dp~1 (SHA 777e5f6a)
- Archivo **presente y completo** en `777e5f6a` (SOURCE de extracción = `develop-problematico~1`).
- **Ausente en `develop`** (tip `003a9b8f`): `git cat-file -e develop:lefthook.yml` → "no existe". Por eso esta feature SÍ debe extraerse.
- No está marcada como YA PORTEADO.

## Criterios de aceptación
1. `lefthook.yml` existe en la raíz del repo BE, idéntico a `777e5f6a:lefthook.yml`.
2. `lefthook install` no falla; los hooks `pre-commit`/`pre-push` quedan registrados en `.git/hooks/`.
3. Un commit con un `.go` mal formateado es bloqueado por el comando `gofmt`.
4. El archivo no rompe el flujo de quien NO tiene Lefthook (es opt-in).
5. No se introduce ningún cambio en código fuente, deps, migraciones ni CI.

## Endpoints / Modelos / UI / DB / Tests afectados
- **Endpoints:** ninguno.
- **Modelos/DTOs:** ninguno.
- **UI:** ninguno (BE).
- **DB / migraciones:** ninguna.
- **Tests:** no agrega tests; el hook `pre-push` *invoca* `go test ./... -short` (depende de los tests existentes del repo, ver feature-025).

## Dependencias
- **Intra-repo:** ninguna dura. Débil/funcional con golangci-lint (Makefile usa `go run github.com/golangci/golangci-lint/v2/...@v2.11.4`) y con tests (feature-025).
- **Cross-repo:** ninguna dura. El FE tiene su propio `lefthook.yml` (mismo feature-022); coordinación cosmética, no técnica.
- **DEPENDE DE:** ninguna (declarado en el brief).

## Riesgos
- **Funcional:** bajo. Es tooling opt-in; no toca runtime.
- **Técnico (medio):** el hook llama `golangci-lint` como binario de sistema con flag `--fast`. El repo estandariza golangci-lint **v2** vía `go run ...@v2.11.4`. En golangci-lint v2 el flag `--fast` fue **eliminado/renombrado** (en v1 existía `--fast`); si el dev tiene v2 instalado, `--fast` puede fallar. Además, versión de sistema ≠ versión pineada del Makefile → resultados de lint inconsistentes con CI.
- **Técnico (bajo):** `--new-from-rev=HEAD` solo reporta issues nuevos respecto a HEAD; semántica correcta pero requiere git disponible.

## DECISIÓN recomendada
**Extraer tal cual** (cherry de archivo entero) — el archivo es inerte y opt-in, riesgo de merge ~nulo. **Pero** dejar anotada como mejora futura la corrección del comando `golangci-lint` para alinearlo con la versión v2 pineada del Makefile (`go run github.com/golangci/golangci-lint/v2/...@v2.11.4 run --new-from-rev=HEAD`, sin `--fast`). Esa corrección NO es bloqueante para mergear, porque el archivo no se ejecuta salvo que el dev haga `lefthook install`.
