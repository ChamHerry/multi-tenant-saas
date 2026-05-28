export type APIKey = {
  id: string
  tenant_id: string
  user_id: string
  name: string
  key_prefix: string
  scopes: string[]
  last_used_at?: string
  expires_at?: string
  created_at: string
  revoked_at?: string
}

export type CreateAPIKeyInput = {
  name: string
  scopes: string[]
  expires_at?: string
}

export type CreatedAPIKey = {
  api_key: APIKey
  raw_key: string
}

export const knownScopes = [
  'tenant:read',
  'tenant:manage',
  'member:read',
  'member:manage',
  'tenant:invitation:manage',
  'tenant:audit:read',
  'tenant:billing:read',
  'tenant:billing:manage',
  'api_key:manage',
] as const
