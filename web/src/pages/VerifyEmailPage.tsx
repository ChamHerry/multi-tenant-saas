import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { Mail, CheckCircle, XCircle } from 'lucide-react'
import { AuthShell } from '@/layouts/AuthShell'
import { verifyEmail, resendVerification } from '@/features/auth/auth-api'
import { Button, Card, CardHeader } from '@/shared/ui'
import { errorMessage } from '@/shared/api/errors'

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token')
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading')
  const [errorMsg, setErrorMsg] = useState('')
  const [resending, setResending] = useState(false)
  const [resendEmail, setResendEmail] = useState('')
  const [resendSent, setResendSent] = useState(false)

  useEffect(() => {
    if (!token) {
      setStatus('error')
      setErrorMsg('Missing verification token.')
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
    } catch (e) {
      // The API returns success regardless for security
      setResendSent(true)
    } finally {
      setResending(false)
    }
  }

  return (
    <AuthShell>
      <Card className="mx-auto max-w-md bg-white/95 text-center hover:border-line hover:shadow-soft">
        {status === 'loading' && (
          <>
            <CardHeader title="Email Verification" description="Verifying your email address..." />
            <div className="flex justify-center py-6">
              <div className="size-10 animate-spin rounded-full border-4 border-brand-soft border-t-brand" />
            </div>
          </>
        )}

        {status === 'success' && (
          <>
            <CardHeader title="Email Verified" description="Your email has been verified successfully." />
            <div className="flex justify-center py-4 text-green-500">
              <CheckCircle className="size-16" />
            </div>
            <p className="mt-2 text-sm text-muted">Redirecting to login in 3 seconds...</p>
            <Link to="/login" className="mt-4 inline-block">
              <Button variant="primary">Go to login</Button>
            </Link>
            {/* Auto-redirect helper */}
            <RedirectAfter delay={3000} to="/login" />
          </>
        )}

        {status === 'error' && (
          <>
            <CardHeader title="Verification Failed" description="We couldn't verify your email address." />
            <div className="flex justify-center py-4 text-red-500">
              <XCircle className="size-16" />
            </div>
            <p className="mt-2 text-sm text-muted">{errorMsg || 'The link may be expired or already used.'}</p>

            <div className="mt-6 space-y-3">
              {resendSent ? (
                <p className="text-sm text-green-600">Verification email sent! Please check your inbox.</p>
              ) : (
                <div className="space-y-2">
                  <input
                    type="email"
                    className="w-full rounded-control border border-line bg-white px-3 py-2 text-sm"
                    placeholder="Enter your email to resend"
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
                    Resend verification email
                  </Button>
                </div>
              )}

              <Link to="/login" className="block">
                <Button variant="primary" className="w-full">
                  Go to login
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
