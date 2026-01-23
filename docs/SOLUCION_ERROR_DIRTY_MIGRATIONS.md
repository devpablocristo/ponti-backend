# Solución: Error de Estado Dirty en Migraciones - Auto-Recuperación

> **⚠️ ACTUALIZACIÓN (2026)**: La estrategia de deploy manual por rama pasó a ser **DB por rama** (`DB_NAME=branch_<slug>`) y el backend ya no hace auto-recuperación por schema.
>
> **Cómo resolver un dirty state en preview hoy**: re-ejecutar el workflow manual con `reset_db=true` para borrar y recrear la DB de esa rama.
>
> Ver: `docs/DEPLOY.md` y `.github/workflows/deploy-cloud-run.yml`.

## 📋 Contexto del Problema

### ¿Qué estaba pasando?

El workflow de GitHub Actions `deploy-cloud-run.yml` tiene configurado un trigger automático para Pull Requests:

```yaml
on:
  pull_request:
    types: [opened, synchronize, reopened]
    branches:
      - develop
      - staging
```

**Estado actual:**
- Los PRs ya NO disparan deploys automáticos (eliminado para evitar deploys innecesarios)
- El workflow solo se ejecuta en `push` a `develop`/`main` (merges) o `workflow_dispatch` (deploy manual)
- Los deploys manuales usan DB aislada (`DB_NAME=branch_<slug>`, schema `public`)
- Los merges a `develop`/`main` usan la DB fija configurada en cada servicio Cloud Run

### El Error Específico

Cuando el workflow se ejecutaba (o se ejecuta en deploy manual), el contenedor de Cloud Run puede fallar al iniciar con el siguiente error:

```
ERROR: (gcloud.run.deploy) The user-provided container failed to start and listen on the port defined provided by the PORT=8080 environment variable within the allocated timeout.
```

**Pero el error real** (visible en los logs de Cloud Run) era:

```
Failed to run SQL migrations: error applying migrations: Dirty database version 79. Fix and force version.
```

## 🔍 Causa Raíz

### ¿Qué es un estado "dirty" en migraciones?

`golang-migrate` marca una migración como "dirty" cuando:

1. **Una migración falla durante su ejecución** (por ejemplo, error de sintaxis SQL, timeout, constraint violation)
2. **El proceso se interrumpe** antes de completar (crash del contenedor, timeout, cancelación manual)
3. **Hay un error de conexión** durante la ejecución de la migración

Cuando una migración queda en estado "dirty":
- `golang-migrate` **bloquea** la ejecución de nuevas migraciones
- Requiere **intervención manual** para limpiar el estado
- La tabla `schema_migrations` tiene `dirty = true` para la versión que falló

### ¿Por qué ocurría en nuestro caso?

**Escenario típico:**
1. Un deploy se ejecuta (manual o automático por merge)
2. El workflow despliega a Cloud Run con schema aislado (ej: `branch_feature-x` o `pr_5` si es antiguo)
3. El contenedor intenta ejecutar migraciones al iniciar
4. Si una migración anterior falló o el contenedor se reinició durante una migración, queda estado dirty
5. El nuevo deploy intenta ejecutar migraciones → detecta estado dirty → **falla**
6. El contenedor no puede iniciar → Cloud Run reporta error de timeout

**Problema adicional:**
- Los schemas aislados (`branch_*`, `pr_*`) son temporales y aislados
- Si un deploy falla, el schema queda con estado dirty
- El siguiente deploy (manual o automático) también falla
- Esto crea un ciclo de fallos hasta que alguien limpie manualmente el estado dirty

## ✅ Solución Implementada

### Estrategia: Recreación de Schema desde Cero

**IMPORTANTE:** La solución NO hace `UPDATE dirty=false` porque eso puede dejar el schema inconsistente.

En su lugar, implementa **recreación completa del schema** cuando se detecta estado dirty:

1. **Schema `public`:** Falla con error claro (requiere intervención manual)
2. **Schemas aislados (`pr_*`, `branch_*`):**
   - Adquiere advisory lock para prevenir concurrencia
   - Ejecuta `DROP SCHEMA <schema> CASCADE` (elimina todo)
   - Ejecuta `CREATE SCHEMA <schema>` (schema limpio)
   - Re-ejecuta todas las migraciones desde 0
   - Libera advisory lock

### Cambios Técnicos Realizados

#### 1. Detección de Errores Dirty

Se modificaron las funciones `runMigrations()` y `runMigrationsWithInstance()` para detectar errores de estado dirty:

```go
// Intentar ejecutar migraciones
err = m.Up()
if err != nil && err != migrate.ErrNoChange {
    // Verificar si es un error de dirty state
    if strings.Contains(err.Error(), "Dirty database version") || strings.Contains(err.Error(), "dirty") {
        // Para public: fallar con error claro
        if schema == "public" {
            return fmt.Errorf("dirty migration state in public schema - manual intervention required: %w", err)
        }
        
        // Para schemas aislados: recrear schema desde cero
        log.Printf("⚠️  Dirty migration state detected for schema %s, recreating schema from scratch...", schema)
        
        if err := recreateSchemaOnDirty(ctx, tempDB, schema); err != nil {
            return fmt.Errorf("failed to recreate schema on dirty state: %w", err)
        }
        
        // Re-ejecutar migraciones desde 0
        // ...
    }
}
```

