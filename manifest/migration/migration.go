package migration

import "embed"

const (
	PublicDir = "postgres/public"
)

// FS contains all SQL migration files used by golang-migrate.
//
//go:embed postgres/public/*.sql
var FS embed.FS
