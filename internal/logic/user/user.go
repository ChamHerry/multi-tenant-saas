package user

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

var (
	internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	allowedProviders  = map[string]struct{}{
		"github":   {},
		"gitlab":   {},
		"gitea":    {},
		"oidc":     {},
		"password": {},
	}
)

type sUser struct{}

func init() {
	service.RegisterUser(&sUser{})
}

func (s *sUser) EnsureUserByIdentity(ctx context.Context, in service.EnsureUserByIdentityInput) (*service.User, error) {
	if err := validateIdentityInput(in); err != nil {
		return nil, err
	}

	existing, err := fetchUserByIdentity(ctx, in.Provider, in.AuthID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if err = touchIdentityLogin(ctx, existing.ID, in); err != nil {
			return nil, err
		}
		return s.GetUser(ctx, existing.ID)
	}
	if in.Email == "" {
		return nil, gerror.New("email is required when creating a new user identity")
	}

	userID := uuid.GenerateV4()
	identityID := uuid.GenerateV4()
	if err = validateInternalID(userID); err != nil {
		return nil, err
	}
	if err = validateInternalID(identityID); err != nil {
		return nil, err
	}
	metadata, err := marshalJSON(defaultMap(in.Metadata), "marshal user metadata")
	if err != nil {
		return nil, err
	}
	rawProfile, err := marshalJSON(defaultMap(in.RawProfile), "marshal raw identity profile")
	if err != nil {
		return nil, err
	}

	if err = dao.Users.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		userCols := dao.Users.Columns()
		if _, err := dao.Users.Ctx(ctx).TX(tx).Data(g.Map{
			userCols.Id:          userID,
			userCols.Email:       in.Email,
			userCols.DisplayName: nullableString(in.DisplayName),
			userCols.AvatarUrl:   nullableString(in.AvatarURL),
			userCols.Status:      "active",
			userCols.Metadata:    metadata,
			userCols.LastLoginAt: "now()",
			userCols.CreatedAt:   "now()",
			userCols.UpdatedAt:   "now()",
		}).Insert(); err != nil {
			return gerror.Wrap(err, "insert public user")
		}

		identCols := dao.UserIdentities.Columns()
		if _, err := dao.UserIdentities.Ctx(ctx).TX(tx).Data(g.Map{
			identCols.Id:            identityID,
			identCols.UserId:        userID,
			identCols.Provider:      in.Provider,
			identCols.AuthId:        in.AuthID,
			identCols.Email:         nullableString(in.Email),
			identCols.EmailVerified: in.EmailVerified,
			identCols.RawProfile:    rawProfile,
			identCols.LastLoginAt:   "now()",
			identCols.CreatedAt:     "now()",
			identCols.UpdatedAt:     "now()",
		}).Insert(); err != nil {
			return gerror.Wrap(err, "insert user identity")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return s.GetUser(ctx, userID)
}

func (s *sUser) GetUser(ctx context.Context, userID string) (*service.User, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	userCols := dao.Users.Columns()
	record, err := dao.Users.Ctx(ctx).
		Fields("users.*, ui.email_verified").
		LeftJoin("user_identities ui", "ui.user_id = users.id AND ui.provider = 'password'").
		Where(userCols.Id, userID).
		Where("users.deleted_at IS NULL").
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select public user")
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("user %s not found", userID)
	}
	return mapUser(record)
}

func (s *sUser) UpdateProfile(ctx context.Context, in service.UpdateProfileInput) (*service.User, error) {
	if err := validateInternalID(in.UserID); err != nil {
		return nil, err
	}
	data := g.Map{"updated_at": "now()"}
	if in.DisplayName != "" {
		data["display_name"] = in.DisplayName
	}
	if in.AvatarURL != "" {
		data["avatar_url"] = in.AvatarURL
	}

	userCols := dao.Users.Columns()
	_, err := dao.Users.Ctx(ctx).
		Where(userCols.Id, in.UserID).
		Where("deleted_at IS NULL").
		Data(data).
		Update()
	if err != nil {
		return nil, gerror.Wrap(err, "update user profile")
	}
	return s.GetUser(ctx, in.UserID)
}

// SetEmailVerified marks the email as verified for a given user identity.
func (s *sUser) SetEmailVerified(ctx context.Context, userID, email string) error {
	if err := validateInternalID(userID); err != nil {
		return err
	}
	identCols := dao.UserIdentities.Columns()
	_, err := dao.UserIdentities.Ctx(ctx).
		Where(identCols.UserId, userID).
		Where("lower("+identCols.Email+") = lower(?)", email).
		Data(g.Map{identCols.EmailVerified: true, identCols.UpdatedAt: "now()"}).
		Update()
	return gerror.Wrap(err, "set email verified")
}

