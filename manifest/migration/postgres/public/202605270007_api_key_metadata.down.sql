DROP INDEX IF EXISTS public.idx_api_keys_tenant_prefix;
ALTER TABLE public.api_keys DROP COLUMN IF EXISTS created_by_user_id;
ALTER TABLE public.api_keys DROP COLUMN IF EXISTS key_prefix;
