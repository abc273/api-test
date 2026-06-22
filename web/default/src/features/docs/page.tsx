import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { LegalDocument } from '@/features/legal/legal-document'
import { getApiDocs, getApiDocsChangelog, getApiDocsMeta } from './api'

const SEEN_VERSION_KEY = 'api_docs_seen_version'

function formatDocTime(timestamp?: number) {
  if (!timestamp) return ''
  return new Date(timestamp * 1000).toLocaleString()
}

export function ApiDocsPage() {
  const { t } = useTranslation()
  const [seenVersion, setSeenVersion] = useState(() => {
    if (typeof window === 'undefined') return ''
    return window.localStorage.getItem(SEEN_VERSION_KEY) ?? ''
  })
  const [showHistory, setShowHistory] = useState(false)

  const { data: metaData } = useQuery({
    queryKey: ['api-docs-meta'],
    queryFn: getApiDocsMeta,
    staleTime: 10 * 60 * 1000,
  })
  const { data: changelogData } = useQuery({
    queryKey: ['api-docs-changelog'],
    queryFn: () => getApiDocsChangelog(1, 10),
    staleTime: 10 * 60 * 1000,
  })

  const meta = metaData?.data
  const version = meta?.version ?? ''
  const hasUnseenVersion = Boolean(version && version !== seenVersion)

  useEffect(() => {
    if (!version || seenVersion) return
    setSeenVersion('')
  }, [seenVersion, version])

  const markVersionAsSeen = () => {
    if (!version || typeof window === 'undefined') return
    window.localStorage.setItem(SEEN_VERSION_KEY, version)
    setSeenVersion(version)
  }

  const headerExtra = meta ? (
    <div className='mt-6 space-y-4'>
      {hasUnseenVersion && (
        <Card className='border-primary/30 bg-primary/5'>
          <CardContent className='flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between'>
            <div>
              <div className='font-medium'>{t('New API docs version available')}</div>
              <p className='text-muted-foreground text-sm'>
                {t('Review the latest API docs changes before integration.')}
              </p>
            </div>
            <Button size='sm' onClick={markVersionAsSeen}>
              {t('Mark as read')}
            </Button>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardContent className='space-y-4 p-4'>
          <div className='flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between'>
            <div className='space-y-2'>
              <div className='flex flex-wrap items-center gap-2'>
                <Badge variant='secondary'>{version || t('No version')}</Badge>
                {meta.published_at > 0 && (
                  <span className='text-muted-foreground text-sm'>
                    {formatDocTime(meta.published_at)}
                  </span>
                )}
              </div>
              {meta.summary && (
                <p className='text-muted-foreground text-sm'>{meta.summary}</p>
              )}
            </div>
            <Button
              type='button'
              variant='outline'
              size='sm'
              onClick={() => setShowHistory((value) => !value)}
            >
              {showHistory ? t('Hide update history') : t('View update history')}
            </Button>
          </div>

          {meta.changed_sections.length > 0 && (
            <div className='flex flex-wrap gap-2'>
              {meta.changed_sections.map((section) => (
                <Badge key={section} variant='outline'>
                  {section}
                </Badge>
              ))}
            </div>
          )}

          {showHistory && (
            <div className='space-y-3 border-t pt-4'>
              {(changelogData?.data?.items ?? []).map((revision) => (
                <div key={revision.version} className='rounded-lg border p-3'>
                  <div className='flex flex-wrap items-center gap-2'>
                    <Badge variant='secondary'>{revision.version}</Badge>
                    <span className='text-muted-foreground text-xs'>
                      {formatDocTime(revision.published_at)}
                    </span>
                  </div>
                  <p className='mt-2 text-sm'>{revision.summary}</p>
                  {revision.change_items && revision.change_items.length > 0 && (
                    <div className='mt-3 space-y-2'>
                      {revision.change_items.map((item) => (
                        <div
                          key={item.id}
                          className='text-muted-foreground text-xs'
                        >
                          <span className='font-medium'>{item.change_type}</span>
                          {item.method && item.endpoint
                            ? ` ${item.method} ${item.endpoint}`
                            : ''}
                          {item.description ? `: ${item.description}` : ''}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  ) : null

  return (
    <LegalDocument
      title={t('API Docs')}
      queryKey='api-docs'
      fetchDocument={getApiDocs}
      emptyMessage={t('The administrator has not configured the API docs yet.')}
      headerExtra={headerExtra}
    />
  )
}
