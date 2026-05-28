DROP INDEX IF EXISTS public.idx_users_tenant_role;

ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_tenant_id_email_key;
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_auth_provider_auth_id_key;
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_role_chk;
ALTER TABLE public.users DROP CONSTRAINT IF EXISTS users_tenant_id_fkey;

ALTER TABLE public.users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE public.users DROP COLUMN IF EXISTS role;
ALTER TABLE public.users DROP COLUMN IF EXISTS auth_provider;
ALTER TABLE public.users DROP COLUMN IF EXISTS auth_id;
