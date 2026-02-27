export type Role = 'user' | 'moderator' | 'admin'

export interface User {
    id: string
    email: string
    username: string
    role: Role
    theme: string
    created_at: string
    updated_at: string
}

export interface Author {
    id: string
    name: string
    status: string
    created_at: string
    updated_at: string
}

export interface Genre {
    id: string
    name: string
    status: string
    created_at: string
}

export interface Edition {
    id: string
    book_id: string
    format: string
    isbn?: string
    language?: string
    publisher?: string
    published_at?: string
    page_count?: number
    status: string
    created_at: string
    updated_at: string
}

export interface Book {
    id: string
    title: string
    description?: string
    cover_url?: string
    status: string
    authors: Author[]
    genres: Genre[]
    editions: Edition[]
    created_at: string
    updated_at: string
}


export interface PaginatedResponse<T> {
    total: number
    page: number
    limit: number
    items: T[]
}

export interface LoginRequest {
    email: string
    password: string
}

export interface RegisterRequest {
    email: string
    username: string
    password: string
}

export interface AuthResponse {
    token: string
    user: User
}

export interface LookupResult {
    title: string
    authors: string[]
    isbn?: string      // we'll use isbn_13 falling back to isbn_10
    isbn_10?: string
    isbn_13?: string
    publisher?: string
    published_date?: string
    page_count?: number
    cover_url?: string
    language?: string
    description?: string
    categories?: string[]
}

export interface UserBook {
    copy_id: string
    reading_status: 'want_to_read' | 'reading' | 'read'
    condition?: string
    added_at: string
    edition_id: string
    format: string
    language?: string
    book: Book
}

export interface Submission {
    id: string
    submitted_by: string
    status: string
    rejection_reason?: string
    reviewed_by?: string
    reviewed_at?: string
    book_id: string
    edition_id: string
    copy_id?: string
    created_at: string
    updated_at: string
}