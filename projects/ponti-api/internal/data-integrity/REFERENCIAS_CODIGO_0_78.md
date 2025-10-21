# 📚 Referencias de Código: Diferencia $0.78

## 🎯 Archivos Involucrados

### 1️⃣ LEFT: Cálculo RAW (GetRawDirectCost)

**Archivo**: `projects/ponti-api/internal/workorder/repository.go`

**Líneas**: 280-316

```go
// GetRawDirectCost calcula el costo directo RAW desde las tablas workorders y workorder_items
// Calcula ∑(Órdenes_de_trabajo.costo_total) como indica el CSV de controles
// Este cálculo es INDEPENDIENTE de las vistas SSOT para validar coherencia
func (r *Repository) GetRawDirectCost(ctx context.Context, projectID int64) (decimal.Decimal, error) {
	// Query RAW: suma directa desde workorders + workorder_items
	// Labor cost: effective_area × labor.price
	// Supply cost: total_used × supply.price (valor almacenado realmente usado)
	q := `
		WITH workorder_costs AS (
		  SELECT 
		    wo.id,
		    -- Costo de la labor (área efectiva × precio de la labor)
		    (wo.effective_area * l.price) AS labor_cost,
		    -- Costo de insumos (suma de items: total_used × price)
		    COALESCE((
		      SELECT SUM(wi.total_used * s.price)    -- ← AQUÍ: total_used
		      FROM public.workorder_items wi
		      JOIN public.supplies s ON s.id = wi.supply_id
		      WHERE wi.workorder_id = wo.id 
		        AND wi.deleted_at IS NULL
		    ), 0) AS supply_cost
		  FROM public.workorders wo
		  JOIN public.labors l ON l.id = wo.labor_id
		  WHERE wo.deleted_at IS NULL
		    AND wo.project_id = ?
		)
		SELECT COALESCE(SUM(labor_cost + supply_cost), 0) AS total_cost
		FROM workorder_costs
	`
	// ...
}
```

**Clave**: Usa `wi.total_used * s.price` directamente.

---

### 2️⃣ RIGHT: Vista v3_workorder_metrics

**Archivo**: `projects/ponti-api/migrations/000117_create_v3_workorder_metrics.up.sql`

**Líneas**: 33-79

```sql
CREATE OR REPLACE VIEW public.v3_workorder_metrics AS
WITH lot_ids AS (
  SELECT DISTINCT
    w.project_id,
    w.field_id,
    w.lot_id
  FROM public.workorders w
  WHERE w.deleted_at IS NULL
)
SELECT
  li.project_id,
  li.field_id,
  li.lot_id,
  
  -- Superficie trabajada (suma de effective_area de workorders)
  v3_lot_ssot.surface_for_lot(li.lot_id) AS surface_ha,
  
  -- Consumos de insumos
  v3_lot_ssot.liters_for_lot(li.lot_id) AS liters,
  v3_lot_ssot.kilograms_for_lot(li.lot_id) AS kilograms,
  
  -- Costos (usa funciones consolidadas de v3_lot_ssot)
  v3_lot_ssot.labor_cost_for_lot(li.lot_id) AS labor_cost_usd,
  v3_lot_ssot.supply_cost_for_lot_base(li.lot_id) AS supplies_cost_usd,  -- ← AQUÍ
  (v3_lot_ssot.labor_cost_for_lot(li.lot_id) + 
   v3_lot_ssot.supply_cost_for_lot_base(li.lot_id)) AS direct_cost_usd,
  -- ...
FROM lot_ids li;
```

**Clave**: Usa `supply_cost_for_lot_base()` que calcula dinámicamente.

---

### 3️⃣ Función: v3_lot_ssot.supply_cost_for_lot_base

**Archivo**: `projects/ponti-api/migrations/000115_create_v3_lot_ssot_schema.up.sql`

**Líneas**: 70-86

```sql
-- 2.2: Costo de insumos base (solo workorder_items, sin movimientos internos)
CREATE OR REPLACE FUNCTION v3_lot_ssot.supply_cost_for_lot_base(p_lot_id bigint)
RETURNS numeric
LANGUAGE sql STABLE AS $$
  SELECT COALESCE(
    SUM(v3_core_ssot.supply_cost(             -- ← Llama a supply_cost()
      wi.final_dose::double precision,
      s.price::numeric,
      w.effective_area::numeric
    )), 0
  )::numeric
  FROM public.workorders w
  LEFT JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
  LEFT JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
  WHERE w.lot_id = p_lot_id
    AND w.deleted_at IS NULL
$$;
```

**Clave**: Pasa `final_dose`, `price`, y `effective_area` a `v3_core_ssot.supply_cost()`.

---

### 4️⃣ Función: v3_core_ssot.supply_cost

**Archivo**: `projects/ponti-api/migrations/000113_____create_v3_core_ssot_schema.up.sql`

**Líneas**: 195-200

