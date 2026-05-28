import { getCurrentTenantId } from '@/features/tenants/tenant-store'
import { readCookie } from '@/shared/lib/cookie'
import { ApiError } from './errors'
import { type ApiEnvelope, isApiEnvelope, isEnvelopeOk } from './envelope'
import { createRequestId } from './request-id'

type BodyInitLike = BodyInit | Record<string, unknown> | unknown[] | undefined

export type ApiRequestOptions = Omit<RequestInit, 'body'> & {
  body?: BodyInitLike
  tenantId?: string
  skipTenant?: boolean
  skipAuth?: boolean
}

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '')
const CSRF_COOKIE = import.meta.env.VITE_CSRF_COOKIE_NAME ?? 'repomind_csrf'

function buildUrl(path: string) {
  if (/^https?:\/\//i.test(path)) return path
  return `${API_BASE_URL}${path.startsWith('/') ? path : `/${path}`}`
}

function normalizeBody(body: BodyInitLike) {
  if (body === undefined || body === null) return undefined
  if (typeof body === 'string' || body instanceof FormData || body instanceof URLSearchParams || body instanceof Blob) {
    return body
  }
  return JSON.stringify(body)
}

function isUnsafeMethod(method?: string) {
  const normalized = (method ?? 'GET').toUpperCase()
  return !['GET', 'HEAD', 'OPTIONS', 'TRACE'].includes(normalized)
}

export async function apiRequest<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
  const requestId = createRequestId()
  const headers = new Headers(options.headers)
  const normalizedBody = normalizeBody(options.body)

  headers.set('X-Request-ID', requestId)
  if (normalizedBody !== undefined && !headers.has('Content-Type') && !(normalizedBody instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  if (!options.skipAuth && isUnsafeMethod(options.method)) {
    const csrfToken = readCookie(CSRF_COOKIE)
    if (csrfToken) headers.set('X-CSRF-Token', csrfToken)
  }

  const tenantId = options.tenantId ?? getCurrentTenantId()
  if (!options.skipTenant && tenantId) {
    headers.set('X-Tenant-ID', tenantId)
  }

  const response = await fetch(buildUrl(path), {
    ...options,
    credentials: options.credentials ?? 'include',
    body: normalizedBody,
    headers,
  })

  const contentType = response.headers.get('content-type') ?? ''
  const payload = contentType.includes('application/json') ? await response.json() : await response.text()

  if (isApiEnvelope(payload)) {
    const envelope = payload as ApiEnvelope<T>
    if (!response.ok || !isEnvelopeOk(envelope)) {
      throw new ApiError(envelope.message || response.statusText, {
        code: envelope.code,
        status: response.status,
        data: envelope.data,
        requestId,
      })
    }
    return envelope.data
  }

  if (!response.ok) {
    throw new ApiError(response.statusText || 'Request failed', {
      status: response.status,
      data: payload,
      requestId,
    })
  }

  return payload as T
}

export async function apiRaw<T>(path: string, options: ApiRequestOptions = {}) {
  return apiRequest<T>(path, { ...options, skipAuth: options.skipAuth ?? true, skipTenant: options.skipTenant ?? true })
}
