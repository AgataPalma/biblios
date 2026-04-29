import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getLibrary, inviteMember, removeMember } from '../api/libraries'
import { getCollections, createCollection } from '../api/collections'
import { useAuth } from '../context/AuthContext'
import type { LibraryMember, Collection } from '../types'
import { Button, Input, Modal, Badge, Spinner, EmptyState, Card } from '../components'

type Tab = 'books' | 'members' | 'collections'

const TABS: { id: Tab; label: string }[] = [
    { id: 'books', label: 'Books' },
    { id: 'members', label: 'Members' },
    { id: 'collections', label: 'Collections' },
]

// ── Books Tab ─────────────────────────────────────────────────────────────────

function BooksTab() {
    return <EmptyState icon="📚" title="No books in this library" />
}

// ── Members Tab ───────────────────────────────────────────────────────────────

interface MembersTabProps {
    members: LibraryMember[]
    libraryId: string
    isOwner: boolean
}

function MembersTab({ members, libraryId, isOwner }: MembersTabProps) {
    const queryClient = useQueryClient()

    const removeMutation = useMutation({
        mutationFn: (userId: string) => removeMember(libraryId, userId),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['library', libraryId] })
        },
    })

    if (members.length === 0) {
        return <EmptyState icon="👥" title="No members" />
    }

    return (
        <div>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <tbody>
                    {members.map((member) => (
                        <tr
                            key={member.user_id}
                            style={{ borderBottom: '1px solid var(--color-border)' }}
                        >
                            {/* Avatar */}
                            <td style={{ padding: '12px 8px', width: '56px' }}>
                                {member.avatar_url ? (
                                    <img
                                        src={member.avatar_url}
                                        alt={member.username}
                                        style={{
                                            width: '40px',
                                            height: '40px',
                                            borderRadius: '50%',
                                            objectFit: 'cover',
                                        }}
                                    />
                                ) : (
                                    <div
                                        style={{
                                            width: '40px',
                                            height: '40px',
                                            borderRadius: '50%',
                                            background: 'var(--color-surface-alt)',
                                            border: '1px solid var(--color-border)',
                                            display: 'flex',
                                            alignItems: 'center',
                                            justifyContent: 'center',
                                            fontSize: '16px',
                                            color: 'var(--color-text-muted)',
                                            fontFamily: 'var(--font-body)',
                                        }}
                                    >
                                        {member.username.charAt(0).toUpperCase()}
                                    </div>
                                )}
                            </td>

                            {/* Username */}
                            <td
                                style={{
                                    padding: '12px 8px',
                                    fontSize: '14px',
                                    color: 'var(--color-text)',
                                    fontFamily: 'var(--font-body)',
                                    fontWeight: 500,
                                }}
                            >
                                {member.username}
                            </td>

                            {/* Role badge */}
                            <td style={{ padding: '12px 8px' }}>
                                <Badge
                                    label={member.role}
                                    variant={member.role === 'owner' ? 'info' : 'default'}
                                />
                            </td>

                            {/* Remove button (owner only, not for self) */}
                            <td style={{ padding: '12px 8px', textAlign: 'right' }}>
                                {isOwner && member.role !== 'owner' && (
                                    <Button
                                        label="Remove"
                                        variant="danger"
                                        onClick={() => removeMutation.mutate(member.user_id)}
                                        isLoading={removeMutation.isPending}
                                    />
                                )}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    )
}

// ── Collections Tab ───────────────────────────────────────────────────────────

interface CollectionsTabProps {
    collections: Collection[]
    libraryId: string
    onNewCollection: () => void
}

function CollectionsTab({ collections, libraryId, onNewCollection }: CollectionsTabProps) {
    const navigate = useNavigate()

    return (
        <div>
            <div
                style={{
                    display: 'flex',
                    justifyContent: 'flex-end',
                    marginBottom: '16px',
                }}
            >
                <Button label="New Collection" onClick={onNewCollection} />
            </div>

            {collections.length === 0 ? (
                <EmptyState icon="📁" title="No collections" />
            ) : (
                <div
                    style={{
                        display: 'grid',
                        gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
                        gap: '12px',
                    }}
                >
                    {collections.map((col) => (
                        <Card
                            key={col.id}
                            hover
                            onClick={() => navigate(`/collections/${col.id}`)}
                            padding="md"
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    flexDirection: 'column',
                                    gap: '8px',
                                }}
                            >
                                <p
                                    style={{
                                        margin: 0,
                                        fontSize: '14px',
                                        fontWeight: 600,
                                        color: 'var(--color-text)',
                                        fontFamily: 'var(--font-heading)',
                                    }}
                                >
                                    {col.name}
                                </p>
                                <p
                                    style={{
                                        margin: 0,
                                        fontSize: '12px',
                                        color: 'var(--color-text-muted)',
                                        fontFamily: 'var(--font-body)',
                                    }}
                                >
                                    {col.book_count ?? 0} book{col.book_count !== 1 ? 's' : ''}
                                </p>
                                <Badge
                                    label={col.visibility}
                                    variant={col.visibility === 'public' ? 'success' : 'muted'}
                                    size="sm"
                                />
                            </div>
                        </Card>
                    ))}
                </div>
            )}
        </div>
    )
}

