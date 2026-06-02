import { FormEvent, useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { KeyRound } from 'lucide-react'
import { AuthShell } from '@/layouts/AuthShell'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { useResetPasswordMutation } from '@/features/auth/auth-hooks'
import { createResetPasswordPayloadSchema } from '@/features/auth/auth-types'
import { Button, Card, CardHeader, Input, Toast } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'
import { errorMessage } from '@/shared/api/errors'

export function ResetPasswordPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const tokenFromURL = searchParams.get('token') ?? ''
  const { messages, schemaTranslator } = useAuthI18n()
  const copy = messages.resetPassword
  const mutation = useResetPasswordMutation()
  const [newPassword, setNewPassword] = useState('')
  const [confirmNewPassword, setConfirmNewPassword] = useState('')
  const [errors, setErrors] = useState<FieldErrors>({})
  const [success, setSuccess] = useState(false)
  const schema = useMemo(() => createResetPasswordPayloadSchema(schemaTranslator), [schemaTranslator])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(schema, {
      token: tokenFromURL,
      new_password: newPassword,
      confirm_new_password: confirmNewPassword,
    })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    try {
      await mutation.mutateAsync({ token: tokenFromURL, new_password: result.data.new_password })
      setSuccess(true)
    } catch {
      // Error handled by Toast
    }
  }

  if (!tokenFromURL) {
    return (
      <AuthShell>
        <Card className="mx-auto max-w-md bg-white/95">
          <CardHeader title={copy.title} />
          <Toast tone="red" message={copy.invalidToken} />
          <Link className="block text-center text-sm font-bold text-brand hover:text-brand-hover" to="/forgot-password">
            {copy.backToLogin}
          </Link>
        </Card>
      </AuthShell>
    )
  }

  if (success) {
    return (
      <AuthShell>
        <Card className="mx-auto max-w-md bg-white/95">
          <CardHeader title={copy.title} />
          <div className="space-y-4">
            <Toast tone="green" message={copy.successMessage} />
            <Button className="w-full" onClick={() => navigate('/login')}>
              {copy.backToLogin}
            </Button>
          </div>
        </Card>
      </AuthShell>
    )
  }

  return (
    <AuthShell>
      <Card className="mx-auto max-w-md bg-white/95 hover:border-line hover:shadow-soft">
        <CardHeader title={copy.title} description={copy.description} />
        <form className="space-y-4" onSubmit={submit}>
          <Input
            label={copy.newPasswordLabel}
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            placeholder={copy.newPasswordPlaceholder}
            autoComplete="new-password"
            autoFocus
            error={errors.new_password}
          />
          <Input
            label={copy.confirmNewPasswordLabel}
            type="password"
            value={confirmNewPassword}
            onChange={(e) => setConfirmNewPassword(e.target.value)}
            placeholder={copy.confirmNewPasswordPlaceholder}
            autoComplete="new-password"
            error={errors.confirm_new_password}
          />
          <Button className="w-full" type="submit" isLoading={mutation.isPending} leftIcon={<KeyRound className="size-4" />}>
            {copy.submit}
          </Button>
          <Toast tone="red" message={mutation.isError ? errorMessage(mutation.error) : undefined} />
        </form>
        <div className="mt-5 text-center text-sm">
          <Link className="font-bold text-brand hover:text-brand-hover" to="/login">
            {copy.backToLogin}
          </Link>
        </div>
      </Card>
    </AuthShell>
  )
}
