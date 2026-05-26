-- Reejecuta el backfill/sync 1:1 de Actores despues de restaurar dumps legacy.
-- Es idempotente y no deduplica por nombre: cada entidad legacy mantiene su actor espejo.

\set ON_ERROR_STOP on

TRUNCATE public.project_responsibles,
         public.project_investor_allocations,
         public.project_admin_cost_allocations,
         public.field_lease_participants
RESTART IDENTITY;

DO $$
DECLARE
    tenant uuid;
    rec record;
    actor_id bigint;
    v_source_key text;
BEGIN
    SELECT id INTO tenant FROM public.auth_tenants WHERE name = 'default' ORDER BY id LIMIT 1;
    IF tenant IS NULL THEN
        RAISE EXCEPTION 'default tenant does not exist';
    END IF;

    FOR rec IN SELECT id, name, created_at, updated_at, deleted_at, created_by::text AS created_by, updated_by::text AS updated_by, deleted_by::text AS deleted_by FROM public.customers LOOP
        v_source_key := rec.id::text;
        IF NOT EXISTS (SELECT 1 FROM public.legacy_actor_map m WHERE m.tenant_id = tenant AND m.source_table = 'customers' AND m.source_key = v_source_key) THEN
            INSERT INTO public.actors (tenant_id, actor_kind, display_name, normalized_name, archived_at, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by)
            VALUES (tenant, 'organization', rec.name, public.normalize_actor_name(rec.name), rec.deleted_at, rec.created_at, rec.updated_at, NULL, rec.created_by, rec.updated_by, rec.deleted_by)
            RETURNING id INTO actor_id;
            INSERT INTO public.actor_roles (actor_id, role, archived_at) VALUES (actor_id, 'cliente', rec.deleted_at) ON CONFLICT DO NOTHING;
            INSERT INTO public.legacy_actor_map (tenant_id, source_table, source_id, source_text, source_key, actor_id, confidence, mapping_status)
            VALUES (tenant, 'customers', rec.id, rec.name, v_source_key, actor_id, 1.0, 'created_new');
        END IF;
    END LOOP;

    FOR rec IN SELECT id, name, created_at, updated_at, deleted_at, created_by::text AS created_by, updated_by::text AS updated_by, deleted_by::text AS deleted_by FROM public.investors LOOP
        v_source_key := rec.id::text;
        IF NOT EXISTS (SELECT 1 FROM public.legacy_actor_map m WHERE m.tenant_id = tenant AND m.source_table = 'investors' AND m.source_key = v_source_key) THEN
            INSERT INTO public.actors (tenant_id, actor_kind, display_name, normalized_name, archived_at, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by)
            VALUES (tenant, 'unknown', rec.name, public.normalize_actor_name(rec.name), rec.deleted_at, rec.created_at, rec.updated_at, NULL, rec.created_by, rec.updated_by, rec.deleted_by)
            RETURNING id INTO actor_id;
            INSERT INTO public.actor_roles (actor_id, role, archived_at) VALUES (actor_id, 'inversor', rec.deleted_at) ON CONFLICT DO NOTHING;
            INSERT INTO public.legacy_actor_map (tenant_id, source_table, source_id, source_text, source_key, actor_id, confidence, mapping_status)
            VALUES (tenant, 'investors', rec.id, rec.name, v_source_key, actor_id, 1.0, 'created_new');
        END IF;
    END LOOP;

    FOR rec IN SELECT id, name, created_at, updated_at, deleted_at, created_by::text AS created_by, updated_by::text AS updated_by, deleted_by::text AS deleted_by FROM public.managers LOOP
        v_source_key := rec.id::text;
        IF NOT EXISTS (SELECT 1 FROM public.legacy_actor_map m WHERE m.tenant_id = tenant AND m.source_table = 'managers' AND m.source_key = v_source_key) THEN
            INSERT INTO public.actors (tenant_id, actor_kind, display_name, normalized_name, archived_at, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by)
            VALUES (tenant, 'natural_person', rec.name, public.normalize_actor_name(rec.name), rec.deleted_at, rec.created_at, rec.updated_at, NULL, rec.created_by, rec.updated_by, rec.deleted_by)
            RETURNING id INTO actor_id;
            INSERT INTO public.actor_roles (actor_id, role, archived_at) VALUES (actor_id, 'responsable', rec.deleted_at) ON CONFLICT DO NOTHING;
            INSERT INTO public.legacy_actor_map (tenant_id, source_table, source_id, source_text, source_key, actor_id, confidence, mapping_status)
            VALUES (tenant, 'managers', rec.id, rec.name, v_source_key, actor_id, 1.0, 'created_new');
        END IF;
    END LOOP;

    FOR rec IN SELECT id, name, created_at, updated_at, deleted_at, created_by::text AS created_by, updated_by::text AS updated_by, deleted_by::text AS deleted_by FROM public.providers LOOP
        v_source_key := rec.id::text;
        IF NOT EXISTS (SELECT 1 FROM public.legacy_actor_map m WHERE m.tenant_id = tenant AND m.source_table = 'providers' AND m.source_key = v_source_key) THEN
            INSERT INTO public.actors (tenant_id, actor_kind, display_name, normalized_name, archived_at, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by)
            VALUES (tenant, 'organization', rec.name, public.normalize_actor_name(rec.name), rec.deleted_at, rec.created_at, rec.updated_at, NULL, rec.created_by, rec.updated_by, rec.deleted_by)
            RETURNING id INTO actor_id;
            INSERT INTO public.actor_roles (actor_id, role, archived_at) VALUES (actor_id, 'proveedor', rec.deleted_at) ON CONFLICT DO NOTHING;
            INSERT INTO public.legacy_actor_map (tenant_id, source_table, source_id, source_text, source_key, actor_id, confidence, mapping_status)
            VALUES (tenant, 'providers', rec.id, rec.name, v_source_key, actor_id, 1.0, 'created_new');
        END IF;
    END LOOP;

    FOR rec IN SELECT DISTINCT contractor_name AS name FROM public.labors WHERE contractor_name IS NOT NULL AND btrim(contractor_name) <> '' LOOP
        v_source_key := public.normalize_actor_name(rec.name);
        IF NOT EXISTS (SELECT 1 FROM public.legacy_actor_map m WHERE m.tenant_id = tenant AND m.source_table = 'labors.contractor_name' AND m.source_key = v_source_key) THEN
            INSERT INTO public.actors (tenant_id, actor_kind, display_name, normalized_name, notes)
            VALUES (tenant, 'unknown', rec.name, public.normalize_actor_name(rec.name), 'Migrado desde texto libre labors.contractor_name')
            RETURNING id INTO actor_id;
            INSERT INTO public.actor_roles (actor_id, role) VALUES (actor_id, 'contratista') ON CONFLICT DO NOTHING;
            INSERT INTO public.legacy_actor_map (tenant_id, source_table, source_text, source_key, actor_id, confidence, mapping_status)
            VALUES (tenant, 'labors.contractor_name', rec.name, v_source_key, actor_id, 0.6, 'manual_review');
        END IF;
    END LOOP;

    FOR rec IN SELECT DISTINCT contractor AS name FROM public.workorders WHERE contractor IS NOT NULL AND btrim(contractor) <> '' LOOP
        v_source_key := public.normalize_actor_name(rec.name);
        IF NOT EXISTS (SELECT 1 FROM public.legacy_actor_map m WHERE m.tenant_id = tenant AND m.source_table = 'workorders.contractor' AND m.source_key = v_source_key) THEN
            INSERT INTO public.actors (tenant_id, actor_kind, display_name, normalized_name, notes)
            VALUES (tenant, 'unknown', rec.name, public.normalize_actor_name(rec.name), 'Migrado desde texto libre workorders.contractor')
            RETURNING id INTO actor_id;
            INSERT INTO public.actor_roles (actor_id, role) VALUES (actor_id, 'contratista') ON CONFLICT DO NOTHING;
            INSERT INTO public.legacy_actor_map (tenant_id, source_table, source_text, source_key, actor_id, confidence, mapping_status)
            VALUES (tenant, 'workorders.contractor', rec.name, v_source_key, actor_id, 0.6, 'manual_review');
        END IF;
    END LOOP;

    FOR rec IN SELECT DISTINCT company AS name FROM public.invoices WHERE company IS NOT NULL AND btrim(company) <> '' LOOP
        v_source_key := public.normalize_actor_name(rec.name);
        IF NOT EXISTS (SELECT 1 FROM public.legacy_actor_map m WHERE m.tenant_id = tenant AND m.source_table = 'invoices.company' AND m.source_key = v_source_key) THEN
            INSERT INTO public.actors (tenant_id, actor_kind, display_name, normalized_name, notes)
            VALUES (tenant, 'organization', rec.name, public.normalize_actor_name(rec.name), 'Migrado desde texto libre invoices.company')
            RETURNING id INTO actor_id;
            INSERT INTO public.actor_roles (actor_id, role) VALUES (actor_id, 'facturador') ON CONFLICT DO NOTHING;
            INSERT INTO public.legacy_actor_map (tenant_id, source_table, source_text, source_key, actor_id, confidence, mapping_status)
            VALUES (tenant, 'invoices.company', rec.name, v_source_key, actor_id, 0.6, 'manual_review');
        END IF;
    END LOOP;
