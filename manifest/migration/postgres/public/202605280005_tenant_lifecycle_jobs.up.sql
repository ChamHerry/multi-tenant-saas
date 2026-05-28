CREATE TABLE IF NOT EXISTS public.tenant_lifecycle_jobs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    requested_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    scheduled_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error_message TEXT NOT NULL DEFAULT '',
    artifact_uri TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tenant_lifecycle_jobs_type_chk CHECK (type IN ('export','purge','quota_recalculate')),
    CONSTRAINT tenant_lifecycle_jobs_status_chk CHECK (status IN ('pending','running','succeeded','failed','cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_tenant_lifecycle_jobs_pending
    ON public.tenant_lifecycle_jobs(status, scheduled_at)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_tenant_lifecycle_jobs_tenant
    ON public.tenant_lifecycle_jobs(tenant_id, created_at DESC);
