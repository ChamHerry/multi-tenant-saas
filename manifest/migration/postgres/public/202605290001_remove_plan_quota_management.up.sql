-- Remove package/plan quota management and runtime quota state.
-- Code must no longer read tenants.plan or any plan/quota tables before this migration runs.

DROP TABLE IF EXISTS public.tenant_usage_reservations;
DROP TABLE IF EXISTS public.tenant_usage_counters;
DROP TABLE IF EXISTS public.plan_entitlements;
DROP TABLE IF EXISTS public.tenant_quotas;
DROP TABLE IF EXISTS public.subscriptions;

ALTER TABLE public.tenants
  DROP COLUMN IF EXISTS plan;

UPDATE public.platform_admins
SET role = 'support', updated_at = now()
WHERE role = 'billing_admin';

ALTER TABLE IF EXISTS public.platform_admins DROP CONSTRAINT IF EXISTS platform_admins_role_chk;
ALTER TABLE IF EXISTS public.platform_admins
  ADD CONSTRAINT platform_admins_role_chk CHECK (role IN ('super_admin','support','auditor'));

ALTER TABLE IF EXISTS public.tenant_lifecycle_jobs DROP CONSTRAINT IF EXISTS tenant_lifecycle_jobs_type_chk;
ALTER TABLE IF EXISTS public.tenant_lifecycle_jobs
  ADD CONSTRAINT tenant_lifecycle_jobs_type_chk CHECK (type IN ('export','purge'));
