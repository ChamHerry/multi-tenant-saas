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
