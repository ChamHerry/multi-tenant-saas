package apikey

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"repomind-temp/internal/logic/rbac"
	"repomind-temp/internal/service"
	"repomind-temp/utility/uuid"
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type sAPIKey struct{}

func init() {
	service.RegisterAPIKey(&sAPIKey{})
}

func (s *sAPIKey) Create(ctx context.Context, in service.CreateAPIKeyInput) (*service.CreatedAPIKey, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}
	if err := ensureActiveMembership(ctx, in.TenantID, in.UserID); err != nil {
		return nil, err
	}
	secret, err := apiKeySecret(ctx)
	if err != nil {
		return nil, err
	}
	rawKey, err := generateRawKey(in.TenantID)
	if err != nil {
		return nil, err
	}
	keyID := uuid.GenerateV4()
	if err = validateInternalID(keyID); err != nil {
		return nil, err
	}
	keyPrefix := rawKeyPrefix(rawKey)
	hash := hashRawKey(rawKey, secret)
	record, err := g.DB().GetOne(ctx, `
INSERT INTO public.api_keys(id, tenant_id, user_id, name, key_hash, key_prefix, scopes, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?::text[], NULLIF(?, '')::timestamptz, now())
RETURNING id, tenant_id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at, revoked_at`,
		keyID, in.TenantID, in.UserID, in.Name, hash, keyPrefix, textArrayLiteral(in.Scopes), expiresAtString(in.ExpiresAt))
	if err != nil {
		return nil, gerror.Wrap(err, "insert api key")
	}
	item, err := mapAPIKey(record)
	if err != nil {
		return nil, err
	}
	return &service.CreatedAPIKey{APIKey: *item, RawKey: rawKey}, nil
}

