package tenantdb

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
)

type sTenantDB struct{}
type sTenantGraph struct{}

func init() {
	service.RegisterTenantDB(&sTenantDB{})
	service.RegisterTenantGraph(&sTenantGraph{})
}

func (s *sTenantDB) Read(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error {
	return publicTransaction(ctx, fn)
}

func (s *sTenantDB) Write(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error {
	return publicTransaction(ctx, fn)
}

func publicTransaction(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return fn(ctx, tx.Ctx(ctx))
	})
}

func (s *sTenantGraph) Query(ctx context.Context, cypher string, args map[string]any) (gdb.Result, error) {
	if len(args) > 0 {
		return nil, gerror.New("shared graph query parameters are not supported by the generic gateway; use a repository-scoped graph service")
	}
	if strings.Contains(cypher, "$$") {
		return nil, gerror.New("shared graph query cannot contain $$ delimiter")
	}
	return nil, gerror.New("generic tenant graph gateway is retired; query public relations/process tables through repository access policy")
}
