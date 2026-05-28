-- Move RepoMind from per-tenant runtime schemas to shared public repository/index tables.
-- The table name public.repos is kept to avoid a parallel repositories_v2 system;
-- from this migration forward it represents a global repository entity.

REVOKE CREATE ON SCHEMA public FROM PUBLIC;

ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS tenants_schema_name_chk;
ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS tenants_graph_name_chk;
ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS tenants_schema_name_key;
ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS tenants_graph_name_key;
ALTER TABLE public.tenants DROP COLUMN IF EXISTS schema_name;
ALTER TABLE public.tenants DROP COLUMN IF EXISTS graph_name;
ALTER TABLE public.tenants DROP COLUMN IF EXISTS tenant_schema_version;
ALTER TABLE public.tenants DROP COLUMN IF EXISTS tenant_schema_dirty;
ALTER TABLE public.tenants DROP COLUMN IF EXISTS tenant_schema_checked_at;

ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS code_host_url VARCHAR(255) NOT NULL DEFAULT 'https://github.com';
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS external_id VARCHAR(160);
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS external_node_id VARCHAR(255);
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS owner_name VARCHAR(255);
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS full_name VARCHAR(512);
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS normalized_remote_key VARCHAR(700);
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS visibility VARCHAR(32) NOT NULL DEFAULT 'unknown';
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS is_fork BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS indexed_commit_hash VARCHAR(80);
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS permission_synced_at TIMESTAMPTZ;
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE public.repos
SET
    code_host_url = COALESCE(NULLIF(code_host_url, ''), 'https://github.com'),
    owner_name = COALESCE(NULLIF(owner_name, ''), NULLIF(namespace, '')),
    full_name = COALESCE(
        NULLIF(full_name, ''),
        CASE
            WHEN NULLIF(namespace, '') IS NOT NULL THEN namespace || '/' || name
            ELSE name
        END
    ),
    normalized_remote_key = COALESCE(
        NULLIF(normalized_remote_key, ''),
        lower(
            provider || ':' || COALESCE(NULLIF(code_host_url, ''), 'https://github.com') || ':' ||
            regexp_replace(
                regexp_replace(
                    regexp_replace(COALESCE(NULLIF(web_url, ''), NULLIF(remote_url, ''), name), '^git@([^:]+):', 'https://\1/'),
                    '^ssh://git@([^/]+)/', 'https://\1/'
                ),
                '\.git$', ''
            )
        )
    ),
    visibility = COALESCE(NULLIF(visibility, ''), 'unknown'),
    metadata = COALESCE(metadata, '{}'::jsonb);

ALTER TABLE public.repos ALTER COLUMN full_name SET NOT NULL;
ALTER TABLE public.repos ALTER COLUMN normalized_remote_key SET NOT NULL;
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_provider_chk;
ALTER TABLE public.repos ADD CONSTRAINT repos_provider_chk CHECK (provider IN ('github','gitlab','gitea','bitbucket','azuredevops','other'));
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_visibility_chk;
ALTER TABLE public.repos ADD CONSTRAINT repos_visibility_chk CHECK (visibility IN ('public','private','internal','unknown'));

CREATE TEMP TABLE repo_canonical_map AS
WITH ranked AS (
    SELECT
        id AS old_repo_id,
        FIRST_VALUE(id) OVER (
            PARTITION BY provider, code_host_url, normalized_remote_key
            ORDER BY created_at ASC, id::text ASC
        ) AS canonical_repo_id
    FROM public.repos
    WHERE deleted_at IS NULL
)
SELECT old_repo_id, canonical_repo_id
FROM ranked;

DELETE FROM public.repo_branches rb
USING repo_canonical_map m, public.repo_branches keep
WHERE m.old_repo_id <> m.canonical_repo_id
  AND rb.repo_id = m.old_repo_id
  AND keep.repo_id = m.canonical_repo_id
  AND keep.branch_name = rb.branch_name;

UPDATE public.repo_branches rb
SET repo_id = m.canonical_repo_id
FROM repo_canonical_map m
WHERE rb.repo_id = m.old_repo_id
  AND m.old_repo_id <> m.canonical_repo_id;

