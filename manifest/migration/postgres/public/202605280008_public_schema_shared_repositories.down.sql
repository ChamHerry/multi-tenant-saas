DROP INDEX IF EXISTS public.idx_analysis_snapshots_repo_time;
DROP INDEX IF EXISTS public.idx_process_steps_symbol;
DROP INDEX IF EXISTS public.idx_processes_repo_type;
DROP INDEX IF EXISTS public.idx_communities_repo_label;
DROP INDEX IF EXISTS public.idx_symbol_search_embedding_hnsw;
DROP INDEX IF EXISTS public.idx_symbol_search_metadata_gin;
DROP INDEX IF EXISTS public.idx_symbol_search_name_trgm;
DROP INDEX IF EXISTS public.idx_symbol_search_content_tsv;
DROP INDEX IF EXISTS public.idx_symbol_search_name_tsv;
DROP INDEX IF EXISTS public.idx_symbol_search_file;
DROP INDEX IF EXISTS public.idx_symbol_search_repo_kind;
DROP INDEX IF EXISTS public.idx_relations_repo_type;
DROP INDEX IF EXISTS public.idx_relations_target;
DROP INDEX IF EXISTS public.idx_relations_source;
DROP INDEX IF EXISTS public.idx_symbols_props_gin;
DROP INDEX IF EXISTS public.idx_symbols_name_trgm;
DROP INDEX IF EXISTS public.idx_symbols_file_path;
DROP INDEX IF EXISTS public.idx_symbols_repo_kind_name;
DROP INDEX IF EXISTS public.idx_files_hash;
DROP INDEX IF EXISTS public.idx_files_repo_language;
DROP INDEX IF EXISTS public.idx_files_repo_path;

DROP TABLE IF EXISTS public.analysis_snapshots;
DROP TABLE IF EXISTS public.process_steps;
DROP TABLE IF EXISTS public.processes;
DROP TABLE IF EXISTS public.symbol_communities;
DROP TABLE IF EXISTS public.communities;
DROP TABLE IF EXISTS public.symbol_search;
DROP TABLE IF EXISTS public.relations;
DROP TABLE IF EXISTS public.symbols;
DROP TABLE IF EXISTS public.files;

ALTER TABLE public.tenant_lifecycle_jobs DROP CONSTRAINT IF EXISTS tenant_lifecycle_jobs_type_chk;
ALTER TABLE public.tenant_lifecycle_jobs ADD CONSTRAINT tenant_lifecycle_jobs_type_chk CHECK (type IN ('export','purge','schema_upgrade','quota_recalculate'));

DROP INDEX IF EXISTS public.idx_webhook_events_received_time;
DROP INDEX IF EXISTS public.idx_analyze_jobs_requested_repo_status;
DROP INDEX IF EXISTS public.idx_repo_access_grants_subject;
DROP INDEX IF EXISTS public.idx_tenant_repo_subscriptions_repo;
DROP INDEX IF EXISTS public.idx_tenant_repo_subscriptions_tenant;
DROP INDEX IF EXISTS public.uq_repos_remote_key;
DROP INDEX IF EXISTS public.uq_repos_external;
DROP INDEX IF EXISTS public.idx_repos_normalized_remote_key;
DROP INDEX IF EXISTS public.idx_repos_visibility_status;
DROP INDEX IF EXISTS public.idx_repos_full_name_trgm;

ALTER TABLE public.webhook_events RENAME COLUMN received_by_tenant_id TO tenant_id;
ALTER TABLE public.analyze_jobs RENAME COLUMN requested_by_tenant_id TO tenant_id;

ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE public.repos ADD COLUMN IF NOT EXISTS credential_id UUID;
UPDATE public.repos r
SET tenant_id = s.tenant_id,
    credential_id = s.credential_id
