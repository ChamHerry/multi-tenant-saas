import type { TenantPermission } from '@/features/access/access-types'

export type Tenant = {
  id: string
  name: string
  slug: string
  plan: string
  status: string
  owner_user_id?: string
  created_at: string
  updated_at: string
}

export type TenantMembership = {
  id: string
  tenant_id: string
  user_id: string
  role: TenantRole
  status: string
  invited_by_user_id?: string
  joined_at?: string
  created_at: string
  updated_at: string
}

export type TenantRole = 'owner' | 'admin' | 'member' | 'viewer'

export type TenantMembershipWithTenant = TenantMembership & {
  tenant_name: string
  tenant_slug: string
  tenant_status: string
}

export type TenantContext = {
  tenant_id: string
  user_id: string
  role: TenantRole
  tenant_slug: string
  tenant_plan: string
  auth_type: string
  scopes?: string[]
  permissions?: TenantPermission[]
  api_key_id?: string
  request_id: string
}

export type CreateTenantInput = {
  name: string
  slug: string
  plan?: string
}

export type UpdateTenantInput = Partial<CreateTenantInput>
