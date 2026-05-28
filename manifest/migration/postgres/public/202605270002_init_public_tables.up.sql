CREATE TABLE IF NOT EXISTS public.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(80) NOT NULL UNIQUE,
    plan VARCHAR(32) NOT NULL DEFAULT 'free',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    max_repos INT NOT NULL DEFAULT 5,
    max_symbols INT NOT NULL DEFAULT 100000,
    max_storage_mb INT NOT NULL DEFAULT 1024,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT tenants_slug_chk CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,78}[a-z0-9]$'),
    CONSTRAINT tenants_status_chk CHECK (status IN ('active','suspended','deleted'))
);

CREATE TABLE IF NOT EXISTS public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    display_name VARCHAR(120),
    role VARCHAR(32) NOT NULL DEFAULT 'member',
    auth_provider VARCHAR(32) NOT NULL,
    auth_id VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT users_role_chk CHECK (role IN ('owner','admin','member','viewer')),
    CONSTRAINT users_status_chk CHECK (status IN ('active','disabled','deleted')),
    UNIQUE (tenant_id, email),
    UNIQUE (auth_provider, auth_id)
);

CREATE TABLE IF NOT EXISTS public.api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    name VARCHAR(120) NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    scopes TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS public.git_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    owner_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    provider VARCHAR(32) NOT NULL,
    base_url VARCHAR(255),
    access_token_cipher BYTEA NOT NULL,
    refresh_token_cipher BYTEA,
    token_expires_at TIMESTAMPTZ,
    scopes TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT git_credentials_provider_chk CHECK (provider IN ('github','gitlab','gitea','bitbucket')),
    CONSTRAINT git_credentials_status_chk CHECK (status IN ('active','expired','revoked','deleted'))
);

CREATE TABLE IF NOT EXISTS public.repos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    credential_id UUID REFERENCES public.git_credentials(id) ON DELETE SET NULL,
    provider VARCHAR(32) NOT NULL,
    remote_url VARCHAR(600) NOT NULL,
    clone_url VARCHAR(600),
    web_url VARCHAR(600),
    namespace VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    default_branch VARCHAR(120) NOT NULL DEFAULT 'main',
    last_commit VARCHAR(80),
    analyze_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    webhook_id VARCHAR(160),
    stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_analyzed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT repos_status_chk CHECK (analyze_status IN ('pending','queued','running','completed','failed','disabled')),
    UNIQUE (tenant_id, remote_url)
);

CREATE TABLE IF NOT EXISTS public.repo_branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    branch_name VARCHAR(255) NOT NULL,
    commit_hash VARCHAR(80),
    is_default BOOLEAN NOT NULL DEFAULT false,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (repo_id, branch_name)
);

CREATE TABLE IF NOT EXISTS public.analyze_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    trigger VARCHAR(32) NOT NULL,
    mode VARCHAR(32) NOT NULL DEFAULT 'incremental',
    status VARCHAR(32) NOT NULL DEFAULT 'queued',
    priority INT NOT NULL DEFAULT 100,
    progress SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    worker_id VARCHAR(120),
    from_commit VARCHAR(80),
    to_commit VARCHAR(80),
    changed_files JSONB NOT NULL DEFAULT '[]'::jsonb,
    stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_code VARCHAR(120),
    error_msg TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    locked_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT analyze_jobs_trigger_chk CHECK (trigger IN ('initial','manual','webhook','schedule','retry')),
    CONSTRAINT analyze_jobs_mode_chk CHECK (mode IN ('full','incremental')),
    CONSTRAINT analyze_jobs_status_chk CHECK (status IN ('queued','running','completed','failed','cancelled'))
);

CREATE TABLE IF NOT EXISTS public.webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    repo_id UUID REFERENCES public.repos(id) ON DELETE SET NULL,
    provider VARCHAR(32) NOT NULL,
    event_type VARCHAR(80) NOT NULL,
    delivery_id VARCHAR(255) NOT NULL,
    payload_hash VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'received',
    analyze_job_id UUID REFERENCES public.analyze_jobs(id) ON DELETE SET NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    error_msg TEXT,
    UNIQUE (provider, delivery_id)
);

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
    max_repos INT NOT NULL,
    max_symbols INT NOT NULL,
    max_storage_mb INT NOT NULL,
    max_daily_requests INT NOT NULL DEFAULT 1000,
    max_concurrent_jobs INT NOT NULL DEFAULT 1,
    current_repos INT NOT NULL DEFAULT 0,
    current_symbols BIGINT NOT NULL DEFAULT 0,
    current_storage_mb BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES public.tenants(id) ON DELETE SET NULL,
    user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    action VARCHAR(120) NOT NULL,
    resource_type VARCHAR(80) NOT NULL,
    resource_id VARCHAR(120),
    ip INET,
    user_agent TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
