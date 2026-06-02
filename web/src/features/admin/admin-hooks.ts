import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { accessKeys } from '@/features/access/access-hooks'
import {
  getPlatformSession,
  grantPlatformAdmin,
  deleteAdminSystemConfig,
  getAdminSystemConfig,
  listAdminAuditLogs,
  listAdminSystemConfigs,
  listAdminTenants,
  listAdminUsers,
  listPlatformAdmins,
  restoreAdminTenant,
  revokePlatformAdmin,
  suspendAdminTenant,
  updateAdminUserStatus,
  upsertAdminSystemConfig,
} from './admin-api'
import type { UpsertSystemConfigPayload } from './admin-types'

export const adminKeys = {
  session: ['admin', 'session'] as const,
  tenants: (params: Record<string, string | number | undefined> = {}) => ['admin', 'tenants', params] as const,
  users: (params: Record<string, string | number | undefined> = {}) => ['admin', 'users', params] as const,
  platformAdmins: (params: Record<string, string | number | undefined> = {}) => ['admin', 'platform-admins', params] as const,
  audit: (params: Record<string, string | number | undefined> = {}) => ['admin', 'audit', params] as const,
  systemConfigs: (params: Record<string, string | number | undefined> = {}) => ['admin', 'system-config', params] as const,
  systemConfig: (key: string) => ['admin', 'system-config', key] as const,
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

export function useAdminSystemConfigs(params: Record<string, string | number | undefined> = {}) {
  return useQuery({ queryKey: adminKeys.systemConfigs(params), queryFn: () => listAdminSystemConfigs(params), retry: false })
}

export function useAdminSystemConfig(key: string, enabled = true) {
  return useQuery({ queryKey: adminKeys.systemConfig(key), queryFn: () => getAdminSystemConfig(key), enabled: enabled && key.length > 0, retry: false })
}

export function useAdminSystemConfigMutations(params: Record<string, string | number | undefined> = {}) {
  const queryClient = useQueryClient()
  const invalidate = (key?: string) => {
    void queryClient.invalidateQueries({ queryKey: adminKeys.systemConfigs(params) })
    void queryClient.invalidateQueries({ queryKey: ['admin', 'system-config'] })
    void queryClient.invalidateQueries({ queryKey: adminKeys.audit({ action: 'system_config.updated' }) })
    void queryClient.invalidateQueries({ queryKey: adminKeys.audit({ action: 'system_config.deleted' }) })
    if (key) void queryClient.invalidateQueries({ queryKey: adminKeys.systemConfig(key) })
  }
  return {
    upsert: useMutation({
      mutationFn: ({ key, payload }: { key: string; payload: UpsertSystemConfigPayload }) => upsertAdminSystemConfig(key, payload),
      onSuccess: (_, variables) => invalidate(variables.key),
    }),
    deleteConfig: useMutation({ mutationFn: deleteAdminSystemConfig, onSuccess: (_, key) => invalidate(key) }),
  }
}
