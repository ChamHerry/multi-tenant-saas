import { useTranslation } from 'react-i18next'
import {
  Activity,
  Building2,
  Inbox,
  KeyRound,
  Server,
  ShieldCheck,
  UserCog,
  Users,
} from 'lucide-react'
import { useAccess } from '@/features/access/access-hooks'
import { useDashboardSummary } from './dashboard-api'
import { KpiCard } from './KpiCard'
import { KpiCardRow } from './KpiCardRow'
import { TrendChart } from './TrendChart'
import { DistributionChart } from './DistributionChart'
import { RecentActivityList } from './RecentActivityList'
import { HealthPanel } from './HealthPanel'
import type { DashboardSummary } from './dashboard-types'

export function DashboardShell() {
  const { t } = useTranslation()
  const { data: summary, isLoading, isError, error } = useDashboardSummary()
  const { data: access } = useAccess()
  const isPlatformAdmin = (access?.platform_admin?.permissions?.length ?? 0) > 0

  if (isError) {
    return (
      <div className="rounded-card border border-danger/20 bg-white p-8 text-center">
        <p className="font-bold text-danger">{t('dashboard.errors.loadFailed')}</p>
        <p className="mt-1 text-sm text-muted">{String(error)}</p>
      </div>
    )
  }

  const scope = summary?.scope ?? 'tenant'

  return (
    <div className="space-y-5">
      <KpiCardRow>
        {scope === 'platform'
          ? renderPlatformKpi(summary!, isLoading, t)
          : renderTenantKpi(summary, isLoading, t)}
      </KpiCardRow>

      <div className="grid gap-4 lg:grid-cols-3">
        {scope === 'platform'
          ? renderPlatformCharts(summary!, isLoading, t)
          : renderTenantCharts(summary, isLoading, t)}
      </div>

      <div className="grid gap-4 lg:grid-cols-[1.5fr_1fr]">
        <RecentActivityList
          activities={summary?.recent_activities ?? []}
          auditLogLink={isPlatformAdmin ? '/admin/audit-logs' : '/tenant/audit-logs'}
          loading={isLoading}
        />
        <HealthPanel
          health={summary?.system_health ?? { api: 'loading', database: 'loading', migration_version: 0 }}
          quickActions={
            isPlatformAdmin
              ? [
                  { label: t('dashboard.platformActions.tenants'), to: '/admin/tenants', icon: <Building2 className="size-3.5" /> },
                  { label: t('dashboard.platformActions.users'), to: '/admin/users', icon: <UserCog className="size-3.5" /> },
                  { label: t('dashboard.platformActions.audit'), to: '/admin/audit-logs', icon: <ShieldCheck className="size-3.5" /> },
                ]
              : [
                  { label: t('dashboard.quickActions.invite'), to: '/tenant/access?tab=invitations', icon: <Inbox className="size-3.5" /> },
                  { label: t('dashboard.quickActions.apiKey'), to: '/tenant/api-keys', icon: <KeyRound className="size-3.5" /> },
                  { label: t('dashboard.quickActions.audit'), to: '/tenant/audit-logs', icon: <Activity className="size-3.5" /> },
                ]
          }
          loading={isLoading}
        />
      </div>
    </div>
  )
}

// Helper render functions (defined in same file to keep module boundaries simple)
function renderTenantKpi(summary: DashboardSummary | undefined, loading: boolean, t: ReturnType<typeof useTranslation>['t']) {
  return (
    <>
      <KpiCard icon={<Users className="size-4" />} label={t('dashboard.kpi.memberCount')}
        value={summary?.kpi.member_count ?? '-'}
        trend={summary?.kpi.member_trend ? t('dashboard.kpi.newThisMonth', { count: summary.kpi.member_trend }) : undefined}
        trendUp={summary?.kpi.member_trend ? summary.kpi.member_trend > 0 : undefined}
        tone="brand" to="/tenant/access?tab=members" loading={loading} />
      <KpiCard icon={<Inbox className="size-4" />} label={t('dashboard.kpi.pendingInvitations')}
        value={summary?.kpi.pending_invitations ?? '-'}
        trend={summary?.kpi.expiring_invitations ? t('dashboard.kpi.expiringSoon', { count: summary.kpi.expiring_invitations }) : undefined}
        trendUp={false}
        tone={(summary?.kpi.pending_invitations ?? 0) > 0 ? 'warning' : 'default'}
        to="/tenant/access?tab=invitations" loading={loading} />
      <KpiCard icon={<KeyRound className="size-4" />} label={t('dashboard.kpi.activeApiKeys')}
        value={summary?.kpi.active_api_keys ?? '-'}
        trend={summary?.kpi.expiring_api_keys ? t('dashboard.kpi.expiringSoon', { count: summary.kpi.expiring_api_keys }) : undefined}
        trendUp={false}
        to="/tenant/api-keys" loading={loading} />
      <KpiCard icon={<Activity className="size-4" />} label={t('dashboard.kpi.monthlyAuditEvents')}
        value={summary?.kpi.monthly_audit_events ?? '-'}
        trend={summary?.kpi.audit_trend_percent !== undefined ? `${summary.kpi.audit_trend_percent}%` : undefined}
        trendUp={summary?.kpi.audit_trend_percent ? summary.kpi.audit_trend_percent > 0 : undefined}
        to="/tenant/audit-logs" loading={loading} />
      <KpiCard icon={<Server className="size-4" />} label={t('dashboard.kpi.systemStatus')}
        value={summary?.system_health.api === 'ok' && summary?.system_health.database === 'ok' ? t('dashboard.health.ok') : '!'}
        tone={summary?.system_health.api === 'ok' && summary?.system_health.database === 'ok' ? 'success' : 'warning'}
        loading={loading} />
    </>
  )
}

