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

export interface Book {
    id: string
    title: string
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
    current_page?: number
    started_reading_at?: string
    finished_reading_at?: string
    owned_by_user: boolean
    borrowed_from?: string   // user ID of the real owner if borrowed
    location?: string
    condition?: string
    added_at: string
    edition_id: string
    format: string
    language?: string
    cover_url?: string       // from book_editions
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

export interface Edition {
    id: string
    book_id: string
    format: string            // hardcover | paperback | ebook | audiobook
    description?: string
    cover_url?: string
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

// ── Notification ──────────────────────────────────────────────────────────────
export interface Notification {
    id: string
    user_id: string
    type: 'invitation' | 'review_like' | 'library_activity' | string
    title: string
    body: string
    data?: Record<string, unknown>
    is_read: boolean
    created_at: string
}

export interface NotificationsResponse {
    notifications: Notification[]
    total: number
    unread_count: number
}

// ── Shelf ─────────────────────────────────────────────────────────────────────
export interface Shelf {
    id: string
    user_id: string
    name: string
    created_at: string
    updated_at: string
    book_count?: number
}

export interface ShelfBook {
    copy_id: string
    shelf_id: string
    added_at: string
    book: Book
    edition_id: string
    format: string
    cover_url?: string
    reading_status: string
}

export interface ShelfDetailResponse {
    shelf: Shelf
    books: ShelfBook[]
    total: number
}

// ── Collection ────────────────────────────────────────────────────────────────
export interface Collection {
    id: string
    library_id: string
    name: string
    description?: string
    visibility: 'public' | 'private'
    is_collaborative: boolean
    created_by: string
    created_at: string
    updated_at: string
    book_count?: number
}

export interface CollectionBook {
    book_id: string
    collection_id: string
    added_by: string
    added_at: string
    book: Book
}

export interface CollectionsResponse {
    collections: Collection[]
    total: number
}

export interface CollectionDetailResponse {
    collection: Collection
    books: CollectionBook[]
    total: number
}

// ── Cooperative Library ───────────────────────────────────────────────────────
export interface CooperativeLibrary {
    id: string
    owner_id: string
    name: string
    description?: string
    visibility: 'private' | 'semi_public' | 'public'
    is_cooperative: boolean
    created_at: string
    updated_at: string
    member_count?: number
    book_count?: number
}

export interface LibraryMember {
    user_id: string
    library_id: string
    username: string
    avatar_url?: string
    role: 'owner' | 'member'
    can_view: boolean
    can_add: boolean
    can_remove: boolean
    can_edit: boolean
    can_invite: boolean
    can_manage_members: boolean
    joined_at: string
}

export interface LibrariesResponse {
    libraries: CooperativeLibrary[]
    total: number
}

export interface LibraryDetailResponse {
    library: CooperativeLibrary
    members: LibraryMember[]
}

// ── Series ────────────────────────────────────────────────────────────────────
export interface Series {
    id: string
    name: string
    description?: string
    status: 'pending' | 'approved'
    created_at: string
    updated_at: string
    book_count?: number
}

export interface SeriesBook {
    book_id: string
    series_id: string
    position?: number
    status: string
    book: Book
}

export interface SeriesResponse {
    series: Series[]
    total: number
    page: number
    limit: number
}

export interface SeriesDetailResponse {
    series: Series
    books: SeriesBook[]
}

// ── Reading ───────────────────────────────────────────────────────────────────
export interface ReadingChallenge {
    id: string
    user_id: string
    title: string
    start_date: string
    end_date: string
    goal_books: number
    current_books: number
    status: 'active' | 'completed' | 'failed'
    created_at: string
    updated_at: string
}

export interface ReadingSession {
    id: string
    user_id: string
    copy_id: string
    date: string
    pages_read: number
    notes?: string
    created_at: string
    book_title?: string
}

export interface ChallengesResponse {
    challenges: ReadingChallenge[]
    total: number
}

export interface SessionsResponse {
    sessions: ReadingSession[]
    total: number
}

// ── Book List ─────────────────────────────────────────────────────────────────
export interface BookList {
    id: string
    user_id: string
    title: string
    description?: string
    tags?: string[]
    visibility: 'private' | 'public'
    created_at: string
    updated_at: string
    book_count?: number
}

export interface BookListItem {
    book_id: string
    list_id: string
    commentary?: string
    sort_order: number
    added_at: string
    book: Book
}

export interface ListsResponse {
    lists: BookList[]
    total: number
}

export interface ListDetailResponse {
    list: BookList
    books: BookListItem[]
    total: number
}
