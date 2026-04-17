ALTER TABLE public.invoices
ADD COLUMN investor_id bigint;

UPDATE public.invoices i
SET investor_id = w.investor_id
FROM public.workorders w
WHERE w.id = i.work_order_id
  AND i.investor_id IS NULL;

ALTER TABLE public.invoices
ALTER COLUMN investor_id SET NOT NULL;

ALTER TABLE ONLY public.invoices
DROP CONSTRAINT IF EXISTS uq_invoices_work_order;

ALTER TABLE ONLY public.invoices
ADD CONSTRAINT fk_invoices_investor
FOREIGN KEY (investor_id) REFERENCES public.investors(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_invoices_investor_id
ON public.invoices (investor_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_invoices_work_order_investor
ON public.invoices (work_order_id, investor_id)
WHERE deleted_at IS NULL;
