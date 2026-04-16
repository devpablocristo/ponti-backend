DROP INDEX IF EXISTS public.uq_invoices_work_order_investor;
DROP INDEX IF EXISTS public.idx_invoices_investor_id;

ALTER TABLE ONLY public.invoices
DROP CONSTRAINT IF EXISTS fk_invoices_investor;

ALTER TABLE ONLY public.invoices
ADD CONSTRAINT uq_invoices_work_order UNIQUE (work_order_id);

ALTER TABLE public.invoices
DROP COLUMN IF EXISTS investor_id;