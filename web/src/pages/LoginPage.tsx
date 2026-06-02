import { FormEvent, useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { LogIn } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { AuthShell } from '@/layouts/AuthShell'
import { accessKeys } from '@/features/access/access-hooks'
import { authKeys, useLoginMutation } from '@/features/auth/auth-hooks'
import { resendVerification } from '@/features/auth/auth-api'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { createLoginPayloadSchema } from '@/features/auth/auth-types'
import { useAuthStore } from '@/features/auth/auth-store'
import { Button, Card, CardHeader, Input, Toast } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'
import { errorMessage } from '@/shared/api/errors'

type LoginLocationState = {
  from?: {
    pathname?: string
    search?: string
  }
}

export function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const loginMutation = useLoginMutation()
  const { messages, schemaTranslator } = useAuthI18n()
  const copy = messages.login
  const lastLoginEmail = useAuthStore((state) => state.lastLoginEmail)
  const setLastLoginEmail = useAuthStore((state) => state.setLastLoginEmail)
  const [email, setEmail] = useState(lastLoginEmail ?? '')
  const [password, setPassword] = useState('')
  const [errors, setErrors] = useState<FieldErrors>({})
  const [emailNotVerified, setEmailNotVerified] = useState(false)
  const [resending, setResending] = useState(false)
  const [resendDone, setResendDone] = useState(false)
  const loginPayloadSchema = useMemo(() => createLoginPayloadSchema(schemaTranslator), [schemaTranslator])
  const from = useMemo(() => {
    const state = location.state as LoginLocationState | null
    const path = state?.from?.pathname && state.from.pathname !== '/login' ? state.from.pathname : '/'
    return `${path}${state?.from?.search ?? ''}`
  }, [location.state])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(loginPayloadSchema, { email: email.trim(), password })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    setLastLoginEmail(result.data.email)
    setEmailNotVerified(false)
    const loginResult = await loginMutation.mutateAsync({ email: result.data.email, password: result.data.password })
    if (!loginResult.user.email_verified) {
      setEmailNotVerified(true)
    }
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: authKeys.me }),
      queryClient.invalidateQueries({ queryKey: authKeys.tenants }),
      queryClient.invalidateQueries({ queryKey: authKeys.session }),
      queryClient.invalidateQueries({ queryKey: accessKeys.snapshot }),
    ])
    navigate(from, { replace: true })
  }

  const handleResend = async () => {
    if (!lastLoginEmail) return
    setResending(true)
    try {
      await resendVerification(lastLoginEmail)
      setResendDone(true)
    } catch {
      setResendDone(true) // API returns success regardless for security
    } finally {
      setResending(false)
    }
  }

  return (
    <AuthShell>
      <Card className="mx-auto max-w-md bg-white/95 hover:border-line hover:shadow-soft">
        <CardHeader title={copy.title} description={copy.description} />
        <form className="space-y-4" onSubmit={submit}>
          <Input
            label={copy.emailLabel}
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder={copy.emailPlaceholder}
            autoComplete="email"
            autoFocus
            error={errors.email}
          />
          <Input
            label={copy.passwordLabel}
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder={copy.passwordPlaceholder}
            autoComplete="current-password"
            error={errors.password}
          />
          <Button className="w-full" type="submit" isLoading={loginMutation.isPending} leftIcon={<LogIn className="size-4" />}>
            {copy.submit}
          </Button>
          <Toast tone="red" message={loginMutation.isError ? errorMessage(loginMutation.error) : undefined} />
          {emailNotVerified && (
            <div className="mt-3 rounded-panel border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800">
              <p className="font-bold">Your email is not yet verified.</p>
              <p className="mt-1">Some features may be limited. Please check your inbox for the verification email.</p>
              {!resendDone ? (
                <button
                  type="button"
                  className="mt-2 font-bold text-brand hover:text-brand-hover disabled:opacity-50"
                  onClick={handleResend}
                  disabled={resending}
                >
                  {resending ? 'Sending...' : 'Resend verification email'}
                </button>
              ) : (
                <p className="mt-2 text-green-700">Verification email sent!</p>
              )}
            </div>
          )}
        </form>
        <div className="mt-5 space-y-3 rounded-panel border border-line bg-surface-soft p-3 text-xs leading-6 text-muted">
          <div>
            <div className="font-bold text-ink">{copy.forgotPasswordTitle}</div>
            <Link className="font-bold text-brand hover:text-brand-hover" to="/forgot-password">
              {copy.forgotPasswordLink}
            </Link>
          </div>
          <div>
            <div className="font-bold text-ink">{copy.noAccountTitle}</div>
            <Link className="font-bold text-brand hover:text-brand-hover" to="/register" state={location.state}>
              {copy.registerLink}
            </Link>
          </div>
          <div>
            <div className="font-bold text-ink">{copy.localAccountTitle}</div>
            <code className="mt-2 block overflow-x-auto rounded-control bg-white p-2 text-[11px] text-brand">
              {copy.localAccountCommand}
            </code>
          </div>
        </div>
      </Card>
    </AuthShell>
  )
}
