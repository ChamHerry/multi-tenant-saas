import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { completeSetup, getSetupState } from './setup-api'
import type { CompleteSetupPayload } from './setup-types'

export const setupKeys = {
  state: ['setup', 'state'] as const,
}

export function useSetupState() {
  return useQuery({ queryKey: setupKeys.state, queryFn: getSetupState, retry: false, staleTime: 10_000 })
}

export function useCompleteSetupMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: CompleteSetupPayload) => completeSetup(payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: setupKeys.state })
    },
  })
}
