package main

import (
	_ "repomind-temp/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	"repomind-temp/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
