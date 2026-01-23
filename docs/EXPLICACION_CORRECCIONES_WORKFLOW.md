# Explicación Detallada: Situación y Correcciones del Workflow de Deploy

## 📋 Situación Original (ANTES de las correcciones)

### Problema 1: Deploy Automático de Pull Requests

**Configuración original:**
```yaml
on:
  pull_request:
    types: [opened, synchronize, reopened]
    branches: [develop, staging]
```

**Qué pasaba:**
- Cada vez que hacías **push a una rama con PR abierto** hacia `develop` o `staging`
- El workflow se ejecutaba **automáticamente** por el evento `synchronize`
- Esto causaba:
  - **Build completo** de la imagen Docker (5-10 minutos)
  - **Push** a Artifact Registry
  - **Deploy** a Cloud Run DEV
  - **Consumo de recursos** innecesario mientras trabajabas
  - **Molestia** porque cada push pequeño disparaba un deploy completo

**Ejemplo de tu caso:**
- Rama: `test/deploy-manual-dev`
- PR #5 abierto hacia `develop`
- Cada `git push` → Deploy automático completo
- Schema usado: `pr_5` (aislado, pero igual consumía recursos)

### Problema 2: Riesgo de Modificar DB de DEV

**Tu preocupación principal:**
- La base de datos de DEV contiene datos importantes
- **NO se puede perder** ni modificar accidentalmente
- Solo debe modificarse cuando se hace **merge a `develop`**

**Configuración original:**
- Los PRs usaban schema `pr_<number>` (aislado) ✅
- Pero el deploy automático era molesto ❌
- No había claridad sobre cuándo se modificaba la DB ❌

## ✅ Correcciones Implementadas

### Corrección 1: Eliminado Deploy Automático de PRs

**Cambio realizado:**
```yaml
# ANTES:
on:
  pull_request:
    types: [opened, synchronize, reopened]
    branches: [develop, staging]

# DESPUÉS:
on:
  push:
    branches: [develop, main]  # Solo estas ramas
  workflow_dispatch:            # Deploy manual
```

**Resultado:**
- ✅ Los PRs **NO** disparan deploys automáticos
- ✅ Puedes trabajar en tu rama sin que se ejecute nada
- ✅ Solo se deploya cuando haces **merge a `develop` o `main`**
- ✅ Deploy manual sigue disponible cuando lo necesites

### Corrección 2: Protección Explícita de la DB de DEV

**Cambio realizado en la lógica de `DB_SCHEMA`:**

```bash
# ⚠️ SEGURIDAD CRÍTICA: La DB de DEV solo se modifica cuando se hace merge a develop
# Los deploys manuales SIEMPRE usan schemas aislados (branch_<slug>)

# ✅ CRÍTICO: workflow_dispatch se verifica ANTES de ref_name
# Esto es importante porque workflow_dispatch puede dispararse desde main/develop
# pero DEBE usar schema aislado para proteger la DB
if [ "$EVENT_NAME" = "workflow_dispatch" ]; then
  DB_SCHEMA="branch_${BRANCH_SLUG}"
  echo "✅ Using isolated schema: ${DB_SCHEMA} - DB dev is SAFE"
# ⚠️ SOLO push a develop/main usa schema public (modifica la DB)
elif [ "${{ github.ref_name }}" = "develop" ] || [ "${{ github.ref_name }}" = "main" ]; then
  DB_SCHEMA="public"
  echo "⚠️  WARNING: Using public schema - this will modify the database!"
else
  # Fallback para otros casos
  DB_SCHEMA="branch_${BRANCH_SLUG}"
  echo "✅ Using isolated schema: ${DB_SCHEMA} - DB dev is SAFE"
fi
```

**Resultado:**
- ✅ **Solo** cuando haces merge a `develop` → usa `public` → modifica DB de dev
- ✅ **Solo** cuando haces merge a `main` → usa `public` → modifica DB de prod
- ✅ **Deploy manual** → usa `branch_<slug>` → **NO modifica** DB de dev
- ✅ Comentarios claros en el código explicando la seguridad

### Corrección 3: Concurrency para Evitar Runs Solapados

**Cambio realizado:**
```yaml
concurrency:
  group: deploy-${{ github.event_name }}-${{ github.event.inputs.branch || github.ref_name }}
  cancel-in-progress: true
```

**Resultado:**
- ✅ Si llega un nuevo deploy del mismo tipo, cancela el anterior
- ✅ Evita deploys simultáneos que podrían causar conflictos
- ✅ Ahorra recursos al cancelar builds innecesarios

## 🔒 Garantías de Seguridad Actuales

### La DB de DEV SOLO se modifica cuando:

