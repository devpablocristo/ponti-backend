# Guía de Pruebas - Schema por Rama

## 🧪 Pruebas Locales

### 1. Prueba Básica: Crear Schema y Ejecutar Migraciones

```bash
# 1. Levantar la base de datos (si usas Docker)
cd projects/ponti-api
docker compose up -d ponti-db

# 2. Esperar a que la DB esté lista
until docker compose exec -T ponti-db pg_isready -U admin -d ponti_api_db >/dev/null 2>&1; do sleep 1; done

# 3. Setear variables de entorno para un schema de prueba
export DB_SCHEMA=pr_test_123
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ponti_api_db
export DB_SSL_MODE=disable
export MIGRATIONS_DIR=file://migrations

# 4. Ejecutar la app (esto creará el schema y ejecutará migraciones)
cd /home/pablo/Projects/Pablo/ponti-backend/projects/ponti-api
go run ./cmd/api/

# Deberías ver en los logs:
# "Initializing schema: pr_test_123"
# "Schema pr_test_123 created (search_path configured via DSN)"
# "Migration lock acquired for schema: pr_test_123"
# "Migrations completed successfully for schema: pr_test_123"
```

### 2. Verificar que el Schema se Creó Correctamente

```bash
# Conectar a la DB
docker compose exec ponti-db psql -U admin -d ponti_api_db

# En psql:
\dn                    # Listar schemas (deberías ver pr_test_123)
\dt pr_test_123.*      # Listar tablas en pr_test_123
SELECT * FROM pr_test_123.schema_migrations;  # Ver migraciones aplicadas

# Verificar que public no cambió
\dt public.*           # Tablas en public (deberían ser las mismas de antes)
```

### 3. Probar Aislamiento entre Schemas

```bash
# Terminal 1: Crear schema pr_1
export DB_SCHEMA=pr_1
go run ./cmd/api/ &
APP_PID_1=$!

# Esperar a que termine de inicializar
sleep 5

# Terminal 2: Crear schema pr_2
export DB_SCHEMA=pr_2
go run ./cmd/api/ &
APP_PID_2=$!

# Esperar
sleep 5

# Verificar en psql:
docker compose exec ponti-db psql -U admin -d ponti_api_db <<EOF
-- Verificar que ambos schemas existen
SELECT schema_name FROM information_schema.schemata 
WHERE schema_name IN ('pr_1', 'pr_2', 'public')
ORDER BY schema_name;

-- Contar tablas en cada schema
SELECT 
    'pr_1' as schema, COUNT(*) as tables 
FROM information_schema.tables 
WHERE table_schema = 'pr_1' AND table_type = 'BASE TABLE'
UNION ALL
SELECT 
    'pr_2' as schema, COUNT(*) as tables 
FROM information_schema.tables 
WHERE table_schema = 'pr_2' AND table_type = 'BASE TABLE'
UNION ALL
SELECT 
    'public' as schema, COUNT(*) as tables 
FROM information_schema.tables 
WHERE table_schema = 'public' AND table_type = 'BASE TABLE';

-- Verificar que cada schema tiene su propia tabla schema_migrations
SELECT 'pr_1' as schema, version FROM pr_1.schema_migrations LIMIT 1
UNION ALL
SELECT 'pr_2' as schema, version FROM pr_2.schema_migrations LIMIT 1;
EOF

# Matar procesos
kill $APP_PID_1 $APP_PID_2 2>/dev/null || true
```

### 4. Probar el Lock de Migraciones

```bash
# Terminal 1: Iniciar migraciones en pr_lock_test (bloquea)
export DB_SCHEMA=pr_lock_test
go run ./cmd/api/ &
APP_PID=$!

# Terminal 2: Intentar migrar el mismo schema simultáneamente
# (debería esperar o fallar con timeout)
export DB_SCHEMA=pr_lock_test
timeout 30 go run ./cmd/api/ || echo "Timeout esperado - lock funcionando"

# Verificar en logs que el segundo proceso esperó el lock
```

### 5. Probar el Script de Cleanup

```bash
# Crear un schema de prueba
export DB_SCHEMA=pr_cleanup_test
go run ./cmd/api/ &
sleep 5
kill %1 2>/dev/null || true

# Verificar que existe
docker compose exec ponti-db psql -U admin -d ponti_api_db -c "\dn pr_cleanup_test"

# Ejecutar cleanup
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=admin
export DB_PASSWORD=admin
export DB_NAME=ponti_api_db
export DB_SSL_MODE=disable

./scripts/cleanup_schema.sh pr_cleanup_test

# Verificar que se eliminó
docker compose exec ponti-db psql -U admin -d ponti_api_db -c "\dn pr_cleanup_test"
# Debería mostrar "No matching schemas found"
```

### 6. Ejecutar Tests Unitarios

```bash
cd projects/ponti-api
go test ./cmd/api/ -v -run 'TestValidateSchemaName|TestQuoteIdentifier|TestHashSchemaName'
```

### 7. Ejecutar Tests de Integración (requiere DB)

