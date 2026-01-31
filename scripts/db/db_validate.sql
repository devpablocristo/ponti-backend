-- Validaciones automáticas de esquema

-- 0) Schemas esperados
DO $$
DECLARE
  v_cnt integer;
BEGIN
  WITH expected(schema_name) AS (
    VALUES ('public'), ('v4_core'), ('v4_ssot'), ('v4_calc'), ('v4_report')
  ),
  missing AS (
    SELECT e.schema_name
    FROM expected e
    LEFT JOIN pg_namespace n ON n.nspname = e.schema_name
    WHERE n.oid IS NULL
  )
  SELECT COUNT(*) INTO v_cnt FROM missing;

  IF v_cnt > 0 THEN
    RAISE EXCEPTION 'Faltan schemas esperados (%).', v_cnt;
  END IF;
END;
$$;

-- 0.1) Tablas esperadas en public
DO $$
DECLARE
  v_cnt integer;
BEGIN
  WITH expected(table_name) AS (
    VALUES
      ('users'),
      ('business_parameters'),
      ('fx_rates'),
      ('customers'),
      ('campaigns'),
      ('projects'),
      ('managers'),
      ('project_managers'),
      ('lease_types'),
      ('fields'),
      ('lots'),
      ('lot_dates'),
      ('crops'),
      ('labor_types'),
      ('labor_categories'),
      ('labors'),
      ('workorders'),
      ('workorder_items'),
      ('types'),
      ('categories'),
      ('supplies'),
      ('stocks'),
      ('supply_movements'),
      ('providers'),
      ('investors'),
      ('project_investors'),
      ('crop_commercializations'),
      ('admin_cost_investors'),
      ('field_investors'),
      ('project_dollar_values'),
      ('invoices')
  ),
  missing AS (
    SELECT e.table_name
    FROM expected e
    LEFT JOIN pg_class c ON c.relname = e.table_name
    LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE c.oid IS NULL OR n.nspname <> 'public'
  )
  SELECT COUNT(*) INTO v_cnt FROM missing;

  IF v_cnt > 0 THEN
    RAISE EXCEPTION 'Faltan tablas esperadas en public (%).', v_cnt;
  END IF;
END;
$$;

-- 1) Constraints UNIQUE duplicadas por mismas columnas
DO $$
DECLARE
  v_cnt integer;
BEGIN
  WITH uniq AS (
    SELECT
      conrelid::regclass AS table_name,
      string_agg(att.attname, ',' ORDER BY att.attname) AS cols,
      COUNT(DISTINCT con.conname) AS cnt
    FROM pg_constraint con
    JOIN unnest(con.conkey) WITH ORDINALITY AS ck(attnum, ord) ON true
    JOIN pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = ck.attnum
    WHERE con.contype = 'u'
    GROUP BY conrelid, con.conkey
  )
  SELECT COUNT(*) INTO v_cnt
  FROM uniq
  WHERE cnt > 1;

  IF v_cnt > 0 THEN
    RAISE EXCEPTION 'Se detectaron % UNIQUE duplicadas', v_cnt;
  END IF;
END;
$$;

-- 2) FKs sin índice (prefijo)
DO $$
DECLARE
  v_cnt integer;
