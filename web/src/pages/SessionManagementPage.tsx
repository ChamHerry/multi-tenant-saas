import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Monitor, Smartphone, Trash2 } from 'lucide-react'
import { useAuthI18n } from '@/features/auth/auth-i18n'
import { useListSessions, useRevokeOtherSessionsMutation, useRevokeSessionMutation } from '@/features/auth/auth-hooks'
import type { SessionListItem } from '@/features/auth/auth-types'
import { useDateTimeFormatter } from '@/shared/i18n'
import { Badge, Button, Card, CardHeader, EmptyState, ErrorView, LoadingView, Table, Td, Th, Toast } from '@/shared/ui'
import { errorMessage } from '@/shared/api/errors'

export function SessionManagementPage() {
  const { t } = useTranslation()
  const { messages } = useAuthI18n()
  const copy = messages.sessions
  const dateTimeFormatter = useDateTimeFormatter()
  const sessions = useListSessions()
  const revokeSession = useRevokeSessionMutation()
  const revokeOthers = useRevokeOtherSessionsMutation()
  const [pendingSessionId, setPendingSessionId] = useState<string>()
  const [successMessage, setSuccessMessage] = useState<string>()

  const handleRevoke = async (session: SessionListItem) => {
    if (session.is_current) return
    if (!window.confirm(copy.confirmRevoke)) return
    setSuccessMessage(undefined)
    setPendingSessionId(session.session_id)
    try {
      await revokeSession.mutateAsync(session.session_id)
      setSuccessMessage(copy.revokeSuccess)
    } finally {
      setPendingSessionId(undefined)
    }
  }

  const handleRevokeOthers = async () => {
    if (!window.confirm(copy.confirmRevokeOthers)) return
    setSuccessMessage(undefined)
    const result = await revokeOthers.mutateAsync()
    setSuccessMessage(copy.revokeOthersSuccess.replace('{count}', String(result.revoked_count)))
  }

  const items = sessions.data?.sessions ?? []
  const otherSessionCount = items.filter((item) => !item.is_current).length
  const firstError = sessions.error ?? revokeSession.error ?? revokeOthers.error

  return (
    <div className="space-y-6">
      {firstError ? <ErrorView error={firstError} title={copy.errors.loadFailed} /> : null}
      <Toast tone="green" message={successMessage} />

      <Card>
        <CardHeader
          title={copy.title}
          description={copy.description}
          action={(
            <Button
              variant="danger"
              size="sm"
              disabled={otherSessionCount === 0}
              isLoading={revokeOthers.isPending}
              onClick={() => void handleRevokeOthers()}
            >
              {copy.revokeOthers}
            </Button>
          )}
        />

        {sessions.isLoading ? <LoadingView label={copy.loading} /> : null}
        {!sessions.isLoading && items.length === 0 ? (
          <EmptyState title={copy.emptyTitle} description={copy.emptyDescription} />
        ) : null}
        {items.length > 0 ? (
          <div className="overflow-x-auto">
            <Table>
              <thead>
                <tr>
                  <Th>{copy.headers.device}</Th>
                  <Th>{copy.headers.ip}</Th>
                  <Th>{copy.headers.lastActive}</Th>
                  <Th>{copy.headers.created}</Th>
                  <Th>{copy.headers.expires}</Th>
                  <Th>{t('common.fields.actions')}</Th>
                </tr>
              </thead>
              <tbody>
                {items.map((item) => (
                  <tr key={item.session_id}>
                    <Td>
                      <div className="flex items-start gap-3">
                        <span className="mt-0.5 rounded-panel bg-brand-soft p-2 text-brand">
                          {item.device.is_mobile ? <Smartphone className="size-4" /> : <Monitor className="size-4" />}
                        </span>
                        <div>
                          <div className="font-bold text-ink">
                            {deviceLabel(item)}
                          </div>
                          <div className="mt-1 text-xs text-subtle">
                            {shortSession(item.session_id)}
                          </div>
                          {item.is_current ? <Badge className="mt-2" tone="green">{copy.currentBadge}</Badge> : null}
                        </div>
                      </div>
                    </Td>
                    <Td>{item.ip}</Td>
                    <Td>{dateTimeFormatter.format(new Date(item.last_active_at))}</Td>
                    <Td>{dateTimeFormatter.format(new Date(item.created_at))}</Td>
                    <Td>{dateTimeFormatter.format(new Date(item.expires_at))}</Td>
                    <Td>
                      {item.is_current ? (
                        <span className="text-xs text-subtle">{copy.currentHint}</span>
                      ) : (
                        <Button
                          size="sm"
                          variant="danger"
                          leftIcon={<Trash2 className="size-3.5" />}
                          isLoading={pendingSessionId === item.session_id && revokeSession.isPending}
                          onClick={() => void handleRevoke(item)}
                        >
                          {copy.revoke}
                        </Button>
                      )}
                    </Td>
                  </tr>
                ))}
              </tbody>
            </Table>
          </div>
        ) : null}

        <p className="mt-4 text-xs leading-5 text-subtle">{copy.privacyNote}</p>
        <Toast tone="red" message={revokeSession.isError ? errorMessage(revokeSession.error) : revokeOthers.isError ? errorMessage(revokeOthers.error) : undefined} />
      </Card>
    </div>
  )
}

function deviceLabel(item: SessionListItem) {
  const browser = item.device.browser_version ? `${item.device.browser} ${item.device.browser_version}` : item.device.browser
  return `${browser} · ${item.device.os}`
}

function shortSession(sessionID: string) {
  if (sessionID.length <= 12) return sessionID
  return `${sessionID.slice(0, 8)}…${sessionID.slice(-4)}`
}
