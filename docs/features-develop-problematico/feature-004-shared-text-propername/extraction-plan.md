# extraction-plan — feature-004 · shared-text / proper-name normalization

## Contexto

| campo | valor |
|---|---|
| repo | `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`) |
| rama base | `develop` (tip `003a9b8f`) |
| SOURCE | `develop-problematico~1` (SHA `777e5f6a`) — NUNCA `develop-problematico` (tip vacío/restore) |
| rango diff fuente | `0972e565..777e5f6a` |
| rama sugerida | `pr/feature-004-shared-text-propername-be` |
| coordinación | BE-independiente. Sin coordinación cross-repo (FE sin cambios). Debe ir **antes** que feature-007 y antes de portar callers customer/project. |

## PR title

`feat(be): paquete internal/shared/text para canonicalización y display de nombres propios`

## PR description (sugerida)

```
Agrega el paquete utilitario internal/shared/text con dos funciones puras:

- CanonicalizeName: forma de almacenamiento (minúsculas español, [a-z0-9ñ ],
  despoja diacríticos preservando ñ/Ñ, colapsa separadores y espacios).
- FormatProperName: forma de display (title-case con conectores españoles en
  minúscula salvo primera palabra, y tokens en mayúscula fija: SRL/SA/SAS/...,
  INTA/YPF/AFIP/ARBA).

Semántica alineada con el helper FE ui/src/lib/properName.ts (repo web, sin
cambios en este PR). Incluye tests table-driven que fijan el contrato.

Base para la normalización de nombres de actores (feature-007) y consumidores
de customer/project.

Sin cambios en go.mod/go.sum: golang.org/x/text ya está presente en develop.
Sin endpoints, sin migraciones, sin cambios FE.
```

## Pasos ordenados

1. Partir de develop limpio:
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout develop`
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core pull` (si corresponde)
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout -b pr/feature-004-shared-text-propername-be`
2. Traer los dos archivos enteros desde el source ref:
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout 777e5f6a -- internal/shared/text/propername.go internal/shared/text/propername_test.go`
3. Verificar que NO se arrastró nada más:
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core status`  (esperado: solo los 2 archivos nuevos staged)
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core diff --check`
4. Compilar y testear (ver validation.md):
   - `go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./internal/shared/text/...`
   - `go -C /home/pablocristo/Proyectos/pablo/ponti/core test ./internal/shared/text/...`
   - opcional, sanity global: `go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./...` (no debería romper nada: el paquete es nuevo y nadie en develop lo importa todavía).
5. Commit y push:
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core add internal/shared/text/`
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core commit -m "feat(be): paquete internal/shared/text para nombres propios"`
   - `git -C /home/pablocristo/Proyectos/pablo/ponti/core push -u origin pr/feature-004-shared-text-propername-be`
6. Abrir PR hacia `develop`.

> Todos los comandos git de arriba son **SUGERENCIAS para un humano**. Este agente no ejecuta mutaciones.

## Archivos enteros vs parciales

- Enteros (whole-file): `internal/shared/text/propername.go`, `internal/shared/text/propername_test.go`.
- Parciales (partial-hunks): **ninguno**.

## Migraciones / tests a incluir

- Migraciones: ninguna.
- Tests: incluir `propername_test.go` (es parte del whole-file extract).

## Dependencias previas

- Ninguna feature previa. La librería `golang.org/x/text` ya está en develop (`v0.37.0`); no hace falta `go mod tidy` salvo sanity.

## Coordinación con el otro repo

- **No aplica** para esta feature (FE sin cambios). No hay PR de FE asociado a 004. (El `properName.ts` del repo FE ya existe; cualquier cambio futuro de paridad sería otra tarea.)

## Comandos git SUGERIDOS (solo lectura para inspección)

```
# inspección (read-only)
git -C /home/pablocristo/Proyectos/pablo/ponti/core show 777e5f6a:internal/shared/text/propername.go
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff 0972e565..777e5f6a -- internal/shared/text/
git -C /home/pablocristo/Proyectos/pablo/ponti/core grep -ln "shared/text" 777e5f6a   # ver callers (NO traer)

# extracción (mutación — ejecuta un humano)
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout develop
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout -b pr/feature-004-shared-text-propername-be
git -C /home/pablocristo/Proyectos/pablo/ponti/core checkout 777e5f6a -- internal/shared/text/propername.go internal/shared/text/propername_test.go
git -C /home/pablocristo/Proyectos/pablo/ponti/core diff --check
```

`git restore -p` NO es necesario: no hay archivos mixtos.

## Qué NO traer

- Ningún caller: `internal/actor/**`, `internal/customer/**`, `internal/project/**` (van con 007/010/011/customer).
- No tocar `go.mod`/`go.sum` (x/text ya presente). Si por hábito se corre `go mod tidy`, revisar que NO cambie versiones (debe quedar `v0.37.0`).

## Qué podría romperse

- Nada en develop hoy: ningún archivo de develop importa `internal/shared/text`. El paquete entra "huérfano" (sin callers) hasta que llegue 007 — eso es esperado y correcto.
- Riesgo solo si `go vet`/linter del CI marca paquete sin uso externo; no es error de compilación.

## Cómo detectar extracción incompleta

- `go build ./internal/shared/text/...` falla → falta un archivo o el import de `x/text`.
- `go test ./internal/shared/text/...` falla en alguna fila → el contrato cambió respecto del source ref (no debería; comparar con `git diff 0972e565..777e5f6a`).

## Qué validar antes del PR

- `git status` muestra exactamente 2 archivos nuevos.
- `go build` y `go test` del paquete OK.
- `go.mod`/`go.sum` SIN cambios.

## Qué hacer después de mergear

- Habilitar feature-007 (y porteo de callers customer/project) para que el paquete tenga consumidores. Confirmar que esos PRs importan `github.com/devpablocristo/ponti-backend/internal/shared/text`.
