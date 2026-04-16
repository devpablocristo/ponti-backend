ALTER TABLE public.stocks
ADD COLUMN has_real_stock_count boolean NOT NULL DEFAULT false;

UPDATE public.stocks
SET has_real_stock_count = CASE
  WHEN real_stock_units <> 0 THEN true
  ELSE false
END;
