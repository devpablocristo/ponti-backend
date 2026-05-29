-- Tenant isolation audit. This script is read-only and should return zero rows
-- for every section before TENANT_STRICT_MODE is enabled.

WITH tenant_owned(table_name) AS (
    VALUES
        ('customers'),
        ('projects'),
        ('campaigns'),
        ('fields'),
        ('lots'),
        ('lot_dates'),
        ('workorders'),
        ('workorder_items'),
        ('workorder_investor_splits'),
        ('workorder_supply_items'),
        ('work_order_drafts'),
        ('project_managers'),
        ('project_investors'),
        ('admin_cost_investors'),
        ('field_investors'),
        ('labors'),
        ('supplies'),
        ('supply_movements'),
        ('stock_movements'),
        ('stocks'),
        ('invoices'),
        ('investors'),
        ('managers'),
        ('providers'),
        ('crops'),
        ('categories'),
        ('class_types'),
        ('lease_types'),
        ('business_parameters'),
        ('crop_commercializations'),
        ('project_dollar_values'),
        ('actors'),
        ('actor_aliases'),
        ('actor_identifiers'),
        ('legacy_actor_map'),
        ('tenant_invites')
)
SELECT 'missing_tenant_id_column' AS check_name, table_name, NULL::bigint AS count
FROM tenant_owned t
WHERE to_regclass('public.' || t.table_name) IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM information_schema.columns c
      WHERE c.table_schema = 'public'
        AND c.table_name = t.table_name
        AND c.column_name = 'tenant_id'
  );

DO $$
DECLARE
    t text;
    n bigint;
BEGIN
    CREATE TEMP TABLE IF NOT EXISTS tenant_audit_results (
        check_name text,
        table_name text,
        count bigint
    );
    DELETE FROM tenant_audit_results;

    FOR t IN
        SELECT table_name
        FROM (VALUES
            ('customers'), ('projects'), ('campaigns'), ('fields'), ('lots'),
            ('lot_dates'), ('workorders'), ('workorder_items'),
            ('workorder_investor_splits'), ('workorder_supply_items'),
            ('work_order_drafts'), ('project_managers'), ('project_investors'),
            ('admin_cost_investors'), ('field_investors'), ('labors'), ('supplies'), ('supply_movements'),
            ('stock_movements'), ('stocks'), ('invoices'), ('investors'),
            ('managers'), ('providers'), ('crops'), ('categories'), ('class_types'),
            ('lease_types'), ('business_parameters'), ('crop_commercializations'), ('project_dollar_values'),
            ('actors'), ('actor_aliases'), ('actor_identifiers'), ('legacy_actor_map'), ('tenant_invites')
        ) AS v(table_name)
    LOOP
        IF to_regclass('public.' || t) IS NOT NULL
           AND EXISTS (
                SELECT 1 FROM information_schema.columns
                WHERE table_schema = 'public' AND table_name = t AND column_name = 'tenant_id'
           )
        THEN
            EXECUTE format('SELECT COUNT(*) FROM public.%I WHERE tenant_id IS NULL', t) INTO n;
            IF n > 0 THEN
                INSERT INTO tenant_audit_results VALUES ('tenant_id_null_rows', t, n);
            END IF;
        END IF;
    END LOOP;
END $$;

SELECT * FROM tenant_audit_results ORDER BY check_name, table_name;

SELECT 'project_customer_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.projects p
JOIN public.customers c ON c.id = p.customer_id
WHERE p.tenant_id IS DISTINCT FROM c.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_campaign_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.projects p
JOIN public.campaigns c ON c.id = p.campaign_id
WHERE p.tenant_id IS DISTINCT FROM c.tenant_id
HAVING COUNT(*) > 0;

SELECT 'field_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.fields f
JOIN public.projects p ON p.id = f.project_id
WHERE f.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'lot_field_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.lots l
JOIN public.fields f ON f.id = l.field_id
WHERE l.tenant_id IS DISTINCT FROM f.tenant_id
HAVING COUNT(*) > 0;

SELECT 'workorder_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.workorders w
JOIN public.projects p ON p.id = w.project_id
WHERE w.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'supply_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.supplies s
JOIN public.projects p ON p.id = s.project_id
WHERE s.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'supply_movement_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.supply_movements sm
JOIN public.projects p ON p.id = sm.project_id
WHERE sm.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'stock_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.stocks s
JOIN public.projects p ON p.id = s.project_id
WHERE s.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_manager_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.project_managers pm
JOIN public.projects p ON p.id = pm.project_id
WHERE pm.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_manager_manager_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.project_managers pm
JOIN public.managers m ON m.id = pm.manager_id
WHERE pm.tenant_id IS DISTINCT FROM m.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_investor_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.project_investors pi
JOIN public.projects p ON p.id = pi.project_id
WHERE pi.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_investor_investor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.project_investors pi
JOIN public.investors i ON i.id = pi.investor_id
WHERE pi.tenant_id IS DISTINCT FROM i.tenant_id
HAVING COUNT(*) > 0;

