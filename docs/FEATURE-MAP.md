TLDR: Mapa real de features y queries allowlist basado en `migrations_v4`.

# Feature Map (migrations_v4)

Este documento lista features reales del dominio con SQL allowlist.
Todas las queries incluyen:
- `project_id` obligatorio
- `LIMIT` obligatorio

## Convenciones
- `entity`: project | campaign | lot
- `window`: all | last_30d | last_7d
- Parametros estandar: `project_id`, `limit`, `date_from`, `date_to`
- SQL usa objetos reales de `migrations_v4`

---

## Costos

### feature_cost_total
- entity: project
- window: all
- source: `v4_report.dashboard_management_balance`
- query_id: `feature_cost_total`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  costos_directos_ejecutados_usd AS cost_total_usd
FROM v4_report.dashboard_management_balance
WHERE project_id = %(project_id)s
LIMIT %(limit)s;
```

### feature_cost_per_ha
- entity: project
- window: all
- source: `v4_report.lot_metrics`
- query_id: `feature_cost_per_ha`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  COALESCE(SUM(direct_cost_total_usd), 0) / NULLIF(SUM(hectares), 0) AS cost_per_ha_usd
FROM v4_report.lot_metrics
WHERE project_id = %(project_id)s
GROUP BY project_id
LIMIT %(limit)s;
```

### feature_cost_total_last_30d
- entity: project
- window: last_30d
- source: `public.workorders`, `public.labors`
- query_id: `feature_cost_total_last_30d`
- params: `project_id`, `limit`
```sql
SELECT
  w.project_id,
  COALESCE(SUM(lb.price * w.effective_area), 0) AS cost_total_usd
FROM public.workorders w
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
WHERE w.project_id = %(project_id)s
  AND w.deleted_at IS NULL
  AND w.effective_area > 0
  AND w.date >= (CURRENT_DATE - INTERVAL '30 days')
GROUP BY w.project_id
LIMIT %(limit)s;
```

### feature_cost_total_last_7d
- entity: project
- window: last_7d
- source: `public.workorders`, `public.labors`
- query_id: `feature_cost_total_last_7d`
- params: `project_id`, `limit`
```sql
SELECT
  w.project_id,
  COALESCE(SUM(lb.price * w.effective_area), 0) AS cost_total_usd
FROM public.workorders w
JOIN public.labors lb ON lb.id = w.labor_id AND lb.deleted_at IS NULL
WHERE w.project_id = %(project_id)s
  AND w.deleted_at IS NULL
  AND w.effective_area > 0
  AND w.date >= (CURRENT_DATE - INTERVAL '7 days')
GROUP BY w.project_id
LIMIT %(limit)s;
```

---

## Insumos

### feature_inputs_by_category
- entity: project
- window: all
- source: `v4_report.field_crop_insumos`
- query_id: `feature_inputs_by_category`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  COALESCE(SUM(semillas_total_usd), 0) AS semillas_total_usd,
  COALESCE(SUM(curasemillas_total_usd), 0) AS curasemillas_total_usd,
  COALESCE(SUM(herbicidas_total_usd), 0) AS herbicidas_total_usd,
  COALESCE(SUM(insecticidas_total_usd), 0) AS insecticidas_total_usd,
  COALESCE(SUM(fungicidas_total_usd), 0) AS fungicidas_total_usd,
  COALESCE(SUM(coadyuvantes_total_usd), 0) AS coadyuvantes_total_usd,
  COALESCE(SUM(fertilizantes_total_usd), 0) AS fertilizantes_total_usd,
  COALESCE(SUM(otros_insumos_total_usd), 0) AS otros_insumos_total_usd,
  COALESCE(SUM(total_insumos_usd), 0) AS total_insumos_usd
FROM v4_report.field_crop_insumos
WHERE project_id = %(project_id)s
GROUP BY project_id
LIMIT %(limit)s;
```

### feature_inputs_last_30d
- entity: project
- window: last_30d
- source: `public.workorders`, `public.workorder_items`
- query_id: `feature_inputs_last_30d`
- params: `project_id`, `limit`
```sql
SELECT
  w.project_id,
  COALESCE(SUM(wi.total_used), 0) AS inputs_total_used
FROM public.workorders w
JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
WHERE w.project_id = %(project_id)s
  AND w.deleted_at IS NULL
  AND w.date >= (CURRENT_DATE - INTERVAL '30 days')
GROUP BY w.project_id
LIMIT %(limit)s;
```

### feature_inputs_last_7d
- entity: project
- window: last_7d
- source: `public.workorders`, `public.workorder_items`
- query_id: `feature_inputs_last_7d`
- params: `project_id`, `limit`
```sql
SELECT
  w.project_id,
  COALESCE(SUM(wi.total_used), 0) AS inputs_total_used
