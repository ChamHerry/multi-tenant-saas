import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { Activity, ArrowRight, DatabaseZap, Server, Users } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { useMe, useMyTenants } from '@/features/auth/auth-hooks'
import { useTenantContext } from '@/features/tenants/tenant-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { apiRaw } from '@/shared/api/client'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { pageTitle, usePageTitle } from '@/layouts/page-title'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'

type Health = { ok: boolean; version?: number; dirty?: boolean; error?: string }

function SystemCard({ title, query, icon }: { title: string; query: ReturnType<typeof useQuery<Health>>; icon: ReactNode }) {
  const { t } = useTranslation()
  const statusLabel = query.data?.ok
    ? t('dashboard.system.status.ok')
    : query.isError
      ? t('dashboard.system.status.error')
      : t('dashboard.system.status.loading')
  const message = query.isLoading
    ? t('dashboard.system.messages.checking')
    : query.isError
      ? t('dashboard.system.messages.unavailable')
      : query.data?.version
        ? t('dashboard.system.messages.migrationVersion', { version: query.data.version })
        : t('dashboard.system.messages.healthy')

  return (
    <Card>
      <div className="flex items-center justify-between gap-3">
        <div className="rounded-panel bg-brand-soft p-2 text-brand">{icon}</div>
        <Badge tone={query.data?.ok ? 'green' : query.isError ? 'red' : 'orange'}>{statusLabel}</Badge>
      </div>
      <h3 className="mt-4 text-lg font-black text-ink">{title}</h3>
      <p className="mt-2 text-sm text-muted">{message}</p>
    </Card>
  )
}

export function DashboardPage() {
  const { t } = useTranslation()
  const health = useQuery<Health>({ queryKey: ['system', 'healthz'], queryFn: () => apiRaw<Health>('/healthz') })
  const ready = useQuery<Health>({
    queryKey: ['system', 'readyz'],
    queryFn: () => apiRaw<Health>('/readyz'),
    retry: false,
  })
  const me = useMe()
  const tenants = useMyTenants()
  const tenantContext = useTenantContext()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const displayName = me.data?.user.display_name || me.data?.user.email || t('dashboard.userFallback')
  const tenantCount = tenants.data?.tenants.length ?? 0

  usePageTitle(me.data ? pageTitle('dashboard.title.welcome', { name: displayName }) : pageTitle('dashboard.title.default'))

  if (me.isLoading) return <LoadingView label={t('common.loading')} />

  return (
    <div className="space-y-6">
      <section className="hero-gradient rounded-[28px] border border-white p-6 shadow-soft sm:p-8">
        <div className="flex flex-wrap gap-3">
          <Link to="/tenants"><Button>{tenantCount === 0 ? t('dashboard.hero.createTenant') : t('dashboard.hero.manageTenants')}</Button></Link>
          {tenantCount > 0 ? (
            <Link to="/tenant/access?tab=members">
              <Button variant="secondary">{t('dashboard.hero.openAccess')}</Button>
            </Link>
          ) : null}
        </div>
        {tenantCount === 0 ? (
          <div className="mt-5 rounded-card border border-warning/20 bg-white/80 p-4 text-sm leading-6 text-muted">
            <div className="font-bold text-ink">{t('dashboard.hero.noTenantsTitle')}</div>
            <p className="mt-1">{t('dashboard.hero.noTenantsDescription')}</p>
          </div>
        ) : null}
      </section>

      <section className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <SystemCard title={t('dashboard.system.healthTitle')} query={health} icon={<Server className="size-5" />} />
        <SystemCard title={t('dashboard.system.readyTitle')} query={ready} icon={<DatabaseZap className="size-5" />} />
        <Card>
          <div className="w-fit rounded-panel bg-accent-purple-soft p-2 text-accent-purple"><Users className="size-5" /></div>
          <h3 className="mt-4 text-lg font-black text-ink">{t('dashboard.cards.tenantsTitle')}</h3>
          <p className="mt-2 text-3xl font-black text-brand">{tenantCount}</p>
        </Card>
        <Card>
          <div className="w-fit rounded-panel bg-success-soft p-2 text-success-strong"><Activity className="size-5" /></div>
          <h3 className="mt-4 text-lg font-black text-ink">{t('dashboard.cards.currentRoleTitle')}</h3>
          <p className="mt-2 text-sm font-bold text-muted">
            {currentTenantId ? tenantContext.data?.tenant_context.role ?? t('dashboard.cards.resolvingRole') : t('dashboard.cards.noTenantSelected')}
          </p>
        </Card>
      </section>

      <section className="grid gap-4 lg:grid-cols-[1fr_0.9fr]">
        <Card>
          <CardHeader title={t('dashboard.cards.currentTenantInfo')} description={t('dashboard.cards.currentTenantDescription')} />
          {tenantContext.isError ? <ErrorView error={tenantContext.error} title={t('dashboard.cards.tenantInfoError')} /> : null}
          {tenantContext.data ? (
            <dl className="grid gap-3 text-sm sm:grid-cols-2">
              {Object.entries({
                tenant_slug: tenantContext.data.tenant_context.tenant_slug,
                role: tenantContext.data.tenant_context.role,
                auth_type: tenantContext.data.tenant_context.auth_type,
              }).map(([key, value]) => (
                <div key={key} className="rounded-panel bg-surface-soft p-3">
                  <dt className="text-xs font-bold uppercase tracking-wide text-subtle">{key}</dt>
                  <dd className="mt-1 break-all font-semibold text-ink">{Array.isArray(value) ? value.join(', ') : String(value ?? t('common.notAvailable'))}</dd>
                </div>
              ))}
            </dl>
          ) : currentTenantId ? <LoadingView label={t('dashboard.cards.tenantInfoLoading')} /> : <p className="text-sm text-muted">{t('dashboard.cards.tenantRequired')}</p>}
        </Card>
        <Card>
          <CardHeader title={t('dashboard.cards.shortcutsTitle')} description={t('dashboard.cards.shortcutsDescription')} />
          <div className="space-y-3">
            {[
              { to: '/tenant/settings', label: t('dashboard.shortcuts.settings'), icon: <ArrowRight className="size-4" /> },
              { to: '/tenant/access?tab=members', label: t('dashboard.shortcuts.access'), icon: <Users className="size-4" /> },
              { to: '/tenant/api-keys', label: t('dashboard.shortcuts.apiKeys'), icon: <ArrowRight className="size-4" /> },
            ].map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className="flex items-center justify-between rounded-panel border border-line bg-surface-soft p-4 text-sm font-bold text-ink transition hover:border-brand-ring hover:bg-brand-soft hover:text-brand"
              >
                {item.label}
                {item.icon}
              </Link>
            ))}
          </div>
        </Card>
      </section>
    </div>
  )
}
