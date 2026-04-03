import axios from 'axios'
import type { ApiResponse, Memory, ListResponse, StatsResponse, CreateMemoryRequest } from './types'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

export const memoriesApi = {
  list: async (params?: { scope?: string; limit?: number; offset?: number }) => {
    const { data } = await client.get<ApiResponse<ListResponse>>('/memories', { params })
    return data
  },

  get: async (id: string) => {
    const { data } = await client.get<ApiResponse<Memory>>(`/memories/${id}`)
    return data
  },

  create: async (memory: CreateMemoryRequest) => {
    const { data } = await client.post<ApiResponse<Memory>>('/memories', memory)
    return data
  },

  delete: async (id: string) => {
    const { data } = await client.delete<ApiResponse<{ id: string }>>(`/memories/${id}`)
    return data
  },

  query: async (q: string, topk = 5, scope?: string) => {
    const { data } = await client.post<ApiResponse<unknown>>('/query', {
      query: q,
      topk,
      scope,
    })
    return data
  },

  stats: async () => {
    const { data } = await client.get<ApiResponse<StatsResponse>>('/stats')
    return data
  },
}