ALTER TABLE public.users ADD COLUMN IF NOT EXISTS tenant_id UUID;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS role VARCHAR(32) DEFAULT 'member';
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(32);
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS auth_id VARCHAR(255);

ALTER TABLE public.users
ADD CONSTRAINT users_tenant_id_fkey
FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;

ALTER TABLE public.users
ADD CONSTRAINT users_role_chk
CHECK (role IN ('owner','admin','member','viewer'));

ALTER TABLE public.users
ADD CONSTRAINT users_tenant_id_email_key
UNIQUE (tenant_id, email);

ALTER TABLE public.users
ADD CONSTRAINT users_auth_provider_auth_id_key
UNIQUE (auth_provider, auth_id);

CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON public.users(tenant_id, role) WHERE deleted_at IS NULL;
