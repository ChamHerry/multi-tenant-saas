import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { accessKeys } from '@/features/access/access-hooks'
import { authKeys } from '@/features/auth/auth-hooks'
import { acceptInvitation, createTenantInvitation, declineInvitation, listMyInvitations, listTenantInvitations, resendTenantInvitation, revokeTenantInvitation } from './invitation-api'
import type { CreateInvitationInput } from './invitation-types'

export const invitationKeys = {
  tenant: (tenantId?: string, status = '') => ['invitations', 'tenant', tenantId, status] as const,
  mine: (status = 'pending') => ['invitations', 'mine', status] as const,
}

export function useTenantInvitations(tenantId?: string, status = '') {
  return useQuery({
    queryKey: invitationKeys.tenant(tenantId, status),
    queryFn: () => listTenantInvitations(tenantId!, status),
    enabled: Boolean(tenantId),
    retry: false,
  })
}

export function useCreateInvitation(tenantId?: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: CreateInvitationInput) => createTenantInvitation(tenantId!, input),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: invitationKeys.tenant(tenantId) }),
  })
}

export function useResendInvitation(tenantId?: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (invitationId: string) => resendTenantInvitation(tenantId!, invitationId),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: invitationKeys.tenant(tenantId) }),
  })
}

export function useRevokeInvitation(tenantId?: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (invitationId: string) => revokeTenantInvitation(tenantId!, invitationId),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: invitationKeys.tenant(tenantId) }),
  })
}

export function useMyInvitations(status = 'pending') {
  return useQuery({ queryKey: invitationKeys.mine(status), queryFn: () => listMyInvitations(status), retry: false })
}

export function useAcceptInvitation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: acceptInvitation,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: invitationKeys.mine('pending') })
      void queryClient.invalidateQueries({ queryKey: authKeys.tenants })
      void queryClient.invalidateQueries({ queryKey: accessKeys.snapshot })
    },
  })
}

export function useDeclineInvitation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: declineInvitation,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: invitationKeys.mine('pending') }),
  })
}
