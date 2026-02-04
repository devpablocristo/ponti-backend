# Estrategia de Deploy y Ramas

## Objetivo

Tener un flujo que:

- sea rápido para desarrollar,
- sea seguro para llegar a producción,
- evite "DEV pantano",
- y garantice que lo que llega a PROD es exactamente lo que se probó en STAGING (mismo artefacto).

---

## Conceptos clave

### Environments (entornos)

| Entorno | Descripción |
|---------|-------------|
| Preview | Efímero, opcional, por PR |
| DEV | Compartido, integración continua |
| STAGING | Pre-prod, release candidates |
| PROD | Producción |

### Bases de datos

| DB | Entorno |
|----|---------|
| `db_pr_<id>` | Preview (efímera) |
| `db_dev` | DEV |
| `db_staging` | STAGING |
| `db_prod` | PROD |

### Regla de oro de promoción

> Se construye el artefacto **una sola vez** para STAGING y se promueve **ese mismo artefacto** a PROD. Nada de "rebuild en main" para prod.

---

## Ramas

### Ramas permanentes

| Rama | Propósito |
|------|-----------|
| `main` | Producción (PROD) |
| `develop` | Integración (DEV) |

### Ramas temporales

| Patrón | Uso |
|--------|-----|
| `feature/*` | Desarrollo de funcionalidades |
| `hotfix/*` | Arreglos urgentes desde main |

---

## Flujo recomendado (end-to-end)

### 1) Desarrollo en `feature/*`

#### 1.1 Crear rama

Desde `develop`:
```
feature/<ticket>-<descripcion-corta>
```

#### 1.2 Trabajo y commits

- Commits chicos y con intención clara.
- PRs chicos y frecuentes.

---

### 2) Preview Dev (opcional y on-demand)

#### ¿Cuándo se usa Preview?

Se usa solo si hace falta, por ejemplo:

- el PR toca migraciones/DB,
- necesitás un entorno aislado para probar,
- querés compartir un link para demo/QA puntual.

#### ¿Cómo se activa?

| Método | Workflow |
|--------|----------|
| Label en PR: `preview` | `deploy-preview.yml` |
| Workflow manual | "Deploy Preview" |

#### Qué crea el Preview

- Deploy del servicio con configuración de Preview
- DB efímera: `db_pr_<id>`
- URL efímera en Cloud Run

#### TTL y limpieza

Se destruye automáticamente al:

- cerrar el PR, o
- vencer **TTL de 48 horas**

**Workflow:** `cleanup-preview.yml`
- Ejecuta cada 6 horas (cron)
- Elimina servicios y DBs con más de 48h de antigüedad
- También limpia al cerrar PR

> **Regla:** si no hay TTL, no existe preview (te explota en costos/ruido).

---

### 3) Merge a develop → Deploy automático a DEV

#### 3.1 PR `feature/*` → `develop`

**Checks requeridos (CI):** `ci-pr.yml`

| Check | Obligatorio | Descripción |
|-------|-------------|-------------|
| `lint` | Sí | golangci-lint |
| `build` | Sí | `go build ./...` |
| `test` | Sí | `go test ./...` |
| `security-scan` | No (opcional) | govulncheck |

Además requiere **1 aprobación** (o lo que definas).

#### 3.2 Deploy automático a DEV

**Workflow:** `deploy-dev.yml`

Al mergear a `develop`:
- se deploya automáticamente a DEV
- usa `db_dev`

#### Migraciones en DEV

En DEV se permiten migraciones automáticas siempre que:
- el proceso sea idempotente,
- y no rompa deploys (ver sección "migraciones seguras").

---

### 4) Release Candidate: PR `develop` → `main` → Deploy automático a STAGING

#### 4.1 Crear PR `develop` → `main`

Este PR representa un **"release candidate" (RC)**.

**Checks requeridos (CI):** `ci-release.yml`

| Check | Obligatorio | Descripción |
|-------|-------------|-------------|
| `lint` | Sí | golangci-lint |
| `build` | Sí | `go build ./...` |
| `test` | Sí | `go test ./...` |
| `security-scan` | Sí | govulncheck |
| `gosec` | Sí | Análisis de seguridad estático |
| `dependency-review` | Sí | GitHub dependency review |

#### 4.2 Build y deploy a STAGING (build único)

**Workflow:** `deploy-staging.yml`

Cuando el PR se mergea a `main`:

- se construye una imagen con **tag inmutable**:
  ```
  app:<git_sha>
  ```
- se deploya a STAGING
- usa `db_staging`

