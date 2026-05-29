export type { APIKey as TenantAPIKey, APIKeyTenantGrant, CreatedAPIKey } from '@/features/api-keys/api-key-types'
export { knownScopes, tenantScopes, userScopes } from '@/features/api-keys/api-key-types'

export type CreateTenantAPIKeyInput = {
  name: string
  scopes: string[]
  expires_at?: string
}