SELECT 'admin_cost_investor_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.admin_cost_investors aci
JOIN public.projects p ON p.id = aci.project_id
WHERE aci.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'admin_cost_investor_investor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.admin_cost_investors aci
JOIN public.investors i ON i.id = aci.investor_id
WHERE aci.tenant_id IS DISTINCT FROM i.tenant_id
HAVING COUNT(*) > 0;

SELECT 'field_investor_field_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.field_investors fi
JOIN public.fields f ON f.id = fi.field_id
WHERE fi.tenant_id IS DISTINCT FROM f.tenant_id
HAVING COUNT(*) > 0;

SELECT 'field_investor_investor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.field_investors fi
JOIN public.investors i ON i.id = fi.investor_id
WHERE fi.tenant_id IS DISTINCT FROM i.tenant_id
HAVING COUNT(*) > 0;

SELECT 'lot_dates_lot_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.lot_dates ld
JOIN public.lots l ON l.id = ld.lot_id
WHERE ld.tenant_id IS DISTINCT FROM l.tenant_id
HAVING COUNT(*) > 0;

SELECT 'workorder_item_workorder_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.workorder_items wi
JOIN public.workorders w ON w.id = wi.workorder_id
WHERE wi.tenant_id IS DISTINCT FROM w.tenant_id
HAVING COUNT(*) > 0;

SELECT 'workorder_split_workorder_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.workorder_investor_splits wis
JOIN public.workorders w ON w.id = wis.workorder_id
WHERE wis.tenant_id IS DISTINCT FROM w.tenant_id
HAVING COUNT(*) > 0;

DO $$
DECLARE
    mismatch_count bigint;
BEGIN
    CREATE TEMP TABLE IF NOT EXISTS tenant_optional_audit_results (
        check_name text,
        table_name text,
        count bigint
    );
    DELETE FROM tenant_optional_audit_results;

    IF to_regclass('public.workorder_supply_items') IS NOT NULL THEN
        EXECUTE $SQL$
            SELECT COUNT(*)
            FROM public.workorder_supply_items wsi
            JOIN public.workorders w ON w.id = wsi.workorder_id
            WHERE wsi.tenant_id IS DISTINCT FROM w.tenant_id
        $SQL$ INTO mismatch_count;

        IF mismatch_count > 0 THEN
            INSERT INTO tenant_optional_audit_results
            VALUES ('workorder_supply_item_workorder_cross_tenant', 'workorder_supply_items', mismatch_count);
        END IF;
    END IF;
END $$;

SELECT * FROM tenant_optional_audit_results ORDER BY check_name, table_name;

SELECT 'invoice_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.invoices i
JOIN public.workorders w ON w.id = i.work_order_id
JOIN public.projects p ON p.id = w.project_id
WHERE i.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'crop_commercialization_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.crop_commercializations cc
JOIN public.projects p ON p.id = cc.project_id
WHERE cc.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_dollar_value_project_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.project_dollar_values pdv
JOIN public.projects p ON p.id = pdv.project_id
WHERE pdv.tenant_id IS DISTINCT FROM p.tenant_id
HAVING COUNT(*) > 0;

SELECT 'actor_alias_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.actor_aliases aa
JOIN public.actors a ON a.id = aa.actor_id
WHERE aa.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'actor_identifier_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.actor_identifiers ai
JOIN public.actors a ON a.id = ai.actor_id
WHERE ai.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'legacy_actor_map_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.legacy_actor_map lam
JOIN public.actors a ON a.id = lam.actor_id
WHERE lam.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'actor_role_orphan_actor' AS check_name, COUNT(*) AS count
FROM public.actor_roles ar
LEFT JOIN public.actors a ON a.id = ar.actor_id
WHERE a.id IS NULL
HAVING COUNT(*) > 0;

SELECT 'actor_person_profile_orphan_actor' AS check_name, COUNT(*) AS count
FROM public.actor_person_profiles app
LEFT JOIN public.actors a ON a.id = app.actor_id
WHERE a.id IS NULL
HAVING COUNT(*) > 0;

