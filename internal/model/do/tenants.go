// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// Tenants is the golang structure of table tenants for DAO operations like Where/Data.
type Tenants struct {
	g.Meta    `orm:"table:tenants, do:true"`
	Id        any         //
	Name      any         //
	Slug      any         //
	Status    any         //
	Metadata  *gjson.Json //
	CreatedAt any         //
	UpdatedAt any         //
	DeletedAt any         //
}
