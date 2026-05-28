package invitation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
	"repomind-temp/utility/uuid"
)

var (
	internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	allowedRoles      = map[string]struct{}{"owner": {}, "admin": {}, "member": {}, "viewer": {}}
)

type sTenantInvitation struct{}

func init() {
	service.RegisterTenantInvitation(&sTenantInvitation{})
}

func (s *sTenantInvitation) Create(ctx context.Context, in service.CreateTenantInvitationInput) (*service.CreatedTenantInvitation, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}
	if err := ensureActiveTenant(ctx, in.TenantID); err != nil {
		return nil, err
	}
	if err := ensureActiveInviter(ctx, in.TenantID, in.InvitedByUserID); err != nil {
		return nil, err
	}
	if err := ensureEmailNotActiveMember(ctx, in.TenantID, in.InviteeEmail); err != nil {
		return nil, err
	}
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	invitationID := uuid.GenerateV4()
	if err = validateInternalID(invitationID); err != nil {
		return nil, err
	}
	expiresIn := in.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 7 * 24 * time.Hour
	}
	expiresAt := time.Now().Add(expiresIn)
	inviteeUserID, _ := findUserIDByEmail(ctx, in.InviteeEmail)
	record, err := g.DB().GetOne(ctx, `
INSERT INTO public.tenant_invitations(
    id, tenant_id, invitee_email, invitee_user_id, role, status, token_hash,
    invited_by_user_id, message, expires_at, created_at, updated_at
) VALUES (?, ?, ?, NULLIF(?, '')::uuid, ?, 'pending', ?, ?, ?, ?, now(), now())
RETURNING id, tenant_id, invitee_email, invitee_user_id, role, status, invited_by_user_id,
          accepted_by_user_id, message, expires_at, accepted_at, declined_at, revoked_at,
          resent_at, created_at, updated_at`,
		invitationID, in.TenantID, normalizeEmail(in.InviteeEmail), inviteeUserID, in.Role, hashToken(token), in.InvitedByUserID, in.Message, expiresAt)
	if err != nil {
		return nil, gerror.Wrap(err, "insert tenant invitation")
	}
	invitation, err := mapInvitation(record)
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{TenantID: in.TenantID, UserID: in.InvitedByUserID, Action: "tenant.invitation.create", ResourceType: "tenant_invitation", ResourceID: invitation.ID, Metadata: map[string]any{"email": invitation.InviteeEmail, "role": invitation.Role}})
	return &service.CreatedTenantInvitation{Invitation: *invitation, Token: token, AcceptURL: acceptURL(ctx, token)}, nil
}

func (s *sTenantInvitation) ListTenantInvitations(ctx context.Context, tenantID string, filter service.TenantInvitationFilter) (*service.TenantInvitationList, error) {
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	where := []string{"tenant_id=?"}
	args := []any{tenantID}
	if status := strings.TrimSpace(filter.Status); status != "" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	return listInvitations(ctx, where, args, filter)
}

