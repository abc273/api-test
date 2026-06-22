import { api } from '@/lib/api'

export type PortraitAssetKind = 'official' | 'virtual'

export interface PortraitAssetFolder {
  id: number
  user_id: number
  external_user_id?: string
  asset_kind: PortraitAssetKind
  name: string
  sort_order: number
  created_time: number
  updated_time: number
}

export interface ApiResponse<T> {
  success: boolean
  message?: string
  data?: T
}

export async function getPortraitAssetFolders(params: {
  asset_kind: PortraitAssetKind
  external_user_id?: string
}): Promise<ApiResponse<PortraitAssetFolder[]>> {
  const res = await api.get('/api/portrait_assets/folders', { params })
  return res.data
}

export async function createPortraitAssetFolder(payload: {
  asset_kind: PortraitAssetKind
  name: string
  external_user_id?: string
  sort_order?: number
}): Promise<ApiResponse<PortraitAssetFolder>> {
  const res = await api.post('/api/portrait_assets/folders', payload)
  return res.data
}

export async function deletePortraitAssetFolder(
  folderId: number,
  params?: { external_user_id?: string }
): Promise<ApiResponse<null>> {
  const res = await api.delete(`/api/portrait_assets/folders/${folderId}`, {
    params,
  })
  return res.data
}

export async function movePortraitAssetsToFolder(payload: {
  asset_kind: PortraitAssetKind
  asset_ids: number[]
  folder_id: number
  external_user_id?: string
}): Promise<ApiResponse<{ moved: number }>> {
  const res = await api.post('/api/portrait_assets/folders/move', payload)
  return res.data
}
