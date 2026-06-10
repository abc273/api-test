import { api } from '@/lib/api'
import type {
  ApiResponse,
  VirtualPortraitAsset,
  VirtualPortraitAssetConfig,
  VirtualPortraitAssetGroup,
  VirtualPortraitAssetPage,
  VirtualPortraitAssetUploadResult,
} from './types'

export async function getVirtualPortraitAssetConfig(): Promise<
  ApiResponse<VirtualPortraitAssetConfig>
> {
  const res = await api.get('/api/portrait_assets/virtual/config')
  return res.data
}

export async function getUserVirtualPortraitAssetGroup(): Promise<
  ApiResponse<VirtualPortraitAssetGroup | null>
> {
  const res = await api.get('/api/portrait_assets/virtual/group')
  return res.data
}

export async function getUserVirtualPortraitAssets(params: {
  p?: number
  page_size?: number
}): Promise<ApiResponse<VirtualPortraitAssetPage>> {
  const res = await api.get('/api/portrait_assets/virtual/assets', { params })
  return res.data
}

export async function uploadVirtualPortraitAssetMaterial(
  file: File
): Promise<ApiResponse<VirtualPortraitAssetUploadResult>> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await api.post('/api/portrait_assets/virtual/upload', formData)
  return res.data
}

export async function createUserVirtualPortraitAsset(payload: {
  name?: string
  asset_url: string
  asset_type: string
}): Promise<ApiResponse<VirtualPortraitAsset>> {
  const res = await api.post('/api/portrait_assets/virtual/assets', payload)
  return res.data
}

export async function syncUserVirtualPortraitAsset(
  id: number
): Promise<ApiResponse<VirtualPortraitAsset>> {
  const res = await api.post(`/api/portrait_assets/virtual/assets/${id}/sync`)
  return res.data
}
