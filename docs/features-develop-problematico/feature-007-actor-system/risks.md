# risks.md — feature-007 actor-system (BE)

## Funcionales

| riesgo | impacto | mitigación |
|---|---|---|
| Sync legacy escribe en tablas de otras features (`customers`, `projects`, `project_responsibles`, `project_investor_allocations`, `project_admin_cost_allocations`) | datos inconsistentes si esas tablas no existen / difieren de su versión final | confirmar que 010 projects está en develop con el esquema esperado; correr migración 223 en una copia antes de prod |
| Unicidad no aplicada si el índice 234 no se crea | duplicados activos por nombre | validar que `git`/migrate aplicó 234; test `repository_tenant_test.go` cubre el conflicto |
| Normalización divergente Go vs SQL | un nombre considerado único por Go pero duplicado por DB (o viceversa) | `normalizeName` (repo.go:1094) y `normalize_actor_name` (migr 223) deben ser idénticos; si se toca uno, tocar el otro (lo dice el SPEC) |

## Técnicos

| riesgo | impacto | mitigación |
|---|---|---|
| Cableado fuera del flist no portado | feature inalcanzable (404 en `/api/v1/actors`), DI sin handler | portar hunks de `wire/wire.go`, `wire/wire_gen.go`, `cmd/api/http_server.go`; verificar 4 hits de `ActorHandler` |
| `wire_gen.go` portado a mano e inconsistente | `wire`/build roto | regenerar con `go run github.com/google/wire/cmd/wire ./wire/...` en vez de copiar hunks |
| Dependencia shared/platform ausente (001-004) | no compila | mergear deps antes; `go build ./...` |

## Integración / cross-repo

| riesgo | impacto | mitigación |
|---|---|---|
| Mergear FE feature-007 antes del BE | FE llama endpoints inexistentes (404) | respetar BE-first; desplegar y verificar `/api/v1/actors` antes del FE |
| Cambio de shape del DTO no coordinado | FE rompe (`archived_at`, `data/page_info`, `duplicate_candidates`, `merge_impact`) | congelar contrato; el DTO ya expone `archived_at` por compat aunque la DB use `deleted_at` |

## Datos / migración

| riesgo | impacto | mitigación |
|---|---|---|
| **234 hace merge destructivo de duplicados activos** (conserva el de menor id, marca el resto `merged_into_actor_id` + `deleted_at`, loguea en `actor_merge_log`) | pérdida/colapso de identidades si había homónimos legítimos | revisar `actor_merge_log` post-migración; correr en staging primero; confirmar con humano si aplica a prod |
| 231 dropea columna `archived_at` de actors/roles/aliases | irreversible parcialmente si hubo datos solo en `archived_at` | 231 backfillea `archived_at`→`deleted_at` antes de dropear; verificar el `.down.sql` |
| 226 `fk_customers_actor ... NOT VALID` | constraint no validada; filas viejas podrían violar | planear `VALIDATE CONSTRAINT` posterior cuando los datos estén limpios |
| backfill 1:1 desde legacy crea actores por cada customer/investor/manager/provider | volumen / posibles colisiones de nombre antes de 234 | el flujo 223→...→234 está diseñado para consolidar al final; respetar el orden |

## Archivos compartidos

| archivo | riesgo | mitigación |
|---|---|---|
| `wire/wire_gen.go` | generado, fácil de romper al portar parcial | regenerar |
| `wire/wire.go` | mezcla deps de todas las features | `git restore -p` solo el hunk del actor |
| `cmd/api/http_server.go` | una línea entre 28 handlers | `git restore -p` solo `deps.ActorHandler.Routes()` |
| migración 223 | crea tablas `project_*` que "pertenecen" a 010 | coordinar quién declara esas tablas para no duplicar CREATE entre PRs |

## Extracción parcial

- Si se trae solo el flist (sin cableado): compila, tests del paquete pasan, pero las
  rutas no existen → falso "verde". **Señal**: `curl /api/v1/actors` → 404.
- Si se omiten down migrations: rollback imposible. Traer up + down siempre.
- Si se omite `repository_tenant_test.go`: se pierde la cobertura de tenancy + conflicto.

## Riesgo de mergear solo un repo

- **Solo BE**: seguro. El backend queda funcional y testeable; el FE simplemente aún no
  lo consume. Recomendado como primer paso (BE-first).
- **Solo FE (sin BE)**: roto. El FE feature-007 llamaría a `/api/v1/actors` inexistente.
  No mergear FE hasta que el BE esté desplegado.
