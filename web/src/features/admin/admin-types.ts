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

export type SystemConfigValueType = 'string' | 'number' | 'bool' | 'json' | 'secret'

export type SystemConfigItem = {
  key: string
  value: string
  value_type: SystemConfigValueType
  description: string
  category: string
  is_encrypted: boolean
  is_secret: boolean
  has_value: boolean
  masked_value?: string
  created_at: string
  updated_at: string
}

export type SystemConfigListResponse = {
  items: SystemConfigItem[]
  total: number
}

export type SystemConfigResponse = {
  config: SystemConfigItem
}

export type UpsertSystemConfigPayload = {
  value: string
  value_provided: boolean
  value_type: SystemConfigValueType
  description: string
}
