ALTER TABLE public.users ADD COLUMN IF NOT EXISTS avatar_url TEXT;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE public.users ALTER COLUMN tenant_id DROP NOT NULL;
ALTER TABLE public.users ALTER COLUMN role DROP NOT NULL;
ALTER TABLE public.users ALTER COLUMN auth_provider DROP NOT NULL;
ALTER TABLE public.users ALTER COLUMN auth_id DROP NOT NULL;

ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_tenant_id_email_key;

CREATE TABLE IF NOT EXISTS public.user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    provider VARCHAR(32) NOT NULL,
    auth_id VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    email_verified BOOLEAN NOT NULL DEFAULT false,
    raw_profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, auth_id),
    CONSTRAINT user_identities_provider_chk CHECK (provider IN ('github','gitlab','gitea','oidc','password'))
);

CREATE TABLE IF NOT EXISTS public.tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL DEFAULT 'member',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    invited_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    joined_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, user_id),
    CONSTRAINT tenant_memberships_role_chk CHECK (role IN ('owner','admin','member','viewer')),
    CONSTRAINT tenant_memberships_status_chk CHECK (status IN ('invited','active','suspended','removed'))
);

INSERT INTO public.user_identities(user_id, provider, auth_id, email, last_login_at, created_at, updated_at)
SELECT id, auth_provider, auth_id, email, last_login_at, created_at, updated_at
FROM public.users
WHERE auth_provider IS NOT NULL AND auth_id IS NOT NULL
ON CONFLICT (provider, auth_id) DO NOTHING;

INSERT INTO public.tenant_memberships(tenant_id, user_id, role, status, joined_at, created_at, updated_at)
SELECT
    tenant_id,
    id,
    COALESCE(role, 'member'),
    CASE status
        WHEN 'active' THEN 'active'
        WHEN 'disabled' THEN 'suspended'
        ELSE 'removed'
    END,
    created_at,
    created_at,
    updated_at
FROM public.users
WHERE tenant_id IS NOT NULL
ON CONFLICT (tenant_id, user_id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_users_metadata_gin ON public.users USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_user_identities_user ON public.user_identities(user_id);
CREATE INDEX IF NOT EXISTS idx_user_identities_email ON public.user_identities(email);
CREATE INDEX IF NOT EXISTS idx_tenant_memberships_user_status
    ON public.tenant_memberships(user_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenant_memberships_tenant_role
    ON public.tenant_memberships(tenant_id, role)
    WHERE deleted_at IS NULL;

ALTER TABLE public.api_keys DROP CONSTRAINT IF EXISTS api_keys_membership_fk;
ALTER TABLE public.api_keys
ADD CONSTRAINT api_keys_membership_fk
FOREIGN KEY (tenant_id, user_id)
REFERENCES public.tenant_memberships(tenant_id, user_id)
NOT VALID;
