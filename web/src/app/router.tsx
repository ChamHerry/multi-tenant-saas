import { Navigate, createBrowserRouter } from 'react-router-dom'
import { RequireAuth } from './RequireAuth'
import { RequireSetup } from './RequireSetup'
import { RequirePermission } from '@/features/access/RequirePermission'
import { TENANT_CONSOLE_HOME } from '@/layouts/console-scope'
import {
  ConsoleEntryRedirect,
  PlatformConsoleIndexRedirect,
  RedirectWithLocation,
  TenantDetailLegacyRedirect,
} from './console-redirects'
import {
  AdminAuditPage,
  AdminSystemConfigPage,
  AdminTenantsPage,
  AdminUsersPage,
  DashboardPage,
  LazyRoute,
  MyInvitationsPage,
  PersonalApiKeysPage,
  SecurityEventsPage,
  SessionManagementPage,
  TenantAccessPage,
  TenantApiKeysPage,
  TenantAuditPage,
  TenantDetailPage,
  TenantsPage,
  TOTPSetupPage,
} from './lazy-routes'

// Eagerly loaded: auth pages and shell
import { LoginPage } from '@/pages/LoginPage'
import { RegisterPage } from '@/pages/RegisterPage'
import { VerifyEmailPage } from '@/pages/VerifyEmailPage'
import { ForgotPasswordPage } from '@/pages/ForgotPasswordPage'
import { ResetPasswordPage } from '@/pages/ResetPasswordPage'
import { TOTPVerifyPage } from '@/pages/TOTPVerifyPage'
import { NotFoundPage } from '@/pages/NotFoundPage'
import { SetupWizardPage } from '@/pages/SetupWizardPage'

export const router = createBrowserRouter([
  {
    path: '/setup',
    element: (
      <RequireSetup setupRoute>
        <SetupWizardPage />
      </RequireSetup>
    ),
    handle: { titleKey: 'routes.setup.title' },
  },
  {
    path: '/login',
    element: (
      <RequireSetup>
        <LoginPage />
      </RequireSetup>
    ),
  },
  {
    path: '/register',
    element: (
      <RequireSetup>
        <RegisterPage />
      </RequireSetup>
    ),
  },
  {
    path: '/verify-email',
    element: (
      <RequireSetup>
        <VerifyEmailPage />
      </RequireSetup>
    ),
  },
  {
    path: '/forgot-password',
    element: (
      <RequireSetup>
        <ForgotPasswordPage />
      </RequireSetup>
    ),
  },
  {
    path: '/reset-password',
    element: (
      <RequireSetup>
        <ResetPasswordPage />
      </RequireSetup>
    ),
  },
  {
    path: '/verify-totp',
    element: (
      <RequireSetup>
        <TOTPVerifyPage />
      </RequireSetup>
    ),
  },
  {
    path: '/',
    element: (
      <RequireSetup>
        <RequireAuth />
      </RequireSetup>
    ),
    children: [
      {
        index: true,
        element: <ConsoleEntryRedirect />,
        handle: { titleKey: 'routes.dashboard.title' },
      },
      {
        path: 'console',
        element: (
          <LazyRoute>
            <RequirePermission tenant={['tenant:read']}>
              <DashboardPage />
            </RequirePermission>
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.dashboard.title' },
      },
      {
        path: 'console/tenants',
        element: (
          <LazyRoute>
            <TenantsPage />
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.tenants.title' },
      },
      {
        path: 'console/settings',
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
        path: 'console/tenants/:tenantId',
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
        path: 'console/access',
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
        path: 'console/api-keys',
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
        path: 'console/audit-logs',
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
        path: 'dashboard',
        element: <Navigate to={TENANT_CONSOLE_HOME} replace />,
        handle: { titleKey: 'routes.dashboard.title' },
      },
      {
        path: 'tenants',
        element: <RedirectWithLocation to="/console/tenants" />,
        handle: { titleKey: 'routes.tenants.title' },
      },
      {
        path: 'tenant/settings',
        element: <RedirectWithLocation to="/console/settings" />,
        handle: { titleKey: 'routes.tenantSettings.title' },
      },
      {
        path: 'tenants/:tenantId',
        element: <TenantDetailLegacyRedirect />,
        handle: { titleKey: 'routes.tenantSettings.title' },
      },
      {
        path: 'tenant/access',
        element: <RedirectWithLocation to="/console/access" />,
        handle: { titleKey: 'routes.tenantAccess.title' },
      },
      {
        path: 'tenant/api-keys',
        element: <RedirectWithLocation to="/console/api-keys" />,
        handle: { titleKey: 'routes.tenantApiKeys.title' },
      },
      {
        path: 'tenant/audit-logs',
        element: <RedirectWithLocation to="/console/audit-logs" />,
        handle: { titleKey: 'routes.tenantAudit.title' },
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
        path: 'me/sessions',
        element: (
          <LazyRoute>
            <SessionManagementPage />
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.sessions.title' },
      },
      {
        path: 'me/security/totp/setup',
        element: (
          <LazyRoute>
            <TOTPSetupPage />
          </LazyRoute>
        ),
        handle: { titleKey: 'routes.totpSetup.title' },
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
        element: <Navigate to="/console/access?tab=members" replace />,
        handle: { titleKey: 'routes.tenantAccess.title' },
      },
      {
        path: 'invitations',
        element: <Navigate to="/console/access?tab=invitations" replace />,
        handle: { titleKey: 'routes.tenantAccess.title' },
      },
      {
        path: 'audit-logs',
        element: <Navigate to="/console/audit-logs" replace />,
        handle: { titleKey: 'routes.tenantAudit.title' },
      },
      {
        path: 'platform-admin-console',
        element: <PlatformConsoleIndexRedirect />,
        handle: { titleKey: 'routes.platformTenants.title' },
      },
      {
        path: 'platform-admin-console/tenants',
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
        path: 'platform-admin-console/users',
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
        path: 'platform-admin-console/audit-logs',
        handle: { titleKey: 'routes.platformAudit.title' },
        element: (
          <LazyRoute>
            <RequirePermission platform={['platform:audit:read']}>
              <AdminAuditPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      {
        path: 'platform-admin-console/system-config',
        handle: { titleKey: 'routes.systemConfig.title' },
        element: (
          <LazyRoute>
            <RequirePermission platform={['platform:config:read']}>
              <AdminSystemConfigPage />
            </RequirePermission>
          </LazyRoute>
        ),
      },
      {
        path: 'admin/tenants',
        element: <RedirectWithLocation to="/platform-admin-console/tenants" />,
        handle: { titleKey: 'routes.platformTenants.title' },
      },
      {
        path: 'admin/users',
        element: <RedirectWithLocation to="/platform-admin-console/users" />,
        handle: { titleKey: 'routes.platformUsers.title' },
      },
      {
        path: 'admin/audit-logs',
        element: <RedirectWithLocation to="/platform-admin-console/audit-logs" />,
        handle: { titleKey: 'routes.platformAudit.title' },
      },
      {
        path: 'admin/system-config',
        element: <RedirectWithLocation to="/platform-admin-console/system-config" />,
        handle: { titleKey: 'routes.systemConfig.title' },
      },
      { path: '*', element: <NotFoundPage />, handle: { titleKey: 'routes.notFound.title' } },
    ],
  },
])
