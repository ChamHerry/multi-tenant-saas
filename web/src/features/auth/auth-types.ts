import { z } from 'zod'

export const userSchema = z.object({
  id: z.string(),
  email: z.string(),
  display_name: z.string(),
  avatar_url: z.string().optional(),
  status: z.string(),
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

export const loginPayloadSchema = z.object({
  email: z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
  password: z.string().min(1, '请输入密码'),
})
export type LoginPayload = z.infer<typeof loginPayloadSchema>

export const registerPayloadSchema = z
  .object({
    email: z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
    password: z.string().min(15, '密码至少 15 个字符'),
    confirmPassword: z.string().min(1, '请确认密码'),
    display_name: z.string().max(50, '显示名称最多 50 字符').optional(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: '两次输入的密码不一致',
    path: ['confirmPassword'],
  })
export type RegisterPayload = z.infer<typeof registerPayloadSchema>

export const changePasswordPayloadSchema = z
  .object({
    old_password: z.string().min(1, '请输入当前密码'),
    new_password: z.string().min(15, '新密码至少 15 个字符'),
    confirm_new_password: z.string().min(1, '请确认新密码'),
  })
  .refine((data) => data.new_password === data.confirm_new_password, {
    message: '两次输入的新密码不一致',
    path: ['confirm_new_password'],
  })
export type ChangePasswordPayload = z.infer<typeof changePasswordPayloadSchema>

export const actionResponseSchema = z.object({
  ok: z.boolean(),
})
export type ActionResponse = z.infer<typeof actionResponseSchema>
