import {
  Building2,
  Inbox,
  KeyRound,
  LayoutDashboard,
  Monitor,
  ScrollText,
  Settings2,
  ShieldCheck,
  UserCog,
  Users,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import type { PlatformPermission, TenantPermission } from '@/features/access/access-types'
import { PLATFORM_CONSOLE_ROOT, TENANT_CONSOLE_HOME, TENANT_CONSOLE_ROOT } from './console-scope'

export type PermissionMode = 'all' | 'any'

export type MenuItem = {
  to: string
  labelKey: string
  icon: LucideIcon
  requiredTenantPermissions?: TenantPermission[]
  requiredPlatformPermissions?: PlatformPermission[]
  permissionMode?: PermissionMode
}

export const accountNavItems: MenuItem[] = [
  { to: '/me/invitations', labelKey: 'nav.items.myInvitations', icon: Inbox },
  { to: '/me/security', labelKey: 'nav.items.securityEvents', icon: ShieldCheck },
  { to: '/me/sessions', labelKey: 'nav.items.sessions', icon: Monitor },
]

export const tenantNavItems: MenuItem[] = [
  { to: TENANT_CONSOLE_HOME, labelKey: 'nav.items.dashboard', icon: LayoutDashboard },
  { to: `${TENANT_CONSOLE_ROOT}/tenants`, labelKey: 'nav.items.tenants', icon: Building2 },
  {
    to: `${TENANT_CONSOLE_ROOT}/settings`,
    labelKey: 'nav.items.tenantSettings',
    icon: Settings2,
    requiredTenantPermissions: ['tenant:read'],
  },
  {
    to: `${TENANT_CONSOLE_ROOT}/access?tab=members`,
    labelKey: 'nav.items.tenantAccess',
    icon: Users,
    requiredTenantPermissions: ['member:read', 'tenant:invitation:manage'],
    permissionMode: 'any',
  },
  {
    to: `${TENANT_CONSOLE_ROOT}/api-keys`,
    labelKey: 'nav.items.tenantApiKeys',
    icon: KeyRound,
    requiredTenantPermissions: ['tenant:read'],
  },
  {
    to: `${TENANT_CONSOLE_ROOT}/audit-logs`,
    labelKey: 'nav.items.tenantAudit',
    icon: ScrollText,
    requiredTenantPermissions: ['tenant:audit:read'],
  },
]

export const platformNavItems: MenuItem[] = [
  {
    to: `${PLATFORM_CONSOLE_ROOT}/tenants`,
    labelKey: 'nav.items.platformTenants',
    icon: ShieldCheck,
    requiredPlatformPermissions: ['platform:tenant:read'],
  },
  {
    to: `${PLATFORM_CONSOLE_ROOT}/users`,
    labelKey: 'nav.items.platformUsers',
    icon: UserCog,
    requiredPlatformPermissions: ['platform:user:read'],
  },
  {
    to: `${PLATFORM_CONSOLE_ROOT}/audit-logs`,
    labelKey: 'nav.items.platformAudit',
    icon: ScrollText,
    requiredPlatformPermissions: ['platform:audit:read'],
  },
  {
    to: `${PLATFORM_CONSOLE_ROOT}/system-config`,
    labelKey: 'nav.items.systemConfig',
    icon: Settings2,
    requiredPlatformPermissions: ['platform:config:read'],
  },
]
