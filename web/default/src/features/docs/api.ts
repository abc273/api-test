import { api } from '@/lib/api'
import type { ApiDocsResponse } from './types'

export async function getApiDocs() {
  const res = await api.get<ApiDocsResponse>('/api/docs')
  return res.data
}
