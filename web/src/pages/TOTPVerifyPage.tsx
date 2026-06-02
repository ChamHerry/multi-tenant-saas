import { FormEvent, useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { ShieldCheck } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { AuthShell } from '@/layouts/AuthShell'
import { accessKeys } from '@/features/access/access-hooks'
import { authKeys, useVerifyTOTPMutation } from '@/features/auth/auth-hooks'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { useAuthStore } from '@/features/auth/auth-store'
import { Button, Card, CardHeader, Input, Toast } from '@/shared/ui'
import { errorMessage } from '@/shared/api/errors'

type VerifyLocationState = {
  from?: string
  email?: string
}

export function TOTPVerifyPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const verifyMutation = useVerifyTOTPMutation()
  const { messages } = useAuthI18n()
  const copy = messages.totp.verify
  const backupCopy = messages.totp.backupCodes
  const pendingToken = useAuthStore((state) => state.pendingTOTPToken)
  const clearPendingTOTPToken = useAuthStore((state) => state.clearPendingTOTPToken)
  const [code, setCode] = useState('')
  const [useBackupCode, setUseBackupCode] = useState(false)
  const [successWarning, setSuccessWarning] = useState('')
  const state = location.state as VerifyLocationState | null
  const searchParams = useMemo(() => new URLSearchParams(location.search), [location.search])
  const challengeToken = searchParams.get('challenge_token') ?? undefined
  const queryEmail = searchParams.get('email') ?? undefined
  const from = useMemo(() => {
    const queryFrom = searchParams.get('from')
    if (queryFrom && queryFrom !== '/login') return queryFrom
    return state?.from && state.from !== '/login' ? state.from : '/'
  }, [searchParams, state?.from])
  const token = pendingToken

  const finishLogin = async (redirectOverride?: string) => {
    if (!challengeToken) {
      clearPendingTOTPToken()
    }
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: authKeys.me }),
      queryClient.invalidateQueries({ queryKey: authKeys.tenants }),
      queryClient.invalidateQueries({ queryKey: authKeys.session }),
      queryClient.invalidateQueries({ queryKey: accessKeys.snapshot }),
    ])
    navigate(redirectOverride || from, { replace: true })
  }

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    if ((!token && !challengeToken) || !code.trim()) return
    setSuccessWarning('')
    const result = await verifyMutation.mutateAsync(
      challengeToken
        ? { challenge_token: challengeToken, code: code.trim() }
        : { totp_token: token, code: code.trim() },
    )
    if (typeof result.backup_codes_remaining === 'number') {
      setSuccessWarning(backupCopy.remainingWarning.replace('{count}', String(result.backup_codes_remaining)))
      if (result.backup_codes_remaining <= 2) {
        return
      }
    }
    await finishLogin(result.redirect_uri)
  }

  return (
    <AuthShell>
      <Card className="mx-auto max-w-md bg-white/95 hover:border-line hover:shadow-soft">
        <CardHeader
          title={copy.title}
          description={(state?.email || queryEmail) ? `${copy.description} (${state?.email || queryEmail})` : copy.description}
        />
        {!token && !challengeToken ? (
          <div className="space-y-4">
            <Toast tone="red" message={copy.errors.tokenExpired} />
            <Link to="/login">
              <Button className="w-full" variant="secondary">{messages.login.submit}</Button>
            </Link>
          </div>
        ) : (
          <form className="space-y-4" onSubmit={submit}>
            <Input
              label={useBackupCode ? copy.backupCodeLabel : copy.codeLabel}
              value={code}
              onChange={(event) => setCode(event.target.value)}
              placeholder={useBackupCode ? copy.backupCodePlaceholder : copy.codePlaceholder}
              inputMode={useBackupCode ? 'text' : 'numeric'}
              autoComplete="one-time-code"
              autoFocus
            />
            <button
              type="button"
              className="text-sm font-bold text-brand hover:text-brand-hover"
              onClick={() => {
                setUseBackupCode((current) => !current)
                setCode('')
                setSuccessWarning('')
              }}
            >
              {useBackupCode ? copy.useAuthenticatorCode : copy.useBackupCode}
            </button>
            <Button className="w-full" type="submit" isLoading={verifyMutation.isPending} leftIcon={<ShieldCheck className="size-4" />}>
              {copy.submit}
            </Button>
            <Toast tone="red" message={verifyMutation.isError ? errorMessage(verifyMutation.error) : undefined} />
            <Toast tone="blue" message={successWarning} />
            {successWarning ? (
              <Button className="w-full" type="button" variant="secondary" onClick={() => finishLogin()}>
                {copy.continue}
              </Button>
            ) : null}
          </form>
        )}
      </Card>
    </AuthShell>
  )
}
