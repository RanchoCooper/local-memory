export interface Memory {
  id: string
  type: MemoryType
  scope: Scope
  media_type: MediaType
  key: string
  value: string
  confidence: number
  related_ids: string[]
  tags: string[]
  metadata: Metadata
  deleted: boolean
  deleted_at: number
  created_at: number
  updated_at: number
}

export type MemoryType = 'preference' | 'fact' | 'event' | 'skill' | 'goal' | 'relationship'
export type Scope = 'global' | 'session' | 'agent'
export type MediaType = 'text' | 'image' | 'audio' | 'video'

export interface Metadata {
  source?: string
  language?: string
  file_path?: string
  file_size?: number
  mime_type?: string
  agent_id?: string
  session_id?: string
  extra?: Record<string, unknown>
}

export interface CreateMemoryRequest {
  type: MemoryType
  scope: Scope
  key: string
  value: string
  confidence?: number
  tags?: string[]
}

export interface ListResponse {
  memories: Memory[]
  total: number
}

export interface StatsResponse {
  total: number
  by_type: Record<string, number>
  by_scope: Record<string, number>
  by_media: Record<string, number>
  deleted: number
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
}