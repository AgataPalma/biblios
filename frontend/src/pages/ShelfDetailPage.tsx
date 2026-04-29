import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getShelf, removeBookFromShelf } from '../api/shelves'
import type { ShelfBook } from '../types'
import { Spinner, EmptyState, Badge, Button } from '../components'

// ── Reading status helpers ────────────────────────────────────────────────────

type ReadingStatus = 'want_to_read' | 'reading' | 'read' | string

function readingStatusLabel(status: ReadingStatus): string {
    if (status === 'reading') return 'Reading'
    if (status === 'read') return 'Read'
    return 'Want to Read'
}

function readingStatusVariant(status: ReadingStatus): 'default' | 'info' | 'success' {
    if (status === 'reading') return 'info'
    if (status === 'read') return 'success'
    return 'default'
}

// ── Book row ──────────────────────────────────────────────────────────────────

interface BookRowProps {
    book: ShelfBook
    shelfId: string
    onRemove: (copyId: string) => void
    isRemoving: boolean
}

function BookRow({ book, shelfId: _shelfId, onRemove, isRemoving }: BookRowProps) {
    const [confirming, setConfirming] = useState(false)
    const [hovered, setHovered] = useState(false)

    const authorNames = book.book.authors?.map(a => a.name).join(', ') ?? ''

    return (
        <div
            onMouseEnter={() => setHovered(true)}
            onMouseLeave={() => setHovered(false)}
            style={{
                display: 'flex',
                alignItems: 'center',
                gap: '16px',
                padding: '12px 0',
                borderBottom: '1px solid var(--color-border)',
                background: hovered ? 'var(--color-surface-alt)' : 'transparent',
                transition: 'background 0.15s ease',
                borderRadius: '4px',
            }}
        >
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
                {book.cover_url ? (
                    <img
                        src={book.cover_url}
                        alt={book.book.title}
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
                    {book.book.title}
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

            {/* Format badge */}
            <div style={{ flexShrink: 0 }}>
                <Badge label={book.format} variant="default" size="sm" />
            </div>

            {/* Reading status badge */}
            <div style={{ flexShrink: 0 }}>
                <Badge
                    label={readingStatusLabel(book.reading_status)}
                    variant={readingStatusVariant(book.reading_status)}
                    size="sm"
                />
            </div>

            {/* Remove button / confirm */}
            <div style={{ flexShrink: 0, display: 'flex', alignItems: 'center', gap: '6px' }}>
                {confirming ? (
                    <>
                        <span
                            style={{
                                fontSize: '12px',
                                color: 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            Confirm?
                        </span>
                        <button
                            title="Confirm remove"
                            disabled={isRemoving}
                            onClick={() => {
                                onRemove(book.copy_id)
                                setConfirming(false)
                            }}
                            style={{
                                width: '28px',
                                height: '28px',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                background: 'rgba(34,197,94,0.15)',
                                border: '1px solid rgba(34,197,94,0.3)',
                                borderRadius: 'var(--border-radius)',
                                cursor: isRemoving ? 'not-allowed' : 'pointer',
                                fontSize: '14px',
                                color: 'var(--color-success)',
                                transition: 'var(--transition)',
                                padding: 0,
                                opacity: isRemoving ? 0.5 : 1,
                            }}
                        >
                            ✓
                        </button>
                        <button
                            title="Cancel"
                            onClick={() => setConfirming(false)}
                            style={{
                                width: '28px',
                                height: '28px',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                background: 'transparent',
                                border: 'none',
                                borderRadius: 'var(--border-radius)',
                                cursor: 'pointer',
                                fontSize: '12px',
                                color: 'var(--color-text-muted)',
                                transition: 'var(--transition)',
                                padding: 0,
                            }}
                        >
                            ✕
                        </button>
                    </>
                ) : (
                    <button
                        title="Remove from shelf"
                        onClick={() => setConfirming(true)}
                        style={{
                            width: '28px',
                            height: '28px',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            background: 'transparent',
                            border: 'none',
                            borderRadius: 'var(--border-radius)',
                            cursor: 'pointer',
                            fontSize: '14px',
                            color: 'var(--color-error)',
                            transition: 'var(--transition)',
                            padding: 0,
                            opacity: hovered ? 1 : 0.4,
                        }}
                    >
                        🗑️
                    </button>
                )}
            </div>
        </div>
    )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function ShelfDetailPage() {
    const { id } = useParams<{ id: string }>()
    const navigate = useNavigate()
    const queryClient = useQueryClient()

    // ── Data fetching ─────────────────────────────────────────────────────────
    const { data, isLoading, isError } = useQuery({
        queryKey: ['shelf', id],
        queryFn: () => getShelf(id!),
        enabled: !!id,
    })

    // ── Remove book mutation ──────────────────────────────────────────────────
    const removeMutation = useMutation({
        mutationFn: (copyId: string) => removeBookFromShelf(id!, copyId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['shelf', id] })
        },
    })

    // ── Render ────────────────────────────────────────────────────────────────
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
                onClick={() => navigate('/shelves')}
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
                <EmptyState icon="⚠️" title="Failed to load shelf" />
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
                            {data.shelf.name}
                        </h1>
                        <p
                            style={{
                                margin: 0,
                                fontSize: '13px',
                                color: 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            {data.total} {data.total === 1 ? 'book' : 'books'}
                        </p>
                    </div>

                    {/* Empty state */}
                    {data.books.length === 0 ? (
                        <EmptyState
                            icon="📚"
                            title="This shelf is empty"
                            description="Add books from your library"
                        />
                    ) : (
                        /* Book list */
                        <div>
                            {data.books.map((book: ShelfBook) => (
                                <BookRow
                                    key={book.copy_id}
                                    book={book}
                                    shelfId={id!}
                                    onRemove={(copyId) => removeMutation.mutate(copyId)}
                                    isRemoving={removeMutation.isPending}
                                />
                            ))}
                        </div>
                    )}
                </>
            )}
        </div>
    )
}