FROM (
    SELECT DISTINCT ON (repo_id) repo_id, tenant_id, credential_id
    FROM public.tenant_repo_subscriptions
    WHERE deleted_at IS NULL
    ORDER BY repo_id, created_at ASC
) s
WHERE s.repo_id = r.id;
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_tenant_id_fkey;
ALTER TABLE public.repos ADD CONSTRAINT repos_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_credential_id_fkey;
ALTER TABLE public.repos ADD CONSTRAINT repos_credential_id_fkey FOREIGN KEY (credential_id) REFERENCES public.git_credentials(id) ON DELETE SET NULL;
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_tenant_id_remote_url_key;
ALTER TABLE public.repos ADD CONSTRAINT repos_tenant_id_remote_url_key UNIQUE (tenant_id, remote_url);

DROP TABLE IF EXISTS public.repo_access_grants;
DROP TABLE IF EXISTS public.tenant_repo_subscriptions;

ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_visibility_chk;
ALTER TABLE public.repos DROP CONSTRAINT IF EXISTS repos_provider_chk;
ALTER TABLE public.repos DROP COLUMN IF EXISTS permission_synced_at;
ALTER TABLE public.repos DROP COLUMN IF EXISTS indexed_commit_hash;
ALTER TABLE public.repos DROP COLUMN IF EXISTS is_archived;
ALTER TABLE public.repos DROP COLUMN IF EXISTS is_fork;
ALTER TABLE public.repos DROP COLUMN IF EXISTS visibility;
ALTER TABLE public.repos DROP COLUMN IF EXISTS normalized_remote_key;
ALTER TABLE public.repos DROP COLUMN IF EXISTS full_name;
ALTER TABLE public.repos DROP COLUMN IF EXISTS owner_name;
ALTER TABLE public.repos DROP COLUMN IF EXISTS external_node_id;
ALTER TABLE public.repos DROP COLUMN IF EXISTS external_id;
ALTER TABLE public.repos DROP COLUMN IF EXISTS code_host_url;
ALTER TABLE public.repos DROP COLUMN IF EXISTS metadata;
ALTER TABLE public.repos ADD CONSTRAINT repos_provider_chk CHECK (provider IN ('github','gitlab','gitea','bitbucket'));

CREATE INDEX IF NOT EXISTS idx_repos_tenant_status ON public.repos(tenant_id, analyze_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_analyze_jobs_tenant_repo_status ON public.analyze_jobs(tenant_id, repo_id, status, created_at DESC);

ALTER TABLE public.tenants ADD COLUMN IF NOT EXISTS schema_name VARCHAR(63);
ALTER TABLE public.tenants ADD COLUMN IF NOT EXISTS graph_name VARCHAR(63);
ALTER TABLE public.tenants ADD COLUMN IF NOT EXISTS tenant_schema_version BIGINT NOT NULL DEFAULT 0;
ALTER TABLE public.tenants ADD COLUMN IF NOT EXISTS tenant_schema_dirty BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE public.tenants ADD COLUMN IF NOT EXISTS tenant_schema_checked_at TIMESTAMPTZ;
UPDATE public.tenants
SET schema_name = COALESCE(schema_name, 'tenant_' || replace(id::text, '-', '_')),
    graph_name = COALESCE(graph_name, 'graph_tenant_' || replace(id::text, '-', '_') || '_code_graph');
ALTER TABLE public.tenants ALTER COLUMN schema_name SET NOT NULL;
ALTER TABLE public.tenants ALTER COLUMN graph_name SET NOT NULL;
ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS tenants_schema_name_key;
ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS tenants_graph_name_key;
ALTER TABLE public.tenants ADD CONSTRAINT tenants_schema_name_key UNIQUE(schema_name);
ALTER TABLE public.tenants ADD CONSTRAINT tenants_graph_name_key UNIQUE(graph_name);
ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS tenants_schema_name_chk;
ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS tenants_graph_name_chk;
ALTER TABLE public.tenants ADD CONSTRAINT tenants_schema_name_chk CHECK (schema_name ~ '^tenant_[a-z0-9_]{8,55}$');
ALTER TABLE public.tenants ADD CONSTRAINT tenants_graph_name_chk CHECK (graph_name ~ '^graph_tenant_[a-z0-9_]{8,55}_code_graph$');
