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

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
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
	var inviteeUserIDVal any
	if inviteeUserID != "" {
		inviteeUserIDVal = inviteeUserID
	}
	cols := dao.TenantInvitations.Columns()
	_, err = dao.TenantInvitations.Ctx(ctx).Data(g.Map{
		cols.Id:              invitationID,
		cols.TenantId:        in.TenantID,
		cols.InviteeEmail:    normalizeEmail(in.InviteeEmail),
		cols.InviteeUserId:   inviteeUserIDVal,
		cols.Role:            in.Role,
		cols.TokenHash:       hashToken(token),
		cols.InvitedByUserId: in.InvitedByUserID,
		cols.Message:         in.Message,
		cols.ExpiresAt:       expiresAt,
	}).Insert()
	if err != nil {
		return nil, gerror.Wrap(err, "insert tenant invitation")
	}
	record, err := dao.TenantInvitations.Ctx(ctx).Where(cols.Id, invitationID).One()
	if err != nil {
		return nil, gerror.Wrap(err, "select created invitation")
	}
	invitation, err := mapInvitation(record)
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{TenantID: in.TenantID, UserID: in.InvitedByUserID, Action: "tenant.invitation.create", ResourceType: "tenant_invitation", ResourceID: invitation.ID, Metadata: map[string]any{"email": invitation.InviteeEmail, "role": invitation.Role}})
	result := &service.CreatedTenantInvitation{Invitation: *invitation, Token: token, AcceptURL: acceptURL(ctx, token)}
	sendInvitationEmailAsync(ctx, in.InvitedByUserID, in.TenantID, result)
	return result, nil
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
	where := []string{"lower(tenant_invitations.invitee_email)=lower(?)"}
	args := []any{email}
	if status := strings.TrimSpace(filter.Status); status != "" {
		where = append(where, "tenant_invitations.status=?")
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
	err = dao.TenantInvitations.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Select invitation within transaction (no FOR UPDATE, transaction isolation handles consistency)
		invCols := dao.TenantInvitations.Columns()
		record, err := dao.TenantInvitations.Ctx(ctx).TX(tx).
			Where(invCols.TokenHash, hashToken(token)).
			Where(invCols.Status, "pending").
			One()
		if err != nil {
			return gerror.Wrap(err, "select invitation for accept")
		}
		if record.IsEmpty() {
			return gerror.NewCode(gcode.CodeNotFound, "invitation not found")
		}
		if time.Now().After(record[invCols.ExpiresAt].Time()) {
			_, _ = dao.TenantInvitations.Ctx(ctx).TX(tx).
				Where(invCols.Id, record[invCols.Id].String()).
				Where(invCols.Status, "pending").
				Data(g.Map{invCols.Status: "expired", invCols.UpdatedAt: "now()"}).
				Update()
			return gerror.NewCode(gcode.CodeInvalidParameter, "invitation is expired")
		}
		if normalizeEmail(record[invCols.InviteeEmail].String()) != normalizeEmail(userEmailValue) {
			return gerror.NewCode(gcode.CodeNotAuthorized, "invitation email does not match current user")
		}

		tenantID := record[invCols.TenantId].String()
		membershipID := uuid.GenerateV4()
		invitedBy := record[invCols.InvitedByUserId].String()
		role := record[invCols.Role].String()
		now := time.Now()

		// Upsert membership: Select existing, then Insert or Update
		memCols := dao.TenantMemberships.Columns()
		existing, err := dao.TenantMemberships.Ctx(ctx).TX(tx).
			Where(memCols.TenantId, tenantID).
			Where(memCols.UserId, userID).
			Where("deleted_at IS NULL").
			One()
		if err != nil {
			return gerror.Wrap(err, "select existing membership")
		}

		if existing.IsEmpty() {
			var invitedByVal any
			if invitedBy != "" {
				invitedByVal = invitedBy
			}
			_, err = dao.TenantMemberships.Ctx(ctx).TX(tx).Data(g.Map{
				memCols.Id:              membershipID,
				memCols.TenantId:        tenantID,
				memCols.UserId:          userID,
				memCols.Role:            role,
				memCols.Status:          "active",
				memCols.InvitedByUserId: invitedByVal,
				memCols.JoinedAt:        now,
				memCols.CreatedAt:       now,
				memCols.UpdatedAt:       now,
			}).Insert()
			if err != nil {
				return gerror.Wrap(err, "accept invitation membership")
			}
		} else {
			data := g.Map{
				memCols.Role:      role,
				memCols.Status:    "active",
				memCols.UpdatedAt: now,
			}
			if invitedBy != "" {
				data[memCols.InvitedByUserId] = invitedBy
			}
			if existing[memCols.Status].String() == "invited" {
				data[memCols.JoinedAt] = now
			}
			_, err = dao.TenantMemberships.Ctx(ctx).TX(tx).
				Where(memCols.TenantId, tenantID).
				Where(memCols.UserId, userID).
				Where("deleted_at IS NULL").
				Data(data).
				Update()
			if err != nil {
				return gerror.Wrap(err, "accept invitation membership")
			}
		}

		// Fetch the membership record
		memberRecord, err := dao.TenantMemberships.Ctx(ctx).TX(tx).
			Where(memCols.TenantId, tenantID).
			Where(memCols.UserId, userID).
			Where("deleted_at IS NULL").
			One()
		if err != nil {
			return gerror.Wrap(err, "select accepted membership")
		}
		membership, err = mapMembershipTx(memberRecord)
		if err != nil {
			return err
		}

		// Mark invitation as accepted (conditional WHERE prevents double-accept)
		_, err = dao.TenantInvitations.Ctx(ctx).TX(tx).
			Where(invCols.Id, record[invCols.Id].String()).
			Where(invCols.Status, "pending").
			Data(g.Map{
				invCols.Status:           "accepted",
				invCols.AcceptedByUserId: userID,
				invCols.AcceptedAt:       now,
				invCols.UpdatedAt:        now,
			}).Update()
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
	cols := dao.TenantInvitations.Columns()
	result, err := dao.TenantInvitations.Ctx(ctx).
		Where(cols.Id, invitationID).
		Where("lower("+cols.InviteeEmail+") = lower(?)", email).
		Where(cols.Status, "pending").
		Data(g.Map{cols.Status: "declined", cols.DeclinedAt: "now()", cols.UpdatedAt: "now()"}).
		Update()
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
	cols := dao.TenantInvitations.Columns()
	result, err := dao.TenantInvitations.Ctx(ctx).
		Where(cols.Id, invitationID).
		Where(cols.TenantId, tenantID).
		Where(cols.Status, "pending").
		Data(g.Map{cols.Status: "revoked", cols.RevokedAt: "now()", cols.UpdatedAt: "now()"}).
		Update()
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
	cols := dao.TenantInvitations.Columns()
	result, err := dao.TenantInvitations.Ctx(ctx).
		Where(cols.Id, invitationID).
		Where(cols.TenantId, tenantID).
		Where(cols.Status, "pending").
		Data(g.Map{
			cols.TokenHash: hashToken(token),
			cols.ExpiresAt: expiresAt,
			cols.ResentAt:  "now()",
			cols.UpdatedAt: "now()",
		}).
		Update()
	if err != nil {
		return nil, gerror.Wrap(err, "resend invitation")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, gerror.NewCode(gcode.CodeNotFound, "pending invitation not found")
	}
	record, err := dao.TenantInvitations.Ctx(ctx).Where(cols.Id, invitationID).One()
	if err != nil {
		return nil, gerror.Wrap(err, "select resent invitation")
	}
	invitation, err := mapInvitation(record)
	if err != nil {
		return nil, err
	}
	_ = service.Audit().Write(ctx, service.AuditLogInput{TenantID: tenantID, UserID: actorUserID, Action: "tenant.invitation.resend", ResourceType: "tenant_invitation", ResourceID: invitationID})
	created := &service.CreatedTenantInvitation{Invitation: *invitation, Token: token, AcceptURL: acceptURL(ctx, token)}
	sendInvitationEmailAsync(ctx, actorUserID, tenantID, created)
	return created, nil
}

