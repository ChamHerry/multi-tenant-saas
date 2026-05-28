export class ApiError extends Error {
  code?: number | string
  status?: number
  data?: unknown
  requestId?: string

  constructor(message: string, options: { code?: number | string; status?: number; data?: unknown; requestId?: string } = {}) {
    super(message)
    this.name = 'ApiError'
    this.code = options.code
    this.status = options.status
    this.data = options.data
    this.requestId = options.requestId
  }
}

export function errorMessage(error: unknown) {
  if (error instanceof ApiError) return error.message
  if (error instanceof Error) return error.message
  return String(error)
}
