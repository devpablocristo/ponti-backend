# risks.md — feature-020 · CI / GitHub workflows (BE)

> Riesgo global: **MEDIO-ALTO**. Los YAML son simples, pero tocan deploy a PROD, auth en DEV y reset de DB. La feature NOTE lo dice: "pueden romper deploy si se traen sin el resto".

## Riesgos funcionales

| # | riesgo | severidad | mitigación |
|---|---|---|---|
| F1 | `deploy-dev.yml` activa `AUTH_ENABLED=true` → el ambiente DEV exige auth y rompe a clientes/FE que hoy no la mandan. | **alta** | Gatear ese hunk (Sub-PR B parcial con `restore -p`): no activar AUTH hasta que feature-008 esté en develop y el FE DEV mande credenciales. Coordinar cross-repo. |
| F2 | Companion/Nexus quedan sin BaseURL/secret si las vars/secrets no existen en GCP → features que llaman al AI fallan en runtime. | media | Confirmar repo vars `COMPANION_BASE_URL_*`/`NEXUS_BASE_URL_*` y secrets `companion-internal-jwt-secret-*` antes de mergear Sub-PR B. |
| F3 | Hardening post-restore ahora corre con `set -euo pipefail` y **falla el job** si el SQL rompe (antes solo warning). | media | Validar `scripts/db/hardening_post_restore.sql` contra el esquema actual antes de correr el reset. |

## Riesgos técnicos

| # | riesgo | severidad | mitigación |
|---|---|---|---|
| T1 | `gcloud run deploy` falla si el secret `companion-internal-jwt-secret-{env}` no existe en Secret Manager. | **alta** | Crear secrets en GCP antes del PR; smoke test post-deploy. |
| T2 | `reset-dev-db-from-prod.yml` referencia `migrations_v4/000224_*.up.sql`; si el path/numeración no existe en develop, el step rompe. | **alta** (solo workflow manual) | Verificar `git ls-tree develop migrations_v4/000224_tenant_security_foundation.up.sql`; postergar Sub-PR C si falta. |
| T3 | Sintaxis YAML / strings largos de `ENV_VARS` (una sola línea enorme) — un error de pegado parte el deploy. | media | `actionlint`/`yamllint`; preferir `git checkout <sha> -- <file>` (copia exacta) sobre edición manual. |
| T4 | `ci-pr.yml` quita `CORE_GOVERNANCE_MODULE`; si OTRO workflow en develop aún lo usa, queda inconsistente. | baja | `git grep CORE_GOVERNANCE_MODULE develop -- .github/` debe quedar vacío. |

## Riesgos de integración

| # | riesgo | mitigación |
|---|---|---|
| I1 | El set `watched` de auditoría/smoke pasa a COMPANION/NEXUS pero el ambiente todavía tiene `AI_SERVICE_URL` taggeada → falso positivo/negativo de alineación. | Mergear renombre de vars y limpieza de env del ambiente en la misma ventana. |
| I2 | Coverage gate sube `coverage.out`; depende de que `go test ./...` compile. Si una feature a medio portear rompe el build, el job de tests falla (aunque el gate sea warning). | El job `build` y `tests` son separados; build roto se detecta en su propio job. |

## Riesgos cross-repo (FE)

| # | riesgo | mitigación |
|---|---|---|
| C1 | Mergear BE con renombre de vars mientras el FE/infra todavía espera `AI_SERVICE_*` → vars huérfanas / config desalineada entre repos. | Coordinar ventana con feature-020-FE; config BE-first. |
| C2 | `AUTH_ENABLED=true` en DEV (BE) sin que el FE DEV mande auth → DEV roto para QA. | Mergear el lado FE de auth (008) primero o simultáneo; gatear el hunk AUTH. |
| C3 | `audit-service-alignment.yml` posiblemente duplicado en FE → divergencia del set `watched`. | Verificar y alinear ambos repos. |

## Riesgos de datos / migración (reset-dev-db)

| # | riesgo | severidad | mitigación |
|---|---|---|---|
| D1 | El reset trunca `users`/`auth_*` y backfillea `tenant_id`. Si la migración 224 o el set de migraciones difiere de lo que asume el YAML, la DEV DB queda inconsistente o el job rompe a mitad. | **alta** (solo workflow manual, ambiente DEV) | Correr primero en una réplica/ventana segura; el target es DEV, no PROD. Validar numeración (`goto 224`) contra develop. |
| D2 | Dedupe de `TABLE DATA`/`SEQUENCE SET`: si pg_restore list tiene un formato distinto, el filtro Python deja pasar duplicados o descarta datos válidos. | media | Revisar `/tmp/prod_data.list` generado en un dry-run. |

## Riesgos de archivos compartidos / extracción parcial

| # | riesgo | mitigación |
|---|---|---|
| S1 | `deploy-dev.yml` mezcla 2 intenciones (COMPANION/NEXUS + AUTH). Traer el archivo entero arrastra AUTH aunque 008 no esté. | Usar `git restore -p` y aceptar solo los hunks listos. |
| S2 | Traer un `deploy-*.yml` a medias (algún hunk omitido) deja env inconsistente y deploy silenciosamente mal configurado. | Comparar hunk-a-hunk: `git diff develop..rama` vs `git diff 0972e565..777e5f6a` por archivo. |

## Riesgo de mergear SOLO este repo / SOLO el otro

- **Solo BE**: deploys rotos si faltan secrets/código; DEV roto por auth. Seguro únicamente para Sub-PR A (`ci-pr.yml`).
- **Solo FE**: si el FE rota a Companion/Nexus pero el BE sigue con `AI_SERVICE_*`, las llamadas FE→BE/AI quedan desalineadas. El job de auditoría (si compartido) marcaría drift.
- **Recomendación**: A se mergea solo sin problema; B y C requieren coordinación cross-repo + features previas + config GCP.
