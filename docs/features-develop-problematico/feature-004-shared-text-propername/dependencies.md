# dependencies — feature-004 · shared-text / proper-name normalization

## Depende de

- **Ninguna feature.** (La nota de la feature lo confirma: "DEPENDE DE: ninguna.")
- Dependencia de librería (ya satisfecha en develop): `golang.org/x/text/unicode/norm`.
  - go.mod develop (`003a9b8f`): `golang.org/x/text v0.37.0`.
  - go.mod source (`777e5f6a` / `0972e565`): `golang.org/x/text v0.36.0`.
  - La API usada (`norm.NFD.String`) es estable entre 0.36 y 0.37 → **no requiere tocar go.mod/go.sum**.

## Bloquea a

| feature | tipo de bloqueo | por qué |
|---|---|---|
| **007 actor-system** [BEFE] | **fuerte** | Normaliza nombres de actores; `internal/actor/handler/dto/actor.go` y `internal/actor/master_link.go` importan `internal/shared/text` en el source ref. Debe mergearse 004 antes que 007. |
| customer (DTOs) | fuerte | `internal/customer/handler/dto/requests.go` importa el paquete en el source ref. |
| **010 projects** [BEFE] / **011 campaign-dto-projectid** | fuerte | `internal/project/handler/dto/project.go` y `internal/project/repository.go` importan el paquete en el source ref. |

> Si cualquiera de esas features se portara antes que 004, los imports de `internal/shared/text` quedarían rotos y no compilaría. Por eso 004 va primero.

## Clasificación de fuerza

- **Fuertes:** el bloqueo hacia 007/010/011/customer (imports directos en source ref).
- **Débiles:** ninguna.
- **Inciertas:** ninguna. (Confianza alta: verificado con `git grep -ln "shared/text" 777e5f6a`.)

## Recursos compartidos

- Archivos compartidos: **ninguno** (no toca wire/cmd/config/go.mod/go.sum/Makefile/handlers/base.go/repository).
- Tipos compartidos: exporta `CanonicalizeName` y `FormatProperName` (funciones libres); cualquier feature posterior depende de esas firmas. Cambiar su semántica afectaría dedupe/display aguas abajo.
- Config / migraciones / APIs compartidas: ninguna.

## Cross-repo

- FE: **sin cambios** en esta feature. El helper `ui/src/lib/properName.ts` (repo `web`) es la referencia semántica espejo; no se modifica aquí. En el cross-repo-map del FE figura como "sin cambios FE".
- No hay PR de FE asociado a 004.

## Recomendación de orden

1. **feature-004 (este PR, BE)** — primero, temprano.
2. Luego 007 (actor-system) y los porteos de callers customer / 010 / 011, que ya pueden importar el paquete.

Mergear 004 solo (sin sus callers) es seguro: queda un paquete sin consumidores en develop, lo cual compila y testea sin problemas.
