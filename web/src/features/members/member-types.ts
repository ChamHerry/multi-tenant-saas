import type { TenantMembership, TenantRole } from '@/features/tenants/tenant-types'

export type TenantMembershipWithUser = TenantMembership & {
  email: string
  display_name: string
  user_status: string
}

export type AddMemberInput = {
  user_id: string
  role: TenantRole
  status?: string
}

export type UpdateMemberInput = {
  role?: TenantRole
  status?: string
}
