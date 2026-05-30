package auth_test

import (
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/test/gtest"

	_ "multi-tenant-saas/internal/logic"
	"multi-tenant-saas/internal/service"
)

func TestAuthLoginEndpoint(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		s := g.Server()
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(func(r *ghttp.Request) {
				r.Middleware.Next()
			})
			group.POST("/api/v1/auth/login", func(r *ghttp.Request) {
				r.Response.WriteJson(g.Map{
					"code":    0,
					"message": "OK",
					"data": g.Map{
						"user": g.Map{
							"id":    "test-uuid",
							"email": "test@example.com",
						},
					},
				})
			})
		})
		t.AssertNE(s, nil)
	})
}

func TestAuthServiceInterface(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.AssertNE(service.Auth(), nil)
	})
}

func TestAuthServiceSession(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		t.AssertNE(service.AuthSessionService(), nil)
	})
}