```sql
-- Costo de insumo
CREATE OR REPLACE FUNCTION v3_core_ssot.supply_cost(
  final_dose double precision, 
  supply_price numeric, 
  effective_area numeric
) 
RETURNS numeric
LANGUAGE sql IMMUTABLE AS $$
  SELECT COALESCE(final_dose,0)::numeric * COALESCE(supply_price,0) * COALESCE(effective_area,0)
$$;
```

**Clave**: Calcula `final_dose × price × effective_area` (NO usa `total_used`).

---

### 5️⃣ Control 1: Órdenes → Dashboard

**Archivo**: `projects/ponti-api/internal/data-integrity/usecases.go`

**Líneas**: 192-234

```go
func (u *UseCases) control1_OrdenesVsDashboard(ctx context.Context, projectID *int64) (domain.IntegrityCheck, error) {
	pID := int64(0)
	if projectID != nil {
		pID = *projectID
	}

	// LEFT: Costos RAW desde workorders
	leftValue, err := u.workorderRepo.GetRawDirectCost(ctx, pID)  // ← Método RAW
	if err != nil {
		return domain.IntegrityCheck{}, err
	}

	// RIGHT: Costos desde dashboard (usa funciones SSOT)
	dashboardFilter := dashboardDomain.DashboardFilter{ProjectID: projectID}
	dashboardData, err := u.dashboardRepo.GetDashboard(ctx, dashboardFilter)
	if err != nil {
		return domain.IntegrityCheck{}, err
	}
	rightValue := dashboardData.ManagementBalance.Summary.DirectCostsExecutedUSD  // ← Método SSOT

	return buildCheck(
		1,
		"Órdenes de trabajo",
		"Costos directos ejecutados",
		"Dashboard",
		"Dashboard.CostosDirectos = ∑(Ordenes.costo_total)",
		"∑(workorders.effective_area × labors.price + workorder_items.total_used × supplies.price)",
		leftValue,
		"Tabla workorders RAW",
		"v3_dashboard_ssot.direct_costs_total_for_project()",
		rightValue,
		"Vista v3_dashboard_management_balance",
		decimal.Zero, // Tolerancia = 0 (debe ser exacto)
	), nil
}
```

**Clave**: Compara directamente los dos métodos.

---

### 6️⃣ Modelo: WorkorderItem

**Archivo**: `projects/ponti-api/internal/workorder/repository/models/workorder.go`

**Líneas**: 45-53

```go
// WorkorderItem GORM model
type WorkorderItem struct {
	ID          int64            `gorm:"primaryKey;autoIncrement"`
	WorkorderID int64            `gorm:"column:workorder_id;index"`
	SupplyID    int64            `gorm:"not null"`
	Supply      supplymod.Supply `gorm:"foreignKey:SupplyID"`
	TotalUsed   decimal.Decimal  `gorm:"not null"`  // ← Valor almacenado
	FinalDose   decimal.Decimal  `gorm:"not null"`  // ← Valor almacenado
}
```

**Clave**: Ambos campos (`TotalUsed` y `FinalDose`) son almacenados, no calculados.

---

### 7️⃣ Migración: Tabla workorder_items

**Archivo**: `projects/ponti-api/migrations/000020_create_workorder_items_table.up.sql`

**Líneas**: 1-10

```sql
CREATE TABLE workorder_items (
  id             BIGSERIAL PRIMARY KEY,
  workorder_id   BIGINT     NOT NULL REFERENCES workorders(id) ON UPDATE CASCADE ON DELETE CASCADE,
  supply_id      BIGINT     NOT NULL REFERENCES supplies(id) ON UPDATE CASCADE ON DELETE RESTRICT,
  total_used     NUMERIC(18,6) NOT NULL,  -- ← 6 decimales
  final_dose     NUMERIC(18,6) NOT NULL,  -- ← 6 decimales
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at     TIMESTAMPTZ
);
```

**Clave**: Ambos campos tienen la misma precisión (18,6).

---

### 8️⃣ DTO: WorkorderItem

**Archivo**: `projects/ponti-api/internal/workorder/handler/dto/workorder.go`

**Líneas**: 12-16

```go
type WorkorderItem struct {
	SupplyID  int64           `json:"supply_id" binding:"required"`
	TotalUsed decimal.Decimal `json:"total_used" binding:"required"`  // ← Viene del frontend
	FinalDose decimal.Decimal `json:"final_dose" binding:"required"`  // ← Viene del frontend
}
```

**Clave**: El frontend envía ambos valores, no se calculan en backend.

---

## 🔄 Flujo de Datos

