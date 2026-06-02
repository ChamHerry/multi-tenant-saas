package service

import "context"

// CreateVerificationTokenInput carries the data needed to create and send a verification email.
type CreateVerificationTokenInput struct {
	UserID string
	Email  string
}

// VerifyEmailInput carries the token for email verification.
type VerifyEmailInput struct {
	Token string
}

// IEmailVerification defines the contract for email verification operations.
type IEmailVerification interface {
	// CreateAndSend creates a verification token and sends the verification email asynchronously.
	CreateAndSend(ctx context.Context, in CreateVerificationTokenInput) error
	// Verify validates the token and marks the email as verified.
	Verify(ctx context.Context, in VerifyEmailInput) error
	// ResendVerification resends the verification email with rate limiting.
	ResendVerification(ctx context.Context, email string) error
}

var localEmailVerification IEmailVerification

func EmailVerification() IEmailVerification {
	if localEmailVerification == nil {
		panic("implement not found for interface IEmailVerification")
	}
	return localEmailVerification
}

func RegisterEmailVerification(i IEmailVerification) {
	localEmailVerification = i
}