#### 4.3 Smoke tests en STAGING

Al finalizar el deploy se ejecutan automáticamente:

| Test | Descripción |
|------|-------------|
| Healthcheck | `GET /ping` con retry |
| API check | `GET /api/v1/customers` (status < 500) |

**Si falla:**
- no se promueve
- se corrige en `develop` y se repite RC

---

### 5) Promote: STAGING → PROD usando el MISMO artefacto

#### 5.1 ¿Qué es "promote"?

Promote es una acción controlada que:

- **NO** recompila
- **NO** reconstruye
- Solo despliega a PROD exactamente la misma imagen/tag que se probó en STAGING

#### 5.2 Cómo se dispara

**Workflow:** `promote-prod.yml` (workflow_dispatch)

Input requerido:
- `commit_sha`: SHA del commit probado en STAGING

**Environment protection:** `prod` (requiere aprobación)

#### 5.3 Deploy a PROD

- Deploy a PROD con la imagen ya validada
- DB: `db_prod`
- Ejecuta smoke tests **"prod-safe"** (read-only)

#### Smoke tests PROD

| Test | Descripción |
|------|-------------|
| Healthcheck | `GET /ping` con retry |
| API read-only | `GET /api/v1/customers` con auth (status < 500) |

---

## Recuperación de DEV (DEV es descartable)

### Principio

> DEV no se arregla. DEV se destruye y se recrea.

Nada de "parchear datos a mano".

### Botón rojo: "Reset DEV"

**Workflow:** `reset-dev.yml` (workflow_dispatch)

#### Paso 1: Drop total de DEV DB
- Borrar `db_dev` completa
- Dejar infra intacta

#### Paso 2: Restore desde Golden Snapshot
Se restaura un snapshot estable, versionado:
```
gs://<bucket>/golden/latest.sql.gz
```

#### Paso 3: Post-restore hardening (opcional)
Script que garantiza que DEV no causa daños:
- desactivar webhooks
- limpiar outbox / colas pendientes
- invalidar tokens
- apagar emails reales
- reset flags peligrosos
- setear endpoints externos a "dummy" o sandbox

Configurado via: `HARDENING_SQL_URI`

#### Paso 4: Smoke tests automáticos
- healthcheck (si `SMOKE_TEST_URL` está configurado)

**Si falla:** DEV queda "no usable" y se vuelve a ejecutar reset.

---

## Golden Snapshot

### Qué es

Un snapshot "known good" de STAGING para restaurar DEV rápido.

### Workflow

`refresh-golden-snapshot.yml`

| Trigger | Descripción |
|---------|-------------|
| Manual | workflow_dispatch |
| Semanal | Lunes 4am UTC |

### Cómo funciona

1. Exporta `db_staging` a GCS
2. Guarda con timestamp: `golden_<db>_<ts>.sql.gz`
3. Copia a `latest.sql.gz` para fácil referencia

### Cómo se mantiene

- Se actualiza luego de un release estable, o semanalmente
- Se versiona y se guarda con checksum
- Se prueba restaurándolo (restores reales, no "teórico")

---

## Migraciones de DB (reglas mínimas)

### Regla de compatibilidad

Para evitar "romper deploy" y poder promover con seguridad:

#### Patrón Expand/Contract (recomendado)

1. **Expand:** agregar columnas/tablas nuevas sin romper lo viejo
2. **Deploy** app compatible con ambas
3. **Backfill** datos
4. **Contract:** remover lo viejo cuando ya no se usa

### Reglas operativas

- No correr migraciones destructivas "en caliente" en horarios críticos
- Todo migration debe ser:
  - repetible
  - idempotente
  - con locks/transactions cuando aplique
- En STAGING y PROD: migraciones con control (por pipeline) y logs claros

### Verificación de migraciones

**Workflow:** `db-verify.yml`

- Se ejecuta en PRs a `develop`
- Levanta PostgreSQL local
- Ejecuta `make db-verify`

---

## Reglas de CI/CD

### Checks en PR (`feature` → `develop`)

**Workflow:** `ci-pr.yml`

| Check | Bloquea PR |
|-------|------------|
| lint | Sí |
| build | Sí |
| test | Sí |
| security-scan | No |

### Checks en PR (`develop` → `main`)

**Workflow:** `ci-release.yml`

| Check | Bloquea PR |
|-------|------------|
| lint | Sí |
| build | Sí |
| test | Sí |
| govulncheck | Sí |
| gosec | Sí |
| dependency-review | Sí |

