# 0003 Registry — Aggregated Read Endpoint In Isolated Module

Specification type: baseline current-state observed decision specification.

Status: Observed Implementation Decision.

No formal decision-record approval evidence was verified.

## Observed Implementation Decision: Aggregated Read Lives In An Isolated `internal/registry` Module

A new module `internal/registry` provides a read-only aggregated search over heterogeneous entity kinds, kept separate from the per-type modules. It exposes a `GET /registry` search and a `PUT /registry/actors/:actor_id/aliases` mutation, wired through the wire set and registered in the HTTP server alongside the other handlers.

Evidence:
- `internal/registry/handler.go` (`Routes`: `GET ""`, `PUT /actors/:actor_id/aliases`)
- `wire/registry_providers.go`, `wire/wire.go`
- `cmd/api/http_server.go` (`deps.RegistryHandler.Routes()`)

## Observed Implementation Decision: Read Is A Typed, Paginated UNION

`GET /registry?q=&type=&status=&page=&per_page=` returns typed rows `{entity_type, id, name, tax, roles, archived}` produced by a `UNION ALL` of typed sources: actors once (with `tax` from the active `TAX_ID` key and `roles` aggregated from `actor_roles`) plus the 4 catalogs (`crops`, `types`, `lease_types`, `campaigns`). `type` selects the sources: `all` emits actors + the 4 catalogs; a role value (`customer`, `provider`, `investor`, `manager`, `contractor`, `biller`, `lessee`) restricts to actors having that role via `EXISTS (... actor_roles ...)`; a catalog value emits only that catalog. `q` matches `display_name` plus active `actor_keys` (`LEGAL_NAME`/`PERSON_NAME`/`ALIAS` ILIKE, `TAX_ID` by digit prefix) plus catalog `name`. Status maps to `deleted_at` (active/archived/all). Pagination is server-side; results are tenant-scoped flag-gated.

Evidence:
- `internal/registry/repository.go` (`SearchRegistry`: `actorRoles`, `catalogTables`, `UNION ALL`, `statusSQL`, `tenantSQL`, count + paged query)
- `internal/registry/handler.go` (`Search`, default `type=all`, default per-page 100)

## Observed Implementation Decision: Mutations Reuse The Per-Type Endpoints

The registry does not own per-type mutations. The only mutation it adds is alias rotation (`PUT /registry/actors/:actor_id/aliases`, replacing the active `ALIAS` key set, with 409 on collision). All other create/update/archive/restore operations go through the existing per-type endpoints (the `/actors` group and the catalog modules), so the registry stays a read aggregator plus an alias editor. The typed `entity_type`/role surface is the base for later abstracting role ↔ type.

Evidence:
- `internal/registry/handler.go` (`SetActorAliases` is the only registry mutation)
- `internal/registry/repository.go` (alias rotation: insert `actor_keys ... key_type 'ALIAS' active true`)
- `internal/actors/handler.go` (per-type actor mutations the registry defers to)

## UNKNOWN

- Whether this decision was formally approved.
- Decision owner and date.
- Whether the registry will absorb additional entity kinds (fields/projects/supplies/labors are noted as future types).
- The final shape of the role ↔ type abstraction the typed UNION is intended to seed.
