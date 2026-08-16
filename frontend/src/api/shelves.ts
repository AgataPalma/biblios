import apiClient from './client'
import type { Shelf, ShelfBooksResponse } from '../types'

export interface CreateShelfPayload {
    name: string
}

export interface UpdateShelfPayload {
    name: string
}

export async function getShelves(): Promise<Shelf[]> {
    const res = await apiClient.get<Shelf[]>('/shelves')
    return res.data
}

export async function getShelf(id: string, page = 1, limit = 20): Promise<ShelfBooksResponse> {
    const res = await apiClient.get<ShelfBooksResponse>(`/shelves/${id}/books`, {
        params: { page, limit },
    })
    return res.data
}

export async function createShelf(data: CreateShelfPayload): Promise<Shelf> {
    const res = await apiClient.post<Shelf>('/shelves', data)
    return res.data
}

export async function updateShelf(id: string, data: UpdateShelfPayload): Promise<Shelf> {
    const res = await apiClient.put<Shelf>(`/shelves/${id}`, data)
    return res.data
}

export async function deleteShelf(id: string): Promise<void> {
    await apiClient.delete(`/shelves/${id}`)
}

export async function addBookToShelf(shelfId: string, copyId: string): Promise<void> {
    await apiClient.post(`/shelves/${shelfId}/books`, { copy_id: copyId })
}

export async function removeBookFromShelf(shelfId: string, copyId: string): Promise<void> {
    await apiClient.delete(`/shelves/${shelfId}/books/${copyId}`)
}
