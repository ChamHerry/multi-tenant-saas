import { FormEvent, useEffect, useState } from 'react'
import QRCode from 'qrcode'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ShieldCheck } from 'lucide-react'
import { authKeys } from '@/features/auth/auth-hooks'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { enableTOTP, setupTOTP } from '@/features/auth/totp-api'
import type { TOTPSetupResponse } from '@/features/auth/totp-types'
import { Button, Card, CardHeader, Input, Toast } from '@/shared/ui'
import { errorMessage } from '@/shared/api/errors'

type SetupStep = 'password' | 'scan' | 'verify' | 'backup'

export function TOTPSetupPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { messages } = useAuthI18n()
  const copy = messages.totp.setup
  const backupCopy = messages.totp.backupCodes
  const [step, setStep] = useState<SetupStep>('password')
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')
  const [setup, setSetup] = useState<TOTPSetupResponse>()
  const [qrDataURL, setQrDataURL] = useState('')
  const [backupCodes, setBackupCodes] = useState<string[]>([])
  const [saved, setSaved] = useState(false)
  const setupMutation = useMutation({ mutationFn: setupTOTP })
  const enableMutation = useMutation({ mutationFn: enableTOTP })

  useEffect(() => {
    let cancelled = false
    if (!setup?.url) {
      return
    }
    QRCode.toDataURL(setup.url, { margin: 1, width: 224 })
      .then((dataURL) => {
        if (!cancelled) setQrDataURL(dataURL)
      })
      .catch(() => {
        if (!cancelled) setQrDataURL('')
      })
    return () => {
      cancelled = true
    }
  }, [setup?.url])

  const startSetup = async (event: FormEvent) => {
    event.preventDefault()
    setQrDataURL('')
    const result = await setupMutation.mutateAsync({ password })
    setSetup(result)
    setStep('scan')
  }

  const verifySetup = async (event: FormEvent) => {
    event.preventDefault()
    const result = await enableMutation.mutateAsync({ code: code.trim() })
    setBackupCodes(result.backup_codes)
    setStep('backup')
    await queryClient.invalidateQueries({ queryKey: authKeys.totpStatus })
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <Card>
        <CardHeader title={copy.title} description={copy.description} />
        {step === 'password' ? (
          <form className="space-y-4" onSubmit={startSetup}>
            <Input
              label={copy.passwordLabel}
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              autoFocus
            />
            <Button type="submit" isLoading={setupMutation.isPending} leftIcon={<ShieldCheck className="size-4" />}>
              {copy.start}
            </Button>
            <Toast tone="red" message={setupMutation.isError ? errorMessage(setupMutation.error) : undefined} />
          </form>
        ) : null}

        {step === 'scan' && setup ? (
          <div className="space-y-4">
            <div>
              <h3 className="font-bold text-ink">{copy.step1Title}</h3>
              <p className="mt-1 text-sm leading-6 text-muted">{copy.step1Description}</p>
            </div>
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start">
              <div className="grid size-60 place-items-center rounded-panel border border-line bg-white p-3">
                {qrDataURL ? <img src={qrDataURL} alt={copy.qrAlt} className="size-56" /> : <span className="text-sm text-muted">{copy.qrLoading}</span>}
              </div>
              <div className="min-w-0 flex-1 rounded-panel border border-line bg-surface-soft p-4">
                <div className="text-xs font-bold uppercase tracking-wide text-muted">{copy.manualEntry}</div>
                <code className="mt-2 block break-all rounded-control bg-white p-3 text-sm font-bold text-brand">{setup.secret}</code>
              </div>
            </div>
            <Button type="button" onClick={() => setStep('verify')}>{copy.next}</Button>
          </div>
        ) : null}

        {step === 'verify' ? (
          <form className="space-y-4" onSubmit={verifySetup}>
            <div>
              <h3 className="font-bold text-ink">{copy.step2Title}</h3>
              <p className="mt-1 text-sm leading-6 text-muted">{copy.step2Description}</p>
            </div>
            <Input
              label={copy.codeLabel}
              value={code}
              onChange={(event) => setCode(event.target.value)}
              placeholder={copy.codePlaceholder}
              inputMode="numeric"
              autoComplete="one-time-code"
              autoFocus
            />
            <Button type="submit" isLoading={enableMutation.isPending}>{copy.verify}</Button>
            <Toast tone="red" message={enableMutation.isError ? errorMessage(enableMutation.error) : undefined} />
          </form>
        ) : null}

        {step === 'backup' ? (
          <div className="space-y-4">
            <div>
              <h3 className="font-bold text-ink">{backupCopy.title}</h3>
              <p className="mt-1 text-sm leading-6 text-muted">{backupCopy.description}</p>
            </div>
            <div className="grid gap-2 sm:grid-cols-2">
              {backupCodes.map((item) => (
                <code key={item} className="rounded-control border border-line bg-surface-soft px-3 py-2 text-center font-bold tracking-widest text-ink">
                  {item}
                </code>
              ))}
            </div>
            <label className="flex items-center gap-2 text-sm text-muted">
              <input type="checkbox" checked={saved} onChange={(event) => setSaved(event.target.checked)} />
              {backupCopy.savedConfirmation}
            </label>
            <Button type="button" disabled={!saved} onClick={() => navigate('/me/security')}>
              {backupCopy.complete}
            </Button>
          </div>
        ) : null}
      </Card>
    </div>
  )
}
