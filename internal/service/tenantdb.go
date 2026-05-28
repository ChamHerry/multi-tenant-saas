package service

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

type ITenantDB interface {
	Read(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error
	Write(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error
}

type ITenantGraph interface {
	Query(ctx context.Context, cypher string, args map[string]any) (gdb.Result, error)
}

var localTenantDB ITenantDB
var localTenantGraph ITenantGraph

func TenantDB() ITenantDB {
	if localTenantDB == nil {
		panic("implement not found for interface ITenantDB")
	}
	return localTenantDB
}

func RegisterTenantDB(i ITenantDB) {
	localTenantDB = i
}

func TenantGraph() ITenantGraph {
	if localTenantGraph == nil {
		panic("implement not found for interface ITenantGraph")
	}
	return localTenantGraph
}

func RegisterTenantGraph(i ITenantGraph) {
	localTenantGraph = i
}
