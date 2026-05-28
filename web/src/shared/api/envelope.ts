export type ApiEnvelope<T> = {
  code: number | string
  message: string
  data: T
}

export function isApiEnvelope(value: unknown): value is ApiEnvelope<unknown> {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Record<string, unknown>
  return 'code' in candidate && 'message' in candidate && 'data' in candidate
}

export function isEnvelopeOk(envelope: ApiEnvelope<unknown>) {
  return envelope.code === 0 || envelope.code === 'OK'
}
