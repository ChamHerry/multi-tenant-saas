import { z } from 'zod'
import type { TenantPermission } from '@/features/access/access-types'

export const tenantRoleSchema = z.enum(['owner', 'admin', 'member', 'viewer'])
export type TenantRole = z.infer<typeof tenantRoleSchema>

export const tenantSchema = z.object({
  id: z.string(),
  name: z.string(),
  slug: z.string(),
  status: z.string(),
  owner_user_id: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
})
export type Tenant = z.infer<typeof tenantSchema>

export const tenantMembershipSchema = z.object({
  id: z.string(),
  tenant_id: z.string(),
  user_id: z.string(),
  role: tenantRoleSchema,
  status: z.string(),
  invited_by_user_id: z.string().optional(),
  joined_at: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
})
export type TenantMembership = z.infer<typeof tenantMembershipSchema>

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
  auth_type: string
  scopes?: string[]
  permissions?: TenantPermission[]
  api_key_id?: string
  request_id: string
}

export const TENANT_SLUG_MAX_LENGTH = 80
const TENANT_SLUG_PATTERN = /^[a-z0-9][a-z0-9-]{1,78}[a-z0-9]$/

export const createTenantInputSchema = z.object({
  name: z.string().min(1, '请输入组织名称').max(100, '组织名称最多 100 字符'),
  slug: z
    .string()
    .min(1, '请输入 Slug')
    .regex(TENANT_SLUG_PATTERN, 'Slug 需为 3-80 位小写字母、数字或短横线，且首尾必须是字母或数字'),
})
export type CreateTenantInput = z.infer<typeof createTenantInputSchema>

export const updateTenantInputSchema = z.object({
  name: z.string().min(1, '组织名称不能为空').max(100, '组织名称最多 100 字符').optional(),
  slug: z
    .string()
    .min(1, 'Slug 不能为空')
    .regex(TENANT_SLUG_PATTERN, 'Slug 需为 3-80 位小写字母、数字或短横线，且首尾必须是字母或数字')
    .optional(),
})
export type UpdateTenantInput = z.infer<typeof updateTenantInputSchema>
