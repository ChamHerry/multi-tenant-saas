// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AuthSessions is the golang structure of table auth_sessions for DAO operations like Where/Data.
type AuthSessions struct {
	g.Meta        `orm:"table:auth_sessions, do:true"`
	Id            any //
	UserId        any //
	SecretHash    any //
	CsrfHash      any //
	UserAgent     any //
	Ip            any //
	LastUsedAt    any //
	ExpiresAt     any //
	IdleExpiresAt any //
	RevokedAt     any //
	RevokeReason  any //
	CreatedAt     any //
	UpdatedAt     any //
}
