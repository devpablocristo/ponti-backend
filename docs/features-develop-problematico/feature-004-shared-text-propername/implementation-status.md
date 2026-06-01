# implementation-status — feature-004 · shared-text / proper-name normalization

## Estado general

| dimensión | valor |
|---|---|
| estado | **completa** (en el source ref `777e5f6a`) |
| % completitud | ~100% para el alcance del paquete |
| confianza | alta |

## Estado en este repo (BE)

- `internal/shared/text/propername.go` — completo. Implementa:
  - `CanonicalizeName(string) string`
  - `FormatProperName(string) string`
  - privados: `stripDiacriticsPreservingEnye`, `formatWord`
  - tablas: `connectors` (18 conectores ES), `uppercaseTokens` (sufijos societarios + siglas agro AR).
- `internal/shared/text/propername_test.go` — completo. Table-driven:
  - `TestCanonicalizeName` (11 casos)
  - `TestFormatProperName` (15 casos)
- En develop (`003a9b8f`) el paquete **no existe todavía** → entra como creación pura.

## Estado en el otro repo (FE)

- **Sin cambios FE.** `ui/src/lib/properName.ts` ya existe en el repo `web` y es la referencia de paridad semántica. No se toca en 004.

## Tests

- Cobertura: ambas funciones públicas cubiertas con tablas de casos representativos (acentos, ñ, siglas societarias, conectores, números/guiones, dobles espacios, vacío, solo-separadores).
- No verificados en ejecución por este agente (solo lectura). Sugerido: `go test ./internal/shared/text/...` (ver validation.md).
- Falta cubrir (mejora futura, no bloqueante): tokens mixtos no contemplados, entrada con dígitos pegados a letras dentro de un token uppercase, comportamiento con runas fuera de Latin (ya colapsan a espacio por diseño).

## Pendientes

- Cablear los callers (NO en 004): actor/customer/project. Pertenece a 007/010/011/customer.

## Clasificación de issues

### BLOQUEANTE para mergear
- Ninguno. El paquete compila y testea aislado; sin dependencias intra-repo; `x/text` ya en develop.

### Mejora futura
- Ampliar `uppercaseTokens`/`connectors` si negocio detecta más siglas/conectores.
- Tests adicionales de borde (ver arriba).
- Considerar exponer una variante que no colapse guiones si algún caller lo necesitara (hoy `25-26` → `25 26`).

### Deuda aceptable
- El paquete queda sin callers en develop hasta que llegue 007. Es esperado.
- Acoplamiento implícito de paridad con `properName.ts` del FE: cambios futuros deben hacerse en ambos repos a mano (no hay test cross-repo que lo garantice).

### Duda humana
- ¿La lista de `uppercaseTokens` (incluye `inta/ypf/afip/arba`) y `connectors` es la versión definitiva acordada con negocio/FE? Verificar paridad exacta con `ui/src/lib/properName.ts` en el repo `web` antes de dar por cerrada la semántica. (No bloquea el merge del util, sí la consistencia BE/FE.)
