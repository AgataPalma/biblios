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

export async function lookupByTitleAuthor(title: string, author: string): Promise<LookupResult> {
    const response = await apiClient.get<LookupResult>(`/books/lookup?title=${encodeURIComponent(title)}&author=${encodeURIComponent(author)}`)
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