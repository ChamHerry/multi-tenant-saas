package dashboard

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"multi-tenant-saas/internal/service"
)

// SummaryResponse is the top-level dashboard response envelope.
type SummaryResponse struct {
	Scope            string             `json:"scope"`
	KPI              KPIData            `json:"kpi"`
	MemberGrowth     []TrendPoint       `json:"member_growth"`
	RoleDistribution []DistributionItem `json:"role_distribution"`
	AuditTrend       []TrendPoint       `json:"audit_trend"`
	GrowthTrend      []GrowthTrendPoint `json:"growth_trend"`
	TenantSizeDist   []DistributionItem `json:"tenant_size_dist"`
	AuditTypeDist    []AuditTypeItem    `json:"audit_type_dist"`
	RecentActivities []ActivityItem     `json:"recent_activities"`
	SystemHealth     SystemHealth       `json:"system_health"`
}

// KPIData holds all KPI counters. Fields omitted when empty (different scopes return different subsets).
type KPIData struct {
	// Tenant scope
	MemberCount        int `json:"member_count,omitempty"`
	PendingInvitations int `json:"pending_invitations,omitempty"`
	ActiveAPIKeys      int `json:"active_api_keys,omitempty"`
	MonthlyAuditEvents int `json:"monthly_audit_events,omitempty"`

	// Platform scope
	TenantCount      int `json:"tenant_count,omitempty"`
	UserCount        int `json:"user_count,omitempty"`
	DailyAuditEvents int `json:"daily_audit_events,omitempty"`

	// Shared / derived
	NewMembersThisMonth int `json:"new_members_this_month,omitempty"`
	NewTenantsThisMonth int `json:"new_tenants_this_month,omitempty"`
	NewUsersThisMonth   int `json:"new_users_this_month,omitempty"`
	ExpiringInvitations int `json:"expiring_invitations,omitempty"`
	ExpiringAPIKeys     int `json:"expiring_api_keys,omitempty"`
	AnomalyEvents       int `json:"anomaly_events,omitempty"`
	AuditTrendPercent   int `json:"audit_trend_percent,omitempty"`
}

// TrendPoint is a single data-point for time-series charts.
type TrendPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// GrowthTrendPoint is a dual-series point for platform growth charts.
type GrowthTrendPoint struct {
	Date    string `json:"date"`
	Tenants int    `json:"tenants"`
	Users   int    `json:"users"`
}

// DistributionItem is a label/count pair for pie/bar charts.
type DistributionItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// AuditTypeItem represents an audit category with its percentage.
type AuditTypeItem struct {
	Category string `json:"category"`
	Percent  int    `json:"percent"`
}

// ActivityItem is a single row in the recent-activity feed.
type ActivityItem struct {
	Action          string    `json:"action"`
	ResourceType    string    `json:"resource_type"`
	UserDisplayName string    `json:"user_display_name"`
	CreatedAt       time.Time `json:"created_at"`
	Metadata        g.Map     `json:"metadata"`
}

// SystemHealth reports the health status of backend dependencies.
type SystemHealth struct {
	API              string `json:"api"`
	Database         string `json:"database"`
	MigrationVersion int    `json:"migration_version"`
}

// ControllerV1 handles dashboard endpoints.
type ControllerV1 struct{}