// GetIdentityByEmail looks up a user identity by provider and email (case-insensitive).
func (s *sUser) GetIdentityByEmail(ctx context.Context, provider, email string) (*service.UserIdentity, error) {
	identCols := dao.UserIdentities.Columns()
	record, err := dao.UserIdentities.Ctx(ctx).
		Where(identCols.Provider, provider).
		Where("lower("+identCols.Email+") = lower(?)", email).
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select user identity by email")
	}
	if record.IsEmpty() {
		return nil, nil
	}
	identity := &service.UserIdentity{
		ID:            record[identCols.Id].String(),
		UserID:        record[identCols.UserId].String(),
		Provider:      record[identCols.Provider].String(),
		AuthID:        record[identCols.AuthId].String(),
		Email:         record[identCols.Email].String(),
		EmailVerified: record[identCols.EmailVerified].Bool(),
		LastLoginAt:   nullableTime(record[identCols.LastLoginAt]),
		CreatedAt:     record[identCols.CreatedAt].Time(),
		UpdatedAt:     record[identCols.UpdatedAt].Time(),
	}
	return identity, nil
}

func fetchUserByIdentity(ctx context.Context, provider, authID string) (*service.User, error) {
	identCols := dao.UserIdentities.Columns()
	userCols := dao.Users.Columns()
	record, err := dao.UserIdentities.Ctx(ctx).
		LeftJoin("users u", "u.id = user_identities.user_id").
		Fields("u.*").
		Where(identCols.Provider, provider).
		Where(identCols.AuthId, authID).
		Where("u.deleted_at IS NULL").
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select user by identity")
	}
	if record.IsEmpty() {
		return nil, nil
	}
	// The LeftJoin returns fields with table prefix; remap to expected keys
	remapped := gdb.Record{
		"id":            record[userCols.Id],
		"email":         record[userCols.Email],
		"display_name":  record[userCols.DisplayName],
		"avatar_url":    record[userCols.AvatarUrl],
		"status":        record[userCols.Status],
		"last_login_at": record[userCols.LastLoginAt],
		"metadata":      record[userCols.Metadata],
		"created_at":    record[userCols.CreatedAt],
		"updated_at":    record[userCols.UpdatedAt],
	}
	return mapUser(remapped)
}

func touchIdentityLogin(ctx context.Context, userID string, in service.EnsureUserByIdentityInput) error {
	rawProfile, err := marshalJSON(defaultMap(in.RawProfile), "marshal raw identity profile")
	if err != nil {
		return err
	}

	// Update user_identities
	identCols := dao.UserIdentities.Columns()
	identData := g.Map{
		identCols.RawProfile:  rawProfile,
		identCols.LastLoginAt: "now()",
		identCols.UpdatedAt:   "now()",
	}
	// Only set email_verified to true; never overwrite an existing true with false.
	if in.EmailVerified {
		identData[identCols.EmailVerified] = true
	}
	if in.Email != "" {
		identData[identCols.Email] = in.Email
	}
	_, err = dao.UserIdentities.Ctx(ctx).
		Where(identCols.Provider, in.Provider).
		Where(identCols.AuthId, in.AuthID).
		Data(identData).
		Update()
	if err != nil {
		return gerror.Wrap(err, "update user identity login")
	}

	// Update users
	userCols := dao.Users.Columns()
	userData := g.Map{
		userCols.LastLoginAt: "now()",
		userCols.UpdatedAt:   "now()",
	}
	if in.Email != "" {
		userData[userCols.Email] = in.Email
	}
	if in.DisplayName != "" {
		userData[userCols.DisplayName] = in.DisplayName
	}
	if in.AvatarURL != "" {
		userData[userCols.AvatarUrl] = in.AvatarURL
	}
	_, err = dao.Users.Ctx(ctx).
		Where(userCols.Id, userID).
		Where("deleted_at IS NULL").
		Data(userData).
		Update()
	if err != nil {
		return gerror.Wrap(err, "update public user login")
	}
	return nil
}

func validateIdentityInput(in service.EnsureUserByIdentityInput) error {
	if _, ok := allowedProviders[in.Provider]; !ok {
		return gerror.Newf("invalid identity provider %q", in.Provider)
	}
	if in.AuthID == "" {
		return gerror.New("auth id is required")
	}
	return nil
}

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.Newf("invalid internal uuid %q", id)
	}
	return nil
}

func mapUser(record gdb.Record) (*service.User, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	metadata := map[string]any{}
	if raw := record["metadata"].String(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &metadata)
	}
	emailVerified := false
	if ev := record["email_verified"]; ev != nil {
		emailVerified = ev.Bool()
	}
	return &service.User{
		ID:            id,
		Email:         record["email"].String(),
		DisplayName:   record["display_name"].String(),
		AvatarURL:     record["avatar_url"].String(),
		Status:        record["status"].String(),
		EmailVerified: emailVerified,
		LastLoginAt:   nullableTime(record["last_login_at"]),
		Metadata:      metadata,
		CreatedAt:     record["created_at"].Time(),
		UpdatedAt:     record["updated_at"].Time(),
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

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func defaultMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	return in
}

func marshalJSON(in map[string]any, message string) (string, error) {
	payload, err := json.Marshal(in)
	if err != nil {
		return "", gerror.Wrap(err, message)
	}
	return string(payload), nil
}
