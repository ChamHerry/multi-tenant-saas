import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  changePassword,
  forgotPassword,
  getAuthSession,
  getMe,
  getMyTenants,
  getOAuthProviders,
  listSessions,
  login,
  logout,
  register,
  resetPassword,
  revokeOtherSessions,
  revokeSession,
  unlockUser,
  verifyTOTP,
} from './auth-api'

export const authKeys = {
  me: ['auth', 'me'] as const,
  session: ['auth', 'session'] as const,
  sessions: ['auth', 'sessions'] as const,
  tenants: ['auth', 'tenants'] as const,
  totpStatus: ['auth', 'totp-status'] as const,
  oauthProviders: ['auth', 'oauth-providers'] as const,
}

export function useMe() {
  return useQuery({
    queryKey: authKeys.me,
    queryFn: getMe,
    retry: false,
  })
}

export function useAuthSession() {
  return useQuery({
    queryKey: authKeys.session,
    queryFn: getAuthSession,
    retry: false,
  })
}

export function useListSessions() {
  return useQuery({
    queryKey: authKeys.sessions,
    queryFn: listSessions,
    retry: false,
  })
}

export function useMyTenants() {
  return useQuery({
    queryKey: authKeys.tenants,
    queryFn: getMyTenants,
    retry: false,
  })
}

export function useOAuthProviders() {
  return useQuery({
    queryKey: authKeys.oauthProviders,
    queryFn: getOAuthProviders,
    retry: false,
  })
}

export function useLoginMutation() {
  return useMutation({ mutationFn: login })
}

export function useVerifyTOTPMutation() {
  return useMutation({ mutationFn: verifyTOTP })
}

export function useRegisterMutation() {
  return useMutation({ mutationFn: register })
}

export function useLogoutMutation() {
  return useMutation({ mutationFn: logout })
}

export function useRevokeSessionMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: revokeSession,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: authKeys.sessions })
      void queryClient.invalidateQueries({ queryKey: authKeys.session })
    },
  })
}

export function useRevokeOtherSessionsMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: revokeOtherSessions,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: authKeys.sessions })
      void queryClient.invalidateQueries({ queryKey: authKeys.session })
    },
  })
}

export function useChangePasswordMutation() {
  return useMutation({ mutationFn: changePassword })
}

export function useForgotPasswordMutation() {
  return useMutation({ mutationFn: forgotPassword })
}

export function useResetPasswordMutation() {
  return useMutation({ mutationFn: resetPassword })
}

export function useUnlockUserMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: unlockUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
    },
  })
}
