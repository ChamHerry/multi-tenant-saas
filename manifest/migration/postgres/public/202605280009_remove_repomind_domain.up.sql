-- Remove RepoMind/code-analysis product domain from the reusable multi-tenant SaaS template.
-- Runtime code must no longer reference these tables before this migration runs.

DELETE FROM public.plan_entitlements
WHERE feature_key IN ('repo.max_count', 'symbol.max_count', 'storage.max_mb');

DELETE FROM public.tenant_usage_counters
WHERE metric IN ('repo.count', 'symbol.count', 'storage.mb');

DELETE FROM public.tenant_usage_reservations
WHERE metric IN ('repo.count', 'symbol.count', 'storage.mb');

ALTER TABLE public.tenants
  DROP COLUMN IF EXISTS max_repos,
  DROP COLUMN IF EXISTS max_symbols,
  DROP COLUMN IF EXISTS max_storage_mb;

ALTER TABLE public.tenant_quotas
  DROP COLUMN IF EXISTS max_repos,
  DROP COLUMN IF EXISTS max_symbols,
  DROP COLUMN IF EXISTS max_storage_mb,
  DROP COLUMN IF EXISTS current_repos,
  DROP COLUMN IF EXISTS current_symbols,
  DROP COLUMN IF EXISTS current_storage_mb;

ALTER TABLE public.tenant_quotas
  ADD COLUMN IF NOT EXISTS max_members INT,
  ADD COLUMN IF NOT EXISTS max_api_keys INT;

DROP TABLE IF EXISTS public.analysis_snapshots;
DROP TABLE IF EXISTS public.process_steps;
DROP TABLE IF EXISTS public.processes;
DROP TABLE IF EXISTS public.symbol_communities;
DROP TABLE IF EXISTS public.communities;
DROP TABLE IF EXISTS public.symbol_search;
DROP TABLE IF EXISTS public.relations;
DROP TABLE IF EXISTS public.symbols;
DROP TABLE IF EXISTS public.files;
DROP TABLE IF EXISTS public.repo_access_grants;
DROP TABLE IF EXISTS public.tenant_repo_subscriptions;
DROP TABLE IF EXISTS public.webhook_events;
DROP TABLE IF EXISTS public.analyze_jobs;
DROP TABLE IF EXISTS public.repo_branches;
DROP TABLE IF EXISTS public.repos;
DROP TABLE IF EXISTS public.git_credentials;