END $$;

WITH unambiguous_customer_actor AS (
    SELECT m.tenant_id, m.source_id, m.actor_id
    FROM public.legacy_actor_map m
    WHERE m.source_table = 'customers'
      AND m.actor_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1
          FROM public.legacy_actor_map other
          WHERE other.tenant_id = m.tenant_id
            AND other.source_table = 'customers'
            AND other.actor_id = m.actor_id
            AND other.source_id IS DISTINCT FROM m.source_id
      )
)
UPDATE public.customers c
SET actor_id = m.actor_id
FROM unambiguous_customer_actor m
WHERE m.source_id = c.id
  AND m.tenant_id = c.tenant_id
  AND c.actor_id IS NULL;

UPDATE public.projects p
SET customer_actor_id = COALESCE(c.actor_id, m.actor_id)
FROM public.customers c
LEFT JOIN public.legacy_actor_map m
  ON m.source_table = 'customers'
 AND m.source_id = c.id
 AND m.tenant_id = c.tenant_id
WHERE c.id = p.customer_id
  AND c.tenant_id = p.tenant_id
  AND COALESCE(c.actor_id, m.actor_id) IS NOT NULL;

UPDATE public.workorders w
SET investor_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = w.investor_id;

UPDATE public.workorders w
SET contractor_actor_id = m.actor_id,
    contractor_name_snapshot = COALESCE(w.contractor_name_snapshot, w.contractor)
