// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SystemConfig is the golang structure of table system_config for DAO operations like Where/Data.
type SystemConfig struct {
	g.Meta      `orm:"table:system_config, do:true"`
	Key         any // key
	Value       any // value
	ValueType   any // value_type
	Description any // description
	IsEncrypted any // is_encrypted
	CreatedAt   any // created_at
	UpdatedAt   any // updated_at
}
