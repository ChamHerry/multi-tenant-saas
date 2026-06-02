package table

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

// PasswordResetTokens defines the fields of table "password_reset_tokens" with their properties.
var PasswordResetTokens = map[string]*gdb.TableField{
	"id": {
		Index:   0,
		Name:    "id",
		Type:    "uuid",
		Null:    false,
		Key:     "pri",
		Default: "gen_random_uuid()",
		Extra:   "",
		Comment: "",
	},
	"user_id": {
		Index:   1,
		Name:    "user_id",
		Type:    "uuid",
		Null:    false,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
	"token_hash": {
		Index:   2,
		Name:    "token_hash",
		Type:    "varchar(64)",
		Null:    false,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
	"expires_at": {
		Index:   3,
		Name:    "expires_at",
		Type:    "timestamptz",
		Null:    false,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
	"used_at": {
		Index:   4,
		Name:    "used_at",
		Type:    "timestamptz",
		Null:    true,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
	"created_at": {
		Index:   5,
		Name:    "created_at",
		Type:    "timestamptz",
		Null:    false,
		Key:     "",
		Default: "now()",
		Extra:   "",
		Comment: "",
	},
	"requested_ip": {
		Index:   6,
		Name:    "requested_ip",
		Type:    "varchar(45)",
		Null:    true,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
}

func SetPasswordResetTokensTableFields(ctx context.Context, db gdb.DB, schema ...string) error {
	return db.GetCore().SetTableFields(ctx, "password_reset_tokens", PasswordResetTokens, schema...)
}
