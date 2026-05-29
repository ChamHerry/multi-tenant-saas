import type { ZodSchema } from 'zod'

export type FieldErrors = Record<string, string>

export function validateForm<T>(schema: ZodSchema<T>, data: unknown):
  | { data: T; errors: undefined }
  | { data: undefined; errors: FieldErrors } {
  const result = schema.safeParse(data)
  if (result.success) return { data: result.data, errors: undefined }
  const errors: FieldErrors = {}
  for (const issue of result.error.issues) {
    const key = issue.path.join('.') || '_root'
    if (!errors[key]) errors[key] = issue.message
  }
  return { data: undefined, errors }
}