FROM public.workorders w
JOIN public.workorder_items wi ON wi.workorder_id = w.id AND wi.deleted_at IS NULL
WHERE w.project_id = %(project_id)s
  AND w.deleted_at IS NULL
  AND w.date >= (CURRENT_DATE - INTERVAL '7 days')
GROUP BY w.project_id
LIMIT %(limit)s;
```

---

## Workorders / Labores

### feature_workorders_count
- entity: project
- window: all
- source: `v4_report.labor_metrics`
- query_id: `feature_workorders_count`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  COALESCE(SUM(total_workorders), 0) AS workorders_count
FROM v4_report.labor_metrics
WHERE project_id = %(project_id)s
GROUP BY project_id
LIMIT %(limit)s;
```

### feature_workorders_last_30d
- entity: project
- window: last_30d
- source: `public.workorders`
- query_id: `feature_workorders_last_30d`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  COUNT(*) AS workorders_count
FROM public.workorders
WHERE project_id = %(project_id)s
  AND deleted_at IS NULL
  AND date >= (CURRENT_DATE - INTERVAL '30 days')
GROUP BY project_id
LIMIT %(limit)s;
```

### feature_workorders_last_7d
- entity: project
- window: last_7d
- source: `public.workorders`
- query_id: `feature_workorders_last_7d`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  COUNT(*) AS workorders_count
FROM public.workorders
WHERE project_id = %(project_id)s
  AND deleted_at IS NULL
  AND date >= (CURRENT_DATE - INTERVAL '7 days')
GROUP BY project_id
LIMIT %(limit)s;
```

---

## Hectareas

### feature_total_hectares
- entity: project
- window: all
- source: `v4_ssot.total_hectares_for_project`
- query_id: `feature_total_hectares`
- params: `project_id`, `limit`
```sql
SELECT
  %(project_id)s::bigint AS project_id,
  v4_ssot.total_hectares_for_project(%(project_id)s) AS total_hectares
LIMIT %(limit)s;
```

### feature_total_hectares_by_lot
- entity: lot
- window: all
- source: `public.lots`
- query_id: `feature_total_hectares_by_lot`
- params: `project_id`, `limit`
```sql
SELECT
  f.project_id,
  l.id AS lot_id,
  l.hectares AS hectares
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
WHERE f.project_id = %(project_id)s
  AND l.deleted_at IS NULL
LIMIT %(limit)s;
```

---

## Stock / Variacion de stock

### feature_stock_variance
- entity: project
- window: all
- source: `public.stocks`
- query_id: `feature_stock_variance`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  COALESCE(SUM(real_stock_units - initial_units), 0) AS stock_variance_units
FROM public.stocks
WHERE project_id = %(project_id)s
  AND deleted_at IS NULL
GROUP BY project_id
LIMIT %(limit)s;
```

### feature_stock_consumed_by_supply
- entity: project
- window: all
- source: `v4_report.stock_consumed_by_supply`
- query_id: `feature_stock_consumed_by_supply`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  supply_id,
  consumed
FROM v4_report.stock_consumed_by_supply
WHERE project_id = %(project_id)s
LIMIT %(limit)s;
```

---

## Campanas / entidades activas

### feature_project_campaign
- entity: project
- window: all
- source: `public.projects`
- query_id: `feature_project_campaign`
- params: `project_id`, `limit`
```sql
SELECT
  id AS project_id,
  campaign_id
FROM public.projects
WHERE id = %(project_id)s
  AND deleted_at IS NULL
LIMIT %(limit)s;
```

### feature_operational_indicators
- entity: project
- window: all
- source: `v4_report.dashboard_operational_indicators`
- query_id: `feature_operational_indicators`
- params: `project_id`, `limit`
```sql
SELECT
  project_id,
  start_date,
  end_date,
  campaign_closing_date,
  first_workorder_id,
  last_workorder_id,
  last_stock_count_date
FROM v4_report.dashboard_operational_indicators
WHERE project_id = %(project_id)s
LIMIT %(limit)s;
```

---

## Gaps detectados y alternativas

1) **Cohortes por region**  
   - No hay columna de region en `projects`, `fields` o `lots` en `migrations_v4`.
   - Alternativa viable: usar `customer` como proxy si aplica.

2) **Cohortes por cultivo a nivel proyecto**
   - Cultivo esta en `lots.current_crop_id`. No existe `project.crop_id`.
   - Alternativa: derivar via `v4_report.dashboard_crop_incidence` o `v4_report.field_crop_metrics`.

3) **Ventanas temporales para costos desde vistas**
   - Las vistas `v4_report.dashboard_management_balance` y `v4_report.lot_metrics` no tienen columna de fecha.
   - Alternativa: usar `public.workorders.date` (ya propuesto en `feature_cost_total_last_7d/30d`).

4) **since_last_event**
   - No hay tabla de eventos generica.
   - Alternativa: usar `v4_ssot.last_workorder_date_for_project` y `v4_ssot.last_stock_count_date_for_project`.
