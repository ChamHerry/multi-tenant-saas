package automigrate

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-migrate/migrate/v4"

	"repomind-temp/internal/service"
	migrationfiles "repomind-temp/manifest/migration"
)

type sAutoMigrate struct{}

func init() {
	service.RegisterAutoMigrate(New())
}

func New() service.IAutoMigrate {
	return &sAutoMigrate{}
}

func (s *sAutoMigrate) Up(ctx context.Context) error {
	enabled := g.Cfg().MustGet(ctx, "automigrate.enabled", true).Bool()
	if !enabled {
		g.Log().Info(ctx, "automigrate disabled")
		return nil
	}
	expected, err := LatestMigrationVersion(migrationfiles.FS, migrationfiles.PublicDir)
	if err != nil {
		return err
	}
	sqlDB, err := publicSQLDB(ctx)
	if err != nil {
		return err
	}
	version, dirty, err := runUp(ctx, migrationTarget{
		Dir:             migrationfiles.PublicDir,
		SchemaName:      publicSchemaName,
		MigrationsTable: publicMigrationsTable,
		SQLDB:           sqlDB,
	}, expected)
	if err != nil {
		return err
	}
	if dirty {
		return gerror.Newf("public database migration is dirty at version %d", version)
	}
	return nil
}

func (s *sAutoMigrate) Status(ctx context.Context) (version uint, dirty bool, err error) {
	sqlDB, err := publicSQLDB(ctx)
	if err != nil {
		return 0, false, err
	}
	m, err := newMigrate(ctx, migrationTarget{
		Dir:             migrationfiles.PublicDir,
		SchemaName:      publicSchemaName,
		MigrationsTable: publicMigrationsTable,
		SQLDB:           sqlDB,
	})
	if err != nil {
		return 0, false, err
	}
	defer func() {
		if closeErr := closeMigrate(m); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	version, dirty, err = m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	return version, dirty, err
}
