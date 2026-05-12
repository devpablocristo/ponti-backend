# Multi-Tenant 100 Evidence

Fecha: 2026-05-12

Este documento define que se puede afirmar sin mentir sobre el alcance
multi-tenant enterprise de Ponti.

## Claim permitido hoy

La base multi-tenant esta implementada y validada localmente: hardening de auth,
helpers de tenant, BFF/UI tenant foundation, actors/AI foundation, migraciones
base, scoping transicional en los principales CRUD operativos y scoping tenant
en reportes analiticos clave.

## Claim no permitido todavia

No decir "100% enterprise multi-tenant completo" hasta cerrar todos los gates:

- `TENANT_STRICT_MODE=true`
- `BFF_REQUIRE_TENANT=true`
- `AI_TENANT_SCOPE=true`
- auditoria SQL contra dump dev sin hallazgos
- golden master contra dump dev con diff numerico `0`
- tests IDOR por modulo tenant-owned
- cobertura IDOR completa por modulo tenant-owned

## Evidencia local verde

Backend:

- `go test ./...`
- `go build ./cmd/api ./cmd/migrate`

BFF:

- `npm run lint`
- `npm test`
- `npm run build`

UI:

- `yarn typecheck`
- `yarn build`
- `yarn lint` con 0 errores y warnings existentes

## Cambios cubiertos en esta tanda

- `tenant_id` agregado a modelos operativos que ya estan en migraciones 224/225:
  - campaigns
  - fields / field investors
  - lots / lot dates
  - labors
  - supplies
  - supply movements
  - stocks
  - workorders / workorder items / workorder investor splits
  - invoices
  - work order drafts
- Scoping transicional con `authz.MaybeTenantScope` en CRUD/list/archive/restore/hard-delete de esos modulos.
- `sharedfilters.ResolveProjectIDs` ahora respeta tenant cuando existe contexto y permite "todos los proyectos del tenant" sin caer en "todos los proyectos de la base".
- `dashboard` resuelve siempre los proyectos via `sharedfilters.ResolveProjectIDs`; un `project_id` directo ya no saltea validacion tenant.
- `report` y `data-integrity` propagan `context.Context` hasta las consultas analiticas para heredar tenant.
- `actors` ya no resuelve `default` en requests normales: list/get/update/archive/restore/hard-delete/merge/deduplicacion usan el tenant del contexto.
- `project` escribe `tenant_id` en proyectos, pivots y relaciones nuevas; sus list/get/update/archive/restore/hard-delete quedan tenant-scoped en modo transicional.
- `scripts/db/tenant_isolation_audit.sql` cubre mas relaciones cross-tenant y constraints globales peligrosos.
- `scripts/db/multi_tenant_golden_master.sql` deja un snapshot numerico reproducible para diff antes/despues.
- Test agregado: `internal/shared/filters` verifica que `ResolveProjectIDs` falle cerrado frente a `project_id`/`field_id` de otro tenant.
- BFF/UI validations siguen verdes.

## Bloqueantes para declarar 100%

- Ejecutar `scripts/db/tenant_isolation_audit.sql` sobre dump fresco de dev.
- Ejecutar golden master completo sobre dump fresco de dev.
- Agregar/ejecutar tests IDOR por cada modulo tenant-owned, no solo helpers compartidos.
- Verificar con DB real que `scripts/db/multi_tenant_golden_master.sql` cubre todas las columnas presentes en dev.
- Activar flags strict solo despues de esos gates.

## Frase final permitida despues de cerrar gates

Se completo 100% del alcance enterprise definido: auth, tenant isolation, RBAC,
BFF, UI, AI, auditoria DB y golden master, sin activar cambios que alteren
calculos.
