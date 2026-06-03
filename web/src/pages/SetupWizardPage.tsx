import { FormEvent, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { CheckCircle2, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { AuthShell } from '@/layouts/AuthShell'
import { useCompleteSetupMutation, useSetupState } from '@/features/setup/setup-hooks'
import type { SetupCheck } from '@/features/setup/setup-types'
import { errorMessage } from '@/shared/api/errors'
import { Button, Card, CardHeader, Input, Toast, LoadingView, ErrorView } from '@/shared/ui'

type SetupForm = {
  email: string
  password: string
  displayName: string
  webBaseUrl: string
}

const stepKeys = ['system', 'admin', 'security', 'review'] as const

export function SetupWizardPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const setup = useSetupState()
  const completeSetup = useCompleteSetupMutation()
  const [step, setStep] = useState(0)
  const [form, setForm] = useState<SetupForm>({
    email: '',
    password: '',
    displayName: '',
    webBaseUrl: window.location.origin,
  })
  const [clientError, setClientError] = useState('')

  const checks = setup.data?.checks
  const canGoNext = useMemo(() => {
    if (step === 0) return Boolean(checks?.database.ok && checks.migrations.ok && checks.encryption.ok)
    if (step === 1) return form.email.trim().length > 0 && form.password.length >= 15
    if (step === 2) return isAbsoluteHTTPURL(form.webBaseUrl)
    return true
  }, [checks, form.email, form.password.length, form.webBaseUrl, step])

  const next = () => {
    setClientError('')
    if (!canGoNext) {
      setClientError(t('setup.errors.completeCurrentStep'))
      return
    }
    setStep((current) => Math.min(current + 1, stepKeys.length - 1))
  }

  const back = () => {
    setClientError('')
    setStep((current) => Math.max(current - 1, 0))
  }

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setClientError('')
    if (!isAbsoluteHTTPURL(form.webBaseUrl)) {
      setClientError(t('setup.errors.invalidBaseUrl'))
      return
    }
    if (form.password.length < 15) {
      setClientError(t('setup.errors.passwordLength'))
      return
    }
    try {
      await completeSetup.mutateAsync({
        admin: {
          email: form.email.trim(),
          password: form.password,
          display_name: form.displayName.trim() || undefined,
        },
        runtime: {
          web_base_url: form.webBaseUrl.trim().replace(/\/$/, ''),
          generate_session_secret: true,
          generate_api_key_secret: true,
        },
      })
      navigate('/login', { replace: true })
    } catch {
      // Rendered by Toast below.
    }
  }

  return (
    <AuthShell>
      <Card className="mx-auto max-w-2xl bg-white/95 hover:border-line hover:shadow-soft">
        <CardHeader title={t('setup.title')} description={t('setup.description')} action={<ShieldCheck className="size-6 text-brand" />} />
        {setup.isLoading ? <LoadingView label={t('setup.loading')} /> : null}
        {setup.isError ? <ErrorView title={t('setup.failed')} error={setup.error} /> : null}
        {setup.data ? (
          <form className="space-y-5" onSubmit={submit}>
            <ol className="grid gap-2 sm:grid-cols-4">
              {stepKeys.map((key, index) => (
                <li
                  key={key}
                  className={index === step ? 'rounded-panel border border-brand bg-brand-soft p-3 text-sm font-bold text-brand' : 'rounded-panel border border-line bg-surface-soft p-3 text-sm font-bold text-muted'}
                >
                  {t(`setup.steps.${key}`)}
                </li>
              ))}
            </ol>

            {step === 0 ? <SystemStep checks={setup.data.checks} missing={setup.data.missing} /> : null}
            {step === 1 ? <AdminStep form={form} setForm={setForm} /> : null}
            {step === 2 ? <SecurityStep form={form} setForm={setForm} /> : null}
            {step === 3 ? <ReviewStep form={form} /> : null}

            <Toast tone="red" message={clientError || (completeSetup.isError ? errorMessage(completeSetup.error) : undefined)} />
            <div className="flex flex-wrap justify-between gap-3">
              <Button type="button" variant="secondary" onClick={back} disabled={step === 0 || completeSetup.isPending}>
                {t('setup.back')}
              </Button>
              {step < stepKeys.length - 1 ? (
                <Button type="button" onClick={next} disabled={!canGoNext}>
                  {t('setup.next')}
                </Button>
              ) : (
                <Button type="submit" isLoading={completeSetup.isPending} leftIcon={<CheckCircle2 className="size-4" />}>
                  {completeSetup.isPending ? t('setup.submitting') : t('setup.submit')}
                </Button>
              )}
            </div>
          </form>
        ) : null}
      </Card>
    </AuthShell>
  )
}

