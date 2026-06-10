import { useTranslation } from 'react-i18next'
import { LegalDocument } from '@/features/legal/legal-document'
import { getApiDocs } from './api'

export function ApiDocsPage() {
  const { t } = useTranslation()

  return (
    <LegalDocument
      title={t('API Docs')}
      queryKey='api-docs'
      fetchDocument={getApiDocs}
      emptyMessage={t(
        'The administrator has not configured the API docs yet.'
      )}
    />
  )
}
