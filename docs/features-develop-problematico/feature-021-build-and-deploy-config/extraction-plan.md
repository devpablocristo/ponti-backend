# extraction-plan.md — feature-021 (BE)

- **repo:** `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`)
- **rama base:** `develop` (tip `003a9b8f`)
- **SOURCE:** `develop-problematico~1` = SHA `777e5f6a` (NUNCA `develop-problematico`)
- **rama sugerida:** `pr/feature-021-build-and-deploy-config-be`
- **merge-base develop↔target:** `0972e565` (divergencia real; `777e5f6a` no es ancestro de develop)

## PR title

`chore(be): alinear config de build/deploy con platform (feature-021)`

## PR description (sugerida)

> Parte BE de feature-021 (build & deploy config). develop ya migró `core→platform`
> de forma independiente; este PR sólo aplica los deltas de config que faltan y que
> NO dependen de código aún no portado:
> - `docker-compose.yml`: remueve el mount local `${CORE_REPO_DIR}:/home/pablo/Projects/Pablo/core` y `GOWORK=off`; agrega NOTA explicando la resolución vía proxy de módulos.
> - `.gitignore`: ignora `scripts/db/schema.snapshot.sql` (generado por `make db-schema-snapshot`, coord feature-019).
>
> NO incluye cambios en `go.mod`/`go.sum` (se regeneran con `go mod tidy` al portar
> 001/002/005/013/023) ni los bumps go-jose/x/net (ya en #124). El agregado de
> `observability/go`+`persistence/gorm/go` al prefetch del `Dockerfile` queda diferido
> hasta que aterricen esas features de código.

## Pasos ordenados

1. Verificar que develop ya está en `platform/*` (lo está) y que `go build ./...` compila en develop antes de tocar nada.
2. Crear rama desde develop.
3. Aplicar SÓLO los hunks portables de `docker-compose.yml` y `.gitignore` (partial).
4. NO tocar `go.mod` / `go.sum`. NO tocar `Dockerfile` (salvo que ya estén en develop 001/005; ver "qué NO traer").
5. `go build ./... && go test ./...` + `docker compose build`.
6. `git diff --check` (whitespace) y abrir PR.

## Archivos enteros vs parciales

- **Parciales (partial-hunks):** `docker-compose.yml`, `.gitignore`. NO usar checkout whole-file: traería de vuelta el revert de `go.work`/tooling local.
- **Enteros:** ninguno.
- **Diferido:** `Dockerfile` (manual-port condicionado), `go.mod`, `go.sum`.

## Migraciones / tests a incluir

Ninguna migración. Ningún test propio. (Tests de develop deben seguir verdes; no son parte de 021.)

## Dependencias previas

- Para `go.mod`/`go.sum`: deben estar mergeadas 001 (platform-tenancy / persistence-gorm), 002 (lifecycle metrics → prometheus), 005/020 (observability), y 013 debe definir si excelize se queda o se va. Hasta entonces, `go mod tidy` es la fuente de verdad, no el diff de dp~1.

## Coordinación con el otro repo (FE-021)

- **Independiente en archivos.** No hay archivo compartido. Orden de merge libre; sugerencia **BE-first** sólo por consistencia de naming de PR. No es bloqueante.

## Comandos git SUGERIDOS (para un humano; NO ejecutar acá)

```bash
# 0) partir de develop limpio
git checkout develop
git pull
git checkout -b pr/feature-021-build-and-deploy-config-be

# 1) traer SÓLO los hunks portables (interactivo, elegir hunks correctos)
git restore -p --source=develop-problematico~1 -- docker-compose.yml
#   -> aceptar: quitar mount core + GOWORK=off + agregar NOTA
#   -> RECHAZAR cualquier hunk que toque otra cosa
git restore -p --source=develop-problematico~1 -- .gitignore
#   -> aceptar SÓLO el hunk que agrega: scripts/db/schema.snapshot.sql
#   -> RECHAZAR los hunks que borran go.work / go.work.sum / /api / scripts/db/*.env

# 2) NO traer go.mod / go.sum / Dockerfile (deps las maneja tidy)
#    Si en el futuro 001/005 ya están en develop:
# git restore -p --source=develop-problematico~1 -- Dockerfile   # sólo el hunk del prefetch

# 3) sanity
git diff --check
go build ./...
go test ./...
docker compose build

# 4) inspección sólo-lectura útil
git show 777e5f6a:docker-compose.yml | head -60
git diff 0972e565..777e5f6a -- .gitignore
```

## Qué NO traer

- `go.mod`, `go.sum` (regenerar con `go mod tidy`).
- Hunks de `.gitignore` que borran `go.work`/`go.work.sum`/`/api`/`scripts/db/*.env`.
- Bumps go-jose/x/net (#124).
- Agregado de `observability/go`/`persistence/gorm/go` al `Dockerfile` mientras su código no esté en develop.

## Qué podría romperse

- Si alguien hace `git checkout 777e5f6a -- go.mod go.sum` whole-file: `go mod tidy` borrará `observability`/`persistence`/`prometheus` (sin código que las use) y `go build` fallará por falta de `excelize`/`cloudsqlconn` (código de develop). Build roto.
- Whole-file de `.gitignore`: deja de ignorar `go.work` y vuelve a ignorar `scripts/db/*.env` → contamina el repo / oculta tooling de 019.

## Cómo detectar extracción incompleta

- `go.mod` cambiado en el PR ⇒ señal de que se mezcló deps; revertir.
- `git grep -n "CORE_REPO_DIR\|/home/pablo/Projects/Pablo/core" docker-compose.yml` debe dar vacío tras el port.
- `go.work` debe seguir en `.gitignore`.

## Qué validar antes del PR

- `go build ./... && go test ./...` verdes (sin tocar go.mod).
- `docker compose build` ok; `docker compose up` levanta el servicio.
- `git diff --stat` del PR sólo toca `docker-compose.yml` y `.gitignore` (y, si aplica, `Dockerfile`).

## Qué hacer después de mergear

- Cuando aterricen 001/002/005/013/023: correr `go mod tidy` y `go mod verify`, commitear `go.mod`/`go.sum` resultantes (ahí, y sólo ahí, aparecen observability/persistence/prometheus y desaparece excelize). Agregar entonces el hunk del prefetch del `Dockerfile`.
