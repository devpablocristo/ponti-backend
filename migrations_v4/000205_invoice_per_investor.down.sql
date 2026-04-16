DROP INDEX IF EXISTS uq_invoices_work_order_investor;
DROP INDEX IF EXISTS idx_invoices_investor_id;

ALTER TABLE ONLY public.invoices
DROP CONSTRAINT IF EXISTS fk_invoices_investor;

DELETE FROM public.invoices i
USING public.invoices dup
WHERE i.work_order_id = dup.work_order_id
  AND i.id < dup.id
  AND i.deleted_at IS NULL
  AND dup.deleted_at IS NULL;

ALTER TABLE ONLY public.invoices
ADD CONSTRAINT uq_invoices_work_order UNIQUE (work_order_id);

ALTER TABLE public.invoices
DROP COLUMN IF EXISTS investor_id;