// NewV1 returns a new ControllerV1 instance.
func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// Summary handles GET /api/v1/dashboard/summary.
func (c *ControllerV1) Summary(r *ghttp.Request) {
	ctx := r.GetCtx()
	bizCtx := service.BizCtx().Get(ctx)

	var (
		scope string
		resp  *SummaryResponse
		err   error
	)

	isPlatformAdmin := bizCtx.PlatformAdmin != nil && len(bizCtx.PlatformAdmin.Permissions) > 0
	hasTenant := bizCtx.Tenant != nil && bizCtx.Tenant.TenantID != ""

	if isPlatformAdmin {
		scope = "platform"
		resp, err = platformSummary(ctx)
	} else if hasTenant {
		scope = "tenant"
		resp, err = tenantSummary(ctx, bizCtx.Tenant.TenantID)
	} else {
		r.Response.WriteJsonExit(ghttp.DefaultHandlerResponse{
			Code:    gcode.CodeNotAuthorized.Code(),
			Message: "no tenant context or platform admin privileges",
		})
		return
	}

	if err != nil {
		r.Response.WriteJsonExit(ghttp.DefaultHandlerResponse{
			Code:    gcode.CodeInternalError.Code(),
			Message: err.Error(),
		})
		return
	}

	resp.Scope = scope

	health, healthErr := healthCheck(ctx)
	if healthErr != nil {
		resp.SystemHealth = SystemHealth{API: "ok", Database: "unavailable"}
	} else {
		resp.SystemHealth = health
	}

	r.Response.WriteJsonExit(ghttp.DefaultHandlerResponse{
		Code:    gcode.CodeOK.Code(),
		Message: "OK",
		Data:    resp,
	})
}

// ---------------------------------------------------------------------------
// Tenant-scoped summary
// ---------------------------------------------------------------------------