function renderPlatformKpi(summary: DashboardSummary, loading: boolean, t: ReturnType<typeof useTranslation>['t']) {
  return (
    <>
      <KpiCard icon={<Building2 className="size-4" />} label={t('dashboard.kpi.tenantCount')}
        value={summary.kpi.tenant_count ?? '-'}
        trend={summary.kpi.tenant_trend ? t('dashboard.kpi.newThisMonth', { count: summary.kpi.tenant_trend }) : undefined}
        trendUp={summary.kpi.tenant_trend ? summary.kpi.tenant_trend > 0 : undefined}
        tone="brand" to="/admin/tenants" loading={loading} />
      <KpiCard icon={<Users className="size-4" />} label={t('dashboard.kpi.userCount')}
        value={summary.kpi.user_count ?? '-'}
        trend={summary.kpi.user_trend ? t('dashboard.kpi.newThisMonth', { count: summary.kpi.user_trend }) : undefined}
        trendUp={summary.kpi.user_trend ? summary.kpi.user_trend > 0 : undefined}
        to="/admin/users" loading={loading} />
      <KpiCard icon={<Inbox className="size-4" />} label={t('dashboard.kpi.pendingInvitations')}
        value={summary.kpi.pending_invitations ?? '-'}
        tone={(summary.kpi.pending_invitations ?? 0) > 0 ? 'warning' : 'default'}
        loading={loading} />
      <KpiCard icon={<Activity className="size-4" />} label={t('dashboard.kpi.dailyAuditEvents')}
        value={summary.kpi.daily_audit_events ?? '-'}
        trend={summary.kpi.anomaly_events ? t('dashboard.kpi.anomalyEvents', { count: summary.kpi.anomaly_events }) : undefined}
        trendUp={false}
        tone={(summary.kpi.anomaly_events ?? 0) > 0 ? 'warning' : 'default'}
        to="/admin/audit-logs" loading={loading} />
      <KpiCard icon={<Server className="size-4" />} label={t('dashboard.kpi.systemStatus')}
        value={summary.system_health.api === 'ok' && summary.system_health.database === 'ok' ? t('dashboard.health.ok') : '!'}
        tone={summary.system_health.api === 'ok' && summary.system_health.database === 'ok' ? 'success' : 'warning'}
        loading={loading} />
    </>
  )
}

function renderTenantCharts(summary: DashboardSummary | undefined, loading: boolean, t: ReturnType<typeof useTranslation>['t']) {
  return (
    <>
      <TrendChart title={t('dashboard.chart.memberGrowth')}
        data={summary?.member_growth ?? []}
        series={[{ key: 'count', color: '#1761ff', name: t('dashboard.chart.members') }]}
        loading={loading} emptyLabel={t('dashboard.empty.noMemberData')} />
      <DistributionChart title={t('dashboard.chart.roleDistribution')}
        data={summary?.role_distribution ?? []}
        loading={loading} emptyLabel={t('dashboard.empty.noMemberData')} />
      <TrendChart title={t('dashboard.chart.auditTrend')}
        data={summary?.audit_trend ?? []}
        series={[{ key: 'count', color: '#7c3aed', name: t('dashboard.chart.events') }]}
        loading={loading} emptyLabel={t('dashboard.empty.noAuditData')} />
    </>
  )
}

function renderPlatformCharts(summary: DashboardSummary, loading: boolean, t: ReturnType<typeof useTranslation>['t']) {
  return (
    <>
      <TrendChart title={t('dashboard.chart.growthTrend')}
        data={summary.growth_trend ?? []}
        series={[
          { key: 'tenants', color: '#1761ff', name: t('dashboard.chart.tenants') },
          { key: 'users', color: '#7c3aed', name: t('dashboard.chart.users') },
        ]}
        loading={loading} emptyLabel={t('dashboard.empty.noData')} />
      <DistributionChart title={t('dashboard.chart.tenantSize')}
        data={summary.tenant_size_distribution ?? []}
        loading={loading} emptyLabel={t('dashboard.empty.noData')} />
      <DistributionChart title={t('dashboard.chart.auditType')}
        data={(summary.audit_type_distribution ?? []).map((item) => ({
          label: item.category,
          count: item.percent,
        }))}
        loading={loading} emptyLabel={t('dashboard.empty.noAuditData')} />
    </>
  )
}
