DROP TABLE IF EXISTS public.auth_login_challenges;
DROP TABLE IF EXISTS public.oauth_states;

DELETE FROM public.system_config
WHERE key LIKE 'oauth.github.%'
   OR key LIKE 'oauth.google.%'
   OR key IN ('oauth.stateTTL', 'oauth.challengeTTL');

ALTER TABLE public.user_identities
    DROP CONSTRAINT IF EXISTS user_identities_provider_chk;

-- Operational caveat: this rollback is only safe when no google identities exist.
ALTER TABLE public.user_identities
    ADD CONSTRAINT user_identities_provider_chk
    CHECK (provider IN ('github','gitlab','gitea','oidc','password'));
