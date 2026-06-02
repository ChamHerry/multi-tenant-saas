INSERT INTO system_config (key, value, value_type, description)
VALUES
  ('auth.lockout.max_attempts', '5', 'number', 'Maximum failed login attempts within window before account lockout'),
  ('auth.lockout.window_minutes', '15', 'number', 'Sliding window in minutes for counting failed login attempts'),
  ('auth.lockout.duration_minutes', '15', 'number', 'Duration in minutes to lock the account after exceeding max_attempts')
ON CONFLICT (key) DO NOTHING;
