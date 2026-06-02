import { Navigate, createBrowserRouter } from 'react-router-dom'
import { RequireAuth } from './RequireAuth'
import { RequirePermission } from '@/features/access/RequirePermission'
import {
  AdminAuditPage,
  AdminTenantsPage,
  AdminUsersPage,
  DashboardPage,
  LazyRoute,
  MyInvitationsPage,
  PersonalApiKeysPage,
  SecurityEventsPage,
  TenantAccessPage,
  TenantApiKeysPage,
  TenantAuditPage,
  TenantDetailPage,
  TenantsPage,
} from './lazy-routes'

// Eagerly loaded: auth pages and shell
import { LoginPage } from '@/pages/LoginPage'
import { RegisterPage } from '@/pages/RegisterPage'
import { VerifyEmailPage } from '@/pages/VerifyEmailPage'
import { NotFoundPage } from '@/pages/NotFoundPage'

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
    path: '/verify-email',
    element: <VerifyEmailPage />,
  },
  {
    path: '/',
    element: <RequireAuth />,
    children: [
      {
        index: true,
        element: (
          <LazyRoute>
            <DashboardPage />
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.dashboard.title' },
      },
      {
        path: 'tenants',
        element: (
          <LazyRoute>
            <TenantsPage />
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.tenants.title' },
      },
      {
        path: 'tenant/settings',
        handle: { titleKey: 'routes.tenantSettings.title' },
        element: (
          <LazyRoute>
            <RequirePermission tenant={['tenant:read']}>
              <TenantDetailPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      {
        path: 'tenants/:tenantId',
        handle: { titleKey: 'routes.tenantSettings.title' },
        element: (
          <LazyRoute>
            <RequirePermission tenant={['tenant:read']}>
              <TenantDetailPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      {
        path: 'tenant/access',
        handle: { titleKey: 'routes.tenantAccess.title' },
        element: (
          <LazyRoute>
            <RequirePermission tenant={['member:read', 'tenant:invitation:manage']} mode="any">
              <TenantAccessPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      {
        path: 'tenant/api-keys',
        element: (
          <LazyRoute>
            <RequirePermission tenant={['tenant:read']}>
              <TenantApiKeysPage />
            </RequirePermission>
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.tenantApiKeys.title' },
      },
      {
        path: 'tenant/audit-logs',
        handle: { titleKey: 'routes.tenantAudit.title' },
        element: (
          <LazyRoute>
            <RequirePermission tenant={['tenant:audit:read']}>
              <TenantAuditPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      {
        path: 'me/invitations',
        element: (
          <LazyRoute>
            <MyInvitationsPage />
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.myInvitations.title' },
      },
      {
        path: 'me/security',
        element: (
          <LazyRoute>
            <SecurityEventsPage />
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.securityEvents.title' },
      },
      {
        path: 'api-keys',
        element: (
          <LazyRoute>
            <PersonalApiKeysPage />
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.tenantApiKeys.title' },
      },
      {
        path: 'members',
        element: <Navigate to="/tenant/access?tab=members" replace />,
        handle: { titleKey: 'routes.tenantAccess.title' },
      },
      {
        path: 'invitations',
        element: <Navigate to="/tenant/access?tab=invitations" replace />,
        handle: { titleKey: 'routes.tenantAccess.title' },
      },
      {
        path: 'audit-logs',
        element: <Navigate to="/tenant/audit-logs" replace />,
        handle: { titleKey: 'routes.tenantAudit.title' },
      },
      {
        path: 'admin/tenants',
        handle: { titleKey: 'routes.platformTenants.title' },
        element: (
          <LazyRoute>
            <RequirePermission platform={['platform:tenant:read']}>
              <AdminTenantsPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      {
        path: 'admin/users',
        handle: { titleKey: 'routes.platformUsers.title' },
        element: (
          <LazyRoute>
            <RequirePermission platform={['platform:user:read']}>
              <AdminUsersPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      {
        path: 'admin/audit-logs',
        handle: { titleKey: 'routes.platformAudit.title' },
        element: (
          <LazyRoute>
            <RequirePermission platform={['platform:audit:read']}>
              <AdminAuditPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      { path: '*', element: <NotFoundPage />, handle: { titleKey: 'routes.notFound.title' } },
    ],
  },
])
