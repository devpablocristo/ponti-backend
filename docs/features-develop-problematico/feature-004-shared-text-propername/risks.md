# risks — feature-004 · shared-text / proper-name normalization

Riesgo global: **bajo**. Util puro, autocontenido, con tests. A continuación por categoría con mitigaciones concretas.

## Funcionales

- **Set de reglas de negocio embebido** (`connectors` y `uppercaseTokens`). Si la lista no coincide con lo que negocio/FE esperan, el display saldrá distinto (p.ej. una sigla societaria nueva no quedaría en mayúscula, o un conector no contemplado se title-caseará).
  - Mitigación: cotejar `connectors`/`uppercaseTokens` contra `ui/src/lib/properName.ts` (repo `web`) antes de cerrar. Los tests fijan el contrato actual; cualquier cambio debe actualizar la tabla de tests.
- **Colapso de guiones y puntos**: `"25-26"` → `"25 26"`, `"S.R.L."` → `"s r l"`. Es por diseño (cualquier no-`[a-z0-9ñ]` colapsa a espacio). Riesgo: si algún caller esperaba conservar guiones (rangos de campaña, lotes con guion), perdería información.
  - Mitigación: revisar en 007/010/011 que ningún campo dependa de conservar separadores; documentado en spec.md (comportamiento esperado).

## Técnicos

- **Dependencia `golang.org/x/text/unicode/norm`**: cambió 0.36 (source) → 0.37 (develop). API usada (`norm.NFD.String`, `unicode.Is(unicode.Mn, r)`, mapeo manual de `̃` a ñ) es estable.
  - Mitigación: tras extraer, `go build`/`go test` del paquete; NO correr `go mod tidy` que baje la versión. Confirmar `go.sum` sin cambios.
- **Manejo de ñ vía U+0303 (combining tilde)**: el helper reconstruye `n+◌̃` → `ñ` y `N+◌̃` → `Ñ` durante NFD. Riesgo de borde con secuencias raras (tilde combinante sin base n/N) — esos casos caen como Mn y se descartan, comportamiento aceptable.
  - Mitigación: cubierto por casos `EL SUEÑO` / `ÑANDÚ` en tests.

## Integración

- El paquete entra **sin callers** en develop hasta 007. No rompe nada, pero un linter estricto podría marcar paquete no usado.
  - Mitigación: aceptar; los tests del propio paquete lo ejercitan, no es "dead code" real.

## Cross-repo

- Paridad BE↔FE solo por convención (doc-comment). No hay verificación automática de que `properName.ts` y `propername.go` produzcan idénticos resultados.
  - Mitigación: si negocio exige paridad estricta, agregar (futuro) un set de fixtures compartido. Fuera de alcance de 004.

## Datos / migración

- Ninguno. No hay migraciones, no se reescriben datos. (OJO: cuando 007/customer empiecen a **persistir** la forma canónica como clave de dedupe, ahí sí habrá implicancias de datos — pero eso es de esas features, no de 004.)

## Archivos compartidos

- Ninguno tocado. Cero riesgo de hunks mezclados, cero conflicto con wire/cmd/config/go.mod.

## Extracción parcial

- Riesgo mínimo: solo 2 archivos whole-file. Señal de incompletitud = `go build ./internal/shared/text/...` falla (falta archivo o import).

## Riesgo de mergear solo este repo / solo el otro

- **Solo BE (este PR):** seguro. Paquete huérfano hasta 007; compila y testea. Recomendado.
- **Solo FE:** no aplica — FE no tiene cambios en 004.
- **Riesgo inverso:** mergear 007/010/011/customer ANTES que 004 rompería la compilación (imports a `internal/shared/text` inexistentes). Mitigación: respetar orden (004 primero).
