# file-list.md — feature-020 · CI / GitHub workflows (BE)

Fuente: `cat /tmp/flists/be-020.txt`. Diff base→source: `git -C <core> diff 0972e565..777e5f6a -- <path>`.
Total de paths en el flist: **6** (todos `.github/workflows/*.yml`, status M).

Leyenda extracción: `whole-file` (traer el archivo entero del SOURCE) · `partial-hunks` (solo algunos hunks, archivo sirve a varias intenciones) · `manual-port` (rehacer a mano) · `do-not-extract-yet` (postergar).

## Propios (de la feature, infra CI/CD pura)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `.github/workflows/ci-pr.yml` | M | GH Actions (CI PR) | Pipeline de PR: lint / build / tests+coverage / security-scan | **partial-hunks** | Mezcla 2 intenciones: (a) remoción `CORE_GOVERNANCE_MODULE` [platform/deps] y (b) coverage gate [feature-025]. Ambos hunks son de bajo riesgo y warning-only; se pueden traer juntos. | bajo | alta |
| `.github/workflows/audit-service-alignment.yml` | M | GH Actions (auditoría) | Verifica que vars `*_URL` no apunten a URLs taggeadas | **partial-hunks** | Solo cambia el set `watched` (AI_SERVICE_URL → COMPANION_BASE_URL/NEXUS_BASE_URL) + quita 1 línea en blanco final. Acoplado al renombre de vars (feature-012). | bajo-medio | alta |

## Compartidos (partial-hunks — diff sirve a varias intenciones / config de runtime de otras features)

| path | status | tipo | rol en la feature | extracción | motivo | riesgo | confianza |
|---|---|---|---|---|---|---|---|
| `.github/workflows/deploy-dev.yml` | M | GH Actions (deploy Cloud Run DEV) | Inyecta env/secrets al contenedor DEV + smoke | **partial-hunks / manual-port** | Mezcla: renombre AI_SERVICE→COMPANION/NEXUS (feature-012), `AUTH_ENABLED=true`+`APP_ENV=dev` (feature-008), set `watched` del smoke (feature-012). Cada bloque depende de código/secrets distintos. | **alto** (cambia AUTH_ENABLED y secrets) | alta sobre el diff, media sobre seguridad de mergear |
| `.github/workflows/deploy-staging.yml` | M | GH Actions (deploy Cloud Run STG) | Inyecta env/secrets STG + smoke | **partial-hunks** | Renombre COMPANION/NEXUS + secret companion + set `watched`. NO toca AUTH (sigue sin AUTH_ENABLED explícito). | medio-alto | alta |
| `.github/workflows/deploy-prod.yml` | M | GH Actions (deploy Cloud Run PROD) | Inyecta env/secrets PROD + smoke | **partial-hunks** | Renombre COMPANION/NEXUS + secret `companion-internal-jwt-secret-prod` + set `watched`. | **alto** (PROD) | alta |

## Requeridos por dependencia (NO en este flist; deben preexistir en develop al mergear)

| artefacto | feature que lo aporta | por qué lo necesita este paquete | extracción |
|---|---|---|---|
| `cmd/config/companion.go` (struct `Companion`/`Nexus`) | feature-012 | El binario debe leer `COMPANION_*`/`NEXUS_*` que el deploy inyecta | do-not-extract-here (referencia) |
| `internal/axis/client.go` | feature-012 | Cliente que consume Companion | do-not-extract-here |
| `cmd/config/auth.go` + middlewares gin | feature-008 | `AUTH_ENABLED=true`/`APP_ENV` en DEV | do-not-extract-here |
| `migrations_v4/000224_tenant_security_foundation.up.sql` + dir `migrations_v4/` | feature-003 / 019 | `reset-dev-db-from-prod.yml` hace `goto 224` y aplica ese .sql | do-not-extract-here |
| tests reales (`*_test.go`) | feature-025 | Alimentan el coverage gate (warning-only) | do-not-extract-here |

## Dudosos

| path | duda | cómo resolver |
|---|---|---|
| `.github/workflows/reset-dev-db-from-prod.yml` | El diff referencia `migrations_v4/000224_*.up.sql` y tablas auth (`users`,`auth_*`,`fx_rates`). ¿Existen en develop hoy? | `git -C <core> ls-tree develop migrations_v4/000224_tenant_security_foundation.up.sql` y revisar features 003/019. Si no existe, **do-not-extract-yet**. |
| `audit-service-alignment.yml` | ¿El job es compartido con el repo FE o duplicado? | Comparar con feature-020-FE; si comparten infra de auditoría, coordinar el renombre simultáneo. |

## NO traer todavía (postergar)

| path | extracción | motivo |
|---|---|---|
| `.github/workflows/reset-dev-db-from-prod.yml` | **do-not-extract-yet** | Acoplado a migración 224 + dir `migrations_v4/` (features 003/019). Es workflow manual (`workflow_dispatch`), no bloquea CI/deploy. Traer recién cuando 003/019 estén en develop. |
| Bloque `AUTH_ENABLED=true` + `APP_ENV=dev` de `deploy-dev.yml` | **partial-hunks, gated** | Solo activar cuando feature-008 (auth) esté porteada y el FE de DEV mande credenciales; si no, rompe el ambiente DEV. |
| Bloque COMPANION/NEXUS de los `deploy-*.yml` | **partial-hunks, gated** | Solo cuando feature-012 esté porteada y existan vars `COMPANION_BASE_URL_*`/`NEXUS_BASE_URL_*` y secret `companion-internal-jwt-secret-*` en GCP. |

## Resumen de archivos COMPARTIDOS (partial-hunks)

Todos los `deploy-*.yml` y `ci-pr.yml` y `audit-service-alignment.yml` son técnicamente compartidos: cada uno mezcla cambios de infra CI con configuración de runtime que pertenece a features 008/012/003/019/025. Por eso la recomendación es desmenuzar por hunks y gatear por dependencia, no traer los archivos enteros de una.
