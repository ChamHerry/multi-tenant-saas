package service

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
)

type IAuth interface {
	Authenticate(ctx context.Context, r *ghttp.Request) (*AuthIdentity, error)
}

var localAuth IAuth

func Auth() IAuth {
	if localAuth == nil {
		panic("implement not found for interface IAuth")
	}
	return localAuth
}

func RegisterAuth(i IAuth) {
	localAuth = i
}
