import { FormEvent, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Mail } from 'lucide-react'
import { AuthShell } from '@/layouts/AuthShell'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { useForgotPasswordMutation } from '@/features/auth/auth-hooks'
import { createForgotPasswordPayloadSchema } from '@/features/auth/auth-types'
import { Button, Card, CardHeader, Input, Toast } from '@/shared/ui'
import { validateForm, type FieldErrors } from '@/shared/lib/validate'
import { errorMessage } from '@/shared/api/errors'

export function ForgotPasswordPage() {
  const { messages, schemaTranslator } = useAuthI18n()
  const copy = messages.forgotPassword
  const mutation = useForgotPasswordMutation()
  const [email, setEmail] = useState('')
  const [errors, setErrors] = useState<FieldErrors>({})
  const [sent, setSent] = useState(false)
  const schema = useMemo(() => createForgotPasswordPayloadSchema(schemaTranslator), [schemaTranslator])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    const result = validateForm(schema, { email: email.trim() })
    if (result.errors) {
      setErrors(result.errors)
      return
    }
    setErrors({})
    await mutation.mutateAsync(result.data)
    setSent(true)
  }

  if (sent) {
    return (
      <AuthShell>
        <Card className="mx-auto max-w-md bg-white/95">
          <CardHeader title={copy.title} />
          <div className="space-y-4">
            <Toast tone="green" message={copy.successMessage} />
            <Link className="block text-center text-sm font-bold text-brand hover:text-brand-hover" to="/login">
              {copy.backToLogin}
            </Link>
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
            label={copy.emailLabel}
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder={copy.emailPlaceholder}
            autoComplete="email"
            autoFocus
            error={errors.email}
          />
          <Button className="w-full" type="submit" isLoading={mutation.isPending} leftIcon={<Mail className="size-4" />}>
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
