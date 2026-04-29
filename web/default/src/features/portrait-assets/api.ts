import { api } from '@/lib/api'
import type { ApiResponse, PortraitAssetJob, PortraitAssetPage } from './types'

export async function getPortraitAssets(params: {
  p?: number
  page_size?: number
}): Promise<ApiResponse<PortraitAssetPage>> {
  const res = await api.get('/api/portrait_assets/', { params })
  return res.data
}

export async function createPortraitAssetJob(
  name: string
): Promise<ApiResponse<PortraitAssetJob>> {
  const res = await api.post('/api/portrait_assets/jobs', { name })
  return res.data
}
