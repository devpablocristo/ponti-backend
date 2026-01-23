# Cambios Implementados - Schema por Rama/PR

## 📋 Resumen Ejecutivo

Se implementaron los cambios requeridos para cerrar correctamente la estrategia de "schema por rama/PR" sin modificar el uso de `public` en `develop` y `main`, ni la estructura general del pipeline.

---

## ✅ Cambios Implementados

### 1. **Agregado trigger `pull_request` al workflow deploy-cloud-run.yml**

**Ubicación:** `.github/workflows/deploy-cloud-run.yml` (líneas 9-13)

**Cambio:**
```yaml
pull_request:
  types: [opened, synchronize, reopened]
  branches:
    - develop
    - staging
```

**Por qué es necesario:**
- Permite que los PRs se deployen automáticamente cuando se abren, sincronizan o reabren
- Solo se ejecuta para PRs cuyo target es `develop` o `staging` (no `main`)
- Ahora los PRs crearán schemas `pr_<number>` que pueden ser limpiados automáticamente

**Impacto:**
- ✅ Los PRs ahora se deployan automáticamente
- ✅ Se crean schemas `pr_<number>` que el cleanup existente puede eliminar
- ✅ No afecta el comportamiento de `push` a `develop`/`main`

---

### 2. **Modificada lógica de determinación de DB_SCHEMA**

**Ubicación:** `.github/workflows/deploy-cloud-run.yml` (líneas 171-203)

**Cambios específicos:**

#### A. Prioridad para `schema_override` manual
```bash
if [ -n "$SCHEMA_OVERRIDE" ]; then
    DB_SCHEMA="$SCHEMA_OVERRIDE"
    # ... usar override directamente
fi
```

#### B. Detección de evento `pull_request`
```bash
if [ "${{ github.event_name }}" = "pull_request" ]; then
    PR_NUMBER="${{ github.event.pull_request.number }}"
    DB_SCHEMA="pr_${PR_NUMBER}"
fi
```

#### C. Eliminación de SHA en `workflow_dispatch`
**Antes:**
```bash
DB_SCHEMA="branch_${BRANCH_SLUG}_${SHORT_SHA}"  # ❌ Incluía SHA
```

**Ahora:**
```bash
DB_SCHEMA="branch_${BRANCH_SLUG}"  # ✅ Sin SHA, schema estable por rama
```

**Por qué es necesario:**
- **Schema estable:** `branch_<slug>` permite reutilizar el mismo schema en múltiples deploys de la misma rama
- **Sin SHA:** Evita crear múltiples schemas innecesarios cuando se hace push múltiples veces
- **PRs con `pr_<number>`:** Permite que el cleanup existente funcione correctamente

**Impacto:**
- ✅ Schemas `branch_*` son reutilizables (mismo schema para la misma rama)
- ✅ Los PRs usan `pr_<number>` que se limpia automáticamente
- ✅ `develop`/`main`/`staging` siguen usando `public`

---

### 3. **Actualizada lógica de "Set deploy variables" para PRs**

**Ubicación:** `.github/workflows/deploy-cloud-run.yml` (líneas 84-97)

**Cambio:**
```bash
# Pull Request: siempre deployar a DEV
if [ "$EVENT_NAME" = "pull_request" ]; then
    # Validar que el PR apunta a develop o staging
    if [ "$PR_BASE_REF" != "develop" ] && [ "$PR_BASE_REF" != "staging" ]; then
        echo "PRs can only target develop or staging branches" >&2
        exit 1
    fi
    deploy_env="${{ vars.DEPLOY_ENV_DEV }}"
    # ... configurar para DEV
fi
```

**Por qué es necesario:**
- Los PRs siempre deben deployarse a DEV (nunca a PROD)
- Validación adicional de seguridad para evitar deploys accidentales
- Consistencia con el comportamiento esperado

**Impacto:**
- ✅ PRs siempre deployan a DEV
- ✅ Validación de seguridad adicional
- ✅ No afecta otros triggers

---

### 4. **Actualizado checkout para PRs**

**Ubicación:** `.github/workflows/deploy-cloud-run.yml` (línea 41)

**Cambio:**
```yaml
ref: ${{ github.event.inputs.branch || github.event.pull_request.head.ref || github.ref_name }}
```

