CREATE TABLE IF NOT EXISTS public.user_password_credentials (
    user_id UUID PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    hash_alg VARCHAR(32) NOT NULL DEFAULT 'bcrypt',
    hash_cost INT NOT NULL DEFAULT 12,
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_password_credentials_hash_alg_chk CHECK (hash_alg IN ('bcrypt')),
    CONSTRAINT user_password_credentials_hash_cost_chk CHECK (hash_cost BETWEEN 10 AND 16)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_identities_password_auth_id_lower
    ON public.user_identities (lower(auth_id))
    WHERE provider = 'password';

CREATE TABLE IF NOT EXISTS public.auth_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    secret_hash TEXT NOT NULL,
    csrf_hash TEXT NOT NULL,
    user_agent TEXT,
    ip INET,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    idle_expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    revoke_reason VARCHAR(80),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_sessions_secret_hash
    ON public.auth_sessions(secret_hash);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_active
    ON public.auth_sessions(user_id, last_used_at DESC)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expiry
    ON public.auth_sessions(expires_at)
    WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS public.auth_login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login_key VARCHAR(255) NOT NULL,
    ip INET,
    failed_count INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_failed_at TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (login_key, ip)
);

CREATE INDEX IF NOT EXISTS idx_auth_login_attempts_locked
    ON public.auth_login_attempts(locked_until)
    WHERE locked_until IS NOT NULL;
