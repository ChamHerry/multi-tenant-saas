package automigrate

import (
	"testing"
	"testing/fstest"

	migrationfiles "repomind-temp/manifest/migration"
)

func TestLatestMigrationVersion(t *testing.T) {
	fsys := fstest.MapFS{
		"migrations/202605270101_init.up.sql":   {Data: []byte("SELECT 1;")},
		"migrations/202605270101_init.down.sql": {Data: []byte("SELECT 1;")},
		"migrations/202605270103_next.up.sql":   {Data: []byte("SELECT 1;")},
		"migrations/not_a_migration.sql":        {Data: []byte("SELECT 1;")},
	}
	version, err := LatestMigrationVersion(fsys, "migrations")
	if err != nil {
		t.Fatalf("LatestMigrationVersion() error = %v", err)
	}
	if version != 202605270103 {
		t.Fatalf("LatestMigrationVersion() = %d, want %d", version, uint64(202605270103))
	}
}

func TestLatestMigrationVersionNoMigrations(t *testing.T) {
	_, err := LatestMigrationVersion(fstest.MapFS{"migrations/readme.md": {Data: []byte("x")}}, "migrations")
	if err == nil {
		t.Fatal("LatestMigrationVersion() expected error for empty migration dir")
	}
}

func TestEmbeddedPublicMigrationVersion(t *testing.T) {
	publicVersion, err := LatestMigrationVersion(migrationfiles.FS, migrationfiles.PublicDir)
	if err != nil {
		t.Fatalf("public LatestMigrationVersion() error = %v", err)
	}
	if publicVersion != 202605280011 {
		t.Fatalf("public version = %d, want %d", publicVersion, uint64(202605280011))
	}
}
