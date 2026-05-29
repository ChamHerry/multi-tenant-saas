import { Navigate } from 'react-router-dom'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { EmptyState } from '@/shared/ui/EmptyState'

export function PersonalApiKeysPage() {
  const tenantId = useTenantStore((state) => state.currentTenantId)

  if (!tenantId) {
    return <EmptyState title="请先创建或加入组织" description="组织 API Keys 已迁移到“组织管理 / 组织 API Keys”。" />
  }

  return <Navigate to="/tenant/api-keys" replace />
}
