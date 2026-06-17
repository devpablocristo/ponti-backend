BEGIN;

-- Pilar 3 — Identity Gate (PARTE V del plan): registro central de identidad.
-- 1 fila por ENTE real (persona/empresa) por tenant; los roles son ATRIBUTO (actor_roles);
-- la unicidad la garantiza el ÍNDICE (no un trigger). Aditivo: NO toca tablas existentes.
-- Track A (actores). Las claves duras (CUIT) + nombre legal viven en actor_keys.

CREATE TABLE public.actors (
	id           bigserial PRIMARY KEY,
	tenant_id    uuid NULL,                                   -- el resolver llena un tenant concreto (OrgID o 'default')
	party_type   text NOT NULL DEFAULT 'unknown' CHECK (party_type IN ('org', 'person', 'unknown')),
	display_name text NOT NULL,                                -- format_proper_name (presentación)
	raw_name     text NOT NULL,                                -- lo que entró
	status       text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
	created_at   timestamptz NOT NULL DEFAULT now(),
	updated_at   timestamptz NOT NULL DEFAULT now(),
	deleted_at   timestamptz NULL,
	created_by   text NULL,
	updated_by   text NULL
);
CREATE INDEX idx_actors_tenant ON public.actors (tenant_id);

-- roles que juega cada actor (ATRIBUTO, no partición) → cross-rol unificado
CREATE TABLE public.actor_roles (
	actor_id   bigint NOT NULL REFERENCES public.actors(id) ON DELETE CASCADE,
	role       text NOT NULL CHECK (role IN ('customer', 'provider', 'investor', 'manager', 'contractor', 'biller', 'lessee')),
	created_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (actor_id, role)
);

-- claves deduplicantes (la unicidad real vive acá, no en un trigger)
CREATE TABLE public.actor_keys (
	id         bigserial PRIMARY KEY,
	actor_id   bigint NOT NULL REFERENCES public.actors(id) ON DELETE CASCADE,
	tenant_id  uuid NULL,
	key_type   text NOT NULL CHECK (key_type IN ('TAX_ID', 'LEGAL_NAME', 'PERSON_NAME', 'ALIAS')),
	key_value  text NOT NULL,
	active     boolean NOT NULL DEFAULT true,
	source     text NOT NULL DEFAULT 'direct' CHECK (source IN ('direct', 'import', 'backfill')),
	created_at timestamptz NOT NULL DEFAULT now()
);

-- LA garantía: unicidad por (tenant, key) SIN rol → mismo CUIT/nombre = UNA identidad
-- (cross-rol unificado). Solo claves ACTIVAS → no obliga a limpiar históricos. El resolver
-- llena tenant_id con un tenant CONCRETO (OrgID o el 'default'), así no hace falta COALESCE
-- ni hardcodear el uuid del default (portable entre entornos).
CREATE UNIQUE INDEX uq_actor_keys_active ON public.actor_keys (tenant_id, key_type, key_value) WHERE active;
CREATE INDEX idx_actor_keys_actor ON public.actor_keys (actor_id);
CREATE INDEX idx_actor_keys_trgm ON public.actor_keys USING gin (key_value gin_trgm_ops)
	WHERE key_type IN ('LEGAL_NAME', 'PERSON_NAME', 'ALIAS');

COMMIT;
