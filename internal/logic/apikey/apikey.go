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

	"multi-tenant-saas/internal/dao"
	"multi-tenant-saas/internal/logic/rbac"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/uuid"
)

var internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type sAPIKey struct{}

func init() {
	service.RegisterAPIKey(&sAPIKey{})
}

func (s *sAPIKey) CreatePersonal(ctx context.Context, in service.CreatePersonalAPIKeyInput) (*service.CreatedAPIKey, error) {
	return s.createPersonal(ctx, in)
}

func (s *sAPIKey) createPersonal(ctx context.Context, in service.CreatePersonalAPIKeyInput) (*service.CreatedAPIKey, error) {
	input, err := validatePersonalCreateInput(in)
	if err != nil {
		return nil, err
	}
	if err = ensureActiveUser(ctx, input.UserID); err != nil {
		return nil, err
	}
	if err = enforcePersonalKeyLimit(ctx, input.UserID); err != nil {
		return nil, err
	}
	for _, grant := range input.Grants {
		if err = ensureActiveMembership(ctx, grant.TenantID, input.UserID); err != nil {
			return nil, err
		}
	}
	secret, err := apiKeySecret(ctx)
	if err != nil {
		return nil, err
	}
	rawKey, err := generateRawKey(input.UserID)
	if err != nil {
		return nil, err
	}
	keyID := uuid.GenerateV4()
	if err = validateInternalID(keyID); err != nil {
		return nil, err
	}
	keyPrefix := rawKeyPrefix(rawKey)
	hash := hashRawKey(rawKey, secret)
	var created *service.APIKey
	err = dao.ApiKeys.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Insert api_key, then Select back
		keyCols := dao.ApiKeys.Columns()
		var expiresAtVal any
		if input.ExpiresAt != nil {
			expiresAtVal = *input.ExpiresAt
		}
		_, err = dao.ApiKeys.Ctx(ctx).TX(tx).Data(g.Map{
			keyCols.Id:              keyID,
			keyCols.TenantId:        nil,
			keyCols.UserId:          input.UserID,
			keyCols.Name:            input.Name,
			keyCols.KeyHash:         hash,
			keyCols.KeyPrefix:       keyPrefix,
			keyCols.Scopes:          gdb.Raw("'" + textArrayLiteral(input.Scopes) + "'"),
			keyCols.ExpiresAt:       expiresAtVal,
			keyCols.KeyType:         service.APIKeyTypePersonal,
			keyCols.CreatedByUserId: input.UserID,
		}).Insert()
		if err != nil {
			return gerror.Wrap(err, "insert api key")
		}
		record, err := dao.ApiKeys.Ctx(ctx).TX(tx).Where(keyCols.Id, keyID).One()
		if err != nil {
			return gerror.Wrap(err, "select created api key")
		}
		item, err := mapAPIKey(record)
		if err != nil {
			return err
		}
		for _, grant := range input.Grants {
			grantID := uuid.GenerateV4()
			if err = validateInternalID(grantID); err != nil {
				return err
			}
			grantCols := dao.ApiKeyTenantGrants.Columns()
			_, err = dao.ApiKeyTenantGrants.Ctx(ctx).TX(tx).Data(g.Map{
				grantCols.Id:             grantID,
				grantCols.ApiKeyId:       keyID,
				grantCols.TenantId:       grant.TenantID,
				grantCols.Scopes:         gdb.Raw("'" + textArrayLiteral(grant.Scopes) + "'"),
				grantCols.GrantedByUserId: input.UserID,
			}).Insert()
			if err != nil {
				return gerror.Wrap(err, "insert api key tenant grant")
			}
			grantRecord, err := dao.ApiKeyTenantGrants.Ctx(ctx).TX(tx).Where(grantCols.Id, grantID).One()
			if err != nil {
				return gerror.Wrap(err, "select created api key tenant grant")
			}
			mappedGrant, err := mapAPIKeyGrant(grantRecord)
			if err != nil {
				return err
			}
			item.TenantGrants = append(item.TenantGrants, *mappedGrant)
		}
		created = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &service.CreatedAPIKey{APIKey: *created, RawKey: rawKey}, nil
}

