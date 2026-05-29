import { Navigate, createBrowserRouter } from 'react-router-dom'
import { RequireAuth } from './RequireAuth'
import { RequirePermission } from '@/features/access/RequirePermission'
import { DashboardPage } from '@/pages/DashboardPage'
import { LoginPage } from '@/pages/LoginPage'
import { RegisterPage } from '@/pages/RegisterPage'
import { TenantsPage } from '@/pages/TenantsPage'
import { TenantDetailPage } from '@/pages/TenantDetailPage'
import { PersonalApiKeysPage } from '@/pages/PersonalApiKeysPage'
import { MyInvitationsPage } from '@/pages/MyInvitationsPage'
import { TenantAuditPage } from '@/pages/TenantAuditPage'
import { SecurityEventsPage } from '@/pages/SecurityEventsPage'
import { AdminTenantsPage } from '@/pages/AdminTenantsPage'
import { AdminUsersPage } from '@/pages/AdminUsersPage'
import { AdminAuditPage } from '@/pages/AdminAuditPage'
import { NotFoundPage } from '@/pages/NotFoundPage'
import { TenantAccessPage } from '@/pages/TenantAccessPage'
import { TenantApiKeysPage } from '@/pages/TenantApiKeysPage'

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
      { index: true, element: <DashboardPage />, handle: { title: '组织概览' } },
      { path: 'tenants', element: <TenantsPage />, handle: { title: '我的组织' } },
      {
        path: 'tenant/settings',
        handle: { title: '组织设置' },
        element: (
          <RequirePermission tenant={['tenant:read']}>
            <TenantDetailPage />
          </RequirePermission>
        ),
      },
      {
        path: 'tenants/:tenantId',
        handle: { title: '组织设置' },
        element: (
          <RequirePermission tenant={['tenant:read']}>
            <TenantDetailPage />
          </RequirePermission>
        ),
      },
      {
        path: 'tenant/access',
        handle: { title: '用户与访问' },
        element: (
          <RequirePermission tenant={['member:read', 'tenant:invitation:manage']} mode="any">
            <TenantAccessPage />
          </RequirePermission>
        ),
      },
      {
        path: 'tenant/api-keys',
        element: (
          <RequirePermission tenant={['tenant:read']}>
            <TenantApiKeysPage />
          </RequirePermission>
        ),
        handle: { title: '组织 API Keys' },
      },
      {
        path: 'tenant/audit-logs',
        handle: { title: '组织审计' },
        element: (
          <RequirePermission tenant={['tenant:audit:read']}>
            <TenantAuditPage />
          </RequirePermission>
        ),
      },
      { path: 'me/invitations', element: <MyInvitationsPage />, handle: { title: '我的邀请' } },
      { path: 'me/security', element: <SecurityEventsPage />, handle: { title: '安全事件' } },
      { path: 'api-keys', element: <PersonalApiKeysPage />, handle: { title: '组织 API Keys' } },
      { path: 'members', element: <Navigate to="/tenant/access?tab=members" replace />, handle: { title: '用户与访问' } },
      { path: 'invitations', element: <Navigate to="/tenant/access?tab=invitations" replace />, handle: { title: '用户与访问' } },
      { path: 'audit-logs', element: <Navigate to="/tenant/audit-logs" replace />, handle: { title: '组织审计' } },
      {
        path: 'admin/tenants',
        handle: { title: '平台组织' },
        element: (
          <RequirePermission platform={['platform:tenant:read']}>
            <AdminTenantsPage />
          </RequirePermission>
        ),
      },
      {
        path: 'admin/users',
        handle: { title: '平台用户与管理员' },
        element: (
          <RequirePermission platform={['platform:user:read']}>
            <AdminUsersPage />
          </RequirePermission>
        ),
      },
      {
        path: 'admin/audit-logs',
        handle: { title: '平台审计' },
        element: (
          <RequirePermission platform={['platform:audit:read']}>
            <AdminAuditPage />
          </RequirePermission>
        ),
      },
      { path: '*', element: <NotFoundPage />, handle: { title: '页面不存在' } },
    ],
  },
])
