-- Restore pre-setup development placeholders only when the keys are absent.
-- Do not overwrite configured/generated secrets.
INSERT INTO public.system_config (key, value, value_type, description, is_encrypted)
VALUES
('auth.session.secret', 'docker-dev-session-secret-change-me', 'string', 'Session HMAC secret (change in prod)', false),
('auth.apiKey.secret', 'docker-dev-apikey-secret-change-me', 'string', 'API Key HMAC secret (change in prod)', false)
ON CONFLICT (key) DO NOTHING;