func (s *sAPIKey) ListPersonal(ctx context.Context, userID, tenantID string) ([]service.APIKey, error) {
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	// List personal API keys that have grants for the given tenant.
	// Select only api_keys columns to avoid PG GROUP BY errors on joined grant columns.
	result, err := dao.ApiKeys.Ctx(ctx).
		Fields("api_keys.*").
		LeftJoin("api_key_tenant_grants akg", "akg.api_key_id = api_keys.id").
		Where("api_keys.user_id", userID).
		Where("akg.tenant_id", tenantID).
		Where("akg.status", "active").
		Where("akg.revoked_at IS NULL").
		Where("api_keys.revoked_at IS NULL").
		Group("api_keys.id").
		OrderDesc("api_keys.created_at").
		All()
	if err != nil {
		return nil, err
	}
	items := make([]service.APIKey, 0, len(result))
	for _, record := range result {
		item, err := mapAPIKey(record)
		if err != nil {
			return nil, err
		}
		grants, err := listKeyGrantsForTenant(ctx, item.ID, tenantID)
		if err != nil {
			return nil, err
		}
		item.TenantGrants = grants
		items = append(items, *item)
	}
	return items, nil
}

func (s *sAPIKey) RevokePersonal(ctx context.Context, userID, tenantID, apiKeyID string) error {
	if err := validateInternalID(userID); err != nil {
		return err
	}
	if err := validateInternalID(tenantID); err != nil {
		return err
	}
	if err := validateInternalID(apiKeyID); err != nil {
		return err
	}
	return dao.ApiKeys.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Check that an active grant exists for this key+tenant
		grantCols := dao.ApiKeyTenantGrants.Columns()
		grantRecord, err := dao.ApiKeyTenantGrants.Ctx(ctx).TX(tx).
			Where(grantCols.ApiKeyId, apiKeyID).
			Where(grantCols.TenantId, tenantID).
			Where(grantCols.Status, "active").
			Where(grantCols.RevokedAt+" IS NULL").
			One()
		if err != nil {
			return gerror.Wrap(err, "check api key tenant grant")
		}
		if grantRecord.IsEmpty() {
			return gerror.Newf("tenant-scoped api key %s not found for user %s in tenant %s", apiKeyID, userID, tenantID)
		}

		// Revoke the key if it belongs to the user
		keyCols := dao.ApiKeys.Columns()
		result, err := dao.ApiKeys.Ctx(ctx).TX(tx).
			Where(keyCols.Id, apiKeyID).
			Where(keyCols.UserId, userID).
			Where(keyCols.RevokedAt+" IS NULL").
			Data(g.Map{keyCols.RevokedAt: time.Now(), keyCols.UpdatedAt: time.Now()}).
			Update()
		if err != nil {
			return gerror.Wrap(err, "revoke tenant-scoped api key")
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return gerror.Newf("tenant-scoped api key %s not found for user %s in tenant %s", apiKeyID, userID, tenantID)
		}

		// Revoke all grants for this key
		_, err = dao.ApiKeyTenantGrants.Ctx(ctx).TX(tx).
			Where(grantCols.ApiKeyId, apiKeyID).
			Where(grantCols.RevokedAt+" IS NULL").
			Data(g.Map{
				grantCols.Status:         "revoked",
				grantCols.RevokedAt:      time.Now(),
				grantCols.RevokedByUserId: userID,
				grantCols.UpdatedAt:      time.Now(),
			}).Update()
		return gerror.Wrap(err, "revoke tenant-scoped api key grants")
	})
}

