import { useEffect, useMemo } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useCanTenant } from '@/features/access/access-hooks'
import { TenantInvitationsPanel } from '@/features/tenant-access/TenantInvitationsPanel'
import { TenantMembersPanel } from '@/features/tenant-access/TenantMembersPanel'
import { useTenantStore } from '@/features/tenants/tenant-store'
import { Button } from '@/shared/ui/Button'
import { EmptyState } from '@/shared/ui/EmptyState'

const tabLabels = {
  members: '成员',
  invitations: '邀请',
} as const

type TenantAccessTab = keyof typeof tabLabels

export function TenantAccessPage() {
  const tenantId = useTenantStore((state) => state.currentTenantId)
  const canReadMembers = useCanTenant('member:read', tenantId)
  const canManageInvitations = useCanTenant('tenant:invitation:manage', tenantId)
  const [searchParams, setSearchParams] = useSearchParams()

  const tabs = useMemo(
    () => [
      { id: 'members' as const, label: tabLabels.members, enabled: canReadMembers },
      { id: 'invitations' as const, label: tabLabels.invitations, enabled: canManageInvitations },
    ],
    [canManageInvitations, canReadMembers],
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
    return <EmptyState title="请先选择组织" description="用户与访问需要明确的组织上下文。" />
  }

  if (!availableTabs.length || !activeTab) {
    return <EmptyState title="无访问权限" description="当前组织角色没有成员或邀请管理权限。" />
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
