//go:build e2e

package auth_test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func scalarString(t *testing.T, sql string, args ...any) string {
	t.Helper()
	value, err := g.DB().GetValue(context.Background(), sql, args...)
	if err != nil {
		t.Fatalf("query scalar string: %v", err)
	}
	return value.String()
}

func scalarInt(t *testing.T, sql string, args ...any) int {
	t.Helper()
	value, err := g.DB().GetValue(context.Background(), sql, args...)
	if err != nil {
		t.Fatalf("query scalar int: %v", err)
	}
	return value.Int()
}
