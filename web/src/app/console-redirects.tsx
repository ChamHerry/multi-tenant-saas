import { Navigate, useLocation, useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAccess } from '@/features/access/access-hooks'
import {
  getPlatformConsoleHome,
  hasPlatformConsoleAccess,
  hasTenantConsoleAccess,
  TENANT_CONSOLE_HOME,
  TENANT_CONSOLE_ROOT,
} from '@/layouts/console-scope'
import { ErrorView, LoadingView } from '@/shared/ui'

export function RedirectWithLocation({ to }: { to: string }) {
  const location = useLocation()
  return <Navigate to={`${to}${location.search}${location.hash}`} replace />
}

export function TenantDetailLegacyRedirect() {
  const { tenantId } = useParams()
  const location = useLocation()
  return <Navigate to={`${TENANT_CONSOLE_ROOT}/tenants/${tenantId ?? ''}${location.search}${location.hash}`} replace />
}

export function ConsoleEntryRedirect() {
  const { t } = useTranslation()
  const access = useAccess()

  if (access.isLoading) return <LoadingView label={t('access.guard.loading')} />
  if (access.isError) return <ErrorView title={t('access.guard.loadFailed')} error={access.error} />

  const platformPermissions = access.data?.platform_admin?.permissions
  const to = hasTenantConsoleAccess(access.data?.tenants)
    ? TENANT_CONSOLE_HOME
    : hasPlatformConsoleAccess(platformPermissions)
      ? getPlatformConsoleHome(platformPermissions)
      : TENANT_CONSOLE_HOME

  return <Navigate to={to} replace />
}

export function PlatformConsoleIndexRedirect() {
  const { t } = useTranslation()
  const access = useAccess()

  if (access.isLoading) return <LoadingView label={t('access.guard.loading')} />
  if (access.isError) return <ErrorView title={t('access.guard.loadFailed')} error={access.error} />

  return <Navigate to={getPlatformConsoleHome(access.data?.platform_admin?.permissions)} replace />
}
