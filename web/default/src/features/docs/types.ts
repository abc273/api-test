export type ApiDocsResponse = {
  success: boolean
  message?: string
  data?: string
}

export type ApiDocChangeItem = {
  id: number
  revision_id: number
  change_type: 'added' | 'changed' | 'deprecated' | 'removed' | 'fixed'
  endpoint?: string
  method?: string
  section?: string
  description: string
  impact?: string
  created_time: number
}

export type ApiDocRevision = {
  id: number
  version: string
  title: string
  summary: string
  changed_sections: string[]
  content?: string
  content_sha256: string
  source_commit?: string
  published_by: number
  published_at: number
  created_time: number
  change_items?: ApiDocChangeItem[]
}

export type ApiDocsMetaResponse = {
  success: boolean
  message?: string
  data?: ApiDocRevision
}

export type ApiDocsChangelogResponse = {
  success: boolean
  message?: string
  data?: {
    page: number
    page_size: number
    total: number
    items: ApiDocRevision[]
  }
}

export type ApiDocDiffLine = {
  type: 'context' | 'added' | 'removed'
  text: string
}

export type ApiDocDiffResult = {
  from_version: string
  to_version: string
  changed: boolean
  added_lines: number
  removed_lines: number
  lines: ApiDocDiffLine[]
}

export type ApiDocDiffResponse = {
  success: boolean
  message?: string
  data?: ApiDocDiffResult
}

export type PublishApiDocsRequest = {
  version: string
  title: string
  summary: string
  changed_sections: string[]
  change_items: Array<{
    change_type: ApiDocChangeItem['change_type']
    endpoint?: string
    method?: string
    section?: string
    description: string
    impact?: string
  }>
  content: string
  source_commit?: string
}

export type PublishApiDocsResponse = {
  success: boolean
  message?: string
  data?: ApiDocRevision
}
