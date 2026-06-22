export type OfficialPortraitAssetStatus =
  | 'pending'
  | 'validate_ready'
  | 'validated'
  | 'asset_processing'
  | 'pending_confirm'
  | 'ready'
  | 'failed'
  | 'expired'
  | 'disabled'

export interface OfficialPortraitAssetJob {
  id: number
  user_id: number
  external_user_id?: string
  folder_id: number
  name: string
  source: 'official'
  status: OfficialPortraitAssetStatus
  invite_url?: string
  validate_result_code?: string
  callback_url?: string
  volc_group_id?: string
  asset_id?: string
  asset_status?: string
  asset_preview?: string
  asset_url?: string
  asset_type?: string
  project_name?: string
  error_message?: string
  created_time: number
  updated_time: number
  qr_expires_time?: number
  ready_time?: number
}

export interface OfficialPortraitAssetPage {
  page: number
  page_size: number
  total: number
  items: OfficialPortraitAssetJob[]
}

export interface OfficialPortraitAssetConfig {
  configured: boolean
  project_name: string
}

export interface OfficialPortraitAssetUploadResult {
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
