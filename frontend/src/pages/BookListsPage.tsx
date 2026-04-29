import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getLists, createList } from '../api/lists'
import type { BookList } from '../types'
import { Button, Input, Modal, Card, Badge, Spinner, EmptyState } from '../components'

export default function BookListsPage() {
    const navigate = useNavigate()
    const queryClient = useQueryClient()

    const [showNewModal, setShowNewModal] = useState(false)
    const [title, setTitle] = useState('')
    const [description, setDescription] = useState('')
    const [visibility, setVisibility] = useState<'private' | 'public'>('private')

    const { data, isLoading, isError } = useQuery({
        queryKey: ['book-lists'],
        queryFn: getLists,
    })

    const lists: BookList[] = data?.lists ?? []

    const createMutation = useMutation({
        mutationFn: () => createList({ title, description: description || undefined, visibility }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['book-lists'] })
            closeModal()
        },
    })

    function closeModal() {
        setShowNewModal(false)
        setTitle('')
        setDescription('')
        setVisibility('private')
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
            {/* Header */}
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    marginBottom: '24px',
                }}
            >
                <h1
                    style={{
                        margin: 0,
                        fontSize: '24px',
                        fontWeight: 700,
                        fontFamily: 'var(--font-heading)',
                        color: 'var(--color-text)',
                    }}
                >
                    My Lists
                </h1>
                <Button label="New List" variant="primary" onClick={() => setShowNewModal(true)} />
            </div>

            {/* Loading */}
            {isLoading && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '48px 0' }}>
                    <Spinner />
                </div>
            )}

            {/* Error */}
            {isError && <EmptyState icon="⚠️" title="Failed to load lists" />}

            {/* Grid */}
            {!isLoading && !isError && lists.length > 0 && (
                <div
                    style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))',
                        gap: '16px',
                    }}
                >
                    {lists.map(list => (
                        <Card
                            key={list.id}
                            hover
                            onClick={() => navigate(`/lists/${list.id}`)}
                        >
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                <span
                                    style={{
                                        fontSize: '15px',
                                        fontWeight: 700,
                                        color: 'var(--color-text)',
                                        fontFamily: 'var(--font-heading)',
                                    }}
                                >
                                    {list.title}
                                </span>
                                {list.description && (
                                    <p
                                        style={{
                                            margin: 0,
                                            fontSize: '13px',
                                            color: 'var(--color-text-muted)',
                                            lineHeight: 1.5,
                                            overflow: 'hidden',
                                            display: '-webkit-box',
                                            WebkitLineClamp: 2,
                                            WebkitBoxOrient: 'vertical',
                                        }}
                                    >
                                        {list.description}
                                    </p>
                                )}
                                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                                    <span
                                        style={{
                                            fontSize: '12px',
                                            color: 'var(--color-text-muted)',
                                            fontFamily: 'var(--font-body)',
                                        }}
                                    >
                                        📚 {list.book_count ?? 0} books
                                    </span>
                                    <Badge
                                        label={list.visibility === 'public' ? 'Public' : 'Private'}
                                        variant={list.visibility === 'public' ? 'success' : 'muted'}
                                        size="sm"
                                    />
                                </div>
                            </div>
                        </Card>
                    ))}
                </div>
            )}

            {/* Empty state */}
            {!isLoading && !isError && lists.length === 0 && (
                <EmptyState
                    icon="📋"
                    title="No lists yet"
                    description="Create a list to curate your favourite books"
                    action={{ label: 'New List', onClick: () => setShowNewModal(true) }}
                />
            )}

            {/* New List Modal */}
            <Modal
                isOpen={showNewModal}
                onClose={closeModal}
                title="New List"
                confirmLabel="Create"
                onConfirm={() => createMutation.mutate()}
                confirmVariant="primary"
                isLoading={createMutation.isPending}
                size="sm"
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Input
                        label="Title"
                        value={title}
                        onChange={setTitle}
                        placeholder="e.g. Best Sci-Fi Novels"
                    />
                    <Input
                        label="Description (optional)"
                        value={description}
                        onChange={setDescription}
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
                            value={visibility}
                            onChange={e => setVisibility(e.target.value as 'private' | 'public')}
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
        </div>
    )
}
