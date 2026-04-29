import apiClient from './client'
import type { BookList, ListsResponse, ListDetailResponse } from '../types'

export interface CreateListPayload {
    title: string
    description?: string
    tags?: string[]
    visibility?: 'private' | 'public'
}

export async function getLists(): Promise<ListsResponse> {
    const res = await apiClient.get<ListsResponse>('/lists')
    return res.data
}

export async function getList(id: string): Promise<ListDetailResponse> {
    const res = await apiClient.get<ListDetailResponse>(`/lists/${id}`)
    return res.data
}

export async function createList(data: CreateListPayload): Promise<BookList> {
    const res = await apiClient.post<BookList>('/lists', data)
    return res.data
}

export async function updateList(id: string, data: Partial<CreateListPayload>): Promise<BookList> {
    const res = await apiClient.put<BookList>(`/lists/${id}`, data)
    return res.data
}

export async function deleteList(id: string): Promise<void> {
    await apiClient.delete(`/lists/${id}`)
}

export async function addBookToList(
    listId: string,
    bookId: string,
    commentary?: string,
): Promise<void> {
    await apiClient.post(`/lists/${listId}/books`, { book_id: bookId, commentary })
}

export async function removeBookFromList(listId: string, bookId: string): Promise<void> {
    await apiClient.delete(`/lists/${listId}/books`, { data: { book_id: bookId } })
}
