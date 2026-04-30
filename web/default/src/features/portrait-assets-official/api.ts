import { api } from '@/lib/api'
import type {
  ApiResponse,
  OfficialPortraitAssetConfig,
  OfficialPortraitAssetJob,
  OfficialPortraitAssetPage,
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
}): Promise<ApiResponse<OfficialPortraitAssetPage>> {
  const res = await api.get('/api/portrait_assets/official/jobs', { params })
  return res.data
}

export async function createOfficialPortraitAssetJob(
  name: string
): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post('/api/portrait_assets/official/jobs', { name })
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
  payload: { asset_url: string; asset_type: string; name?: string }
): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post(
    `/api/portrait_assets/official/jobs/${id}/asset`,
    payload
  )
  return res.data
}

export async function syncOfficialPortraitAssetJob(
  id: number
): Promise<ApiResponse<OfficialPortraitAssetJob>> {
  const res = await api.post(`/api/portrait_assets/official/jobs/${id}/sync`)
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
