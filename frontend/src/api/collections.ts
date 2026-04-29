import apiClient from './client'
import type { Collection, CollectionsResponse, CollectionDetailResponse } from '../types'

export interface CreateCollectionPayload {
    name: string
    description?: string
    visibility?: 'public' | 'private'
    is_collaborative?: boolean
}

export async function getCollections(libraryId: string): Promise<CollectionsResponse> {
    const res = await apiClient.get<CollectionsResponse>(`/libraries/${libraryId}/collections`)
    return res.data
}

export async function getCollection(id: string): Promise<CollectionDetailResponse> {
    const res = await apiClient.get<CollectionDetailResponse>(`/collections/${id}`)
    return res.data
}

export async function createCollection(
    libraryId: string,
    data: CreateCollectionPayload,
): Promise<Collection> {
    const res = await apiClient.post<Collection>(`/libraries/${libraryId}/collections`, data)
    return res.data
}

export async function updateCollection(
    id: string,
    data: Partial<CreateCollectionPayload>,
): Promise<Collection> {
    const res = await apiClient.put<Collection>(`/collections/${id}`, data)
    return res.data
}

export async function deleteCollection(id: string): Promise<void> {
    await apiClient.delete(`/collections/${id}`)
}

export async function addBookToCollection(collectionId: string, bookId: string): Promise<void> {
    await apiClient.post(`/collections/${collectionId}/books`, { book_id: bookId })
}

export async function removeBookFromCollection(collectionId: string, bookId: string): Promise<void> {
    await apiClient.delete(`/collections/${collectionId}/books`, { data: { book_id: bookId } })
}