```bash
# Setear variables de test
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=admin
export TEST_DB_PASSWORD=admin
export TEST_DB_NAME=ponti_api_db
export TEST_DB_SSL_MODE=disable
export TEST_MIGRATIONS_DIR=file://migrations

# Ejecutar tests
cd projects/ponti-api
go test ./cmd/api/ -v -run 'TestSchemaIsolation|TestMigrationLock'
```

---

## 🚀 Pruebas en GitHub Actions

### 1. Probar Deploy con Schema de PR

**Opción A: Crear un PR de prueba**

1. Crear una rama nueva:
   ```bash
   git checkout -b test/pr-schema-123
   git commit --allow-empty -m "Test PR schema"
   git push origin test/pr-schema-123
   ```

2. Crear un PR en GitHub (#123 por ejemplo)

3. **IMPORTANTE:** El workflow actual NO se dispara automáticamente en PRs.
   Para probar, necesitas:
   
   **Opción 1:** Agregar trigger de PR al workflow (recomendado):
   ```yaml
   on:
     push:
       branches: [develop, staging, main]
     pull_request:
       types: [opened, synchronize]
     workflow_dispatch:
   ```
   
   **Opción 2:** Usar `workflow_dispatch` manualmente:
   - Ir a Actions → Deploy to Cloud Run
   - Click "Run workflow"
   - Seleccionar la rama `test/pr-schema-123`
   - Ejecutar

4. Verificar en los logs del workflow:
   ```
   DB_SCHEMA set to: pr_123  (o branch_... si no es PR)
   ```

5. Verificar en Cloud Run:
   ```bash
   gcloud run services describe ponti-backend-dev \
     --project=new-ponti-dev \
     --region=us-central1 \
     --format="value(spec.template.spec.containers[0].env)" | grep DB_SCHEMA
   ```

### 2. Probar Deploy a Develop/Main (debería usar public)

```bash
# Mergear a develop
git checkout develop
git merge test/pr-schema-123
git push origin develop

# Verificar en logs del workflow:
# DB_SCHEMA set to: public

# Verificar en Cloud Run que DB_SCHEMA=public
```

### 3. Probar Cleanup al Cerrar PR

1. Cerrar el PR de prueba (#123)
2. Verificar que se ejecuta el workflow `cleanup-schema.yml`
3. Verificar en los logs:
   ```
   Cleaning up schema: pr_123
   Schema pr_123 dropped successfully
   ```
4. Verificar en la DB que el schema se eliminó:
   ```bash
   # Conectar a Cloud SQL (ajusta credenciales)
   psql -h <DB_HOST> -U <DB_USER> -d <DB_NAME> -c "\dn pr_123"
   # Debería mostrar "No matching schemas found"
   ```

---

## ✅ Checklist de Verificación

### Local
- [ ] Schema se crea correctamente con `DB_SCHEMA=pr_test_123`
- [ ] Migraciones se ejecutan en el schema correcto
- [ ] Tabla `schema_migrations` está en el schema correcto (`pr_test_123.schema_migrations`)
- [ ] Schema `public` no se modifica
- [ ] Dos schemas diferentes (`pr_1` y `pr_2`) están aislados
- [ ] Lock de migraciones funciona (dos procesos no pueden migrar el mismo schema simultáneamente)
- [ ] Script de cleanup elimina el schema correctamente
- [ ] Tests unitarios pasan

### GitHub Actions / Cloud Run
- [ ] Deploy de PR crea schema `pr_<number>`
- [ ] Deploy a `develop` usa `public`
- [ ] Deploy a `main` usa `public`
- [ ] Variable `DB_SCHEMA` se inyecta correctamente en Cloud Run
- [ ] Cleanup elimina schema al cerrar PR
- [ ] Logs de Cloud Run muestran el schema correcto

---

## 🔍 Comandos de Verificación Rápida

### Ver schemas activos en la DB:
```sql
SELECT schema_name, 
       (SELECT COUNT(*) FROM information_schema.tables 
        WHERE table_schema = s.schema_name 
        AND table_type = 'BASE TABLE') as table_count
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

### Ver locks activos de migraciones:
```sql
SELECT 
    locktype, 
    objid, 
    mode, 
    granted,
    pg_advisory_unlock_all() -- Solo para testing, no ejecutar en producción
FROM pg_locks 
WHERE locktype = 'advisory';
```

---

## 🐛 Troubleshooting

### Problema: "Schema no se crea"
- Verificar que `DB_SCHEMA` está seteado
- Verificar logs de la app: debería mostrar "Initializing schema: ..."
- Verificar permisos del usuario de DB

### Problema: "Migraciones no se ejecutan"
- Verificar que `MIGRATIONS_DIR` está correcto
- Verificar logs: debería mostrar "Migration lock acquired"
- Verificar que no hay otro proceso migrando el mismo schema

### Problema: "Cleanup no funciona"
- Verificar que el script tiene permisos de ejecución: `chmod +x scripts/cleanup_schema.sh`
- Verificar variables de entorno en el workflow
- Verificar que la DB es accesible desde GitHub Actions (IP pública o Cloud SQL Proxy)

### Problema: "search_path no funciona"
- Verificar que el DSN tiene `options=-c search_path='...',public`
- Verificar logs de conexión de GORM
- Probar manualmente: `SET search_path TO pr_test_123, public;`
