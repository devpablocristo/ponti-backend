# notes-for-future-agent.md — feature-021 (BE)

## Resumen corto

feature-021 (BE) son 5 archivos: `Dockerfile`, `docker-compose.yml`, `.gitignore`, `go.mod`, `go.sum`. La intención (core→platform en build/deploy) **YA está en develop** por otra rama (commits `9a5e465b`, `e7bb89d8`, `93f77883`). Lo que falta NO es config pura: es la **consecuencia de deps** de features de código aún no portadas. Por eso la decisión es **partir + postergar**: portar 2 hunks chicos hoy, cerrar `go.mod`/`go.sum` con `go mod tidy` después.

## Qué está en FE y qué en BE

- **BE-021 (este paquete):** Dockerfile / docker-compose / .gitignore / go.mod / go.sum.
- **FE-021 (espejo):** vite / tailwind / eslint / knip / tsconfig / lockfiles / generated client. **Sin archivos compartidos con BE.** Coordinación de orden, no de contenido.

## Archivos esenciales / peligrosos / mezclados

- **Peligrosos (NO portar como config):** `go.mod`, `go.sum`. Su delta arrastra deps de 001/002/005 y borra excelize/cloudsqlconn que develop usa. Regenerar con `go mod tidy`, nunca a mano.
- **Mezclados:** `.gitignore` (mezcla 021 + 019 + tooling local `go.work`). `Dockerfile` (mezcla 021 + 005/020 prefetch).
- **Esenciales y portables HOY:** hunk de `docker-compose.yml` (quitar mount `core` + `GOWORK=off`, agregar NOTA) y hunk de `.gitignore` (ignorar `scripts/db/schema.snapshot.sql`).

## Decisiones ya tomadas

- NO tocar `go.mod`/`go.sum` en el PR de 021.
- Excluir bumps go-jose/x/net (DONE #124).
- Diferir el agregado de `observability/go`+`persistence/gorm/go` al prefetch del `Dockerfile` hasta que 001/005 estén en develop.
- Usar `git restore -p` (hunk a hunk), nunca `git checkout whole-file`, para compose y gitignore.

## Dudas abiertas

- ¿feature-013 (csv-export) mantiene `excelize/v2` o migra a CSV puro? Determina si el `go.mod` final conserva excelize. Revisar el paquete de feature-013.
- Estado real del FE-021 (no inspeccionado desde acá).

## Comandos para mirar primero

```bash
git -C <repo> merge-base develop 777e5f6a          # == 0972e565 (divergencia)
git -C <repo> diff 0972e565..777e5f6a -- docker-compose.yml .gitignore
diff <(git show develop:go.mod) <(git show 777e5f6a:go.mod)   # ver el delta real vs develop
git -C <repo> grep -l "platform/observability/go" 777e5f6a -- '*.go'   # prueba de acoplamiento a código
git -C <repo> grep -l "excelize" develop -- '*.go'             # develop aún usa excelize
```

## Errores a evitar

- Hacer `git checkout 777e5f6a -- go.mod go.sum` → build roto (tidy borra/falta deps).
- Tomar `.gitignore` whole-file → se pierde el ignore de `go.work` (tooling 019).
- Re-aplicar bumps go-jose/x/net (#124).
- Asumir que `develop-problematico` (tip) sirve: es un restore vacío. Usar SIEMPRE `develop-problematico~1` = `777e5f6a`.

## Camino más seguro

1. Rama desde develop.
2. `git restore -p --source=develop-problematico~1 -- docker-compose.yml .gitignore` eligiendo SÓLO los hunks portables.
3. `go build ./... && go test ./... && docker compose build`.
4. PR que toca SÓLO esos 2 archivos.
5. El cierre de `go.mod`/`go.sum` y del prefetch del Dockerfile va en un PR posterior, tras mergear 001/002/005/013/023, vía `go mod tidy`.

## PR del otro repo: orden

- FE-021 y BE-021 son **independientes**; cualquiera puede ir primero. No hay bloqueo técnico. Sólo trackearlos bajo el mismo feature-021 para el release notes.
