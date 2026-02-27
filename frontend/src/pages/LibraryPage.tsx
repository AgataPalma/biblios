import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getMyLibrary, updateReadingStatus, removeCopy } from '../api/books'
import { Card, Badge, Spinner, EmptyState, Modal } from '../components'
import type { UserBook } from '../types'

const STATUS_LABELS: Record<string, string> = {
    want_to_read: 'Want to Read',
    reading: 'Reading',
    read: 'Read',
}

const STATUS_VARIANTS: Record<string, 'default' | 'info' | 'success'> = {
    want_to_read: 'default',
    reading: 'info',
    read: 'success',
}

export default function LibraryPage() {
    const navigate = useNavigate()
    const queryClient = useQueryClient()
    const [page, setPage] = useState(1)
    const [filter, setFilter] = useState<string>('all')
    const [removeTarget, setRemoveTarget] = useState<UserBook | null>(null)

    const { data, isLoading } = useQuery({
        queryKey: ['my-library', page],
        queryFn: () => getMyLibrary(page, 20),
    })

    const statusMutation = useMutation({
        mutationFn: ({ copyId, status }: { copyId: string; status: string }) =>
            updateReadingStatus(copyId, status),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['my-library'] }),
    })

    const removeMutation = useMutation({
        mutationFn: (copyId: string) => removeCopy(copyId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['my-library'] })
            queryClient.invalidateQueries({ queryKey: ['my-books'] })
            setRemoveTarget(null)
        },
    })

    const books: UserBook[] = data?.books ?? []
    const filtered = filter === 'all' ? books : books.filter(b => b.reading_status === filter)

    const counts = books.reduce((acc, b) => {
        acc[b.reading_status] = (acc[b.reading_status] ?? 0) + 1
        return acc
    }, {} as Record<string, number>)

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
                        My Library
                    </h1>
                    <p style={{
                        margin: '4px 0 0',
                        fontSize: '13px',
                        color: 'var(--color-text-muted)',
                        fontFamily: 'var(--font-body)',
                    }}>
                        {data?.total ?? 0} book{data?.total !== 1 ? 's' : ''} in your collection
                    </p>
                </div>
                <button
                    onClick={() => navigate('/books')}
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
                        fontFamily: 'var(--font-body)',
                    }}
                >
                    ➕ Add books
                </button>
            </div>

            {/* Filter tabs */}
            <div style={{
                display: 'flex',
                gap: '4px',
                marginBottom: '24px',
                background: 'var(--color-surface-alt)',
                padding: '4px',
                borderRadius: 'var(--border-radius)',
                width: 'fit-content',
            }}>
                {[
                    { key: 'all', label: `All (${books.length})` },
                    { key: 'want_to_read', label: `Want to Read (${counts.want_to_read ?? 0})` },
                    { key: 'reading', label: `Reading (${counts.reading ?? 0})` },
                    { key: 'read', label: `Read (${counts.read ?? 0})` },
                ].map(tab => (
                    <button
                        key={tab.key}
                        onClick={() => setFilter(tab.key)}
                        style={{
                            padding: '6px 14px',
                            border: 'none',
                            borderRadius: 'var(--border-radius)',
                            background: filter === tab.key ? 'var(--color-surface)' : 'transparent',
                            color: filter === tab.key ? 'var(--color-text)' : 'var(--color-text-muted)',
                            fontSize: '12px',
                            fontWeight: filter === tab.key ? 600 : 400,
                            cursor: 'pointer',
                            transition: 'var(--transition)',
                            boxShadow: filter === tab.key ? 'var(--shadow-sm)' : 'none',
                            fontFamily: 'var(--font-body)',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        {tab.label}
                    </button>
                ))}
            </div>

            {/* Content */}
            {isLoading ? (
                <div style={{ display: 'flex', justifyContent: 'center', paddingTop: '80px' }}>
                    <Spinner size="lg" label="Loading your library..." />
                </div>
            ) : filtered.length === 0 ? (
                <Card padding="lg">
                    <EmptyState
                        icon="📚"
                        title={filter === 'all' ? 'Your library is empty' : `No books with status "${STATUS_LABELS[filter]}"`}
                        description={filter === 'all' ? 'Browse the catalogue and add books to your collection.' : 'Change the filter to see other books.'}
                        action={filter === 'all' ? { label: 'Browse catalogue', onClick: () => navigate('/books') } : undefined}
                    />
                </Card>
            ) : (
                <div style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
                    gap: '16px',
                }}>
                    {filtered.map(userBook => (
                        <LibraryBookCard
                            key={userBook.copy_id}
                            userBook={userBook}
                            onStatusChange={(status) => statusMutation.mutate({
                                copyId: userBook.copy_id,
                                status,
                            })}
                            onNavigate={() => navigate(`/books/${userBook.book.id}`)}
                            onRemove={() => setRemoveTarget(userBook)}
                            isUpdating={statusMutation.isPending}
                        />
                    ))}
                </div>
            )}

            {/* Confirm remove modal */}
            <Modal
                isOpen={!!removeTarget}
                onClose={() => setRemoveTarget(null)}
                title="Remove from library"
                confirmLabel="Remove"
                confirmVariant="danger"
                onConfirm={() => removeTarget && removeMutation.mutate(removeTarget.copy_id)}
                isLoading={removeMutation.isPending}
                size="sm"
            >
                <p style={{
                    margin: 0,
                    fontSize: '14px',
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-body)',
                    lineHeight: '1.6',
                }}>
                    Remove <strong>{removeTarget?.book.title}</strong> from your library? This cannot be undone.
                </p>
            </Modal>
        </div>
    )
}

