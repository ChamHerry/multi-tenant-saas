import { useEffect, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useCanTenant } from '@/features/access/access-hooks'
import { TenantInvitationsPanel } from '@/features/tenant-access/TenantInvitationsPanel'
import { TenantMembersPanel } from '@/features/tenant-access/TenantMembersPanel'
import { useTenantI18n } from '@/features/tenants/tenant-i18n'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Button } from '@/shared/ui/Button'
import { EmptyState } from '@/shared/ui/EmptyState'

const tabLabels = ['members', 'invitations'] as const

type TenantAccessTab = (typeof tabLabels)[number]

export function TenantAccessPage() {
  const tenantId = useTenantStore((state) => state.currentTenantId)
  const canReadMembers = useCanTenant('member:read', tenantId)
  const canManageInvitations = useCanTenant('tenant:invitation:manage', tenantId)
  const [searchParams, setSearchParams] = useSearchParams()
  const { t } = useTenantI18n()

  const tabs = useMemo(
    () => [
      { id: 'members' as const, label: t('tenantAccess.tabs.members', '成员'), enabled: canReadMembers },
      { id: 'invitations' as const, label: t('tenantAccess.tabs.invitations', '邀请'), enabled: canManageInvitations },
    ],
    [canManageInvitations, canReadMembers, t],
  )
  const availableTabs = tabs.filter((tab) => tab.enabled)
  const requestedTab = searchParams.get('tab') as TenantAccessTab | null
  const activeTab = availableTabs.find((tab) => tab.id === requestedTab)?.id ?? availableTabs[0]?.id

  useEffect(() => {
    if (!availableTabs.length || activeTab === requestedTab) {
      return
    }
    setSearchParams({ tab: activeTab }, { replace: true })
  }, [activeTab, availableTabs.length, requestedTab, setSearchParams])

  if (!tenantId) {
    return <EmptyState title={t('tenantAccess.noTenantTitle', '请先选择组织')} description={t('tenantAccess.noTenantDescription', '用户与访问需要明确的组织上下文。')} />
  }

  if (!availableTabs.length || !activeTab) {
    return <EmptyState title={t('tenantAccess.noAccessTitle', '无访问权限')} description={t('tenantAccess.noAccessDescription', '当前组织角色没有成员或邀请管理权限。')} />
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap gap-2">
        {tabs
          .filter((tab) => tab.enabled)
          .map((tab) => (
            <Button
              key={tab.id}
              type="button"
              variant={tab.id === activeTab ? 'primary' : 'secondary'}
              onClick={() => setSearchParams({ tab: tab.id })}
            >
              {tab.label}
            </Button>
          ))}
      </div>

      {activeTab === 'members' ? <TenantMembersPanel /> : <TenantInvitationsPanel />}
    </div>
  )
}
