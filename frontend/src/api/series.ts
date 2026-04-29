import apiClient from './client'
import type { Series, SeriesResponse, SeriesDetailResponse } from '../types'

export interface CreateSeriesPayload {
    name: string
    description?: string
}

export async function getSeries(page = 1, limit = 20): Promise<SeriesResponse> {
    const res = await apiClient.get<SeriesResponse>('/series', { params: { page, limit } })
    return res.data
}

export async function getSeriesById(id: string): Promise<SeriesDetailResponse> {
    const res = await apiClient.get<SeriesDetailResponse>(`/series/${id}`)
    return res.data
}

export async function createSeries(data: CreateSeriesPayload): Promise<Series> {
    const res = await apiClient.post<Series>('/series', data)
    return res.data
}
