// Package logic is the unified side-effect import surface for service implementations.
//
// Each logic subpackage registers its implementation into internal/service from init().
// Import this package once from application entrypoints instead of scattering blank imports
// across main/cmd/tests.
package logic

import (
	_ "repomind-temp/internal/logic/access"
	_ "repomind-temp/internal/logic/apikey"
	_ "repomind-temp/internal/logic/audit"
	_ "repomind-temp/internal/logic/auditquery"
	_ "repomind-temp/internal/logic/auth"
	_ "repomind-temp/internal/logic/authsession"
	_ "repomind-temp/internal/logic/automigrate"
	_ "repomind-temp/internal/logic/invitation"
	_ "repomind-temp/internal/logic/lifecycle"
	_ "repomind-temp/internal/logic/membership"
	_ "repomind-temp/internal/logic/passwordauth"
	_ "repomind-temp/internal/logic/platformadmin"
	_ "repomind-temp/internal/logic/quota"
	_ "repomind-temp/internal/logic/rbac"
	_ "repomind-temp/internal/logic/tenant"
	_ "repomind-temp/internal/logic/user"
)
