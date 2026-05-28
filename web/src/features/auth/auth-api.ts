import { apiRequest } from '@/shared/api/client'
import type { ActionResponse, AuthSessionResponse, ChangePasswordPayload, LoginPayload, MeResponse, RegisterPayload } from './auth-types'
import type { TenantMembershipWithTenant } from '@/features/tenants/tenant-types'

export function login(payload: LoginPayload) {
  return apiRequest<MeResponse>('/api/v1/auth/login', { method: 'POST', body: payload, skipAuth: true, skipTenant: true })
}

export function register(payload: RegisterPayload) {
  return apiRequest<MeResponse>('/api/v1/auth/register', { method: 'POST', body: payload, skipAuth: true, skipTenant: true })
}

export function logout() {
  return apiRequest<ActionResponse>('/api/v1/auth/logout', { method: 'POST', skipTenant: true })
}

export function getAuthSession() {
  return apiRequest<AuthSessionResponse>('/api/v1/auth/session', { skipTenant: true })
}

export function changePassword(payload: ChangePasswordPayload) {
  return apiRequest<ActionResponse>('/api/v1/auth/password/change', { method: 'POST', body: payload, skipTenant: true })
}

export function getMe() {
  return apiRequest<MeResponse>('/api/v1/me', { skipTenant: true })
}

export function getMyTenants() {
  return apiRequest<{ tenants: TenantMembershipWithTenant[] }>('/api/v1/me/tenants', { skipTenant: true })
}