### Environments protections

| Environment | Aprobación |
|-------------|------------|
| dev | No requiere |
| stg | Opcional |
| prod | **Obligatoria** |

---

## Convenciones de tags e imágenes

### Tag inmutable por commit

```
app:<git_sha>
```

Recomendado para trazabilidad.

### Tag semver para releases (opcional)

```
vX.Y.Z
```

Apuntando al commit que ya fue a staging.

---

## Rollback (plan claro)

### Si algo sale mal en PROD

1. **Elegir** el tag anterior conocido (último prod estable)
2. **Deploy** con promote inverso (sin rebuild)
3. **Correr** smoke tests

> **Importante:** rollback de app ≠ rollback de DB. Si hay migraciones incompatibles, necesitás estrategia expand/contract.

---

## "Definition of Done" para features con DB

Si un PR toca DB:

- [ ] Migración aplicada en preview (si usaste preview)
- [ ] Smoke tests pasan en DEV
- [ ] Plan de migración segura (expand/contract) si es destructivo
- [ ] Scripts de hardening no se rompen

---

## Resumen operativo (checklist)

### Feature

- [ ] Rama `feature/*`
- [ ] Preview opcional si label `preview` o toca DB
- [ ] PR a `develop`
- [ ] CI pasa (lint, build, test)

### DEV

- [ ] Deploy auto desde `develop`
- [ ] Usa `db_dev`
- [ ] Si se rompe: Reset DEV (drop + restore golden + hardening + smoke)

### Release

- [ ] PR `develop` → `main`
- [ ] CI release pasa (lint, build, test, security)
- [ ] Build único y deploy a staging
- [ ] Smoke tests OK

### Prod

- [ ] Promote a prod con mismo artefacto
- [ ] Smoke tests prod-safe OK
- [ ] Rollback por artefacto si hace falta

---

## Anexo: Workflows implementados

| Workflow | Archivo | Trigger |
|----------|---------|---------|
| CI PR a develop | `ci-pr.yml` | PR a `develop` |
| CI Release | `ci-release.yml` | PR a `main` |
| DB Verify | `db-verify.yml` | PR a `develop` + manual |
| Deploy DEV | `deploy-dev.yml` | Push a `develop` |
| Deploy STAGING | `deploy-staging.yml` | Push a `main` |
| Deploy Preview | `deploy-preview.yml` | Label `preview` + manual |
| Cleanup Preview | `cleanup-preview.yml` | PR closed + cada 6h (TTL 48h) |
| Promote to PROD | `promote-prod.yml` | Manual (requiere SHA) |
| Reset DEV | `reset-dev.yml` | Manual |
| Refresh Golden Snapshot | `refresh-golden-snapshot.yml` | Manual + semanal |

---

## Variables de entorno requeridas

### Por environment en GitHub

| Variable | dev | stg | prod |
|----------|-----|-----|------|
| `GCP_PROJECT_ID_*` | ✓ | ✓ | ✓ |
| `GCP_REGION` | ✓ | ✓ | ✓ |
| `ARTIFACT_REGISTRY` | ✓ | ✓ | ✓ |
| `IMAGE_NAME` | ✓ | ✓ | ✓ |
| `SERVICE_NAME_*` | ✓ | ✓ | ✓ |
| `CLOUDSQL_INSTANCE_*` | ✓ | ✓ | ✓ |
| `DB_NAME_*` | ✓ | ✓ | ✓ |
| `WIF_PROVIDER_*` | ✓ | ✓ | ✓ |
| `WIF_SERVICE_ACCOUNT_*` | ✓ | ✓ | ✓ |

### Secrets

| Secret | dev | stg | prod |
|--------|-----|-----|------|
| `DB_PASSWORD_*` | ✓ | ✓ | ✓ |
| `X_API_KEY_*` | ✓ | ✓ | ✓ |

### Variables específicas

| Variable | Workflow | Descripción |
|----------|----------|-------------|
| `PREVIEW_SERVICE_PREFIX` | deploy-preview, cleanup | Prefijo servicios preview |
| `PREVIEW_BUCKET` | cleanup | Bucket para seeds preview |
| `GOLDEN_SNAPSHOT_URI` | reset-dev | URI del golden snapshot |
| `GOLDEN_SNAPSHOT_BUCKET` | refresh-golden | Bucket para golden |
| `HARDENING_SQL_URI` | reset-dev | SQL de hardening (opcional) |
| `SMOKE_TEST_URL` | reset-dev | URL para smoke test (opcional) |