1. **Merge a `develop`** (push directo o merge de PR)
   - Trigger: `push` a `develop`
   - Schema: `public`
   - Acción: Modifica la DB de dev
   - ⚠️ **Este es el único caso donde se modifica la DB de dev**

2. **Merge a `main`** (push directo o merge de PR)
   - Trigger: `push` a `main`
   - Schema: `public`
   - Acción: Modifica la DB de prod
   - ⚠️ **Este es el único caso donde se modifica la DB de prod**

### La DB de DEV NO se modifica cuando:

1. **Deploy manual** (`workflow_dispatch`)
   - Trigger: Manual desde GitHub Actions
   - Schema: `branch_<slug>` (ej: `branch_test-deploy-manual-dev`)
   - Acción: **NO modifica** DB de dev (schema aislado)
   - ✅ **Seguro**

2. **Push a cualquier otra rama**
   - Trigger: Ninguno (no se ejecuta automáticamente)
   - Acción: No pasa nada
   - ✅ **Seguro**

3. **Pull Request abierto/sincronizado**
   - Trigger: Ninguno (eliminado)
   - Acción: No pasa nada
   - ✅ **Seguro**

## 📊 Comparativa: Antes vs Después

| Escenario | ANTES | DESPUÉS |
|-----------|-------|---------|
| Push a rama con PR abierto | ❌ Deploy automático (molesto) | ✅ No hace nada |
| Merge a `develop` | ✅ Deploy automático (correcto) | ✅ Deploy automático (correcto) |
| Merge a `main` | ✅ Deploy automático (correcto) | ✅ Deploy automático (correcto) |
| Deploy manual | ✅ Disponible | ✅ Disponible |
| Schema usado en deploy manual | `branch_<slug>` (aislado) | `branch_<slug>` (aislado) |
| Schema usado en merge a develop | `public` (modifica DB) | `public` (modifica DB) |
| Protección de DB de dev | ⚠️ Implícita | ✅ Explícita con comentarios |

## 🎯 Comportamiento Final

### Flujo Normal de Trabajo:

1. **Crear rama y trabajar:**
   ```bash
   git checkout -b feature/mi-feature
   # Trabajas, haces commits, pushes...
   # ✅ NO se ejecuta ningún deploy automático
   ```

2. **Abrir PR hacia `develop`:**
   - PR se abre
   - ✅ NO se ejecuta ningún deploy automático
   - Puedes seguir trabajando sin molestias

3. **Hacer deploy manual si necesitas probar:**
   - GitHub Actions → Deploy to Cloud Run → Run workflow
   - Seleccionas tu rama: `feature/mi-feature`
   - ✅ Se deploya con schema `branch_feature-mi-feature`
   - ✅ **NO modifica** la DB de dev (schema aislado)

4. **Mergear PR a `develop`:**
   - Haces merge del PR
   - Esto dispara un `push` a `develop`
   - ✅ Se ejecuta deploy automático
   - ✅ Usa schema `public`
   - ⚠️ **Modifica** la DB de dev (esto es correcto, es un merge)

## 🔍 Verificación de Seguridad

### Cómo verificar que la DB está protegida:

1. **Revisar triggers del workflow:**
   ```yaml
   on:
     push:
       branches: [develop, main]  # Solo estas ramas
   ```
   ✅ Solo se ejecuta en merge a `develop` o `main`

2. **Revisar lógica de schema:**
   ```bash
   if [ "${{ github.ref_name }}" = "develop" ] || [ "${{ github.ref_name }}" = "main" ]; then
     DB_SCHEMA="public"  # Solo aquí se modifica la DB
   else
     DB_SCHEMA="branch_${BRANCH_SLUG}"  # Schema aislado, seguro
   fi
   ```
   ✅ Solo `develop` y `main` usan `public`

3. **Logs del workflow:**
   - Cuando usa `public`: `⚠️  WARNING: Using public schema - this will modify the database!`
   - Cuando usa schema aislado: `✅ Using isolated schema: branch_xxx - DB dev is SAFE`

## ✅ Resumen Final

**Problema resuelto:**
- ❌ Antes: Cada push a PR disparaba deploy automático (molesto)
- ✅ Ahora: Solo merge a `develop`/`main` dispara deploy automático

**Seguridad garantizada:**
- ✅ La DB de DEV solo se modifica cuando se hace merge a `develop`
- ✅ Los deploys manuales usan schemas aislados (no tocan la DB de dev)
- ✅ Comentarios explícitos en el código documentan la seguridad
- ✅ Logs claros indican cuándo se modifica la DB

**Tu preocupación resuelta:**
- ✅ **La base de datos de DEV está protegida**
- ✅ **Solo se modifica cuando haces merge a `develop`**
- ✅ **No se puede perder ni modificar accidentalmente**