func (s *sTenantInvitation) ListMyInvitations(ctx context.Context, userID string, email string, filter service.TenantInvitationFilter) (*service.TenantInvitationList, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	if email == "" {
		var err error
		email, err = userEmail(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	where := []string{"lower(i.invitee_email)=lower(?)"}
	args := []any{email}
	if status := strings.TrimSpace(filter.Status); status != "" {
		where = append(where, "i.status=?")
		args = append(args, status)
	}
	return listInvitations(ctx, where, args, filter)
}

func (s *sTenantInvitation) Accept(ctx context.Context, token string, userID string) (*service.TenantMembership, error) {
	if strings.TrimSpace(token) == "" {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "invitation token is required")
	}
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	userEmailValue, err := userEmail(ctx, userID)
	if err != nil {
		return nil, err
	}
	var membership *service.TenantMembership
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		record, err := tx.Ctx(ctx).GetOne(`
SELECT id, tenant_id, invitee_email, role, status, invited_by_user_id, expires_at
FROM public.tenant_invitations
WHERE token_hash=?
FOR UPDATE`, hashToken(token))
		if err != nil {
			return gerror.Wrap(err, "select invitation for accept")
		}
		if record.IsEmpty() {
			return gerror.NewCode(gcode.CodeNotFound, "invitation not found")
		}
		if record["status"].String() != "pending" {
			return gerror.NewCodef(gcode.CodeInvalidParameter, "invitation is %s", record["status"].String())
		}
		if time.Now().After(record["expires_at"].Time()) {
			_, _ = tx.Ctx(ctx).Exec(`UPDATE public.tenant_invitations SET status='expired', updated_at=now() WHERE id=?`, record["id"].String())
			return gerror.NewCode(gcode.CodeInvalidParameter, "invitation is expired")
		}
		if normalizeEmail(record["invitee_email"].String()) != normalizeEmail(userEmailValue) {
			return gerror.NewCode(gcode.CodeNotAuthorized, "invitation email does not match current user")
		}
		tenantID := record["tenant_id"].String()
		membershipID := uuid.GenerateV4()
		memberRecord, err := tx.Ctx(ctx).GetOne(`
INSERT INTO public.tenant_memberships(
    id, tenant_id, user_id, role, status, invited_by_user_id, joined_at, created_at, updated_at
) VALUES (?, ?, ?, ?, 'active', NULLIF(?, '')::uuid, now(), now(), now())
ON CONFLICT (tenant_id, user_id) DO UPDATE
SET role=EXCLUDED.role,
    status='active',
    invited_by_user_id=EXCLUDED.invited_by_user_id,
    joined_at=COALESCE(public.tenant_memberships.joined_at, now()),
    deleted_at=NULL,
    updated_at=now()
RETURNING id, tenant_id, user_id, role, status, invited_by_user_id, joined_at, created_at, updated_at`,
			membershipID, tenantID, userID, record["role"].String(), record["invited_by_user_id"].String())
		if err != nil {
			return gerror.Wrap(err, "accept invitation membership")
		}
		membership, err = mapMembership(memberRecord)
		if err != nil {
			return err
		}
		_, err = tx.Ctx(ctx).Exec(`
UPDATE public.tenant_invitations
SET status='accepted', invitee_user_id=?, accepted_by_user_id=?, accepted_at=now(), updated_at=now()
WHERE id=?`, userID, userID, record["id"].String())
		if err != nil {
			return gerror.Wrap(err, "mark invitation accepted")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{TenantID: membership.TenantID, UserID: userID, Action: "tenant.invitation.accept", ResourceType: "tenant_invitation", ResourceID: membership.ID, Metadata: map[string]any{"role": membership.Role}})
	return membership, nil
}

func (s *sTenantInvitation) Decline(ctx context.Context, invitationID string, userID string) error {
	if err := validateInternalID(invitationID); err != nil {
		return err
	}
	if err := validateInternalID(userID); err != nil {
		return err
	}
	email, err := userEmail(ctx, userID)
	if err != nil {
		return err
	}
	result, err := g.DB().Exec(ctx, `
UPDATE public.tenant_invitations
SET status='declined', declined_at=now(), updated_at=now()
WHERE id=? AND lower(invitee_email)=lower(?) AND status='pending'`, invitationID, email)
	if err != nil {
		return gerror.Wrap(err, "decline invitation")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.NewCode(gcode.CodeNotFound, "pending invitation not found")
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{UserID: userID, Action: "tenant.invitation.decline", ResourceType: "tenant_invitation", ResourceID: invitationID})
	return nil
}

func (s *sTenantInvitation) Revoke(ctx context.Context, tenantID string, invitationID string, actorUserID string) error {
	if err := validateTenantActor(tenantID, invitationID, actorUserID); err != nil {
		return err
	}
	result, err := g.DB().Exec(ctx, `
UPDATE public.tenant_invitations
SET status='revoked', revoked_at=now(), updated_at=now()
WHERE tenant_id=? AND id=? AND status='pending'`, tenantID, invitationID)
	if err != nil {
		return gerror.Wrap(err, "revoke invitation")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.NewCode(gcode.CodeNotFound, "pending invitation not found")
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.invitation.revoke", ResourceType: "tenant_invitation", ResourceID: invitationID})
	return nil
}

func (s *sTenantInvitation) Resend(ctx context.Context, tenantID string, invitationID string, actorUserID string) (*service.CreatedTenantInvitation, error) {
	if err := validateTenantActor(tenantID, invitationID, actorUserID); err != nil {
		return nil, err
	}
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	record, err := g.DB().GetOne(ctx, `
UPDATE public.tenant_invitations
SET token_hash=?, expires_at=?, resent_at=now(), updated_at=now()
WHERE tenant_id=? AND id=? AND status='pending'
RETURNING id, tenant_id, invitee_email, invitee_user_id, role, status, invited_by_user_id,
          accepted_by_user_id, message, expires_at, accepted_at, declined_at, revoked_at,
          resent_at, created_at, updated_at`, hashToken(token), expiresAt, tenantID, invitationID)
	if err != nil {
		return nil, gerror.Wrap(err, "resend invitation")
	}
	if record.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotFound, "pending invitation not found")
	}
	invitation, err := mapInvitation(record)
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.invitation.resend", ResourceType: "tenant_invitation", ResourceID: invitationID})
	return &service.CreatedTenantInvitation{Invitation: *invitation, Token: token, AcceptURL: acceptURL(ctx, token)}, nil
}

