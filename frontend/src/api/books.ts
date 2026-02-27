import apiClient from './client'
import type { Book, LookupResult, Submission } from '../types'

export interface BooksResponse {
    books: Book[]
    total: number
    page: number
    limit: number
}

export interface SubmitBookPayload {
    title: string
    description?: string
    cover_url?: string
    authors: string[]
    genres: string[]
    edition: {
        format: string
        isbn?: string
        language: string
        publisher?: string
        page_count?: number
        narrator?: string
        duration_minutes?: number
    }
    condition?: string
}

export async function listBooks(page: number = 1, limit: number = 20): Promise<BooksResponse> {
    const response = await apiClient.get<BooksResponse>(`/books?page=${page}&limit=${limit}`)
    return response.data
}

export async function getBook(id: string): Promise<Book> {
    const response = await apiClient.get<Book>(`/books/${id}`)
    return response.data
}

export async function submitBook(data: SubmitBookPayload): Promise<Submission> {
    const response = await apiClient.post<Submission>('/books', data)
    return response.data
}

export async function lookupByISBN(isbn: string): Promise<LookupResult> {
    const response = await apiClient.get<LookupResult>(`/books/lookup?isbn=${isbn}`)
    return response.data
}

export async function lookupByTitleAuthor(title: string, author?: string): Promise<LookupResult> {
    const params = new URLSearchParams({ title })
    if (author) params.append('author', author)
    const response = await apiClient.get<LookupResult>(`/books/lookup?${params.toString()}`)
    return response.data
}

export async function checkDuplicate(isbn: string): Promise<{ exists: boolean; edition?: object }> {
    const response = await apiClient.get(`/books/check?isbn=${isbn}`)
    return response.data
}

export async function getMyBooks(page: number = 1, limit: number = 20): Promise<BooksResponse> {
    const response = await apiClient.get<BooksResponse>(`/users/me/books?page=${page}&limit=${limit}`)
    return response.data
}

export async function getMyLibrary(page = 1, limit = 20) {
    const res = await apiClient.get('/users/me/library', { params: { page, limit } })
    return res.data
}

export async function updateReadingStatus(copyId: string, status: string): Promise<void> {
    await apiClient.put(`/books/copies/${copyId}/status`, { status })
}

export async function removeCopy(copyId: string): Promise<void> {
    await apiClient.delete(`/books/copies/${copyId}`)
}

export async function getPendingSubmissions(page = 1, limit = 20) {
    const res = await apiClient.get('/moderation/submissions', { params: { page, limit } })
    return res.data
}

export async function approveSubmission(id: string): Promise<void> {
    await apiClient.put(`/moderation/submissions/${id}/approve`)
}

export async function rejectSubmission(id: string, reason: string): Promise<void> {
    await apiClient.put(`/moderation/submissions/${id}/reject`, { reason })
}