FROM public.legacy_actor_map m
WHERE m.source_table = 'workorders.contractor'
  AND m.source_key = public.normalize_actor_name(w.contractor);

UPDATE public.workorder_investor_splits s
SET actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = s.investor_id;

UPDATE public.stocks s
SET investor_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = s.investor_id;

UPDATE public.supply_movements sm
SET investor_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = sm.investor_id;

UPDATE public.supply_movements sm
SET provider_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'providers'
  AND m.source_id = sm.provider_id;

UPDATE public.labors l
SET contractor_actor_id = m.actor_id,
    contractor_name_snapshot = COALESCE(l.contractor_name_snapshot, l.contractor_name)
FROM public.legacy_actor_map m
WHERE m.source_table = 'labors.contractor_name'
  AND m.source_key = public.normalize_actor_name(l.contractor_name);

UPDATE public.invoices i
SET investor_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = i.investor_id;

UPDATE public.invoices i
SET company_actor_id = m.actor_id,
    company_name_snapshot = COALESCE(i.company_name_snapshot, i.company)
FROM public.legacy_actor_map m
WHERE m.source_table = 'invoices.company'
  AND m.source_key = public.normalize_actor_name(i.company);

INSERT INTO public.project_responsibles (project_id, actor_id, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by)
SELECT pm.project_id, m.actor_id, pm.created_at, pm.updated_at, pm.deleted_at, pm.created_by::text, pm.updated_by::text, pm.deleted_by::text
FROM public.project_managers pm
JOIN public.legacy_actor_map m ON m.source_table = 'managers' AND m.source_id = pm.manager_id
ON CONFLICT DO NOTHING;

INSERT INTO public.project_investor_allocations (project_id, actor_id, percentage, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by)
SELECT pi.project_id, m.actor_id, pi.percentage, pi.created_at, pi.updated_at, pi.deleted_at, pi.created_by::text, pi.updated_by::text, pi.deleted_by::text
FROM public.project_investors pi
JOIN public.legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = pi.investor_id
ON CONFLICT DO NOTHING;

INSERT INTO public.project_admin_cost_allocations (project_id, actor_id, percentage, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by)
SELECT aci.project_id, m.actor_id, aci.percentage, aci.created_at, aci.updated_at, aci.deleted_at, aci.created_by::text, aci.updated_by::text, aci.deleted_by::text
FROM public.admin_cost_investors aci
JOIN public.legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = aci.investor_id
ON CONFLICT DO NOTHING;

INSERT INTO public.field_lease_participants (field_id, actor_id, percentage, created_at, updated_at, deleted_at, created_by, updated_by, deleted_by)
SELECT fi.field_id, m.actor_id, fi.percentage, fi.created_at, fi.updated_at, fi.deleted_at, fi.created_by::text, fi.updated_by::text, fi.deleted_by::text
FROM public.field_investors fi
JOIN public.legacy_actor_map m ON m.source_table = 'investors' AND m.source_id = fi.investor_id
ON CONFLICT DO NOTHING;

SELECT 'actors' AS subject, COUNT(*) AS count FROM public.actors
UNION ALL
SELECT 'legacy_actor_map', COUNT(*) FROM public.legacy_actor_map
UNION ALL
SELECT 'project_investor_allocations', COUNT(*) FROM public.project_investor_allocations
UNION ALL
SELECT 'project_responsibles', COUNT(*) FROM public.project_responsibles;
