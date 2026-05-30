-- Convert audit_logs to a range-partitioned table by month.
-- Monthly partitions are created inline (not via PL/pgSQL function) to avoid
-- golang-migrate splitting on semicolons inside $$ dollar-quoted function bodies.

-- 1. Rename existing table
ALTER TABLE IF EXISTS public.audit_logs RENAME TO audit_logs_legacy;

-- 2. Create partitioned table with same schema
CREATE TABLE IF NOT EXISTS public.audit_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    tenant_id UUID,
    user_id UUID,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    ip VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
) PARTITION BY RANGE (created_at);

-- 3. Create default partition
CREATE TABLE IF NOT EXISTS public.audit_logs_default PARTITION OF public.audit_logs DEFAULT;

-- 4. Copy data from legacy table (if any)
INSERT INTO public.audit_logs SELECT * FROM public.audit_logs_legacy ON CONFLICT DO NOTHING;

-- 5. Drop legacy table
DROP TABLE IF EXISTS public.audit_logs_legacy;

-- 6. Detach default, create monthly partitions, migrate data, reattach default.
ALTER TABLE public.audit_logs DETACH PARTITION public.audit_logs_default;

CREATE TABLE IF NOT EXISTS public.audit_logs_202605 PARTITION OF public.audit_logs FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE IF NOT EXISTS public.audit_logs_202606 PARTITION OF public.audit_logs FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE IF NOT EXISTS public.audit_logs_202607 PARTITION OF public.audit_logs FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE IF NOT EXISTS public.audit_logs_202608 PARTITION OF public.audit_logs FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');

INSERT INTO public.audit_logs SELECT * FROM public.audit_logs_default ON CONFLICT DO NOTHING;
DROP TABLE IF EXISTS public.audit_logs_default;
CREATE TABLE IF NOT EXISTS public.audit_logs_default PARTITION OF public.audit_logs DEFAULT;
