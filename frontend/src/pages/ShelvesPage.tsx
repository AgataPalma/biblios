import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getShelves, createShelf, updateShelf, deleteShelf } from '../api/shelves'
import type { Shelf } from '../types'
import { Button, Input, Modal, Card, Badge, Spinner, EmptyState } from '../components'

export default function ShelvesPage() {
    const navigate = useNavigate()
    const queryClient = useQueryClient()

    // ── New shelf modal state ─────────────────────────────────────────────────
    const [showNewModal, setShowNewModal] = useState(false)
    const [newShelfName, setNewShelfName] = useState('')

    // ── Rename modal state ────────────────────────────────────────────────────
    const [renameTarget, setRenameTarget] = useState<Shelf | null>(null)
    const [renameValue, setRenameValue] = useState('')

    // ── Delete modal state ────────────────────────────────────────────────────
    const [deleteTarget, setDeleteTarget] = useState<Shelf | null>(null)

    // ── Data fetching ─────────────────────────────────────────────────────────
    const { data: shelves, isLoading, isError } = useQuery({
        queryKey: ['shelves'],
        queryFn: getShelves,
    })

    // ── Mutations ─────────────────────────────────────────────────────────────
    const createMutation = useMutation({
        mutationFn: () => createShelf({ name: newShelfName }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['shelves'] })
            setShowNewModal(false)
            setNewShelfName('')
        },
    })

    const renameMutation = useMutation({
        mutationFn: () => updateShelf(renameTarget!.id, { name: renameValue }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['shelves'] })
            setRenameTarget(null)
        },
    })

    const deleteMutation = useMutation({
        mutationFn: () => deleteShelf(deleteTarget!.id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['shelves'] })
            setDeleteTarget(null)
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
            {/* Page header */}
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
                    My Shelves
                </h1>
                <Button
                    label="New Shelf"
                    variant="primary"
                    onClick={() => setShowNewModal(true)}
                />
            </div>

            {/* Loading */}
            {isLoading && (
                <div style={{ display: 'flex', justifyContent: 'center', padding: '48px 0' }}>
                    <Spinner />
                </div>
            )}

            {/* Error */}
            {isError && (
                <EmptyState icon="⚠️" title="Failed to load shelves" />
            )}

            {/* Shelf grid */}
            {!isLoading && !isError && shelves && shelves.length > 0 && (
                <div
                    style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
                        gap: '16px',
                    }}
                >
                    {shelves.map(shelf => (
                        <Card
                            key={shelf.id}
                            hover
                            onClick={() => navigate(`/shelves/${shelf.id}`)}
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    flexDirection: 'column',
                                    gap: '8px',
                                }}
                            >
                                {/* Name + action buttons row */}
                                <div
                                    style={{
                                        display: 'flex',
                                        alignItems: 'flex-start',
                                        justifyContent: 'space-between',
                                        gap: '8px',
                                    }}
                                >
                                    <span
                                        style={{
                                            fontSize: '15px',
                                            fontWeight: 700,
                                            color: 'var(--color-text)',
                                            fontFamily: 'var(--font-heading)',
                                            lineHeight: 1.3,
                                            wordBreak: 'break-word',
                                        }}
                                    >
                                        {shelf.name}
                                    </span>
                                    <div style={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                                        {/* Edit button */}
                                        <button
                                            title="Rename shelf"
                                            onClick={e => {
                                                e.stopPropagation()
                                                setRenameTarget(shelf)
                                                setRenameValue(shelf.name)
                                            }}
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
                                                color: 'var(--color-text-muted)',
                                                transition: 'var(--transition)',
                                                padding: 0,
                                            }}
                                        >
                                            ✏️
                                        </button>
                                        {/* Delete button */}
                                        <button
                                            title="Delete shelf"
                                            onClick={e => {
                                                e.stopPropagation()
                                                setDeleteTarget(shelf)
                                            }}
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
                                            }}
                                        >
                                            🗑️
                                        </button>
                                    </div>
                                </div>

                                {/* Book count badge */}
                                <Badge
                                    label={`${shelf.book_count ?? 0} books`}
                                    variant="default"
                                />
                            </div>
                        </Card>
                    ))}
                </div>
            )}

            {/* Empty state */}
            {!isLoading && !isError && shelves && shelves.length === 0 && (
                <EmptyState
                    icon="🗂️"
                    title="No shelves yet"
                    description="Create a shelf to organise your books"
                    action={{
                        label: 'New Shelf',
                        onClick: () => setShowNewModal(true),
                    }}
                />
            )}

            {/* ── New Shelf Modal ─────────────────────────────────────────────── */}
            <Modal
                isOpen={showNewModal}
                onClose={() => {
                    setShowNewModal(false)
                    setNewShelfName('')
                }}
                title="New Shelf"
                confirmLabel="Create"
                onConfirm={() => createMutation.mutate()}
                confirmVariant="primary"
                isLoading={createMutation.isPending}
                size="sm"
            >
                <Input
                    label="Shelf name"
                    value={newShelfName}
                    onChange={setNewShelfName}
                    placeholder="e.g. Favourites"
                />
            </Modal>

            {/* ── Rename Modal ────────────────────────────────────────────────── */}
            <Modal
                isOpen={renameTarget !== null}
                onClose={() => setRenameTarget(null)}
                title="Rename Shelf"
                confirmLabel="Save"
                onConfirm={() => renameMutation.mutate()}
                confirmVariant="primary"
                isLoading={renameMutation.isPending}
                size="sm"
            >
                <Input
                    label="Shelf name"
                    value={renameValue}
                    onChange={setRenameValue}
                    placeholder="Shelf name"
                />
            </Modal>

            {/* ── Delete Confirmation Modal ───────────────────────────────────── */}
            <Modal
                isOpen={deleteTarget !== null}
                onClose={() => setDeleteTarget(null)}
                title="Delete Shelf"
                confirmLabel="Delete"
                onConfirm={() => deleteMutation.mutate()}
                confirmVariant="danger"
                isLoading={deleteMutation.isPending}
                size="sm"
            >
                <p style={{ margin: 0, color: 'var(--color-text)', fontSize: '14px' }}>
                    Are you sure you want to delete{' '}
                    <strong>{deleteTarget?.name}</strong>? This action cannot be undone.
                </p>
            </Modal>
        </div>
    )
}
