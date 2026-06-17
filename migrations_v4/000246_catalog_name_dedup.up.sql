BEGIN;

-- Pilar 3 — Track B: prevención de duplicados de NOMBRE en catálogos (no-actores).
-- Reusa normalize_name + el patrón del trigger de migr 240. NO toca duplicados existentes
-- (solo bloquea nuevos en INSERT / cambio de nombre / reactivación). Escape hatch
-- app.bypass_name_dedup='on'.
--
-- Alcance (decidido por el esquema real):
--   crops / types / lease_types  -> ya tienen unique per-tenant EXACTO; se sube a NORMALIZADO.
--   categories                   -> el nombre repite legítimamente entre type_id; dedup por
--                                   (tenant, type_id, normalize_name) con función dedicada.
--   business_parameters          -> EXCLUIDO: dedup por `key` (identificador); normalize_name
--                                   colapsaría claves distintas (rate_1 == rate1). Queda su
--                                   unique exacto per-tenant (migr 236).

-- Catálogos planos: trigger genérico (por nombre, per-tenant) de migr 240.
CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.crops       FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_name();
CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.types       FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_name();
CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.lease_types FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_name();

-- categories: dedup de nombre normalizado dentro de (tenant, type_id).
CREATE OR REPLACE FUNCTION public.prevent_duplicate_category_name() RETURNS trigger AS $$
DECLARE
	def_tenant uuid;
	hit        int;
BEGIN
	IF current_setting('app.bypass_name_dedup', true) = 'on' THEN
		RETURN NEW;
	END IF;

	IF TG_OP = 'UPDATE'
		AND NEW.name IS NOT DISTINCT FROM OLD.name
		AND NEW.type_id IS NOT DISTINCT FROM OLD.type_id
		AND NOT (OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL) THEN
		RETURN NEW;
	END IF;

	IF NEW.deleted_at IS NOT NULL THEN
		RETURN NEW;
	END IF;

	SELECT id INTO def_tenant FROM public.auth_tenants WHERE name = 'default' LIMIT 1;

	SELECT 1 INTO hit
	FROM public.categories
	WHERE deleted_at IS NULL
	  AND id <> NEW.id
	  AND type_id IS NOT DISTINCT FROM NEW.type_id
	  AND normalize_name(name) = normalize_name(NEW.name)
	  AND COALESCE(tenant_id, def_tenant) = COALESCE(NEW.tenant_id, def_tenant)
	LIMIT 1;

	IF hit IS NOT NULL THEN
		RAISE EXCEPTION 'duplicate category name (normalized) "%" already exists for this type/tenant', NEW.name
			USING ERRCODE = 'unique_violation';
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.categories FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_category_name();

COMMIT;
