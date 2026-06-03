// Package logic is the unified side-effect import surface for service implementations.
//
// Each logic subpackage registers its implementation into internal/service from init().
// Import this package once from application entrypoints instead of scattering blank imports
// across main/cmd/tests.
package logic

import (
	_ "multi-tenant-saas/internal/logic/access"
	_ "multi-tenant-saas/internal/logic/apikey"
	_ "multi-tenant-saas/internal/logic/audit"
	_ "multi-tenant-saas/internal/logic/auditquery"
	_ "multi-tenant-saas/internal/logic/auth"
	_ "multi-tenant-saas/internal/logic/authsession"
	_ "multi-tenant-saas/internal/logic/automigrate"
	_ "multi-tenant-saas/internal/logic/bizctx"
	_ "multi-tenant-saas/internal/logic/config"
	_ "multi-tenant-saas/internal/logic/email"
	_ "multi-tenant-saas/internal/logic/emailverification"
	_ "multi-tenant-saas/internal/logic/invitation"
	_ "multi-tenant-saas/internal/logic/lifecycle"
	_ "multi-tenant-saas/internal/logic/membership"
	_ "multi-tenant-saas/internal/logic/oauth"
	_ "multi-tenant-saas/internal/logic/passwordauth"
	_ "multi-tenant-saas/internal/logic/platformadmin"
	_ "multi-tenant-saas/internal/logic/rbac"
	_ "multi-tenant-saas/internal/logic/systemconfigadmin"
	_ "multi-tenant-saas/internal/logic/systemsetup"
	_ "multi-tenant-saas/internal/logic/tenant"
	_ "multi-tenant-saas/internal/logic/totp"
	_ "multi-tenant-saas/internal/logic/user"
)