func (s *sTenantInvitation) ExpirePending(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	cols := dao.TenantInvitations.Columns()

	// Select IDs of pending expired invitations, then Update by IDs
	records, err := dao.TenantInvitations.Ctx(ctx).
		Fields(cols.Id).
		Where(cols.Status, "pending").
		Where(cols.ExpiresAt+" < ?", now).
		OrderAsc(cols.ExpiresAt).
		Limit(limit).
		All()
	if err != nil {
		return 0, gerror.Wrap(err, "select pending invitations")
	}
	if len(records) == 0 {
		return 0, nil
	}

	ids := make([]string, 0, len(records))
	for _, r := range records {
		ids = append(ids, r[cols.Id].String())
	}

	result, err := dao.TenantInvitations.Ctx(ctx).
		Where(cols.Id+" IN(?)", ids).
		Where(cols.Status, "pending").
		Data(g.Map{cols.Status: "expired", cols.UpdatedAt: "now()"}).
		Update()
	if err != nil {
		return 0, gerror.Wrap(err, "expire pending invitations")
	}
	rows, _ := result.RowsAffected()
	return int(rows), nil
}

func listInvitations(ctx context.Context, where []string, args []any, filter service.TenantInvitationFilter) (*service.TenantInvitationList, error) {
	limit, offset := normalizeLimitOffset(filter.Limit, filter.Offset)
	whereSQL := strings.Join(where, " AND ")

	total, err := dao.TenantInvitations.Ctx(ctx).
		LeftJoin("tenants t", "t.id = tenant_invitations.tenant_id").
		Where(whereSQL, args...).
		Count()
	if err != nil {
		return nil, gerror.Wrap(err, "count invitations")
	}

	rows, err := dao.TenantInvitations.Ctx(ctx).
		LeftJoin("tenants t", "t.id = tenant_invitations.tenant_id").
		Fields("tenant_invitations.*, t.name AS tenant_name, t.slug AS tenant_slug").
		Where(whereSQL, args...).
		OrderDesc("tenant_invitations.created_at").
		Limit(limit).
		Offset(offset).
		All()
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
	return &service.TenantInvitationList{Invitations: items, Total: total}, nil
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
	cols := dao.Tenants.Columns()
	record, err := dao.Tenants.Ctx(ctx).
		Where(cols.Id, tenantID).
		Where(cols.Status, "active").
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return gerror.Wrap(err, "select active tenant")
	}
	if record.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotFound, "active tenant not found")
	}
	return nil
}