func (s *sTenantInvitation) ExpirePending(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	result, err := g.DB().Exec(ctx, `
WITH expired AS (
    SELECT id FROM public.tenant_invitations
    WHERE status='pending' AND expires_at < ?
    ORDER BY expires_at ASC
    LIMIT ?
)
UPDATE public.tenant_invitations i
SET status='expired', updated_at=now()
FROM expired
WHERE i.id=expired.id`, now, limit)
	if err != nil {
		return 0, gerror.Wrap(err, "expire pending invitations")
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func listInvitations(ctx context.Context, where []string, args []any, filter service.TenantInvitationFilter) (*service.TenantInvitationList, error) {
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL := strings.Join(where, " AND ")
	countRecord, err := g.DB().GetOne(ctx, `
SELECT count(*) AS total
FROM public.tenant_invitations i
JOIN public.tenants t ON t.id=i.tenant_id
WHERE `+whereSQL, args...)
	if err != nil {
		return nil, gerror.Wrap(err, "count invitations")
	}
	rowsArgs := append(append([]any{}, args...), limit, offset)
	rows, err := g.DB().GetAll(ctx, `
SELECT i.id, i.tenant_id, t.name AS tenant_name, t.slug AS tenant_slug,
       i.invitee_email, i.invitee_user_id, i.role, i.status, i.invited_by_user_id,
       i.accepted_by_user_id, i.message, i.expires_at, i.accepted_at, i.declined_at,
       i.revoked_at, i.resent_at, i.created_at, i.updated_at
FROM public.tenant_invitations i
JOIN public.tenants t ON t.id=i.tenant_id
WHERE `+whereSQL+`
ORDER BY i.created_at DESC
LIMIT ? OFFSET ?`, rowsArgs...)
	if err != nil {
		return nil, gerror.Wrap(err, "list invitations")
	}
	items := make([]service.TenantInvitation, 0, len(rows))
	for _, row := range rows {
		item, err := mapInvitation(row)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return &service.TenantInvitationList{Invitations: items, Total: countRecord["total"].Int()}, nil
}

func validateCreateInput(in service.CreateTenantInvitationInput) error {
	if err := validateInternalID(in.TenantID); err != nil {
		return err
	}
	if err := validateInternalID(in.InvitedByUserID); err != nil {
		return err
	}
	if _, ok := allowedRoles[in.Role]; !ok {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "invalid tenant role %q", in.Role)
	}
	if !validEmail(in.InviteeEmail) {
		return gerror.NewCode(gcode.CodeInvalidParameter, "valid invitee email is required")
	}
	return nil
}

func validateTenantActor(tenantID, invitationID, actorUserID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if err := validateInternalID(invitationID); err != nil {
		return err
	}
	return validateInternalID(actorUserID)
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.NewCodef(gcode.CodeInvalidParameter, "invalid internal uuid %q", id)
	}
	return nil
}

func normalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	email = normalizeEmail(email)
	return len(email) >= 3 && strings.Contains(email, "@") && !strings.ContainsAny(email, " \t\n\r")
}

func ensureActiveTenant(ctx context.Context, tenantID string) error {
	record, err := g.DB().GetOne(ctx, `SELECT 1 FROM public.tenants WHERE id=? AND status='active' AND deleted_at IS NULL`, tenantID)
	if err != nil {
		return gerror.Wrap(err, "select active tenant")
	}
	if record.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotFound, "active tenant not found")
	}
	return nil
}

