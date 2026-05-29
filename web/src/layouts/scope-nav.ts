import { hasAll, hasAny } from '@/features/access/can'
import type { PlatformPermission, TenantPermission } from '@/features/access/access-types'
import {
  accountNavItems,
  platformNavItems,
  tenantNavItems,
  type MenuItem,
  type PermissionMode,
} from './menu-config'

export type ScopeNavSection = {
  title: string
  tone: 'default' | 'platform'
  items: MenuItem[]
}

type BuildScopeNavSectionsInput = {
  tenantPermissions?: TenantPermission[]
  platformPermissions?: PlatformPermission[]
}

function matchesPermissions<T extends string>(
  owned: readonly T[] | undefined,
  required: readonly T[] | undefined,
  mode: PermissionMode = 'all',
) {
  if (!required || required.length === 0) {
    return true
  }
  return mode === 'any' ? hasAny(owned, required) : hasAll(owned, required)
}

function isItemVisible(
  item: MenuItem,
  tenantPermissions: readonly TenantPermission[] | undefined,
  platformPermissions: readonly PlatformPermission[] | undefined,
) {
  if (item.requiredPlatformPermissions?.length) {
    return matchesPermissions(platformPermissions, item.requiredPlatformPermissions, item.permissionMode)
  }
  return matchesPermissions(tenantPermissions, item.requiredTenantPermissions, item.permissionMode)
}

export function buildScopeNavSections({
  tenantPermissions,
  platformPermissions,
}: BuildScopeNavSectionsInput): ScopeNavSection[] {
  const sections: ScopeNavSection[] = [
    {
      title: '我的账户',
      tone: 'default',
      items: accountNavItems,
    },
    {
      title: '组织管理',
      tone: 'default',
      items: tenantNavItems.filter((item) => isItemVisible(item, tenantPermissions, platformPermissions)),
    },
    {
      title: '平台管理',
      tone: 'platform',
      items: platformNavItems.filter((item) => isItemVisible(item, tenantPermissions, platformPermissions)),
    },
  ]

  return sections.filter((section) => section.items.length > 0)
}
