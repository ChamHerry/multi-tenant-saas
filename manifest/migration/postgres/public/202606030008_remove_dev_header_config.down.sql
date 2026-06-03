INSERT INTO public.system_config (key, value, value_type, description, is_encrypted, created_at, updated_at)
VALUES ('auth.devHeader.enabled', 'false', 'bool', 'Removed legacy dev X-User-ID header auth toggle', false, now(), now())
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    value_type = EXCLUDED.value_type,
    description = EXCLUDED.description,
    is_encrypted = EXCLUDED.is_encrypted,
    updated_at = now();
