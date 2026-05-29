import { apiRequest } from '@/shared/api/client'
import type { CreateTenantAPIKeyInput, CreatedAPIKey, TenantAPIKey } from './tenant-api-key-types'

export function listTenantAPIKeys(tenantId: string) {
  return apiRequest<{ api_keys: TenantAPIKey[] }>(`/api/v1/tenants/${tenantId}/api-keys`, {
    tenantId,
  })
}

export function createTenantAPIKey(tenantId: string, input: CreateTenantAPIKeyInput) {
  return apiRequest<CreatedAPIKey>(`/api/v1/tenants/${tenantId}/api-keys`, {
    method: 'POST',
    tenantId,
    body: input,
  })
}

export function revokeTenantAPIKey(tenantId: string, apiKeyId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/tenants/${tenantId}/api-keys/${apiKeyId}`, {
    method: 'DELETE',
    tenantId,
  })
}
