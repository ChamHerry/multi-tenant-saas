import { apiRequest } from '@/shared/api/client'
import type {
  TOTPDisableBody,
  TOTPEnableBody,
  TOTPEnableResponse,
  TOTPRegenerateBody,
  TOTPRegenerateResponse,
  TOTPSetupBody,
  TOTPSetupResponse,
  TOTPStatusResponse,
} from './totp-types'

export function setupTOTP(payload: TOTPSetupBody) {
  return apiRequest<TOTPSetupResponse>('/api/v1/me/totp/setup', {
    method: 'POST',
    body: payload,
    skipTenant: true,
  })
}

export function enableTOTP(payload: TOTPEnableBody) {
  return apiRequest<TOTPEnableResponse>('/api/v1/me/totp/enable', {
    method: 'POST',
    body: payload,
    skipTenant: true,
  })
}

export function disableTOTP(payload: TOTPDisableBody) {
  return apiRequest<{ ok: boolean }>('/api/v1/me/totp/disable', {
    method: 'POST',
    body: payload,
    skipTenant: true,
  })
}

export function regenerateBackupCodes(payload: TOTPRegenerateBody) {
  return apiRequest<TOTPRegenerateResponse>('/api/v1/me/totp/backup-codes/regenerate', {
    method: 'POST',
    body: payload,
    skipTenant: true,
  })
}

export function getTOTPStatus() {
  return apiRequest<TOTPStatusResponse>('/api/v1/me/totp/status', { skipTenant: true })
}
