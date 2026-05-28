import { apiRequest } from '@/shared/api/client'
import type { APIKey, CreateAPIKeyInput, CreatedAPIKey } from './api-key-types'

export function listAPIKeys(tenantId: string) {
  return apiRequest<{ api_keys: APIKey[] }>('/api/v1/api-keys', { tenantId })
}

export function createAPIKey(tenantId: string, input: CreateAPIKeyInput) {
  return apiRequest<CreatedAPIKey>('/api/v1/api-keys', {
    method: 'POST',
    tenantId,
    body: input,
  })
}

export function revokeAPIKey(tenantId: string, apiKeyId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/api-keys/${apiKeyId}`, {
    method: 'DELETE',
    tenantId,
  })
}
