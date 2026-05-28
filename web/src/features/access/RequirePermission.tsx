import type { ReactNode } from 'react'
import { useParams } from 'react-router-dom'
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
  const currentTenantId = useTenantStore((state) => state.currentTenantId)
  const access = useAccess()

  if (access.isLoading) return <LoadingView label="正在加载权限上下文..." />
  if (access.isError) return <ErrorView title="权限上下文加载失败" error={access.error} />

  if (platform && platform.length > 0) {
    const permissions = access.data?.platform_admin?.permissions
    const allowed = mode === 'all' ? hasAll(permissions, platform) : hasAny(permissions, platform)
    if (!allowed) {
      return <EmptyState title="无平台权限" description="当前账号没有访问该平台管理页面所需的权限。真实 API 仍由后端授权保护。" />
    }
  }

  if (tenant && tenant.length > 0) {
    const tenantId = params.tenantId ?? currentTenantId
    const tenantAccess = access.data?.tenants.find((item) => item.tenant_id === tenantId)
    const allowed = mode === 'all' ? hasAll(tenantAccess?.permissions, tenant) : hasAny(tenantAccess?.permissions, tenant)
    if (!tenantId) {
      return <EmptyState title="请先选择组织" description="该页面需要组织上下文。可以在顶部选择已有组织，或先创建组织。" />
    }
    if (!allowed) {
      return <EmptyState title="无组织权限" description="当前组织角色没有访问该页面所需的权限。真实 API 仍由后端 RBAC 保护。" />
    }
  }

  return children
}
