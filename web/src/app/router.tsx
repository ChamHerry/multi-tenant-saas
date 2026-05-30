import { lazy, Suspense } from 'react'
import { Navigate, createBrowserRouter } from 'react-router-dom'
import { RequireAuth } from './RequireAuth'
import { RequirePermission } from '@/features/access/RequirePermission'

// Eagerly loaded: auth pages and shell
import { LoginPage } from '@/pages/LoginPage'
import { RegisterPage } from '@/pages/RegisterPage'
import { NotFoundPage } from '@/pages/NotFoundPage'

// Lazily loaded: main app pages (code-split into separate chunks)
const DashboardPage = lazy(() => import('@/pages/DashboardPage').then(m => ({ default: m.DashboardPage })))
const TenantsPage = lazy(() => import('@/pages/TenantsPage').then(m => ({ default: m.TenantsPage })))
const TenantDetailPage = lazy(() => import('@/pages/TenantDetailPage').then(m => ({ default: m.TenantDetailPage })))
const PersonalApiKeysPage = lazy(() => import('@/pages/PersonalApiKeysPage').then(m => ({ default: m.PersonalApiKeysPage })))
const MyInvitationsPage = lazy(() => import('@/pages/MyInvitationsPage').then(m => ({ default: m.MyInvitationsPage })))
const TenantAuditPage = lazy(() => import('@/pages/TenantAuditPage').then(m => ({ default: m.TenantAuditPage })))
const SecurityEventsPage = lazy(() => import('@/pages/SecurityEventsPage').then(m => ({ default: m.SecurityEventsPage })))
const TenantAccessPage = lazy(() => import('@/pages/TenantAccessPage').then(m => ({ default: m.TenantAccessPage })))
const TenantApiKeysPage = lazy(() => import('@/pages/TenantApiKeysPage').then(m => ({ default: m.TenantApiKeysPage })))

// Admin pages — separate chunk
const AdminTenantsPage = lazy(() => import('@/pages/AdminTenantsPage').then(m => ({ default: m.AdminTenantsPage })))
const AdminUsersPage = lazy(() => import('@/pages/AdminUsersPage').then(m => ({ default: m.AdminUsersPage })))
const AdminAuditPage = lazy(() => import('@/pages/AdminAuditPage').then(m => ({ default: m.AdminAuditPage })))

function LoadingFallback() {
  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-muted-foreground text-sm">Loading…</div>
    </div>
  )
}

function Lazy({ children }: { children: React.ReactNode }) {
  return <Suspense fallback={<LoadingFallback />}>{children}</Suspense>
}

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/register',
    element: <RegisterPage />,
  },
  {
    path: '/',
    element: <RequireAuth />,
    children: [
      { index: true, element: <Lazy><DashboardPage /></Lazy>, handle: { title: '组织概览' } },
      { path: 'tenants', element: <Lazy><TenantsPage /></Lazy>, handle: { title: '我的组织' } },
      {
        path: 'tenant/settings',
        handle: { title: '组织设置' },
        element: (
          <Lazy>
            <RequirePermission tenant={['tenant:read']}>
              <TenantDetailPage />
            </RequirePermission>
          </Lazy>
        ),
      },
      {
        path: 'tenants/:tenantId',
        handle: { title: '组织设置' },
        element: (
          <Lazy>
            <RequirePermission tenant={['tenant:read']}>
              <TenantDetailPage />
            </RequirePermission>
          </Lazy>
        ),
      },
      {
        path: 'tenant/access',
        handle: { title: '用户与访问' },
        element: (
          <Lazy>
            <RequirePermission tenant={['member:read', 'tenant:invitation:manage']} mode="any">
              <TenantAccessPage />
            </RequirePermission>
          </Lazy>
        ),
      },
      {
        path: 'tenant/api-keys',
        element: (
          <Lazy>
            <RequirePermission tenant={['tenant:read']}>
              <TenantApiKeysPage />
            </RequirePermission>
          </Lazy>
        ),
        handle: { title: '组织 API Keys' },
      },
      {
        path: 'tenant/audit-logs',
        handle: { title: '组织审计' },
        element: (
          <Lazy>
            <RequirePermission tenant={['tenant:audit:read']}>
              <TenantAuditPage />
            </RequirePermission>
          </Lazy>
        ),
      },
      { path: 'me/invitations', element: <Lazy><MyInvitationsPage /></Lazy>, handle: { title: '我的邀请' } },
      { path: 'me/security', element: <Lazy><SecurityEventsPage /></Lazy>, handle: { title: '安全事件' } },
      { path: 'api-keys', element: <Lazy><PersonalApiKeysPage /></Lazy>, handle: { title: '组织 API Keys' } },
      { path: 'members', element: <Navigate to="/tenant/access?tab=members" replace />, handle: { title: '用户与访问' } },
      { path: 'invitations', element: <Navigate to="/tenant/access?tab=invitations" replace />, handle: { title: '用户与访问' } },
      { path: 'audit-logs', element: <Navigate to="/tenant/audit-logs" replace />, handle: { title: '组织审计' } },
      {
        path: 'admin/tenants',
        handle: { title: '平台组织' },
        element: (
          <Lazy>
            <RequirePermission platform={['platform:tenant:read']}>
              <AdminTenantsPage />
            </RequirePermission>
          </Lazy>
        ),
      },
      {
        path: 'admin/users',
        handle: { title: '平台用户与管理员' },
        element: (
          <Lazy>
            <RequirePermission platform={['platform:user:read']}>
              <AdminUsersPage />
            </RequirePermission>
          </Lazy>
        ),
      },
      {
        path: 'admin/audit-logs',
        handle: { title: '平台审计' },
        element: (
          <Lazy>
            <RequirePermission platform={['platform:audit:read']}>
              <AdminAuditPage />
            </RequirePermission>
          </Lazy>
        ),
      },
      { path: '*', element: <NotFoundPage />, handle: { title: '页面不存在' } },
    ],
  },
])
