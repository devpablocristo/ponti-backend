# validation — feature-004 · shared-text / proper-name normalization

## Checklist pre-PR (BE)

- [ ] `git status` muestra **exactamente** 2 archivos nuevos: `internal/shared/text/propername.go` y `internal/shared/text/propername_test.go`.
- [ ] `go.mod` y `go.sum` **sin cambios** (x/text sigue en `v0.37.0`).
- [ ] Ningún caller (`internal/actor/**`, `internal/customer/**`, `internal/project/**`) modificado.
- [ ] Build del paquete:
  - `go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./internal/shared/text/...`
- [ ] Tests del paquete (todas las filas deben pasar):
  - `go -C /home/pablocristo/Proyectos/pablo/ponti/core test ./internal/shared/text/...`
  - con verbose: `go -C /home/pablocristo/Proyectos/pablo/ponti/core test -v ./internal/shared/text/...`
- [ ] Sanity global (no debe romper nada; paquete nuevo sin callers en develop):
  - `go -C /home/pablocristo/Proyectos/pablo/ponti/core build ./...`
  - `go -C /home/pablocristo/Proyectos/pablo/ponti/core vet ./internal/shared/text/...`
- [ ] `git diff --check` sin warnings de whitespace.

## Verificación manual (comportamiento clave)

Comparar salida contra estas tablas (idénticas a `propername_test.go`):

`CanonicalizeName`:
- `"  AGRO LAJITAS  "` → `"agro lajitas"`
- `"María Ángeles"` → `"maria angeles"`
- `"EL SUEÑO"` → `"el sueño"`
- `"ÑANDÚ"` → `"ñandu"`
- `"AGRO LAJITAS S.R.L."` → `"agro lajitas s r l"`
- `"JIMENES 25-26"` → `"jimenes 25 26"`
- `"///"` → `""`, `""` → `""`

`FormatProperName`:
- `"AGRO LAJITAS SRL"` → `"Agro Lajitas SRL"`
- `"juan de la torre"` → `"Juan de la Torre"`
- `"y griega"` → `"Y Griega"`
- `"inta pergamino"` → `"INTA Pergamino"`
- `"soalen sa"` → `"Soalen SA"`
- `"LOTE 1"` → `"Lote 1"`

## Tests sugeridos

- BE: `go test ./internal/shared/text/...` (ya incluido en el extract). Opcional añadir `-run TestCanonicalizeName` / `-run TestFormatProperName`.
- FE: **no aplica** (sin cambios FE). No correr yarn test/build/e2e para esta feature.

## Casos borde a revisar

- Entrada con tilde combinante sobre n/N → debe dar ñ/Ñ (cubierto por `EL SUEÑO`/`ÑANDÚ`).
- Solo separadores (`"///"`, `"   "`) → cadena vacía.
- Conector como primera palabra (`"y griega"`) → se capitaliza (`"Y Griega"`).
- Token uppercase combinado con conector (`"perez y gomez sh"`) → `"Perez y Gomez SH"`.
- Números y guiones → guiones colapsan a espacio (`"25-26"` → `"25 26"`).

## Qué revisar en UI / API / DB / env

- UI: nada (FE sin cambios).
- API: nada (sin endpoints).
- DB: nada (sin migraciones).
- Env: nada.

## Qué validar en el otro repo

- Nada que portar. Único chequeo recomendado (consistencia, no bloqueante): que `ui/src/lib/properName.ts` (repo `web`) produzca las mismas salidas que la tabla de arriba, para garantizar paridad BE↔FE. Si difiere, es señal de divergencia semántica a resolver (no en este PR).

## Señales de incompletitud / incompatibilidad

- Falla de compilación `package text` → falta archivo o import `golang.org/x/text/unicode/norm`.
- Falla de test en alguna fila → el contenido extraído difiere del source ref; comparar con `git diff 0972e565..777e5f6a -- internal/shared/text/`.
- `go.mod` modificado tras `go mod tidy` (baja x/text) → revertir, no es necesario.
- Aparecen cambios en `internal/actor`/`internal/customer`/`internal/project` → se arrastró un caller por error; quitarlos del PR.
