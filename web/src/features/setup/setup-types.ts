import type { User } from '@/features/auth/auth-types'

export type SetupCheck = {
  ok: boolean
  message?: string
}

export type SetupState = {
  initialized: boolean
  requires_setup: boolean
  legacy_initialized: boolean
  initialized_at?: string
  initialized_by_user_id?: string
  version?: string
  missing: string[]
  checks: {
    database: SetupCheck
    migrations: SetupCheck
    encryption: SetupCheck
    redis: SetupCheck
  }
}

export type CompleteSetupPayload = {
  admin: {
    email: string
    password: string
    display_name?: string
  }
  runtime: {
    server_env: 'local' | 'test' | 'prod'
    web_base_url: string
    generate_session_secret: true
    generate_api_key_secret: true
  }
}

export type CompleteSetupResponse = {
  initialized: boolean
  user: User
}
