ALTER TABLE public.api_keys DROP CONSTRAINT IF EXISTS api_keys_membership_fk;

ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS key_type VARCHAR(32) NOT NULL DEFAULT 'personal';
ALTER TABLE public.api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_chk;
ALTER TABLE public.api_keys
    ADD CONSTRAINT api_keys_key_type_chk CHECK (key_type = 'personal');

ALTER TABLE public.api_keys ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.api_keys ALTER COLUMN tenant_id DROP NOT NULL;

UPDATE public.api_keys
SET created_by_user_id = COALESCE(created_by_user_id, user_id)
WHERE created_by_user_id IS NULL;

CREATE TABLE IF NOT EXISTS public.api_key_tenant_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL REFERENCES public.api_keys(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    scopes TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    granted_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    revoked_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT api_key_tenant_grants_status_chk CHECK (status IN ('active','revoked'))
);

INSERT INTO public.api_key_tenant_grants(api_key_id, tenant_id, scopes, status, granted_by_user_id, created_at, updated_at)
SELECT
    ak.id,
    ak.tenant_id,
    COALESCE(ak.scopes, ARRAY[]::TEXT[]),
    CASE WHEN ak.revoked_at IS NULL THEN 'active' ELSE 'revoked' END,
    COALESCE(ak.created_by_user_id, ak.user_id),
    ak.created_at,
    now()
FROM public.api_keys ak
WHERE ak.tenant_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1
      FROM public.api_key_tenant_grants g
      WHERE g.api_key_id = ak.id
        AND g.tenant_id = ak.tenant_id
        AND g.revoked_at IS NULL
  );

UPDATE public.api_key_tenant_grants g
SET revoked_at = COALESCE(g.revoked_at, ak.revoked_at),
    revoked_by_user_id = COALESCE(g.revoked_by_user_id, ak.created_by_user_id, ak.user_id),
    updated_at = now()
FROM public.api_keys ak
WHERE g.api_key_id = ak.id
  AND ak.revoked_at IS NOT NULL
  AND g.revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_api_keys_user_active
    ON public.api_keys(user_id, created_at DESC)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix
    ON public.api_keys(key_prefix);
CREATE UNIQUE INDEX IF NOT EXISTS uq_api_key_tenant_grants_active
    ON public.api_key_tenant_grants(api_key_id, tenant_id)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_key_tenant_grants_tenant
    ON public.api_key_tenant_grants(tenant_id, status, created_at DESC)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_key_tenant_grants_api_key
    ON public.api_key_tenant_grants(api_key_id, status);
