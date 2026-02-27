import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { listBooks } from '../api/books'
import { Card, Badge, Spinner, EmptyState } from '../components'
import type { Book } from '../types'

export default function BooksPage() {
    const navigate = useNavigate()
    const [page, setPage] = useState<number>(1)
    const limit = 20

    const { data, isLoading, isError } = useQuery({
        queryKey: ['books', page],
        queryFn: () => listBooks(page, limit),
        placeholderData: prev => prev,
    })

    const totalPages = data ? Math.ceil(data.total / limit) : 1

    return (
        <div style={{
            minHeight: 'calc(100vh - 56px)',
            padding: '32px 24px',
            maxWidth: '1100px',
            margin: '0 auto',
        }}>
            {/* Header */}
            <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: '24px',
                flexWrap: 'wrap',
                gap: '12px',
            }}>
                <div>
                    <h1 style={{
                        margin: 0,
                        fontSize: '24px',
                        fontWeight: 700,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                    }}>
                        Book Catalogue
                    </h1>
                    <p style={{
                        margin: '4px 0 0',
                        fontSize: '13px',
                        color: 'var(--color-text-muted)',
                        fontFamily: 'var(--font-body)',
                    }}>
                        {data ? `${data.total} approved book${data.total !== 1 ? 's' : ''}` : ''}
                    </p>
                </div>

                <button
                    onClick={() => navigate('/books/add')}
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                        padding: '10px 18px',
                        background: 'var(--color-primary)',
                        color: 'var(--color-primary-text)',
                        border: 'none',
                        borderRadius: 'var(--border-radius)',
                        fontSize: '13px',
                        fontWeight: 600,
                        cursor: 'pointer',
                        transition: 'var(--transition)',
                        fontFamily: 'var(--font-body)',
                    }}
                    onMouseEnter={e => {
                        (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary-hover)'
                    }}
                    onMouseLeave={e => {
                        (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary)'
                    }}
                >
                    <span>➕</span>
                    <span>Add a book</span>
                </button>
            </div>

            {/* Content */}
            {isLoading ? (
                <div style={{
                    display: 'flex',
                    justifyContent: 'center',
                    paddingTop: '80px',
                }}>
                    <Spinner size="lg" label="Loading catalogue..." />
                </div>
            ) : isError ? (
                <Card padding="lg">
                    <EmptyState
                        icon="⚠️"
                        title="Failed to load books"
                        description="Something went wrong. Please try again."
                        action={{ label: 'Retry', onClick: () => window.location.reload() }}
                    />
                </Card>
            ) : !data?.books?.length ? (
                <Card padding="lg">
                    <EmptyState
                        icon="📭"
                        title="No books yet"
                        description="The catalogue is empty. Be the first to add a book!"
                        action={{ label: 'Add a book', onClick: () => navigate('/books/add') }}
                    />
                </Card>
            ) : (
                <>
                    {/* Book grid */}
                    <div style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
                        gap: '16px',
                        marginBottom: '32px',
                    }}>
                        {data.books.map(book => (
                            <BookCard
                                key={book.id}
                                book={book}
                                onClick={() => navigate(`/books/${book.id}`)}
                            />
                        ))}
                    </div>

                    {/* Pagination */}
                    {totalPages > 1 && (
                        <Pagination
                            page={page}
                            totalPages={totalPages}
                            onPageChange={setPage}
                        />
                    )}
                </>
            )}
        </div>
    )
}

// ─── Book Card ────────────────────────────────────────────────────────────────

