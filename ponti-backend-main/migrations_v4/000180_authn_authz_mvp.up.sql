BEGIN;

ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS idp_sub text,
    ADD COLUMN IF NOT EXISTS idp_email text;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_idp_sub
    ON public.users (idp_sub)
    WHERE idp_sub IS NOT NULL;

CREATE TABLE IF NOT EXISTS public.auth_tenants (
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.auth_roles (
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.auth_permissions (
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS public.auth_role_permissions (
    role_id bigint NOT NULL REFERENCES public.auth_roles(id) ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES public.auth_permissions(id) ON DELETE CASCADE,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS public.auth_memberships (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    tenant_id bigint NOT NULL REFERENCES public.auth_tenants(id) ON DELETE CASCADE,
    role_id bigint NOT NULL REFERENCES public.auth_roles(id),
    status text NOT NULL DEFAULT 'active',
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, tenant_id)
);

INSERT INTO public.auth_roles (name)
VALUES ('admin'), ('manager'), ('viewer')
ON CONFLICT (name) DO NOTHING;

INSERT INTO public.auth_permissions (name)
VALUES ('api.read'), ('api.write')
ON CONFLICT (name) DO NOTHING;

INSERT INTO public.auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.auth_roles r
JOIN public.auth_permissions p ON p.name IN ('api.read', 'api.write')
WHERE r.name IN ('admin', 'manager')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO public.auth_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM public.auth_roles r
JOIN public.auth_permissions p ON p.name = 'api.read'
WHERE r.name = 'viewer'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO public.auth_tenants (name)
VALUES ('default')
ON CONFLICT (name) DO NOTHING;

COMMIT;
