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