```
┌─────────────────────────────────────────────────────────────────┐
│                      FRONTEND (JavaScript)                      │
│  - Usuario ingresa/calcula TotalUsed                            │
│  - Puede ser: manual, calculado, redondeado                     │
│  - Envía via JSON: { total_used, final_dose }                   │
└────────────────┬────────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    BACKEND (Go + GORM)                          │
│  - Recibe DTO WorkorderItem                                     │
│  - Almacena directamente en DB (NO calcula)                     │
│  - workorder_items.total_used = valor enviado                   │
└────────────────┬────────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    BASE DE DATOS (PostgreSQL)                   │
│                                                                  │
│  ┌───────────────────────────┬─────────────────────────────┐   │
│  │         RAW (LEFT)        │       SSOT (RIGHT)          │   │
│  ├───────────────────────────┼─────────────────────────────┤   │
│  │ GetRawDirectCost()        │ v3_workorder_metrics        │   │
│  │                            │                             │   │
│  │ SUM(                      │ SUM(                         │   │
│  │   total_used × price      │   final_dose × area × price │   │
│  │ )                          │ )                           │   │
│  │                            │                             │   │
│  │ = $24,604.39              │ = $24,605.17                │   │
│  └───────────────────────────┴─────────────────────────────┘   │
│                            │                                    │
│                            ▼                                    │
│                    DIFERENCIA = -$0.78                          │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔍 Comparación Directa

| Aspecto | LEFT (RAW) | RIGHT (SSOT) |
|---------|------------|--------------|
| **Función/Query** | `GetRawDirectCost()` | `v3_workorder_metrics.direct_cost_usd` |
| **Ubicación** | `workorder/repository.go:287` | Migración 000117 + 000115 + 000113 |
| **Cálculo Insumos** | `total_used × price` | `final_dose × effective_area × price` |
| **Campo Usado** | `workorder_items.total_used` | `workorder_items.final_dose` + `workorders.effective_area` |
| **Valor** | **$24,604.39** | **$24,605.17** |
| **Naturaleza** | Almacenado | Calculado |
| **Ventaja** | Refleja valor real guardado | Matemáticamente consistente |
| **Desventaja** | Puede desincronizarse | No refleja ajustes manuales |

---

## 🎯 Puntos Clave para Debuggear

1. **Frontend**: ¿Cómo se calcula `total_used` antes de enviarlo?
   - Ver código JavaScript de creación/edición de workorders
   - Buscar redondeos: `Math.round()`, `toFixed()`, etc.

2. **Backend**: ¿Hay validación de `total_used` vs. `final_dose × area`?
   - Ver handlers en `workorder/handler.go`
   - Buscar validaciones en use cases

3. **Base de Datos**: ¿Hay triggers o constraints?
   - Buscar en migraciones: `trigger`, `constraint`, `check`
   - Verificar si se actualiza automáticamente

4. **Auditoría**: ¿Cuándo se introdujeron las diferencias?
   - Revisar `workorder_items.created_at` de items con diferencia
   - Ver si hay un patrón temporal (ej: después de un deploy)

---

## 📝 Notas de Implementación

### Si se decide que `total_used` DEBE ser exacto:

**Opción 1: Validación en Frontend**
```javascript
// En el form de workorder
const calculatedTotalUsed = finalDose * effectiveArea;
if (Math.abs(totalUsed - calculatedTotalUsed) > 0.01) {
  alert('Error: total_used no coincide con el cálculo');
}
```

**Opción 2: Recalcular en Backend**
```go
// En CreateWorkorder o UpdateWorkorder
for i := range o.Items {
    o.Items[i].TotalUsed = o.Items[i].FinalDose.Mul(o.EffectiveArea)
}
```

**Opción 3: Trigger en PostgreSQL**
```sql
CREATE OR REPLACE FUNCTION sync_total_used()
RETURNS TRIGGER AS $$
BEGIN
  NEW.total_used := NEW.final_dose * (
    SELECT effective_area 
    FROM workorders 
    WHERE id = NEW.workorder_id
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER before_insert_or_update_workorder_items
BEFORE INSERT OR UPDATE ON workorder_items
FOR EACH ROW
EXECUTE FUNCTION sync_total_used();
```

### Si se decide que `total_used` PUEDE diferir:

**Opción 1: Ajustar Tolerancia**
```go
// En usecases.go, línea 232
decimal.NewFromInt(1), // Cambiar de 0 a 1 USD
```

**Opción 2: Documentar Diferencia**
- Agregar comentario en código explicando por qué puede diferir
- Actualizar README o documentación técnica
- Agregar test que valide que la diferencia está en rango esperado

---

## 🚀 Archivos a Revisar si se Modifica

Si decides cambiar la lógica de `total_used`:

1. ✅ `workorder/handler/dto/workorder.go` - DTO de entrada
2. ✅ `workorder/usecases/domain/workorder.go` - Dominio
3. ✅ `workorder/repository/models/workorder.go` - Modelo GORM
4. ✅ `workorder/repository.go` - Lógica de guardado
5. ✅ `data-integrity/usecases.go` - Controles (ajustar tolerancia)
6. ⚠️ Migraciones - Si agregas trigger/constraint
7. ⚠️ Frontend - Validación de inputs

---

**Última actualización**: 2025-10-21

