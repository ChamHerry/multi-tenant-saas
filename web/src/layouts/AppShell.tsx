import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link, NavLink, Outlet, useMatches, useNavigate } from 'react-router-dom'
import { Building2, LogOut } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { useAccess } from '@/features/access/access-hooks'
import { useLogoutMutation, useMe } from '@/features/auth/auth-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Badge } from '@/shared/ui/Badge'
import { Button } from '@/shared/ui/Button'
import { Select } from '@/shared/ui/Select'
import { ErrorView } from '@/shared/ui/StatusView'
import { cn } from '@/shared/lib/cn'
import { buildScopeNavSections } from './scope-nav'
import { PageTitleContext } from './page-title'

function getRouteTitle(matches: { handle?: unknown }[]) {
  for (const match of [...matches].reverse()) {
    const title = (match.handle as { title?: unknown } | undefined)?.title
    if (typeof title === 'string' && title.trim()) return title
  }
  return '组织概览'
}

export function AppShell() {
  const navigate = useNavigate()
  const matches = useMatches()
  const queryClient = useQueryClient()
  const { data: me, error: meError } = useMe()
  const { data: access, error: accessError } = useAccess()
  const logoutMutation = useLogoutMutation()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const setCurrentTenantId = useTenantStore((state) => state.setCurrentTenantId)
  const [pageTitleOverride, setPageTitleOverride] = useState<string>()
  const tenants = useMemo(() => access?.tenants ?? [], [access?.tenants])
  const currentMembership = tenants.find((tenant) => tenant.tenant_id === currentTenantId)
  const routeTitle = getRouteTitle(matches)
  const setPageTitle = useCallback((title?: string) => {
    setPageTitleOverride(title?.trim() || undefined)
  }, [])
  const pageTitleContext = useMemo(() => ({ setPageTitle }), [setPageTitle])
  const pageTitle = pageTitleOverride ?? routeTitle
  const scopeSections = useMemo(
    () =>
      buildScopeNavSections({
        tenantPermissions: currentMembership?.permissions,
        platformPermissions: access?.platform_admin?.permissions,
      }),
    [access?.platform_admin?.permissions, currentMembership?.permissions],
  )

  useEffect(() => {
    if (tenants.length === 0) {
      if (currentTenantId) setCurrentTenantId(undefined)
      return
    }
    const hasSelected =
      currentTenantId &&
      tenants.some(
        (tenant) => tenant.tenant_id === currentTenantId && tenant.status === 'active' && tenant.tenant_status === 'active',
      )
    if (!hasSelected) {
      const firstActive =
        tenants.find((tenant) => tenant.status === 'active' && tenant.tenant_status === 'active') ?? tenants[0]
      setCurrentTenantId(firstActive.tenant_id)
    }
  }, [currentTenantId, setCurrentTenantId, tenants])

  const logout = async () => {
    try {
      await logoutMutation.mutateAsync()
    } finally {
      setCurrentTenantId(undefined)
      queryClient.clear()
      navigate('/login', { replace: true })
    }
  }

  return (
    <div className="min-h-screen bg-page">
      <aside className="fixed inset-y-0 left-0 z-20 hidden w-[18rem] border-r border-line bg-white/85 px-4 py-5 shadow-soft backdrop-blur lg:block">
        <Link to="/" className="flex items-center gap-3 rounded-card bg-brand-soft p-3">
          <div className="brand-gradient grid size-10 place-items-center rounded-panel text-sm font-black text-white">MT</div>
          <div>
            <div className="text-base font-black text-ink">SaaS Template</div>
            <div className="text-xs font-semibold text-brand">Scope-aware Console</div>
          </div>
        </Link>

        <div className="mt-6 rounded-card border border-line bg-surface p-4 shadow-soft">
          <div className="text-xs font-bold uppercase tracking-wide text-subtle">当前组织</div>
          <div className="mt-3">
            <Select
              value={currentTenantId ?? ''}
              onChange={(event) => setCurrentTenantId(event.target.value || undefined)}
              className="bg-white"
              aria-label="切换当前组织"
              disabled={tenants.length === 0}
            >
              <option value="">{tenants.length === 0 ? '暂无组织' : '选择组织'}</option>
              {tenants.map((tenant) => (
                <option key={tenant.tenant_id} value={tenant.tenant_id}>
                  {tenant.tenant_name}
                </option>
              ))}
            </Select>
          </div>
          {currentMembership ? (
            <div className="mt-3 flex items-center justify-between gap-2 rounded-panel bg-surface-soft px-3 py-2 text-sm">
              <div className="min-w-0">
                <div className="truncate font-bold text-ink">{currentMembership.tenant_name}</div>
                <div className="truncate text-xs text-subtle">{currentMembership.tenant_slug}</div>
              </div>
              <Badge tone="green">{currentMembership.role}</Badge>
            </div>
          ) : (
            <div className="mt-3 rounded-panel bg-surface-soft px-3 py-3 text-sm text-muted">
              <div>未选择可用组织。</div>
              <div className="mt-1 text-xs text-subtle">可先创建或加入组织，再访问组织管理菜单。</div>
              <div className="mt-3">
                <Link to="/tenants">
                  <Button variant="secondary" size="sm" leftIcon={<Building2 className="size-4" />}>
                    管理组织
                  </Button>
                </Link>
              </div>
            </div>
          )}
        </div>

        <nav className="mt-6 space-y-5 pb-32">
          {scopeSections.map((section) => (
            <div key={section.title}>
              <div className="px-3 pb-2 text-xs font-black uppercase tracking-wide text-subtle">{section.title}</div>
              <div className="space-y-1">
                {section.items.map((item) => {
                  const Icon = item.icon
                  const isPlatform = section.tone === 'platform'
                  return (
                    <NavLink
                      key={item.to}
                      to={item.to}
                      end={item.to === '/'}
                      className={({ isActive }) =>
                        cn(
                          'flex items-center gap-3 rounded-panel px-3 py-2.5 text-sm font-bold transition',
                          isPlatform
                            ? 'text-muted hover:bg-red-50 hover:text-danger'
                            : 'text-muted hover:bg-brand-soft hover:text-brand',
                          isActive &&
                            (isPlatform
                              ? 'bg-danger text-white shadow-soft hover:bg-danger hover:text-white'
                              : 'bg-brand text-white shadow-soft hover:bg-brand hover:text-white'),
                        )
                      }
                    >
                      <Icon className="size-4" />
                      {item.label}
                    </NavLink>
                  )
                })}
              </div>
            </div>
          ))}
        </nav>

        <div className="absolute inset-x-4 bottom-5 rounded-card border border-line bg-surface-soft p-4">
          <div className="text-xs font-bold uppercase tracking-wide text-subtle">当前用户</div>
          <div className="mt-2 truncate text-sm font-bold text-ink">{me?.user.display_name || me?.user.email || 'Dev User'}</div>
          <div className="mt-1 truncate text-xs text-muted">{me?.user.id}</div>
        </div>
      </aside>

      <div className="lg:pl-[18rem]">
        <PageTitleContext.Provider value={pageTitleContext}>
          <header className="sticky top-0 z-10 border-b border-line bg-page/90 px-4 py-3 backdrop-blur sm:px-6">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <h1 className="text-xl font-black text-ink">{pageTitle}</h1>
                  {currentMembership ? <Badge tone="green">{currentMembership.role}</Badge> : <Badge tone="orange">无组织上下文</Badge>}
                </div>
              </div>
              <div className="flex flex-wrap items-center justify-end gap-2">
                {currentMembership ? (
                  <div className="rounded-panel bg-white px-3 py-2 text-sm text-muted shadow-soft">
                    当前组织：<span className="font-bold text-ink">{currentMembership.tenant_name}</span>
                  </div>
                ) : null}
                <Button
                  variant="ghost"
                  onClick={() => void logout()}
                  isLoading={logoutMutation.isPending}
                  leftIcon={<LogOut className="size-4" />}
                >
                  退出
                </Button>
              </div>
            </div>
          </header>
          <main className="mx-auto max-w-7xl p-4 sm:p-6">
            {meError ? (
              <ErrorView title="当前用户认证失败" error={meError} />
            ) : accessError ? (
              <ErrorView title="权限上下文加载失败" error={accessError} />
            ) : (
              <Outlet />
            )}
          </main>
        </PageTitleContext.Provider>
      </div>
    </div>
  )
}
