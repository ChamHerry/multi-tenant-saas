import { Navigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { EmptyState } from '@/shared/ui/EmptyState'

export function PersonalApiKeysPage() {
  const { t } = useTranslation()
  const tenantId = useTenantStore((state) => state.currentTenantId)

  if (!tenantId) {
    return <EmptyState title={t('apiKeys.personal.requireTenantTitle')} description={t('apiKeys.personal.requireTenantDescription')} />
  }

  return <Navigate to="/tenant/api-keys" replace />
}
