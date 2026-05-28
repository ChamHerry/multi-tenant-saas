-- Re-introduce legacy tenant API Key quota metadata only for rollback. This
-- does not recreate the removed organization API Key HTTP/frontend surface.

ALTER TABLE public.tenant_quotas
  ADD COLUMN IF NOT EXISTS max_api_keys INT;

INSERT INTO public.plan_entitlements(plan, feature_key, enabled, limit_value)
VALUES
    ('free', 'api_key.max_count', true, 5),
    ('pro', 'api_key.max_count', true, 25),
    ('enterprise', 'api_key.max_count', true, NULL)
ON CONFLICT (plan, feature_key) DO NOTHING;

ALTER TABLE public.api_keys DROP CONSTRAINT IF EXISTS api_keys_key_type_chk;
ALTER TABLE public.api_keys
  ADD CONSTRAINT api_keys_key_type_chk CHECK (key_type IN ('personal','tenant_service'));
