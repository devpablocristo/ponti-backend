-- ========================================
-- MIGRATION 000218 RECONCILE MAIN WITH DEVELOP SCHEMA (UP)
-- ========================================
-- Reproduce en forma idempotente el efecto neto de las migraciones 000204-000213
-- para entornos que ya avanzaron a 000217 sin haberlas ejecutado en orden.

BEGIN;

-- ---------------------------------------------------------------------------
-- 1) invoices por inversor
-- ---------------------------------------------------------------------------

ALTER TABLE public.invoices
    ADD COLUMN IF NOT EXISTS investor_id bigint;

UPDATE public.invoices i
SET investor_id = w.investor_id
FROM public.workorders w
WHERE w.id = i.work_order_id
  AND i.investor_id IS NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.invoices
        WHERE investor_id IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot reconcile invoices.investor_id: null rows remain after backfill';
    END IF;
END $$;

ALTER TABLE public.invoices
    ALTER COLUMN investor_id SET NOT NULL;

ALTER TABLE ONLY public.invoices
    DROP CONSTRAINT IF EXISTS uq_invoices_work_order;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_invoices_investor'
          AND conrelid = 'public.invoices'::regclass
    ) THEN
        ALTER TABLE ONLY public.invoices
            ADD CONSTRAINT fk_invoices_investor
            FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON DELETE RESTRICT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_invoices_investor_id
    ON public.invoices (investor_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_invoices_work_order_investor
    ON public.invoices (work_order_id, investor_id)
    WHERE deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- 2) work order drafts y tablas relacionadas
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS public.work_order_drafts (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    number character varying(100) NOT NULL,
    date date NOT NULL,
    customer_id bigint NOT NULL,
    project_id bigint NOT NULL,
    campaign_id bigint,
    field_id bigint NOT NULL,
    lot_id bigint NOT NULL,
    crop_id bigint NOT NULL,
    labor_id bigint NOT NULL,
    contractor character varying(100) NOT NULL,
    effective_area numeric(18,6) NOT NULL,
    observations text,
    investor_id bigint NOT NULL,
    status character varying(30) NOT NULL DEFAULT 'draft',
    reviewed_by bigint,
    published_work_order_id bigint,
    review_notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    created_by bigint,
    updated_by bigint,
    deleted_by bigint,
    is_digital boolean
);

