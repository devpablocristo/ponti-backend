Fix Labores Lista Labores:

el vista 74 corrige el problema de duplicados de labores:

## �� **¿QUÉ CORRIGE Y CÓMO LO HACE?**

### ** PROBLEMA ORIGINAL:**

#### **Vista `fix_labors_list` con duplicación:**
```sql
-- PROBLEMA: LEFT JOIN sin filtro de mes causa duplicación
LEFT JOIN project_dollar_values pdv ON pdv.project_id = w.project_id 
    AND pdv.deleted_at IS NULL
```

**❌ Resultado:** Si hay 3 meses de dólar promedio para el mismo proyecto, cada labor aparece 3 veces.

---

### **✅ SOLUCIÓN IMPLEMENTADA:**

#### **1. Función Helper:**
```sql
CREATE OR REPLACE FUNCTION get_project_dollar_value(p_project_id BIGINT, p_month VARCHAR)
RETURNS DECIMAL AS $$
BEGIN
  RETURN (
    SELECT average_value 
    FROM project_dollar_values 
    WHERE project_id = p_project_id 
      AND month = p_month 
      AND deleted_at IS NULL
    LIMIT 1
  );
END;
$$ LANGUAGE plpgsql IMMUTABLE;
```

** Cómo funciona:**
- **Entrada:** `project_id` y `month` específicos
- **Salida:** Un solo valor de dólar para ese proyecto y mes
- **Ventaja:** Evita duplicación porque retorna solo 1 valor

#### **2. Vista Corregida:**
```sql
-- ANTES (con duplicación):
LEFT JOIN project_dollar_values pdv ON pdv.project_id = w.project_id 
    AND pdv.deleted_at IS NULL
COALESCE(pdv.average_value, get_default_fx_rate()) AS usd_avg_value

-- DESPUÉS (sin duplicación):
-- FIX: Usar función que obtiene el valor específico del mes
COALESCE(
    get_project_dollar_value(w.project_id, '01'), -- Mes específico
    get_default_fx_rate()
) AS usd_avg_value
```

---

### ** COMPARACIÓN DETALLADA:**

#### **❌ ANTES (con duplicación):**
```sql
-- Si hay 3 meses de dólar promedio:
-- Mes 01: $1000
-- Mes 02: $1100  
-- Mes 03: $1200

-- LEFT JOIN sin filtro de mes:
-- Labor 1 → 3 registros (uno por cada mes)
-- Labor 2 → 3 registros (uno por cada mes)
-- Labor 3 → 3 registros (uno por cada mes)

-- Resultado: 9 registros en lugar de 3
```

#### **✅ DESPUÉS (sin duplicación):**
```sql
-- Función con mes específico:
get_project_dollar_value(project_id, '01') → $1000 (solo 1 valor)

-- Labor 1 → 1 registro
-- Labor 2 → 1 registro  
-- Labor 3 → 1 registro

-- Resultado: 3 registros (correcto)
```

---

### ** MECANISMO DE CORRECCIÓN:**

#### **1. Eliminación del JOIN problemático:**
```sql
-- ELIMINADO:
LEFT JOIN project_dollar_values pdv ON pdv.project_id = w.project_id 
    AND pdv.deleted_at IS NULL
```

#### **2. Reemplazo por función parametrizada:**
```sql
-- IMPLEMENTADO:
COALESCE(
    get_project_dollar_value(w.project_id, '01'), -- Mes específico
    get_default_fx_rate()
) AS usd_avg_value
```

#### **3. Cálculos corregidos:**
```sql
-- Todos los cálculos usan la función:
-- Costo U$/Ha en pesos:
(lb.price * COALESCE(
    get_project_dollar_value(w.project_id, '01'), -- Sin duplicación
    get_default_fx_rate()
)) AS usd_cost_ha

-- Total U Neto en pesos:
(lb.price * COALESCE(
    get_project_dollar_value(w.project_id, '01'), -- Sin duplicación
    get_default_fx_rate()
) * w.effective_area) AS usd_net_total
```

---

### ** VENTAJAS DE LA SOLUCIÓN:**

#### **1. Eliminación de duplicación:**
- **Antes:** N labores × M meses = N×M registros
- **Después:** N labores × 1 mes = N registros

#### **2. Flexibilidad:**
- **Mes configurable:** Puedes cambiar '01' por cualquier mes
- **Fallback:** Si no hay datos, usa `get_default_fx_rate()`

#### **3. Performance:**
- **Función optimizada:** `IMMUTABLE` para mejor performance
- **Sin JOINs innecesarios:** Reduce complejidad de la consulta

#### **4. Mantenibilidad:**
- **Función reutilizable:** Se puede usar en otras vistas
- **Lógica centralizada:** Un solo lugar para obtener valores de dólar

---

### ** RESUMEN:**

**La corrección elimina la duplicación de labores** que ocurría cuando había múltiples meses de dólar promedio para el mismo proyecto, **reemplazando el LEFT JOIN problemático por una función parametrizada** que obtiene un solo valor específico por proyecto y mes, **garantizando que cada labor aparezca una sola vez** en la vista `fix_labors_list`.