import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getList, updateList, deleteList, removeBookFromList } from '../api/lists'
import type { BookListItem } from '../types'
import { Button, Input, Modal, Badge, Spinner, EmptyState } from '../components'

// ── Book card ─────────────────────────────────────────────────────────────────

interface BookCardProps {
    item: BookListItem
    onRemove: (bookId: string) => void
    isRemoving: boolean
}

function BookCard({ item, onRemove, isRemoving }: BookCardProps) {
    const [confirming, setConfirming] = useState(false)
    const coverUrl = item.book.editions?.find(e => e.cover_url)?.cover_url
    const authorNames = item.book.authors?.map(a => a.name).join(', ') ?? ''

    return (
        <div
            style={{
                background: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--border-radius)',
                overflow: 'hidden',
                display: 'flex',
                flexDirection: 'column',
            }}
        >
            {/* Cover */}
            <div
                style={{
                    width: '100%',
                    aspectRatio: '2/3',
                    background: 'var(--color-surface-alt)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '32px',
                    overflow: 'hidden',
                    flexShrink: 0,
                }}
            >
                {coverUrl ? (
                    <img
                        src={coverUrl}
                        alt={item.book.title}
                        style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }}
                    />
                ) : (
                    <span>📖</span>
                )}
            </div>

            {/* Info */}
            <div style={{ padding: '10px', display: 'flex', flexDirection: 'column', gap: '4px', flex: 1 }}>
                <p
                    style={{
                        margin: 0,
                        fontSize: '13px',
                        fontWeight: 700,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        whiteSpace: 'nowrap',
                    }}
                >
                    {item.book.title}
                </p>
                {authorNames && (
                    <p
                        style={{
                            margin: 0,
                            fontSize: '11px',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        {authorNames}
                    </p>
                )}
                {item.commentary && (
                    <p
                        style={{
                            margin: '4px 0 0',
                            fontSize: '11px',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                            fontStyle: 'italic',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        "{item.commentary}"
                    </p>
                )}

                {/* Remove button */}
                <div style={{ marginTop: 'auto', paddingTop: '8px' }}>
                    {confirming ? (
                        <div style={{ display: 'flex', gap: '6px', alignItems: 'center' }}>
                            <span style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>Remove?</span>
                            <button
                                disabled={isRemoving}
                                onClick={() => { onRemove(item.book_id); setConfirming(false) }}
                                style={{
                                    padding: '2px 8px',
                                    fontSize: '11px',
                                    background: 'var(--color-error)',
                                    color: '#fff',
                                    border: 'none',
                                    borderRadius: 'var(--border-radius)',
                                    cursor: 'pointer',
                                }}
                            >
                                ✓
                            </button>
                            <button
                                onClick={() => setConfirming(false)}
                                style={{
                                    padding: '2px 8px',
                                    fontSize: '11px',
                                    background: 'transparent',
                                    color: 'var(--color-text-muted)',
                                    border: '1px solid var(--color-border)',
                                    borderRadius: 'var(--border-radius)',
                                    cursor: 'pointer',
                                }}
                            >
                                ✕
                            </button>
                        </div>
                    ) : (
                        <button
                            onClick={() => setConfirming(true)}
                            style={{
                                background: 'transparent',
                                border: 'none',
                                cursor: 'pointer',
                                color: 'var(--color-error)',
                                fontSize: '13px',
                                padding: 0,
                                opacity: 0.6,
                                transition: 'var(--transition)',
                            }}
                            onMouseEnter={e => (e.currentTarget.style.opacity = '1')}
                            onMouseLeave={e => (e.currentTarget.style.opacity = '0.6')}
                            title="Remove from list"
                        >
                            🗑️
                        </button>
                    )}
                </div>
            </div>
        </div>
    )
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function BookListDetailPage() {
    const { id } = useParams<{ id: string }>()
    const navigate = useNavigate()
    const queryClient = useQueryClient()

    const [showEditModal, setShowEditModal] = useState(false)
    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const [editTitle, setEditTitle] = useState('')
    const [editDescription, setEditDescription] = useState('')
    const [editVisibility, setEditVisibility] = useState<'private' | 'public'>('private')

    const { data, isLoading, isError } = useQuery({
        queryKey: ['book-list', id],
        queryFn: () => getList(id!),
        enabled: !!id,
    })

    const updateMutation = useMutation({
        mutationFn: () =>
            updateList(id!, {
                title: editTitle,
                description: editDescription || undefined,
                visibility: editVisibility,
            }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['book-list', id] })
            setShowEditModal(false)
        },
    })

    const deleteMutation = useMutation({
        mutationFn: () => deleteList(id!),
        onSuccess: () => navigate('/lists'),
    })

    const removeMutation = useMutation({
        mutationFn: (bookId: string) => removeBookFromList(id!, bookId),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['book-list', id] }),
    })

    function openEditModal() {
        if (!data) return
        setEditTitle(data.list.title)
        setEditDescription(data.list.description ?? '')
        setEditVisibility(data.list.visibility)
        setShowEditModal(true)
    }

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
            {/* Back */}
            <button
                onClick={() => navigate('/lists')}
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
                }}
            >
                ← Back
            </button>

            {isLoading && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '48px 0' }}>
                    <Spinner />
                </div>
            )}

            {isError && <EmptyState icon="⚠️" title="Failed to load list" />}

            {!isLoading && !isError && data && (
                <>
                    {/* Header */}
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'flex-start',
                            justifyContent: 'space-between',
                            gap: '16px',
                            marginBottom: '28px',
                            flexWrap: 'wrap',
                        }}
                    >
                        <div>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '6px' }}>
                                <h1
                                    style={{
                                        margin: 0,
                                        fontSize: '24px',
                                        fontWeight: 700,
                                        fontFamily: 'var(--font-heading)',
                                        color: 'var(--color-text)',
                                    }}
                                >
                                    {data.list.title}
                                </h1>
                                <Badge
                                    label={data.list.visibility === 'public' ? 'Public' : 'Private'}
                                    variant={data.list.visibility === 'public' ? 'success' : 'muted'}
                                    size="sm"
                                />
                            </div>
                            {data.list.description && (
                                <p
                                    style={{
                                        margin: 0,
                                        fontSize: '14px',
                                        color: 'var(--color-text-muted)',
                                        fontFamily: 'var(--font-body)',
                                    }}
                                >
                                    {data.list.description}
                                </p>
                            )}
                        </div>
                        <div style={{ display: 'flex', gap: '8px' }}>
                            <Button label="Edit" variant="secondary" onClick={openEditModal} />
                            <Button label="Delete" variant="danger" onClick={() => setShowDeleteModal(true)} />
                        </div>
                    </div>

                    {/* Book grid */}
                    {data.books.length === 0 ? (
                        <EmptyState
                            icon="📚"
                            title="This list is empty"
                            description="Add books from the catalogue"
                        />
                    ) : (
                        <div
                            style={{
                                display: 'grid',
                                gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))',
                                gap: '16px',
                            }}
                        >
                            {data.books.map((item: BookListItem) => (
                                <BookCard
                                    key={item.book_id}
                                    item={item}
                                    onRemove={bookId => removeMutation.mutate(bookId)}
                                    isRemoving={removeMutation.isPending}
                                />
                            ))}
                        </div>
                    )}
                </>
            )}

            {/* Edit Modal */}
            <Modal
                isOpen={showEditModal}
                onClose={() => setShowEditModal(false)}
                title="Edit List"
                confirmLabel="Save"
                onConfirm={() => updateMutation.mutate()}
                isLoading={updateMutation.isPending}
                size="sm"
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Input label="Title" value={editTitle} onChange={setEditTitle} placeholder="List title" />
                    <Input
                        label="Description (optional)"
                        value={editDescription}
                        onChange={setEditDescription}
                        placeholder="What's this list about?"
                    />
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                        <label
                            style={{
                                fontSize: '14px',
                                fontWeight: 500,
                                color: 'var(--color-text)',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            Visibility
                        </label>
                        <select
                            value={editVisibility}
                            onChange={e => setEditVisibility(e.target.value as 'private' | 'public')}
                            style={{
                                padding: '8px 12px',
                                background: 'var(--input-bg)',
                                border: '1px solid var(--color-border)',
                                borderRadius: 'var(--border-radius)',
                                color: 'var(--color-text)',
                                fontFamily: 'var(--font-body)',
                                fontSize: '14px',
                                outline: 'none',
                                cursor: 'pointer',
                            }}
                        >
                            <option value="private">Private</option>
                            <option value="public">Public</option>
                        </select>
                    </div>
                </div>
            </Modal>

            {/* Delete Modal */}
            <Modal
                isOpen={showDeleteModal}
                onClose={() => setShowDeleteModal(false)}
                title="Delete List"
                confirmLabel="Delete"
                confirmVariant="danger"
                onConfirm={() => deleteMutation.mutate()}
                isLoading={deleteMutation.isPending}
                size="sm"
            >
                <p style={{ margin: 0, color: 'var(--color-text)', fontSize: '14px' }}>
                    Are you sure you want to delete <strong>{data?.list.title}</strong>? This cannot be undone.
                </p>
            </Modal>
        </div>
    )
}