function BookCard({ book, onClick }: { book: Book; onClick: () => void }) {
    // Generate a deterministic colour from the book title for the cover
    const colors = [
        '#2563eb', '#16a34a', '#dc2626', '#9333ea',
        '#ea580c', '#0891b2', '#d97706', '#4f46e5',
        '#0d9488', '#be185d',
    ]
    const colorIndex = book.title
        .split('')
        .reduce((a, c) => a + c.charCodeAt(0), 0) % colors.length
    const coverColor = colors[colorIndex]

    return (
        <div
            onClick={onClick}
            style={{
                background: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--border-radius)',
                overflow: 'hidden',
                cursor: 'pointer',
                transition: 'var(--transition)',
                boxShadow: 'var(--shadow-sm)',
                display: 'flex',
                flexDirection: 'column',
            }}
            onMouseEnter={e => {
                (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-md)'
                ;(e.currentTarget as HTMLDivElement).style.transform = 'translateY(-3px)'
            }}
            onMouseLeave={e => {
                (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-sm)'
                ;(e.currentTarget as HTMLDivElement).style.transform = 'translateY(0)'
            }}
        >
            {/* Cover */}
            <div style={{
                height: '160px',
                background: book.cover_url
                    ? `url(${book.cover_url}) center/cover no-repeat`
                    : `linear-gradient(135deg, ${coverColor}dd, ${coverColor}88)`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
                position: 'relative',
            }}>
                {!book.cover_url && (
                    <span style={{ fontSize: '40px', opacity: 0.6 }}>📖</span>
                )}
                {/* Genre badges */}
                {book.genres && book.genres.length > 0 && (
                    <div style={{
                        position: 'absolute',
                        bottom: '8px',
                        left: '8px',
                        display: 'flex',
                        gap: '4px',
                        flexWrap: 'wrap',
                    }}>
                        <Badge label={book.genres[0].name} variant="info" size="sm" />
                    </div>
                )}
            </div>

            {/* Info */}
            <div style={{
                padding: '12px',
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                gap: '4px',
            }}>
                <p style={{
                    margin: 0,
                    fontSize: '13px',
                    fontWeight: 600,
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-body)',
                    overflow: 'hidden',
                    display: '-webkit-box',
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: 'vertical',
                    lineHeight: '1.4',
                }}>
                    {book.title}
                </p>
                <p style={{
                    margin: 0,
                    fontSize: '11px',
                    color: 'var(--color-text-muted)',
                    fontFamily: 'var(--font-body)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                }}>
                    {book.authors && book.authors.length > 0
                        ? book.authors.map(a => a.name).join(', ')
                        : 'Unknown author'}
                </p>
            </div>
        </div>
    )
}

// ─── Pagination ───────────────────────────────────────────────────────────────

function Pagination({
                        page,
                        totalPages,
                        onPageChange,
                    }: {
    page: number
    totalPages: number
    onPageChange: (p: number) => void
}) {
    const pages: (number | '...')[] = []

    if (totalPages <= 7) {
        for (let i = 1; i <= totalPages; i++) pages.push(i)
    } else {
        pages.push(1)
        if (page > 3) pages.push('...')
        for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) {
            pages.push(i)
        }
        if (page < totalPages - 2) pages.push('...')
        pages.push(totalPages)
    }

    return (
        <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '6px',
        }}>
            <PageBtn
                label="←"
                onClick={() => onPageChange(page - 1)}
                disabled={page === 1}
            />
            {pages.map((p, i) =>
                    p === '...' ? (
                        <span
                            key={`ellipsis-${i}`}
                            style={{ color: 'var(--color-text-muted)', padding: '0 4px' }}
                        >
            …
          </span>
                    ) : (
                        <PageBtn
                            key={p}
                            label={String(p)}
                            onClick={() => onPageChange(p as number)}
                            active={p === page}
                        />
                    )
            )}
            <PageBtn
                label="→"
                onClick={() => onPageChange(page + 1)}
                disabled={page === totalPages}
            />
        </div>
    )
}

function PageBtn({
                     label,
                     onClick,
                     active = false,
                     disabled = false,
                 }: {
    label: string
    onClick: () => void
    active?: boolean
    disabled?: boolean
}) {
    return (
        <button
            onClick={onClick}
            disabled={disabled}
            style={{
                minWidth: '36px',
                height: '36px',
                padding: '0 10px',
                background: active ? 'var(--color-primary)' : 'var(--color-surface)',
                color: active ? 'var(--color-primary-text)' : 'var(--color-text)',
                border: `1px solid ${active ? 'var(--color-primary)' : 'var(--color-border)'}`,
                borderRadius: 'var(--border-radius)',
                fontSize: '13px',
                fontWeight: active ? 600 : 400,
                cursor: disabled ? 'not-allowed' : 'pointer',
                opacity: disabled ? 0.4 : 1,
                transition: 'var(--transition)',
                fontFamily: 'var(--font-body)',
            }}
        >
            {label}
        </button>
    )
}