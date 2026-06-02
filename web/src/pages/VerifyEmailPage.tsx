import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Mail, CheckCircle, XCircle } from 'lucide-react'
import { AuthShell } from '@/layouts/AuthShell'
import { verifyEmail, resendVerification } from '@/features/auth/auth-api'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { Button, Card, CardHeader } from '@/shared/ui'
import { errorMessage } from '@/shared/api/errors'

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const { messages } = useAuthI18n()
  const copy = messages.verifyEmail
  const token = searchParams.get('token')
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [errorMsg, setErrorMsg] = useState('')
  const [resending, setResending] = useState(false)
  const [resendEmail, setResendEmail] = useState('')
  const [resendSent, setResendSent] = useState(false)
  const effectiveStatus = token ? status : 'error'
  const effectiveErrorMsg = token ? errorMsg : copy.missingToken

  useEffect(() => {
    if (!token) {
      return
    }
    verifyEmail(token)
      .then(() => setStatus('success'))
      .catch((e) => {
        setStatus('error')
        setErrorMsg(errorMessage(e))
      })
  }, [token])

  const handleResend = async () => {
    if (!resendEmail.trim()) return
    setResending(true)
    try {
      await resendVerification(resendEmail.trim())
      setResendSent(true)
    } catch {
      // The API returns success regardless for security
      setResendSent(true)
    } finally {
      setResending(false)
    }
  }

  return (
    <AuthShell>
      <Card className="mx-auto max-w-md bg-white/95 text-center hover:border-line hover:shadow-soft">
        {effectiveStatus === 'loading' && (
          <>
            <CardHeader title={copy.title} description={copy.verifying} />
            <div className="flex justify-center py-6">
              <div className="size-10 animate-spin rounded-full border-4 border-brand-soft border-t-brand" />
            </div>
          </>
        )}

        {effectiveStatus === 'success' && (
          <>
            <CardHeader title={copy.successTitle} description={copy.successDescription} />
            <div className="flex justify-center py-4 text-green-500">
              <CheckCircle className="size-16" />
            </div>
            <p className="mt-2 text-sm text-muted">{copy.redirecting}</p>
            <Link to="/login" className="mt-4 inline-block">
              <Button variant="primary">{copy.goToLogin}</Button>
            </Link>
            {/* Auto-redirect helper */}
            <RedirectAfter delay={3000} to="/login" />
          </>
        )}

        {effectiveStatus === 'error' && (
          <>
            <CardHeader title={copy.errorTitle} description={copy.errorDescription} />
            <div className="flex justify-center py-4 text-red-500">
              <XCircle className="size-16" />
            </div>
            <p className="mt-2 text-sm text-muted">{effectiveErrorMsg || copy.error}</p>

            <div className="mt-6 space-y-3">
              {resendSent ? (
                <p className="text-sm text-green-600">{copy.resendSent}</p>
              ) : (
                <div className="space-y-2">
                  <input
                    type="email"
                    className="w-full rounded-control border border-line bg-white px-3 py-2 text-sm"
                    placeholder={copy.resendEmailPlaceholder}
                    value={resendEmail}
                    onChange={(e) => setResendEmail(e.target.value)}
                  />
                  <Button
                    variant="secondary"
                    className="w-full"
                    onClick={handleResend}
                    isLoading={resending}
                    leftIcon={<Mail className="size-4" />}
                  >
                    {copy.resend}
                  </Button>
                </div>
              )}

              <Link to="/login" className="block">
                <Button variant="primary" className="w-full">
                  {copy.goToLogin}
                </Button>
              </Link>
            </div>
          </>
        )}
      </Card>
    </AuthShell>
  )
}

function RedirectAfter({ delay, to }: { delay: number; to: string }) {
  useEffect(() => {
    const timer = setTimeout(() => {
      window.location.href = to
    }, delay)
    return () => clearTimeout(timer)
  }, [delay, to])
  return null
}
