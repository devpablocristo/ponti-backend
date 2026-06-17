BEGIN;

-- Pilar 3 (alcance acotado por decisión del usuario): PREVENIR nuevos duplicados de
-- identidad por nombre. NO toca los duplicados existentes (eso es limpieza posterior),
-- NO agrega índice único (fallaría con los dups que ya existen).
--
-- Vector: el unique por-tenant de nombre EXACTO (migr 234/235) no atrapa casi-iguales
-- ("Acme SA" vs "acme sa" vs "Acme  SA"). Acá se normaliza y se bloquea en creación.

-- normalize_name: única fuente de normalización, preparada para ESPAÑOL.
--   lower -> sacar SOLO tildes de vocales + ç (translate explícito, NO unaccent que
--   convertiría ñ->n) -> dejar SOLO letras+números+ñ (quita espacios y puntuación).
-- La ñ se CONSERVA ("Niño" != "Nino", "Peña" != "Pena"). IMMUTABLE (sirve para indexar).
CREATE OR REPLACE FUNCTION public.normalize_name(input text) RETURNS text AS $$
	SELECT regexp_replace(
		translate(
			lower(coalesce(input, '')),
			'áàäâãéèëêíìïîóòöôõúùüûç',
			'aaaaaeeeeiiiiooooouuuuc'
		),
		'[^a-z0-9ñ]', '', 'g'
	);
$$ LANGUAGE sql IMMUTABLE;

-- prevent_duplicate_name: trigger genérico (reusado por todas las tablas vía TG_TABLE_NAME).
-- Rechaza una fila que quede ACTIVA con un normalize_name que ya tenga otra fila activa
-- del MISMO tenant. COALESCE(tenant_id, default) cubre el caso flag-off (tenant_id NULL).
-- Solo chequea en INSERT, cambio de nombre, o reactivación (restore) -> los UPDATE benignos
-- (dual-write de tenant_id, archive, touch de updated_at) sobre filas con dup pre-existente
-- NO se rompen. Escape hatch app.bypass_name_dedup='on' para restores/bulk loads.
CREATE OR REPLACE FUNCTION public.prevent_duplicate_name() RETURNS trigger AS $$
DECLARE
	def_tenant uuid;
	hit        int;
BEGIN
	IF current_setting('app.bypass_name_dedup', true) = 'on' THEN
		RETURN NEW;
	END IF;

	IF TG_OP = 'UPDATE'
		AND NEW.name IS NOT DISTINCT FROM OLD.name
		AND NOT (OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL) THEN
		RETURN NEW; -- update que no cambia nombre ni reactiva: no re-chequear
	END IF;

	-- Solo aplica si la fila resultante queda activa.
	IF NEW.deleted_at IS NOT NULL THEN
		RETURN NEW;
	END IF;

	SELECT id INTO def_tenant FROM public.auth_tenants WHERE name = 'default' LIMIT 1;

	EXECUTE format(
		'SELECT 1 FROM public.%I
		   WHERE deleted_at IS NULL
		     AND id <> $1
		     AND normalize_name(name) = normalize_name($2)
		     AND COALESCE(tenant_id, $3) = COALESCE($4, $3)
		   LIMIT 1',
		TG_TABLE_NAME)
	INTO hit
	USING NEW.id, NEW.name, def_tenant, NEW.tenant_id;

	IF hit IS NOT NULL THEN
		RAISE EXCEPTION 'duplicate name (normalized) "%" already exists in % for this tenant', NEW.name, TG_TABLE_NAME
			USING ERRCODE = 'unique_violation';
	END IF;

	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.customers FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_name();
CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.campaigns FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_name();
CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.managers  FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_name();
CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.investors FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_name();
CREATE TRIGGER trg_prevent_dup_name BEFORE INSERT OR UPDATE ON public.providers FOR EACH ROW EXECUTE FUNCTION public.prevent_duplicate_name();

COMMIT;
