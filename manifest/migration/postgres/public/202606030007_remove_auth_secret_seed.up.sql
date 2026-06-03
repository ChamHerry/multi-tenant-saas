-- Remove unsafe development secret seeds now that first-run setup generates encrypted secrets.
-- This only deletes legacy unencrypted change-me placeholders and never removes
-- setup-generated encrypted secret values.
DELETE FROM public.system_config
WHERE key IN ('auth.session.secret', 'auth.apiKey.secret')
  AND is_encrypted = false
  AND value ILIKE '%change-me%';
