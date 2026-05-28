-- ========================================
-- MIGRATION 000223 ACTORS SAFE MIGRATION (UP)
-- ========================================
-- Nota: esta migracion crea el modelo nuevo de Actores sin cambiar lecturas
-- productivas ni formulas existentes. El backfill es 1:1 por entidad legacy.

BEGIN;

CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

INSERT INTO public.auth_tenants (name)
VALUES ('default')
ON CONFLICT (name) DO NOTHING;

CREATE OR REPLACE FUNCTION public.normalize_actor_name(input text)
RETURNS text
LANGUAGE sql
STABLE
AS $$
  SELECT NULLIF(
    regexp_replace(
      lower(unaccent(btrim(COALESCE(input, '')))),
      '\s+',
      ' ',
      'g'
    ),
    ''
  )
$$;

CREATE TABLE IF NOT EXISTS public.actors (
    id bigserial PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE RESTRICT,
    actor_kind text NOT NULL DEFAULT 'unknown',
    display_name text NOT NULL,
    normalized_name text NOT NULL,
    primary_email text,
    primary_phone text,
    notes text,
    archived_at timestamp with time zone,
    merged_into_actor_id bigint REFERENCES public.actors(id) ON DELETE RESTRICT,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by text,
    updated_by text,
    deleted_by text,
    CONSTRAINT chk_actors_kind CHECK (actor_kind IN ('natural_person', 'organization', 'other', 'unknown')),
    CONSTRAINT chk_actors_not_self_merged CHECK (merged_into_actor_id IS NULL OR merged_into_actor_id <> id)
);

CREATE INDEX IF NOT EXISTS idx_actors_tenant ON public.actors (tenant_id);
CREATE INDEX IF NOT EXISTS idx_actors_archived_at ON public.actors (archived_at);
CREATE INDEX IF NOT EXISTS idx_actors_merged_into ON public.actors (merged_into_actor_id);
CREATE INDEX IF NOT EXISTS idx_actors_normalized_name_trgm
    ON public.actors USING gin (normalized_name gin_trgm_ops);

