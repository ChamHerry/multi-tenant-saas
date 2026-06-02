package cmd

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"multi-tenant-saas/internal/controller"
	_ "multi-tenant-saas/internal/logic"
	"multi-tenant-saas/internal/middleware"
	"multi-tenant-saas/internal/service"
	"multi-tenant-saas/utility/crypto"
	credis "multi-tenant-saas/utility/redis"
)

var (
	internalIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 1. Validate bootstrap config (DB + server + encryptionKey)
			if err = validateBootstrapConfig(ctx); err != nil {
				return err
			}

			// 2. Initialize encryption (reads encryptionKey from YAML)
			if err = crypto.InitEncryption(ctx); err != nil {
				return err
			}

			// 3. Run database migrations
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}

			// 4. Initialize Redis (non-fatal: app runs without Redis if unavailable)
			if err = credis.GetCacheManager().InitAdapter(ctx, "default"); err != nil {
				g.Log().Warningf(ctx, "[cmd] Redis init failed, config will query DB directly: %v", err)
			}

			// 5. Validate runtime auth config (reads from system_config table)
			if err = validateRuntimeAuthConfig(ctx); err != nil {
				return err
			}

			s := g.Server()
			configureStaticFileService(ctx, s)
			registerRoutes(s)
			startBackgroundJobs(ctx)
			s.Run()
			return nil
		},
	}
	TenantCreate = &gcmd.Command{
		Name:  "tenant-create",
		Usage: "tenant-create --name <name> --slug <slug> --owner-user <uuid> [--system]",
		Brief: "create tenant metadata and owner membership; owner-user is required unless --system is set",
		Arguments: []gcmd.Argument{
			{Name: "name", Brief: "tenant display name"},
			{Name: "slug", Brief: "tenant slug, lowercase letters/numbers/hyphen"},
			{Name: "owner-user", Brief: "owner user UUID"},
			{Name: "system", Brief: "allow ownerless system tenant"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			name, err := requiredOption(parser, "name")
			if err != nil {
				return err
			}
			slug, err := requiredOption(parser, "slug")
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			tenant, err := service.TenantProvision().CreateTenant(ctx, service.CreateTenantInput{
				Name:            name,
				Slug:            slug,
				OwnerUserID:     optionalOption(parser, "owner-user"),
				SystemOwnerless: optionBool(parser, "system"),
			})
			if err != nil {
				return err
			}
			g.Dump(tenant)
			return nil
		},
	}
	UserUpsert = &gcmd.Command{
		Name:  "user-upsert",
		Usage: "user-upsert --provider <github|gitlab|gitea|oidc|password> --auth-id <id> --email <email> [--name <display>]",
		Brief: "create or update a global user by identity",
		Arguments: []gcmd.Argument{
			{Name: "provider", Brief: "identity provider"},
			{Name: "auth-id", Brief: "identity id from provider"},
			{Name: "email", Brief: "primary email"},
			{Name: "name", Brief: "display name"},
			{Name: "email-verified", Brief: "whether provider verified email"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			provider, err := requiredOption(parser, "provider")
			if err != nil {
				return err
			}
			authID, err := requiredOption(parser, "auth-id")
			if err != nil {
				return err
			}
			email, err := requiredOption(parser, "email")
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			user, err := service.UserService().EnsureUserByIdentity(ctx, service.EnsureUserByIdentityInput{
				Provider:      provider,
				AuthID:        authID,
				Email:         email,
				EmailVerified: optionBool(parser, "email-verified"),
				DisplayName:   optionalOption(parser, "name"),
			})
			if err != nil {
				return err
			}
			g.Dump(user)
			return nil
		},
	}
	UserPasswordCreate = &gcmd.Command{
		Name:  "user-password-create",
		Usage: "user-password-create --email <email> --password <password> [--name <display>] [--email-verified true]",
		Brief: "create or update a password login user",
		Arguments: []gcmd.Argument{
			{Name: "email", Brief: "login email"},
			{Name: "password", Brief: "initial password"},
			{Name: "name", Brief: "display name"},
			{Name: "email-verified", Brief: "mark email as verified"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			email, err := requiredOption(parser, "email")
			if err != nil {
				return err
			}
			password, err := requiredOption(parser, "password")
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			user, err := service.PasswordAuth().CreatePasswordUser(ctx, service.CreatePasswordUserInput{
				Email:         email,
				Password:      password,
				DisplayName:   optionalOption(parser, "name"),
				EmailVerified: optionBool(parser, "email-verified"),
			})
			if err != nil {
				return err
			}
			g.Dump(user)
			return nil
		},
	}
	UserPasswordSet = &gcmd.Command{
		Name:  "user-password-set",
		Usage: "user-password-set --email <email> --password <password>",
		Brief: "reset an existing password user's password and revoke sessions",
		Arguments: []gcmd.Argument{
			{Name: "email", Brief: "login email"},
			{Name: "password", Brief: "new password"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			email, err := requiredOption(parser, "email")
			if err != nil {
				return err
			}
			password, err := requiredOption(parser, "password")
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			user, err := service.PasswordAuth().SetPassword(ctx, email, password)
			if err != nil {
				return err
			}
			g.Dump(user)
			return nil
		},
	}
	UserSessionRevoke = &gcmd.Command{
		Name:  "user-session-revoke",
		Usage: "user-session-revoke --user <uuid> [--session <uuid>]",
		Brief: "revoke a specific auth session or all sessions of a user",
		Arguments: []gcmd.Argument{
			{Name: "user", Brief: "user UUID"},
			{Name: "session", Brief: "session UUID"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			userID, err := userIDFromParser(parser)
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			if sessionID := optionalOption(parser, "session"); sessionID != "" {
				if !internalIDPattern.MatchString(sessionID) {
					return gerror.Newf("invalid session uuid %q", sessionID)
				}
				return service.AuthSessionService().Revoke(ctx, sessionID, "manual_revoke")
			}
			return service.AuthSessionService().RevokeUserSessions(ctx, userID, "", "manual_revoke")
		},
	}
	PlatformAdminGrant = &gcmd.Command{
		Name:  "platform-admin-grant",
		Usage: "platform-admin-grant --user <uuid> --role <super_admin|support|auditor> [--actor <uuid>]",
		Brief: "bootstrap or update a platform admin user",
		Arguments: []gcmd.Argument{
			{Name: "user", Brief: "target user UUID"},
			{Name: "role", Brief: "platform admin role"},
			{Name: "actor", Brief: "actor user UUID; defaults to target user for first bootstrap"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			userID, err := userIDFromParser(parser)
			if err != nil {
				return err
			}
			role, err := requiredOption(parser, "role")
			if err != nil {
				return err
			}
			actorID := optionalOption(parser, "actor")
			if actorID == "" {
				actorID = userID
			}
			if !internalIDPattern.MatchString(actorID) {
				return gerror.Newf("invalid actor uuid %q", actorID)
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			if err = service.PlatformAdminService().Grant(ctx, actorID, userID, role); err != nil {
				return err
			}
			g.Dump(g.Map{"ok": true, "user_id": userID, "role": role})
			return nil
		},
	}
	TenantLifecycleRun = &gcmd.Command{
		Name:  "tenant-lifecycle-run",
		Usage: "tenant-lifecycle-run [--limit 20]",
		Brief: "run due tenant lifecycle jobs such as delayed purge/export placeholders",
		Arguments: []gcmd.Argument{
			{Name: "limit", Brief: "max pending jobs", Default: "20"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			if err := service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			return service.TenantLifecycle().RunPendingJobs(ctx, parser.GetOpt("limit", 20).Int())
		},
	}

	TenantMemberAdd = &gcmd.Command{
		Name:  "tenant-member-add",
		Usage: "tenant-member-add --tenant <uuid> --user <uuid> --role <owner|admin|member|viewer> [--status active]",
		Brief: "add or update a user's membership in a tenant",
		Arguments: []gcmd.Argument{
			{Name: "tenant", Brief: "tenant UUID"},
			{Name: "user", Brief: "user UUID"},
			{Name: "role", Brief: "tenant role"},
			{Name: "status", Brief: "membership status", Default: "active"},
			{Name: "invited-by", Brief: "inviter user UUID"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			tenantID, err := tenantIDFromParser(parser)
			if err != nil {
				return err
			}
			userID, err := userIDFromParser(parser)
			if err != nil {
				return err
			}
			role, err := requiredOption(parser, "role")
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			membership, err := service.TenantMembershipService().AddMember(ctx, service.AddTenantMemberInput{
				TenantID:        tenantID,
				UserID:          userID,
				Role:            role,
				Status:          parser.GetOpt("status", "active").String(),
				InvitedByUserID: optionalOption(parser, "invited-by"),
			})
			if err != nil {
				return err
			}
			g.Dump(membership)
			return nil
		},
	}
	TenantMemberList = &gcmd.Command{
		Name:  "tenant-member-list",
		Usage: "tenant-member-list --tenant <uuid>",
		Brief: "list members in one tenant",
		Arguments: []gcmd.Argument{
			{Name: "tenant", Brief: "tenant UUID"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			tenantID, err := tenantIDFromParser(parser)
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			members, err := service.TenantMembershipService().ListTenantMembers(ctx, tenantID)
			if err != nil {
				return err
			}
			g.Dump(members)
			return nil
		},
	}
	UserTenantList = &gcmd.Command{
		Name:  "user-tenant-list",
		Usage: "user-tenant-list --user <uuid>",
		Brief: "list tenants a user belongs to",
		Arguments: []gcmd.Argument{
			{Name: "user", Brief: "user UUID"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			userID, err := userIDFromParser(parser)
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			tenants, err := service.TenantMembershipService().ListUserTenants(ctx, userID)
			if err != nil {
				return err
			}
			g.Dump(tenants)
			return nil
		},
	}
	TenantContextResolve = &gcmd.Command{
		Name:  "tenant-context-resolve",
		Usage: "tenant-context-resolve --user <uuid> --tenant <uuid-or-slug>",
		Brief: "resolve and validate an active tenant context for a user",
		Arguments: []gcmd.Argument{
			{Name: "user", Brief: "user UUID"},
			{Name: "tenant", Brief: "tenant UUID or slug"},
		},
		Func: func(ctx context.Context, parser *gcmd.Parser) error {
			userID, err := userIDFromParser(parser)
			if err != nil {
				return err
			}
			tenantSelector, err := requiredOption(parser, "tenant")
			if err != nil {
				return err
			}
			if err = service.AutoMigrate().Up(ctx); err != nil {
				return err
			}
			tenantCtx, err := service.TenantMembershipService().ResolveTenantContext(ctx, userID, tenantSelector)
			if err != nil {
				return err
			}
			g.Dump(tenantCtx)
			return nil
		},
	}
)

func registerRoutes(s *ghttp.Server) {
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.RequestContext)
		group.Middleware(middleware.HandlerResponse)
		controller.RegisterRootRoutes(group)
	})
	s.Group("/metrics", func(group *ghttp.RouterGroup) {
		group.GET("/", middleware.MetricsHandler)
	})
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.RequestContext)
		group.Middleware(middleware.HandlerResponse)
		group.Middleware(middleware.Tracing)
		group.Middleware(middleware.Metrics)
		controller.RegisterRoutes(group)
	})
}

func init() {
	if err := Main.AddCommand(TenantCreate, UserUpsert, UserPasswordCreate, UserPasswordSet, UserSessionRevoke, PlatformAdminGrant, TenantLifecycleRun, TenantMemberAdd, TenantMemberList, UserTenantList, TenantContextResolve); err != nil {
		panic(err)
	}
}

func requiredOption(parser *gcmd.Parser, name string) (string, error) {
	value := parser.GetOpt(name)
	if value == nil || value.String() == "" {
		return "", gerror.Newf("missing required option --%s", name)
	}
	return value.String(), nil
}

func tenantIDFromParser(parser *gcmd.Parser) (string, error) {
	value, err := requiredOption(parser, "tenant")
	if err != nil {
		return "", err
	}
	if !internalIDPattern.MatchString(value) {
		return "", gerror.Newf("invalid tenant uuid %q", value)
	}
	return value, nil
}

func userIDFromParser(parser *gcmd.Parser) (string, error) {
	value, err := requiredOption(parser, "user")
	if err != nil {
		return "", err
	}
	if !internalIDPattern.MatchString(value) {
		return "", gerror.Newf("invalid user uuid %q", value)
	}
	return value, nil
}

func optionalOption(parser *gcmd.Parser, name string) string {
	value := parser.GetOpt(name)
	if value == nil {
		return ""
	}
	return value.String()
}

func optionBool(parser *gcmd.Parser, name string) bool {
	value := parser.GetOpt(name)
	if value == nil {
		return false
	}
	if strings.TrimSpace(value.String()) == "" {
		return true
	}
	return value.Bool()
}

// validateBootstrapConfig checks only the minimal YAML configs needed before DB/Redis are available.
func validateBootstrapConfig(ctx context.Context) error {
	link := strings.TrimSpace(g.Cfg().MustGet(ctx, "database.default.link", "").String())
	if link == "" {
		return gerror.New("database.default.link is required in config.yaml")
	}
	addr := strings.TrimSpace(g.Cfg().MustGet(ctx, "server.address", "").String())
	if addr == "" {
		return gerror.New("server.address is required in config.yaml")
	}
	encKey := strings.TrimSpace(g.Cfg().MustGet(ctx, "encryptionKey", "").String())
	if encKey == "" {
		return gerror.New("encryptionKey is required in config.yaml")
	}
	return nil
}

func validateRuntimeAuthConfig(ctx context.Context) error {
	env := service.Config().GetString(ctx, "server.env", "local")
	devHeader := service.Config().GetBool(ctx, "auth.devHeader.enabled", false)
	if devHeader && env != "local" && env != "test" {
		return gerror.New("auth.devHeader.enabled is only allowed when server.env is local or test")
	}
	passwordEnabled := service.Config().GetBool(ctx, "auth.password.enabled", true)
	sessionSecret := service.Config().GetString(ctx, "auth.session.secret", "")
	if passwordEnabled && sessionSecret == "" {
		return gerror.New("auth.session.secret is required when password login is enabled")
	}
	if env != "local" && env != "test" {
		if strings.Contains(sessionSecret, "change-me") {
			return gerror.New("auth.session.secret must be changed outside local/test")
		}
		apiKeySecret := service.Config().GetString(ctx, "auth.apiKey.secret", "")
		if apiKeySecret == "" || strings.Contains(apiKeySecret, "change-me") {
			return gerror.New("auth.apiKey.secret must be configured outside local/test")
		}
	}
	return nil
}

// startBackgroundJobs launches background goroutines for periodic tasks
// such as expiring pending invitations and cleaning up expired verification tokens.
func startBackgroundJobs(ctx context.Context) {
	go func() {
		interval := service.Config().GetDuration(ctx, "invitation.autoExpireInterval", 5*time.Minute)
		if interval <= 0 {
			interval = 5 * time.Minute
		}
		g.Log().Infof(ctx, "[background] invitation auto-expire started with interval %s", interval)
		for {
			time.Sleep(interval)
			expired, err := service.TenantInvitationService().ExpirePending(ctx, time.Now(), 100)
			if err != nil {
				g.Log().Warningf(ctx, "[background] invitation expire error: %v", err)
				continue
			}
			if expired > 0 {
				g.Log().Infof(ctx, "[background] expired %d pending invitations", expired)
			}
		}
	}()

	// Clean up expired email verification tokens every hour.
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		g.Log().Infof(ctx, "[background] email verification token cleanup started with interval 1h")
		for range ticker.C {
			result, err := g.DB().Exec(ctx,
				"DELETE FROM email_verification_tokens WHERE expires_at < now() - INTERVAL '1 hour'")
			if err != nil {
				g.Log().Warningf(ctx, "[background] email verification token cleanup error: %v", err)
				continue
			}
			if n, _ := result.RowsAffected(); n > 0 {
				g.Log().Infof(ctx, "[background] cleaned up %d expired verification tokens", n)
			}
		}
	}()
}