func (s *sAPIKey) ResolveTenantGrant(ctx context.Context, apiKeyID, tenantID string) (*service.APIKeyTenantGrant, error) {
	if err := validateInternalID(apiKeyID); err != nil {
		return nil, err
	}
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	grantCols := dao.ApiKeyTenantGrants.Columns()
	record, err := dao.ApiKeyTenantGrants.Ctx(ctx).
		LeftJoin("tenants t", "t.id = api_key_tenant_grants.tenant_id").
		LeftJoin("api_keys ak", "ak.id = api_key_tenant_grants.api_key_id").
		LeftJoin("users u", "u.id = ak.user_id").
		Fields("api_key_tenant_grants.*, t.name AS tenant_name, t.slug AS tenant_slug, u.id AS user_id, ak.name AS key_name, ak.key_prefix").
		Where(grantCols.ApiKeyId, apiKeyID).
		Where(grantCols.TenantId, tenantID).
		Where(grantCols.Status, "active").
		Where(grantCols.RevokedAt + " IS NULL").
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select api key tenant grant")
	}
	if record.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key is not granted to this tenant")
	}
	return mapAPIKeyGrant(record)
}

func (s *sAPIKey) Authenticate(ctx context.Context, rawKey string) (*service.AuthIdentity, error) {
	if rawKey == "" || !isSupportedRawKey(rawKey) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "invalid api key")
	}
	secret, err := apiKeySecret(ctx)
	if err != nil {
		return nil, err
	}
	hash := hashRawKey(rawKey, secret)
	cols := dao.ApiKeys.Columns()
	record, err := dao.ApiKeys.Ctx(ctx).
		Where(cols.KeyHash, hash).
		Where(cols.RevokedAt+" IS NULL").
		Where("(expires_at IS NULL OR expires_at > NOW())").
		One()
	if err != nil {
		return nil, gerror.Wrap(err, "select api key")
	}
	if record.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key is invalid, revoked, expired, or not attached to an active user")
	}
	// Verify user is active
	userCols := dao.Users.Columns()
	userRecord, err := dao.Users.Ctx(ctx).
		Where(userCols.Id, record[cols.UserId].String()).
		Where(userCols.Status, "active").
		Where("deleted_at IS NULL").
		One()
	if err != nil || userRecord.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key is invalid, revoked, expired, or not attached to an active user")
	}
	scopes := parseTextArray(record["scopes"].String())
	if len(scopes) == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "api key has no scopes")
	}
	apiKeyID := record["id"].String()
	_, _ = dao.ApiKeys.Ctx(ctx).Where(cols.Id, apiKeyID).Data(g.Map{cols.LastUsedAt: "now()"}).Update()
	return &service.AuthIdentity{
		UserID:   record["user_id"].String(),
		TenantID: "",
		Type:     "api_key",
		Scopes:   scopes,
		APIKeyID: apiKeyID,
		KeyType:  record["key_type"].String(),
	}, nil
}

func validatePersonalCreateInput(in service.CreatePersonalAPIKeyInput) (service.CreatePersonalAPIKeyInput, error) {
	if err := validateInternalID(in.UserID); err != nil {
		return in, err
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return in, gerror.NewCode(gcode.CodeMissingParameter, "api key name is required")
	}
	scopes, err := normalizeKeyScopes(in.Scopes)
	if err != nil {
		return in, err
	}
	in.Scopes = scopes
	for i, grant := range in.Grants {
		if err = validateInternalID(grant.TenantID); err != nil {
			return in, err
		}
		grantScopes, err := normalizeGrantScopes(grant.Scopes, scopes)
		if err != nil {
			return in, err
		}
		in.Grants[i].Scopes = grantScopes
	}
	return in, nil
}

func normalizeKeyScopes(scopes []string) ([]string, error) {
	items, err := normalizeScopes(scopes, false)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "api key scopes are required")
	}
	return items, nil
}

func normalizeGrantScopes(scopes []string, keyScopes []string) ([]string, error) {
	items, err := normalizeScopes(scopes, true)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return items, nil
	}
	for _, scope := range items {
		if !scopeIncludedInKey(scope, keyScopes) {
			return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "grant scope %q is not included in api key scopes", scope)
		}
	}
	return items, nil
}

