import apiClient from './client'
import type { Collection, CollectionBooksResponse } from '../types'

export interface CreateCollectionPayload {
    name: string
    description?: string
    cover_colour?: string
    is_public?: boolean
    is_collaborative?: boolean
}

export interface UpdateCollectionPayload {
    name?: string
    description?: string
    is_public?: boolean
}

export async function getCollections(libraryId: string): Promise<Collection[]> {
    const res = await apiClient.get<Collection[]>(`/libraries/${libraryId}/collections`)
    return res.data
}

export async function getCollection(libraryId: string, collectionId: string): Promise<Collection> {
    const res = await apiClient.get<Collection>(`/libraries/${libraryId}/collections/${collectionId}`)
    return res.data
}

export async function getCollectionBooks(
    libraryId: string,
    collectionId: string,
    page = 1,
    limit = 20,
): Promise<CollectionBooksResponse> {
    const res = await apiClient.get<CollectionBooksResponse>(
        `/libraries/${libraryId}/collections/${collectionId}/books`,
        { params: { page, limit } },
    )
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
    libraryId: string,
    collectionId: string,
    data: UpdateCollectionPayload,
): Promise<Collection> {
    const res = await apiClient.put<Collection>(`/libraries/${libraryId}/collections/${collectionId}`, data)
    return res.data
}

export async function deleteCollection(libraryId: string, collectionId: string): Promise<void> {
    await apiClient.delete(`/libraries/${libraryId}/collections/${collectionId}`)
}

export async function addBookToCollection(libraryId: string, collectionId: string, copyId: string): Promise<void> {
    await apiClient.post(`/libraries/${libraryId}/collections/${collectionId}/books`, { copy_id: copyId })
}

export async function removeBookFromCollection(libraryId: string, collectionId: string, copyId: string): Promise<void> {
    await apiClient.delete(`/libraries/${libraryId}/collections/${collectionId}/books/${copyId}`)
}
