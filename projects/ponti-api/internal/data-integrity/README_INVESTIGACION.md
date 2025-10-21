# 🔍 Investigación: Diferencia de $0.78 en Controles 1-4

## 📂 Documentación Generada

Esta carpeta contiene la investigación completa sobre el origen de la diferencia de **$0.78** en los Controles de Data-Integrity 1-4 (Proyecto 11).

---

## 📄 Documentos Disponibles

### 1. 📊 **RESUMEN_0_78.md**
**Para**: Lectura rápida (5 minutos)

Resumen ejecutivo con:
- ⚡ Causa raíz en 1 línea
- 📊 Verificación numérica
- 🎯 Código del problema
- 🔬 Ejemplo real
- 💡 Soluciones recomendadas

👉 **Empieza aquí si quieres entender rápido el problema**

---

### 2. 📖 **ANALISIS_DIFERENCIA_0_78.md**
**Para**: Análisis completo y detallado

Incluye:
- 🔍 Resumen ejecutivo
- 📊 Valores observados
- 🔬 Análisis detallado de métodos LEFT/RIGHT
- 🎯 Causa raíz identificada
- 📋 Top 5 items con mayor impacto
- 🤔 Posibles causas (redondeo, manual, precisión)
- 📋 Comparación de arquitectura
- 🎯 ¿Cuál es el correcto?
- 💡 Recomendaciones (corto/mediano/largo plazo)
- 📄 Query de verificación
- ✅ Conclusión

👉 **Lee esto para entender todo el contexto y decisiones**

---

### 3. 🗂️ **QUERIES_INVESTIGACION_0_78.sql**
**Para**: Investigación técnica adicional

Contiene 11 queries SQL:

1. ✅ Verificación básica de diferencias
2. 📋 Items con diferencias (detallado)
3. 📊 Resumen por tipo de insumo
4. 📊 Resumen por workorder
5. ⚖️ Comparación RAW vs SSOT
6. 📈 Estadísticas generales
7. ✅ Items perfectamente sincronizados
8. 🔍 Análisis de redondeo
9. 📅 Auditoría temporal
10. 🔧 Corrección propuesta (NO ejecutar)
11. 📤 Exportar para Excel

👉 **Usa esto para investigar más profundo en la DB**

---

### 4. 📚 **REFERENCIAS_CODIGO_0_78.md**
**Para**: Desarrollo y modificaciones

Incluye:
- 🎯 8 archivos involucrados con líneas exactas
- 🔄 Flujo de datos completo (Frontend → Backend → DB)
- 🔍 Comparación directa LEFT vs RIGHT
- 🎯 Puntos clave para debuggear
- 📝 Notas de implementación (si se decide modificar)
- 🚀 Archivos a revisar si se modifica

👉 **Lee esto antes de modificar código relacionado**

---

## 🎯 Resumen Ultra-Rápido (30 segundos)

```
PROBLEMA:
  LEFT (workorders RAW):     $24,604.39
  RIGHT (dashboard/lotes):    $24,605.17
  DIFERENCIA:                      -$0.78

CAUSA:
  total_used (almacenado) ≠ final_dose × effective_area (calculado)
  en 21 items de workorder_items

SOLUCIÓN RÁPIDA:
  Ajustar tolerancia de Controles 1-4 de 0 a ±1 USD
```

---

## 🚀 Cómo Usar Esta Documentación

### Si eres **Product Owner / Manager**:
1. Lee `RESUMEN_0_78.md` (5 min)
2. Decide: ¿`total_used` debe ser exacto o puede diferir?
3. Comunica decisión al equipo técnico

### Si eres **Developer**:
1. Lee `RESUMEN_0_78.md` (5 min)
2. Lee `ANALISIS_DIFERENCIA_0_78.md` (15 min)
3. Consulta `REFERENCIAS_CODIGO_0_78.md` si vas a modificar
4. Usa `QUERIES_INVESTIGACION_0_78.sql` para verificar en DB

### Si eres **QA / Tester**:
1. Lee `RESUMEN_0_78.md` (5 min)
2. Usa queries de `QUERIES_INVESTIGACION_0_78.sql` para verificar
3. Específicamente query #11 para exportar a Excel

### Si eres **Data Analyst**:
1. Usa `QUERIES_INVESTIGACION_0_78.sql` completo
2. Query #2: Detalle de items con diferencias
3. Query #3: Resumen por categoría
4. Query #8: Análisis de patrones de redondeo

---

## 📊 Datos Clave

| Métrica | Valor |
|---------|-------|
| **Diferencia total** | -$0.78 |
| **Items afectados** | 21 items |
| **Mayor diferencia individual** | -$1.31 (Workorder 49, METSULFURON) |
| **Rango de diferencias** | -$1.31 a +$0.29 |
| **Total items en proyecto 11** | XXX items |
| **% items con diferencia** | ~X% |

---

## 🎯 Decisión Requerida

El equipo debe decidir:

### Opción A: `total_used` DEBE ser exacto
- ✅ Más consistencia matemática
- ❌ Requiere cambios en frontend/backend
- 🔧 Acción: Implementar validación o trigger
- ⏱️ Tiempo estimado: 2-3 días

### Opción B: `total_used` PUEDE diferir
- ✅ No requiere cambios
- ✅ Permite ajustes manuales de usuarios
- ❌ Controles 1-4 siempre tendrán diferencia
- 🔧 Acción: Ajustar tolerancia a ±1 USD
- ⏱️ Tiempo estimado: 5 minutos

---

## 🔗 Enlaces Rápidos

- [Controles de Data-Integrity](./usecases.go)
- [Repository Workorders](../workorder/repository.go)
- [Migraciones SSOT](../../migrations/)

---

## ✍️ Autor y Fecha

- **Investigado por**: AI Assistant (Claude Sonnet 4.5)
- **Fecha**: 2025-10-21
- **Proyecto**: Ponti Backend
- **Módulo**: Data-Integrity
- **Versión**: 1.0

---

## 📝 Notas Adicionales

- ⚠️ Esta investigación se basa en el estado del código al 2025-10-21
- ⚠️ Los queries SQL están probados en proyecto 11
- ⚠️ Antes de ejecutar queries de modificación (UPDATE), hacer backup
- ✅ Todos los queries de SELECT son seguros para ejecutar

---

## 🔄 Historial de Cambios

| Fecha | Versión | Cambios |
|-------|---------|---------|
| 2025-10-21 | 1.0 | Investigación inicial completa |

---

**¿Preguntas?** Contacta al equipo de desarrollo.

