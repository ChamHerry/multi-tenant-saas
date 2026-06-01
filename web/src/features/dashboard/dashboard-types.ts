// web/src/features/dashboard/dashboard-types.ts

export interface TrendPoint {
  date: string
  count: number
}

export interface GrowthTrendPoint {
  date: string
  tenants: number
  users: number
}

export interface DistributionItem {
  label: string
  count: number
}

export interface AuditTypeItem {
  category: string
  percent: number
}

export interface ActivityItem {
  action: string
  resource_type: string
  user_display_name?: string
  created_at: string
  metadata?: Record<string, unknown>
}

export interface SystemHealth {
  api: string
  database: string
  migration_version: number
}

export interface KPIData {
  // 租户视图
  member_count?: number
  member_trend?: number
  pending_invitations?: number
  expiring_invitations?: number
  active_api_keys?: number
  expiring_api_keys?: number
  monthly_audit_events?: number
  audit_trend_percent?: number
  // 平台视图
  tenant_count?: number
  tenant_trend?: number
  user_count?: number
  user_trend?: number
  daily_audit_events?: number
  anomaly_events?: number
}

export interface DashboardSummary {
  scope: 'tenant' | 'platform'
  kpi: KPIData
  member_growth?: TrendPoint[]
  role_distribution?: DistributionItem[]
  audit_trend?: TrendPoint[]
  growth_trend?: GrowthTrendPoint[]
  tenant_size_distribution?: DistributionItem[]
  audit_type_distribution?: AuditTypeItem[]
  recent_activities: ActivityItem[]
  system_health: SystemHealth
}