#### 2. Función de Recreación de Schema

Se creó la función `recreateSchemaOnDirty()` que:

- **Adquiere advisory lock** antes de hacer DROP/CREATE (previene concurrencia)
- **Ejecuta `DROP SCHEMA CASCADE`** para eliminar todo el contenido
- **Ejecuta `CREATE SCHEMA`** para crear schema limpio
- **Solo funciona** para schemas aislados (`pr_*`, `branch_*`)
- **Nunca** modifica el schema `public` (seguridad)

```go
func recreateSchemaOnDirty(ctx context.Context, sqlDB *sql.DB, schema string) error {
    if schema == "public" {
        return fmt.Errorf("cannot recreate public schema (requires manual intervention)")
    }

    // Adquirir advisory lock para prevenir concurrencia
    lockID := hashSchemaName(schema)
    // ... adquirir lock ...
    
    // DROP SCHEMA CASCADE - elimina todo el contenido
    dropSQL := fmt.Sprintf(`DROP SCHEMA IF EXISTS %s CASCADE`, quoteIdentifier(schema))
    // ... ejecutar DROP ...
    
    // CREATE SCHEMA - crear schema limpio
    createSQL := fmt.Sprintf(`CREATE SCHEMA %s`, quoteIdentifier(schema))
    // ... ejecutar CREATE ...
    
    // Liberar lock
    // ...
}
```

#### 3. Advisory Lock para Concurrencia

**Problema:** En Cloud Run, múltiples instancias pueden intentar migrar simultáneamente.

**Solución:** Advisory lock mejorado con loop y timeout configurable:
- **Lock ID:** `int64/bigint` compatible (hash determinístico del schema name)
- **Estrategia:** Loop con `pg_try_advisory_lock` (no bloqueante) en lugar de `pg_advisory_lock` bloqueante
- **Timeout configurable:** 5 minutos por defecto
- **Backoff progresivo:** 100ms → 500ms → 1s (reduce carga en DB)
- **Logs mejorados:** Incluyen schema name, lock_id, wait time
- **Unlock garantizado:** Siempre se libera en `defer`

**Implementación:**
```go
lockID := hashSchemaName(schema)  // Hash determinístico → int64
// Loop con pg_try_advisory_lock hasta adquirir o timeout
for {
    acquired := pg_try_advisory_lock(lockID)
    if acquired {
        break
    }
    // Backoff progresivo y logs periódicos
    time.Sleep(backoff)
}
// ... ejecutar migraciones o recrear schema ...
defer pg_advisory_unlock(lockID)  // Garantizado
```

**Logs de ejemplo:**
```
Attempting to acquire migration lock for schema: pr_5 (lock_id: 1234567890)
✅ Migration lock acquired immediately for schema: pr_5 (lock_id: 1234567890)
...
🔓 Migration lock released for schema: pr_5 (lock_id: 1234567890)
```

Si espera:
```
Attempting to acquire migration lock for schema: pr_5 (lock_id: 1234567890)
⏳ Waiting for migration lock (schema: pr_5, lock_id: 1234567890, waited: 5s)...
✅ Migration lock acquired for schema: pr_5 (lock_id: 1234567890, waited: 12.5s)
```

**Cambio importante:** Ahora también se usa para `public` (antes solo para schemas aislados):

#### 4. Re-ejecución de Migraciones

Después de recrear el schema:

1. Se **recrea** la instancia de migrate (nueva conexión)
2. Se **re-ejecutan** todas las migraciones desde 0 (`m.Up()`)
3. Si el reintento falla, se reporta el error real (no el dirty state)
4. Esto garantiza que el schema quede en un estado consistente

### Flujo Completo de la Solución

```
1. PR se abre/sincroniza
   ↓
2. Workflow ejecuta deploy a Cloud Run
   ↓
3. Contenedor inicia → intenta ejecutar migraciones
   ↓
4. ¿Hay estado dirty?
   ├─ NO → Ejecuta migraciones normalmente ✅
   └─ SÍ → ¿Es schema public?
       ├─ SÍ → ❌ Falla con error claro (requiere intervención manual)
       └─ NO → Adquiere advisory lock
                ↓
                DROP SCHEMA CASCADE
                ↓
                CREATE SCHEMA
                ↓
                Re-ejecuta migraciones desde 0
                ↓
                Libera advisory lock
                ↓
                ✅ Éxito
```

## 🛡️ Consideraciones de Seguridad

### ¿Por qué no auto-limpiar `public`?

El schema `public` es el schema principal de producción. Si tiene estado dirty:

- **Puede indicar un problema real** que necesita investigación
- **Auto-limpiar podría ocultar problemas** críticos
- **Requiere decisión consciente** del equipo

Por lo tanto:
- ✅ **Schemas aislados** (`pr_*`, `branch_*`): Recreación automática habilitada
- ❌ **Schema `public`**: Recreación deshabilitada (requiere intervención manual)

