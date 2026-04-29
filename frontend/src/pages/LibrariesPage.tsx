import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getLibraries, createLibrary } from '../api/libraries'
import type { CooperativeLibrary } from '../types'
import { Button, Input, Modal, Card, Badge, Spinner, EmptyState } from '../components'

export default function LibrariesPage() {
    const navigate = useNavigate()
    const queryClient = useQueryClient()

    // ── New library modal state ───────────────────────────────────────────────
    const [showNewModal, setShowNewModal] = useState(false)
    const [name, setName] = useState('')
    const [description, setDescription] = useState('')
    const [visibility, setVisibility] = useState<'private' | 'semi_public' | 'public'>('private')
    const [isCooperative, setIsCooperative] = useState(false)

    // ── Data fetching ─────────────────────────────────────────────────────────
    const { data, isLoading, isError } = useQuery({
        queryKey: ['libraries'],
        queryFn: getLibraries,
    })

    const libraries = data?.libraries ?? []

    // ── Create mutation ───────────────────────────────────────────────────────
    const createMutation = useMutation({
        mutationFn: () =>
            createLibrary({ name, description, visibility, is_cooperative: isCooperative }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['libraries'] })
            setShowNewModal(false)
            resetForm()
        },
    })

    function resetForm() {
        setName('')
        setDescription('')
        setVisibility('private')
        setIsCooperative(false)
    }

    function handleCloseModal() {
        setShowNewModal(false)
        resetForm()
    }

    function visibilityBadge(lib: CooperativeLibrary) {
        if (lib.visibility === 'public') return <Badge variant="success" label="Public" />
        if (lib.visibility === 'semi_public') return <Badge variant="warning" label="Semi-Public" />
        return <Badge variant="default" label="Private" />
    }

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
                    Libraries
                </h1>
                <Button
                    label="New Library"
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
                <EmptyState icon="⚠️" title="Failed to load libraries" />
            )}

            {/* Library cards grid */}
            {!isLoading && !isError && libraries.length > 0 && (
                <div
                    style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))',
                        gap: '16px',
                    }}
                >
                    {libraries.map(lib => (
                        <Card
                            key={lib.id}
                            hover
                            onClick={() => navigate(`/libraries/${lib.id}`)}
                        >
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px' }}>
                                {/* Name */}
                                <span
                                    style={{
                                        fontSize: '16px',
                                        fontWeight: 700,
                                        color: 'var(--color-text)',
                                        fontFamily: 'var(--font-heading)',
                                        lineHeight: 1.3,
                                        wordBreak: 'break-word',
                                    }}
                                >
                                    {lib.name}
                                </span>

                                {/* Description */}
                                {lib.description && (
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
                                        {lib.description}
                                    </p>
                                )}

                                {/* Visibility badge */}
                                <div>{visibilityBadge(lib)}</div>

                                {/* Stats */}
                                <div
                                    style={{
                                        display: 'flex',
                                        gap: '12px',
                                        fontSize: '12px',
                                        color: 'var(--color-text-muted)',
                                    }}
                                >
                                    <span>👥 {lib.member_count ?? 0} members</span>
                                    <span>📚 {lib.book_count ?? 0} books</span>
                                </div>
                            </div>
                        </Card>
                    ))}
                </div>
            )}

            {/* Empty state */}
            {!isLoading && !isError && libraries.length === 0 && (
                <EmptyState
                    icon="🏛️"
                    title="No libraries yet"
                    description="Create a library to share your books"
                    action={{
                        label: 'New Library',
                        onClick: () => setShowNewModal(true),
                    }}
                />
            )}

            {/* ── New Library Modal ───────────────────────────────────────────── */}
            <Modal
                isOpen={showNewModal}
                onClose={handleCloseModal}
                title="New Library"
                confirmLabel="Create"
                onConfirm={() => createMutation.mutate()}
                confirmVariant="primary"
                isLoading={createMutation.isPending}
                size="md"
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Input
                        label="Name"
                        value={name}
                        onChange={setName}
                        placeholder="e.g. Community Reads"
                    />

                    <Input
                        label="Description"
                        value={description}
                        onChange={setDescription}
                        placeholder="What is this library for?"
                    />

                    {/* Visibility select */}
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
                            onChange={e =>
                                setVisibility(e.target.value as 'private' | 'semi_public' | 'public')
                            }
                            style={{
                                padding: '8px 12px',
                                fontSize: '14px',
                                background: 'var(--input-bg)',
                                border: '1px solid var(--color-border)',
                                borderRadius: 'var(--border-radius)',
                                color: 'var(--color-text)',
                                fontFamily: 'var(--font-body)',
                                cursor: 'pointer',
                                outline: 'none',
                                transition: 'var(--transition)',
                            }}
                        >
                            <option value="private">Private</option>
                            <option value="semi_public">Semi-Public</option>
                            <option value="public">Public</option>
                        </select>
                    </div>

                    {/* Cooperative checkbox */}
                    <label
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: '8px',
                            fontSize: '14px',
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-body)',
                            cursor: 'pointer',
                        }}
                    >
                        <input
                            type="checkbox"
                            checked={isCooperative}
                            onChange={e => setIsCooperative(e.target.checked)}
                            style={{ width: '16px', height: '16px', cursor: 'pointer' }}
                        />
                        Cooperative library
                    </label>
                </div>
            </Modal>
        </div>
    )
}
