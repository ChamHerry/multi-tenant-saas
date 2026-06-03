CREATE TABLE IF NOT EXISTS public.system_config (
    key           VARCHAR(255)  NOT NULL PRIMARY KEY,
    value         TEXT          NOT NULL DEFAULT '',
    value_type    VARCHAR(32)   NOT NULL DEFAULT 'string',
    description   TEXT          NOT NULL DEFAULT '',
    is_encrypted  BOOLEAN       NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);

COMMENT ON TABLE  public.system_config IS 'Global system configuration';
COMMENT ON COLUMN public.system_config.key IS 'Configuration key, e.g. auth.session.absoluteTTL';
COMMENT ON COLUMN public.system_config.value IS 'Configuration value (encrypted for secret type)';
COMMENT ON COLUMN public.system_config.value_type IS 'Value type: string | number | bool | json | secret';
COMMENT ON COLUMN public.system_config.is_encrypted IS 'Whether value is AES-256-GCM encrypted';

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('auth.password.enabled',               'true',   'bool',   'Password login toggle'),
('auth.password.registrationEnabled',   'true',   'bool',   'Password registration toggle'),
('auth.password.minLength',             '15',     'number', 'Minimum password length'),
('auth.password.bcryptCost',            '12',     'number', 'Bcrypt cost factor (10-16)'),
('auth.password.lockThreshold',         '10',     'number', 'Failed login lock threshold'),
('auth.password.lockDuration',          '15m',   'string',  'Login lock duration')
ON CONFLICT (key) DO NOTHING;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('auth.session.absoluteTTL',            '168h',  'string',  'Session absolute TTL'),
('auth.session.idleTTL',                '12h',   'string',  'Session idle TTL'),
('auth.session.cookie.name',            'saas_template_session', 'string', 'Session cookie name'),
('auth.session.cookie.csrfName',        'saas_template_csrf',   'string', 'CSRF cookie name'),
('auth.session.cookie.path',            '/',     'string',  'Cookie path'),
('auth.session.cookie.domain',          '',      'string',  'Cookie domain (empty = auto)'),
('server.env',                          'local', 'string',  'Server environment (local/test/prod)')
ON CONFLICT (key) DO NOTHING;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('auth.apiKey.maxPersonalKeysPerUser',  '10',    'number',  'Max personal API keys per user')
ON CONFLICT (key) DO NOTHING;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('rateLimit.enabled',                   'true',  'bool',    'Rate limit toggle'),
('rateLimit.rps',                       '20.0',  'number',  'Requests per second'),
('rateLimit.burst',                     '40',    'number',  'Max burst size')
ON CONFLICT (key) DO NOTHING;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('email.smtp.host',                     '',                  'string', 'SMTP server (empty = skip)'),
('email.smtp.port',                     '587',               'number', 'SMTP port'),
('email.smtp.username',                 '',                  'string', 'SMTP username'),
('email.from',                          'noreply@example.com','string', 'From address')
ON CONFLICT (key) DO NOTHING;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('web.baseUrl', 'http://127.0.0.1:5173', 'string', 'Frontend base URL for invitation links')
ON CONFLICT (key) DO NOTHING;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('tenant.lifecycle.purgeDelayHours',    '720',                'number', 'Soft-delete purge delay hours'),
('tenant.lifecycle.exportDir',          '/tmp/tenant-exports', 'string', 'Export directory for purges')
ON CONFLICT (key) DO NOTHING;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('invitation.autoExpireInterval',       '5m',   'string',  'Invitation auto-expire check interval')
ON CONFLICT (key) DO NOTHING;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('automigrate.enabled',                 'true',  'bool',    'Auto-run migrations at startup')
ON CONFLICT (key) DO NOTHING;
