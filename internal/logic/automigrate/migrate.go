package automigrate

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	migrationfiles "repomind-temp/manifest/migration"
)

const (
	publicSchemaName      = "public"
	publicMigrationsTable = "schema_migrations"
)

type migrationTarget struct {
	Dir             string
	SchemaName      string
	MigrationsTable string
	SQLDB           *sql.DB
}

func newMigrate(ctx context.Context, target migrationTarget) (*migrate.Migrate, error) {
	sourceDriver, err := iofs.New(migrationfiles.FS, target.Dir)
	if err != nil {
		return nil, gerror.Wrapf(err, "create migration source %s", target.Dir)
	}
	conn, err := target.SQLDB.Conn(ctx)
	if err != nil {
		_ = sourceDriver.Close()
		return nil, gerror.Wrapf(err, "get migration connection for schema %s", target.SchemaName)
	}
	databaseDriver, err := postgres.WithConnection(ctx, conn, &postgres.Config{
		SchemaName:            target.SchemaName,
		MigrationsTable:       target.MigrationsTable,
		MultiStatementEnabled: true,
	})
	if err != nil {
		_ = sourceDriver.Close()
		_ = conn.Close()
		return nil, gerror.Wrapf(err, "create postgres migration driver for schema %s", target.SchemaName)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", databaseDriver)
	if err != nil {
		_, _ = sourceDriver.Close(), databaseDriver.Close()
		return nil, gerror.Wrap(err, "create migrate instance")
	}
	m.Log = newGFLogger(ctx, true)
	return m, nil
}

func closeMigrate(m *migrate.Migrate) error {
	if m == nil {
		return nil
	}
	sourceErr, databaseErr := m.Close()
	if sourceErr != nil && databaseErr != nil {
		return gerror.Wrapf(sourceErr, "close migration source: %v", databaseErr)
	}
	if sourceErr != nil {
		return gerror.Wrap(sourceErr, "close migration source")
	}
	if databaseErr != nil {
		return gerror.Wrap(databaseErr, "close migration database")
	}
	return nil
}

func bindGracefulStop(ctx context.Context, m *migrate.Migrate) func() {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			select {
			case m.GracefulStop <- true:
			default:
			}
		case <-done:
		}
	}()
	return func() { close(done) }
}

func readVersion(m *migrate.Migrate) (version uint, dirty bool, err error) {
	version, dirty, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, err
}

func runUp(ctx context.Context, target migrationTarget, expectedVersion uint64) (uint64, bool, error) {
	m, err := newMigrate(ctx, target)
	if err != nil {
		return 0, false, err
	}
	defer func() {
		if err := closeMigrate(m); err != nil {
			g.Log().Warning(ctx, err)
		}
	}()
	stop := bindGracefulStop(ctx, m)
	defer stop()

	version, dirty, err := readVersion(m)
	if err != nil {
		return 0, false, gerror.Wrap(err, "read migration version")
	}
	if dirty {
		return uint64(version), true, gerror.Newf("database migration is dirty at version %d for schema %s", version, target.SchemaName)
	}
	if expectedVersion > 0 && uint64(version) > expectedVersion {
		return uint64(version), false, gerror.Newf("database schema %s version %d is newer than code version %d", target.SchemaName, version, expectedVersion)
	}

	err = m.Up()
	switch {
	case err == nil:
		g.Log().Infof(ctx, "migration completed for schema %s", target.SchemaName)
	case errors.Is(err, migrate.ErrNoChange):
		g.Log().Infof(ctx, "migration no change for schema %s", target.SchemaName)
	default:
		return uint64(version), false, gerror.Wrapf(err, "migration up failed for schema %s", target.SchemaName)
	}

	version, dirty, err = readVersion(m)
	if err != nil {
		return 0, false, gerror.Wrap(err, "read migration version after up")
	}
	return uint64(version), dirty, nil
}
