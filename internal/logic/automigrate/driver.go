package automigrate

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func isPostgresType(dbType string) bool {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "pgsql", "postgres", "postgresql":
		return true
	default:
		return false
	}
}

func postgresTypeFromNode(node *gdb.ConfigNode) string {
	if node == nil {
		return ""
	}
	if node.Type != "" {
		return node.Type
	}
	if i := strings.Index(node.Link, ":"); i > 0 {
		return node.Link[:i]
	}
	return ""
}

func requirePostgresDefaultDB() (gdb.DB, error) {
	db := g.DB()
	node := db.GetConfig()
	if !isPostgresType(postgresTypeFromNode(node)) {
		return nil, gerror.Newf("automigrate only supports postgres/pgsql, got %q", postgresTypeFromNode(node))
	}
	return db, nil
}

func publicSQLDB(ctx context.Context) (*sql.DB, error) {
	db, err := requirePostgresDefaultDB()
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.Master()
	if err != nil {
		return nil, gerror.Wrap(err, "get goframe default master sql db")
	}
	if err = sqlDB.PingContext(ctx); err != nil {
		return nil, gerror.Wrap(err, "ping public database")
	}
	return sqlDB, nil
}
