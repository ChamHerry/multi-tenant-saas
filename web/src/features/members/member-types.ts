import { z } from 'zod'
import { tenantRoleSchema } from '@/features/tenants/tenant-types'
import type { TenantMembership } from '@/features/tenants/tenant-types'

export type TenantMembershipWithUser = TenantMembership & {
  email: string
  display_name: string
  user_status: string
}

export const addMemberInputSchema = z.object({
  user_id: z.string().min(1, '请输入用户 ID'),
  role: tenantRoleSchema,
  status: z.string().optional(),
})
export type AddMemberInput = z.infer<typeof addMemberInputSchema>

export const updateMemberInputSchema = z.object({
  role: tenantRoleSchema.optional(),
  status: z.string().optional(),
})
export type UpdateMemberInput = z.infer<typeof updateMemberInputSchema>
