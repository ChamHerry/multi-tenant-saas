CREATE TABLE IF NOT EXISTS public.user_totp_configs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL UNIQUE REFERENCES public.users(id) ON DELETE CASCADE,
    secret_encrypted TEXT NOT NULL,
    enabled          BOOLEAN NOT NULL DEFAULT FALSE,
    enabled_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_totp_configs_user_enabled
    ON public.user_totp_configs(user_id)
    WHERE enabled = TRUE;

CREATE TABLE IF NOT EXISTS public.user_totp_backup_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    code_hash  TEXT NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_backup_codes_unused
    ON public.user_totp_backup_codes(user_id)
    WHERE used_at IS NULL;

INSERT INTO public.system_config (key, value, value_type, description)
VALUES
  ('auth.totp.enabled', 'true', 'bool', 'TOTP 2FA global feature flag'),
  ('auth.totp.issuer', 'Multi-Tenant-SaaS', 'string', 'Issuer displayed in authenticator apps'),
  ('auth.totp.tokenTTL', '5m', 'string', 'Temporary TOTP login token TTL'),
  ('auth.totp.rateLimitAttempts', '5', 'number', 'Maximum failed TOTP verify attempts per window'),
  ('auth.totp.rateLimitWindow', '5m', 'string', 'TOTP verify attempt rate-limit window'),
  ('auth.totp.backupCodeCount', '10', 'number', 'Number of generated TOTP backup codes'),
  ('auth.totp.revokeSessionsOnChange', 'false', 'bool', 'Revoke all active sessions after TOTP enable/disable/backup-code regeneration')
ON CONFLICT (key) DO NOTHING;
