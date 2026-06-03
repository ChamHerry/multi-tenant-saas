CREATE TABLE IF NOT EXISTS public.system_setup (
    id                     SMALLINT      NOT NULL PRIMARY KEY DEFAULT 1,
    status                 VARCHAR(32)   NOT NULL DEFAULT 'initialized',
    version                VARCHAR(64)   NOT NULL DEFAULT '',
    initialized_by_user_id UUID          NULL REFERENCES public.users(id) ON DELETE SET NULL,
    initialized_at         TIMESTAMPTZ   NOT NULL DEFAULT now(),
    metadata               JSONB         NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT system_setup_singleton_chk CHECK (id = 1),
    CONSTRAINT system_setup_status_chk CHECK (status IN ('initialized'))
);

COMMENT ON TABLE public.system_setup IS 'Singleton installation lock for the first-run setup wizard';
COMMENT ON COLUMN public.system_setup.id IS 'Singleton id. Must always be 1.';
COMMENT ON COLUMN public.system_setup.status IS 'Setup status. Row existence means initialized.';
COMMENT ON COLUMN public.system_setup.version IS 'Application/setup version that completed initialization.';
COMMENT ON COLUMN public.system_setup.initialized_by_user_id IS 'Platform administrator user that completed setup.';
COMMENT ON COLUMN public.system_setup.initialized_at IS 'Timestamp when setup was completed.';
COMMENT ON COLUMN public.system_setup.metadata IS 'Non-secret setup metadata. Must never contain generated secrets.';
