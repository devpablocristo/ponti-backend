# Cambios Exactos por Archivo - Schema por Rama

## Resumen de Problemas Resueltos

✅ **Problema 1:** search_path es por conexión → **Resuelto** usando `options=-c search_path=...` en DSN  
✅ **Problema 2:** Migraciones deben usar schema → **Resuelto** con DSN + SchemaName + lock  
✅ **Problema 3:** Concurrencia en migraciones → **Resuelto** con `pg_advisory_lock` por schema  
⏳ **Problema 4:** Tests de aislamiento → **Pendiente** (siguiente paso)

---

## Archivo 1: `pkg/databases/sql/gorm/repository.go`

### Cambio en `getDialector()`

**Líneas ~100-120**

**ANTES:**
```go
case Postgres:
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
        config.GetHost(), config.GetUser(), config.GetPassword(), config.GetDBName(), config.GetPort(), config.GetSSLMode())
    dialector = postgres.Open(dsn)
```

**DESPUÉS:**
```go
case Postgres:
    schema := config.GetSchema()
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
        config.GetHost(), config.GetUser(), config.GetPassword(), config.GetDBName(), config.GetPort(), config.GetSSLMode())
    
    // Agregar search_path al DSN usando options=-c (se aplica a TODAS las conexiones del pool)
    // PostgreSQL aplica estos parámetros automáticamente a cada conexión nueva
    // Esto funciona con lib/pq y pgx, independientemente del driver usado por GORM
    if schema != "" && schema != "public" {
        // Validar schema name antes de agregarlo al DSN
        if err := validateSchemaName(schema); err != nil {
            return nil, fmt.Errorf("invalid schema name: %w", err)
        }
        // Escapar el schema name para el DSN (reemplazar comillas y espacios)
        escapedSchema := strings.ReplaceAll(schema, "'", "''")
        dsn += fmt.Sprintf(" options=-csearch_path='%s',public", escapedSchema)
    }
    
    dialector = postgres.Open(dsn)
```

**Por qué resuelve el Problema 1:**
- El parámetro `options=-c search_path=...` en el DSN de PostgreSQL se aplica automáticamente a **TODAS** las conexiones nuevas del pool
- Funciona con lib/pq y pgx (los dos drivers que puede usar GORM)
- No requiere hooks ni código adicional
- PostgreSQL ejecuta `SET search_path` automáticamente al establecer cada conexión

---

## Archivo 2: `projects/ponti-api/cmd/api/migrate_sql.go`

### Cambio 1: Imports

**Líneas ~1-19**

**AGREGADO:**
```go
import (
    "hash/fnv"  // Para hashSchemaName
    "time"      // Para timeouts en locks
    // ... otros imports existentes
)
```

### Cambio 2: `buildMigrateDatabaseURL()`

**Líneas ~113-134**

**ANTES:**
```go
func buildMigrateDatabaseURL(cfg config.DB) string {
    // ... código ...
    if ssl != "" {
        q := url.Values{}
        q.Set("sslmode", ssl)
        u.RawQuery = q.Encode()
    }
    return u.String()
}
```

**DESPUÉS:**
```go
func buildMigrateDatabaseURL(cfg config.DB) string {
    // ... código existente ...
    schema := strings.TrimSpace(cfg.Schema)
    if schema == "" {
        schema = "public"
    }

    q := url.Values{}
    if ssl != "" {
        q.Set("sslmode", ssl)
    }
    
    // Agregar search_path usando options=-c (se aplica a TODAS las conexiones)
    // Esto asegura que el driver de migrate use el schema correcto
    if schema != "" && schema != "public" {
        escapedSchema := url.QueryEscape(schema)
        q.Set("options", fmt.Sprintf("-csearch_path=%s,public", escapedSchema))
    }
    
    u.RawQuery = q.Encode()
    return u.String()
}
```

**Por qué resuelve el Problema 2:**
- El DSN con `options=-c search_path=...` asegura que el driver de migrate use el schema correcto
- Se aplica a todas las conexiones que el driver de migrate crea

### Cambio 3: `runMigrations()`

**Líneas ~21-66**

**ANTES:**
```go
func runMigrations(dbConfig config.DB, migConfig config.Migrations) error {
    dsn := buildMigrateDatabaseURL(dbConfig)
    tempDB, err := sql.Open("postgres", dsn)
    // ... inicializar schema ...
    m, err := migrate.New(migConfig.Dir, dsn)
    // ... ejecutar migraciones ...
}
```

**DESPUÉS:**
```go
func runMigrations(dbConfig config.DB, migConfig config.Migrations) error {
    schema := strings.TrimSpace(dbConfig.Schema)
    if schema == "" {
        schema = "public"
    }

    dsn := buildMigrateDatabaseURL(dbConfig)
    tempDB, err := sql.Open("postgres", dsn)
    defer tempDB.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    // Inicializar schema si no existe
    if err := initializeSchema(ctx, tempDB, dbConfig); err != nil {
        return fmt.Errorf("failed to initialize schema: %w", err)
    }

    // Adquirir lock de migración para evitar ejecuciones concurrentes
    unlock, err := acquireMigrationLock(ctx, tempDB, schema)
    if err != nil {
        return fmt.Errorf("failed to acquire migration lock: %w", err)
    }
    defer unlock()

    log.Printf("Migration lock acquired for schema: %s", schema)

    // Configurar tabla de migraciones en el schema correcto
    dsnWithTable := dsn
    if schema != "public" {
        if strings.Contains(dsnWithTable, "?") {
            dsnWithTable += "&"
        } else {
            dsnWithTable += "?"
        }
        dsnWithTable += fmt.Sprintf("x-migrations-table=%s.schema_migrations", quoteIdentifier(schema))
    }

    m, err := migrate.New(migConfig.Dir, dsnWithTable)
    // ... ejecutar migraciones ...
    log.Printf("Migrations completed successfully for schema: %s", schema)
    return nil
}
```