UPDATE public.analyze_jobs aj
SET repo_id = m.canonical_repo_id
FROM repo_canonical_map m
WHERE aj.repo_id = m.old_repo_id
  AND m.old_repo_id <> m.canonical_repo_id;

UPDATE public.webhook_events we
SET repo_id = m.canonical_repo_id
FROM repo_canonical_map m
WHERE we.repo_id = m.old_repo_id
  AND m.old_repo_id <> m.canonical_repo_id;

CREATE TABLE IF NOT EXISTS public.tenant_repo_subscriptions (
    tenant_id UUID NOT NULL REFERENCES public.tenants(id) ON DELETE CASCADE,
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    added_by_user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    credential_id UUID REFERENCES public.git_credentials(id) ON DELETE SET NULL,
    purpose VARCHAR(32) NOT NULL DEFAULT 'workspace',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (tenant_id, repo_id),
    CONSTRAINT tenant_repo_subscriptions_purpose_chk CHECK (purpose IN ('workspace','favorite','private_import','watch','migration')),
    CONSTRAINT tenant_repo_subscriptions_status_chk CHECK (status IN ('active','disabled','deleted'))
);

INSERT INTO public.tenant_repo_subscriptions(tenant_id, repo_id, credential_id, purpose, status, created_at, updated_at)
SELECT DISTINCT ON (r.tenant_id, m.canonical_repo_id)
    r.tenant_id,
    m.canonical_repo_id,
    r.credential_id,
    'migration',
    CASE WHEN r.deleted_at IS NULL THEN 'active' ELSE 'deleted' END,
    r.created_at,
    now()
FROM public.repos r
JOIN repo_canonical_map m ON m.old_repo_id = r.id
WHERE r.tenant_id IS NOT NULL
ON CONFLICT (tenant_id, repo_id) DO UPDATE
SET credential_id = COALESCE(EXCLUDED.credential_id, public.tenant_repo_subscriptions.credential_id),
    status = EXCLUDED.status,
    updated_at = now();

CREATE TABLE IF NOT EXISTS public.repo_access_grants (
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    subject_type VARCHAR(32) NOT NULL,
    subject_id UUID NOT NULL,
    permission VARCHAR(32) NOT NULL,
    source VARCHAR(64) NOT NULL,
    external_account_id VARCHAR(255),
    synced_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (repo_id, subject_type, subject_id, source),
    CONSTRAINT repo_access_grants_subject_chk CHECK (subject_type IN ('user','tenant','code_host_account','api_key','platform_role')),
    CONSTRAINT repo_access_grants_permission_chk CHECK (permission IN ('read','write','admin'))
);

INSERT INTO public.repo_access_grants(repo_id, subject_type, subject_id, permission, source, synced_at, created_at, updated_at)
SELECT repo_id, 'tenant', tenant_id, 'read', 'migration_legacy_subscription', now(), now(), now()
FROM public.tenant_repo_subscriptions
WHERE deleted_at IS NULL AND status = 'active'
ON CONFLICT (repo_id, subject_type, subject_id, source) DO UPDATE
SET permission = EXCLUDED.permission,
    synced_at = now(),
    updated_at = now();

DELETE FROM public.repos r
USING repo_canonical_map m
WHERE r.id = m.old_repo_id
  AND m.old_repo_id <> m.canonical_repo_id;

DROP INDEX IF EXISTS public.idx_repos_tenant_status;
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_tenant_id_remote_url_key;
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_tenant_id_fkey;
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_credential_id_fkey;
ALTER TABLE public.repos DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE public.repos DROP COLUMN IF EXISTS credential_id;

ALTER TABLE public.analyze_jobs RENAME COLUMN tenant_id TO requested_by_tenant_id;
ALTER TABLE public.webhook_events RENAME COLUMN tenant_id TO received_by_tenant_id;

