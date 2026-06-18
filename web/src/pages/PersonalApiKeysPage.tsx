import { Navigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { EmptyState } from '@/shared/ui/EmptyState'
import { TENANT_CONSOLE_ROOT } from '@/layouts/console-scope'

export function PersonalApiKeysPage() {
  const { t } = useTranslation()
  const tenantId = useTenantStore((state) => state.currentTenantId)

  if (!tenantId) {
    return <EmptyState title={t('apiKeys.personal.requireTenantTitle')} description={t('apiKeys.personal.requireTenantDescription')} />
  }

  return <Navigate to={`${TENANT_CONSOLE_ROOT}/api-keys`} replace />
}
