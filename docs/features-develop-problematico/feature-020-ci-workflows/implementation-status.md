# implementation-status.md — feature-020 · CI / GitHub workflows (BE)

## Estado global

- **Estado**: **completa en el SOURCE REF** (`777e5f6a`) a nivel de los 6 YAML. Los archivos están en su forma final/coherente.
- **% completitud (de los YAML en sí)**: ~100%.
- **% completitud para mergear con seguridad a develop**: ~40% — porque la mayoría de los cambios dependen de features aún no confirmadas en develop (012/008/003/019) y de config externa (vars/secrets GCP) que no está en git.

## Estado en este repo (BE)

| archivo | cambio | estado intrínseco | mergeable a develop hoy |
|---|---|---|---|
| `ci-pr.yml` | remoción governance + coverage gate | completo | **SÍ** (bajo riesgo; coverage warning-only) |
| `audit-service-alignment.yml` | set `watched` COMPANION/NEXUS | completo | condicionado a renombre de vars (junto con deploys) |
| `deploy-staging.yml` | COMPANION/NEXUS + secret | completo | solo si 012 + vars/secrets STG listos |
| `deploy-prod.yml` | COMPANION/NEXUS + secret | completo | solo si 012 + vars/secrets PROD listos |
| `deploy-dev.yml` | COMPANION/NEXUS + secret + `AUTH_ENABLED=true` + `APP_ENV=dev` | completo | solo si 012 **y** 008 + vars/secrets DEV listos + FE DEV manda auth |
| `reset-dev-db-from-prod.yml` | goto 224 + backfill tenant + hardening psql | completo | solo si 003/019 (migración 224, migrations_v4, hardening sql) en develop |

## Estado en el otro repo (FE)

- **Desconocido desde este paquete**. Hay un feature-020-FE espejo (`.github/workflows` del FE). Pendiente: confirmar si comparte el renombre AI_SERVICE→COMPANION/NEXUS y/o el job de auditoría. Coordinar con el agente/paquete del FE.

## Tests

- No hay tests propios de la feature (es infra). El `ci-pr.yml` ejecuta `go test ./... -coverprofile=coverage.out` — depende de tests aportados por otras features (025). El gate es warning-only, no falla aunque la cobertura sea 0%.

## Pendientes

### BLOQUEANTE para mergear (cada sub-PR)
- **Sub-PR B**: confirmar `cmd/config/companion.go` en develop (012) + existencia de repo vars `COMPANION_BASE_URL_*`/`NEXUS_BASE_URL_*` y secrets `companion-internal-jwt-secret-*` en GCP.
- **Sub-PR B (DEV)**: confirmar feature-008 (auth) en develop antes de activar `AUTH_ENABLED=true`; confirmar que el FE DEV manda credenciales.
- **Sub-PR C**: confirmar `migrations_v4/000224_tenant_security_foundation.up.sql` y dir `migrations_v4/` en develop (003/019).

### Mejora futura
- Convertir el coverage gate de warning a enforcement (subir threshold) cuando feature-025 acumule tests.

### Deuda aceptable
- `audit-service-alignment.yml` solo quita una línea en blanco final además del set `watched` — cosmético.
- `REQUIRE_AUTH_SMOKE: "0"` se mantiene en todos los deploys (smoke no exige auth aún).

### Duda humana
- ¿La numeración de migraciones en develop coincide con `goto 224`? (memoria: el set viejo tenía otra numeración → el reset script ya había necesitado fixes). **Revisar antes de Sub-PR C.**
- ¿`audit-service-alignment.yml` es el mismo job en FE? Coordinar.

## Bugs / observaciones

- Ningún bug detectado en los YAML. El cambio de hardening de `gcloud sql import` → `psql` por el proxy es una corrección intencional (el import user no puede otorgar privilegios sobre `fx_rates`), no un bug.
- El `set +e`/`code=$?` permisivo del hardening viejo se reemplaza por `set -euo pipefail` → ahora el hardening **falla el job** si rompe (antes solo warning). Esto es endurecimiento intencional pero cambia el comportamiento: tenerlo presente al validar el reset.
