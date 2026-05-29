import { z } from 'zod'
export type { APIKey as TenantAPIKey, APIKeyTenantGrant, CreatedAPIKey } from '@/features/api-keys/api-key-types'
export { knownScopes, tenantScopes, userScopes } from '@/features/api-keys/api-key-types'

export const createTenantAPIKeyInputSchema = z.object({
  name: z.string().min(1, '请输入名称').max(100, '名称最多 100 字符'),
  scopes: z.array(z.string()).min(1, '至少选择一个 scope'),
  expires_at: z.string().optional(),
})
export type CreateTenantAPIKeyInput = z.infer<typeof createTenantAPIKeyInputSchema>
