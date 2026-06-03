// Package controller provides centralized HTTP route registration.
package controller

import (
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/controller/admin"
	"multi-tenant-saas/internal/controller/apikey"
	"multi-tenant-saas/internal/controller/audit"
	authcontroller "multi-tenant-saas/internal/controller/auth"
	"multi-tenant-saas/internal/controller/dashboard"
	"multi-tenant-saas/internal/controller/hello"
	"multi-tenant-saas/internal/controller/invitation"
	"multi-tenant-saas/internal/controller/me"
	"multi-tenant-saas/internal/controller/member"
	setupcontroller "multi-tenant-saas/internal/controller/setup"
	"multi-tenant-saas/internal/controller/tenant"
	totpcontroller "multi-tenant-saas/internal/controller/totp"
	"multi-tenant-saas/internal/middleware"
	"multi-tenant-saas/internal/service"
)

// RouteConfig describes one route group and its shared middleware chain.
type RouteConfig struct {
	Prefix      string
	Middlewares []ghttp.HandlerFunc
	Controllers []interface{}
}

// RegisterRootRoutes registers non-versioned routes.
func RegisterRootRoutes(group *ghttp.RouterGroup) {
	registerHealthRoutes(group)
	group.Bind(hello.NewV1())
}

// RegisterRoutes registers all /api/v1 routes.
func RegisterRoutes(group *ghttp.RouterGroup) {
	group.Middleware(middleware.InitGuard)
	for _, route := range getAllRoutes() {
		route := route
		group.Group(route.Prefix, func(subGroup *ghttp.RouterGroup) {
			for _, mw := range route.Middlewares {
				subGroup.Middleware(mw)
			}
			subGroup.Bind(route.Controllers...)
		})
	}
}

func getAllRoutes() []RouteConfig {
	return []RouteConfig{
		{
			Prefix:      "",
			Middlewares: []ghttp.HandlerFunc{middleware.RateLimit},
			Controllers: []interface{}{setupcontroller.NewV1()},
		},
		{
			Prefix:      "",
			Middlewares: []ghttp.HandlerFunc{middleware.RateLimit},
			Controllers: []interface{}{authcontroller.NewPublicV1(), totpcontroller.NewAuthPublicV1(), authcontroller.NewOAuthV1()},
		},
		{
			Prefix: "",
			Middlewares: []ghttp.HandlerFunc{
				middleware.RateLimit,
				middleware.Auth,
				middleware.CSRF,
			},
			Controllers: []interface{}{authcontroller.NewV1()},
		},
		{
			Prefix: "",
			Middlewares: []ghttp.HandlerFunc{
				middleware.RateLimit,
				middleware.Auth,
				middleware.CSRF,
				middleware.TenantResolver,
			},
			Controllers: []interface{}{
				me.NewV1(),
				tenant.NewV1(),
				member.NewV1(),
				invitation.NewV1(),
				audit.NewV1(),
				apikey.NewV1(),
				totpcontroller.NewMeTotpV1(),
			},
		},
		{
			Prefix: "/dashboard",
			Middlewares: []ghttp.HandlerFunc{
				middleware.RateLimit,
				middleware.Auth,
				middleware.CSRF,
				middleware.TenantResolver,
			},
			Controllers: []interface{}{
				dashboard.NewV1(),
			},
		},
		{
			Prefix: "",
			Middlewares: []ghttp.HandlerFunc{
				middleware.RateLimit,
				middleware.Auth,
				middleware.CSRF,
				middleware.PlatformAdmin,
			},
			Controllers: []interface{}{
				admin.NewV1(),
			},
		},
	}
}

func registerHealthRoutes(group *ghttp.RouterGroup) {
	group.GET("/healthz", func(r *ghttp.Request) {
		r.Response.WriteJsonExit(g.Map{"ok": true})
	})
	group.GET("/readyz", func(r *ghttp.Request) {
		ctx := r.GetCtx()
		version, dirty, err := service.AutoMigrate().Status(ctx)
		if err != nil || dirty {
			r.Response.Status = http.StatusServiceUnavailable
			r.Response.WriteJsonExit(g.Map{"ok": false, "version": version, "dirty": dirty, "error": err})
			return
		}
		if _, err := g.DB().GetOne(ctx, "SELECT 1"); err != nil {
			r.Response.Status = http.StatusServiceUnavailable
			r.Response.WriteJsonExit(g.Map{"ok": false, "version": version, "db": err.Error()})
			return
		}
		setupState, err := service.SystemSetup().State(ctx)
		if err != nil {
			r.Response.Status = http.StatusServiceUnavailable
			r.Response.WriteJsonExit(g.Map{"ok": false, "version": version, "db": "ok", "setup": err.Error()})
			return
		}
		if setupState.RequiresSetup {
			r.Response.WriteJsonExit(g.Map{"ok": true, "version": version, "dirty": dirty, "db": "ok", "setup_required": true, "code": service.SetupCodeRequired})
			return
		}
		r.Response.WriteJsonExit(g.Map{"ok": true, "version": version, "dirty": dirty, "db": "ok", "setup_required": false})
	})
}
