# spec — feature-004 · shared-text / proper-name normalization

| campo | valor |
|---|---|
| id | feature-004 |
| slug | shared-text-propername |
| nombre | Shared text / proper-name normalization |
| tipo | feature (utilidad compartida BE) |
| repo | Backend Go — `ponti-backend` (`/home/pablocristo/Proyectos/pablo/ponti/core`) |
| existe en BE | Sí (paquete nuevo `internal/shared/text`) |
| existe en FE | No hay carpeta espejo en este repo. El doc-comment del Go referencia `ui/src/lib/properName.ts`, que vive en el repo FE separado (`web`). En el cross-repo-map del FE figura como "sin cambios FE" para esta feature. |
| merge | BE independiente (no depende de ninguna otra feature) |
| source ref | `develop-problematico~1` (SHA `777e5f6a`) — NUNCA usar `develop-problematico` (tip = restore/vacío) |
| rango diff (fuente de verdad) | `0972e565..777e5f6a` |
| rama destino | `develop` (tip `003a9b8f`) |

## Resumen

Paquete utilitario Go nuevo y autocontenido (`internal/shared/text`) con dos funciones puras para normalizar nombres propios de entidades (cliente, proyecto, manager, inversor, campo, lote, cultivo, temporada, display name de actor):

- `CanonicalizeName(value string) string` — forma de **almacenamiento**: minúsculas en español, solo `[a-z0-9ñ ]`, despoja diacríticos (preservando ñ/Ñ), colapsa cualquier otro carácter a un solo espacio, colapsa espacios y hace trim.
- `FormatProperName(value string) string` — forma de **display**: canonicaliza, luego title-case por palabra, con excepciones: conectores españoles (`de`, `del`, `con`, ...) en minúscula salvo que sean la primera palabra, y tokens en mayúscula fija (sufijos societarios `srl/sa/sas/...` y siglas agro argentinas `inta/ypf/afip/arba`).

## Objetivo

Tener una única fuente de verdad en BE para canonicalizar (clave de almacenamiento/deduplicación) y formatear (presentación) nombres de entidades, con semántica idéntica al helper FE `properName.ts`. Es la base que consumen otras features (notablemente 007 actor-system para normalizar nombres de actores; también customer/project en el árbol del source ref).

## Problema

Sin este util, cada módulo (customer, project, actor) implementaría su propia normalización ad-hoc, con riesgo de:
- claves de deduplicación inconsistentes (mismo nombre escrito distinto se guarda distinto),
- display inconsistente entre BE y FE,
- manejo divergente de ñ, acentos, siglas societarias y conectores.

## Alcance en este repo (BE)

- Crear el paquete `internal/shared/text` con `propername.go` y `propername_test.go`.
- Nada más. No toca routers, wire, main, config, migraciones ni go.mod/go.sum.

## Alcance en el otro repo (FE)

- **Sin cambios FE en esta feature.** El archivo `ui/src/lib/properName.ts` ya existe en el FE y es la referencia semántica; no se modifica como parte de feature-004. Solo se menciona en el cross-repo-map del FE como "sin cambios FE".

## Fuera de alcance

- Cualquier cableado de llamadas a estas funciones desde repositorios/handlers (eso pertenece a 007/010/011/customer y se porta con esas features).
- Cambios en `go.mod`/`go.sum` (la dependencia `golang.org/x/text` ya está presente en develop, ver abajo).
- Lógica de matching/fuzzy o linking de actores (master_link) — pertenece a 007.

## Comportamiento esperado (extraído de los tests `propername_test.go`)

`CanonicalizeName`:
- `"  AGRO LAJITAS  "` → `"agro lajitas"`
- `"María Ángeles"` → `"maria angeles"`
- `"EL SUEÑO"` → `"el sueño"` (preserva ñ)
- `"ÑANDÚ"` → `"ñandu"`
- `"AGRO LAJITAS S.R.L."` → `"agro lajitas s r l"`
- `"JIMENES 25-26"` → `"jimenes 25 26"`
- `"E.VEDOYA"` → `"e vedoya"`; `"J.M. PEREZ"` → `"j m perez"`
- `"  doble   espacio  "` → `"doble espacio"`
- `""` → `""`; `"///"` → `""`