function SystemStep({ checks, missing }: { checks: Record<string, SetupCheck>; missing: string[] }) {
  const { t } = useTranslation()
  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-base font-bold text-ink">{t('setup.system.title')}</h3>
        <p className="mt-1 text-sm text-muted">{t('setup.system.description')}</p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        {Object.entries(checks).map(([key, check]) => (
          <div key={key} className="rounded-panel border border-line bg-surface-soft p-3 text-sm">
            <div className="font-bold text-ink">{t(`setup.checks.${key}`)}</div>
            <div className={check.ok ? 'mt-1 text-success-strong' : 'mt-1 text-danger'}>
              {check.ok ? t('setup.checks.ok') : check.message || t('setup.checks.notOk')}
            </div>
          </div>
        ))}
      </div>
      <div className="rounded-panel border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800">
        <div className="font-bold">{t('setup.missingTitle')}</div>
        {missing.length ? (
          <ul className="mt-2 list-inside list-disc">
            {missing.map((item) => (
              <li key={item}>{missingLabel(t, item)}</li>
            ))}
          </ul>
        ) : (
          <p className="mt-1">{t('setup.noMissing')}</p>
        )}
      </div>
    </div>
  )
}

function AdminStep({ form, setForm }: { form: SetupForm; setForm: (next: SetupForm) => void }) {
  const { t } = useTranslation()
  return (
    <div className="grid gap-4">
      <Input label={t('setup.admin.email')} type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} autoComplete="email" required />
      <Input label={t('setup.admin.displayName')} value={form.displayName} onChange={(event) => setForm({ ...form, displayName: event.target.value })} autoComplete="name" />
      <Input label={t('setup.admin.password')} type="password" value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} autoComplete="new-password" hint={t('setup.admin.passwordHint')} required />
    </div>
  )
}

function SecurityStep({ form, setForm }: { form: SetupForm; setForm: (next: SetupForm) => void }) {
  const { t } = useTranslation()
  return (
    <div className="grid gap-4">
      <Input label={t('setup.security.webBaseUrl')} value={form.webBaseUrl} onChange={(event) => setForm({ ...form, webBaseUrl: event.target.value })} placeholder={t('setup.security.webBaseUrlPlaceholder')} required />
      <div className="rounded-panel border border-line bg-surface-soft p-3 text-sm text-muted">
        <div className="font-bold text-ink">{t('setup.security.generatedSecretsTitle')}</div>
        <p className="mt-1">{t('setup.security.generatedSecretsDescription')}</p>
      </div>
    </div>
  )
}

function ReviewStep({ form }: { form: SetupForm }) {
  const { t } = useTranslation()
  return (
    <div className="space-y-3 rounded-panel border border-line bg-surface-soft p-4 text-sm">
      <div>
        <div className="font-bold text-ink">{t('setup.review.admin')}</div>
        <div className="text-muted">{form.email}</div>
      </div>
      <div>
        <div className="font-bold text-ink">{t('setup.review.runtime')}</div>
        <div className="text-muted">{form.webBaseUrl}</div>
      </div>
      <div>
        <div className="font-bold text-ink">{t('setup.review.secrets')}</div>
        <div className="text-muted">{t('setup.review.secretsDescription')}</div>
      </div>
    </div>
  )
}

function isAbsoluteHTTPURL(value: string) {
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

function missingLabel(t: (key: string) => string, item: string) {
  switch (item) {
    case 'admin':
      return t('setup.missing.admin')
    case 'auth.session.secret':
      return t('setup.missing.sessionSecret')
    case 'auth.apiKey.secret':
      return t('setup.missing.apiKeySecret')
    default:
      return item
  }
}
