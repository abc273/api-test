export type PortraitAssetStatus =
  | 'pending'
  | 'qr_ready'
  | 'waiting_upload'
  | 'waiting_accept'
  | 'pending_confirm'
  | 'ready'
  | 'failed'
  | 'disabled'
  | 'expired'

export interface PortraitAssetJob {
  id: number
  user_id: number
  name: string
  status: PortraitAssetStatus
  invite_url?: string
  qr_image?: string
  volc_group_id?: string
  asset_id?: string
  asset_status?: string
  asset_preview?: string
  error_message?: string
  created_time: number
  updated_time: number
  accept_time?: number
  qr_expires_time?: number
  ready_time?: number
  queue_position?: number
}

export interface PortraitAssetPage {
  page: number
  page_size: number
  total: number
  items: PortraitAssetJob[]
}

export interface ApiResponse<T> {
  success: boolean
  message?: string
  data?: T
}