**Por qué es necesario:**
- En eventos `pull_request`, `github.ref_name` apunta a la rama base (ej: `develop`)
- Necesitamos hacer checkout de la rama del PR (`head.ref`)
- Mantiene compatibilidad con `workflow_dispatch` y `push`

**Impacto:**
- ✅ Checkout correcto de la rama del PR
- ✅ Compatible con todos los triggers

---

### 5. **Agregado input opcional `schema_override` en workflow_dispatch**

**Ubicación:** `.github/workflows/deploy-cloud-run.yml` (líneas 20-23)

**Cambio:**
```yaml
schema_override:
  description: Schema override (opcional, ej: custom_schema)
  required: false
  type: string
```

**Por qué es necesario:**
- Permite override manual del schema cuando se necesita un comportamiento específico
- Útil para testing o casos especiales
- No afecta el comportamiento por defecto

**Impacto:**
- ✅ Flexibilidad para casos especiales
- ✅ No afecta el comportamiento normal

---

### 6. **Actualizado environment protection para PRs**

**Ubicación:** `.github/workflows/deploy-cloud-run.yml` (línea 35)

**Cambio:**
```yaml
environment: ${{ github.event_name == 'workflow_dispatch' && vars.DEPLOY_ENV_DEV || github.event_name == 'pull_request' && vars.DEPLOY_ENV_DEV || ... }}
```

**Por qué es necesario:**
- Los PRs deben usar el environment DEV (sin protección)
- Mantiene la protección para `main` (PROD)

**Impacto:**
- ✅ PRs no requieren aprobación (deploy automático a DEV)
- ✅ `main` mantiene protección si está configurada

---

### 7. **Creado workflow opcional de garbage collector**

**Ubicación:** `.github/workflows/garbage-collect-schemas.yml` (nuevo archivo)

**Características:**
- **Trigger:** Cron semanal (domingos 2 AM UTC) + `workflow_dispatch` manual
- **Función:** Limpia schemas `branch_*` que tienen más de N días (default: 7)
- **Seguridad:** Nunca borra `public`, `pr_*`, ni schemas reservados
- **Lógica:** Usa la fecha de `schema_migrations.installed_on` para determinar antigüedad

**Por qué es necesario:**
- Los schemas `branch_*` no se limpian automáticamente cuando se mergea una rama
- Evita acumulación de schemas huérfanos en la base de datos
- Mantiene la DB limpia sin intervención manual

**Impacto:**
- ✅ Limpieza automática de schemas antiguos
- ✅ Configurable (días de antigüedad)
- ✅ Ejecutable manualmente si es necesario

---

## 📊 Matriz de Comportamiento Actualizado

| Escenario | Trigger | DB_SCHEMA | Deploy a | ¿Altera DB dev? | Cleanup automático |
|-----------|---------|-----------|----------|-----------------|-------------------|
| **PR abierto/sincronizado** | `pull_request` | `pr_<number>` | DEV | ❌ No | ✅ Sí (al cerrar PR) |
| **Deploy manual rama x** | `workflow_dispatch` | `branch_<slug>` | DEV | ❌ No | ⚠️ Manual o cron |
| **Push a develop** | `push` | `public` | DEV | ✅ Sí | ❌ No |
| **Push a staging** | `push` | `public` | DEV | ✅ Sí | ❌ No |
| **Push a main** | `push` | `public` | PROD | ❌ No | ❌ No |
| **Cerrar PR** | `pull_request: closed` | N/A | N/A | N/A | ✅ Sí (`pr_<number>`) |

---

## 🔄 Flujos de Trabajo Actualizados

### Flujo 1: PR Automático (NUEVO)
```
1. Crear PR: feature/nueva-funcionalidad → develop
2. Push a la rama del PR
3. Workflow se ejecuta automáticamente (trigger: pull_request)
4. Resultado:
   - Schema creado: pr_<number>
   - Deploy a DEV
   - DB dev NO alterada
5. Probar en el schema aislado
6. Mergear PR → cleanup-schema.yml borra pr_<number>
```