CREATE TABLE IF NOT EXISTS public.work_order_draft_items (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    draft_id bigint NOT NULL,
    supply_id bigint NOT NULL,
    supply_name character varying(100) NOT NULL,
    total_used numeric(18,6) NOT NULL,
    final_dose numeric(18,6) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS public.work_order_draft_investor_splits (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    draft_id bigint NOT NULL,
    investor_id bigint NOT NULL,
    percentage numeric(9,6) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

ALTER TABLE public.work_order_drafts
    ADD COLUMN IF NOT EXISTS is_digital boolean;

ALTER TABLE public.work_order_draft_items
    ADD COLUMN IF NOT EXISTS supply_name character varying(100);

UPDATE public.work_order_draft_items wodi
SET supply_name = s.name
FROM public.supplies s
WHERE s.id = wodi.supply_id
  AND wodi.supply_name IS NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.work_order_draft_items
        WHERE supply_name IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot reconcile work_order_draft_items.supply_name: null rows remain after backfill';
    END IF;
END $$;

ALTER TABLE public.work_order_draft_items
    ALTER COLUMN supply_name SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_work_order_drafts_project_id
    ON public.work_order_drafts (project_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_work_order_drafts_status
    ON public.work_order_drafts (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_work_order_drafts_is_digital
    ON public.work_order_drafts (is_digital);

CREATE INDEX IF NOT EXISTS idx_work_order_draft_items_draft_id
    ON public.work_order_draft_items (draft_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_work_order_draft_investor_splits_draft_id
    ON public.work_order_draft_investor_splits (draft_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_work_order_draft_investor_splits_draft_investor
    ON public.work_order_draft_investor_splits (draft_id, investor_id)
    WHERE deleted_at IS NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_customer'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_customer
            FOREIGN KEY (customer_id) REFERENCES public.customers(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_project'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_project
            FOREIGN KEY (project_id) REFERENCES public.projects(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_campaign'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_campaign
            FOREIGN KEY (campaign_id) REFERENCES public.campaigns(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_field'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_field
            FOREIGN KEY (field_id) REFERENCES public.fields(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_lot'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_lot
            FOREIGN KEY (lot_id) REFERENCES public.lots(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_crop'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_crop
            FOREIGN KEY (crop_id) REFERENCES public.crops(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_labor'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_labor
            FOREIGN KEY (labor_id) REFERENCES public.labors(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_investor'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_investor
            FOREIGN KEY (investor_id) REFERENCES public.investors(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_drafts_published_work_order'
          AND conrelid = 'public.work_order_drafts'::regclass
    ) THEN
        ALTER TABLE public.work_order_drafts
            ADD CONSTRAINT fk_work_order_drafts_published_work_order
            FOREIGN KEY (published_work_order_id) REFERENCES public.workorders(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_draft_items_draft'
          AND conrelid = 'public.work_order_draft_items'::regclass
    ) THEN
        ALTER TABLE public.work_order_draft_items
            ADD CONSTRAINT fk_work_order_draft_items_draft
            FOREIGN KEY (draft_id) REFERENCES public.work_order_drafts(id)
            ON UPDATE CASCADE ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_draft_items_supply'
          AND conrelid = 'public.work_order_draft_items'::regclass
    ) THEN
        ALTER TABLE public.work_order_draft_items
            ADD CONSTRAINT fk_work_order_draft_items_supply
            FOREIGN KEY (supply_id) REFERENCES public.supplies(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_draft_investor_splits_draft'
          AND conrelid = 'public.work_order_draft_investor_splits'::regclass
    ) THEN
        ALTER TABLE public.work_order_draft_investor_splits
            ADD CONSTRAINT fk_work_order_draft_investor_splits_draft
            FOREIGN KEY (draft_id) REFERENCES public.work_order_drafts(id)
            ON UPDATE CASCADE ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_work_order_draft_investor_splits_investor'
          AND conrelid = 'public.work_order_draft_investor_splits'::regclass
    ) THEN
        ALTER TABLE public.work_order_draft_investor_splits
            ADD CONSTRAINT fk_work_order_draft_investor_splits_investor
            FOREIGN KEY (investor_id) REFERENCES public.investors(id)
            ON UPDATE CASCADE ON DELETE RESTRICT;
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- 3) business insights
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS public.business_insight_candidates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    kind text NOT NULL,
    event_type text NOT NULL,
    entity_type text NOT NULL,
    entity_id text NOT NULL DEFAULT '',
    fingerprint text NOT NULL,
    severity text NOT NULL DEFAULT 'info',
    status text NOT NULL DEFAULT 'new',
    title text NOT NULL,
    body text NOT NULL,
    evidence_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    occurrence_count integer NOT NULL DEFAULT 1,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    first_notified_at timestamptz,
    last_notified_at timestamptz,
    resolved_at timestamptz,
    last_actor text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT business_insight_candidates_fingerprint_uniq UNIQUE (tenant_id, fingerprint)
);

CREATE INDEX IF NOT EXISTS idx_business_insight_candidates_tenant_status
    ON public.business_insight_candidates (tenant_id, status, last_seen_at DESC);

CREATE INDEX IF NOT EXISTS idx_business_insight_candidates_tenant_entity
    ON public.business_insight_candidates (tenant_id, entity_type, entity_id);

CREATE TABLE IF NOT EXISTS public.business_insight_reads (
    insight_id uuid NOT NULL REFERENCES public.business_insight_candidates(id) ON DELETE CASCADE,
    user_id text NOT NULL,
    read_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (insight_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_business_insight_reads_user
    ON public.business_insight_reads (user_id);

-- ---------------------------------------------------------------------------
-- 4) flags y snapshots de insumos
-- ---------------------------------------------------------------------------

ALTER TABLE public.supplies
    ADD COLUMN IF NOT EXISTS is_pending boolean NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_supplies_pending_notdel
    ON public.supplies USING btree (project_id, is_pending, name)
    WHERE deleted_at IS NULL;

ALTER TABLE public.workorder_items
    ADD COLUMN IF NOT EXISTS supply_name character varying(100);

UPDATE public.workorder_items wi
SET supply_name = s.name
FROM public.supplies s
WHERE s.id = wi.supply_id
  AND wi.supply_name IS NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.workorder_items
        WHERE supply_name IS NULL
    ) THEN
        RAISE EXCEPTION 'cannot reconcile workorder_items.supply_name: null rows remain after backfill';
    END IF;
END $$;

ALTER TABLE public.workorder_items
    ALTER COLUMN supply_name SET NOT NULL;

-- ---------------------------------------------------------------------------
-- 5) balance de gestion con columnas *_stock
-- ---------------------------------------------------------------------------

DROP VIEW IF EXISTS v4_report.dashboard_management_balance;
DROP VIEW IF EXISTS v4_report.dashboard_management_balance_field;

CREATE VIEW v4_report.dashboard_management_balance AS
SELECT p.id AS project_id,
    COALESCE(sum(v4_ssot.income_net_total_for_lot(l.id)), 0::numeric) AS income_usd,
    v4_ssot.operating_result_total_for_project(p.id) AS operating_result_usd,
    v4_ssot.renta_pct(v4_ssot.operating_result_total_for_project(p.id), v4_ssot.total_costs_for_project(p.id)) AS operating_result_pct,
    v4_ssot.direct_costs_total_for_project(p.id) AS costos_directos_ejecutados_usd,
    v4_ssot.supply_movements_invested_total_for_project(p.id) + COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0::numeric) AS costos_directos_invertidos_usd,
    v4_ssot.supply_movements_invested_total_for_project(p.id) + COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0::numeric) - v4_ssot.direct_costs_total_for_project(p.id) AS costos_directos_stock_usd,
    COALESCE(sc.semillas_ejecutados_usd, 0::numeric) AS semillas_ejecutados_usd,
    v4_ssot.seeds_invested_for_project_mb(p.id) AS semillas_invertidos_usd,
    v4_ssot.seeds_invested_for_project_mb(p.id) - COALESCE(sc.semillas_ejecutados_usd, 0::numeric) AS semillas_stock_usd,
    COALESCE(sc.agroquimicos_ejecutados_usd, 0::numeric) AS agroquimicos_ejecutados_usd,
    v4_ssot.agrochemicals_invested_for_project_mb(p.id) AS agroquimicos_invertidos_usd,
    v4_ssot.agrochemicals_invested_for_project_mb(p.id) - COALESCE(sc.agroquimicos_ejecutados_usd, 0::numeric) AS agroquimicos_stock_usd,
    COALESCE(sc.fertilizantes_ejecutados_usd, 0::numeric) AS fertilizantes_ejecutados_usd,
    COALESCE(fi.fertilizantes_invertidos_usd, 0::numeric) AS fertilizantes_invertidos_usd,
    COALESCE(fi.fertilizantes_invertidos_usd, 0::numeric) - COALESCE(sc.fertilizantes_ejecutados_usd, 0::numeric) AS fertilizantes_stock_usd,
    COALESCE(sum(v4_ssot.labor_cost_for_lot(l.id)), 0::numeric) AS labores_ejecutados_usd,
    COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0::numeric) AS labores_invertidos_usd,
    COALESCE(sum(v4_ssot.labor_cost_pre_harvest_for_lot(l.id)), 0::numeric) - COALESCE(sum(v4_ssot.labor_cost_for_lot(l.id)), 0::numeric) AS labores_stock_usd,
    v4_ssot.lease_executed_for_project(p.id) AS arriendo_ejecutados_usd,
    v4_ssot.lease_invested_for_project(p.id) AS arriendo_invertidos_usd,
    v4_ssot.lease_invested_for_project(p.id) - v4_ssot.lease_executed_for_project(p.id) AS arriendo_stock_usd,
    v4_ssot.admin_cost_total_for_project(p.id) AS estructura_ejecutados_usd,
    v4_ssot.admin_cost_total_for_project(p.id) AS estructura_invertidos_usd,
    0::numeric AS estructura_stock_usd,
    COALESCE(sc.semillas_ejecutados_usd, 0::numeric) AS semilla_cost,
    COALESCE(sc.agroquimicos_ejecutados_usd, 0::numeric) AS insumos_cost,
    COALESCE(sum(v4_ssot.labor_cost_for_lot(l.id)), 0::numeric) AS labores_cost,
    COALESCE(sc.fertilizantes_ejecutados_usd, 0::numeric) AS fertilizantes_cost
   FROM projects p
     LEFT JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
     LEFT JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
     LEFT JOIN v4_calc.dashboard_supply_costs_by_project sc ON sc.project_id = p.id
     LEFT JOIN v4_calc.dashboard_fertilizers_invested_by_project fi ON fi.project_id = p.id
  WHERE p.deleted_at IS NULL
  GROUP BY p.id, sc.semillas_ejecutados_usd, sc.agroquimicos_ejecutados_usd, sc.fertilizantes_ejecutados_usd, fi.fertilizantes_invertidos_usd;

CREATE VIEW v4_report.dashboard_management_balance_field AS
WITH lots_base AS (
    SELECT p.id AS project_id, p.customer_id, p.campaign_id, f.id AS field_id, l.id AS lot_id, l.hectares
    FROM projects p
    JOIN fields f ON f.project_id = p.id AND f.deleted_at IS NULL
    JOIN lots l ON l.field_id = f.id AND l.deleted_at IS NULL
    WHERE p.deleted_at IS NULL
), income_totals AS (
    SELECT project_id, field_id, sum(v4_ssot.income_net_total_for_lot(lot_id)) AS income_usd
    FROM lots_base GROUP BY project_id, field_id
), direct_costs AS (
    SELECT project_id, field_id, sum(v4_ssot.direct_cost_for_lot(lot_id)) AS direct_costs_usd
    FROM lots_base GROUP BY project_id, field_id
), rent_totals AS (
    SELECT project_id, field_id, sum(v4_ssot.rent_per_ha_for_lot(lot_id) * hectares) AS rent_total_usd
    FROM lots_base GROUP BY project_id, field_id
), admin_totals AS (
    SELECT project_id, field_id, sum(v4_ssot.admin_cost_per_ha_for_lot(lot_id) * hectares) AS admin_total_usd
    FROM lots_base GROUP BY project_id, field_id
), field_hectares AS (
    SELECT project_id, field_id, sum(hectares) AS hectares
    FROM lots_base GROUP BY project_id, field_id
), project_hectares AS (
    SELECT project_id, sum(hectares) AS hectares
    FROM lots_base GROUP BY project_id
), supply_costs_field AS (
    SELECT project_id, field_id,
        sum(semillas_usd) AS semillas_usd,
        sum(total_insumos_usd - semillas_usd - fertilizantes_usd) AS agroquimicos_usd,
        sum(fertilizantes_usd) AS fertilizantes_usd,
        sum(total_insumos_usd) AS total_supply_usd
    FROM v4_calc.field_crop_supply_costs_by_lot GROUP BY project_id, field_id
), supply_costs_project AS (
    SELECT project_id,
        sum(semillas_usd) AS semillas_usd,
        sum(total_insumos_usd - semillas_usd - fertilizantes_usd) AS agroquimicos_usd,
        sum(fertilizantes_usd) AS fertilizantes_usd,
        sum(total_insumos_usd) AS total_supply_usd
    FROM v4_calc.field_crop_supply_costs_by_lot GROUP BY project_id
), labor_costs AS (
    SELECT project_id, field_id,
        sum(v4_ssot.labor_cost_for_lot(lot_id)) AS labor_total_usd,
        sum(v4_ssot.labor_cost_pre_harvest_for_lot(lot_id)) AS labor_pre_harvest_usd
    FROM lots_base GROUP BY project_id, field_id
), project_invested AS (
    SELECT DISTINCT project_id,
        v4_ssot.supply_movements_invested_total_for_project(project_id) AS supply_invested_usd,
        v4_ssot.seeds_invested_for_project_mb(project_id) AS seeds_invested_usd,
        v4_ssot.agrochemicals_invested_for_project_mb(project_id) AS agrochemicals_invested_usd
    FROM lots_base
), project_fertilizers_invested AS (
    SELECT project_id, fertilizantes_invertidos_usd
    FROM v4_calc.dashboard_fertilizers_invested_by_project
)
SELECT lb.project_id, lb.customer_id, lb.campaign_id, lb.field_id,
    COALESCE(it.income_usd, 0::numeric) AS income_usd,
    COALESCE(it.income_usd, 0::numeric) - COALESCE(dc.direct_costs_usd, 0::numeric) - COALESCE(rt.rent_total_usd, 0::numeric) - COALESCE(ad.admin_total_usd, 0::numeric) AS operating_result_usd,
    v4_ssot.renta_pct(COALESCE(it.income_usd, 0::numeric) - COALESCE(dc.direct_costs_usd, 0::numeric) - COALESCE(rt.rent_total_usd, 0::numeric) - COALESCE(ad.admin_total_usd, 0::numeric), COALESCE(dc.direct_costs_usd, 0::numeric) + COALESCE(rt.rent_total_usd, 0::numeric) + COALESCE(ad.admin_total_usd, 0::numeric)) AS operating_result_pct,
    COALESCE(dc.direct_costs_usd, 0::numeric) AS costos_directos_ejecutados_usd,
    COALESCE(pi.supply_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.total_supply_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.total_supply_usd, 0::numeric), scp.total_supply_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END + COALESCE(lc.labor_pre_harvest_usd, 0::numeric) AS costos_directos_invertidos_usd,
    COALESCE(pi.supply_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.total_supply_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.total_supply_usd, 0::numeric), scp.total_supply_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END + COALESCE(lc.labor_pre_harvest_usd, 0::numeric) - COALESCE(dc.direct_costs_usd, 0::numeric) AS costos_directos_stock_usd,
    COALESCE(scf.semillas_usd, 0::numeric) AS semillas_ejecutados_usd,
    COALESCE(pi.seeds_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.semillas_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.semillas_usd, 0::numeric), scp.semillas_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END AS semillas_invertidos_usd,
    COALESCE(pi.seeds_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.semillas_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.semillas_usd, 0::numeric), scp.semillas_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END - COALESCE(scf.semillas_usd, 0::numeric) AS semillas_stock_usd,
    COALESCE(scf.agroquimicos_usd, 0::numeric) AS agroquimicos_ejecutados_usd,
    COALESCE(pi.agrochemicals_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.agroquimicos_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.agroquimicos_usd, 0::numeric), scp.agroquimicos_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END AS agroquimicos_invertidos_usd,
    COALESCE(pi.agrochemicals_invested_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.agroquimicos_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.agroquimicos_usd, 0::numeric), scp.agroquimicos_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END - COALESCE(scf.agroquimicos_usd, 0::numeric) AS agroquimicos_stock_usd,
    COALESCE(scf.fertilizantes_usd, 0::numeric) AS fertilizantes_ejecutados_usd,
    COALESCE(pfi.fertilizantes_invertidos_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.fertilizantes_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.fertilizantes_usd, 0::numeric), scp.fertilizantes_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END AS fertilizantes_invertidos_usd,
    COALESCE(pfi.fertilizantes_invertidos_usd, 0::numeric) *
        CASE WHEN COALESCE(scp.fertilizantes_usd, 0::numeric) > 0::numeric
            THEN v4_core.safe_div(COALESCE(scf.fertilizantes_usd, 0::numeric), scp.fertilizantes_usd)
            ELSE v4_core.safe_div(COALESCE(fh.hectares, 0::numeric), COALESCE(ph.hectares, 0::numeric))
        END - COALESCE(scf.fertilizantes_usd, 0::numeric) AS fertilizantes_stock_usd,
    COALESCE(lc.labor_total_usd, 0::numeric) AS labores_ejecutados_usd,
    COALESCE(lc.labor_pre_harvest_usd, 0::numeric) AS labores_invertidos_usd,
    COALESCE(lc.labor_pre_harvest_usd, 0::numeric) - COALESCE(lc.labor_total_usd, 0::numeric) AS labores_stock_usd,
    COALESCE(rt.rent_total_usd, 0::numeric) AS arriendo_ejecutados_usd,
    COALESCE(rt.rent_total_usd, 0::numeric) AS arriendo_invertidos_usd,
    0::numeric AS arriendo_stock_usd,
    COALESCE(ad.admin_total_usd, 0::numeric) AS estructura_ejecutados_usd,
    COALESCE(ad.admin_total_usd, 0::numeric) AS estructura_invertidos_usd,
    0::numeric AS estructura_stock_usd,
    COALESCE(scf.semillas_usd, 0::numeric) AS semilla_cost,
    COALESCE(scf.agroquimicos_usd, 0::numeric) AS insumos_cost,
    COALESCE(lc.labor_total_usd, 0::numeric) AS labores_cost,
    COALESCE(scf.fertilizantes_usd, 0::numeric) AS fertilizantes_cost
FROM (SELECT DISTINCT project_id, customer_id, campaign_id, field_id FROM lots_base) lb
LEFT JOIN income_totals it ON it.project_id = lb.project_id AND it.field_id = lb.field_id
LEFT JOIN direct_costs dc ON dc.project_id = lb.project_id AND dc.field_id = lb.field_id
LEFT JOIN rent_totals rt ON rt.project_id = lb.project_id AND rt.field_id = lb.field_id
LEFT JOIN admin_totals ad ON ad.project_id = lb.project_id AND ad.field_id = lb.field_id
LEFT JOIN field_hectares fh ON fh.project_id = lb.project_id AND fh.field_id = lb.field_id
LEFT JOIN project_hectares ph ON ph.project_id = lb.project_id
LEFT JOIN supply_costs_field scf ON scf.project_id = lb.project_id AND scf.field_id = lb.field_id
LEFT JOIN supply_costs_project scp ON scp.project_id = lb.project_id
LEFT JOIN labor_costs lc ON lc.project_id = lb.project_id AND lc.field_id = lb.field_id
LEFT JOIN project_invested pi ON pi.project_id = lb.project_id
LEFT JOIN project_fertilizers_invested pfi ON pfi.project_id = lb.project_id;

COMMIT;
