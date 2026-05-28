import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { accessKeys } from '@/features/access/access-hooks'
import {
  getPlatformSession,
  grantPlatformAdmin,
  listAdminAuditLogs,
  listAdminTenants,
  listAdminUsers,
  listPlatformAdmins,
  restoreAdminTenant,
  revokePlatformAdmin,
  suspendAdminTenant,
  updateAdminUserStatus,
} from './admin-api'

export const adminKeys = {
  session: ['admin', 'session'] as const,
  tenants: (params: Record<string, string | number | undefined> = {}) => ['admin', 'tenants', params] as const,
  users: (params: Record<string, string | number | undefined> = {}) => ['admin', 'users', params] as const,
  platformAdmins: (params: Record<string, string | number | undefined> = {}) => ['admin', 'platform-admins', params] as const,
  audit: (params: Record<string, string | number | undefined> = {}) => ['admin', 'audit', params] as const,
}

export function usePlatformSession() {
  return useQuery({ queryKey: adminKeys.session, queryFn: getPlatformSession, retry: false })
}

export function useAdminTenants(params: Record<string, string | number | undefined> = {}) {
  return useQuery({ queryKey: adminKeys.tenants(params), queryFn: () => listAdminTenants(params), retry: false })
}

export function useAdminTenantActions(params: Record<string, string | number | undefined> = {}) {
  const queryClient = useQueryClient()
  const invalidate = () => void queryClient.invalidateQueries({ queryKey: adminKeys.tenants(params) })
  return {
    suspend: useMutation({ mutationFn: suspendAdminTenant, onSuccess: invalidate }),
    restore: useMutation({ mutationFn: restoreAdminTenant, onSuccess: invalidate }),
  }
}

export function useAdminUsers(params: Record<string, string | number | undefined> = {}) {
  return useQuery({ queryKey: adminKeys.users(params), queryFn: () => listAdminUsers(params), retry: false })
}

export function usePlatformAdmins(params: Record<string, string | number | undefined> = {}, enabled = true) {
  return useQuery({ queryKey: adminKeys.platformAdmins(params), queryFn: () => listPlatformAdmins(params), enabled, retry: false })
}

export function usePlatformAdminMutations(params: Record<string, string | number | undefined> = {}) {
  const queryClient = useQueryClient()
  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: adminKeys.users(params) })
    void queryClient.invalidateQueries({ queryKey: adminKeys.platformAdmins(params) })
    void queryClient.invalidateQueries({ queryKey: accessKeys.snapshot })
  }
  return {
    grant: useMutation({ mutationFn: ({ userId, role }: { userId: string; role: string }) => grantPlatformAdmin(userId, role), onSuccess: invalidate }),
    revoke: useMutation({ mutationFn: revokePlatformAdmin, onSuccess: invalidate }),
    updateUserStatus: useMutation({ mutationFn: ({ userId, status }: { userId: string; status: string }) => updateAdminUserStatus(userId, status), onSuccess: invalidate }),
  }
}

export function useAdminAuditLogs(params: Record<string, string | number | undefined> = {}) {
  return useQuery({ queryKey: adminKeys.audit(params), queryFn: () => listAdminAuditLogs(params), retry: false })
}