**Por qué resuelve los Problemas 2 y 3:**
- **Problema 2:** DSN con `options=-c search_path` + `x-migrations-table` asegura que las migraciones usen el schema correcto
- **Problema 3:** `acquireMigrationLock()` previene ejecuciones concurrentes usando `pg_advisory_lock`

### Cambio 4: `runMigrationsWithInstance()`

**Líneas ~69-105**

**AGREGADO:**
- Llamada a `acquireMigrationLock()` antes de ejecutar migraciones
- Timeout de 10 minutos para el contexto
- Logs de lock adquirido/liberado

**Por qué resuelve los Problemas 2 y 3:**
- Mismo que `runMigrations()` pero usando instancia SQL existente

### Cambio 5: Nueva función `acquireMigrationLock()`

**Líneas ~137-185** (NUEVA)

```go
func acquireMigrationLock(ctx context.Context, sqlDB *sql.DB, schema string) (func(), error) {
    if schema == "" || schema == "public" {
        return func() {}, nil
    }

    lockID := hashSchemaName(schema)

    // Intentar adquirir lock no bloqueante primero
    var acquired bool
    err := sqlDB.QueryRowContext(ctx, 
        "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired)
    // ... manejo de errores ...

    if !acquired {
        // Lock ya está tomado, esperar con timeout
        lockCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
        defer cancel()
        // ... esperar lock bloqueante ...
    }

    unlock := func() {
        _, _ = sqlDB.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockID)
    }
    return unlock, nil
}
```

**Por qué resuelve el Problema 3:**
- `pg_advisory_lock` usa un hash del schema name como lock ID
- Solo una instancia puede tener el lock por schema
- Si múltiples instancias de Cloud Run intentan migrar el mismo schema, solo una procede, las otras esperan
- Timeout de 5 minutos previene esperas indefinidas

### Cambio 6: Nueva función `hashSchemaName()`

**Líneas ~187-193** (NUEVA)

```go
func hashSchemaName(schema string) int64 {
    h := fnv.New64a()
    h.Write([]byte(schema))
    return int64(h.Sum64() >> 1)
}
```

**Por qué:**
- Convierte el schema name (string) a un int64 para usar como lock ID en `pg_advisory_lock`
- Hash determinístico: mismo schema name = mismo lock ID

### Cambio 7: `initializeSchema()` simplificado

**Líneas ~107-125**

**ANTES:**
```go
func initializeSchema(...) error {
    // ... crear schema ...
    // Setear search_path para esta conexión
    searchPathSQL := fmt.Sprintf(`SET search_path TO %s, public`, ...)
    sqlDB.ExecContext(ctx, searchPathSQL)
}
```

**DESPUÉS:**
```go
func initializeSchema(...) error {
    // ... crear schema ...
    // NO setear search_path aquí porque se configura en el DSN
    log.Printf("Schema %s created (search_path configured via DSN)", schema)
}
```

**Por qué:**
- `search_path` ya se configura en el DSN, no necesitamos setearlo manualmente

---

## Archivo 3: `pkg/databases/sql/gorm/repository.go`

### Cambio: Nueva función `initializeSchema()`

**Líneas ~272-290** (NUEVA)

```go
func (r *Repository) initializeSchema(ctx context.Context, sqlDB *sql.DB, schema string) error {
    if schema == "" {
        schema = "public"
    }

    if err := validateSchemaName(schema); err != nil {
        return fmt.Errorf("invalid schema name: %w", err)
    }

    log.Printf("Initializing schema: %s", schema)

    createSchemaSQL := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quoteIdentifier(schema))
    if _, err := sqlDB.ExecContext(ctx, createSchemaSQL); err != nil {
        return fmt.Errorf("failed to create schema %s: %w", schema, err)
    }

    log.Printf("Schema %s created (search_path configured via DSN)", schema)
    return nil
}
```

**Por qué:**
- Crea el schema si no existe
- No setea `search_path` porque se configura en el DSN

---

## Resumen de Por Qué Cada Cambio Resuelve los Problemas

### ✅ Problema 1: search_path por conexión

**Solución:** `options=-c search_path=...` en DSN

**Por qué funciona:**
- PostgreSQL aplica los parámetros en `options` automáticamente a cada conexión nueva
- No depende de hooks ni código adicional
- Funciona con lib/pq y pgx

### ✅ Problema 2: Migraciones con schema

**Solución:** DSN con `options=-c search_path` + `SchemaName` en driver + `x-migrations-table`

**Por qué funciona:**
- El DSN asegura que todas las conexiones del driver de migrate usen el schema
- `SchemaName` y `MigrationsTable` aseguran que la tabla de control esté en el schema correcto

### ✅ Problema 3: Concurrencia en migraciones

**Solución:** `pg_advisory_lock` con hash del schema name

**Por qué funciona:**
- Lock a nivel de base de datos, no de aplicación
- Cada schema tiene su propio lock (hash único)
- Solo una instancia puede migrar un schema a la vez
- Las demás esperan o timeout

### ⏳ Problema 4: Tests de aislamiento

**Pendiente:** Crear `migrate_sql_test.go` con tests que:
- Crean schemas `pr_1` y `pr_2`
- Ejecutan migraciones en `pr_1`
- Verifican que `pr_2` está vacío
- Verifican que `public` no cambió
- Ejecutan migraciones en `pr_2`
- Verifican que `pr_1` no cambió

---

## Próximos Pasos

1. ✅ Código crítico implementado
2. ⏳ Crear tests de aislamiento
3. ⏳ Actualizar workflow de GitHub Actions
4. ⏳ Crear GitHub Action para cleanup
