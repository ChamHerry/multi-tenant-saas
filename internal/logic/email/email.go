package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/service"
)

func init() {
	service.RegisterEmail(&sEmail{})
}

type sEmail struct{}

func (s *sEmail) SendInvitation(ctx context.Context, in service.SendInvitationInput) error {
	host := service.Config().GetString(ctx, "email.smtp.host", "")
	if host == "" {
		g.Log().Infof(ctx, "[email] invitation to %s for tenant %s skipped (no SMTP configured)", in.ToEmail, in.TenantName)
		return nil
	}
	port := service.Config().GetInt(ctx, "email.smtp.port", 587)
	username := service.Config().GetString(ctx, "email.smtp.username", "")
	password := service.Config().GetString(ctx, "email.smtp.password", "")
	from := service.Config().GetString(ctx, "email.from", "noreply@example.com")

	body := buildInvitationBody(from, in)
	addr := fmt.Sprintf("%s:%d", host, port)
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	g.Log().Infof(ctx, "[email] sending invitation to %s for tenant %s via %s", in.ToEmail, in.TenantName, addr)
	return smtp.SendMail(addr, auth, from, []string{in.ToEmail}, []byte(body))
}

func (s *sEmail) SendEmailVerification(ctx context.Context, in service.SendEmailVerificationInput) error {
	host := service.Config().GetString(ctx, "email.smtp.host", "")
	if host == "" {
		g.Log().Infof(ctx, "[email] verification email to %s skipped (no SMTP configured) — verify URL: %s", in.ToEmail, in.VerifyURL)
		return nil
	}
	port := service.Config().GetInt(ctx, "email.smtp.port", 587)
	username := service.Config().GetString(ctx, "email.smtp.username", "")
	password := service.Config().GetString(ctx, "email.smtp.password", "")
	from := service.Config().GetString(ctx, "email.from", "noreply@example.com")

	body := buildVerificationBody(from, in)
	addr := fmt.Sprintf("%s:%d", host, port)
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	g.Log().Infof(ctx, "[email] sending verification email to %s via %s", in.ToEmail, addr)
	return smtp.SendMail(addr, auth, from, []string{in.ToEmail}, []byte(body))
}

func (s *sEmail) SendPasswordReset(ctx context.Context, in service.SendPasswordResetInput) error {
	host := service.Config().GetString(ctx, "email.smtp.host", "")
	if host == "" {
		g.Log().Infof(ctx, "[email] password reset to %s skipped (no SMTP) — reset URL: %s", in.ToEmail, in.ResetURL)
		return nil
	}
	port := service.Config().GetInt(ctx, "email.smtp.port", 587)
	username := service.Config().GetString(ctx, "email.smtp.username", "")
	password := service.Config().GetString(ctx, "email.smtp.password", "")
	from := service.Config().GetString(ctx, "email.from", "noreply@example.com")

	body := buildPasswordResetBody(from, in)
	addr := fmt.Sprintf("%s:%d", host, port)
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	g.Log().Infof(ctx, "[email] sending password reset to %s via %s", in.ToEmail, addr)
	return smtp.SendMail(addr, auth, from, []string{in.ToEmail}, []byte(body))
}

func (s *sEmail) SendPasswordChanged(ctx context.Context, in service.SendPasswordChangedInput) error {
	host := service.Config().GetString(ctx, "email.smtp.host", "")
	if host == "" {
		g.Log().Infof(ctx, "[email] password changed notification to %s skipped (no SMTP)", in.ToEmail)
		return nil
	}
	port := service.Config().GetInt(ctx, "email.smtp.port", 587)
	username := service.Config().GetString(ctx, "email.smtp.username", "")
	password := service.Config().GetString(ctx, "email.smtp.password", "")
	from := service.Config().GetString(ctx, "email.from", "noreply@example.com")

	body := buildPasswordChangedBody(from, in)
	addr := fmt.Sprintf("%s:%d", host, port)
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	g.Log().Infof(ctx, "[email] sending password changed notification to %s via %s", in.ToEmail, addr)
	return smtp.SendMail(addr, auth, from, []string{in.ToEmail}, []byte(body))
}

func buildPasswordResetBody(from string, in service.SendPasswordResetInput) string {
	appName := "SaaS Console"
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", in.ToEmail))
	b.WriteString(fmt.Sprintf("Subject: Password Reset Request — %s\r\n", appName))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	if in.UserName != "" {
		b.WriteString(fmt.Sprintf("Hello %s,\r\n\r\n", in.UserName))
	}
	b.WriteString("We received a request to reset your password.\r\n\r\n")
	b.WriteString(fmt.Sprintf("Click the link below to reset your password (valid for 1 hour):\r\n%s\r\n\r\n", in.ResetURL))
	b.WriteString("If you did not request this, please ignore this email and your password will remain unchanged.\r\n")
	b.WriteString(fmt.Sprintf("—\r\n%s Team\r\n", appName))
	return b.String()
}

func buildPasswordChangedBody(from string, in service.SendPasswordChangedInput) string {
	appName := "SaaS Console"
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", in.ToEmail))
	b.WriteString(fmt.Sprintf("Subject: Your Password Has Been Changed — %s\r\n", appName))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	if in.UserName != "" {
		b.WriteString(fmt.Sprintf("Hello %s,\r\n\r\n", in.UserName))
	}
	b.WriteString("Your password was recently changed.\r\n\r\n")
	b.WriteString("If you made this change, no further action is needed.\r\n")
	b.WriteString("If you did NOT change your password, please contact support immediately.\r\n")
	b.WriteString(fmt.Sprintf("\r\n—\r\n%s Team\r\n", appName))
	return b.String()
}

func buildVerificationBody(from string, in service.SendEmailVerificationInput) string {
	appName := "SaaS Console"
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", in.ToEmail))
	b.WriteString(fmt.Sprintf("Subject: Verify your email address for %s\r\n", appName))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString("Hello,\r\n")
	b.WriteString("\r\n")
	b.WriteString("Please verify your email address by clicking the link below:\r\n")
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("%s\r\n", in.VerifyURL))
	b.WriteString("\r\n")
	b.WriteString("This link will expire in 24 hours.\r\n")
	b.WriteString("\r\n")
	b.WriteString("If you didn't create an account, you can safely ignore this email.\r\n")
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("—\r\n%s Team\r\n", appName))
	return b.String()
}

func buildInvitationBody(from string, in service.SendInvitationInput) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", in.ToEmail))
	b.WriteString("Subject: You've been invited to join a workspace\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("%s has invited you to join %s as a %s.\r\n", in.InviterName, in.TenantName, in.Role))
	if in.AcceptURL != "" {
		b.WriteString(fmt.Sprintf("\r\nAccept the invitation:\r\n%s\r\n", in.AcceptURL))
	}
	return b.String()
}
