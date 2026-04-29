import apiClient from './client'
import type { Genre } from '../types'

export async function getGenres(): Promise<Genre[]> {
    const res = await apiClient.get<Genre[]>('/genres')
    return res.data
}
