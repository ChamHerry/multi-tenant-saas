import { apiRequest } from '@/shared/api/client'
import type { AccessSnapshot } from './access-types'

export function getAccessSnapshot() {
  return apiRequest<AccessSnapshot>('/api/v1/me/access', { skipTenant: true })
}
