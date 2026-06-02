import { apiRequest } from '@/shared/api/client'
import type { ActionResponse, AuthSessionResponse, ChangePasswordPayload, LoginPayload, MeResponse } from './auth-types'
import type { TenantMembershipWithTenant } from '@/features/tenants/tenant-types'

export type RegisterBody = {
  email: string
  password: string
  display_name?: string
}

export function login(payload: LoginPayload) {
  return apiRequest<MeResponse>('/api/v1/auth/login', { method: 'POST', body: payload, skipAuth: true, skipTenant: true })
}

export function register(payload: RegisterBody) {
  return apiRequest<MeResponse>('/api/v1/auth/register', { method: 'POST', body: payload, skipAuth: true, skipTenant: true })
}

export function logout() {
  return apiRequest<ActionResponse>('/api/v1/auth/logout', { method: 'POST', skipTenant: true })
}

export function getAuthSession() {
  return apiRequest<AuthSessionResponse>('/api/v1/auth/session', { skipTenant: true })
}

export function changePassword(payload: Omit<ChangePasswordPayload, 'confirm_new_password'>) {
  return apiRequest<ActionResponse>('/api/v1/auth/password/change', { method: 'POST', body: payload, skipTenant: true })
}

export function getMe() {
  return apiRequest<MeResponse>('/api/v1/me', { skipTenant: true })
}

export function getMyTenants() {
  return apiRequest<{ tenants: TenantMembershipWithTenant[] }>('/api/v1/me/tenants', { skipTenant: true })
}

export function verifyEmail(token: string) {
  return apiRequest<{ ok: boolean; message: string }>(
    '/api/v1/auth/verify-email',
    { method: 'POST', body: { token }, skipAuth: true, skipTenant: true },
  )
}

export function resendVerification(email: string) {
  return apiRequest<{ ok: boolean; message: string }>(
    '/api/v1/auth/resend-verification',
    { method: 'POST', body: { email }, skipAuth: true, skipTenant: true },
  )
}
