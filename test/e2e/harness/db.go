//go:build e2e

package harness

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/testutil"
)

func ResetDatabase(t *testing.T) context.Context {
	t.Helper()
	ctx := context.Background()
	testutil.TruncateAllTables(ctx, t)
	testutil.MarkSystemSetupInitialized(ctx, t)
	return ctx
}

func ScalarString(t *testing.T, sql string, args ...any) string {
	t.Helper()
	v, err := g.DB().GetValue(context.Background(), sql, args...)
	if err != nil {
		t.Fatalf("query scalar string: %v", err)
	}
	return v.String()
}

func ScalarInt(t *testing.T, sql string, args ...any) int {
	t.Helper()
	v, err := g.DB().GetValue(context.Background(), sql, args...)
	if err != nil {
		t.Fatalf("query scalar int: %v", err)
	}
	return v.Int()
}
