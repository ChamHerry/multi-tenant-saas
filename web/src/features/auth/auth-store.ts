import { create } from 'zustand'

export type AuthState = {
  lastLoginEmail?: string
  setLastLoginEmail: (email: string) => void
  clearAuthState: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  lastLoginEmail: undefined,
  setLastLoginEmail: (email) => set({ lastLoginEmail: email.trim() || undefined }),
  clearAuthState: () => set({ lastLoginEmail: undefined }),
}))
