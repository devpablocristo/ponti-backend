-- ========================================
-- MIGRATION 000199 WORK ORDER DRAFTS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

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
    deleted_by bigint
);

CREATE TABLE IF NOT EXISTS public.work_order_draft_items (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    draft_id bigint NOT NULL,
    supply_id bigint NOT NULL,
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

CREATE INDEX IF NOT EXISTS idx_work_order_drafts_project_id
    ON public.work_order_drafts (project_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_work_order_drafts_status
    ON public.work_order_drafts (status)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_work_order_draft_items_draft_id
    ON public.work_order_draft_items (draft_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_work_order_draft_investor_splits_draft_id
    ON public.work_order_draft_investor_splits (draft_id)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_work_order_draft_investor_splits_draft_investor
    ON public.work_order_draft_investor_splits (draft_id, investor_id)
    WHERE deleted_at IS NULL;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_customer
    FOREIGN KEY (customer_id) REFERENCES public.customers(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_project
    FOREIGN KEY (project_id) REFERENCES public.projects(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_campaign
    FOREIGN KEY (campaign_id) REFERENCES public.campaigns(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_field
    FOREIGN KEY (field_id) REFERENCES public.fields(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_lot
    FOREIGN KEY (lot_id) REFERENCES public.lots(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_crop
    FOREIGN KEY (crop_id) REFERENCES public.crops(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_labor
    FOREIGN KEY (labor_id) REFERENCES public.labors(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_investor
    FOREIGN KEY (investor_id) REFERENCES public.investors(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_drafts
    ADD CONSTRAINT fk_work_order_drafts_published_work_order
    FOREIGN KEY (published_work_order_id) REFERENCES public.workorders(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_draft_items
    ADD CONSTRAINT fk_work_order_draft_items_draft
    FOREIGN KEY (draft_id) REFERENCES public.work_order_drafts(id)
    ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE public.work_order_draft_items
    ADD CONSTRAINT fk_work_order_draft_items_supply
    FOREIGN KEY (supply_id) REFERENCES public.supplies(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE public.work_order_draft_investor_splits
    ADD CONSTRAINT fk_work_order_draft_investor_splits_draft
    FOREIGN KEY (draft_id) REFERENCES public.work_order_drafts(id)
    ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE public.work_order_draft_investor_splits
    ADD CONSTRAINT fk_work_order_draft_investor_splits_investor
    FOREIGN KEY (investor_id) REFERENCES public.investors(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

COMMIT;
