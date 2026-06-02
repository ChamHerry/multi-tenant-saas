import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { changePassword, forgotPassword, getAuthSession, getMe, getMyTenants, login, logout, register, resetPassword, unlockUser, verifyTOTP } from './auth-api'

export const authKeys = {
  me: ['auth', 'me'] as const,
  session: ['auth', 'session'] as const,
  tenants: ['auth', 'tenants'] as const,
  totpStatus: ['auth', 'totp-status'] as const,
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

export function useMyTenants() {
  return useQuery({
    queryKey: authKeys.tenants,
    queryFn: getMyTenants,
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
