import apiClient from './client'
import type { Book, LookupResult, LookupResultsPage, Submission, Review, ReviewsResponse } from '../types'


export interface BooksResponse {
    books: Book[]
    total: number
    page: number
    limit: number
}

export interface UpdateBookPayload {
    title?: string
    description?: string
    cover_url?: string
    authors?: string[]
    genres?: string[]
    edition?: {
        format: string
        isbn?: string
        language: string
        publisher?: string
        edition?: string          // edition name/label
        page_count?: number
        narrator?: string         // audiobook narrator name (backend creates Narrator)
        translator?: string       // translator name (backend creates Translator)
        duration_minutes?: number
    }

}

export interface SubmitBookPayload {
    title: string
    description?: string
    cover_url?: string
    authors: string[]
    genres: string[]
    catalogue_only?: boolean
    edition: {
        format: string
        isbn?: string
        asin?: string
        language: string
        publisher?: string
        edition?: string          // edition name/label e.g. "Illustrated", "10th Anniversary"
        published_at?: string     // e.g. "2001" or "2001-09-01"
        page_count?: number
        file_format?: string      // ebook: EPUB | PDF | MOBI | AZW3
        narrator?: string         // audiobook narrator name
        translators?: string[]    // translator names (multiple allowed)
        duration_minutes?: number
        audio_format?: string     // audiobook: MP3 | AAC | WMA | FLAC
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

// Upload a cover image file for a book (mod/admin only).
// Sends multipart/form-data to POST /books/{id}/cover and returns the stored public URL.
export async function uploadBookCover(bookId: string, file: File): Promise<string> {
    const formData = new FormData()
    formData.append('cover', file)
    const res = await apiClient.post<{ cover_url: string }>(`/books/${bookId}/cover`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data.cover_url
}

export async function lookupByTitleAuthor(
    title: string,
    author?: string,
    page: number = 1
): Promise<LookupResultsPage> {
    const params = new URLSearchParams({ title, page: String(page) })
    if (author) params.append('author', author)
    const response = await apiClient.get<LookupResultsPage>(`/books/lookup?${params.toString()}`)
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


export async function getBookReviews(
    bookId: string,
    page = 1,
    limit = 10,
): Promise<ReviewsResponse> {
    const res = await apiClient.get<ReviewsResponse>(
        `/books/${bookId}/reviews?page=${page}&limit=${limit}`,
    )
    return res.data
}

export async function submitReview(
    bookId: string,
    payload: { rating: number; body?: string; is_public?: boolean },
): Promise<Review> {
    const res = await apiClient.post<Review>(`/books/${bookId}/reviews`, payload)
    return res.data
}

export async function updateMyReview(
    bookId: string,
    payload: { rating: number; body?: string; is_public?: boolean },
): Promise<Review> {
    const res = await apiClient.put<Review>(`/books/${bookId}/reviews/me`, payload)
    return res.data
}

// Add a copy of an existing edition to the user's library
export async function addCopy(
    editionId: string,
    opts?: { condition?: string }
): Promise<void> {
    await apiClient.post('/books/copies', { edition_id: editionId, ...opts })
}

// Update book fields — moderator/admin only (PUT /books/{id})
export async function updateBook(id: string, data: UpdateBookPayload): Promise<Book> {
    const res = await apiClient.put<Book>(`/books/${id}`, data)
    return res.data
}

// PLACEHOLDER — wishlist backend not yet implemented
export async function addToWishlist(_bookId: string): Promise<void> {
    alert('Wishlist feature coming soon!')
}

export async function removeFromWishlist(_bookId: string): Promise<void> {
    alert('Wishlist feature coming soon!')
}