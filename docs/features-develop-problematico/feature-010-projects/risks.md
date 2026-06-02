# risks.md — feature-010 projects (BE)

## Funcionales
- **Cambio de contrato HTTP (alto):** `DELETE /api/v1/projects/:id` → `DELETE /api/v1/projects/:id/hard`. Un FE que siga llamando al path viejo recibe 404. **Mitigación:** BE-first + coordinar el port del path en FE-010 en el mismo tren; opcionalmente dejar un alias temporal (no presente en el SOURCE — evaluar).
- **Hard-delete ahora bloquea (medio):** antes borraba en cascada; ahora exige proyecto archivado (409 "must be archived before hard delete") y sin dependientes activos (409 con conteo). El FE/usuarios que esperaban borrado directo cambian de flujo. **Mitigación:** documentar los 409 y reflejarlos en UI.
- **Canonicalización de nombres (bajo/medio):** `CanonicalizeName` puede alterar nombres que antes pasaban con sólo `TrimSpace` (mayúsculas/espacios internos), pudiendo colisionar con unique constraints o reescribir filas existentes (rename en `ensure*ForUpdate`). El código de campaña queda sólo trimmed a propósito. **Mitigación:** revisar `text/propername.go` (004) y probar con nombres reales del cliente.

## Técnicos
- **No compila aislado (alto):** dependencias `internal/actor`, `internal/shared/{lifecycle,authz,text}`, `platform/persistence/gorm/go` están **MISSING en develop**. **Mitigación:** mergear deps primero (ver `dependencies.md`); checklist de verificación en `extraction-plan.md` paso 0.
- **Domain con `ActorID` (alto):** DTOs/modelos referencian `ActorID` en `customerdom/managerdom/investordom`, structs que en develop NO lo tienen y NO están en mi flist. **Mitigación:** asegurar 007/011 mergeados (aportan esos campos).
- **`GetRawAdminCostTotal` doble-ownership (medio):** agregada aquí, consumida por 018. **Mitigación:** confirmar con 018 que no la duplica ni la espera con otra firma.

## Integración
- **Orden de merge (alto):** si 010 entra antes que 007/008/009/004/tenancy, rompe el build de develop. **Mitigación:** respetar orden tenancy→004→008→007→009→010.
- **011 (campaign-dto-projectid) (medio):** comparte DTO/domain de project; si 011 va antes choca. **Mitigación:** 010 antes de 011.

## Cross-repo
- **Mergear solo BE (medio):** seguro a nivel build (BE-first), pero deja el FE-010 desincronizado hasta que porte el path `/hard` y `actor_id`. **Riesgo:** ventana en la que la UI de delete no funciona.
- **Mergear solo FE (alto):** NO hacerlo. El FE-010 que llama `/hard` o manda `actor_id` falla contra un BE sin esta feature. **Mitigación:** BE-first estricto.

## Datos / migración
- **Columnas inexistentes en runtime (alto):** sin migraciones de `tenant_id` (todo el grafo), `actor_id` (customers), `actors`/`legacy_actor_map`, columnas `archive_*`, las queries fallan en runtime (`column ... does not exist`). **Mitigación:** las migraciones las aportan 001/003/007/009; verificar que estén aplicadas en cada entorno antes de exponer la feature.
- **Rename masivo (medio):** `ensure*ForUpdate` puede reescribir `name` de customers/managers/investors existentes al canonicalizar; un `23505` (unique) se traduce a 409. **Mitigación:** correr en staging con datos representativos.

## Archivos compartidos
- En mi flist NO hay archivos compartidos (no `go.mod`, no `wire/*`, no `cmd/*`, no `shared/handlers/*`). Riesgo de conflicto de merge por hunks compartidos: **bajo**. El riesgo se concentra en las DEPENDENCIAS, no en estos 9 archivos.

## Extracción parcial
- **Riesgo de traer solo algunos de los 9 (medio):** los tests y el repository están acoplados (helpers/firmas). Traer `repository.go` sin `usecases.go`/`handler.go` rompe el port. **Mitigación:** extraer los 9 enteros juntos (whole-file), nunca por hunks sueltos.
- **Señal de extracción incompleta:** `undefined: actorsync.* / lifecycle.* / authz.* / tenancy.* / text.CanonicalizeName`, o `(customerdom.Customer).ActorID undefined`.
