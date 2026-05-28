ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS key_prefix VARCHAR(32);
ALTER TABLE public.api_keys ADD COLUMN IF NOT EXISTS created_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_prefix ON public.api_keys(tenant_id, key_prefix) WHERE revoked_at IS NULL;
