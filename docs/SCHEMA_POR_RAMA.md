# Schema por Rama - Documentación Técnica

> **⚠️ DEPRECADO**: La estrategia recomendada y la implementada para deploy manual por rama ahora es **DB por rama** (`DB_NAME=branch_<slug>`) y **schema `public`**.
>
> Motivo: existen migraciones que modifican `public.*` (views, etc.). Con “schema por rama” dentro de la misma DB, `public` queda compartido y puede romper previews.
>
> Ver: `docs/DEPLOY.md` (sección “Deploy manual por rama”) y el workflow `.github/workflows/deploy-cloud-run.yml`.

## 📋 Resumen

Sistema de aislamiento de datos y migraciones usando **schemas de PostgreSQL** por rama. Cada deploy manual de feature branch usa su propio schema, evitando conflictos con `develop`.

> **Nota:** Los PRs ya NO disparan deploys automáticos. Solo se deploya cuando se hace merge a `develop` o `main`, o mediante `workflow_dispatch` (deploy manual).

---

## 🏗️ Arquitectura

```
┌─────────────────────────────────────────┐
│         PostgreSQL Database             │
│  ┌───────────────────────────────────┐ │
│  │ Schema: public (compartido)        │ │
│  │ - Extensiones (pgvector, etc)      │ │
│  │ - Funciones SSOT compartidas       │ │
│  └───────────────────────────────────┘ │
│  ┌───────────────────────────────────┐ │
│  │ Schema: pr_123 (aislado)          │ │
│  │ - Todas las tablas                │ │
│  │ - Todas las vistas                │ │
│  │ - schema_migrations               │ │
│  └───────────────────────────────────┘ │
│  ┌───────────────────────────────────┐ │
│  │ Schema: branch_feature_abc (aislado)│ │
│  │ - Todas las tablas                 │ │
│  │ - Todas las vistas                 │ │
│  │ - schema_migrations                │ │
│  └───────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

---

## 🔧 Implementación

### 1. Variable de Entorno: `DB_SCHEMA`

**Ubicación:** `projects/ponti-api/cmd/config/db.go`

```go
type DB struct {
    // ... otros campos ...
    Schema string `envconfig:"DB_SCHEMA" default:"public"` // Schema de PostgreSQL
}
```

**Comportamiento:**
- Si `DB_SCHEMA` no está definida → usa `"public"` (comportamiento legacy)
- Si `DB_SCHEMA` está definida → usa ese schema

---

### 2. Inicialización de Schema

**Ubicación:** `pkg/databases/sql/gorm/repository.go`

Al conectar a la base de datos:

1. **Crear schema si no existe:**
   ```sql
   CREATE SCHEMA IF NOT EXISTS "pr_123";
   ```

2. **Setear search_path:**
   ```sql
   SET search_path TO "pr_123", public;
   ```

**Por qué incluir `public` en search_path:**
- Las extensiones (ej: `pgvector`) viven en `public`
- Las funciones SSOT compartidas pueden estar en `public`
- Permite acceso a recursos compartidos sin duplicar código

---

### 3. Migraciones con Schema

**Ubicación:** `projects/ponti-api/cmd/api/migrate_sql.go`

Las migraciones se ejecutan dentro del schema especificado:

```go
driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
    DatabaseName:    dbConfig.Name,
    SchemaName:      schema,
    MigrationsTable: fmt.Sprintf("%s.schema_migrations", quoteIdentifier(schema)),
})
```

**Tabla de control:** `pr_123.schema_migrations` (una por schema)

---

### 4. Naming Conventions

#### Opción A: `pr_<number>` (Recomendado para PRs)

**Formato:** `pr_123`

**Ventajas:**
- ✅ Corto y claro
- ✅ Fácil de identificar en logs
- ✅ Número único por PR
- ✅ Fácil de limpiar (solo necesitas el número)

**Desventajas:**
- ⚠️ Requiere extraer número del PR (GitHub API o variable de entorno)

**Ejemplo de uso:**
```bash
# En GitHub Actions workflow
DB_SCHEMA=pr_123  # Extraído de ${{ github.event.pull_request.number }}
```

---

#### Opción B: `branch_<slug>_<short_sha>`

**Formato:** `branch_feature_new-migration_a1b2c3d`

**Ventajas:**
- ✅ Identifica rama y commit
- ✅ Único automáticamente
- ✅ Útil para debugging

**Desventajas:**
- ⚠️ Nombres largos (límite PostgreSQL: 63 caracteres)
- ⚠️ Requiere sanitización de nombre de rama

**Ejemplo de uso:**
```bash
# Slug de rama + primeros 7 caracteres del SHA
BRANCH_SLUG=$(echo "$GITHUB_HEAD_REF" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g')
SHORT_SHA=$(echo "$GITHUB_SHA" | cut -c1-7)
DB_SCHEMA="branch_${BRANCH_SLUG}_${SHORT_SHA}"
```

---

**Para deploys manuales (`workflow_dispatch`):** `branch_<slug>` (sin SHA, schema estable por rama)

> **Nota:** El workflow actual solo usa `branch_<slug>` para deploys manuales. Los schemas `pr_*` ya no se crean automáticamente, pero el workflow `cleanup-schema.yml` puede limpiar schemas `pr_*` antiguos si existen.

**Implementación sugerida en GitHub Actions:**
```yaml
- name: Set DB_SCHEMA
  run: |
    if [ -n "${{ github.event.pull_request.number }}" ]; then
      echo "DB_SCHEMA=pr_${{ github.event.pull_request.number }}" >> $GITHUB_ENV
    else
      BRANCH_SLUG=$(echo "${{ github.head_ref }}" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | cut -c1-30)
      SHORT_SHA=$(echo "${{ github.sha }}" | cut -c1-7)
      echo "DB_SCHEMA=branch_${BRANCH_SLUG}_${SHORT_SHA}" >> $GITHUB_ENV
    fi
```

---

### 5. Deploy por Rama en Cloud Run

**En GitHub Actions workflow:**

```yaml
- name: Deploy to Cloud Run
  run: |
    # Determinar schema name
    if [ -n "${{ github.event.pull_request.number }}" ]; then
      DB_SCHEMA="pr_${{ github.event.pull_request.number }}"
    else
      BRANCH_SLUG=$(echo "${{ github.head_ref }}" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | cut -c1-30)
      SHORT_SHA=$(echo "${{ github.sha }}" | cut -c1-7)
      DB_SCHEMA="branch_${BRANCH_SLUG}_${SHORT_SHA}"
    fi
    
    gcloud run deploy "${{ env.SERVICE_NAME }}" \
      --project="${{ env.GCP_PROJECT_ID }}" \
      --region="${{ vars.GCP_REGION }}" \
      --image="$image_uri" \
      --service-account="${{ env.CLOUD_RUN_SERVICE_ACCOUNT }}" \
      --update-env-vars="DB_SCHEMA=${DB_SCHEMA},..." \
      --allow-unauthenticated
```

**Para develop/main (sin schema):**
- No setear `DB_SCHEMA` → usa `public` por defecto
- O setear explícitamente: `DB_SCHEMA=public`

---

### 6. Cleanup de Schemas

#### Script SQL: `scripts/cleanup_schema.sql`

```sql
-- Validaciones de seguridad incluidas
DROP SCHEMA IF EXISTS "pr_123" CASCADE;
```

#### Script Bash: `scripts/cleanup_schema.sh`

```bash
./scripts/cleanup_schema.sh pr_123
```

#### Cuándo ejecutar cleanup:

**Opción A: Manual al cerrar PR**
- GitHub Action que se ejecuta cuando se cierra un PR
- Usa `github.event.pull_request.number` para determinar schema

**Opción B: Automático con TTL**
- Script cron que elimina schemas sin actividad por X días
- Verifica última migración o última conexión

**Ejemplo GitHub Action para cleanup:**
```yaml
name: Cleanup Schema on PR Close

on:
  pull_request:
    types: [closed]

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - name: Drop schema
        run: |
          SCHEMA="pr_${{ github.event.pull_request.number }}"
          # Ejecutar script de cleanup
          ./scripts/cleanup_schema.sh "$SCHEMA"
```

---

## 🔒 Seguridad

### Validaciones Implementadas

1. **Nombres reservados bloqueados:**
   - `public`, `pg_catalog`, `pg_toast`, `information_schema`, `pg_temp`, `pg_toast_temp`

2. **Validación de caracteres:**
   - Solo: letras, números, guiones bajos (`_`), guiones (`-`)
   - No puede empezar con número

3. **Quote de identificadores:**
   - Todos los nombres de schema se escapan con comillas dobles
   - Previene SQL injection

---

## 📝 Flujo Completo

### Escenario: Deploy de PR #123

1. **GitHub Actions detecta PR:**
   ```yaml
   DB_SCHEMA=pr_123
   ```

2. **Deploy a Cloud Run:**
   - Variable `DB_SCHEMA=pr_123` inyectada

3. **App inicia:**
   - GORM se conecta a DB
   - Ejecuta `CREATE SCHEMA IF NOT EXISTS "pr_123"`
   - Ejecuta `SET search_path TO "pr_123", public`

4. **Migraciones se ejecutan:**
   - Tabla de control: `pr_123.schema_migrations`
   - Todas las tablas se crean en `pr_123`

5. **App funciona:**
   - Todas las queries usan `pr_123` por defecto
   - Extensiones de `public` siguen disponibles

6. **PR se cierra:**
   - GitHub Action ejecuta cleanup
   - `DROP SCHEMA "pr_123" CASCADE`
   - Schema eliminado completamente

---

## ⚠️ Consideraciones Importantes

### 1. Extensiones Compartidas

Las extensiones (ej: `pgvector`) deben crearse **una sola vez** en `public`:

```sql
-- Ejecutar una vez en la DB
CREATE EXTENSION IF NOT EXISTS vector SCHEMA public;
```

Luego todos los schemas pueden usarlas porque `public` está en `search_path`.

---

### 2. Funciones SSOT Compartidas

Si tienes funciones en `v3_*_ssot` schemas, asegúrate de que:
- Estén en schemas compartidos (no en `pr_123`)
- O que las vistas las referencien con nombre completo: `v3_core_ssot.function_name()`

---

### 3. Migraciones Existentes

**IMPORTANTE:** Las migraciones existentes **NO** necesitan cambios si:
- No asumen `public` explícitamente
- Usan nombres de tabla sin schema (se resuelven por `search_path`)

**Si una migración tiene:**
```sql
-- ❌ MAL (asume public)
CREATE TABLE public.users (...)
```

**Debe cambiarse a:**
```sql
-- ✅ BIEN (usa search_path actual)
CREATE TABLE users (...)
```

---

### 4. GORM Models

Los modelos de GORM **NO** necesitan cambios. GORM usa `search_path` automáticamente.

---

### 5. Vistas con Referencias entre Schemas

Si una vista en `pr_123` necesita referenciar algo de `public`:

```sql
-- ✅ BIEN: nombre completo
CREATE VIEW pr_123.my_view AS
SELECT * FROM public.shared_table;

-- ✅ BIEN: search_path lo resuelve si está en public
CREATE VIEW pr_123.my_view AS
SELECT * FROM shared_table;  -- Busca en pr_123 primero, luego public
```

---

## 🧪 Testing

### Probar localmente:

```bash
# Setear schema
export DB_SCHEMA=pr_test_123

# Ejecutar app
go run ./cmd/api/

# Verificar que se creó el schema
psql -d ponti_api_db -c "\dn"

# Verificar tablas en el schema
psql -d ponti_api_db -c "\dt pr_test_123.*"

# Limpiar
./scripts/cleanup_schema.sh pr_test_123
```

---

## 📊 Monitoreo

### Ver schemas activos:

```sql
SELECT schema_name, 
       (SELECT COUNT(*) FROM information_schema.tables 
        WHERE table_schema = s.schema_name) as table_count
FROM information_schema.schemata s
WHERE schema_name LIKE 'pr_%' OR schema_name LIKE 'branch_%'
ORDER BY schema_name;
```

### Ver tamaño de schemas:

```sql
SELECT schemaname,
       pg_size_pretty(SUM(pg_total_relation_size(schemaname||'.'||tablename))::bigint) AS size
FROM pg_tables
WHERE schemaname LIKE 'pr_%' OR schemaname LIKE 'branch_%'
GROUP BY schemaname
ORDER BY SUM(pg_total_relation_size(schemaname||'.'||tablename)) DESC;
```

---

## ✅ Checklist de Implementación

- [x] Variable `DB_SCHEMA` agregada a config
- [x] Inicialización de schema en GORM
- [x] Migraciones configuradas para usar schema
- [x] Validaciones de seguridad implementadas
- [x] Scripts de cleanup creados
- [ ] Actualizar GitHub Actions workflow para inyectar `DB_SCHEMA`
- [ ] Crear GitHub Action para cleanup automático
- [ ] Documentar naming convention elegida
- [ ] Probar deploy completo end-to-end
