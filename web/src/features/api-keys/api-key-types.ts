import { z } from 'zod'
import { defaultSchemaTranslator, type SchemaTranslator } from '@/shared/lib/schema-translator'

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

export function createAPIKeyTenantGrantInputSchema(t: SchemaTranslator = defaultSchemaTranslator) {
  return z.object({
    tenant_id: z.string().min(1, t('tenantRequired', 'Select an organization')),
    scopes: z.array(z.string()).min(1, t('scopeRequired', 'Select at least one scope')),
  })
}
export const apiKeyTenantGrantInputSchema = createAPIKeyTenantGrantInputSchema()
export type CreateAPIKeyTenantGrantInput = z.infer<ReturnType<typeof createAPIKeyTenantGrantInputSchema>>

export function createPersonalAPIKeyInputSchema(t: SchemaTranslator = defaultSchemaTranslator) {
  const grantSchema = createAPIKeyTenantGrantInputSchema(t)
  return z.object({
    name: z.string().min(1, t('nameRequired', 'Enter a name')).max(100, t('nameMax', 'Name must be at most {max} characters', { max: 100 })),
    scopes: z.array(z.string()).min(1, t('scopeRequired', 'Select at least one scope')),
    grants: z.array(grantSchema).optional(),
    expires_at: z.string().optional(),
  })
}
export const personalAPIKeyInputSchema = createPersonalAPIKeyInputSchema()
export type CreatePersonalAPIKeyInput = z.infer<ReturnType<typeof createPersonalAPIKeyInputSchema>>

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
