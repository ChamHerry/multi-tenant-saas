DROP INDEX IF EXISTS public.idx_auth_login_attempts_locked;
DROP TABLE IF EXISTS public.auth_login_attempts;

DROP INDEX IF EXISTS public.idx_auth_sessions_expiry;
DROP INDEX IF EXISTS public.idx_auth_sessions_user_active;
DROP INDEX IF EXISTS public.idx_auth_sessions_secret_hash;
DROP TABLE IF EXISTS public.auth_sessions;

DROP INDEX IF EXISTS public.idx_user_identities_password_auth_id_lower;
DROP TABLE IF EXISTS public.user_password_credentials;
