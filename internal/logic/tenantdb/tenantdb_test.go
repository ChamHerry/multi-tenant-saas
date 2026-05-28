package tenantdb

import (
	"context"
	"testing"

	"repomind-temp/internal/service"
)

func TestTenantContextHelpers(t *testing.T) {
	ctx := context.Background()
	if _, ok := service.TenantContextFromCtx(ctx); ok {
		t.Fatal("TenantContextFromCtx(empty)=true")
	}
	tc := &service.TenantContext{TenantID: "t1", UserID: "u1", Role: "owner"}
	ctx = service.WithTenantContext(ctx, tc)
	got, ok := service.TenantContextFromCtx(ctx)
	if !ok || got != tc {
		t.Fatal("TenantContextFromCtx did not return stored pointer")
	}
	if _, err := service.MustTenantContext(context.Background()); err == nil {
		t.Fatal("MustTenantContext(empty) expected error")
	}
}

func TestGraphQueryRejectsDelimiter(t *testing.T) {
	ctx := service.WithTenantContext(context.Background(), &service.TenantContext{TenantID: "tenant-id", UserID: "user-id"})
	_, err := (&sTenantGraph{}).Query(ctx, "RETURN '$$'", nil)
	if err == nil {
		t.Fatal("Query with $$ delimiter expected error")
	}
}

func TestGraphQueryRetired(t *testing.T) {
	ctx := service.WithTenantContext(context.Background(), &service.TenantContext{TenantID: "tenant-id", UserID: "user-id"})
	_, err := (&sTenantGraph{}).Query(ctx, "MATCH (n) RETURN n", nil)
	if err == nil {
		t.Fatal("generic tenant graph gateway should be retired")
	}
}
