// web/src/features/dashboard/dashboard-api.ts
import { useQuery } from '@tanstack/react-query'
import { apiRequest } from '@/shared/api/client'
import type { DashboardSummary } from './dashboard-types'

export const dashboardKeys = {
  summary: ['dashboard', 'summary'] as const,
}

export function getDashboardSummary() {
  return apiRequest<DashboardSummary>('/api/v1/dashboard/summary')
}

export function useDashboardSummary() {
  return useQuery({
    queryKey: dashboardKeys.summary,
    queryFn: getDashboardSummary,
    staleTime: 30_000,
    retry: 1,
  })
}
