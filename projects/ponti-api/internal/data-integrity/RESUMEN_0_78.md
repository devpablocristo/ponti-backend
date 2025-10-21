# 🔍 RESUMEN EJECUTIVO: Diferencia de $0.78

## ⚡ Causa Raíz (1 línea)
**`total_used` (almacenado) ≠ `final_dose × effective_area` (calculado)** en 21 items de workorder_items del proyecto 11.

---

## 📊 Verificación Numérica

```
LEFT (RAW - GetRawDirectCost):  $24,604.39
RIGHT (SSOT - v3_workorder_metrics):  $24,605.17
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
DIFERENCIA:                     -$0.78
```

---

## 🎯 El Problema en Código

### LEFT (Controles 1-4):
```go
// repository.go línea 287-315
SELECT SUM(
  (wo.effective_area * l.price) +      // Labor
  (wi.total_used * s.price)            // ← USA TOTAL_USED (almacenado)
)
```

### RIGHT (Dashboard, Lotes, Informes):
```sql
-- v3_lot_ssot.supply_cost_for_lot_base() línea 71-86
SELECT SUM(
  v3_core_ssot.supply_cost(
    wi.final_dose,                     // ← USA FINAL_DOSE
    s.price,
    w.effective_area                   // ← MULTIPLICADO POR AREA
  )
)
-- Cálculo: final_dose × price × effective_area
```

---

## 🔬 Ejemplo Real (Mayor Impacto)

**Workorder #49, Item #122** (METSULFURON):

```
total_used almacenado:         0.300000 unidades
final_dose × effective_area:   0.003000 × 115 = 0.345000 unidades
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Diferencia:                    -0.045000 unidades
× Precio ($29.15):             = -$1.31
```

---

## 📋 Top 5 Items con Diferencias

| WO | Insumo | total_used | calculated | Impacto |
|----|--------|------------|------------|---------|
| 49 | METSULFURON | 0.300 | 0.345 | **-$1.31** 👈 |
| 54 | METSULFURON | 0.200 | 0.190 | +$0.29 |
| 49 | 2-4D ETILHEXILICO | 15.560 | 15.525 | +$0.17 |
| 41 | METSULFURON | 0.050 | 0.045 | +$0.15 |
| 48 | CURASEMILLAS | 19.000 | 18.968 | +$0.12 |
| ... | ... | ... | ... | ... |
| **Σ 21 items** | | | | **-$0.78** |

---

## 🤔 ¿Por Qué Existe Esta Diferencia?

### Posibles Causas:
1. ⭐ **Frontend redondea antes de enviar**:
   - JS: `0.003 × 115 = 0.345`
   - Frontend: `Math.round(0.345, 1) = 0.3`
   - DB almacena: `0.300000`

2. 👤 **Usuarios ingresan manualmente** valores ajustados

3. 🔢 **Precisión JavaScript vs PostgreSQL** (floating point vs NUMERIC)

4. 🔄 **Cambios posteriores** en `effective_area` desincronizaron `total_used`

---

## ✅ ¿Cuál Valor es "Correcto"?

### Depende de la Semántica:

| Escenario | Correcto | Acción |
|-----------|----------|--------|
| **`total_used` DEBE ser `final_dose × area`** | ❌ Bug de datos | Corregir datos o validación |
| **`total_used` PUEDE diferir** (ajustes manuales) | ✅ Ambos OK | Documentar diferencia |

---

## 💡 Solución Recomendada

### ⚡ Solución Inmediata (Sin código):
```
Ajustar tolerancia de Controles 1-4 de 0 a ±1 USD
(igual que controles 9, 11-13)
```

### 🔧 Mejora a Mediano Plazo:
1. Investigar con usuarios la semántica de `total_used`
2. Validar en frontend: `|total_used - (final_dose × area)| < 0.01`
3. Alertar si diferencia > umbral

### 🏗️ Mejora Arquitectónica:
```sql
-- Opción: hacer total_used un campo calculado
ALTER TABLE workorder_items
  ADD CONSTRAINT total_used_check
  CHECK (ABS(total_used - (final_dose * 
    (SELECT effective_area FROM workorders WHERE id = workorder_id))) < 0.01);
```

O usar un trigger para auto-calcular.

---

## 📁 Archivos Relevantes

```
workorder/repository.go:283-316          ← GetRawDirectCost (LEFT)
migrations/000115_...lot_ssot.up.sql:71  ← supply_cost_for_lot_base (RIGHT)
migrations/000113_...core_ssot.up.sql:196 ← supply_cost() function
```

---

## 🎬 Conclusión

**NO es un bug de redondeo SQL.**  
**ES una diferencia arquitectónica** entre valor almacenado vs. calculado.

La diferencia de $0.78 está **completamente explicada** y proviene de **21 items** donde el frontend/usuario almacenó un `total_used` que no coincide exactamente con el cálculo matemático `final_dose × effective_area`.

---

**Análisis completo**: Ver `ANALISIS_DIFERENCIA_0_78.md`