BEGIN
  WITH fks AS (
    SELECT
      con.oid,
      con.conrelid,
      con.conkey,
      con.conname,
      con.conrelid::regclass AS table_name
    FROM pg_constraint con
    WHERE con.contype = 'f'
  ),
  fk_cols AS (
    SELECT
      fks.oid,
      fks.conrelid,
      fks.conname,
      fks.table_name,
      array_agg(att.attname ORDER BY ord) AS cols
    FROM fks
    JOIN unnest(fks.conkey) WITH ORDINALITY AS ck(attnum, ord) ON true
    JOIN pg_attribute att ON att.attrelid = fks.conrelid AND att.attnum = ck.attnum
    GROUP BY fks.oid, fks.conrelid, fks.conname, fks.table_name
  ),
  idx_cols AS (
    SELECT
      idx.indrelid,
      array_agg(att.attname ORDER BY ord) AS cols
    FROM pg_index idx
    JOIN unnest(idx.indkey) WITH ORDINALITY AS ik(attnum, ord) ON true
    JOIN pg_attribute att ON att.attrelid = idx.indrelid AND att.attnum = ik.attnum
    GROUP BY idx.indrelid, idx.indexrelid
  ),
  fk_without_index AS (
    SELECT f.table_name, f.conname, f.cols
    FROM fk_cols f
    LEFT JOIN idx_cols i
      ON i.indrelid = f.conrelid
     AND i.cols[1:array_length(f.cols, 1)] = f.cols
    WHERE i.indrelid IS NULL
  )
  SELECT COUNT(*) INTO v_cnt FROM fk_without_index;

  IF v_cnt > 0 THEN
    RAISE EXCEPTION 'Se detectaron % FKs sin índice', v_cnt;
  END IF;
END;
$$;

-- 2.1) Vistas esperadas en v4_calc y v4_report
DO $$
DECLARE
  v_cnt integer;
BEGIN
  WITH expected(schema_name, view_name) AS (
    VALUES
      ('v4_calc', 'workorder_metrics'),
      ('v4_calc', 'workorder_metrics_raw'),
      ('v4_calc', 'lot_base_costs'),
      ('v4_calc', 'lot_base_income'),
      ('v4_calc', 'field_crop_lot_base'),
      ('v4_calc', 'field_crop_supply_costs_by_lot'),
      ('v4_calc', 'field_crop_labor_costs_by_lot'),
      ('v4_calc', 'field_crop_aggregated'),
      ('v4_calc', 'field_crop_metrics_lot_base'),
      ('v4_calc', 'field_crop_metrics_aggregated'),
      ('v4_calc', 'dashboard_fertilizers_invested_by_project'),
      ('v4_calc', 'dashboard_supply_costs_by_project'),
      ('v4_calc', 'investor_contribution_categories'),
      ('v4_calc', 'investor_real_contributions'),
      ('v4_report', 'lot_metrics'),
      ('v4_report', 'lot_list'),
      ('v4_report', 'labor_list'),
      ('v4_report', 'labor_metrics'),
      ('v4_report', 'field_crop_cultivos'),
      ('v4_report', 'field_crop_economicos'),
      ('v4_report', 'field_crop_insumos'),
      ('v4_report', 'field_crop_labores'),
      ('v4_report', 'field_crop_metrics'),
      ('v4_report', 'field_crop_rentabilidad'),
      ('v4_report', 'summary_results'),
      ('v4_report', 'dashboard_management_balance'),
      ('v4_report', 'dashboard_metrics'),
      ('v4_report', 'dashboard_crop_incidence'),
      ('v4_report', 'dashboard_operational_indicators'),
      ('v4_report', 'investor_project_base'),
      ('v4_report', 'investor_contribution_categories'),
      ('v4_report', 'investor_distributions'),
      ('v4_report', 'investor_contribution_data'),
      ('v4_report', 'dashboard_contributions_progress'),
      ('v4_report', 'workorder_list'),
      ('v4_report', 'workorder_metrics'),
      ('v4_report', 'stock_consumed_by_supply')
  ),
  missing AS (
    SELECT e.schema_name, e.view_name
    FROM expected e
    LEFT JOIN pg_views v ON v.schemaname = e.schema_name AND v.viewname = e.view_name
    WHERE v.viewname IS NULL
  )
  SELECT COUNT(*) INTO v_cnt FROM missing;

  IF v_cnt > 0 THEN
    RAISE EXCEPTION 'Faltan vistas esperadas (%).', v_cnt;
  END IF;
END;
$$;

-- 3) Vistas válidas (intenta SELECT 1)
DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT schemaname, viewname
    FROM pg_views
    WHERE schemaname IN ('v4_calc', 'v4_report')
  LOOP
    EXECUTE format('SELECT 1 FROM %I.%I LIMIT 1', r.schemaname, r.viewname);
  END LOOP;
END;
$$;
