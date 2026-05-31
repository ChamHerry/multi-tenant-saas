import { z } from 'zod'
import {
  defaultTenantSchemaTranslator,
  tenantValidationFallback,
  type SchemaTranslator,
} from '@/features/tenants/tenant-i18n'
import { tenantRoleSchema } from '@/features/tenants/tenant-types'
import type { TenantMembership } from '@/features/tenants/tenant-types'

export type TenantMembershipWithUser = TenantMembership & {
  email: string
  display_name: string
  user_status: string
}

export function createAddMemberInputSchema(t: SchemaTranslator = defaultTenantSchemaTranslator) {
  return z.object({
    user_id: z.string().min(1, t('memberUserIdRequired', tenantValidationFallback.memberUserIdRequired)),
    role: tenantRoleSchema,
    status: z.string().optional(),
  })
}
export const addMemberInputSchema = createAddMemberInputSchema()
export type AddMemberInput = z.infer<ReturnType<typeof createAddMemberInputSchema>>

export const updateMemberInputSchema = z.object({
  role: tenantRoleSchema.optional(),
  status: z.string().optional(),
})
export type UpdateMemberInput = z.infer<typeof updateMemberInputSchema>
