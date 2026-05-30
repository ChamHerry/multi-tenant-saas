package bizctx

import (
	"context"
	"testing"

	"multi-tenant-saas/internal/service"
)

func TestGetReturnsEmptyWhenNotSet(t *testing.T) {
	ctx := context.Background()
	assertEqual(t, "", service.BizCtx().GetRequestID(ctx))
	assertEqual(t, "", service.BizCtx().GetUserID(ctx))
	assertEqual(t, "", service.BizCtx().GetClientIP(ctx))
	assertEqual(t, "", service.BizCtx().GetUserAgent(ctx))
	assertEqual(t, "", service.BizCtx().GetTenantID(ctx))
	assertEqual(t, "", service.BizCtx().GetRole(ctx))
	assertEqual(t, "", service.BizCtx().GetPlatformRole(ctx))
	assertEqual(t, "", service.BizCtx().GetAuthType(ctx))
	if service.BizCtx().IsPlatformAdmin(ctx) {
		t.Fatal("expected IsPlatformAdmin=false when not set")
	}
	if perms := service.BizCtx().GetPermissions(ctx); perms != nil {
		t.Fatalf("expected nil permissions, got %v", perms)
	}
}

func TestSetAndGetRequestMetadata(t *testing.T) {
	ctx := context.WithValue(context.Background(), service.BizContextKey("repomind.biz_context"), &service.BizContext{})
	service.BizCtx().SetRequestID(ctx, "req-123")
	service.BizCtx().SetTraceID(ctx, "trace-456")
	service.BizCtx().SetClientIP(ctx, "1.2.3.4")
	service.BizCtx().SetUserAgent(ctx, "TestAgent/1.0")

	assertEqual(t, "req-123", service.BizCtx().GetRequestID(ctx))
	assertEqual(t, "trace-456", service.BizCtx().GetTraceID(ctx))
	assertEqual(t, "1.2.3.4", service.BizCtx().GetClientIP(ctx))
	assertEqual(t, "TestAgent/1.0", service.BizCtx().GetUserAgent(ctx))
}

func TestSetAndGetIdentity(t *testing.T) {
	ctx := context.WithValue(context.Background(), service.BizContextKey("repomind.biz_context"), &service.BizContext{})
	service.BizCtx().SetIdentity(ctx, &service.AuthIdentity{
		UserID: "user-789",
		Type:   "session",
	})

	assertEqual(t, "user-789", service.BizCtx().GetUserID(ctx))
	assertEqual(t, "session", service.BizCtx().GetAuthType(ctx))
	assertEqual(t, "", service.BizCtx().GetTenantID(ctx))
}

func TestSetAndGetTenant(t *testing.T) {
	ctx := context.WithValue(context.Background(), service.BizContextKey("repomind.biz_context"), &service.BizContext{})
	service.BizCtx().SetTenant(ctx, &service.TenantContext{
		TenantID:    "t-1",
		Role:        "owner",
		Permissions: []service.Permission{service.PermissionTenantRead, service.PermissionTenantManage},
	})

	assertEqual(t, "t-1", service.BizCtx().GetTenantID(ctx))
	assertEqual(t, "owner", service.BizCtx().GetRole(ctx))
	if len(service.BizCtx().GetPermissions(ctx)) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(service.BizCtx().GetPermissions(ctx)))
	}
	if service.BizCtx().IsPlatformAdmin(ctx) {
		t.Fatal("expected IsPlatformAdmin=false")
	}
}

func TestSetAndGetPlatformAdmin(t *testing.T) {
	ctx := context.WithValue(context.Background(), service.BizContextKey("repomind.biz_context"), &service.BizContext{})
	service.BizCtx().SetPlatformAdmin(ctx, &service.PlatformAdminContext{
		UserID: "admin-1",
		Role:   "super_admin",
	})

	assertEqual(t, "super_admin", service.BizCtx().GetPlatformRole(ctx))
	if !service.BizCtx().IsPlatformAdmin(ctx) {
		t.Fatal("expected IsPlatformAdmin=true")
	}
}

func TestGetReturnsNilIdentityFieldsWhenNotSet(t *testing.T) {
	ctx := context.WithValue(context.Background(), service.BizContextKey("repomind.biz_context"), &service.BizContext{})
	// No identity set — should return empty strings.
	assertEqual(t, "", service.BizCtx().GetUserID(ctx))
	assertEqual(t, "", service.BizCtx().GetAuthType(ctx))
}

func TestFullBizContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), service.BizContextKey("repomind.biz_context"), &service.BizContext{
		RequestID: "req-full",
		TraceID:   "trace-full",
		ClientIP:  "10.0.0.1",
		UserAgent: "Mozilla/5.0",
		Identity: &service.AuthIdentity{
			UserID:    "u-1",
			Type:      "api_key",
			APIKeyID:  "ak-1",
			SessionID: "",
		},
		Tenant: &service.TenantContext{
			TenantID: "t-1",
			Role:     "member",
		},
		PlatformAdmin: nil,
	})

	bc := service.BizCtx().Get(ctx)
	assertEqual(t, "req-full", bc.RequestID)
	assertEqual(t, "trace-full", bc.TraceID)
	assertEqual(t, "10.0.0.1", bc.ClientIP)
	assertEqual(t, "Mozilla/5.0", bc.UserAgent)
	assertEqual(t, "u-1", bc.Identity.UserID)
	assertEqual(t, "t-1", bc.Tenant.TenantID)
	if bc.PlatformAdmin != nil {
		t.Fatal("expected nil PlatformAdmin")
	}

	// Convenience methods
	assertEqual(t, "req-full", service.BizCtx().GetRequestID(ctx))
	assertEqual(t, "u-1", service.BizCtx().GetUserID(ctx))
	assertEqual(t, "api_key", service.BizCtx().GetAuthType(ctx))
	assertEqual(t, "t-1", service.BizCtx().GetTenantID(ctx))
	assertEqual(t, "member", service.BizCtx().GetRole(ctx))
	if service.BizCtx().IsPlatformAdmin(ctx) {
		t.Fatal("expected IsPlatformAdmin=false")
	}
}

func assertEqual(t *testing.T, expected, actual string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
