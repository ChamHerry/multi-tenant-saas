CREATE INDEX IF NOT EXISTS idx_audit_logs_user_time
    ON public.audit_logs(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_action_time
    ON public.audit_logs(tenant_id, action, created_at DESC);
