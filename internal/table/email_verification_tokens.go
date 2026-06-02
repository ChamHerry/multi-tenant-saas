package table

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
)

// EmailVerificationTokens defines the fields of table "email_verification_tokens" with their properties.
var EmailVerificationTokens = map[string]*gdb.TableField{
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
	"email": {
		Index:   2,
		Name:    "email",
		Type:    "varchar(255)",
		Null:    false,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
	"token_hash": {
		Index:   3,
		Name:    "token_hash",
		Type:    "varchar(128)",
		Null:    false,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
	"expires_at": {
		Index:   4,
		Name:    "expires_at",
		Type:    "timestamptz",
		Null:    false,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
	"used_at": {
		Index:   5,
		Name:    "used_at",
		Type:    "timestamptz",
		Null:    true,
		Key:     "",
		Default: nil,
		Extra:   "",
		Comment: "",
	},
	"created_at": {
		Index:   6,
		Name:    "created_at",
		Type:    "timestamptz",
		Null:    false,
		Key:     "",
		Default: "now()",
		Extra:   "",
		Comment: "",
	},
}

func SetEmailVerificationTokensTableFields(ctx context.Context, db gdb.DB, schema ...string) error {
	return db.GetCore().SetTableFields(ctx, "email_verification_tokens", EmailVerificationTokens, schema...)
}
