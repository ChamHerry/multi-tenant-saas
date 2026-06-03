# Go E2E Tests

This tree contains Go `testing`-based end-to-end tests for the multi-tenant SaaS auth/setup/tenant/frontend flows.

## Run

```bash
# Required dependency for the local GoFrame test server mode.
docker compose up -d postgres

# Full Go e2e suite. `-p=1` keeps package-level shared DB state serialized.
go test -tags=e2e -p=1 -count=1 -timeout 15m -v ./test/e2e/...
```

`test/e2e/harness` serializes package suites with a cross-process lock so the suite also remains safe when callers forget `-p=1`.

## Scope

- `harness/`: package setup, DB/server lock, config, diagnostics, HTTP helpers.
- `auth/`: session, password reset, email verification, account lockout, TOTP, OAuth.
- `setup/`: setup wizard, platform admin bootstrap, system config.
- `tenant/`: tenant CRUD, members, invitations, API keys, permission/middleware checks.
- `frontend/`: built SPA artifact smoke checks and optional live app fallback checks.

Legacy bash e2e scripts under `hack/` have been removed after their scenarios were migrated into this Go suite.
