// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AuthLoginAttempts is the golang structure of table auth_login_attempts for DAO operations like Where/Data.
type AuthLoginAttempts struct {
	g.Meta        `orm:"table:auth_login_attempts, do:true"`
	Id            any //
	LoginKey      any //
	Ip            any //
	FailedCount   any //
	LockedUntil   any //
	LastFailedAt  any //
	LastSuccessAt any //
	CreatedAt     any //
	UpdatedAt     any //
}
