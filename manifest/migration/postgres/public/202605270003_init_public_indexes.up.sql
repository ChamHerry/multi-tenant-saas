CREATE INDEX IF NOT EXISTS idx_tenants_status ON public.tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_metadata_gin ON public.tenants USING GIN (metadata);

CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON public.users(tenant_id, role) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_email ON public.users(email);

CREATE INDEX IF NOT EXISTS idx_api_keys_tenant ON public.api_keys(tenant_id) WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_time ON public.audit_logs(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_time ON public.audit_logs(action, created_at DESC);
