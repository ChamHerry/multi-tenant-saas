import { FormEvent, useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { UserPlus } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { AuthShell } from '@/layouts/AuthShell'
import { accessKeys } from '@/features/access/access-hooks'
import { authKeys, useRegisterMutation } from '@/features/auth/auth-hooks'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { createRegisterPayloadSchema } from '@/features/auth/auth-types'
import { useAuthStore } from '@/features/auth/auth-store'
import { Button, Card, CardHeader, Input, Toast } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'
import { errorMessage } from '@/shared/api/errors'

type RegisterLocationState = {
  from?: {
    pathname?: string
    search?: string
  }
}

export function RegisterPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const registerMutation = useRegisterMutation()
  const { messages, schemaTranslator } = useAuthI18n()
  const copy = messages.register
  const setLastLoginEmail = useAuthStore((state) => state.setLastLoginEmail)
  const [email, setEmail] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [errors, setErrors] = useState<FieldErrors>({})
  const [verificationSent, setVerificationSent] = useState(false)
  const registerPayloadSchema = useMemo(() => createRegisterPayloadSchema(schemaTranslator), [schemaTranslator])
  const from = useMemo(() => {
    const state = location.state as RegisterLocationState | null
    const path = state?.from?.pathname && !['/login', '/register', '/'].includes(state.from.pathname) ? state.from.pathname : '/tenants'
    return `${path}${state?.from?.search ?? ''}`
  }, [location.state])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(registerPayloadSchema, {
      email: email.trim(),
      password,
      confirmPassword,
      display_name: displayName.trim() || undefined,
    })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    await registerMutation.mutateAsync({
      email: result.data.email,
      password: result.data.password,
      display_name: result.data.display_name,
    })
    setLastLoginEmail(result.data.email)
    setVerificationSent(true)
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: authKeys.me }),
      queryClient.invalidateQueries({ queryKey: authKeys.tenants }),
      queryClient.invalidateQueries({ queryKey: authKeys.session }),
      queryClient.invalidateQueries({ queryKey: accessKeys.snapshot }),
    ])
    navigate(from, { replace: true })
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
            label={copy.displayNameLabel}
            value={displayName}
            onChange={(event) => setDisplayName(event.target.value)}
            placeholder={copy.displayNamePlaceholder}
            autoComplete="name"
            error={errors.display_name}
          />
          <Input
            label={copy.passwordLabel}
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder={copy.passwordPlaceholder}
            autoComplete="new-password"
            error={errors.password}
          />
          <Input
            label={copy.confirmPasswordLabel}
            type="password"
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
            placeholder={copy.confirmPasswordPlaceholder}
            autoComplete="new-password"
            error={errors.confirmPassword}
          />
          <Button className="w-full" type="submit" isLoading={registerMutation.isPending} leftIcon={<UserPlus className="size-4" />}>
            {copy.submit}
          </Button>
          <Toast tone="red" message={registerMutation.isError ? errorMessage(registerMutation.error) : undefined} />
          {verificationSent && (
            <div className="mt-3 rounded-panel border border-green-200 bg-green-50 p-3 text-sm text-green-800">
              Registration successful! Please check your email and verify your address.
            </div>
          )}
        </form>
        <div className="mt-5 rounded-panel border border-line bg-surface-soft p-3 text-xs leading-6 text-muted">
          <div className="font-bold text-ink">{copy.hasAccountTitle}</div>
          <Link className="font-bold text-brand hover:text-brand-hover" to="/login" state={location.state}>
            {copy.loginLink}
          </Link>
          <p className="mt-2">{copy.registrationDisabledHint}</p>
        </div>
      </Card>
    </AuthShell>
  )
}
