CREATE INDEX IF NOT EXISTS idx_tenants_status ON public.tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenants_metadata_gin ON public.tenants USING GIN (metadata);

CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON public.users(tenant_id, role) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_email_trgm ON public.users USING GIN (email gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_api_keys_tenant ON public.api_keys(tenant_id) WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_git_credentials_tenant_provider ON public.git_credentials(tenant_id, provider) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_repos_tenant_status ON public.repos(tenant_id, analyze_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_repos_remote_trgm ON public.repos USING GIN (remote_url gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_repos_stats_gin ON public.repos USING GIN (stats);

CREATE INDEX IF NOT EXISTS idx_repo_branches_repo_default ON public.repo_branches(repo_id, is_default);

CREATE INDEX IF NOT EXISTS idx_analyze_jobs_pickup ON public.analyze_jobs(priority, created_at) WHERE status = 'queued';
CREATE INDEX IF NOT EXISTS idx_analyze_jobs_tenant_repo_status ON public.analyze_jobs(tenant_id, repo_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analyze_jobs_running_timeout ON public.analyze_jobs(locked_at) WHERE status = 'running';

CREATE INDEX IF NOT EXISTS idx_webhook_events_repo_time ON public.webhook_events(repo_id, received_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhook_events_payload_gin ON public.webhook_events USING GIN (payload);

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_time ON public.audit_logs(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_time ON public.audit_logs(action, created_at DESC);
