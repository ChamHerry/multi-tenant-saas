import { apiRequest } from '@/shared/api/client'
import type {
  AdminAuditListResponse,
  AdminTenantListResponse,
  AdminUserListResponse,
  PlatformAdminListResponse,
  PlatformSessionResponse,
  SystemConfigListResponse,
  SystemConfigResponse,
  UpsertSystemConfigPayload,
} from './admin-types'
import type { Tenant } from '@/features/tenants/tenant-types'

function qs(params: Record<string, string | number | undefined>) {
  const query = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== '') query.set(key, String(value))
  })
  const value = query.toString()
  return value ? `?${value}` : ''
}

export function getPlatformSession() {
  return apiRequest<PlatformSessionResponse>('/api/v1/admin/session', { skipTenant: true })
}

export function listAdminTenants(params: Record<string, string | number | undefined> = {}) {
  return apiRequest<AdminTenantListResponse>(`/api/v1/admin/tenants${qs(params)}`, { skipTenant: true })
}

export function getAdminTenant(tenantId: string) {
  return apiRequest<{ tenant: Tenant }>(`/api/v1/admin/tenants/${tenantId}`, { skipTenant: true })
}

export function suspendAdminTenant(tenantId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/admin/tenants/${tenantId}/suspend`, { method: 'POST', skipTenant: true })
}

export function restoreAdminTenant(tenantId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/admin/tenants/${tenantId}/restore`, { method: 'POST', skipTenant: true })
}

export function listAdminUsers(params: Record<string, string | number | undefined> = {}) {
  return apiRequest<AdminUserListResponse>(`/api/v1/admin/users${qs(params)}`, { skipTenant: true })
}

export function updateAdminUserStatus(userId: string, status: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/admin/users/${userId}/status`, { method: 'PATCH', skipTenant: true, body: { status } })
}

export function listPlatformAdmins(params: Record<string, string | number | undefined> = {}) {
  return apiRequest<PlatformAdminListResponse>(`/api/v1/admin/platform-admins${qs(params)}`, { skipTenant: true })
}

export function grantPlatformAdmin(userId: string, role: string) {
  return apiRequest<{ ok: boolean }>('/api/v1/admin/platform-admins', { method: 'POST', skipTenant: true, body: { user_id: userId, role } })
}

export function revokePlatformAdmin(userId: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/admin/platform-admins/${userId}`, { method: 'DELETE', skipTenant: true })
}

export function listAdminAuditLogs(params: Record<string, string | number | undefined> = {}) {
  return apiRequest<AdminAuditListResponse>(`/api/v1/admin/audit-logs${qs(params)}`, { skipTenant: true })
}

export function listAdminSystemConfigs(params: Record<string, string | number | undefined> = {}) {
  return apiRequest<SystemConfigListResponse>(`/api/v1/admin/system-config${qs(params)}`, { skipTenant: true })
}

export function getAdminSystemConfig(key: string) {
  return apiRequest<SystemConfigResponse>(`/api/v1/admin/system-config/${encodeURIComponent(key)}`, { skipTenant: true })
}

export function upsertAdminSystemConfig(key: string, payload: UpsertSystemConfigPayload) {
  return apiRequest<SystemConfigResponse>(`/api/v1/admin/system-config/${encodeURIComponent(key)}`, {
    method: 'PUT',
    skipTenant: true,
    body: payload,
  })
}

export function deleteAdminSystemConfig(key: string) {
  return apiRequest<{ ok: boolean }>(`/api/v1/admin/system-config/${encodeURIComponent(key)}`, { method: 'DELETE', skipTenant: true })
}
