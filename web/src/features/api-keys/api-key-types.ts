import { z } from 'zod'

export type APIKeyTenantGrant = {
  id: string
  api_key_id: string
  tenant_id: string
  tenant_slug?: string
  tenant_name?: string
  user_id?: string
  key_name?: string
  key_prefix?: string
  scopes: string[]
  status: 'active' | 'revoked' | string
  granted_by_user_id?: string
  revoked_by_user_id?: string
  created_at: string
  updated_at: string
  revoked_at?: string
}

export type APIKey = {
  id: string
  tenant_id?: string
  user_id: string
  name: string
  key_type: 'personal' | string
  key_prefix: string
  scopes: string[]
  last_used_at?: string
  expires_at?: string
  created_at: string
  revoked_at?: string
  created_by_user_id?: string
  tenant_grants?: APIKeyTenantGrant[]
}

export const createAPIKeyTenantGrantInputSchema = z.object({
  tenant_id: z.string().min(1, '请选择租户'),
  scopes: z.array(z.string()).min(1, '至少选择一个 scope'),
})
export type CreateAPIKeyTenantGrantInput = z.infer<typeof createAPIKeyTenantGrantInputSchema>

export const createPersonalAPIKeyInputSchema = z.object({
  name: z.string().min(1, '请输入名称').max(100, '名称最多 100 字符'),
  scopes: z.array(z.string()).min(1, '至少选择一个 scope'),
  grants: z.array(createAPIKeyTenantGrantInputSchema).optional(),
  expires_at: z.string().optional(),
})
export type CreatePersonalAPIKeyInput = z.infer<typeof createPersonalAPIKeyInputSchema>

export type CreatedAPIKey = {
  api_key: APIKey
  raw_key: string
}

export const tenantScopes = [
  'tenant:read',
  'tenant:manage',
  'member:read',
  'member:manage',
  'tenant:invitation:manage',
  'tenant:audit:read',
] as const

export const userScopes = [
  'user:read',
  'user:tenant:read',
  'user:security:read',
  'api_key:self_manage',
] as const

export const knownScopes = [...userScopes, ...tenantScopes] as const
