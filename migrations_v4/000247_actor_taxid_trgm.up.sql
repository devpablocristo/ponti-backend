BEGIN;

-- Pilar 3 / admin de actores: índice trigram sobre las claves TAX_ID, para permitir
-- búsqueda por CUIT/CUIL parcial (el índice trgm de migr 241 era solo para nombres).
CREATE INDEX IF NOT EXISTS idx_actor_keys_taxid_trgm
  ON public.actor_keys USING gin (key_value gin_trgm_ops)
  WHERE key_type = 'TAX_ID';

COMMIT;