// ── Invite Member Modal ───────────────────────────────────────────────────────

interface InviteMemberModalProps {
    isOpen: boolean
    onClose: () => void
    libraryId: string
}

function InviteMemberModal({ isOpen, onClose, libraryId }: InviteMemberModalProps) {
    const [email, setEmail] = useState('')
    const [successMsg, setSuccessMsg] = useState('')
    const [errorMsg, setErrorMsg] = useState('')

    const mutation = useMutation({
        mutationFn: () => inviteMember(libraryId, email),
        onSuccess: () => {
            setSuccessMsg('Invitation sent successfully.')
            setErrorMsg('')
            setEmail('')
        },
        onError: (err: unknown) => {
            const msg = err instanceof Error ? err.message : 'Failed to send invitation.'
            setErrorMsg(msg)
            setSuccessMsg('')
        },
    })

    function handleClose() {
        setEmail('')
        setSuccessMsg('')
        setErrorMsg('')
        onClose()
    }

    return (
        <Modal
            isOpen={isOpen}
            onClose={handleClose}
            title="Invite Member"
            confirmLabel="Send Invite"
            onConfirm={() => mutation.mutate()}
            isLoading={mutation.isPending}
        >
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                <Input
                    label="Email address"
                    type="email"
                    value={email}
                    onChange={setEmail}
                    placeholder="member@example.com"
                />
                {successMsg && (
                    <p
                        style={{
                            margin: 0,
                            fontSize: '14px',
                            color: 'var(--color-success)',
                            fontFamily: 'var(--font-body)',
                        }}
                    >
                        {successMsg}
                    </p>
                )}
                {errorMsg && (
                    <p
                        style={{
                            margin: 0,
                            fontSize: '14px',
                            color: 'var(--color-error)',
                            fontFamily: 'var(--font-body)',
                        }}
                    >
                        {errorMsg}
                    </p>
                )}
            </div>
        </Modal>
    )
}

// ── New Collection Modal ──────────────────────────────────────────────────────

interface NewCollectionModalProps {
    isOpen: boolean
    onClose: () => void
    libraryId: string
}

function NewCollectionModal({ isOpen, onClose, libraryId }: NewCollectionModalProps) {
    const queryClient = useQueryClient()
    const [name, setName] = useState('')
    const [description, setDescription] = useState('')
    const [visibility, setVisibility] = useState<'public' | 'private'>('public')
    const [errorMsg, setErrorMsg] = useState('')

    const mutation = useMutation({
        mutationFn: () => createCollection(libraryId, { name, description, visibility }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['collections', libraryId] })
            handleClose()
        },
        onError: (err: unknown) => {
            const msg = err instanceof Error ? err.message : 'Failed to create collection.'
            setErrorMsg(msg)
        },
    })

    function handleClose() {
        setName('')
        setDescription('')
        setVisibility('public')
        setErrorMsg('')
        onClose()
    }

    function handleConfirm() {
        if (!name.trim()) {
            setErrorMsg('Name is required.')
            return
        }
        setErrorMsg('')
        mutation.mutate()
    }

    return (
        <Modal
            isOpen={isOpen}
            onClose={handleClose}
            title="New Collection"
            confirmLabel="Create"
            onConfirm={handleConfirm}
            isLoading={mutation.isPending}
        >
            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                <Input
                    label="Name"
                    value={name}
                    onChange={setName}
                    placeholder="Collection name"
                />
                <Input
                    label="Description (optional)"
                    value={description}
                    onChange={setDescription}
                    placeholder="What's this collection about?"
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
                        onChange={(e) => setVisibility(e.target.value as 'public' | 'private')}
                        style={{
                            background: 'var(--input-bg)',
                            border: '1px solid var(--color-border)',
                            borderRadius: 'var(--border-radius)',
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-body)',
                            fontSize: '14px',
                            padding: '8px 12px',
                            outline: 'none',
                            cursor: 'pointer',
                            width: '100%',
                        }}
                    >
                        <option value="public">Public</option>
                        <option value="private">Private</option>
                    </select>
                </div>

                {errorMsg && (
                    <p
                        style={{
                            margin: 0,
                            fontSize: '14px',
                            color: 'var(--color-error)',
                            fontFamily: 'var(--font-body)',
                        }}
                    >
                        {errorMsg}
                    </p>
                )}
            </div>
        </Modal>
    )
}

// ── Library Detail Page ───────────────────────────────────────────────────────

