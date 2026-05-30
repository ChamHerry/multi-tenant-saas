package main

import (
	_ "multi-tenant-saas/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"multi-tenant-saas/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
