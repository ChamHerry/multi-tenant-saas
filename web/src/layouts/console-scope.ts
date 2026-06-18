import type { PlatformPermission, TenantAccess } from '@/features/access/access-types'
import { hasAny } from '@/features/access/can'

export const CONSOLE_SCOPE_TENANT = 'tenant'
export const CONSOLE_SCOPE_PLATFORM = 'platform'

export type ConsoleScope = typeof CONSOLE_SCOPE_TENANT | typeof CONSOLE_SCOPE_PLATFORM

export const TENANT_CONSOLE_ROOT = '/console'
export const PLATFORM_CONSOLE_ROOT = '/platform-admin-console'
export const TENANT_CONSOLE_HOME = TENANT_CONSOLE_ROOT
export const PLATFORM_CONSOLE_HOME = `${PLATFORM_CONSOLE_ROOT}/tenants`

const PLATFORM_HOME_BY_PERMISSION: Array<{
  permissions: PlatformPermission[]
  to: string
}> = [
  { permissions: ['platform:tenant:read'], to: `${PLATFORM_CONSOLE_ROOT}/tenants` },
  { permissions: ['platform:user:read'], to: `${PLATFORM_CONSOLE_ROOT}/users` },
  { permissions: ['platform:audit:read'], to: `${PLATFORM_CONSOLE_ROOT}/audit-logs` },
  { permissions: ['platform:config:read'], to: `${PLATFORM_CONSOLE_ROOT}/system-config` },
]

export function isTenantConsolePath(pathname: string) {
  return pathname === TENANT_CONSOLE_ROOT || pathname.startsWith(`${TENANT_CONSOLE_ROOT}/`)
}

export function isPlatformConsolePath(pathname: string) {
  return pathname === PLATFORM_CONSOLE_ROOT || pathname.startsWith(`${PLATFORM_CONSOLE_ROOT}/`)
}

export function getConsoleScopeFromPath(pathname: string): ConsoleScope {
  return isPlatformConsolePath(pathname) ? CONSOLE_SCOPE_PLATFORM : CONSOLE_SCOPE_TENANT
}

export function getPlatformConsoleHome(permissions?: readonly PlatformPermission[]) {
  return PLATFORM_HOME_BY_PERMISSION.find((entry) => hasAny(permissions, entry.permissions))?.to ?? PLATFORM_CONSOLE_HOME
}

export function hasTenantConsoleAccess(tenants?: readonly TenantAccess[]) {
  return Boolean(
    tenants?.some((tenant) => tenant.status === 'active' && tenant.tenant_status === 'active'),
  )
}

export function hasPlatformConsoleAccess(permissions?: readonly PlatformPermission[]) {
  return PLATFORM_HOME_BY_PERMISSION.some((entry) => hasAny(permissions, entry.permissions))
}
