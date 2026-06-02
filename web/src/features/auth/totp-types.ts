import { z } from 'zod'

export const totpSetupResponseSchema = z.object({
  secret: z.string(),
  url: z.string(),
})
export type TOTPSetupResponse = z.infer<typeof totpSetupResponseSchema>

export const totpEnableResponseSchema = z.object({
  ok: z.boolean(),
  backup_codes: z.array(z.string()),
})
export type TOTPEnableResponse = z.infer<typeof totpEnableResponseSchema>

export const totpStatusResponseSchema = z.object({
  enabled: z.boolean(),
  setup_initiated: z.boolean(),
  enabled_at: z.string().nullable(),
  backup_codes_remaining: z.number(),
  created_at: z.string().nullable(),
})
export type TOTPStatusResponse = z.infer<typeof totpStatusResponseSchema>

export const verifyTOTPPayloadSchema = z.object({
  totp_token: z.string().min(1),
  code: z.string().min(1),
})
export type VerifyTOTPPayload = z.infer<typeof verifyTOTPPayloadSchema>

export type TOTPSetupBody = { password: string }
export type TOTPEnableBody = { code: string }
export type TOTPDisableBody = { password?: string; code?: string }
export type TOTPRegenerateBody = { password: string }
export type TOTPRegenerateResponse = { backup_codes: string[] }