### Flujo 2: Deploy Manual de Feature Branch (ACTUALIZADO)
```
1. Crear rama: feature/nueva-funcionalidad
2. Push a GitHub
3. En GitHub Actions → Deploy to Cloud Run → Run workflow
   - Seleccionar rama: feature/nueva-funcionalidad
4. Resultado:
   - Schema creado: branch_feature-nueva-funcionalidad (sin SHA)
   - Deploy a DEV
   - DB dev NO alterada
5. Múltiples deploys de la misma rama → reutilizan el mismo schema
6. (Opcional) Garbage collector limpia schemas antiguos (>7 días)
```

### Flujo 3: Merge a Develop (SIN CAMBIOS)
```
1. Mergear feature/nueva-funcionalidad → develop
2. Push a develop dispara workflow automáticamente
3. Resultado:
   - Schema usado: public
   - Deploy a DEV
   - DB dev SÍ alterada (migraciones en public)
```

### Flujo 4: PR con Cleanup (AHORA FUNCIONAL)
```
1. Crear PR: feature/nueva-funcionalidad → develop
2. Workflow deploya automáticamente → crea pr_<number>
3. Mergear o cerrar PR
4. Workflow cleanup-schema.yml se ejecuta
5. Resultado:
   - ✅ Borra pr_<number> correctamente
   - ✅ Schema limpiado automáticamente
```

---

## 🔍 Comparación Antes vs Después

### Antes
- ❌ PRs no se deployaban automáticamente
- ❌ Schemas `branch_<slug>_<sha>` se creaban con SHA (múltiples schemas por rama)
- ❌ Cleanup solo funcionaba para `pr_<number>` (que nunca se creaba)
- ❌ Schemas `branch_*` nunca se limpiaban automáticamente

### Después
- ✅ PRs se deployan automáticamente (opened, synchronize, reopened)
- ✅ Schemas `branch_<slug>` son estables (sin SHA, reutilizables)
- ✅ Cleanup funciona correctamente para `pr_<number>` (ahora se crean)
- ✅ Garbage collector limpia schemas `branch_*` antiguos automáticamente

---

## ⚠️ Notas Importantes

1. **Backward Compatibility:**
   - ✅ `develop`/`main`/`staging` siguen usando `public`
   - ✅ `workflow_dispatch` sigue funcionando igual (solo cambió el formato del schema)
   - ✅ No se requieren cambios en el código de la aplicación

2. **Schemas Existentes:**
   - Los schemas `branch_<slug>_<sha>` existentes seguirán existiendo
   - El garbage collector los limpiará si tienen más de 7 días
   - Los nuevos deploys usarán `branch_<slug>` (sin SHA)

3. **Garbage Collector:**
   - Es opcional y puede deshabilitarse si no se necesita
   - Se ejecuta semanalmente (domingos 2 AM UTC)
   - Puede ejecutarse manualmente con `workflow_dispatch`

4. **Schema Override:**
   - Solo disponible en `workflow_dispatch`
   - Útil para casos especiales de testing
   - No afecta el comportamiento normal

---

## 🧪 Testing Recomendado

1. **Crear un PR** y verificar que se deploya automáticamente con `pr_<number>`
2. **Hacer deploy manual** de una feature branch y verificar que usa `branch_<slug>` (sin SHA)
3. **Cerrar un PR** y verificar que el cleanup borra `pr_<number>`
4. **Ejecutar garbage collector manualmente** y verificar que limpia schemas antiguos
5. **Verificar que `develop` y `main` siguen usando `public`**

---

## 📝 Archivos Modificados

1. `.github/workflows/deploy-cloud-run.yml` - Workflow principal (modificado)
2. `.github/workflows/garbage-collect-schemas.yml` - Garbage collector (nuevo)

---

## ✅ Checklist de Implementación

- [x] Agregado trigger `pull_request` con eventos opened, synchronize, reopened
- [x] Modificada lógica de DB_SCHEMA para PRs → `pr_<number>`
- [x] Eliminado SHA de schemas `branch_*` en `workflow_dispatch`
- [x] Actualizada lógica de deploy variables para PRs
- [x] Agregado input opcional `schema_override`
- [x] Actualizado checkout para PRs
- [x] Actualizado environment protection para PRs
- [x] Creado workflow de garbage collector
- [x] Mantenida compatibilidad backward
- [x] Documentación completa

---

**Todos los cambios están implementados y listos para usar.** 🚀
