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
    ('free', 'repo.max_count', true, 5),
    ('free', 'symbol.max_count', true, 100000),
    ('free', 'storage.max_mb', true, 1024),
    ('free', 'api_key.max_count', true, 5),
    ('free', 'member.max_count', true, 5),
    ('free', 'audit.retention_days', true, 30),
    ('pro', 'repo.max_count', true, 50),
    ('pro', 'symbol.max_count', true, 1000000),
    ('pro', 'storage.max_mb', true, 10240),
    ('pro', 'api_key.max_count', true, 25),
    ('pro', 'member.max_count', true, 25),
    ('pro', 'audit.retention_days', true, 180),
    ('enterprise', 'repo.max_count', true, NULL),
    ('enterprise', 'symbol.max_count', true, NULL),
    ('enterprise', 'storage.max_mb', true, NULL),
    ('enterprise', 'api_key.max_count', true, NULL),
    ('enterprise', 'member.max_count', true, NULL),
    ('enterprise', 'audit.retention_days', true, 365)
ON CONFLICT (plan, feature_key) DO NOTHING;
