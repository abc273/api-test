export type VirtualPortraitAssetGroupStatus = 'creating' | 'active' | 'failed'

export type VirtualPortraitAssetStatus = 'processing' | 'active' | 'failed'

export interface VirtualPortraitAssetConfig {
  configured: boolean
  project_name: string
}

export interface VirtualPortraitAssetGroup {
  id: number
  user_id: number
  name: string
  description?: string
  project_name: string
  volc_group_id?: string
  status: VirtualPortraitAssetGroupStatus
  error_message?: string
  created_time: number
  updated_time: number
}

export interface VirtualPortraitAsset {
  id: number
  user_id: number
  group_id: number
  name: string
  asset_type: 'Image' | 'Video' | 'Audio'
  source_url?: string
  preview_url?: string
  project_name: string
  volc_group_id?: string
  volc_asset_id?: string
  status: VirtualPortraitAssetStatus
  volc_status?: string
  error_message?: string
  created_time: number
  updated_time: number
  ready_time?: number
}

export interface VirtualPortraitAssetPage {
  page: number
  page_size: number
  total: number
  items: VirtualPortraitAsset[]
}

export interface VirtualPortraitAssetUploadResult {
  url: string
  file_name: string
  content_type: string
  asset_type: 'Image' | 'Video' | 'Audio'
  size: number
}

export interface ApiResponse<T> {
  success: boolean
  message?: string
  data?: T
}
