import { z } from 'zod'
import type { SchemaTranslator } from '@/shared/lib/schema-translator'
import { defaultAuthMessages, defaultAuthSchemaTranslator } from './auth-i18n'

export const userSchema = z.object({
  id: z.string(),
  email: z.string(),
  display_name: z.string(),
  avatar_url: z.string().optional(),
  status: z.string(),
  email_verified: z.boolean().optional(),
  last_login_at: z.string().optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  created_at: z.string(),
  updated_at: z.string(),
})
export type User = z.infer<typeof userSchema>

export const authSessionSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  user_agent: z.string().optional(),
  ip: z.string().optional(),
  last_used_at: z.string(),
  expires_at: z.string(),
  idle_expires_at: z.string(),
  revoked_at: z.string().optional(),
  created_at: z.string(),
  updated_at: z.string(),
})
export type AuthSession = z.infer<typeof authSessionSchema>

export const meResponseSchema = z.object({
  user: userSchema,
})
export type MeResponse = z.infer<typeof meResponseSchema>

export const authSessionResponseSchema = z.object({
  user: userSchema,
  session: authSessionSchema.optional(),
})
export type AuthSessionResponse = z.infer<typeof authSessionResponseSchema>

const validationFallback = defaultAuthMessages.validation

export function createLoginPayloadSchema(t: SchemaTranslator = defaultAuthSchemaTranslator) {
  return z.object({
    email: z.string().min(1, t('emailRequired', validationFallback.emailRequired)).email(t('emailInvalid', validationFallback.emailInvalid)),
    password: z.string().min(1, t('passwordRequired', validationFallback.passwordRequired)),
  })
}

export const loginPayloadSchema = createLoginPayloadSchema()
export type LoginPayload = z.infer<ReturnType<typeof createLoginPayloadSchema>>

export function createRegisterPayloadSchema(t: SchemaTranslator = defaultAuthSchemaTranslator) {
  return z
    .object({
      email: z.string().min(1, t('emailRequired', validationFallback.emailRequired)).email(t('emailInvalid', validationFallback.emailInvalid)),
      password: z.string().min(15, t('passwordMin', validationFallback.passwordMin, { min: 15 })),
      confirmPassword: z.string().min(1, t('confirmPasswordRequired', validationFallback.confirmPasswordRequired)),
      display_name: z.string().max(50, t('displayNameMax', validationFallback.displayNameMax, { max: 50 })).optional(),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: t('passwordsMustMatch', validationFallback.passwordsMustMatch),
      path: ['confirmPassword'],
    })
}
export const registerPayloadSchema = createRegisterPayloadSchema()
export type RegisterPayload = z.infer<ReturnType<typeof createRegisterPayloadSchema>>

export function createChangePasswordPayloadSchema(t: SchemaTranslator = defaultAuthSchemaTranslator) {
  return z
    .object({
      old_password: z.string().min(1, t('currentPasswordRequired', validationFallback.currentPasswordRequired)),
      new_password: z.string().min(15, t('newPasswordMin', validationFallback.newPasswordMin, { min: 15 })),
      confirm_new_password: z.string().min(1, t('confirmNewPasswordRequired', validationFallback.confirmNewPasswordRequired)),
    })
    .refine((data) => data.new_password === data.confirm_new_password, {
      message: t('newPasswordsMustMatch', validationFallback.newPasswordsMustMatch),
      path: ['confirm_new_password'],
    })
}
export const changePasswordPayloadSchema = createChangePasswordPayloadSchema()
export type ChangePasswordPayload = z.infer<ReturnType<typeof createChangePasswordPayloadSchema>>

export const actionResponseSchema = z.object({
  ok: z.boolean(),
})
export type ActionResponse = z.infer<typeof actionResponseSchema>

export function createForgotPasswordPayloadSchema(t: SchemaTranslator = defaultAuthSchemaTranslator) {
  return z.object({
    email: z.string()
      .min(1, t('emailRequired', validationFallback.emailRequired))
      .email(t('emailInvalid', validationFallback.emailInvalid)),
  })
}
export const forgotPasswordPayloadSchema = createForgotPasswordPayloadSchema()
export type ForgotPasswordPayload = z.infer<ReturnType<typeof createForgotPasswordPayloadSchema>>

export function createResetPasswordPayloadSchema(t: SchemaTranslator = defaultAuthSchemaTranslator) {
  return z
    .object({
      token: z.string().min(64, t('tokenRequired', validationFallback.tokenRequired)),
      new_password: z.string().min(15, t('newPasswordMin', validationFallback.newPasswordMin, { min: 15 })),
      confirm_new_password: z.string().min(1, t('confirmNewPasswordRequired', validationFallback.confirmNewPasswordRequired)),
    })
    .refine((data) => data.new_password === data.confirm_new_password, {
      message: t('newPasswordsMustMatch', validationFallback.newPasswordsMustMatch),
      path: ['confirm_new_password'],
    })
}
export const resetPasswordPayloadSchema = createResetPasswordPayloadSchema()
export type ResetPasswordPayload = z.infer<ReturnType<typeof createResetPasswordPayloadSchema>>
