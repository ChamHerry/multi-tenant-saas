// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"

	"github.com/google/uuid"
)

// EmailVerificationTokens is the golang structure for table email_verification_tokens.
type EmailVerificationTokens struct {
	Id        uuid.UUID `json:"id"         orm:"id"         description:""` //
	UserId    uuid.UUID `json:"user_id"    orm:"user_id"    description:""` //
	Email     string    `json:"email"      orm:"email"      description:""` //
	TokenHash string    `json:"token_hash" orm:"token_hash" description:""` //
	ExpiresAt time.Time `json:"expires_at" orm:"expires_at" description:""` //
	UsedAt    time.Time `json:"used_at"    orm:"used_at"    description:""` //
	CreatedAt time.Time `json:"created_at" orm:"created_at" description:""` //
}
