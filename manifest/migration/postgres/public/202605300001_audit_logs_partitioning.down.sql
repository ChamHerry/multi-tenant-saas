-- Revert: recreate audit_logs as a regular table and drop partitions.

-- Collect all partition names
DO $$
DECLARE
    part RECORD;
BEGIN
    FOR part IN
        SELECT c.relname
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'public'
          AND c.relkind = 'r'
          AND c.relname LIKE 'audit_logs_2%'
    LOOP
        EXECUTE 'DROP TABLE IF EXISTS public.' || quote_ident(part.relname);
    END LOOP;
END $$;

-- Drop the partitioned table and recreate as regular
DROP TABLE IF EXISTS public.audit_logs;

CREATE TABLE public.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    user_id UUID,
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    ip VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

DROP FUNCTION IF EXISTS public.create_monthly_partitions(INT);
