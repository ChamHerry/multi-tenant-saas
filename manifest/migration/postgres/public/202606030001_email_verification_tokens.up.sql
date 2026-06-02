CREATE TABLE IF NOT EXISTS public.email_verification_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    email       VARCHAR(255) NOT NULL,
    token_hash  VARCHAR(128) NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_evt_token_hash ON public.email_verification_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_evt_user_id ON public.email_verification_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_evt_email ON public.email_verification_tokens(email);
CREATE INDEX IF NOT EXISTS idx_evt_expires_at ON public.email_verification_tokens(expires_at)
    WHERE used_at IS NULL;
