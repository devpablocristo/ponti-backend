-- ========================================
-- MIGRATION 000190 WORKORDER INVESTOR SPLITS (UP)
-- ========================================
-- Nota: SQL en inglés, comentarios en español

BEGIN;

-- Tabla para soportar "Dividir aporte" sin duplicar workorders.
-- Un workorder puede tener N inversores con un porcentaje asignado.
CREATE TABLE IF NOT EXISTS public.workorder_investor_splits (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    workorder_id bigint NOT NULL,
    investor_id bigint NOT NULL,
    percentage numeric(9,6) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);

-- Evitar duplicados del mismo inversor en la misma OT.
CREATE UNIQUE INDEX IF NOT EXISTS ux_workorder_investor_splits_workorder_investor
    ON public.workorder_investor_splits (workorder_id, investor_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_investor_splits_workorder_id
    ON public.workorder_investor_splits (workorder_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_workorder_investor_splits_investor_id
    ON public.workorder_investor_splits (investor_id)
    WHERE deleted_at IS NULL;

ALTER TABLE public.workorder_investor_splits
    ADD CONSTRAINT fk_workorder_investor_splits_workorder
    FOREIGN KEY (workorder_id) REFERENCES public.workorders(id)
    ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE public.workorder_investor_splits
    ADD CONSTRAINT fk_workorder_investor_splits_investor
    FOREIGN KEY (investor_id) REFERENCES public.investors(id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

COMMIT;

