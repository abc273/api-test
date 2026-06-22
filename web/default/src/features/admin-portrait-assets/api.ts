import { api } from '@/lib/api'
import type {
  OfficialPortraitAssetStatus,
  OfficialPortraitAssetPage,
} from '@/features/portrait-assets-official/types'
import type {
  VirtualPortraitAssetPage,
  VirtualPortraitAssetStatus,
} from '@/features/portrait-assets-virtual/types'

export interface PortraitAssetOwner {
  user_id: number
  username?: string
  email?: string
}

export type AdminOfficialPortraitAssetItem =
  OfficialPortraitAssetPage['items'][number] & {
    owner: PortraitAssetOwner
  }

export type AdminVirtualPortraitAssetItem =
  VirtualPortraitAssetPage['items'][number] & {
    owner: PortraitAssetOwner
  }

export interface AdminOfficialPortraitAssetPage {
  page: number
  page_size: number
  total: number
  items: AdminOfficialPortraitAssetItem[]
}

export interface AdminVirtualPortraitAssetPage {
  page: number
  page_size: number
  total: number
  items: AdminVirtualPortraitAssetItem[]
}

export interface ApiResponse<T> {
  success: boolean
  message?: string
  data?: T
}

export interface AdminPortraitAssetListParams {
  p?: number
  page_size?: number
  user_id?: number
  external_user_id?: string
  status?: string
  keyword?: string
  asset_id?: string
}

export async function getAdminOfficialPortraitAssets(
  params: AdminPortraitAssetListParams & { status?: OfficialPortraitAssetStatus }
): Promise<ApiResponse<AdminOfficialPortraitAssetPage>> {
  const res = await api.get('/api/admin/portrait_assets/official/jobs', {
    params,
  })
  return res.data
}

export async function syncAdminOfficialPortraitAsset(
  id: number
): Promise<ApiResponse<AdminOfficialPortraitAssetItem>> {
  const res = await api.post(`/api/admin/portrait_assets/official/jobs/${id}/sync`)
  return res.data
}

export async function deleteAdminOfficialPortraitAsset(
  id: number
): Promise<ApiResponse<null>> {
  const res = await api.delete(`/api/admin/portrait_assets/official/jobs/${id}`)
  return res.data
}

export async function getAdminVirtualPortraitAssets(
  params: AdminPortraitAssetListParams & { status?: VirtualPortraitAssetStatus }
): Promise<ApiResponse<AdminVirtualPortraitAssetPage>> {
  const res = await api.get('/api/admin/portrait_assets/virtual/assets', {
    params,
  })
  return res.data
}

export async function syncAdminVirtualPortraitAsset(
  id: number
): Promise<ApiResponse<AdminVirtualPortraitAssetItem>> {
  const res = await api.post(`/api/admin/portrait_assets/virtual/assets/${id}/sync`)
  return res.data
}

export async function deleteAdminVirtualPortraitAsset(
  id: number
): Promise<ApiResponse<null>> {
  const res = await api.delete(`/api/admin/portrait_assets/virtual/assets/${id}`)
  return res.data
}
