BEGIN;

-- Ciclo de vida del tenant (soporte Modelo 2: cada cliente de Ponti = un tenant).
-- Aditivo y reversible.
--   status: 'active' | 'suspended' (deja de pagar => suspended).
--   deleted_at: soft-delete (archive/restore/offboard), patrón CRUDAR del repo.
ALTER TABLE public.auth_tenants
	ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'active';

-- CHECK por separado (idempotente) para no fallar si la columna ya existía.
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint WHERE conname = 'auth_tenants_status_check'
	) THEN
		ALTER TABLE public.auth_tenants
			ADD CONSTRAINT auth_tenants_status_check CHECK (status IN ('active','suspended'));
	END IF;
END$$;

ALTER TABLE public.auth_tenants
	ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

CREATE INDEX IF NOT EXISTS idx_auth_tenants_status
	ON public.auth_tenants (status) WHERE deleted_at IS NULL;

COMMIT;
