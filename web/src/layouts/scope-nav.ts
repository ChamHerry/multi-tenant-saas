import { hasAll, hasAny } from '@/features/access/can'
import type { PlatformPermission, TenantPermission } from '@/features/access/access-types'
import {
  CONSOLE_SCOPE_PLATFORM,
  type ConsoleScope,
} from './console-scope'
import {
  accountNavItems,
  platformNavItems,
  tenantNavItems,
  type MenuItem,
  type PermissionMode,
} from './menu-config'

export type ScopeNavSection = {
  titleKey: string
  tone: 'default' | 'platform'
  items: MenuItem[]
}

type BuildScopeNavSectionsInput = {
  tenantPermissions?: TenantPermission[]
  platformPermissions?: PlatformPermission[]
  consoleScope: ConsoleScope
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
  consoleScope,
}: BuildScopeNavSectionsInput): ScopeNavSection[] {
  const tenantItems = tenantNavItems.filter((item) => isItemVisible(item, tenantPermissions, platformPermissions))
  const platformItems = platformNavItems.filter((item) => isItemVisible(item, tenantPermissions, platformPermissions))
  const sections: ScopeNavSection[] = []

  if (consoleScope === CONSOLE_SCOPE_PLATFORM) {
    if (platformItems.length > 0) {
      sections.push({
        titleKey: 'nav.sections.platform',
        tone: 'platform',
        items: platformItems,
      })
    }
    return sections.filter((section) => section.items.length > 0)
  }

  sections.push({
    titleKey: 'nav.sections.account',
    tone: 'default',
    items: accountNavItems,
  })

  if (tenantItems.length > 0) {
    sections.push({
      titleKey: 'nav.sections.tenant',
      tone: 'default',
      items: tenantItems,
    })
  }
  return sections.filter((section) => section.items.length > 0)
}
