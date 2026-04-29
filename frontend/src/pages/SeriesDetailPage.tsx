import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getSeriesById } from '../api/series'
import type { SeriesBook } from '../types'
import { Spinner, EmptyState, Badge } from '../components'

// ── Reading status helpers ────────────────────────────────────────────────────

function statusLabel(status: string): string {
    if (status === 'reading') return 'Reading'
    if (status === 'read') return 'Read'
    if (status === 'want_to_read') return 'Want to Read'
    return status
}

function statusVariant(status: string): 'default' | 'info' | 'success' {
    if (status === 'reading') return 'info'
    if (status === 'read') return 'success'
    return 'default'
}

// ── Book row ──────────────────────────────────────────────────────────────────

function BookRow({ seriesBook }: { seriesBook: SeriesBook }) {
    const { book, position, status } = seriesBook
    const authorNames = book.authors?.map(a => a.name).join(', ') ?? ''

    // Find a cover URL from the first edition that has one
    const coverUrl = book.editions?.find(e => e.cover_url)?.cover_url

    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                gap: '16px',
                padding: '12px 0',
                borderBottom: '1px solid var(--color-border)',
            }}
        >
            {/* Position number */}
            <div
                style={{
                    width: '32px',
                    flexShrink: 0,
                    fontSize: '18px',
                    fontWeight: 700,
                    color: 'var(--color-text-muted)',
                    textAlign: 'center',
                }}
            >
                {position ?? '—'}
            </div>

            {/* Cover thumbnail */}
            <div
                style={{
                    width: '40px',
                    height: '60px',
                    flexShrink: 0,
                    borderRadius: '4px',
                    overflow: 'hidden',
                    background: 'var(--color-surface-alt)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '20px',
                }}
            >
                {coverUrl ? (
                    <img
                        src={coverUrl}
                        alt={book.title}
                        style={{
                            width: '100%',
                            height: '100%',
                            objectFit: 'cover',
                            display: 'block',
                        }}
                    />
                ) : (
                    <span>📖</span>
                )}
            </div>

            {/* Title + author */}
            <div style={{ flex: 1, minWidth: 0 }}>
                <div
                    style={{
                        fontWeight: 700,
                        fontSize: '14px',
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                    }}
                >
                    {book.title}
                </div>
                {authorNames && (
                    <div
                        style={{
                            fontSize: '12px',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                            marginTop: '2px',
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                        }}
                    >
                        {authorNames}
                    </div>
                )}
            </div>

            {/* Reading status badge */}
            {status && (
                <div style={{ flexShrink: 0 }}>
                    <Badge
                        label={statusLabel(status)}
                        variant={statusVariant(status)}
                        size="sm"
                    />
                </div>
            )}
        </div>
    )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function SeriesDetailPage() {
    const { id } = useParams<{ id: string }>()
    const navigate = useNavigate()

    const { data, isLoading, isError } = useQuery({
        queryKey: ['series', id],
        queryFn: () => getSeriesById(id!),
        enabled: !!id,
    })

    // Sort books by position ascending, nulls last
    const sortedBooks = data?.books
        ? [...data.books].sort((a, b) => {
              if (a.position == null && b.position == null) return 0
              if (a.position == null) return 1
              if (b.position == null) return -1
              return a.position - b.position
          })
        : []

    return (
        <div
            style={{
                maxWidth: '900px',
                margin: '0 auto',
                padding: '32px',
                fontFamily: 'var(--font-body)',
                color: 'var(--color-text)',
            }}
        >
            {/* Back button */}
            <button
                onClick={() => navigate('/series')}
                style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    color: 'var(--color-text-muted)',
                    fontSize: '14px',
                    fontFamily: 'var(--font-body)',
                    padding: '0 0 20px 0',
                    display: 'flex',
                    alignItems: 'center',
                    gap: '4px',
                    transition: 'var(--transition)',
                }}
            >
                ← Back
            </button>

            {/* Loading */}
            {isLoading && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '48px 0' }}>
                    <Spinner />
                </div>
            )}

            {/* Error */}
            {isError && (
                <EmptyState icon="⚠️" title="Failed to load series" />
            )}

            {/* Content */}
            {!isLoading && !isError && data && (
                <>
                    {/* Page header */}
                    <div style={{ marginBottom: '28px' }}>
                        <h1
                            style={{
                                margin: '0 0 6px 0',
                                fontSize: '24px',
                                fontWeight: 700,
                                fontFamily: 'var(--font-heading)',
                                color: 'var(--color-text)',
                            }}
                        >
                            {data.series.name}
                        </h1>
                        {data.series.description && (
                            <p
                                style={{
                                    margin: 0,
                                    fontSize: '13px',
                                    color: 'var(--color-text-muted)',
                                    fontFamily: 'var(--font-body)',
                                    lineHeight: 1.5,
                                }}
                            >
                                {data.series.description}
                            </p>
                        )}
                    </div>

                    {/* Empty state */}
                    {sortedBooks.length === 0 ? (
                        <EmptyState icon="📚" title="No books in this series" />
                    ) : (
                        <div>
                            {sortedBooks.map((seriesBook: SeriesBook) => (
                                <BookRow
                                    key={seriesBook.book_id}
                                    seriesBook={seriesBook}
                                />
                            ))}
                        </div>
                    )}
                </>
            )}
        </div>
    )
}