### ¿Por qué DROP/CREATE en lugar de UPDATE dirty=false?

**Problema con UPDATE dirty=false:**
- Puede dejar el schema en estado inconsistente
- Si la migración falló a mitad de camino, el schema puede tener tablas parciales
- Reintentar migraciones puede fallar por objetos ya existentes

**Solución con DROP/CREATE:**
- Garantiza estado completamente limpio
- Elimina cualquier inconsistencia previa
- Re-ejecutar desde 0 asegura consistencia total
- Es seguro para schemas aislados (no afecta producción)

## 📊 Beneficios de la Solución

1. **Deploys automáticos de PR funcionan sin intervención manual**
   - Los PRs pueden sincronizarse múltiples veces sin fallar por estado dirty

2. **Garantía de consistencia**
   - Recrear desde cero elimina cualquier inconsistencia
   - No hay riesgo de dejar el schema en estado parcial

3. **Prevención de concurrencia**
   - Advisory lock previene que múltiples instancias migren simultáneamente
   - Especialmente importante en Cloud Run con múltiples réplicas

4. **Seguridad mantenida**
   - El schema `public` nunca se auto-recrea, protegiendo producción

5. **Mejor experiencia de desarrollo**
   - Los desarrolladores no necesitan limpiar manualmente schemas de PR
   - El sistema se recupera automáticamente de fallos transitorios

## 🔄 Comportamiento Final

### Escenarios Cubiertos

#### Escenario 1: Deploy Normal (sin estado dirty)
```
Deploy ejecutado → Migraciones ejecutan → ✅ Éxito
```

#### Escenario 2: Deploy con Estado Dirty Previo
```
Deploy ejecutado → Detecta dirty → Adquiere lock → DROP SCHEMA CASCADE → CREATE SCHEMA → Re-ejecuta migraciones desde 0 → ✅ Éxito
```

#### Escenario 3: Deploy Manual de Rama
```
workflow_dispatch → Deploy → Schema branch_<slug> → Si hay dirty → Adquiere lock → DROP/CREATE → Re-ejecuta migraciones → ✅ Éxito
```

#### Escenario 4: Deploy a Producción (public)
```
Push a main → Deploy → Schema public → Si hay dirty → ❌ Falla con mensaje claro
```

## 📝 Archivos Modificados

- `projects/ponti-api/cmd/api/migrate_sql.go`
  - Función `runMigrations()`: Agregado manejo de errores dirty con recreación de schema
  - Función `runMigrationsWithInstance()`: Agregado manejo de errores dirty con recreación de schema
  - Nueva función `recreateSchemaOnDirty()`: Implementa DROP/CREATE con advisory lock mejorado
  - Función `acquireMigrationLock()`: Wrapper con timeout de 5 minutos
  - Función `acquireMigrationLockWithTimeout()`: Loop con `pg_try_advisory_lock`, backoff progresivo, logs mejorados

## 🧪 Cómo Probar

1. **Simular estado dirty manualmente:**
   ```sql
   UPDATE pr_5.schema_migrations SET dirty = true WHERE version = 79;
   ```

2. **Hacer push al PR o ejecutar deploy manual**

3. **Verificar en logs:**
   - Debe aparecer: `⚠️ Dirty migration state detected for schema pr_5, recreating schema from scratch...`
   - Debe aparecer: `Attempting to acquire migration lock for schema: pr_5 (lock_id: ...)`
   - Debe aparecer: `✅ Migration lock acquired immediately for schema: pr_5 (lock_id: ...)`
   - Debe aparecer: `Dropping schema pr_5 (CASCADE)...`
   - Debe aparecer: `Creating fresh schema pr_5...`
   - Debe aparecer: `✅ Schema pr_5 recreated successfully (ready for fresh migrations)`
   - Debe aparecer: `Re-running migrations from scratch for schema pr_5...`
   - Debe aparecer: `Migrations completed successfully for schema: pr_5`
   - Debe aparecer: `🔓 Migration lock released for schema: pr_5 (lock_id: ...)`

4. **Verificar que el contenedor inicia correctamente**

5. **Verificar en la base de datos:**
   ```sql
   -- El schema debe existir y estar limpio
   SELECT * FROM pr_5.schema_migrations ORDER BY version DESC LIMIT 5;
   -- Debe mostrar migraciones completas sin dirty=true
   ```

## 🎯 Conclusión

La solución implementa **recreación automática de schemas** cuando se detecta estado dirty, permitiendo que los deploys funcionen sin intervención manual, mientras mantiene la seguridad al requerir intervención manual para el schema `public` de producción.

**Estrategia clave:** En lugar de limpiar el estado dirty (que puede dejar inconsistencias), el sistema recrea completamente el schema desde cero, garantizando un estado limpio y consistente.

**Importante:** La lógica de `DB_SCHEMA` ahora verifica `workflow_dispatch` ANTES de `ref_name` para garantizar que los deploys manuales siempre usen schemas aislados, incluso si se disparan desde `main`/`develop`.

**Concurrencia:** El uso de advisory locks previene que múltiples instancias de Cloud Run migren simultáneamente, tanto en el camino normal como durante la recreación de schemas.
