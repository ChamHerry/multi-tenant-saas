-- Technical rollback for the removed package/plan quota schema. This recreates
-- structure needed by older code; it does not restore historical tenant data.

ALTER TABLE public.tenants
  ADD COLUMN IF NOT EXISTS plan VARCHAR(32) NOT NULL DEFAULT 'free';

CREATE TABLE IF NOT EXISTS public.subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    plan VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (tenant_id, status) DEFERRABLE INITIALLY IMMEDIATE
);

CREATE TABLE IF NOT EXISTS public.tenant_quotas (
    tenant_id UUID PRIMARY KEY REFERENCES public.tenants(id) ON DELETE CASCADE,
    max_daily_requests INT NOT NULL DEFAULT 1000,
    max_concurrent_jobs INT NOT NULL DEFAULT 1,
    max_members INT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.plan_entitlements (
    plan VARCHAR(32) NOT NULL,
    feature_key VARCHAR(120) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    limit_value BIGINT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(plan, feature_key)
);

CREATE TABLE IF NOT EXISTS public.tenant_usage_counters (
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    metric VARCHAR(80) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    period_end TIMESTAMPTZ NOT NULL DEFAULT '2999-12-31 00:00:00+00',
    used BIGINT NOT NULL DEFAULT 0,
    reserved BIGINT NOT NULL DEFAULT 0,
    limit_snapshot BIGINT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(tenant_id, metric, period_start),
    CONSTRAINT tenant_usage_counters_non_negative_chk CHECK (used >= 0 AND reserved >= 0)
);

CREATE TABLE IF NOT EXISTS public.tenant_usage_reservations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    metric VARCHAR(80) NOT NULL,
    delta BIGINT NOT NULL CHECK (delta > 0),
    status VARCHAR(32) NOT NULL DEFAULT 'reserved',
    resource_type VARCHAR(80) NOT NULL,
    resource_id TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tenant_usage_reservations_status_chk CHECK (status IN ('reserved','committed','released','expired'))
);

CREATE INDEX IF NOT EXISTS idx_tenant_usage_reservations_tenant_status
    ON public.tenant_usage_reservations(tenant_id, status, expires_at);

INSERT INTO public.plan_entitlements(plan, feature_key, enabled, limit_value)
VALUES
    ('free', 'member.max_count', true, 5),
    ('free', 'audit.retention_days', true, 30),
    ('pro', 'member.max_count', true, 25),
    ('pro', 'audit.retention_days', true, 180),
    ('enterprise', 'member.max_count', true, NULL),
    ('enterprise', 'audit.retention_days', true, 365)
ON CONFLICT (plan, feature_key) DO NOTHING;

ALTER TABLE IF EXISTS public.platform_admins DROP CONSTRAINT IF EXISTS platform_admins_role_chk;
ALTER TABLE IF EXISTS public.platform_admins
  ADD CONSTRAINT platform_admins_role_chk CHECK (role IN ('super_admin','support','billing_admin','auditor'));

ALTER TABLE IF EXISTS public.tenant_lifecycle_jobs DROP CONSTRAINT IF EXISTS tenant_lifecycle_jobs_type_chk;
ALTER TABLE IF EXISTS public.tenant_lifecycle_jobs
  ADD CONSTRAINT tenant_lifecycle_jobs_type_chk CHECK (type IN ('export','purge','quota_recalculate'));
