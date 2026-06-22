import { api } from '@/lib/api'
import type {
  ApiResponse,
  OfficialPortraitAssetConfig,
  OfficialPortraitAssetJob,
  OfficialPortraitAssetPage,
  OfficialPortraitAssetUploadResult,
} from './types'

export async function getOfficialPortraitAssetConfig(): Promise<
  ApiResponse<OfficialPortraitAssetConfig>
> {
  const res = await api.get('/api/portrait_assets/official/config')
  return res.data
}

export async function getOfficialPortraitAssets(params: {
  p?: number
  page_size?: number
  external_user_id?: string
  folder_id?: number
}): Promise<ApiResponse<OfficialPortraitAssetPage>> {
  const res = await api.get('/api/portrait_assets/official/jobs', { params })
  return res.data
}

export async function createOfficialPortraitAssetJob(payload: {
  name: string
  callback_url?: string
  external_user_id?: string
  folder_id?: number
}): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post('/api/portrait_assets/official/jobs', payload)
  return res.data
}

export async function refreshOfficialPortraitValidation(
  id: number
): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post(
    `/api/portrait_assets/official/jobs/${id}/validation`
  )
  return res.data
}

export async function submitOfficialPortraitAsset(
  id: number,
  payload: {
    asset_url: string
    asset_type: string
    name?: string
    folder_id?: number
  }
): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post(
    `/api/portrait_assets/official/jobs/${id}/asset`,
    payload
  )
  return res.data
}

export async function uploadOfficialPortraitAssetMaterial(
  file: File
): Promise<ApiResponse<OfficialPortraitAssetUploadResult>> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await api.post('/api/portrait_assets/official/upload', formData)
  return res.data
}

export async function syncOfficialPortraitAssetJob(
  id: number
): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post(`/api/portrait_assets/official/jobs/${id}/sync`)
  return res.data
}

export async function deleteOfficialPortraitAssetJob(
  id: number,
  params?: {
    external_user_id?: string
  }
): Promise<ApiResponse<null>> {
  const res = await api.delete(`/api/portrait_assets/official/jobs/${id}`, {
    params,
  })
  return res.data
}

export async function confirmOfficialPortraitAsset(
  id: number
): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post(`/api/portrait_assets/official/jobs/${id}/confirm`)
  return res.data
}

export async function rejectOfficialPortraitAsset(
  id: number
): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post(`/api/portrait_assets/official/jobs/${id}/reject`)
  return res.data
}
