package user

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/service"
	"repomind-temp/utility/uuid"
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

	db := g.DB()
	existing, err := fetchUserByIdentity(ctx, db, in.Provider, in.AuthID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if err = touchIdentityLogin(ctx, db, existing.ID, in); err != nil {
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

	if err = db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO public.users(id, email, display_name, avatar_url, status, metadata, last_login_at, created_at, updated_at)
VALUES (?, ?, ?, ?, 'active', ?::jsonb, now(), now(), now())`,
			userID, in.Email, nullableString(in.DisplayName), nullableString(in.AvatarURL), metadata); err != nil {
			return gerror.Wrap(err, "insert public user")
		}
		if _, err := tx.Ctx(ctx).Exec(`
INSERT INTO public.user_identities(
    id, user_id, provider, auth_id, email, email_verified, raw_profile, last_login_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?::jsonb, now(), now(), now())`,
			identityID, userID, in.Provider, in.AuthID, nullableString(in.Email), in.EmailVerified, rawProfile); err != nil {
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
	record, err := g.DB().GetOne(ctx, `
SELECT id, email, display_name, avatar_url, status, last_login_at, metadata, created_at, updated_at
FROM public.users
WHERE id=? AND deleted_at IS NULL`, userID)
	if err != nil {
		return nil, gerror.Wrap(err, "select public user")
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("user %s not found", userID)
	}
	return mapUser(record)
}

func fetchUserByIdentity(ctx context.Context, db gdb.DB, provider, authID string) (*service.User, error) {
	record, err := db.GetOne(ctx, `
SELECT u.id, u.email, u.display_name, u.avatar_url, u.status, u.last_login_at, u.metadata, u.created_at, u.updated_at
FROM public.user_identities i
JOIN public.users u ON u.id = i.user_id
WHERE i.provider=? AND i.auth_id=? AND u.deleted_at IS NULL`, provider, authID)
	if err != nil {
		return nil, gerror.Wrap(err, "select user by identity")
	}
	if record.IsEmpty() {
		return nil, nil
	}
	return mapUser(record)
}

func touchIdentityLogin(ctx context.Context, db gdb.DB, userID string, in service.EnsureUserByIdentityInput) error {
	rawProfile, err := marshalJSON(defaultMap(in.RawProfile), "marshal raw identity profile")
	if err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `
UPDATE public.user_identities
SET email=COALESCE(NULLIF(?, ''), email),
    email_verified=?,
    raw_profile=?::jsonb,
    last_login_at=now(),
    updated_at=now()
WHERE provider=? AND auth_id=?`, in.Email, in.EmailVerified, rawProfile, in.Provider, in.AuthID); err != nil {
		return gerror.Wrap(err, "update user identity login")
	}
	if _, err = db.Exec(ctx, `
UPDATE public.users
SET email=COALESCE(NULLIF(?, ''), email),
    display_name=COALESCE(NULLIF(?, ''), display_name),
    avatar_url=COALESCE(NULLIF(?, ''), avatar_url),
    last_login_at=now(),
    updated_at=now()
WHERE id=? AND deleted_at IS NULL`, in.Email, in.DisplayName, in.AvatarURL, userID); err != nil {
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
	return &service.User{
		ID:          id,
		Email:       record["email"].String(),
		DisplayName: record["display_name"].String(),
		AvatarURL:   record["avatar_url"].String(),
		Status:      record["status"].String(),
		LastLoginAt: nullableTime(record["last_login_at"]),
		Metadata:    metadata,
		CreatedAt:   record["created_at"].Time(),
		UpdatedAt:   record["updated_at"].Time(),
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
