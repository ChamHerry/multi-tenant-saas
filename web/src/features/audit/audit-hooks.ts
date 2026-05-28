import { useMutation, useQuery } from '@tanstack/react-query'
import { exportTenantAuditLogs, listPlatformAuditLogs, listSecurityEvents, listTenantAuditLogs } from './audit-api'

export const auditKeys = {
  tenant: (tenantId?: string, params: Record<string, string | number | undefined> = {}) => ['audit', 'tenant', tenantId, params] as const,
  security: (params: Record<string, string | number | undefined> = {}) => ['audit', 'security', params] as const,
  platform: (params: Record<string, string | number | undefined> = {}) => ['audit', 'platform', params] as const,
}

export function useTenantAuditLogs(tenantId?: string, params: Record<string, string | number | undefined> = {}) {
  return useQuery({ queryKey: auditKeys.tenant(tenantId, params), queryFn: () => listTenantAuditLogs(tenantId!, params), enabled: Boolean(tenantId), retry: false })
}

export function useExportTenantAuditLogs(tenantId?: string) {
  return useMutation({ mutationFn: (params: Record<string, string | number | undefined> = {}) => exportTenantAuditLogs(tenantId!, params) })
}

export function useSecurityEvents(params: Record<string, string | number | undefined> = {}) {
  return useQuery({ queryKey: auditKeys.security(params), queryFn: () => listSecurityEvents(params), retry: false })
}

export function usePlatformAuditLogs(params: Record<string, string | number | undefined> = {}) {
  return useQuery({ queryKey: auditKeys.platform(params), queryFn: () => listPlatformAuditLogs(params), retry: false })
}
