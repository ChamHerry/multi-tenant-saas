CREATE TABLE IF NOT EXISTS public.platform_admins (
    user_id UUID PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT platform_admins_role_chk CHECK (role IN ('super_admin','support','billing_admin','auditor')),
    CONSTRAINT platform_admins_status_chk CHECK (status IN ('active','suspended'))
);

CREATE INDEX IF NOT EXISTS idx_platform_admins_status
    ON public.platform_admins(status, role);