func tenantSummary(ctx context.Context, tenantID string) (*SummaryResponse, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	sevenDaysFromNow := now.AddDate(0, 0, 7)
	thirtyDaysFromNow := now.AddDate(0, 0, 30)

	resp := &SummaryResponse{
		GrowthTrend:    []GrowthTrendPoint{},
		TenantSizeDist: []DistributionItem{},
		AuditTypeDist:  []AuditTypeItem{},
	}

	// --- KPI: members ---
	memberCount, err := g.DB().Model("tenant_memberships").Ctx(ctx).
		Where("tenant_id=? AND status='active'", tenantID).
		Count()
	if err != nil {
		return nil, err
	}

	newMembersThisMonth, err := g.DB().Model("tenant_memberships").Ctx(ctx).
		Where("tenant_id=? AND status='active' AND created_at >= ?", tenantID, monthStart).
		Count()
	if err != nil {
		return nil, err
	}

	// --- KPI: invitations ---
	pendingInvitations, err := g.DB().Model("tenant_invitations").Ctx(ctx).
		Where("tenant_id=? AND status='pending'", tenantID).
		Count()
	if err != nil {
		return nil, err
	}

	expiringInvitations, err := g.DB().Model("tenant_invitations").Ctx(ctx).
		Where("tenant_id=? AND status='pending' AND expires_at IS NOT NULL AND expires_at <= ?", tenantID, sevenDaysFromNow).
		Count()
	if err != nil {
		return nil, err
	}

	// --- KPI: api keys ---
	activeAPIKeys, err := g.DB().Model("api_keys").Ctx(ctx).
		Where("tenant_id=? AND revoked_at IS NULL", tenantID).
		Count()
	if err != nil {
		return nil, err
	}

	expiringAPIKeys, err := g.DB().Model("api_keys").Ctx(ctx).
		Where("tenant_id=? AND revoked_at IS NULL AND expires_at IS NOT NULL AND expires_at <= ?", tenantID, thirtyDaysFromNow).
		Count()
	if err != nil {
		return nil, err
	}

	// --- KPI: audit ---
	monthlyAuditEvents, err := g.DB().Model("audit_logs").Ctx(ctx).
		Where("tenant_id=? AND created_at >= ?", tenantID, monthStart).
		Count()
	if err != nil {
		return nil, err
	}

	lastMonthStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	lastMonthCount, err := g.DB().Model("audit_logs").Ctx(ctx).
		Where("tenant_id=? AND created_at >= ? AND created_at < ?", tenantID, lastMonthStart, monthStart).
		Count()
	if err != nil {
		return nil, err
	}

	auditTrendPercent := 0
	if lastMonthCount > 0 {
		auditTrendPercent = int(float64(monthlyAuditEvents-lastMonthCount) / float64(lastMonthCount) * 100)
	}

	resp.KPI = KPIData{
		MemberCount:         memberCount,
		NewMembersThisMonth: newMembersThisMonth,
		PendingInvitations:  pendingInvitations,
		ExpiringInvitations: expiringInvitations,
		ActiveAPIKeys:       activeAPIKeys,
		ExpiringAPIKeys:     expiringAPIKeys,
		MonthlyAuditEvents:  monthlyAuditEvents,
		AuditTrendPercent:   auditTrendPercent,
	}

	// --- Member growth (30 days) ---
	memberGrowth, err := queryTrend(ctx, "tenant_memberships",
		"tenant_id=? AND status='active' AND created_at >= ?",
		tenantID, thirtyDaysAgo,
	)
	if err != nil {
		return nil, err
	}
	resp.MemberGrowth = memberGrowth

	// --- Role distribution ---
	roleRows, err := g.DB().GetAll(ctx,
		`SELECT role as label, COUNT(*) as count
		 FROM tenant_memberships
		 WHERE tenant_id=? AND status='active'
		 GROUP BY role
		 ORDER BY count DESC`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	roleDist := make([]DistributionItem, 0, len(roleRows))
	for _, row := range roleRows {
		roleDist = append(roleDist, DistributionItem{
			Label: row["label"].String(),
			Count: row["count"].Int(),
		})
	}
	resp.RoleDistribution = roleDist

	// --- Audit trend (30 days) ---
	auditTrend, err := queryTrend(ctx, "audit_logs",
		"tenant_id=? AND created_at >= ?",
		tenantID, thirtyDaysAgo,
	)
	if err != nil {
		return nil, err
	}
	resp.AuditTrend = auditTrend

	// --- Recent activities ---
	activities, err := queryRecentActivities(ctx,
		`SELECT al.action, al.resource_type, COALESCE(u.display_name,'') as user_display_name, al.created_at, al.metadata
		 FROM audit_logs al
		 LEFT JOIN users u ON al.user_id = u.id
		 WHERE al.tenant_id = ?
		 ORDER BY al.created_at DESC
		 LIMIT 5`,
		tenantID,
	)
	if err != nil {
		return nil, err
	}
	resp.RecentActivities = activities

	return resp, nil
}

// ---------------------------------------------------------------------------
// Platform-scoped summary
// ---------------------------------------------------------------------------

func platformSummary(ctx context.Context) (*SummaryResponse, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	twentyFourHoursAgo := now.Add(-24 * time.Hour)

	resp := &SummaryResponse{
		MemberGrowth:     []TrendPoint{},
		RoleDistribution: []DistributionItem{},
		AuditTrend:       []TrendPoint{},
	}

	// --- KPI: tenants ---
	tenantCount, err := g.DB().Model("tenants").Ctx(ctx).
		Where("status='active'").
		Count()
	if err != nil {
		return nil, err
	}

	newTenantsThisMonth, err := g.DB().Model("tenants").Ctx(ctx).
		Where("status='active' AND created_at >= ?", monthStart).
		Count()
	if err != nil {
		return nil, err
	}

	// --- KPI: users ---
	userCount, err := g.DB().Model("users").Ctx(ctx).
		Where("status='active'").
		Count()
	if err != nil {
		return nil, err
	}

	newUsersThisMonth, err := g.DB().Model("users").Ctx(ctx).
		Where("status='active' AND created_at >= ?", monthStart).
		Count()
	if err != nil {
		return nil, err
	}

	// --- KPI: invitations (global) ---
	pendingInvitations, err := g.DB().Model("tenant_invitations").Ctx(ctx).
		Where("status='pending'").
		Count()
	if err != nil {
		return nil, err
	}

	// --- KPI: audit ---
	dailyAuditEvents, err := g.DB().Model("audit_logs").Ctx(ctx).
		Where("created_at >= ?", twentyFourHoursAgo).
		Count()
	if err != nil {
		return nil, err
	}

	anomalyEvents, err := g.DB().Model("audit_logs").Ctx(ctx).
		Where("created_at >= ? AND (action LIKE 'security.%' OR action = 'auth.login_failed')", twentyFourHoursAgo).
		Count()
	if err != nil {
		return nil, err
	}

	resp.KPI = KPIData{
		TenantCount:         tenantCount,
		NewTenantsThisMonth: newTenantsThisMonth,
		UserCount:           userCount,
		NewUsersThisMonth:   newUsersThisMonth,
		PendingInvitations:  pendingInvitations,
		DailyAuditEvents:    dailyAuditEvents,
		AnomalyEvents:       anomalyEvents,
	}

	// --- Growth trend (30 days): tenants + users merged ---
	growthTrend, err := queryGrowthTrend(ctx, thirtyDaysAgo)
	if err != nil {
		return nil, err
	}
	resp.GrowthTrend = growthTrend

	// --- Tenant size distribution ---
	sizeRows, err := g.DB().GetAll(ctx,
		`SELECT CASE
			WHEN cnt <= 5  THEN '1-5'
			WHEN cnt <= 20 THEN '6-20'
			WHEN cnt <= 100 THEN '21-100'
			ELSE '101+'
		END as label, COUNT(*) as count
		FROM (
			SELECT tenant_id, COUNT(*) as cnt
			FROM tenant_memberships
			WHERE status='active'
			GROUP BY tenant_id
		) sub
		GROUP BY label
		ORDER BY count DESC`,
	)
	if err != nil {
		return nil, err
	}
	tenantSizeDist := make([]DistributionItem, 0, len(sizeRows))
	for _, row := range sizeRows {
		tenantSizeDist = append(tenantSizeDist, DistributionItem{
			Label: row["label"].String(),
			Count: row["count"].Int(),
		})
	}
	resp.TenantSizeDist = tenantSizeDist

	// --- Audit type distribution ---
	typeRows, err := g.DB().GetAll(ctx,
		`SELECT CASE
			WHEN action LIKE 'auth.%' THEN 'auth'
			WHEN action LIKE 'member.%' THEN 'member'
			WHEN action LIKE 'tenant.%' THEN 'tenant'
			WHEN action LIKE 'api_key.%' THEN 'api_key'
			WHEN action LIKE 'security.%' THEN 'security'
			ELSE 'other'
		END as category, COUNT(*) as count
		FROM audit_logs
		WHERE created_at >= ?
		GROUP BY category`,
		twentyFourHoursAgo,
	)
	if err != nil {
		return nil, err
	}
	totalAudit := 0
	for _, row := range typeRows {
		totalAudit += row["count"].Int()
	}
	auditTypeDist := make([]AuditTypeItem, 0, len(typeRows))
	for _, row := range typeRows {
		cnt := row["count"].Int()
		pct := 0
		if totalAudit > 0 {
			pct = cnt * 100 / totalAudit
		}
		auditTypeDist = append(auditTypeDist, AuditTypeItem{
			Category: row["category"].String(),
			Percent:  pct,
		})
	}
	resp.AuditTypeDist = auditTypeDist

	// --- Recent activities ---
	activities, err := queryRecentActivities(ctx,
		`SELECT al.action, al.resource_type, COALESCE(u.display_name,'') as user_display_name, al.created_at, al.metadata
		 FROM audit_logs al
		 LEFT JOIN users u ON al.user_id = u.id
		 ORDER BY al.created_at DESC
		 LIMIT 5`,
	)
	if err != nil {
		return nil, err
	}
	resp.RecentActivities = activities

	return resp, nil
}

// ---------------------------------------------------------------------------
// Health check
// ---------------------------------------------------------------------------

func healthCheck(ctx context.Context) (SystemHealth, error) {
	version, dirty, err := service.AutoMigrate().Status(ctx)
	if err != nil || dirty {
		return SystemHealth{}, err
	}

	if _, dbErr := g.DB().GetOne(ctx, "SELECT 1"); dbErr != nil {
		return SystemHealth{
			API:              "ok",
			Database:         dbErr.Error(),
			MigrationVersion: int(version),
		}, nil
	}

	return SystemHealth{
		API:              "ok",
		Database:         "ok",
		MigrationVersion: int(version),
	}, nil
}

// ---------------------------------------------------------------------------
// Shared query helpers
// ---------------------------------------------------------------------------

// queryTrend returns daily counts for a given table over a time range.
// The whereClause should include the time-bound placeholder and args must supply all values.
func queryTrend(ctx context.Context, table, whereClause string, args ...interface{}) ([]TrendPoint, error) {
	sql := `SELECT to_char(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count
		FROM ` + table + `
		WHERE ` + whereClause + `
		GROUP BY date
		ORDER BY date ASC`

	rows, err := g.DB().GetAll(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	points := make([]TrendPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, TrendPoint{
			Date:  row["date"].String(),
			Count: row["count"].Int(),
		})
	}
	return points, nil
}

// queryGrowthTrend returns daily new-tenant and new-user counts merged into GrowthTrendPoints.
func queryGrowthTrend(ctx context.Context, thirtyDaysAgo time.Time) ([]GrowthTrendPoint, error) {
	tenantRows, err := g.DB().GetAll(ctx,
		`SELECT to_char(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count
		 FROM tenants
		 WHERE status='active' AND created_at >= ?
		 GROUP BY date
		 ORDER BY date ASC`,
		thirtyDaysAgo,
	)
	if err != nil {
		return nil, err
	}

	userRows, err := g.DB().GetAll(ctx,
		`SELECT to_char(created_at, 'YYYY-MM-DD') as date, COUNT(*) as count
		 FROM users
		 WHERE status='active' AND created_at >= ?
		 GROUP BY date
		 ORDER BY date ASC`,
		thirtyDaysAgo,
	)
	if err != nil {
		return nil, err
	}

	// Merge tenant and user counts by date.
	dateMap := make(map[string]*GrowthTrendPoint)
	for _, row := range tenantRows {
		d := row["date"].String()
		if _, ok := dateMap[d]; !ok {
			dateMap[d] = &GrowthTrendPoint{Date: d}
		}
		dateMap[d].Tenants = row["count"].Int()
	}
	for _, row := range userRows {
		d := row["date"].String()
		if _, ok := dateMap[d]; !ok {
			dateMap[d] = &GrowthTrendPoint{Date: d}
		}
		dateMap[d].Users = row["count"].Int()
	}

	// Fill in all dates in the 30-day range so the series has no gaps.
	points := make([]GrowthTrendPoint, 0, 31)
	for i := 0; i <= 30; i++ {
		d := thirtyDaysAgo.AddDate(0, 0, i).Format("2006-01-02")
		pt := GrowthTrendPoint{Date: d}
		if existing, ok := dateMap[d]; ok {
			pt.Tenants = existing.Tenants
			pt.Users = existing.Users
		}
		points = append(points, pt)
	}

	return points, nil
}

// queryRecentActivities runs a join query against audit_logs + users and returns ActivityItem rows.
func queryRecentActivities(ctx context.Context, sql string, args ...interface{}) ([]ActivityItem, error) {
	rows, err := g.DB().GetAll(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	activities := make([]ActivityItem, 0, len(rows))
	for _, row := range rows {
		item := ActivityItem{
			Action:          row["action"].String(),
			ResourceType:    row["resource_type"].String(),
			UserDisplayName: row["user_display_name"].String(),
			CreatedAt:       row["created_at"].Time(),
		}

		metaStr := row["metadata"].String()
		if metaStr != "" && metaStr != "null" {
			var meta g.Map
			if err := json.Unmarshal([]byte(metaStr), &meta); err == nil {
				item.Metadata = meta
			}
		}
		if item.Metadata == nil {
			item.Metadata = g.Map{}
		}

		activities = append(activities, item)
	}

	return activities, nil
}
