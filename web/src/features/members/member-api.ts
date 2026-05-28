import { apiRequest } from '@/shared/api/client'
import type { AddMemberInput, TenantMembershipWithUser, UpdateMemberInput } from './member-types'
import type { TenantMembership } from '@/features/tenants/tenant-types'

export function listMembers(tenantId: string) {
  return apiRequest<{ members: TenantMembershipWithUser[] }>(`/api/v1/tenants/${tenantId}/members`, { tenantId })
}

export function addMember(tenantId: string, input: AddMemberInput) {
  return apiRequest<{ member: TenantMembership }>(`/api/v1/tenants/${tenantId}/members`, {
    method: 'POST',
    tenantId,
    body: input,
  })
}

export function updateMember(tenantId: string, userId: string, input: UpdateMemberInput) {
  return apiRequest<{ member: TenantMembership }>(`/api/v1/tenants/${tenantId}/members/${userId}`, {
    method: 'PATCH',
    tenantId,
    body: input,
  })
}

export function removeMember(tenantId: string, userId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/tenants/${tenantId}/members/${userId}`, {
    method: 'DELETE',
    tenantId,
  })
}
