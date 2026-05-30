import {
  Building2,
  Inbox,
  KeyRound,
  LayoutDashboard,
  ScrollText,
  Settings2,
  ShieldCheck,
  UserCog,
  Users,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import type { PlatformPermission, TenantPermission } from '@/features/access/access-types'

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
]

export const tenantNavItems: MenuItem[] = [
  { to: '/', labelKey: 'nav.items.dashboard', icon: LayoutDashboard },
  { to: '/tenants', labelKey: 'nav.items.tenants', icon: Building2 },
  {
    to: '/tenant/settings',
    labelKey: 'nav.items.tenantSettings',
    icon: Settings2,
    requiredTenantPermissions: ['tenant:read'],
  },
  {
    to: '/tenant/access?tab=members',
    labelKey: 'nav.items.tenantAccess',
    icon: Users,
    requiredTenantPermissions: ['member:read', 'tenant:invitation:manage'],
    permissionMode: 'any',
  },
  {
    to: '/tenant/api-keys',
    labelKey: 'nav.items.tenantApiKeys',
    icon: KeyRound,
    requiredTenantPermissions: ['tenant:read'],
  },
  {
    to: '/tenant/audit-logs',
    labelKey: 'nav.items.tenantAudit',
    icon: ScrollText,
    requiredTenantPermissions: ['tenant:audit:read'],
  },
]

export const platformNavItems: MenuItem[] = [
  {
    to: '/admin/tenants',
    labelKey: 'nav.items.platformTenants',
    icon: ShieldCheck,
    requiredPlatformPermissions: ['platform:tenant:read'],
  },
  {
    to: '/admin/users',
    labelKey: 'nav.items.platformUsers',
    icon: UserCog,
    requiredPlatformPermissions: ['platform:user:read'],
  },
  {
    to: '/admin/audit-logs',
    labelKey: 'nav.items.platformAudit',
    icon: ScrollText,
    requiredPlatformPermissions: ['platform:audit:read'],
  },
]
