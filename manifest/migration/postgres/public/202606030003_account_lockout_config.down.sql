DELETE FROM system_config WHERE key IN (
  'auth.lockout.max_attempts',
  'auth.lockout.window_minutes',
  'auth.lockout.duration_minutes'
);