export default function LibraryDetailPage() {
    const { id } = useParams<{ id: string }>()
    const navigate = useNavigate()
    const { user } = useAuth()
    const [activeTab, setActiveTab] = useState<Tab>('books')
    const [inviteModalOpen, setInviteModalOpen] = useState(false)
    const [newCollectionModalOpen, setNewCollectionModalOpen] = useState(false)

    const {
        data,
        isLoading: libraryLoading,
        isError: libraryError,
    } = useQuery({
        queryKey: ['library', id],
        queryFn: () => getLibrary(id!),
        enabled: !!id,
    })

    const {
        data: collectionsData,
        isLoading: collectionsLoading,
        isError: collectionsError,
    } = useQuery({
        queryKey: ['collections', id],
        queryFn: () => getCollections(id!),
        enabled: !!id,
    })

    const isLoading = libraryLoading || collectionsLoading
    const isError = libraryError || collectionsError
    const isOwner = !!user && !!data && user.id === data.library.owner_id

    if (isLoading) {
        return (
            <div
                style={{
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center',
                    minHeight: '300px',
                }}
            >
                <Spinner />
            </div>
        )
    }

    if (isError || !data) {
        return (
            <div
                style={{
                    maxWidth: '1000px',
                    margin: '0 auto',
                    padding: '32px',
                }}
            >
                <EmptyState icon="⚠️" title="Failed to load library" />
            </div>
        )
    }

    const { library, members } = data
    const collections = collectionsData?.collections ?? []

    return (
        <div
            style={{
                maxWidth: '1000px',
                margin: '0 auto',
                padding: '32px',
                fontFamily: 'var(--font-body)',
            }}
        >
            {/* Back button */}
            <button
                onClick={() => navigate('/libraries')}
                style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    color: 'var(--color-text-muted)',
                    fontSize: '14px',
                    fontFamily: 'var(--font-body)',
                    padding: '0',
                    marginBottom: '20px',
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: '4px',
                    transition: 'var(--transition)',
                }}
                onMouseEnter={(e) => {
                    (e.currentTarget as HTMLButtonElement).style.color = 'var(--color-text)'
                }}
                onMouseLeave={(e) => {
                    (e.currentTarget as HTMLButtonElement).style.color = 'var(--color-text-muted)'
                }}
            >
                ← Back
            </button>

            {/* Page header */}
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
                    <h1
                        style={{
                            margin: '0 0 8px',
                            fontSize: '28px',
                            fontWeight: 700,
                            color: 'var(--color-text)',
                            fontFamily: 'var(--font-heading)',
                        }}
                    >
                        {library.name}
                    </h1>
                    {library.description && (
                        <p
                            style={{
                                margin: 0,
                                fontSize: '14px',
                                color: 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                                lineHeight: 1.5,
                            }}
                        >
                            {library.description}
                        </p>
                    )}
                </div>

                {/* Invite Member button (owner only) */}
                {isOwner && (
                    <Button
                        label="Invite Member"
                        onClick={() => setInviteModalOpen(true)}
                    />
                )}
            </div>

            {/* Tab navigation */}
            <nav
                style={{
                    display: 'flex',
                    borderBottom: '1px solid var(--color-border)',
                    marginBottom: '28px',
                    gap: '0',
                }}
                role="tablist"
            >
                {TABS.map((tab) => {
                    const isActive = tab.id === activeTab
                    return (
                        <button
                            key={tab.id}
                            role="tab"
                            aria-selected={isActive}
                            onClick={() => setActiveTab(tab.id)}
                            style={{
                                background: 'none',
                                border: 'none',
                                borderBottom: isActive
                                    ? '2px solid var(--color-primary)'
                                    : '2px solid transparent',
                                padding: '10px 16px',
                                cursor: 'pointer',
                                fontSize: '14px',
                                fontWeight: isActive ? 600 : 400,
                                color: isActive ? 'var(--color-primary)' : 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                                transition: 'var(--transition)',
                                marginBottom: '-1px',
                            }}
                        >
                            {tab.label}
                        </button>
                    )
                })}
            </nav>

            {/* Tab content */}
            <div role="tabpanel">
                {activeTab === 'books' && <BooksTab />}
                {activeTab === 'members' && (
                    <MembersTab
                        members={members}
                        libraryId={id!}
                        isOwner={isOwner}
                    />
                )}
                {activeTab === 'collections' && (
                    <CollectionsTab
                        collections={collections}
                        libraryId={id!}
                        onNewCollection={() => setNewCollectionModalOpen(true)}
                    />
                )}
            </div>

            {/* Modals */}
            <InviteMemberModal
                isOpen={inviteModalOpen}
                onClose={() => setInviteModalOpen(false)}
                libraryId={id!}
            />
            <NewCollectionModal
                isOpen={newCollectionModalOpen}
                onClose={() => setNewCollectionModalOpen(false)}
                libraryId={id!}
            />
        </div>
    )
}
