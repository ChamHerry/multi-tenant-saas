// Package controller provides centralized HTTP route registration.
package controller

import (
	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"repomind-temp/internal/controller/admin"
	"repomind-temp/internal/controller/apikey"
	"repomind-temp/internal/controller/audit"
	authcontroller "repomind-temp/internal/controller/auth"
	"repomind-temp/internal/controller/hello"
	"repomind-temp/internal/controller/invitation"
	"repomind-temp/internal/controller/me"
	"repomind-temp/internal/controller/member"
	"repomind-temp/internal/controller/tenant"
	"repomind-temp/internal/middleware"
	"repomind-temp/internal/service"
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
			Controllers: []interface{}{authcontroller.NewPublicV1()},
		},
		{
			Prefix: "",
			Middlewares: []ghttp.HandlerFunc{
				middleware.Auth,
				middleware.CSRF,
			},
			Controllers: []interface{}{authcontroller.NewV1()},
		},
		{
			Prefix: "",
			Middlewares: []ghttp.HandlerFunc{
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
			},
		},
		{
			Prefix: "",
			Middlewares: []ghttp.HandlerFunc{
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
		version, dirty, err := service.AutoMigrate().Status(r.GetCtx())
		if err != nil || dirty {
			r.Response.Status = http.StatusServiceUnavailable
			r.Response.WriteJsonExit(g.Map{"ok": false, "version": version, "dirty": dirty, "error": err})
		}
		r.Response.WriteJsonExit(g.Map{"ok": true, "version": version, "dirty": dirty})
	})
}
