export type Role = 'user' | 'moderator' | 'admin'

export interface User {
    id: string
    email: string
    username: string
    role: Role
    theme: string
    created_at: string
    updated_at: string
    bio?: string
    avatar_url?: string
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

export interface LookupResultsPage {
    results: LookupResult[]
    total: number
    page: number
    page_size: number
}

export interface LookupFilters {
    language?: string
    publisher?: string
    author?: string
    yearFrom?: number
    yearTo?: number
    format?: string   // for future use — APIs don't return format, but we can filter by it post-fetch
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

// ── Replace the existing Edition interface with this ────────────────────────
export interface Edition {
    id: string
    book_id: string
    format: string            // hardcover | paperback | ebook | audiobook
    isbn?: string
    asin?: string
    language?: string
    publisher?: string
    edition?: string          // "2nd Edition", "Anniversary Edition" etc
    published_at?: string
    page_count?: number
    file_format?: string      // EPUB | PDF | MOBI | AZW3
    duration_minutes?: number
    audio_format?: string     // MP3 | AAC | WMA | FLAC
    status: string
    translators?: Translator[]
    narrators?: Narrator[]
    created_at: string
    updated_at: string
}

// ── These already exist in book.go — ensure they exist in types/index.ts ────
export interface Narrator {
    id: string
    name: string
    status: string
    created_at: string
    updated_at: string
}

export interface Translator {
    id: string
    name: string
    status: string
    created_at: string
    updated_at: string
}

// ── Add Review interface ─────────────────────────────────────────────────────
export interface Review {
    id: string
    book_id: string
    user_id?: string          // nullable after migration 000016 (anonymisation)
    username?: string         // joined from users table in the API response
    avatar_url?: string
    rating: number            // 1–5
    body?: string
    is_public: boolean
    created_at: string
    updated_at: string
}

export interface ReviewsResponse {
    reviews: Review[]
    total: number
    page: number
    limit: number
    average_rating?: number   // computed by the backend
}