# risks.md — feature-005 · be-config-modularization

## Funcionales

- **Cambio de default `Auth.AutoProvision` true→false** (`AUTH_AUTO_PROVISION_MEMBERSHIP`).
  - Riesgo: si features 001/008 ya leyeran este flag en `develop`, el auto-provisioning de membresías deja de ocurrir por default.
  - Mitigación: como en esta feature **no hay consumer**, no hay efecto al mergear sola. Comunicar a 001/008. Deploys que dependan de auto-provision deben setear `AUTH_AUTO_PROVISION_MEMBERSHIP=true` explícito.
- **`Auth.RequireTenantHeader` default `true`** (`AUTH_REQUIRE_TENANT_HEADER`).
  - Riesgo: cuando aterricen los middlewares (001/008), las requests sin `X-Tenant-Id` serán rechazadas salvo tenant implícito.
  - Mitigación: documentado; el comportamiento real lo activa el middleware, no este paquete.

## Técnicos

- **Campos/structs huérfanos:** `Companion`, `Nexus`, `Reporting`, `Security`, `CORSOriginList`, etc. quedan sin uso hasta 012/023/021/001.
  - Riesgo: linters muy estrictos (no Go por defecto) podrían marcar unused. `go build`/`go vet` NO fallan.
  - Mitigación: aceptar como deuda; documentado en implementation-status.
- **Referencia residual a `config.AI`:** si en `develop` existiera código consumiendo `config.AI`, borrar `ai.go` rompe el build.
  - Mitigación: antes de borrar, `git -C ... grep -nE "config\.AI" -- cmd internal wire` (en SOURCE no hay hits en código).

## Integración

- El paquete habilita pero no integra: los clientes Companion/Nexus reales viven en `axis/` y se cablean en 012/023. El `.env.example` referencia secretos (`COMPANION_INTERNAL_JWT_SECRET`, `NEXUS_INTERNAL_JWT_SECRET`, `REVIEW_API_KEY`) que deben coincidir con `axis/docker-compose.yml`/`axis/.env`.
  - Mitigación: son valores de ejemplo dev; en prod inyectar desde GCP Secret Manager. No es responsabilidad de esta feature.

## Cross-repo

- **Ninguno.** Solo-BE. No hay riesgo de desincronización con el FE.

## Datos / migración

- **Ninguno.** No hay migraciones ni cambios de datos.

## Archivos compartidos

- **`.env.example`** — el mayor riesgo de la feature. Hunks de 005/012/019/021 conviven.
  - Riesgo: arrastrar el bloque `DB_*_PROD` (feature-019) o pisar hunks de deploy (021).
  - Mitigación: `git restore -p` aceptando solo COMPANION_*/NEXUS_*/REVIEW_* y el header; rechazar `# PROD data source for local DB reset`.
- **`cmd/config/loadconfig.go`** — agregador. En el rango de 005 solo cambió el bloque de campos; whole-file es seguro **si** `develop` no tiene otros hunks pendientes. Verificar `git diff develop..SOURCE -- cmd/config/loadconfig.go`.

## Extracción parcial

- Riesgo de traer consumers por error (wire/, cmd/api/http_server.go, middlewares) → build roto por símbolos de features futuras no presentes.
  - Mitigación: ceñirse al flist; lista de "do-not-extract-yet" en file-list.md.

## Riesgo de mergear solo este repo / solo el otro

- **Solo BE (este repo):** seguro. Es leaf; no rompe nada y no requiere FE. Recomendado mergear primero.
- **Solo el otro repo (FE):** N/A — no hay cambios FE asociados.
- **Riesgo de NO mergear esta antes de 012/023:** 012 y 023 no compilarán (faltan `config.Companion`/`config.Nexus`). Orden estricto 005 primero.
