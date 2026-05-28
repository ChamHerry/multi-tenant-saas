ALTER TABLE public.api_keys DROP CONSTRAINT IF EXISTS api_keys_membership_fk;

DROP INDEX IF EXISTS public.idx_tenant_memberships_tenant_role;
DROP INDEX IF EXISTS public.idx_tenant_memberships_user_status;
DROP INDEX IF EXISTS public.idx_user_identities_email;
DROP INDEX IF EXISTS public.idx_user_identities_user;
DROP INDEX IF EXISTS public.idx_users_metadata_gin;

DROP TABLE IF EXISTS public.tenant_memberships;
DROP TABLE IF EXISTS public.user_identities;