CREATE TABLE IF NOT EXISTS public.actor_person_profiles (
    actor_id bigint PRIMARY KEY REFERENCES public.actors(id) ON DELETE CASCADE,
    first_name text,
    last_name text,
    birth_date date,
    document_type text,
    document_number text,
    normalized_document_number text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS public.actor_organization_profiles (
    actor_id bigint PRIMARY KEY REFERENCES public.actors(id) ON DELETE CASCADE,
    legal_name text,
    normalized_legal_name text,
    trade_name text,
    normalized_trade_name text,
    legal_entity_type text,
    tax_condition text,
    fiscal_address text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_actor_org_legal_name_trgm
    ON public.actor_organization_profiles USING gin (normalized_legal_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_actor_org_trade_name_trgm
    ON public.actor_organization_profiles USING gin (normalized_trade_name gin_trgm_ops);

CREATE TABLE IF NOT EXISTS public.actor_identifiers (
    id bigserial PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE RESTRICT,
    actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE CASCADE,
    country text NOT NULL DEFAULT 'AR',
    identifier_type text NOT NULL,
    identifier_value text NOT NULL,
    normalized_identifier_value text NOT NULL,
    is_primary boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_actor_identifiers_unique_value
    ON public.actor_identifiers (tenant_id, country, identifier_type, normalized_identifier_value)
    WHERE normalized_identifier_value IS NOT NULL AND normalized_identifier_value <> '';
CREATE INDEX IF NOT EXISTS idx_actor_identifiers_actor ON public.actor_identifiers (actor_id);

CREATE TABLE IF NOT EXISTS public.actor_roles (
    actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE CASCADE,
    role text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    archived_at timestamp with time zone,
    PRIMARY KEY (actor_id, role),
    CONSTRAINT chk_actor_roles_role CHECK (role IN (
      'cliente',
      'responsable',
      'inversor',
      'arrendatario',
      'proveedor',
      'contratista',
      'facturador'
    ))
);

CREATE INDEX IF NOT EXISTS idx_actor_roles_role ON public.actor_roles (role) WHERE archived_at IS NULL;

CREATE TABLE IF NOT EXISTS public.actor_aliases (
    id bigserial PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE RESTRICT,
    actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE CASCADE,
    alias text NOT NULL,
    normalized_alias text NOT NULL,
    source text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    archived_at timestamp with time zone
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_actor_aliases_actor_alias
    ON public.actor_aliases (actor_id, normalized_alias)
    WHERE archived_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actor_aliases_trgm
    ON public.actor_aliases USING gin (normalized_alias gin_trgm_ops);

CREATE TABLE IF NOT EXISTS public.actor_relationships (
    id bigserial PRIMARY KEY,
    from_actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE CASCADE,
    to_actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE CASCADE,
    relationship_type text NOT NULL,
    start_date date,
    end_date date,
    status text NOT NULL DEFAULT 'active',
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT chk_actor_relationship_not_self CHECK (from_actor_id <> to_actor_id)
);

CREATE INDEX IF NOT EXISTS idx_actor_relationships_from ON public.actor_relationships (from_actor_id);
CREATE INDEX IF NOT EXISTS idx_actor_relationships_to ON public.actor_relationships (to_actor_id);

CREATE TABLE IF NOT EXISTS public.actor_merge_log (
    id bigserial PRIMARY KEY,
    from_actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE RESTRICT,
    to_actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE RESTRICT,
    merged_by text,
    merged_at timestamp with time zone DEFAULT now() NOT NULL,
    reason text,
    impact jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS public.legacy_actor_map (
    id bigserial PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE RESTRICT,
    source_table text NOT NULL,
    source_id bigint,
    source_text text,
    source_key text NOT NULL,
    actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE RESTRICT,
    confidence numeric(5,4) NOT NULL DEFAULT 1.0,
    mapping_status text NOT NULL DEFAULT 'created_new',
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chk_legacy_actor_map_status CHECK (mapping_status IN ('auto_matched', 'manual_review', 'created_new', 'ignored'))
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_legacy_actor_map_source
    ON public.legacy_actor_map (tenant_id, source_table, source_key);
CREATE INDEX IF NOT EXISTS idx_legacy_actor_map_actor ON public.legacy_actor_map (actor_id);

-- Columnas paralelas. No reemplazan las columnas legacy en esta fase.
ALTER TABLE public.projects
    ADD COLUMN IF NOT EXISTS customer_actor_id bigint;

ALTER TABLE public.workorders
    ADD COLUMN IF NOT EXISTS investor_actor_id bigint,
    ADD COLUMN IF NOT EXISTS contractor_actor_id bigint,
    ADD COLUMN IF NOT EXISTS contractor_name_snapshot text;

ALTER TABLE public.workorder_investor_splits
    ADD COLUMN IF NOT EXISTS actor_id bigint;

ALTER TABLE public.stocks
    ADD COLUMN IF NOT EXISTS investor_actor_id bigint;

ALTER TABLE public.supply_movements
    ADD COLUMN IF NOT EXISTS investor_actor_id bigint,
    ADD COLUMN IF NOT EXISTS provider_actor_id bigint;

ALTER TABLE public.labors
    ADD COLUMN IF NOT EXISTS contractor_actor_id bigint,
    ADD COLUMN IF NOT EXISTS contractor_name_snapshot text;

ALTER TABLE public.invoices
    ADD COLUMN IF NOT EXISTS investor_actor_id bigint,
    ADD COLUMN IF NOT EXISTS company_actor_id bigint,
    ADD COLUMN IF NOT EXISTS company_name_snapshot text;

CREATE TABLE IF NOT EXISTS public.project_responsibles (
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES public.projects(id) ON DELETE CASCADE,
    actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE RESTRICT,
    start_date date,
    end_date date,
    status text NOT NULL DEFAULT 'active',
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by text,
    updated_by text,
    deleted_by text
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_project_responsibles_active
    ON public.project_responsibles (project_id, actor_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS public.project_investor_allocations (
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES public.projects(id) ON DELETE CASCADE,
    actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE RESTRICT,
    percentage integer NOT NULL,
    start_date date,
    end_date date,
    status text NOT NULL DEFAULT 'active',
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by text,
    updated_by text,
    deleted_by text
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_project_investor_allocations_active
    ON public.project_investor_allocations (project_id, actor_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS public.project_admin_cost_allocations (
    id bigserial PRIMARY KEY,
    project_id bigint NOT NULL REFERENCES public.projects(id) ON DELETE CASCADE,
    actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE RESTRICT,
    percentage integer NOT NULL,
    start_date date,
    end_date date,
    status text NOT NULL DEFAULT 'active',
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by text,
    updated_by text,
    deleted_by text
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_project_admin_cost_allocations_active
    ON public.project_admin_cost_allocations (project_id, actor_id)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS public.field_lease_participants (
    id bigserial PRIMARY KEY,
    field_id bigint NOT NULL REFERENCES public.fields(id) ON DELETE CASCADE,
    actor_id bigint NOT NULL REFERENCES public.actors(id) ON DELETE RESTRICT,
    percentage integer NOT NULL,
    start_date date,
    end_date date,
    status text NOT NULL DEFAULT 'active',
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by text,
    updated_by text,
    deleted_by text
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_field_lease_participants_active
    ON public.field_lease_participants (field_id, actor_id)
    WHERE deleted_at IS NULL;

-- Backfill 1:1. A proposito no se deduplica entre tablas ni por nombre.
DO $$
DECLARE
    tenant uuid;
    rec record;
    actor_id bigint;
    v_source_key text;
BEGIN
    SELECT id INTO tenant FROM public.auth_tenants WHERE name = 'default' ORDER BY id LIMIT 1;

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

UPDATE public.projects p
SET customer_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'customers'
  AND m.source_id = p.customer_id
  AND p.customer_actor_id IS NULL;

UPDATE public.workorders w
SET investor_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = w.investor_id
  AND w.investor_actor_id IS NULL;

UPDATE public.workorders w
SET contractor_actor_id = m.actor_id,
    contractor_name_snapshot = COALESCE(w.contractor_name_snapshot, w.contractor)
FROM public.legacy_actor_map m
WHERE m.source_table = 'workorders.contractor'
  AND m.source_key = public.normalize_actor_name(w.contractor)
  AND w.contractor_actor_id IS NULL;

UPDATE public.workorder_investor_splits s
SET actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = s.investor_id
  AND s.actor_id IS NULL;

UPDATE public.stocks s
SET investor_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = s.investor_id
  AND s.investor_actor_id IS NULL;

UPDATE public.supply_movements sm
SET investor_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = sm.investor_id
  AND sm.investor_actor_id IS NULL;

UPDATE public.supply_movements sm
SET provider_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'providers'
  AND m.source_id = sm.provider_id
  AND sm.provider_actor_id IS NULL;

UPDATE public.labors l
SET contractor_actor_id = m.actor_id,
    contractor_name_snapshot = COALESCE(l.contractor_name_snapshot, l.contractor_name)
FROM public.legacy_actor_map m
WHERE m.source_table = 'labors.contractor_name'
  AND m.source_key = public.normalize_actor_name(l.contractor_name)
  AND l.contractor_actor_id IS NULL;

UPDATE public.invoices i
SET investor_actor_id = m.actor_id
FROM public.legacy_actor_map m
WHERE m.source_table = 'investors'
  AND m.source_id = i.investor_id
  AND i.investor_actor_id IS NULL;

UPDATE public.invoices i
SET company_actor_id = m.actor_id,
    company_name_snapshot = COALESCE(i.company_name_snapshot, i.company)
FROM public.legacy_actor_map m
WHERE m.source_table = 'invoices.company'
  AND m.source_key = public.normalize_actor_name(i.company)
  AND i.company_actor_id IS NULL;

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

ALTER TABLE public.projects
    ADD CONSTRAINT fk_projects_customer_actor
    FOREIGN KEY (customer_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.workorders
    ADD CONSTRAINT fk_workorders_investor_actor
    FOREIGN KEY (investor_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID,
    ADD CONSTRAINT fk_workorders_contractor_actor
    FOREIGN KEY (contractor_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.workorder_investor_splits
    ADD CONSTRAINT fk_workorder_investor_splits_actor
    FOREIGN KEY (actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.stocks
    ADD CONSTRAINT fk_stocks_investor_actor
    FOREIGN KEY (investor_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.supply_movements
    ADD CONSTRAINT fk_supply_movements_investor_actor
    FOREIGN KEY (investor_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID,
    ADD CONSTRAINT fk_supply_movements_provider_actor
    FOREIGN KEY (provider_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.labors
    ADD CONSTRAINT fk_labors_contractor_actor
    FOREIGN KEY (contractor_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID;
ALTER TABLE public.invoices
    ADD CONSTRAINT fk_invoices_investor_actor
    FOREIGN KEY (investor_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID,
    ADD CONSTRAINT fk_invoices_company_actor
    FOREIGN KEY (company_actor_id) REFERENCES public.actors(id) ON DELETE RESTRICT NOT VALID;

CREATE INDEX IF NOT EXISTS idx_projects_customer_actor_id ON public.projects (customer_actor_id);
CREATE INDEX IF NOT EXISTS idx_workorders_investor_actor_id ON public.workorders (investor_actor_id);
CREATE INDEX IF NOT EXISTS idx_workorders_contractor_actor_id ON public.workorders (contractor_actor_id);
CREATE INDEX IF NOT EXISTS idx_workorder_splits_actor_id ON public.workorder_investor_splits (actor_id);
CREATE INDEX IF NOT EXISTS idx_stocks_investor_actor_id ON public.stocks (investor_actor_id);
CREATE INDEX IF NOT EXISTS idx_supply_movements_investor_actor_id ON public.supply_movements (investor_actor_id);
CREATE INDEX IF NOT EXISTS idx_supply_movements_provider_actor_id ON public.supply_movements (provider_actor_id);
CREATE INDEX IF NOT EXISTS idx_labors_contractor_actor_id ON public.labors (contractor_actor_id);
CREATE INDEX IF NOT EXISTS idx_invoices_investor_actor_id ON public.invoices (investor_actor_id);
CREATE INDEX IF NOT EXISTS idx_invoices_company_actor_id ON public.invoices (company_actor_id);

-- Vista liviana de control estructural. Las validaciones numericas completas viven
-- en scripts/db/actors_golden_master.sql para ejecutarse contra datos reales.
CREATE OR REPLACE VIEW public.actor_migration_coverage AS
SELECT 'projects.customer_actor_id' AS subject,
       COUNT(*) FILTER (WHERE customer_id IS NOT NULL) AS legacy_count,
       COUNT(*) FILTER (WHERE customer_id IS NOT NULL AND customer_actor_id IS NOT NULL) AS actor_count
FROM public.projects
UNION ALL
SELECT 'workorders.investor_actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL),
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL AND investor_actor_id IS NOT NULL)
FROM public.workorders
UNION ALL
SELECT 'workorder_investor_splits.actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL),
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL AND actor_id IS NOT NULL)
FROM public.workorder_investor_splits
UNION ALL
SELECT 'stocks.investor_actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL),
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL AND investor_actor_id IS NOT NULL)
FROM public.stocks
UNION ALL
SELECT 'supply_movements.investor_actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL),
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL AND investor_actor_id IS NOT NULL)
FROM public.supply_movements
UNION ALL
SELECT 'supply_movements.provider_actor_id',
       COUNT(*) FILTER (WHERE provider_id IS NOT NULL),
       COUNT(*) FILTER (WHERE provider_id IS NOT NULL AND provider_actor_id IS NOT NULL)
FROM public.supply_movements
UNION ALL
SELECT 'invoices.investor_actor_id',
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL),
       COUNT(*) FILTER (WHERE investor_id IS NOT NULL AND investor_actor_id IS NOT NULL)
FROM public.invoices;

COMMIT;
