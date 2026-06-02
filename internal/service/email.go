package service

import "context"

// SendEmailVerificationInput carries the data needed to construct a verification email.
type SendEmailVerificationInput struct {
	ToEmail     string
	DisplayName string
	VerifyURL   string
}

// IEmail defines the contract for sending emails. Implementations may include
// SMTP, an external API (SendGrid / SES), or a no-op fallback for development.
type IEmail interface {
	// SendInvitation sends an invitation email to the given address.
	// It receives enough context to build the invitation link and
	// tenant information for the email body.
	SendInvitation(ctx context.Context, in SendInvitationInput) error
	// SendEmailVerification sends a verification email with a verification link.
	SendEmailVerification(ctx context.Context, in SendEmailVerificationInput) error
}

// SendInvitationInput carries the data needed to construct an invitation email.
type SendInvitationInput struct {
	ToEmail       string
	InviterName   string
	TenantName    string
	AcceptURL     string
	Role          string
	InvitationID  string
}

var localEmail IEmail

func Email() IEmail {
	if localEmail == nil {
		panic("implement not found for interface IEmail")
	}
	return localEmail
}

func RegisterEmail(i IEmail) {
	localEmail = i
}
