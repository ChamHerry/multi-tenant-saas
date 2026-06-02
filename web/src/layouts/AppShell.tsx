import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, NavLink, Outlet, useMatches, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Building2, CircleUser, LogOut } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { useAccess } from '@/features/access/access-hooks'
import { useLogoutMutation, useMe } from '@/features/auth/auth-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { LanguageSwitcher } from '@/shared/i18n'
import { Badge, Button, ErrorView } from '@/shared/ui'
import { cn } from '@/shared/lib/cn'
import { buildScopeNavSections } from './scope-nav'
import { PageTitleContext, resolvePageTitle, type PageTitleDescriptor } from './page-title'

type RouteTitleHandle = {
  titleKey?: string
  titleValues?: PageTitleDescriptor['values']
}

function getRouteTitleDescriptor(matches: { handle?: unknown }[]): PageTitleDescriptor {
  for (const match of [...matches].reverse()) {
    const handle = match.handle as RouteTitleHandle | undefined
    if (typeof handle?.titleKey === 'string' && handle.titleKey.trim()) {
      return { key: handle.titleKey, values: handle.titleValues }
    }
  }
  return { key: 'routes.dashboard.title' }
}

export function AppShell() {
  const navigate = useNavigate()
  const matches = useMatches()
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data: me, error: meError } = useMe()
  const { data: access, error: accessError } = useAccess()
  const logoutMutation = useLogoutMutation()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const setCurrentTenantId = useTenantStore((state) => state.setCurrentTenantId)
  const [pageTitleOverride, setPageTitleOverride] = useState<PageTitleDescriptor>()
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false)
  const [isTenantMenuOpen, setIsTenantMenuOpen] = useState(false)
  const userMenuRef = useRef<HTMLDivElement>(null)
  const tenantMenuRef = useRef<HTMLDivElement>(null)
  const tenants = useMemo(() => access?.tenants ?? [], [access?.tenants])
  const currentMembership = tenants.find((tenant) => tenant.tenant_id === currentTenantId)
  const currentUserLabel = me?.user.display_name || me?.user.email || t('appShell.user.devUser')
  const routeTitle = getRouteTitleDescriptor(matches)
  const setPageTitle = useCallback((title?: PageTitleDescriptor) => {
    setPageTitleOverride(title)
  }, [])
  const pageTitleContext = useMemo(() => ({ setPageTitle }), [setPageTitle])
  const pageTitle = resolvePageTitle(t, pageTitleOverride ?? routeTitle)
  const scopeSections = useMemo(
    () =>
      buildScopeNavSections({
        tenantPermissions: currentMembership?.permissions,
        platformPermissions: access?.platform_admin?.permissions,
      }),
    [access?.platform_admin?.permissions, currentMembership?.permissions],
  );

  useEffect(() => {
    if (tenants.length === 0) {
      if (currentTenantId) setCurrentTenantId(undefined);
      return;
    }
    const hasSelected =
      currentTenantId &&
      tenants.some(
        (tenant) =>
          tenant.tenant_id === currentTenantId && tenant.status === 'active' && tenant.tenant_status === 'active',
      )
    if (!hasSelected) {
      const firstActive =
        tenants.find(
          (tenant) =>
            tenant.status === "active" && tenant.tenant_status === "active",
        ) ?? tenants[0];
      setCurrentTenantId(firstActive.tenant_id);
    }
  }, [currentTenantId, setCurrentTenantId, tenants]);

  useEffect(() => {
    if (!isUserMenuOpen && !isTenantMenuOpen) {
      return;
    }
    const handlePointerDown = (event: MouseEvent) => {
      const target = event.target as Node;
      if (
        !userMenuRef.current?.contains(target) &&
        !tenantMenuRef.current?.contains(target)
      ) {
        setIsUserMenuOpen(false);
        setIsTenantMenuOpen(false);
      }
    };
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setIsUserMenuOpen(false);
        setIsTenantMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [isTenantMenuOpen, isUserMenuOpen]);

  const chooseTenant = (tenantId: string) => {
    setCurrentTenantId(tenantId);
    setIsTenantMenuOpen(false);
  };

  const logout = async () => {
    try {
      await logoutMutation.mutateAsync();
    } finally {
      setCurrentTenantId(undefined);
      queryClient.clear();
      navigate("/login", { replace: true });
    }
  };

  return (
    <div className="min-h-screen bg-page">
      <aside className="fixed inset-y-0 left-0 z-20 hidden w-[18rem] flex-col border-r border-line bg-white/85 px-4 py-5 shadow-soft backdrop-blur lg:flex">
        {currentMembership ? (
          <div ref={tenantMenuRef} className="relative">
            <button
              type="button"
              className="flex w-full items-center justify-between gap-2 rounded-panel bg-surface-soft px-3 py-2 text-left text-sm transition hover:bg-brand-soft"
              onClick={() => {
                setIsTenantMenuOpen((current) => !current)
                setIsUserMenuOpen(false)
              }}
            >
              <div className="min-w-0">
                <div className="truncate font-bold text-ink">{currentMembership.tenant_name}</div>
                <div className="truncate text-xs text-subtle">{currentMembership.tenant_slug}</div>
              </div>
              <Badge tone="green">{currentMembership.role}</Badge>
            </button>
            {isTenantMenuOpen ? (
              <div className="absolute inset-x-0 top-full z-20 mt-2 rounded-card border border-line bg-white p-2 shadow-brand">
                <div className="px-2 py-1 text-xs font-bold uppercase tracking-wide text-subtle">
                  {t('appShell.tenant.selectLabel')}
                </div>
                <div className="mt-1 space-y-1">
                  {tenants.map((tenant) => {
                    const isCurrent = tenant.tenant_id === currentTenantId
                    return (
                      <button
                        key={tenant.tenant_id}
                        type="button"
                        className={cn(
                          'flex w-full items-center justify-between gap-2 rounded-panel px-3 py-2 text-left text-sm transition hover:bg-brand-soft',
                          isCurrent && 'bg-brand-soft',
                        )}
                        onClick={() => chooseTenant(tenant.tenant_id)}
                      >
                        <div className="min-w-0">
                          <div className="truncate font-bold text-ink">{tenant.tenant_name}</div>
                          <div className="truncate text-xs text-subtle">{tenant.tenant_slug}</div>
                        </div>
                        {isCurrent ? <Badge tone="green">{t('appShell.tenant.currentBadge')}</Badge> : null}
                      </button>
                    )
                  })}
                </div>
              </div>
            ) : null}
          </div>
        ) : (
          <div className="mt-3 rounded-panel bg-surface-soft px-3 py-3 text-sm text-muted">
            <div>{t('appShell.tenant.emptyTitle')}</div>
            <div className="mt-1 text-xs text-subtle">{t('appShell.tenant.emptyDescription')}</div>
            <div className="mt-3">
              <Link to="/tenants">
                <Button variant="secondary" size="sm" leftIcon={<Building2 className="size-4" />}>
                  {t('appShell.tenant.manageCta')}
                </Button>
              </Link>
            </div>
          </div>
        )}

        <nav className="mt-6 space-y-5">
          {scopeSections.map((section) => (
            <div key={section.titleKey}>
              <div className="px-3 pb-2 text-xs font-black uppercase tracking-wide text-subtle">
                {t(section.titleKey)}
              </div>
              <div className="space-y-1">
                {section.items.map((item) => {
                  const Icon = item.icon;
                  const isPlatform = section.tone === "platform";
                  return (
                    <NavLink
                      key={item.to}
                      to={item.to}
                      end={item.to === "/"}
                      className={({ isActive }) =>
                        cn(
                          "flex items-center gap-3 rounded-panel px-3 py-2.5 text-sm font-bold transition",
                          isPlatform
                            ? "text-muted hover:bg-red-50 hover:text-danger"
                            : "text-muted hover:bg-brand-soft hover:text-brand",
                          isActive &&
                            (isPlatform
                              ? "bg-danger text-white shadow-soft hover:bg-danger hover:text-white"
                              : "bg-brand text-white shadow-soft hover:bg-brand hover:text-white"),
                        )
                      }
                    >
                      <Icon className="size-4" />
                      {t(item.labelKey)}
                    </NavLink>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>

        <div className="mt-auto pt-4">
          <div ref={userMenuRef} className="relative">
            <button
              type="button"
              className="flex w-full items-center gap-3 rounded-panel px-3 py-2.5 text-sm font-bold transition text-muted hover:bg-brand-soft hover:text-brand"
              onClick={() => {
                setIsUserMenuOpen((current) => !current);
                setIsTenantMenuOpen(false);
              }}
            >
              <CircleUser className="size-4" />
              <span className="truncate">{currentUserLabel}</span>
            </button>
            {isUserMenuOpen ? (
              <div className="absolute bottom-full inset-x-0 z-20 mb-2 rounded-card border border-line bg-white p-2 shadow-brand">
                <button
                  type="button"
                  className="flex w-full items-center gap-3 rounded-panel px-3 py-2 text-left text-sm font-bold text-muted transition hover:bg-brand-soft hover:text-brand disabled:cursor-not-allowed disabled:opacity-55"
                  onClick={() => void logout()}
                  disabled={logoutMutation.isPending}
                >
                  <LogOut className="size-4" />
                  <span>
                    {logoutMutation.isPending ? t('appShell.user.logoutPending') : t('appShell.user.logout')}
                  </span>
                </button>
              </div>
            ) : null}
          </div>
        </div>
      </aside>

      <div className="lg:pl-[18rem]">
        <PageTitleContext.Provider value={pageTitleContext}>
          <header className="sticky top-0 z-10 border-b border-line bg-page/90 px-4 py-3 backdrop-blur sm:px-6">
            <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <div className="flex items-center gap-2">
                  <h1 className="text-xl font-black text-ink">{pageTitle}</h1>
                </div>
              </div>
              <div className="flex flex-wrap items-center justify-end gap-2">
                <LanguageSwitcher />
              </div>
            </div>
          </header>
          <main className="mx-auto max-w-7xl p-4 sm:p-6">
            {meError ? (
              <ErrorView title={t('appShell.errors.meLoad')} error={meError} />
            ) : accessError ? (
              <ErrorView title={t('appShell.errors.accessLoad')} error={accessError} />
            ) : (
              <Outlet />
            )}
          </main>
        </PageTitleContext.Provider>
      </div>
    </div>
  );
}