SELECT 'actor_organization_profile_orphan_actor' AS check_name, COUNT(*) AS count
FROM public.actor_organization_profiles aop
LEFT JOIN public.actors a ON a.id = aop.actor_id
WHERE a.id IS NULL
HAVING COUNT(*) > 0;

SELECT 'actor_relationship_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.actor_relationships ar
JOIN public.actors fa ON fa.id = ar.from_actor_id
JOIN public.actors ta ON ta.id = ar.to_actor_id
WHERE fa.tenant_id IS DISTINCT FROM ta.tenant_id
HAVING COUNT(*) > 0;

SELECT 'actor_merge_log_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.actor_merge_log aml
JOIN public.actors fa ON fa.id = aml.from_actor_id
JOIN public.actors ta ON ta.id = aml.to_actor_id
WHERE fa.tenant_id IS DISTINCT FROM ta.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_customer_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.projects p
JOIN public.actors a ON a.id = p.customer_actor_id
WHERE p.customer_actor_id IS NOT NULL
  AND p.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'workorder_investor_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.workorders w
JOIN public.actors a ON a.id = w.investor_actor_id
WHERE w.investor_actor_id IS NOT NULL
  AND w.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'workorder_contractor_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.workorders w
JOIN public.actors a ON a.id = w.contractor_actor_id
WHERE w.contractor_actor_id IS NOT NULL
  AND w.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'workorder_split_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.workorder_investor_splits wis
JOIN public.actors a ON a.id = wis.actor_id
WHERE wis.actor_id IS NOT NULL
  AND wis.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'stock_investor_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.stocks s
JOIN public.actors a ON a.id = s.investor_actor_id
WHERE s.investor_actor_id IS NOT NULL
  AND s.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'supply_movement_investor_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.supply_movements sm
JOIN public.actors a ON a.id = sm.investor_actor_id
WHERE sm.investor_actor_id IS NOT NULL
  AND sm.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'supply_movement_provider_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.supply_movements sm
JOIN public.actors a ON a.id = sm.provider_actor_id
WHERE sm.provider_actor_id IS NOT NULL
  AND sm.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'labor_contractor_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.labors l
JOIN public.actors a ON a.id = l.contractor_actor_id
WHERE l.contractor_actor_id IS NOT NULL
  AND l.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'invoice_investor_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.invoices i
JOIN public.actors a ON a.id = i.investor_actor_id
WHERE i.investor_actor_id IS NOT NULL
  AND i.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'invoice_company_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.invoices i
JOIN public.actors a ON a.id = i.company_actor_id
WHERE i.company_actor_id IS NOT NULL
  AND i.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_responsible_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.project_responsibles pr
JOIN public.projects p ON p.id = pr.project_id
JOIN public.actors a ON a.id = pr.actor_id
WHERE p.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_investor_allocation_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.project_investor_allocations pia
JOIN public.projects p ON p.id = pia.project_id
JOIN public.actors a ON a.id = pia.actor_id
WHERE p.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'project_admin_cost_allocation_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.project_admin_cost_allocations paca
JOIN public.projects p ON p.id = paca.project_id
JOIN public.actors a ON a.id = paca.actor_id
WHERE p.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'field_lease_participant_actor_cross_tenant' AS check_name, COUNT(*) AS count
FROM public.field_lease_participants flp
JOIN public.fields f ON f.id = flp.field_id
JOIN public.actors a ON a.id = flp.actor_id
WHERE f.tenant_id IS DISTINCT FROM a.tenant_id
HAVING COUNT(*) > 0;

SELECT 'global_unique_name_constraint' AS check_name, rel.relname AS table_name, COUNT(*) AS count
FROM pg_constraint c
JOIN pg_class rel ON rel.oid = c.conrelid
JOIN pg_namespace nsp ON nsp.oid = rel.relnamespace
WHERE nsp.nspname = 'public'
  AND c.contype = 'u'
  AND rel.relname IN (
      'customers', 'projects', 'campaigns', 'fields', 'lots',
      'investors', 'managers', 'providers', 'supplies', 'labors',
      'crops', 'categories', 'class_types', 'lease_types'
  )
  AND (
      SELECT array_agg(att.attname ORDER BY ord.ordinality)
      FROM unnest(c.conkey) WITH ORDINALITY AS ord(attnum, ordinality)
      JOIN pg_attribute att ON att.attrelid = rel.oid AND att.attnum = ord.attnum
  ) = ARRAY['name']::name[]
GROUP BY rel.relname
HAVING COUNT(*) > 0;