func normalizeScopes(scopes []string, allowEmpty bool) ([]string, error) {
	seen := map[string]struct{}{}
	items := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			return nil, gerror.NewCode(gcode.CodeInvalidParameter, "api key scope cannot be empty")
		}
		if !rbac.KnownPermission(scope) {
			return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "unknown api key scope %q", scope)
		}
		if _, ok := seen[scope]; ok {
			return nil, gerror.NewCodef(gcode.CodeInvalidParameter, "duplicate api key scope %q", scope)
		}
		seen[scope] = struct{}{}
		items = append(items, scope)
	}
	if !allowEmpty && len(items) == 0 {
		return nil, gerror.NewCode(gcode.CodeMissingParameter, "api key scopes are required")
	}
	return items, nil
}

func scopeIncludedInKey(scope string, keyScopes []string) bool {
	for _, keyScope := range keyScopes {
		if keyScope == "*" || keyScope == scope {
			return true
		}
	}
	return false
}

func ensureActiveUser(ctx context.Context, userID string) error {
	cols := dao.Users.Columns()
	record, err := dao.Users.Ctx(ctx).
		Where(cols.Id, userID).
		Where(cols.Status, "active").
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return gerror.Wrap(err, "select active api key user")
	}
	if record.IsEmpty() {
		return gerror.Newf("active user %s not found", userID)
	}
	return nil
}

func ensureActiveMembership(ctx context.Context, tenantID, userID string) error {
	cols := dao.TenantMemberships.Columns()
	record, err := dao.TenantMemberships.Ctx(ctx).
		Where(cols.TenantId, tenantID).
		Where(cols.UserId, userID).
		Where(cols.Status, "active").
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return gerror.Wrap(err, "select active api key membership")
	}
	if record.IsEmpty() {
		return gerror.Newf("active membership tenant=%s user=%s not found", tenantID, userID)
	}
	return nil
}

func enforcePersonalKeyLimit(ctx context.Context, userID string) error {
	limit := service.Config().GetInt(ctx, "auth.apiKey.maxPersonalKeysPerUser", 10)
	if limit <= 0 {
		return nil
	}
	cols := dao.ApiKeys.Columns()
	count, err := dao.ApiKeys.Ctx(ctx).
		Where(cols.UserId, userID).
		Where(cols.KeyType, service.APIKeyTypePersonal).
		Where(cols.RevokedAt+" IS NULL").
		Where("(expires_at IS NULL OR expires_at > NOW())").
		Count()
	if err != nil {
		return err
	}
	if count >= limit {
		return gerror.NewCodef(gcode.CodeNotAuthorized, "personal api key limit exceeded: limit=%d", limit)
	}
	return nil
}

