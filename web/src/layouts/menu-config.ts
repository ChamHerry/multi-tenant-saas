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
  label: string
  icon: LucideIcon
  requiredTenantPermissions?: TenantPermission[]
  requiredPlatformPermissions?: PlatformPermission[]
  permissionMode?: PermissionMode
}

export const accountNavItems: MenuItem[] = [
  { to: '/me/invitations', label: '我的邀请', icon: Inbox },
  { to: '/me/security', label: '安全事件', icon: ShieldCheck },
]

export const tenantNavItems: MenuItem[] = [
  { to: '/', label: '组织概览', icon: LayoutDashboard },
  { to: '/tenants', label: '我的组织', icon: Building2 },
  {
    to: '/tenant/settings',
    label: '组织设置',
    icon: Settings2,
    requiredTenantPermissions: ['tenant:read'],
  },
  {
    to: '/tenant/access?tab=members',
    label: '用户与访问',
    icon: Users,
    requiredTenantPermissions: ['member:read', 'tenant:invitation:manage'],
    permissionMode: 'any',
  },
  {
    to: '/tenant/api-keys',
    label: '组织 API Keys',
    icon: KeyRound,
    requiredTenantPermissions: ['tenant:read'],
  },
  {
    to: '/tenant/audit-logs',
    label: '组织审计',
    icon: ScrollText,
    requiredTenantPermissions: ['tenant:audit:read'],
  },
]

export const platformNavItems: MenuItem[] = [
  {
    to: '/admin/tenants',
    label: '平台组织',
    icon: ShieldCheck,
    requiredPlatformPermissions: ['platform:tenant:read'],
  },
  {
    to: '/admin/users',
    label: '平台用户与管理员',
    icon: UserCog,
    requiredPlatformPermissions: ['platform:user:read'],
  },
  {
    to: '/admin/audit-logs',
    label: '平台审计',
    icon: ScrollText,
    requiredPlatformPermissions: ['platform:audit:read'],
  },
]