func ensureActiveInviter(ctx context.Context, tenantID, inviterUserID string) error {
	cols := dao.TenantMemberships.Columns()
	record, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.UserId, inviterUserID).
		Where(cols.Status, "active").
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return gerror.Wrap(err, "select active inviter membership")
	}
	if record.IsEmpty() {
		return gerror.NewCode(gcode.CodeNotAuthorized, "active inviter membership not found")
	}
	return nil
}

func ensureEmailNotActiveMember(ctx context.Context, tenantID, email string) error {
	// JOIN query to check by email
	record, err := dao.TenantMemberships.Ctx(ctx).
		LeftJoin("users u", "u.id = tenant_memberships.user_id").
		Where("tenant_memberships.tenant_id", tenantID).
		Where("lower(u.email) = lower(?)", email).
		Where("tenant_memberships.status", "active").
		Where("tenant_memberships.deleted_at IS NULL").
		One()
	if err != nil {
		return gerror.Wrap(err, "select active member by email")
	}
	if !record.IsEmpty() {
		return gerror.NewCode(gcode.CodeInvalidParameter, "invitee is already an active tenant member")
	}
	return nil
}

func userEmail(ctx context.Context, userID string) (string, error) {
	cols := dao.Users.Columns()
	value, err := dao.Users.Ctx(ctx).
		Where(cols.Id, userID).
		Where(cols.Status, "active").
		Where("deleted_at IS NULL").
		Value(cols.Email)
	if err != nil {
		return "", gerror.Wrap(err, "select user email")
	}
	if value.IsNil() || value.String() == "" {
		return "", gerror.NewCode(gcode.CodeNotFound, "active user not found")
	}
	return value.String(), nil
}

func findUserIDByEmail(ctx context.Context, email string) (string, error) {
	cols := dao.UserIdentities.Columns()
	value, err := dao.UserIdentities.Ctx(ctx).
		Where(cols.Provider, "password").
		Where("lower("+cols.Email+") = lower(?)", email).
		Value(cols.UserId)
	if err != nil {
		return "", gerror.Wrap(err, "select user by email")
	}
	if value.IsNil() || value.String() == "" {
		return "", nil
	}
	return value.String(), nil
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
	base := strings.TrimRight(service.Config().GetString(ctx, "web.baseUrl", ""), "/")
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

func mapMembershipTx(record gdb.Record) (*service.TenantMembership, error) {
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

// sendInvitationEmailAsync fires the invitation email in a goroutine so
// it never blocks or fails the HTTP response.
func sendInvitationEmailAsync(ctx context.Context, inviterUserID, tenantID string, created *service.CreatedTenantInvitation) {
	inviterName := inviterUserID
	if value, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, inviterUserID).Value(dao.Users.Columns().DisplayName); err == nil && !value.IsNil() {
		if name := value.String(); name != "" {
			inviterName = name
		}
	}
	tenantName := tenantID
	if value, err := dao.Tenants.Ctx(ctx).Where(dao.Tenants.Columns().Id, tenantID).Value(dao.Tenants.Columns().Name); err == nil && !value.IsNil() {
		tenantName = value.String()
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log().Warningf(ctx, "[email] panic sending invitation email: %v", r)
			}
		}()
		if err := service.Email().SendInvitation(ctx, service.SendInvitationInput{
			ToEmail:      created.Invitation.InviteeEmail,
			InviterName:  inviterName,
			TenantName:   tenantName,
			AcceptURL:    created.AcceptURL,
			Role:         created.Invitation.Role,
			InvitationID: created.Invitation.ID,
		}); err != nil {
			g.Log().Warningf(ctx, "[email] failed to send invitation to %s: %v", created.Invitation.InviteeEmail, err)
		}
	}()
}
