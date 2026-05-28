import { Link } from 'react-router-dom'
import { Activity, ArrowRight, DatabaseZap, Server, Users } from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { useMe, useMyTenants } from '@/features/auth/auth-hooks'
import { useTenantContext } from '@/features/tenants/tenant-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { apiRaw } from '@/shared/api/client'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Card, CardHeader } from '@/shared/ui/Card'
import { usePageTitle } from '@/layouts/page-title'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'

type Health = { ok: boolean; version?: number; dirty?: boolean; error?: string }

function SystemCard({ title, query, icon }: { title: string; query: ReturnType<typeof useQuery<Health>>; icon: React.ReactNode }) {
  return (
    <Card>
      <div className="flex items-center justify-between gap-3">
        <div className="rounded-panel bg-brand-soft p-2 text-brand">{icon}</div>
        <Badge tone={query.data?.ok ? 'green' : query.isError ? 'red' : 'orange'}>{query.data?.ok ? 'OK' : query.isError ? 'ERROR' : 'LOADING'}</Badge>
      </div>
      <h3 className="mt-4 text-lg font-black text-ink">{title}</h3>
      <p className="mt-2 text-sm text-muted">
        {query.isLoading ? '检查中...' : query.isError ? '服务不可用' : query.data?.version ? `migration version ${query.data.version}` : '服务正常'}
      </p>
    </Card>
  )
}

export function DashboardPage() {
  const health = useQuery<Health>({ queryKey: ['system', 'healthz'], queryFn: () => apiRaw<Health>('/healthz') })
  const ready = useQuery<Health>({ queryKey: ['system', 'readyz'], queryFn: () => apiRaw<Health>('/readyz'), retry: false })
  const me = useMe()
  const tenants = useMyTenants()
  const tenantContext = useTenantContext()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const displayName = me.data?.user.display_name || me.data?.user.email || 'Dev User'

  usePageTitle(me.data ? `欢迎，${displayName}` : '仪表盘')

  if (me.isLoading) return <LoadingView label="加载当前用户..." />

  return (
    <div className="space-y-6">
      <section className="hero-gradient rounded-[28px] border border-white p-6 shadow-soft sm:p-8">
        <div className="flex flex-wrap gap-3">
          <Link to="/tenants"><Button>{(tenants.data?.tenants.length ?? 0) === 0 ? '创建第一个组织' : '管理组织'}</Button></Link>
          {(tenants.data?.tenants.length ?? 0) > 0 ? <Link to="/members"><Button variant="secondary">成员管理</Button></Link> : null}
        </div>
        {(tenants.data?.tenants.length ?? 0) === 0 ? (
          <div className="mt-5 rounded-card border border-warning/20 bg-white/80 p-4 text-sm leading-6 text-muted">
            <div className="font-bold text-ink">还没有选择组织</div>
            <p className="mt-1">请进入“组织”页面创建或选择组织。后端会在需要时自动准备运行环境。</p>
          </div>
        ) : null}
      </section>

      <section className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <SystemCard title="HTTP Health" query={health} icon={<Server className="size-5" />} />
        <SystemCard title="DB Ready" query={ready} icon={<DatabaseZap className="size-5" />} />
        <Card>
          <div className="rounded-panel bg-accent-purple-soft p-2 text-accent-purple w-fit"><Users className="size-5" /></div>
          <h3 className="mt-4 text-lg font-black text-ink">我的组织</h3>
          <p className="mt-2 text-3xl font-black text-brand">{tenants.data?.tenants.length ?? 0}</p>
        </Card>
        <Card>
          <div className="rounded-panel bg-success-soft p-2 text-success-strong w-fit"><Activity className="size-5" /></div>
          <h3 className="mt-4 text-lg font-black text-ink">当前组织角色</h3>
          <p className="mt-2 text-sm font-bold text-muted">{currentTenantId ? tenantContext.data?.tenant_context.role ?? '解析中' : '未选择组织'}</p>
        </Card>
      </section>

      <section className="grid gap-4 lg:grid-cols-[1fr_0.9fr]">
        <Card>
          <CardHeader title="当前组织信息" description="API client 会自动使用已选择的组织。" />
          {tenantContext.isError ? <ErrorView error={tenantContext.error} title="组织信息加载失败" /> : null}
          {tenantContext.data ? (
            <dl className="grid gap-3 text-sm sm:grid-cols-2">
              {Object.entries({
                tenant_slug: tenantContext.data.tenant_context.tenant_slug,
                role: tenantContext.data.tenant_context.role,
                auth_type: tenantContext.data.tenant_context.auth_type,
              }).map(([key, value]) => (
                <div key={key} className="rounded-panel bg-surface-soft p-3">
                  <dt className="text-xs font-bold uppercase tracking-wide text-subtle">{key}</dt>
                  <dd className="mt-1 break-all font-semibold text-ink">{Array.isArray(value) ? value.join(', ') : String(value ?? '-')}</dd>
                </div>
              ))}
            </dl>
          ) : currentTenantId ? <LoadingView label="加载组织信息..." /> : <p className="text-sm text-muted">请先选择或创建一个组织。</p>}
        </Card>
        <Card>
          <CardHeader title="快捷入口" description="覆盖当前已经实现的组织级 HTTP 能力。" />
          <div className="space-y-3">
            {[
              { to: '/tenants', label: '创建/更新组织', icon: <ArrowRight className="size-4" /> },
              { to: '/members', label: '添加/更新/移除组织成员', icon: <Users className="size-4" /> },
            ].map((item) => (
              <Link key={item.to} to={item.to} className="flex items-center justify-between rounded-panel border border-line bg-surface-soft p-4 text-sm font-bold text-ink transition hover:border-brand-ring hover:bg-brand-soft hover:text-brand">
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
