DELETE FROM public.system_config WHERE key IN (
  'auth.totp.enabled',
  'auth.totp.issuer',
  'auth.totp.tokenTTL',
  'auth.totp.rateLimitAttempts',
  'auth.totp.rateLimitWindow',
  'auth.totp.backupCodeCount',
  'auth.totp.revokeSessionsOnChange'
);

DROP TABLE IF EXISTS public.user_totp_backup_codes;
DROP TABLE IF EXISTS public.user_totp_configs;