func (s *sAPIKey) List(ctx context.Context, tenantID string) ([]service.APIKey, error) {
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	result, err := g.DB().GetAll(ctx, `
SELECT id, tenant_id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at, revoked_at
FROM public.api_keys
WHERE tenant_id=?
ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, gerror.Wrap(err, "list api keys")
	}
	items := make([]service.APIKey, 0, len(result))
	for _, record := range result {
		item, err := mapAPIKey(record)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (s *sAPIKey) Revoke(ctx context.Context, tenantID, apiKeyID string) error {
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if err := validateInternalID(apiKeyID); err != nil {
		return err
	}
	result, err := g.DB().Exec(ctx, `
UPDATE public.api_keys
SET revoked_at=COALESCE(revoked_at, now())
WHERE tenant_id=? AND id=?`, tenantID, apiKeyID)
	if err != nil {
		return gerror.Wrap(err, "revoke api key")
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return gerror.Newf("api key %s not found in tenant %s", apiKeyID, tenantID)
	}
	return nil
}

func (s *sAPIKey) Authenticate(ctx context.Context, rawKey string) (*service.AuthIdentity, error) {
	if rawKey == "" || !strings.HasPrefix(rawKey, "rpm_") {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "invalid api key")
	}
	secret, err := apiKeySecret(ctx)
	if err != nil {
		return nil, err
	}
	hash := hashRawKey(rawKey, secret)
	record, err := g.DB().GetOne(ctx, `
SELECT ak.id, ak.tenant_id, ak.user_id, ak.scopes
FROM public.api_keys ak
JOIN public.users u ON u.id = ak.user_id
JOIN public.tenants t ON t.id = ak.tenant_id
JOIN public.tenant_memberships tm ON tm.tenant_id = ak.tenant_id AND tm.user_id = ak.user_id
WHERE ak.key_hash=?
  AND ak.revoked_at IS NULL
  AND (ak.expires_at IS NULL OR ak.expires_at > now())
  AND u.status='active'
  AND u.deleted_at IS NULL
  AND t.status='active'
  AND t.deleted_at IS NULL
  AND tm.status='active'
  AND tm.deleted_at IS NULL
LIMIT 1`, hash)
	if err != nil {
		return nil, gerror.Wrap(err, "select api key")
	}
	if record.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key is invalid, revoked, expired, or not attached to an active membership")
	}
	scopes := parseTextArray(record["scopes"].String())
	if len(scopes) == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key has no scopes")
	}
	apiKeyID := record["id"].String()
	_, _ = g.DB().Exec(ctx, `UPDATE public.api_keys SET last_used_at=now() WHERE id=?`, apiKeyID)
	return &service.AuthIdentity{
		UserID:   record["user_id"].String(),
		TenantID: record["tenant_id"].String(),
		Type:     "api_key",
		Scopes:   scopes,
		APIKeyID: apiKeyID,
	}, nil
}

func validateCreateInput(in service.CreateAPIKeyInput) error {
	if err := validateInternalID(in.TenantID); err != nil {
		return err
	}
	if err := validateInternalID(in.UserID); err != nil {
		return err
	}
	if strings.TrimSpace(in.Name) == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, "api key name is required")
	}
	if len(in.Scopes) == 0 {
		return gerror.NewCode(gcode.CodeMissingParameter, "api key scopes are required")
	}
	seen := map[string]struct{}{}
	for _, scope := range in.Scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			return gerror.NewCode(gcode.CodeInvalidParameter, "api key scope cannot be empty")
		}
		if !rbac.KnownPermission(scope) {
			return gerror.NewCodef(gcode.CodeInvalidParameter, "unknown api key scope %q", scope)
		}
		if _, ok := seen[scope]; ok {
			return gerror.NewCodef(gcode.CodeInvalidParameter, "duplicate api key scope %q", scope)
		}
		seen[scope] = struct{}{}
	}
	return nil
}

func ensureActiveMembership(ctx context.Context, tenantID, userID string) error {
	record, err := g.DB().GetOne(ctx, `
SELECT 1
FROM public.tenant_memberships tm
JOIN public.users u ON u.id = tm.user_id
JOIN public.tenants t ON t.id = tm.tenant_id
WHERE tm.tenant_id=? AND tm.user_id=?
  AND tm.status='active' AND tm.deleted_at IS NULL
  AND u.status='active' AND u.deleted_at IS NULL
  AND t.status='active' AND t.deleted_at IS NULL
LIMIT 1`, tenantID, userID)
	if err != nil {
		return gerror.Wrap(err, "select active api key membership")
	}
	if record.IsEmpty() {
		return gerror.Newf("active membership tenant=%s user=%s not found", tenantID, userID)
	}
	return nil
}

func apiKeySecret(ctx context.Context) (string, error) {
	secret := strings.TrimSpace(g.Cfg().MustGet(ctx, "auth.apiKey.secret", "").String())
	if secret == "" {
		return "", gerror.NewCode(gcode.CodeMissingConfiguration, "auth.apiKey.secret is required")
	}
	return secret, nil
}

func generateRawKey(tenantID string) (string, error) {
	if err := validateInternalID(tenantID); err != nil {
		return "", err
	}
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", gerror.Wrap(err, "generate api key random bytes")
	}
	shortTenant := strings.ReplaceAll(tenantID, "-", "")[:8]
	return "rpm_" + shortTenant + "_" + hex.EncodeToString(buf), nil
}

func rawKeyPrefix(rawKey string) string {
	parts := strings.Split(rawKey, "_")
	if len(parts) < 2 {
		return rawKey
	}
	return parts[0] + "_" + parts[1]
}

func hashRawKey(rawKey, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(rawKey))
	return hex.EncodeToString(mac.Sum(nil))
}

func mapAPIKey(record gdb.Record) (*service.APIKey, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	tenantID := record["tenant_id"].String()
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	userID := record["user_id"].String()
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	return &service.APIKey{
		ID:         id,
		TenantID:   tenantID,
		UserID:     userID,
		Name:       record["name"].String(),
		KeyPrefix:  record["key_prefix"].String(),
		Scopes:     parseTextArray(record["scopes"].String()),
		LastUsedAt: nullableTime(record["last_used_at"]),
		ExpiresAt:  nullableTime(record["expires_at"]),
		CreatedAt:  record["created_at"].Time(),
		RevokedAt:  nullableTime(record["revoked_at"]),
	}, nil
}

func textArrayLiteral(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.ReplaceAll(item, `\`, `\\`)
		item = strings.ReplaceAll(item, `"`, `\"`)
		quoted = append(quoted, `"`+item+`"`)
	}
	return `{` + strings.Join(quoted, `,`) + `}`
}

func expiresAtString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTextArray(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "{}" || value == "[]" {
		return nil
	}
	if strings.HasPrefix(value, "[") {
		var items []string
		if err := json.Unmarshal([]byte(value), &items); err == nil {
			return items
		}
	}
	value = strings.TrimPrefix(strings.TrimSuffix(value, "}"), "{")
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, ` "`)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
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

func validateInternalID(id string) error {
	if !internalIDPattern.MatchString(id) {
		return gerror.Newf("invalid internal uuid %q", id)
	}
	return nil
}
