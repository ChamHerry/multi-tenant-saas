import type { User } from '@/features/auth/auth-types'
import type { TenantRole } from '@/features/tenants/tenant-types'

export type TenantPermission =
  | 'tenant:read'
  | 'tenant:manage'
  | 'member:read'
  | 'member:manage'
  | 'tenant:invitation:manage'
  | 'tenant:audit:read'

export type PlatformPermission =
  | 'platform:tenant:read'
  | 'platform:tenant:manage'
  | 'platform:user:read'
  | 'platform:user:manage'
  | 'platform:audit:read'
  | 'platform:admin:manage'
  | 'platform:config:read'
  | 'platform:config:manage'

export type TenantAccess = {
  tenant_id: string
  tenant_name: string
  tenant_slug: string
  tenant_status: string
  role: TenantRole
  status: string
  permissions: TenantPermission[]
}

export type PlatformAdminAccess = {
  user_id: string
  role: string
  status: string
  permissions: PlatformPermission[]
}

export type AccessSnapshot = {
  user: User
  tenants: TenantAccess[]
  platform_admin: PlatformAdminAccess | null
}