`FormatProperName`:
- `"agro lajitas srl"` / `"AGRO LAJITAS SRL"` → `"Agro Lajitas SRL"`
- `"juan de la torre"` → `"Juan de la Torre"`
- `"y griega"` → `"Y Griega"` (conector como primera palabra se capitaliza)
- `"María Ángeles"` → `"Maria Angeles"`; `"EL SUEÑO"` → `"El Sueño"`; `"ÑANDÚ"` → `"Ñandu"`
- `"soalen sa"` → `"Soalen SA"`; `"perez y gomez sh"` → `"Perez y Gomez SH"`
- `"inta pergamino"` → `"INTA Pergamino"`
- `"LOTE 1"` → `"Lote 1"`; `""` → `""`

## Estado en dp~1 (SHA 777e5f6a)

- Paquete **completo y con tests** (191 líneas creadas: 134 código + 57 test).
- En el source ref ya hay consumidores del paquete: `internal/actor/handler/dto/actor.go`, `internal/actor/master_link.go`, `internal/customer/handler/dto/requests.go`, `internal/project/handler/dto/project.go`, `internal/project/repository.go`. Esos consumidores NO son parte de feature-004 (pertenecen a 007/010/011/customer).

## Criterios de aceptación

1. Existen `internal/shared/text/propername.go` y `internal/shared/text/propername_test.go` idénticos al source ref.
2. `go build ./internal/shared/text/...` compila en develop.
3. `go test ./internal/shared/text/...` pasa (todas las filas de ambas tablas de test).
4. No se modificó ningún otro archivo (ni go.mod, ni callers, ni config).

## Endpoints / modelos / UI / DB / tests afectados

- Endpoints: ninguno (utilidad pura).
- Modelos/DTOs/tipos: define solo funciones libres `CanonicalizeName` y `FormatProperName`; no exporta tipos.
- UI: ninguna (FE sin cambios).
- DB / migraciones: ninguna.
- Tests: `internal/shared/text/propername_test.go` (table-driven, 2 funciones de test).

## Dependencias

- Intra-repo: **ninguna**. No depende de otras features.
- Dependencia de librería: `golang.org/x/text/unicode/norm` — **ya presente** en develop (`go.mod` develop: `golang.org/x/text v0.37.0`; en source ref era `v0.36.0`). No requiere tocar go.mod/go.sum.
- Cross-repo: ninguna (FE sin cambios).
- **Bloquea a:** feature-007 (actor-system) que normaliza nombres de actores, y de hecho cualquier caller que lo consuma (customer/project en el árbol del source ref).

## Riesgos

- Funcional: bajo. Funciones puras, deterministas, con tests que fijan el contrato. El único matiz de comportamiento es el set de `connectors` y `uppercaseTokens` (decisiones de negocio sobre conectores españoles y siglas argentinas) — ver risks.md.
- Técnico: muy bajo. Solo stdlib + `golang.org/x/text/unicode/norm`. La versión de `x/text` cambió de 0.36 → 0.37 entre source y develop; la API `norm.NFD.String` y `unicode.Is(unicode.Mn, r)` es estable.
- Extracción: muy bajo. Dos archivos nuevos, whole-file, sin hunks compartidos.

## DECISIÓN recomendada

**Extraer tal cual (whole-file), como PR BE independiente y temprano.** Es un util chico, autocontenido, con tests, sin dependencias intra-repo y sin tocar archivos compartidos. Debe mergearse antes que feature-007 (y antes de portar callers de customer/project) para no dejar imports rotos. No requiere arreglos previos ni partición.
