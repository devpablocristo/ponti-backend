# Cambios Críticos para Schema por Rama

## Resumen de Problemas y Soluciones

### Problema 1: search_path es por conexión
**Riesgo:** Solo setear search_path una vez no funciona con pools de conexiones. Cada conexión nueva del pool no tendrá el search_path configurado.

**Solución:** Usar el parámetro `options` en el DSN de PostgreSQL que se aplica automáticamente a TODAS las conexiones del pool, independientemente del driver (lib/pq o pgx).

### Problema 2: Migraciones deben usar el mismo schema
**Riesgo:** El driver del migrator puede usar una conexión diferente sin search_path configurado.

**Solución:** Configurar el DSN de migraciones con `options=-c search_path=...` y usar `SchemaName` en el driver de migrate.

### Problema 3: Concurrencia en migraciones
**Riesgo:** Múltiples instancias de Cloud Run pueden intentar correr migraciones simultáneamente, causando conflictos.

**Solución:** Usar `pg_advisory_lock` con un hash del schema name para lockear migraciones por schema.

### Problema 4: Tests de aislamiento
**Riesgo:** Sin tests, no podemos verificar que los schemas están realmente aislados.

**Solución:** Tests que crean schemas pr_1 y pr_2, ejecutan migraciones, y verifican que no se afectan entre sí ni a public.

---

## Archivos Modificados

### 1. `pkg/databases/sql/gorm/repository.go`

**Cambio:** Modificar `getDialector()` para agregar `options=-c search_path=...` al DSN.

**Por qué resuelve el problema 1:**
- PostgreSQL aplica los parámetros en `options` a nivel de conexión automáticamente
- Funciona con lib/pq y pgx
- Se aplica a TODAS las conexiones nuevas del pool sin código adicional

```go
case Postgres:
    schema := config.GetSchema()
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
        config.GetHost(), config.GetUser(), config.GetPassword(), config.GetDBName(), config.GetPort(), config.GetSSLMode())
    
    // Agregar search_path al DSN usando options (se aplica a TODAS las conexiones)
    if schema != "" && schema != "public" {
        if err := validateSchemaName(schema); err != nil {
            return nil, fmt.Errorf("invalid schema name: %w", err)
        }
        // Usar options=-c para configurar parámetros de PostgreSQL
        // Esto se aplica automáticamente a todas las conexiones del pool
        dsn += fmt.Sprintf(" options=-csearch_path=%s,public", schema)
    }
    
    dialector = postgres.Open(dsn)
```

---

### 2. `projects/ponti-api/cmd/api/migrate_sql.go`

**Cambios:**
1. Modificar `buildMigrateDatabaseURL()` para agregar `options=-c search_path=...`
2. Agregar función `acquireMigrationLock()` con `pg_advisory_lock`
3. Modificar `runMigrations()` y `runMigrationsWithInstance()` para usar el lock

**Por qué resuelve los problemas 2 y 3:**

**Problema 2 (migraciones con schema):**
- El DSN con `options=-c search_path=...` asegura que el driver de migrate use el schema correcto
- `SchemaName` en el Config del driver asegura que la tabla de control esté en el schema correcto

**Problema 3 (concurrencia):**
- `pg_advisory_lock` usa un hash del schema name como lock ID
- Solo una instancia puede tener el lock por schema
- Si otra instancia intenta migrar el mismo schema, esperará hasta que se libere el lock

```go
// acquireMigrationLock adquiere un lock de migración usando pg_advisory_lock
// El lock ID se deriva del schema name para que cada schema tenga su propio lock
func acquireMigrationLock(ctx context.Context, sqlDB *sql.DB, schema string) (func(), error) {
    // Calcular hash del schema name para usar como lock ID
    // Usamos un hash simple para convertir string a int64
    lockID := hashSchemaName(schema)
    
    // Intentar adquirir el lock (bloquea hasta obtenerlo o timeout)
    var acquired bool
    err := sqlDB.QueryRowContext(ctx, 
        "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
    if err != nil {
        return nil, fmt.Errorf("failed to acquire migration lock: %w", err)
    }
    
    if !acquired {
        // Lock ya está tomado, esperar con timeout
        ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
        defer cancel()
        
        err := sqlDB.QueryRowContext(ctx,
            "SELECT pg_advisory_lock($1)", lockID).Scan()
        if err != nil && err != sql.ErrNoRows {
            return nil, fmt.Errorf("failed to wait for migration lock: %w", err)
        }
    }
    
    // Retornar función de unlock
    unlock := func() {
        _, _ = sqlDB.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockID)
    }
    
    return unlock, nil
}

// hashSchemaName convierte un schema name a un int64 para usar como lock ID
func hashSchemaName(schema string) int64 {
    h := fnv.New64a()
    h.Write([]byte(schema))
    return int64(h.Sum64() >> 1) // Convertir a int64 (pg_advisory_lock usa bigint)
}
```

---

### 3. `projects/ponti-api/cmd/api/migrate_sql_test.go` (NUEVO)

**Por qué resuelve el problema 4:**
- Tests que verifican aislamiento entre schemas
- Tests que verifican que public no se modifica
- Tests que verifican que las migraciones funcionan correctamente

```go
func TestSchemaIsolation(t *testing.T) {
    // Setup: crear schemas pr_1 y pr_2
    // Ejecutar migraciones en pr_1
    // Verificar que pr_2 está vacío
    // Verificar que public no cambió
    // Ejecutar migraciones en pr_2
    // Verificar que pr_1 no cambió
}
```

---

### 4. `.github/workflows/deploy-cloud-run.yml`

**Cambios:**
1. Agregar step para determinar `DB_SCHEMA`
2. Inyectar `DB_SCHEMA` en Cloud Run
3. Agregar job de cleanup al cerrar PR

**Por qué resuelve el deploy:**
- Determina automáticamente el schema según el contexto (PR vs branch)
- Inyecta la variable en Cloud Run para que la app la use
- Limpia schemas automáticamente al cerrar PRs

---

## Implementación Completa

Ahora voy a implementar todos estos cambios.
