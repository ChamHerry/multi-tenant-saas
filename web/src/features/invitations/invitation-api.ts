import { apiRequest } from '@/shared/api/client'
import type { AcceptInvitationResponse, CreateInvitationInput, CreateInvitationResponse, InvitationListResponse } from './invitation-types'

export function listTenantInvitations(tenantId: string, status = '') {
  const query = status ? `?status=${encodeURIComponent(status)}` : ''
  return apiRequest<InvitationListResponse>(`/api/v1/tenants/${tenantId}/invitations${query}`, { tenantId })
}

export function createTenantInvitation(tenantId: string, input: CreateInvitationInput) {
  return apiRequest<CreateInvitationResponse>(`/api/v1/tenants/${tenantId}/invitations`, {
    method: 'POST',
    tenantId,
    body: input,
  })
}

export function resendTenantInvitation(tenantId: string, invitationId: string) {
  return apiRequest<CreateInvitationResponse>(`/api/v1/tenants/${tenantId}/invitations/${invitationId}/resend`, {
    method: 'POST',
    tenantId,
  })
}

export function revokeTenantInvitation(tenantId: string, invitationId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/tenants/${tenantId}/invitations/${invitationId}/revoke`, {
    method: 'POST',
    tenantId,
  })
}

export function listMyInvitations(status = 'pending') {
  const query = status ? `?status=${encodeURIComponent(status)}` : ''
  return apiRequest<InvitationListResponse>(`/api/v1/me/invitations${query}`, { skipTenant: true })
}

export function acceptInvitation(token: string) {
  return apiRequest<AcceptInvitationResponse>('/api/v1/invitations/accept', { method: 'POST', skipTenant: true, body: { token } })
}

export function declineInvitation(invitationId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/me/invitations/${invitationId}/decline`, { method: 'POST', skipTenant: true })
}
