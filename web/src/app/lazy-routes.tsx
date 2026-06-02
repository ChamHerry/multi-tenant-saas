import { lazy, Suspense, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

export const DashboardPage = lazy(() => import('@/pages/DashboardPage').then((m) => ({ default: m.DashboardPage })))
export const TenantsPage = lazy(() => import('@/pages/TenantsPage').then((m) => ({ default: m.TenantsPage })))
export const TenantDetailPage = lazy(() => import('@/pages/TenantDetailPage').then((m) => ({ default: m.TenantDetailPage })))
export const PersonalApiKeysPage = lazy(() =>
  import('@/pages/PersonalApiKeysPage').then((m) => ({ default: m.PersonalApiKeysPage })),
)
export const MyInvitationsPage = lazy(() =>
  import('@/pages/MyInvitationsPage').then((m) => ({ default: m.MyInvitationsPage })),
)
export const TenantAuditPage = lazy(() => import('@/pages/TenantAuditPage').then((m) => ({ default: m.TenantAuditPage })))
export const SecurityEventsPage = lazy(() =>
  import('@/pages/SecurityEventsPage').then((m) => ({ default: m.SecurityEventsPage })),
)
export const SessionManagementPage = lazy(() =>
  import('@/pages/SessionManagementPage').then((m) => ({ default: m.SessionManagementPage })),
)
export const TOTPSetupPage = lazy(() => import('@/pages/TOTPSetupPage').then((m) => ({ default: m.TOTPSetupPage })))
export const TenantAccessPage = lazy(() => import('@/pages/TenantAccessPage').then((m) => ({ default: m.TenantAccessPage })))
export const TenantApiKeysPage = lazy(() =>
  import('@/pages/TenantApiKeysPage').then((m) => ({ default: m.TenantApiKeysPage })),
)
export const AdminTenantsPage = lazy(() => import('@/pages/AdminTenantsPage').then((m) => ({ default: m.AdminTenantsPage })))
export const AdminUsersPage = lazy(() => import('@/pages/AdminUsersPage').then((m) => ({ default: m.AdminUsersPage })))
export const AdminAuditPage = lazy(() => import('@/pages/AdminAuditPage').then((m) => ({ default: m.AdminAuditPage })))

function LoadingFallback() {
  const { t } = useTranslation()

  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-muted-foreground text-sm">{t('common.loading')}</div>
    </div>
  )
}

export function LazyRoute({ children }: { children: ReactNode }) {
  return <Suspense fallback={<LoadingFallback />}>{children}</Suspense>
}