DROP INDEX IF EXISTS public.idx_analyze_jobs_tenant_repo_status;
CREATE INDEX IF NOT EXISTS idx_repos_remote_trgm ON public.repos USING GIN (remote_url gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_repos_full_name_trgm ON public.repos USING GIN (full_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_repos_visibility_status ON public.repos(visibility, analyze_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_repos_normalized_remote_key ON public.repos(provider, code_host_url, normalized_remote_key) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_repos_external ON public.repos(provider, code_host_url, external_id) WHERE external_id IS NOT NULL AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_repos_remote_key ON public.repos(provider, code_host_url, normalized_remote_key) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenant_repo_subscriptions_tenant ON public.tenant_repo_subscriptions(tenant_id, status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tenant_repo_subscriptions_repo ON public.tenant_repo_subscriptions(repo_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_repo_access_grants_subject ON public.repo_access_grants(subject_type, subject_id, permission, expires_at);
CREATE INDEX IF NOT EXISTS idx_analyze_jobs_requested_repo_status ON public.analyze_jobs(requested_by_tenant_id, repo_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhook_events_received_time ON public.webhook_events(received_by_tenant_id, received_at DESC);

ALTER TABLE public.tenant_lifecycle_jobs DROP CONSTRAINT IF EXISTS tenant_lifecycle_jobs_type_chk;
DELETE FROM public.tenant_lifecycle_jobs WHERE type = 'schema_upgrade';
ALTER TABLE public.tenant_lifecycle_jobs ADD CONSTRAINT tenant_lifecycle_jobs_type_chk CHECK (type IN ('export','purge','quota_recalculate'));

CREATE TABLE IF NOT EXISTS public.files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    commit_hash VARCHAR(80) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    language VARCHAR(80),
    size_bytes BIGINT NOT NULL DEFAULT 0,
    content_hash VARCHAR(128) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (repo_id, commit_hash, path),
    UNIQUE (repo_id, id)
);

CREATE TABLE IF NOT EXISTS public.symbols (
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    uid VARCHAR(512) NOT NULL,
    file_id UUID,
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(64) NOT NULL,
    file_path VARCHAR(1024) NOT NULL,
    start_line INT,
    end_line INT,
    content TEXT,
    parameter_count INT,
    return_type VARCHAR(512),
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    content_hash VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (repo_id, uid),
    CONSTRAINT symbols_file_fk FOREIGN KEY (repo_id, file_id) REFERENCES public.files(repo_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    source_uid VARCHAR(512) NOT NULL,
    target_uid VARCHAR(512) NOT NULL,
    relation_type VARCHAR(80) NOT NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    reason TEXT,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (repo_id, source_uid, target_uid, relation_type),
    CONSTRAINT relations_source_fk FOREIGN KEY (repo_id, source_uid) REFERENCES public.symbols(repo_id, uid) ON DELETE CASCADE,
    CONSTRAINT relations_target_fk FOREIGN KEY (repo_id, target_uid) REFERENCES public.symbols(repo_id, uid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.symbol_search (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL,
    symbol_uid VARCHAR(512) NOT NULL,
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(64) NOT NULL,
    file_path VARCHAR(1024) NOT NULL,
    content TEXT,
    summary TEXT,
    community_id VARCHAR(120),
    process_id VARCHAR(120),
    name_tsv TSVECTOR GENERATED ALWAYS AS (to_tsvector('simple', coalesce(name, ''))) STORED,
    content_tsv TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', coalesce(content, '') || ' ' || coalesce(summary, ''))) STORED,
    embedding VECTOR(1536),
    embedding_model VARCHAR(120),
    embedding_dimension INT NOT NULL DEFAULT 1536,
    search_boost NUMERIC(8,4) NOT NULL DEFAULT 1.0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (repo_id, symbol_uid),
    CONSTRAINT symbol_search_symbol_fk FOREIGN KEY (repo_id, symbol_uid) REFERENCES public.symbols(repo_id, uid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.communities (
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    id VARCHAR(120) NOT NULL,
    label VARCHAR(255) NOT NULL,
    cohesion DOUBLE PRECISION,
    symbol_count INT NOT NULL DEFAULT 0,
    keywords TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    description TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (repo_id, id)
);

CREATE TABLE IF NOT EXISTS public.symbol_communities (
    repo_id UUID NOT NULL,
    symbol_uid VARCHAR(512) NOT NULL,
    community_id VARCHAR(120) NOT NULL,
    score DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    PRIMARY KEY (repo_id, symbol_uid, community_id),
    CONSTRAINT symbol_communities_symbol_fk FOREIGN KEY (repo_id, symbol_uid) REFERENCES public.symbols(repo_id, uid) ON DELETE CASCADE,
    CONSTRAINT symbol_communities_community_fk FOREIGN KEY (repo_id, community_id) REFERENCES public.communities(repo_id, id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.processes (
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    id VARCHAR(120) NOT NULL,
    label VARCHAR(255) NOT NULL,
    process_type VARCHAR(80) NOT NULL,
    step_count INT NOT NULL DEFAULT 0,
    communities TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    entry_point_uid VARCHAR(512),
    terminal_uid VARCHAR(512),
    summary TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (repo_id, id)
);

CREATE TABLE IF NOT EXISTS public.process_steps (
    repo_id UUID NOT NULL,
    process_id VARCHAR(120) NOT NULL,
    step_index INT NOT NULL,
    symbol_uid VARCHAR(512) NOT NULL,
    relation_to_next VARCHAR(80),
    PRIMARY KEY (repo_id, process_id, step_index),
    CONSTRAINT process_steps_process_fk FOREIGN KEY (repo_id, process_id) REFERENCES public.processes(repo_id, id) ON DELETE CASCADE,
    CONSTRAINT process_steps_symbol_fk FOREIGN KEY (repo_id, symbol_uid) REFERENCES public.symbols(repo_id, uid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.analysis_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES public.repos(id) ON DELETE CASCADE,
    commit_hash VARCHAR(80) NOT NULL,
    analyze_job_id UUID REFERENCES public.analyze_jobs(id) ON DELETE SET NULL,
    files_count INT NOT NULL DEFAULT 0,
    symbols_count INT NOT NULL DEFAULT 0,
    relations_count INT NOT NULL DEFAULT 0,
    processes_count INT NOT NULL DEFAULT 0,
    communities_count INT NOT NULL DEFAULT 0,
    stats JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (repo_id, commit_hash)
);

CREATE INDEX IF NOT EXISTS idx_files_repo_path ON public.files(repo_id, path);
CREATE INDEX IF NOT EXISTS idx_files_repo_language ON public.files(repo_id, language);
CREATE INDEX IF NOT EXISTS idx_files_hash ON public.files(content_hash);

CREATE INDEX IF NOT EXISTS idx_symbols_repo_kind_name ON public.symbols(repo_id, kind, name);
CREATE INDEX IF NOT EXISTS idx_symbols_file_path ON public.symbols(file_path);
CREATE INDEX IF NOT EXISTS idx_symbols_name_trgm ON public.symbols USING GIN (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_symbols_props_gin ON public.symbols USING GIN (properties);

CREATE INDEX IF NOT EXISTS idx_relations_source ON public.relations(repo_id, source_uid, relation_type);
CREATE INDEX IF NOT EXISTS idx_relations_target ON public.relations(repo_id, target_uid, relation_type);
CREATE INDEX IF NOT EXISTS idx_relations_repo_type ON public.relations(repo_id, relation_type);

CREATE INDEX IF NOT EXISTS idx_symbol_search_repo_kind ON public.symbol_search(repo_id, kind);
CREATE INDEX IF NOT EXISTS idx_symbol_search_file ON public.symbol_search(repo_id, file_path);
CREATE INDEX IF NOT EXISTS idx_symbol_search_name_tsv ON public.symbol_search USING GIN (name_tsv);
CREATE INDEX IF NOT EXISTS idx_symbol_search_content_tsv ON public.symbol_search USING GIN (content_tsv);
CREATE INDEX IF NOT EXISTS idx_symbol_search_name_trgm ON public.symbol_search USING GIN (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_symbol_search_metadata_gin ON public.symbol_search USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_symbol_search_embedding_hnsw ON public.symbol_search USING HNSW (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);

CREATE INDEX IF NOT EXISTS idx_communities_repo_label ON public.communities(repo_id, label);
CREATE INDEX IF NOT EXISTS idx_processes_repo_type ON public.processes(repo_id, process_type);
CREATE INDEX IF NOT EXISTS idx_process_steps_symbol ON public.process_steps(repo_id, symbol_uid);
CREATE INDEX IF NOT EXISTS idx_analysis_snapshots_repo_time ON public.analysis_snapshots(repo_id, created_at DESC);

DROP TABLE IF EXISTS repo_canonical_map;
