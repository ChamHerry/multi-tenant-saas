CREATE TABLE IF NOT EXISTS public.tenant_invitations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    invitee_email TEXT NOT NULL,
    invitee_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    role VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    token_hash TEXT NOT NULL UNIQUE,
    invited_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    accepted_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    message TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    declined_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    resent_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tenant_invitations_role_chk CHECK (role IN ('owner','admin','member','viewer')),
    CONSTRAINT tenant_invitations_status_chk CHECK (status IN ('pending','accepted','declined','revoked','expired')),
    CONSTRAINT tenant_invitations_email_chk CHECK (position('@' in invitee_email) > 1)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_invitations_pending_email
    ON public.tenant_invitations(tenant_id, lower(invitee_email))
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_tenant_invitations_tenant_status
    ON public.tenant_invitations(tenant_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tenant_invitations_email_status
    ON public.tenant_invitations(lower(invitee_email), status, expires_at);
