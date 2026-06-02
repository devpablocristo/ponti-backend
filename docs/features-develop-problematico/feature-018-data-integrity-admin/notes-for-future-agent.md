# feature-018 · data-integrity-admin · notes-for-future-agent (BE)

## Resumen corto

FULL-STACK. En BE es un refactor del módulo `internal/data-integrity`: pasa de 17 controles
encadenados (módulo-vs-módulo, `usecases.go` de 1068 líneas, timeout 8 min) a **5 controles
independientes** SSOT-dashboard-vs-RAW-tabla-base (343 líneas, paralelo, timeout 30 s). Mi flist
son 8 archivos, TODOS dentro de `internal/data-integrity/`. El módulo es portable whole-file,
pero **NO compila solo**: necesita dependencias que viven en archivos de otras features.

## Qué está en FE y qué está en BE

- **BE (este paquete)**: módulo `internal/data-integrity` (handler, dto, dominio, usecases, tests).
- **FE (otro repo, mismo feature-018)**: `pages/admin/data-integrity` + hook `useDatabase`. Consume `GET /data-integrity/costs-check?project_id=`.

## Archivos esenciales (los de mi flist — whole-file)

- `internal/data-integrity/usecases.go` (corazón: 5 controles + `buildCheck`, define 6 ports).
- `internal/data-integrity/handler.go` (ruta `costs-check`, timeout 30s).
- `internal/data-integrity/handler/dto/integrity_check.go` (+check_type/severity/recommendation).
- `internal/data-integrity/usecases/domain/types.go` (dominio sin json tags).
- 4 archivos `*_test.go`.

## Archivos PELIGROSOS / mezclados (NO están en mi flist, son partial-hunks)

- `internal/report/repository.go` → solo `GetRawNetIncome`.
- `internal/supply/repository.go` → solo `GetRawSupplyInvestment`.
- `internal/project/repository.go` → solo `GetRawAdminCostTotal` (archivo de +1642 líneas; muy mezclado, cuidado).
- `internal/lot/repository.go` → solo `GetRawLeaseExecuted`.
- `internal/work-order/repository.go` → `GetRawDirectCost` (ya en develop, sin tenant).
- `wire/data_integrity_providers.go` → solo el bloque DataIntegrity.
- `wire/wire_gen.go` → **regenerar, no copiar**.
- `internal/shared/authz/*` → lo trae tenancy (001/003), NO 018.

## Decisiones ya tomadas

- Extracción del módulo = **whole-file** (los 8 de la flist).
- Dependencias compartidas = **partial-hunks** (solo los `GetRaw*`), wire = regenerar.
- DECISIÓN global: **arreglar antes** (traer dependencias) — no es "extraer tal cual".
- Orden cross-repo: **BE-first**.
- tentative-prices / lot-metrics: **excluidos** (ya DONE, sin overlap con mi flist).
- `wire/wire.go` (ActorSet): **no traer** (es feature-007).

## Dudas abiertas (para humano)

1. ¿La tenantización de `GetRawDirectCost` (work-order) la dueña es 018 o tenancy? No bloquea compilación, sí coherencia de datos.
2. ¿La limpieza de json-tags de `usecases/domain/types.go` va con 018 o con 027? Recomendado: con 018.
3. ¿El FE necesita que el BE emita `WARNING/SKIPPED`? Hoy solo `OK/ERROR`.

## Comandos para mirar primero

```bash
cat /tmp/flists/be-018.txt
git -C <repo> diff 0972e565..777e5f6a -- internal/data-integrity/usecases.go | less
git -C <repo> show 777e5f6a:internal/data-integrity/usecases.go | head -170
git -C <repo> ls-tree -r --name-only develop -- internal/shared/authz        # ¿está la base de tenancy?
git -C <repo> grep -n "GetRawNetIncome\|GetRawSupplyInvestment\|GetRawAdminCostTotal\|GetRawLeaseExecuted" develop   # ¿existen los RAW en develop? (deberían dar 0)
git -C <repo> diff develop..777e5f6a -- wire/data_integrity_providers.go
```

## Errores a evitar

- Creer que `go test ./internal/data-integrity/...` verde = listo. Los tests usan mocks; el gate real es `go build ./...`.
- Copiar `wire_gen.go` a mano: desincroniza. Regenerar.
- Traer los repos compartidos enteros: arrastra cambios de 001/003/010/etc. Solo los hunks `GetRaw*`.
- Mergear FE-first: el FE espera campos que el BE viejo no envía.
- Usar `develop-problematico` (tip vacío). Siempre `develop-problematico~1` = `777e5f6a`.

## Camino más seguro

1. Confirmar tenancy (`internal/shared/authz`) en develop (de 001/003).
2. Branch `pr/feature-018-data-integrity-admin-be` desde develop.
3. `git checkout develop-problematico~1 -- <8 archivos del módulo>`.
4. `git restore -p --source=develop-problematico~1 -- <4 repos>` aceptando solo los `GetRaw*`.
5. `git restore -p ... -- wire/data_integrity_providers.go` (bloque DataIntegrity).
6. `go generate ./wire/...` ; `go build ./...` ; `go test ./internal/data-integrity/...`.
7. PR BE-first; luego coordinar FE 018.

## Qué PR del otro repo va antes/después

- **Antes (mismo repo BE)**: features 001/003 (tenancy + `internal/shared/authz`).
- **Después (repo FE)**: FE `feature-018` (`pages/admin/data-integrity` + `useDatabase`), que ya espera el DTO nuevo.
