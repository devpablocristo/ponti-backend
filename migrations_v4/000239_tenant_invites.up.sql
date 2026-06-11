BEGIN;

-- U4 (Pilar 2): invitaciones de tenant. Un admin/tenant_owner (o platform-admin)
-- invita un email con un rol; el invitado acepta con el token y obtiene su membership.
-- Solo se guarda el sha256 del token (el token plano se devuelve una vez al crear).
CREATE TABLE IF NOT EXISTS public.tenant_invites (
	id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tenant_id   uuid NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
	email       text NOT NULL,
	role_id     uuid NOT NULL REFERENCES public.auth_roles(id) ON DELETE RESTRICT,
	token_hash  text NOT NULL,
	status      text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'revoked')),
	expires_at  timestamptz NOT NULL,
	created_by  text,        -- idp_sub del que invitó
	accepted_by uuid,        -- users.id del que aceptó
	accepted_at timestamptz,
	created_at  timestamptz NOT NULL DEFAULT now(),
	updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_invites_token_hash ON public.tenant_invites (token_hash);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant ON public.tenant_invites (tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_invites_pending ON public.tenant_invites (tenant_id, email) WHERE status = 'pending';

-- U4: invariante >=1 tenant_owner activo por tenant. Trigger defensivo: bloquea
-- borrar/demotar/desactivar al ÚLTIMO tenant_owner activo de un tenant. Para tenants
-- SIN owner (ej. 'default' hoy) es no-op. No depende de TENANT_ENFORCEMENT (es una
-- regla de integridad de datos).
CREATE OR REPLACE FUNCTION public.enforce_min_one_tenant_owner() RETURNS trigger AS $$
DECLARE
	owner_role_id uuid;
	remaining     int;
BEGIN
	SELECT id INTO owner_role_id FROM public.auth_roles WHERE name = 'tenant_owner';
	IF owner_role_id IS NULL THEN
		RETURN COALESCE(NEW, OLD);
	END IF;

	-- Solo importa si la fila vieja era un owner ACTIVO.
	IF OLD.role_id = owner_role_id AND OLD.status = 'active' THEN
		SELECT count(*) INTO remaining
		FROM public.auth_memberships
		WHERE tenant_id = OLD.tenant_id
		  AND role_id = owner_role_id
		  AND status = 'active'
		  AND user_id <> OLD.user_id;

		IF remaining = 0 AND (
			TG_OP = 'DELETE'
			OR NEW.role_id <> owner_role_id
			OR NEW.status <> 'active'
		) THEN
			RAISE EXCEPTION 'cannot remove the last active tenant_owner of tenant %', OLD.tenant_id
				USING ERRCODE = 'check_violation';
		END IF;
	END IF;

	RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_min_one_tenant_owner ON public.auth_memberships;
CREATE TRIGGER trg_min_one_tenant_owner
	BEFORE UPDATE OR DELETE ON public.auth_memberships
	FOR EACH ROW EXECUTE FUNCTION public.enforce_min_one_tenant_owner();

COMMIT;