func ensureActiveInviter(ctx context.Context, tenantID, inviterUserID string) error {
	record, err := g.DB().GetOne(ctx, `
SELECT 1 FROM public.tenant_memberships
WHERE tenant_id=? AND user_id=? AND status='active' AND deleted_at IS NULL`, tenantID, inviterUserID)
	if err != nil {
		return gerror.Wrap(err, "select active inviter membership")
	}
	if record.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotAuthorized, "active inviter membership not found")
	}
	return nil
}

func ensureEmailNotActiveMember(ctx context.Context, tenantID, email string) error {
	record, err := g.DB().GetOne(ctx, `
SELECT 1
FROM public.tenant_memberships tm
JOIN public.users u ON u.id=tm.user_id
WHERE tm.tenant_id=? AND lower(u.email)=lower(?) AND tm.status='active' AND tm.deleted_at IS NULL AND u.deleted_at IS NULL
LIMIT 1`, tenantID, email)
	if err != nil {
		return gerror.Wrap(err, "select active member by email")
	}
	if !record.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "invitee is already an active tenant member")
	}
	return nil
}

func userEmail(ctx context.Context, userID string) (string, error) {
	record, err := g.DB().GetOne(ctx, `SELECT email FROM public.users WHERE id=? AND status='active' AND deleted_at IS NULL`, userID)
	if err != nil {
		return "", gerror.Wrap(err, "select user email")
	}
	if record.IsEmpty() {
		return "", gerror.NewCode(gcode.CodeNotFound, "active user not found")
	}
	return record["email"].String(), nil
}

func findUserIDByEmail(ctx context.Context, email string) (string, error) {
	record, err := g.DB().GetOne(ctx, `SELECT id FROM public.users WHERE lower(email)=lower(?) AND deleted_at IS NULL LIMIT 1`, email)
	if err != nil {
		return "", gerror.Wrap(err, "select user by email")
	}
	if record.IsEmpty() {
		return "", nil
	}
	return record["id"].String(), nil
}

func generateToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", gerror.Wrap(err, "generate invitation token")
	}
	return hex.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func acceptURL(ctx context.Context, token string) string {
	base := strings.TrimRight(g.Cfg().MustGet(ctx, "web.baseUrl", "").String(), "/")
	if base == "" {
		return ""
	}
	return base + "/me/invitations?token=" + token
}

func mapInvitation(record gdb.Record) (*service.TenantInvitation, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	tenantID := record["tenant_id"].String()
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	return &service.TenantInvitation{
		ID:               id,
		TenantID:         tenantID,
		TenantName:       record["tenant_name"].String(),
		TenantSlug:       record["tenant_slug"].String(),
		InviteeEmail:     record["invitee_email"].String(),
		InviteeUserID:    record["invitee_user_id"].String(),
		Role:             record["role"].String(),
		Status:           record["status"].String(),
		InvitedByUserID:  record["invited_by_user_id"].String(),
		AcceptedByUserID: record["accepted_by_user_id"].String(),
		Message:          record["message"].String(),
		ExpiresAt:        record["expires_at"].Time(),
		AcceptedAt:       nullableTime(record["accepted_at"]),
		DeclinedAt:       nullableTime(record["declined_at"]),
		RevokedAt:        nullableTime(record["revoked_at"]),
		ResentAt:         nullableTime(record["resent_at"]),
		CreatedAt:        record["created_at"].Time(),
		UpdatedAt:        record["updated_at"].Time(),
	}, nil
}

func mapMembership(record gdb.Record) (*service.TenantMembership, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	return &service.TenantMembership{
		ID:              id,
		TenantID:        record["tenant_id"].String(),
		UserID:          record["user_id"].String(),
		Role:            record["role"].String(),
		Status:          record["status"].String(),
		InvitedByUserID: record["invited_by_user_id"].String(),
		JoinedAt:        nullableTime(record["joined_at"]),
		CreatedAt:       record["created_at"].Time(),
		UpdatedAt:       record["updated_at"].Time(),
	}, nil
}

func nullableTime(value any) *time.Time {
	v, ok := value.(interface {
		IsNil() bool
		Time(...string) time.Time
	})
	if !ok || v.IsNil() {
		return nil
	}
	t := v.Time()
	return &t
}
