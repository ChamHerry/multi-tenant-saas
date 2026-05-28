DROP INDEX IF EXISTS public.idx_api_key_tenant_grants_api_key;
DROP INDEX IF EXISTS public.idx_api_key_tenant_grants_tenant;
DROP INDEX IF EXISTS public.uq_api_key_tenant_grants_active;
DROP INDEX IF EXISTS public.idx_api_keys_prefix;
DROP INDEX IF EXISTS public.idx_api_keys_user_active;

DROP TABLE IF EXISTS public.api_key_tenant_grants;

ALTER TABLE public.api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_chk;
ALTER TABLE public.api_keys DROP COLUMN IF EXISTS key_type;
