# Multi-tenant SaaS Platform Template

A GoFrame + React template for reusable multi-tenant SaaS control-plane development.

Core capabilities:

- password login, sessions, CSRF protection
- users, tenant memberships, invitations and tenant context resolution
- tenant RBAC and platform-admin RBAC
- API key management
- tenant/platform audit logs
- subscriptions, plan entitlements and generic quota usage
- PostgreSQL migrations with runtime readiness checks

This repository intentionally excludes product-specific domains so future projects can add their own tenant-bound business tables on top of the shared SaaS foundation.

## Docker

Start the SPA, Go API and PostgreSQL with one command:

```sh
scripts/manage-docker.sh
```

Use `scripts/manage-docker.sh help` to see start, stop, logs, status, psql and cleanup commands.

The app is available at `http://127.0.0.1:8000`. PostgreSQL is only exposed inside the Compose network by default to avoid local port conflicts.

The Docker build keeps npm cache, Go module cache and Go build cache in separate BuildKit layers. Frontend dependency changes, frontend source changes and backend source changes are also split into separate stages to keep rebuilds fast.
