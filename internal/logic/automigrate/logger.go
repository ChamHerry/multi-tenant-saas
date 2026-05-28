package automigrate

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type migrateLogger struct {
	ctx     context.Context
	verbose bool
}

func newGFLogger(ctx context.Context, verbose bool) migrateLogger {
	return migrateLogger{ctx: ctx, verbose: verbose}
}

func (l migrateLogger) Printf(format string, v ...any) {
	g.Log().Infof(l.ctx, "migrate: "+format, v...)
}

func (l migrateLogger) Verbose() bool {
	return l.verbose
}