// ─── Library Book Card ────────────────────────────────────────────────────────

function LibraryBookCard({
                             userBook,
                             onStatusChange,
                             onNavigate,
                             onRemove,
                             isUpdating,
                         }: {
    userBook: UserBook
    onStatusChange: (status: string) => void
    onNavigate: () => void
    onRemove: () => void
    isUpdating: boolean
}) {
    const { book } = userBook

    const colors = [
        '#2563eb', '#16a34a', '#dc2626', '#9333ea',
        '#ea580c', '#0891b2', '#d97706', '#4f46e5',
    ]
    const colorIndex = book.title
        .split('')
        .reduce((a, c) => a + c.charCodeAt(0), 0) % colors.length
    const coverColor = colors[colorIndex]

    return (
        <Card padding="sm" hover>
            <div style={{ display: 'flex', gap: '12px' }}>
                {/* Cover */}
                <div
                    onClick={onNavigate}
                    style={{
                        width: '56px',
                        height: '76px',
                        borderRadius: '4px',
                        flexShrink: 0,
                        background: book.cover_url
                            ? `url(${book.cover_url}) center/cover no-repeat`
                            : `linear-gradient(135deg, ${coverColor}dd, ${coverColor}88)`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: '22px',
                        border: '1px solid var(--color-border)',
                        cursor: 'pointer',
                    }}
                >
                    {!book.cover_url && '📖'}
                </div>

                {/* Info */}
                <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: '6px' }}>
                    <p
                        onClick={onNavigate}
                        style={{
                            margin: 0,
                            fontSize: '13px',
                            fontWeight: 600,
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-body)',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                            cursor: 'pointer',
                        }}
                    >
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
                        {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                    </p>

                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <Badge
                            label={STATUS_LABELS[userBook.reading_status]}
                            variant={STATUS_VARIANTS[userBook.reading_status]}
                            size="sm"
                        />
                        {userBook.format && (
                            <Badge label={userBook.format} variant="muted" size="sm" />
                        )}
                    </div>
                </div>
            </div>

            {/* Status selector */}
            <div style={{
                marginTop: '10px',
                paddingTop: '10px',
                borderTop: '1px solid var(--color-border)',
                display: 'flex',
                gap: '4px',
                justifyContent: 'space-between',
                alignItems: 'center',
            }}>
                <div style={{ display: 'flex', gap: '4px' }}>
                    {(['want_to_read', 'reading', 'read'] as const).map(status => (
                        <button
                            key={status}
                            disabled={isUpdating}
                            onClick={() => onStatusChange(status)}
                            style={{
                                padding: '4px 8px',
                                border: `1px solid ${userBook.reading_status === status ? 'var(--color-primary)' : 'var(--color-border)'}`,
                                borderRadius: 'var(--border-radius)',
                                background: userBook.reading_status === status ? 'var(--color-primary)' : 'transparent',
                                color: userBook.reading_status === status ? 'var(--color-primary-text)' : 'var(--color-text-muted)',
                                fontSize: '10px',
                                cursor: isUpdating ? 'not-allowed' : 'pointer',
                                fontFamily: 'var(--font-body)',
                                fontWeight: userBook.reading_status === status ? 600 : 400,
                                transition: 'var(--transition)',
                                whiteSpace: 'nowrap',
                            }}
                        >
                            {status === 'want_to_read' ? '🔖' : status === 'reading' ? '📖' : '✅'}
                        </button>
                    ))}
                </div>

                <button
                    onClick={onRemove}
                    style={{
                        background: 'none',
                        border: 'none',
                        cursor: 'pointer',
                        color: 'var(--color-error)',
                        fontSize: '14px',
                        padding: '4px',
                        opacity: 0.6,
                        transition: 'var(--transition)',
                    }}
                    onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.opacity = '1'}
                    onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.opacity = '0.6'}
                    title="Remove from library"
                >
                    🗑️
                </button>
            </div>
        </Card>
    )
}