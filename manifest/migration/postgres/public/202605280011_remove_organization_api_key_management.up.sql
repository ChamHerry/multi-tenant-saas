-- Remove tenant/organization API Key management quota surface while keeping
-- user-bound personal API keys and api_key_tenant_grants as the runtime tenant
-- allowlist for API key requests.

DELETE FROM public.plan_entitlements
WHERE feature_key = 'api_key.max_count';

DELETE FROM public.tenant_usage_counters
WHERE metric = 'api_key.count';

DELETE FROM public.tenant_usage_reservations
WHERE metric = 'api_key.count';

ALTER TABLE public.tenant_quotas
  DROP COLUMN IF EXISTS max_api_keys;

UPDATE public.api_keys
SET key_type = 'personal'
WHERE key_type IS DISTINCT FROM 'personal';

ALTER TABLE public.api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_chk;
ALTER TABLE public.api_keys
  ADD CONSTRAINT api_keys_key_type_chk CHECK (key_type = 'personal');
