import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { useParams, useSearchParams } from 'react-router-dom'
import { EmptyState } from '@/shared/ui/EmptyState'
import { ErrorView, LoadingView } from '@/shared/ui/StatusView'
import { hasAll, hasAny } from './can'
import { useAccess } from './access-hooks'
import { useTenantStore } from '@/features/tenants/tenant-store'
import type { PlatformPermission, TenantPermission } from './access-types'

type RequirePermissionProps = {
  children: ReactNode
  tenant?: TenantPermission[]
  platform?: PlatformPermission[]
  mode?: 'all' | 'any'
}

export function RequirePermission({ children, tenant, platform, mode = 'all' }: RequirePermissionProps) {
  const params = useParams()
  const [searchParams] = useSearchParams()
  const { t } = useTranslation()
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const access = useAccess()

  if (access.isLoading) return <LoadingView label={t('access.guard.loading')} />
  if (access.isError) return <ErrorView title={t('access.guard.loadFailed')} error={access.error} />

  if (platform && platform.length > 0) {
    const permissions = access.data?.platform_admin?.permissions
    const allowed = mode === 'all' ? hasAll(permissions, platform) : hasAny(permissions, platform)
    if (!allowed) {
      return (
        <EmptyState
          title={t('access.guard.noPlatformPermissionTitle')}
          description={t('access.guard.noPlatformPermissionDescription')}
        />
      )
    }
  }

  if (tenant && tenant.length > 0) {
    const tenantId = params.tenantId ?? searchParams.get('tenantId') ?? currentTenantId
    const tenantAccess = access.data?.tenants.find((item) => item.tenant_id === tenantId)
    const allowed =
      mode === 'all' ? hasAll(tenantAccess?.permissions, tenant) : hasAny(tenantAccess?.permissions, tenant)
    if (!tenantId) {
      return (
        <EmptyState
          title={t('access.guard.selectTenantTitle')}
          description={t('access.guard.selectTenantDescription')}
        />
      )
    }
    if (!allowed) {
      return (
        <EmptyState
          title={t('access.guard.noTenantPermissionTitle')}
          description={t('access.guard.noTenantPermissionDescription')}
        />
      )
    }
  }

  return children
}
