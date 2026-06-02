package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/crypto"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	_ "multi-tenant-saas/internal/logic"
)

const (
	testDBHost     = "127.0.0.1"
	testDBPort     = "55432"
	testDBUser     = "saas_template"
	testDBPass     = "secret"
	testDBMainName = "saas_template"
	testDBTestName = "saas_template_test"
)

// SetupTestDB ensures the test database exists and GoFrame is configured to use it.
// It runs migrations to create all tables. Call this once from TestMain.
func SetupTestDB(t *testing.T) gdb.DB {
	t.Helper()
	ctx := context.Background()
	if err := crypto.InitEncryption(ctx); err != nil {
		t.Fatalf("init encryption: %v", err)
	}

	// 1. Connect to main database via database/sql to create test database.
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		testDBUser, testDBPass, testDBHost, testDBPort, testDBMainName)
	adminDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("connect admin db: %v", err)
	}
	defer adminDB.Close()

	// Retry connecting — docker-compose postgres might still be starting.
	for range 30 {
		if err = adminDB.PingContext(ctx); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		t.Fatalf("ping admin db after 30 retries: %v", err)
	}

	// 2. Create test database if not exists.
	_, _ = adminDB.ExecContext(ctx, fmt.Sprintf(
		"CREATE DATABASE %s OWNER %s", testDBTestName, testDBUser))

	// 3. Override GoFrame database config to use test database.
	testLink := fmt.Sprintf("pgsql:%s:%s@tcp(%s:%s)/%s",
		testDBUser, testDBPass, testDBHost, testDBPort, testDBTestName)
	gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			{
				Link:             testLink,
				Type:             "pgsql",
				Debug:            false,
				MaxIdleConnCount: 5,
				MaxOpenConnCount: 10,
			},
		},
	})

	// 4. Verify connection.
	db := g.DB()
	sqlDB, err := db.Master()
	if err != nil {
		t.Fatalf("get sql.DB from goframe: %v", err)
	}
	for range 10 {
		if err = sqlDB.PingContext(ctx); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("ping test db: %v", err)
	}
	t.Log("test database connected:", testDBTestName)

	// 5. Run migrations via the project's automigrate service.
	if err = service.AutoMigrate().Up(ctx); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	t.Log("migrations applied")

	return db
}

// TruncateAllTables truncates all user-data tables for test isolation.
func TruncateAllTables(ctx context.Context, t *testing.T) {
	t.Helper()
	tables := []string{
		"audit_logs",
		"auth_login_attempts",
		"auth_sessions",
		"user_totp_backup_codes",
		"user_totp_configs",
		"tenant_lifecycle_jobs",
		"tenant_invitations",
		"tenant_memberships",
		"api_key_tenant_grants",
		"api_keys",
		"tenants",
		"platform_admins",
		"user_password_credentials",
		"user_identities",
		"users",
	}
	db := g.DB()
	for _, table := range tables {
		if _, err := db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			t.Fatalf("truncate %s: %v", table, err)
		}
	}
}
