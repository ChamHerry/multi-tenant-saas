import { create } from 'zustand'

export type AuthState = {
  lastLoginEmail?: string
  pendingTOTPToken?: string
  setLastLoginEmail: (email: string) => void
  setPendingTOTPToken: (token: string) => void
  clearPendingTOTPToken: () => void
  clearAuthState: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  lastLoginEmail: undefined,
  pendingTOTPToken: undefined,
  setLastLoginEmail: (email) => set({ lastLoginEmail: email.trim() || undefined }),
  setPendingTOTPToken: (token) => set({ pendingTOTPToken: token.trim() || undefined }),
  clearPendingTOTPToken: () => set({ pendingTOTPToken: undefined }),
  clearAuthState: () => set({ lastLoginEmail: undefined, pendingTOTPToken: undefined }),
}))
