import type { User } from '@/features/auth/auth-types'
import type { AuditLog } from '@/features/audit/audit-types'
import type { Tenant } from '@/features/tenants/tenant-types'
import type { PlatformPermission } from '@/features/access/access-types'

export type { PlatformPermission }

export type PlatformAdminContext = {
  user_id: string
  role: string
  status: string
  permissions: PlatformPermission[]
}

export type PlatformAdmin = {
  user_id: string
  role: string
  status: string
  created_by_user_id?: string
  created_at: string
  updated_at: string
}

export type PlanEntitlement = {
  plan: string
  feature_key: string
  enabled: boolean
  limit_value?: number
  metadata: Record<string, unknown>
  created_at: string
  updated_at: string
}

export type PlatformSessionResponse = {
  platform_admin: PlatformAdminContext
}

export type AdminTenantListResponse = {
  items: Tenant[]
  total: number
}

export type AdminUserListResponse = {
  items: User[]
  total: number
}

export type PlatformAdminListResponse = {
  items: PlatformAdmin[]
  total: number
}

export type AdminAuditListResponse = {
  logs: AuditLog[]
  total: number
}

export type AdminPlanListResponse = {
  items: PlanEntitlement[]
  total: number
}
