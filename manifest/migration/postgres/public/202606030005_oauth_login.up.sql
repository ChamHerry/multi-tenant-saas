CREATE TABLE IF NOT EXISTS public.oauth_states (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state        VARCHAR(255) NOT NULL UNIQUE,
    provider     VARCHAR(32)  NOT NULL,
    redirect_uri TEXT         NOT NULL DEFAULT '/',
    metadata     JSONB        NOT NULL DEFAULT '{}'::jsonb,
    expires_at   TIMESTAMPTZ  NOT NULL,
    used_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT oauth_states_provider_chk CHECK (provider IN ('github', 'google'))
);

CREATE INDEX IF NOT EXISTS idx_oauth_states_active_state
    ON public.oauth_states(state)
    WHERE used_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_at
    ON public.oauth_states(expires_at)
    WHERE used_at IS NULL;

CREATE TABLE IF NOT EXISTS public.auth_login_challenges (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_hash VARCHAR(128) NOT NULL UNIQUE,
    user_id        UUID         NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    provider       VARCHAR(32)  NOT NULL,
    auth_id        TEXT         NOT NULL,
    login_key      TEXT         NOT NULL DEFAULT '',
    redirect_uri   TEXT         NOT NULL DEFAULT '/',
    metadata       JSONB        NOT NULL DEFAULT '{}'::jsonb,
    expires_at     TIMESTAMPTZ  NOT NULL,
    consumed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT auth_login_challenges_provider_chk CHECK (provider IN ('password','github','google'))
);

CREATE INDEX IF NOT EXISTS idx_auth_login_challenges_user_active
    ON public.auth_login_challenges(user_id)
    WHERE consumed_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_auth_login_challenges_expires_at
    ON public.auth_login_challenges(expires_at)
    WHERE consumed_at IS NULL;

INSERT INTO public.system_config (key, value, value_type, description) VALUES
('oauth.github.enabled',      'false',      'bool',   'Enable GitHub OAuth login'),
('oauth.github.clientId',     '',           'string', 'GitHub OAuth App Client ID'),
('oauth.github.clientSecret', '',           'secret', 'GitHub OAuth App Client Secret'),
('oauth.github.redirectUri',  '',           'string', 'GitHub OAuth callback URI defaults to request host when empty'),
('oauth.github.scopes',       'user:email', 'string', 'GitHub OAuth scopes, space-separated'),
('oauth.github.authUrl',      '',           'string', 'GitHub OAuth authorization endpoint override for tests'),
('oauth.github.tokenUrl',     '',           'string', 'GitHub OAuth token endpoint override for tests'),
('oauth.github.userUrl',      '',           'string', 'GitHub user profile API endpoint override for tests'),
('oauth.github.emailsUrl',    '',           'string', 'GitHub user emails API endpoint override for tests'),
('oauth.google.enabled',      'false',      'bool',   'Enable Google OAuth login'),
('oauth.google.clientId',     '',           'string', 'Google OAuth Client ID'),
('oauth.google.clientSecret', '',           'secret', 'Google OAuth Client Secret'),
('oauth.google.redirectUri',  '',           'string', 'Google OAuth callback URI defaults to request host when empty'),
('oauth.google.scopes',       'openid email profile', 'string', 'Google OAuth scopes, space-separated'),
('oauth.google.authUrl',      '',           'string', 'Google OAuth authorization endpoint override for tests'),
('oauth.google.tokenUrl',     '',           'string', 'Google OAuth token endpoint override for tests'),
('oauth.google.userInfoUrl',  '',           'string', 'Google userinfo API endpoint override for tests'),
('oauth.stateTTL',            '10m',        'string', 'OAuth state time-to-live'),
('oauth.challengeTTL',        '5m',         'string', 'OAuth TOTP challenge time-to-live')
ON CONFLICT (key) DO NOTHING;

ALTER TABLE public.user_identities
    DROP CONSTRAINT IF EXISTS user_identities_provider_chk;

ALTER TABLE public.user_identities
    ADD CONSTRAINT user_identities_provider_chk
    CHECK (provider IN ('github','gitlab','gitea','oidc','password','google'));
