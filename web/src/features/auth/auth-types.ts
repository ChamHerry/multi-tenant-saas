export type User = {
  id: string
  email: string
  display_name: string
  avatar_url?: string
  status: string
  last_login_at?: string
  metadata?: Record<string, unknown>
  created_at: string
  updated_at: string
}

export type AuthSession = {
  id: string
  user_id: string
  user_agent?: string
  ip?: string
  last_used_at: string
  expires_at: string
  idle_expires_at: string
  revoked_at?: string
  created_at: string
  updated_at: string
}

export type MeResponse = {
  user: User
}

export type AuthSessionResponse = {
  user: User
  session?: AuthSession
}

export type LoginPayload = {
  email: string
  password: string
}

export type RegisterPayload = {
  email: string
  password: string
  display_name?: string
}

export type ChangePasswordPayload = {
  old_password: string
  new_password: string
}

export type ActionResponse = {
  ok: boolean
}
