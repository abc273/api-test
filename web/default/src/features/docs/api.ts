import { api } from '@/lib/api'
import type {
  ApiDocDiffResponse,
  ApiDocsChangelogResponse,
  ApiDocsMetaResponse,
  ApiDocsResponse,
  PublishApiDocsRequest,
  PublishApiDocsResponse,
} from './types'

export async function getApiDocs() {
  const res = await api.get<ApiDocsResponse>('/api/docs')
  return res.data
}

export async function getApiDocsMeta() {
  const res = await api.get<ApiDocsMetaResponse>('/api/docs/meta')
  return res.data
}

export async function getApiDocsChangelog(page = 1, pageSize = 20) {
  const res = await api.get<ApiDocsChangelogResponse>('/api/docs/changelog', {
    params: { p: page, page_size: pageSize },
  })
  return res.data
}

export async function publishApiDocs(request: PublishApiDocsRequest) {
  const res = await api.post<PublishApiDocsResponse>(
    '/api/docs/publish',
    request
  )
  return res.data
}

export async function diffApiDocs(request: {
  from_version?: string
  from_content?: string
  to_version?: string
  to_content: string
}) {
  const res = await api.post<ApiDocDiffResponse>('/api/docs/diff', request)
  return res.data
}
