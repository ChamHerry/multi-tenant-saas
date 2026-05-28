import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { accessKeys } from '@/features/access/access-hooks'
import { addMember, listMembers, removeMember, updateMember } from './member-api'
import type { AddMemberInput, UpdateMemberInput } from './member-types'

export const memberKeys = {
  list: (tenantId?: string) => ['members', 'list', tenantId] as const,
}

export function useMembers(tenantId?: string) {
  return useQuery({
    queryKey: memberKeys.list(tenantId),
    queryFn: () => listMembers(tenantId!),
    enabled: Boolean(tenantId),
    retry: false,
  })
}

export function useAddMember(tenantId?: string) {
  const queryClient = useQueryClient()
  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: memberKeys.list(tenantId) })
    void queryClient.invalidateQueries({ queryKey: accessKeys.snapshot })
  }
  return useMutation({
    mutationFn: (input: AddMemberInput) => addMember(tenantId!, input),
    onSuccess: invalidate,
  })
}

export function useUpdateMember(tenantId?: string) {
  const queryClient = useQueryClient()
  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: memberKeys.list(tenantId) })
    void queryClient.invalidateQueries({ queryKey: accessKeys.snapshot })
  }
  return useMutation({
    mutationFn: ({ userId, input }: { userId: string; input: UpdateMemberInput }) => updateMember(tenantId!, userId, input),
    onSuccess: invalidate,
  })
}

export function useRemoveMember(tenantId?: string) {
  const queryClient = useQueryClient()
  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: memberKeys.list(tenantId) })
    void queryClient.invalidateQueries({ queryKey: accessKeys.snapshot })
  }
  return useMutation({
    mutationFn: (userId: string) => removeMember(tenantId!, userId),
    onSuccess: invalidate,
  })
}
