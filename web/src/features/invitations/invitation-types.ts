import { z } from 'zod'
import { tenantRoleSchema } from '@/features/tenants/tenant-types'
import type { TenantMembership } from '@/features/tenants/tenant-types'

export type TenantInvitation = {
  id: string
  tenant_id: string
  tenant_name?: string
  tenant_slug?: string
  invitee_email: string
  invitee_user_id?: string
  role: z.infer<typeof tenantRoleSchema>
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

export const createInvitationSchema = z.object({
  invitee_email: z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
  role: tenantRoleSchema,
  message: z.string().max(500, '留言最多 500 字符').optional(),
  expires_hours: z.number().int().min(1).max(720).optional(),
})
export type CreateInvitationInput = z.infer<typeof createInvitationSchema>

export type CreateInvitationResponse = {
  invitation: TenantInvitation
  token?: string
  accept_url?: string
}

export type AcceptInvitationResponse = {
  member: TenantMembership
}

export const acceptInvitationTokenSchema = z.object({
  token: z.string().min(1, '请输入邀请 token'),
})
