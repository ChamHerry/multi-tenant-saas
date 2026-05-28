import type { TenantMembership, TenantRole } from '@/features/tenants/tenant-types'

export type TenantInvitation = {
  id: string
  tenant_id: string
  tenant_name?: string
  tenant_slug?: string
  invitee_email: string
  invitee_user_id?: string
  role: TenantRole
  status: 'pending' | 'accepted' | 'declined' | 'revoked' | 'expired'
  invited_by_user_id?: string
  accepted_by_user_id?: string
  message: string
  expires_at: string
  accepted_at?: string
  declined_at?: string
  revoked_at?: string
  resent_at?: string
  created_at: string
  updated_at: string
}

export type InvitationListResponse = {
  invitations: TenantInvitation[]
  total: number
}

export type CreateInvitationInput = {
  invitee_email: string
  role: TenantRole
  message?: string
  expires_hours?: number
}

export type CreateInvitationResponse = {
  invitation: TenantInvitation
  token?: string
  accept_url?: string
}

export type AcceptInvitationResponse = {
  member: TenantMembership
}
