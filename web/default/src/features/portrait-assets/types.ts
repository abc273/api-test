export type PortraitAssetStatus =
  | 'pending'
  | 'qr_ready'
  | 'waiting_upload'
  | 'waiting_accept'
  | 'ready'
  | 'failed'
  | 'disabled'

export interface PortraitAssetJob {
  id: number
  user_id: number
  name: string
  status: PortraitAssetStatus
  invite_url?: string
  qr_image?: string
  volc_group_id?: string
  asset_id?: string
  error_message?: string
  created_time: number
  updated_time: number
  ready_time?: number
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
