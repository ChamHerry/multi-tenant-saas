import { z } from 'zod'
import type { TenantPermission } from '@/features/access/access-types'
import {
  defaultTenantSchemaTranslator,
  tenantValidationFallback,
  type SchemaTranslator,
} from './tenant-i18n'

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

export function createCreateTenantInputSchema(t: SchemaTranslator = defaultTenantSchemaTranslator) {
  return z.object({
    name: z
      .string()
      .min(1, t('tenantNameRequired', tenantValidationFallback.tenantNameRequired))
      .max(100, t('tenantNameMax', tenantValidationFallback.tenantNameMax, { max: 100 })),
    slug: z
      .string()
      .min(1, t('tenantSlugRequired', tenantValidationFallback.tenantSlugRequired))
      .regex(TENANT_SLUG_PATTERN, t('tenantSlugPattern', tenantValidationFallback.tenantSlugPattern)),
  })
}

export const createTenantInputSchema = createCreateTenantInputSchema()
export type CreateTenantInput = z.infer<ReturnType<typeof createCreateTenantInputSchema>>

export function createUpdateTenantInputSchema(t: SchemaTranslator = defaultTenantSchemaTranslator) {
  return z.object({
    name: z
      .string()
      .min(1, t('tenantNameCannotBeEmpty', tenantValidationFallback.tenantNameCannotBeEmpty))
      .max(100, t('tenantNameMax', tenantValidationFallback.tenantNameMax, { max: 100 }))
      .optional(),
    slug: z
      .string()
      .min(1, t('tenantSlugCannotBeEmpty', tenantValidationFallback.tenantSlugCannotBeEmpty))
      .regex(TENANT_SLUG_PATTERN, t('tenantSlugPattern', tenantValidationFallback.tenantSlugPattern))
      .optional(),
  })
}

export const updateTenantInputSchema = createUpdateTenantInputSchema()
export type UpdateTenantInput = z.infer<ReturnType<typeof createUpdateTenantInputSchema>>