func listKeyGrants(ctx context.Context, apiKeyID string) ([]service.APIKeyTenantGrant, error) {
	grantCols := dao.ApiKeyTenantGrants.Columns()
	result, err := dao.ApiKeyTenantGrants.Ctx(ctx).
		LeftJoin("tenants t", "t.id = api_key_tenant_grants.tenant_id").
		LeftJoin("api_keys ak", "ak.id = api_key_tenant_grants.api_key_id").
		LeftJoin("users u", "u.id = ak.user_id").
		Fields("api_key_tenant_grants.*, t.name AS tenant_name, t.slug AS tenant_slug, u.id AS user_id, ak.name AS key_name, ak.key_prefix").
		Where(grantCols.ApiKeyId, apiKeyID).
		Where(grantCols.Status, "active").
		Where(grantCols.RevokedAt + " IS NULL").
		All()
	if err != nil {
		return nil, err
	}
	items := make([]service.APIKeyTenantGrant, 0, len(result))
	for _, record := range result {
		item, err := mapAPIKeyGrant(record)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func listKeyGrantsForTenant(ctx context.Context, apiKeyID, tenantID string) ([]service.APIKeyTenantGrant, error) {
	grantCols := dao.ApiKeyTenantGrants.Columns()
	result, err := dao.ApiKeyTenantGrants.Ctx(ctx).
		LeftJoin("tenants t", "t.id = api_key_tenant_grants.tenant_id").
		LeftJoin("api_keys ak", "ak.id = api_key_tenant_grants.api_key_id").
		LeftJoin("users u", "u.id = ak.user_id").
		Fields("api_key_tenant_grants.*, t.name AS tenant_name, t.slug AS tenant_slug, u.id AS user_id, ak.name AS key_name, ak.key_prefix").
		Where(grantCols.ApiKeyId, apiKeyID).
		Where(grantCols.TenantId, tenantID).
		Where(grantCols.Status, "active").
		Where(grantCols.RevokedAt + " IS NULL").
		All()
	if err != nil {
		return nil, err
	}
	items := make([]service.APIKeyTenantGrant, 0, len(result))
	for _, record := range result {
		item, err := mapAPIKeyGrant(record)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func apiKeySecret(ctx context.Context) (string, error) {
	secret := strings.TrimSpace(service.Config().GetString(ctx, "auth.apiKey.secret", ""))
	if secret == "" {
		return "", gerror.NewCode(gcode.CodeMissingConfiguration, "auth.apiKey.secret is required")
	}
	return secret, nil
}

func generateRawKey(subjectID string) (string, error) {
	if err := validateInternalID(subjectID); err != nil {
		return "", err
	}
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", gerror.Wrap(err, "generate api key random bytes")
	}
	shortSubject := strings.ReplaceAll(subjectID, "-", "")[:8]
	return "saas_" + shortSubject + "_" + hex.EncodeToString(buf), nil
}

func isSupportedRawKey(rawKey string) bool {
	return strings.HasPrefix(rawKey, "saas_") || strings.HasPrefix(rawKey, "st_") || strings.HasPrefix(rawKey, "rpm_")
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
	if tenantID != "" {
		if err := validateInternalID(tenantID); err != nil {
			return nil, err
		}
	}
	userID := record["user_id"].String()
	if err := validateInternalID(userID); err != nil {
		return nil, err
	}
	createdByUserID := record["created_by_user_id"].String()
	if createdByUserID != "" {
		if err := validateInternalID(createdByUserID); err != nil {
			return nil, err
		}
	}
	keyType := record["key_type"].String()
	if keyType == "" {
		keyType = service.APIKeyTypePersonal
	}
	return &service.APIKey{
		ID:              id,
		TenantID:        tenantID,
		UserID:          userID,
		Name:            record["name"].String(),
		KeyType:         keyType,
		KeyPrefix:       record["key_prefix"].String(),
		Scopes:          parseTextArray(record["scopes"].String()),
		LastUsedAt:      nullableTime(record["last_used_at"]),
		ExpiresAt:       nullableTime(record["expires_at"]),
		CreatedAt:       record["created_at"].Time(),
		RevokedAt:       nullableTime(record["revoked_at"]),
		CreatedByUserID: createdByUserID,
	}, nil
}

func mapAPIKeyGrant(record gdb.Record) (*service.APIKeyTenantGrant, error) {
	id := record["id"].String()
	if err := validateInternalID(id); err != nil {
		return nil, err
	}
	apiKeyID := record["api_key_id"].String()
	if err := validateInternalID(apiKeyID); err != nil {
		return nil, err
	}
	tenantID := record["tenant_id"].String()
	if err := validateInternalID(tenantID); err != nil {
		return nil, err
	}
	grantedByUserID := record["granted_by_user_id"].String()
	if grantedByUserID != "" {
		if err := validateInternalID(grantedByUserID); err != nil {
			return nil, err
		}
	}
	revokedByUserID := record["revoked_by_user_id"].String()
	if revokedByUserID != "" {
		if err := validateInternalID(revokedByUserID); err != nil {
			return nil, err
		}
	}
	return &service.APIKeyTenantGrant{
		ID:              id,
		APIKeyID:        apiKeyID,
		TenantID:        tenantID,
		TenantSlug:      record["tenant_slug"].String(),
		TenantName:      record["tenant_name"].String(),
		UserID:          record["user_id"].String(),
		KeyName:         record["key_name"].String(),
		KeyPrefix:       record["key_prefix"].String(),
		Scopes:          parseTextArray(record["scopes"].String()),
		Status:          record["status"].String(),
		GrantedByUserID: grantedByUserID,
		RevokedByUserID: revokedByUserID,
		CreatedAt:       record["created_at"].Time(),
		UpdatedAt:       record["updated_at"].Time(),
		RevokedAt:       nullableTime(record["revoked_at"]),
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
