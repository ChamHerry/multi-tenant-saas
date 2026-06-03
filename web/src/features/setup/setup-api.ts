import { apiRaw } from '@/shared/api/client'
import type { CompleteSetupPayload, CompleteSetupResponse, SetupState } from './setup-types'

export function getSetupState() {
  return apiRaw<SetupState>('/api/v1/setup/state')
}

export function completeSetup(payload: CompleteSetupPayload) {
  return apiRaw<CompleteSetupResponse>('/api/v1/setup/complete', { method: 'POST', body: payload })
}
