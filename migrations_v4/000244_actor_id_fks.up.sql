BEGIN;

-- Pilar 3 — Identity Gate: FK actor_id en los 7 portadores de identidad → actors.
-- Aditivo (columna nullable), ON DELETE SET NULL (la identidad es secundaria al portador;
-- no debe bloquear cascadas de hard-delete), NOT VALID→VALIDATE (patrón migr 237; nacen
-- todas NULL → VALIDATE trivial). El ROL lo da QUÉ columna referencia + la tabla actor_roles.
-- La CUIT/nombre del actor vive en actor_keys (no se duplica en estas tablas).

-- 4 roles con tabla propia
ALTER TABLE public.customers ADD COLUMN IF NOT EXISTS actor_id bigint;
ALTER TABLE public.investors ADD COLUMN IF NOT EXISTS actor_id bigint;
ALTER TABLE public.managers  ADD COLUMN IF NOT EXISTS actor_id bigint;
ALTER TABLE public.providers ADD COLUMN IF NOT EXISTS actor_id bigint;
-- ex-texto-libre: contractor (workorders + labors) y biller (invoices)
ALTER TABLE public.workorders ADD COLUMN IF NOT EXISTS contractor_actor_id bigint;
ALTER TABLE public.labors     ADD COLUMN IF NOT EXISTS contractor_actor_id bigint;
ALTER TABLE public.invoices   ADD COLUMN IF NOT EXISTS biller_actor_id bigint;

ALTER TABLE public.customers  ADD CONSTRAINT fk_customers_actor  FOREIGN KEY (actor_id)            REFERENCES public.actors(id) ON DELETE SET NULL NOT VALID;
ALTER TABLE public.investors  ADD CONSTRAINT fk_investors_actor  FOREIGN KEY (actor_id)            REFERENCES public.actors(id) ON DELETE SET NULL NOT VALID;
ALTER TABLE public.managers   ADD CONSTRAINT fk_managers_actor   FOREIGN KEY (actor_id)            REFERENCES public.actors(id) ON DELETE SET NULL NOT VALID;
ALTER TABLE public.providers  ADD CONSTRAINT fk_providers_actor  FOREIGN KEY (actor_id)            REFERENCES public.actors(id) ON DELETE SET NULL NOT VALID;
ALTER TABLE public.workorders ADD CONSTRAINT fk_workorders_actor FOREIGN KEY (contractor_actor_id) REFERENCES public.actors(id) ON DELETE SET NULL NOT VALID;
ALTER TABLE public.labors     ADD CONSTRAINT fk_labors_actor     FOREIGN KEY (contractor_actor_id) REFERENCES public.actors(id) ON DELETE SET NULL NOT VALID;
ALTER TABLE public.invoices   ADD CONSTRAINT fk_invoices_actor   FOREIGN KEY (biller_actor_id)     REFERENCES public.actors(id) ON DELETE SET NULL NOT VALID;

ALTER TABLE public.customers  VALIDATE CONSTRAINT fk_customers_actor;
ALTER TABLE public.investors  VALIDATE CONSTRAINT fk_investors_actor;
ALTER TABLE public.managers   VALIDATE CONSTRAINT fk_managers_actor;
ALTER TABLE public.providers  VALIDATE CONSTRAINT fk_providers_actor;
ALTER TABLE public.workorders VALIDATE CONSTRAINT fk_workorders_actor;
ALTER TABLE public.labors     VALIDATE CONSTRAINT fk_labors_actor;
ALTER TABLE public.invoices   VALIDATE CONSTRAINT fk_invoices_actor;

CREATE INDEX IF NOT EXISTS idx_customers_actor  ON public.customers  (actor_id);
CREATE INDEX IF NOT EXISTS idx_investors_actor  ON public.investors  (actor_id);
CREATE INDEX IF NOT EXISTS idx_managers_actor   ON public.managers   (actor_id);
CREATE INDEX IF NOT EXISTS idx_providers_actor  ON public.providers  (actor_id);
CREATE INDEX IF NOT EXISTS idx_workorders_actor ON public.workorders (contractor_actor_id);
CREATE INDEX IF NOT EXISTS idx_labors_actor     ON public.labors     (contractor_actor_id);
CREATE INDEX IF NOT EXISTS idx_invoices_actor   ON public.invoices   (biller_actor_id);

COMMIT;
