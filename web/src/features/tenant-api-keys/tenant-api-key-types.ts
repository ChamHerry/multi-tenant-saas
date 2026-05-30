import { z } from 'zod'
import { defaultSchemaTranslator, type SchemaTranslator } from '@/shared/lib/schema-translator'
export type { APIKey as TenantAPIKey, APIKeyTenantGrant, CreatedAPIKey } from '@/features/api-keys/api-key-types'
export { knownScopes, tenantScopes, userScopes } from '@/features/api-keys/api-key-types'

export function createTenantAPIKeyInputSchema(t: SchemaTranslator = defaultSchemaTranslator) {
  return z.object({
    name: z.string().min(1, t('nameRequired', 'Enter a name')).max(100, t('nameMax', 'Name must be at most {max} characters', { max: 100 })),
    scopes: z.array(z.string()).min(1, t('scopeRequired', 'Select at least one scope')),
    expires_at: z.string().optional(),
  })
}
export const tenantAPIKeyInputSchema = createTenantAPIKeyInputSchema()
export type CreateTenantAPIKeyInput = z.infer<ReturnType<typeof createTenantAPIKeyInputSchema>>
