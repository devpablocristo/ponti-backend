# notes-for-future-agent — feature-004 · shared-text / proper-name normalization

## Resumen corto

Util Go nuevo y minúsculo: paquete `internal/shared/text` con `CanonicalizeName` (forma de almacenamiento) y `FormatProperName` (forma de display) para nombres de entidades. Dos archivos, ambos `A` (created), whole-file, con tests. Extracción de riesgo bajo. Es lo primero que debería mergearse de su cadena.

## Qué está en FE y qué en BE

- **BE:** todo (los 2 archivos del flist). Único repo con cambios en 004.
- **FE:** sin cambios. Existe `ui/src/lib/properName.ts` en el repo `web` como referencia semántica espejo, pero NO se toca en esta feature. En el cross-repo-map del FE va como "sin cambios FE".

## Archivos

- Esenciales: `internal/shared/text/propername.go` (impl), `internal/shared/text/propername_test.go` (contrato).
- Peligrosos / mezclados: **ninguno**. No hay archivos compartidos (no toca wire/cmd/config/go.mod/go.sum).
- NO traer (callers de otras features, presentes en source ref `777e5f6a`): `internal/actor/handler/dto/actor.go`, `internal/actor/master_link.go`, `internal/customer/handler/dto/requests.go`, `internal/project/handler/dto/project.go`, `internal/project/repository.go`.

## Decisiones ya tomadas

- Extraer **tal cual / whole-file** como PR BE independiente (`pr/feature-004-shared-text-propername-be`).
- NO tocar go.mod/go.sum: `golang.org/x/text` ya está en develop (`v0.37.0`; source usaba `v0.36.0`, API estable).
- SOURCE = `develop-problematico~1` (SHA `777e5f6a`). Nunca usar el tip `develop-problematico` (restore/vacío).

## Dudas abiertas

- Paridad exacta de `connectors` y `uppercaseTokens` con `properName.ts` del FE. No bloquea el merge del util, sí la consistencia BE/FE a largo plazo. Verificar manualmente si negocio lo exige.
- Decisión de colapsar guiones/puntos a espacio (`25-26`→`25 26`): confirmar que ningún caller de 007/010/011 dependa de conservarlos.

## Comandos a mirar primero (read-only)

```
cat /tmp/flists/be-004.txt
git -C /home/pablocristo/Proyectos/pablo/ponti/core show 777e5f6a:internal/shared/text/propername.go
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff 0972e565..777e5f6a -- internal/shared/text/
git -C /home/pablocristo/Proyectos/pablo/ponti/core grep -ln "shared/text" 777e5f6a   # ver callers (NO extraer aquí)
git -C /home/pablocristo/Proyectos/pablo/ponti/core ls-tree -r 003a9b8f --name-only -- internal/shared/text   # vacío en develop
```

## Errores a evitar

- Arrastrar callers (`internal/actor|customer|project`) al PR de 004 → no compilaría aislado y mezcla features.
- Correr `go mod tidy` y bajar `x/text` por inercia → dejar `v0.37.0`.
- Confundir `develop-problematico` (tip vacío) con el SOURCE correcto (`develop-problematico~1` = `777e5f6a`).

## Camino más seguro

1. Branch desde develop → `git checkout 777e5f6a -- internal/shared/text/propername.go internal/shared/text/propername_test.go`.
2. `go build ./internal/shared/text/...` + `go test ./internal/shared/text/...`.
3. Confirmar `git status` = solo 2 archivos, go.mod/go.sum intactos.
4. PR a develop.

## Qué PR del otro repo va antes/después

- Antes: nada (004 es raíz de su cadena, sin dependencias).
- Después: feature-007 (actor-system) y los porteos de callers de customer/010/011, que importan este paquete. 004 debe mergear ANTES que todos ellos para no romper imports.
- FE: ningún PR de FE asociado.
