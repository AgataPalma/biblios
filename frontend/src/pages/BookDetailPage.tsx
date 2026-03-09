import { useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getBook, addCopy, updateBook, uploadBookCover } from '../api/books'
import { Badge, Card, Modal, Spinner } from '../components'
import { useAuth } from '../context/AuthContext'
import type { Book, Edition } from '../types'

// ─── Colour helper ────────────────────────────────────────────────────────────

const PALETTE = [
    '#2563eb', '#16a34a', '#dc2626', '#9333ea',
    '#ea580c', '#0891b2', '#d97706', '#4f46e5',
    '#0d9488', '#be185d',
]
function pickColor(title: string): string {
    const idx = title.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % PALETTE.length
    return PALETTE[idx]
}

// ─── BookDetailPage ───────────────────────────────────────────────────────────

export default function BookDetailPage() {
    const { id }      = useParams<{ id: string }>()
    const navigate    = useNavigate()
    const queryClient = useQueryClient()
    const { user }    = useAuth()
    const canModerate = user?.role === 'moderator' || user?.role === 'admin'

    // Add to library modal
    const [addOpen,           setAddOpen]           = useState(false)
    const [selectedEditionId, setSelectedEditionId] = useState('')
    const [selectedCondition, setSelectedCondition] = useState('good')
    const [addError,          setAddError]          = useState('')

    // Inline edit state (mod/admin only)
    const [editing,      setEditing]      = useState(false)
    const [editTitle,    setEditTitle]    = useState('')
    const [editDesc,     setEditDesc]     = useState('')
    const [saveError,    setSaveError]    = useState('')
    // Cover: file selected by user — stored as File object.
    // Uploaded via uploadBookCover (base64 placeholder until real endpoint exists).
    const [coverFile,    setCoverFile]    = useState<File | null>(null)
    const [coverPreview, setCoverPreview] = useState<string | undefined>()
    const fileInputRef = useRef<HTMLInputElement>(null)

    const { data: book, isLoading, isError } = useQuery({
        queryKey: ['book', id],
        queryFn: () => getBook(id!),
        enabled: !!id,
    })

    const addCopyMutation = useMutation({
        mutationFn: () => addCopy(selectedEditionId, { condition: selectedCondition }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['my-library'] })
            setAddOpen(false)
            setAddError('')
        },
        onError: () => setAddError('Failed to add to library. Please try again.'),
    })

    const saveMutation = useMutation({
        mutationFn: async () => {
            // If a new cover image was chosen, upload it first (placeholder implementation)
            let newCoverUrl: string | undefined = undefined
            if (coverFile && id) {
                newCoverUrl = await uploadBookCover(id, coverFile)
            }
            return updateBook(id!, {
                title:       editTitle.trim()  || undefined,
                description: editDesc.trim()   || undefined,
                cover_url:   newCoverUrl,
            })
        },
        onSuccess: (updated: Book) => {
            queryClient.setQueryData(['book', id], updated)
            queryClient.invalidateQueries({ queryKey: ['books'] })
            setEditing(false)
            setSaveError('')
            setCoverFile(null)
            setCoverPreview(undefined)
        },
        onError: () => setSaveError('Failed to save changes. Please try again.'),
    })

    function startEditing() {
        if (!book) return
        setEditTitle(book.title)
        setEditDesc(book.description ?? '')
        setCoverFile(null)
        setCoverPreview(undefined)
        setSaveError('')
        setEditing(true)
    }

    function cancelEditing() {
        setEditing(false)
        setSaveError('')
        setCoverFile(null)
        setCoverPreview(undefined)
    }

    function openAddModal() {
        if (!book) return
        setSelectedEditionId(book.editions?.[0]?.id ?? '')
        setSelectedCondition('good')
        setAddError('')
        setAddOpen(true)
    }

    function handleCoverFile(e: React.ChangeEvent<HTMLInputElement>) {
        const file = e.target.files?.[0]
        if (!file) return
        setCoverFile(file)
        const reader = new FileReader()
        reader.onload = ev => setCoverPreview(ev.target?.result as string)
        reader.readAsDataURL(file)
    }

    // ── Loading / error states ──────────────────────────────────────────────

    if (isLoading) {
        return (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 'calc(100vh - 56px)' }}>
                <Spinner size="lg" label="Loading book..." />
            </div>
        )
    }

    if (isError || !book) {
        return (
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: 'calc(100vh - 56px)', gap: '16px' }}>
                <span style={{ fontSize: '48px' }}>📭</span>
                <h2 style={{ margin: 0, color: 'var(--color-text)', fontFamily: 'var(--font-heading)' }}>Book not found</h2>
                <button
                    onClick={() => navigate('/books')}
                    style={{ background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', padding: '10px 20px', cursor: 'pointer', fontFamily: 'var(--font-body)', fontSize: '14px' }}
                >
                    Back to catalogue
                </button>
            </div>
        )
    }

    const coverColor  = pickColor(book.title)
    // Support both new cover_image_url and legacy cover_url field names
    const storedCover = (book as any).cover_image_url ?? book.cover_url
    const displayCover = coverPreview ?? storedCover
    const primaryEdition = book.editions?.[0]

    return (
        <div style={{ minHeight: 'calc(100vh - 56px)', padding: '32px 24px', maxWidth: '900px', margin: '0 auto' }}>
            {/* Back button */}
            <button
                onClick={() => navigate('/books')}
                style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', fontSize: '13px', fontFamily: 'var(--font-body)', padding: 0, marginBottom: '24px', display: 'flex', alignItems: 'center', gap: '6px' }}
            >
                ← Back to catalogue
            </button>

            <div style={{ display: 'grid', gridTemplateColumns: '220px 1fr', gap: '32px', alignItems: 'start' }}>

                {/* ── Left column — cover + edition info ── */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    {/* Cover */}
                    <div style={{ position: 'relative' }}>
                        <div style={{ width: '220px', height: '300px', borderRadius: '6px', overflow: 'hidden', boxShadow: 'var(--shadow-lg)', background: displayCover ? undefined : `linear-gradient(135deg, ${coverColor}dd, ${coverColor}88)`, display: 'flex', alignItems: 'center', justifyContent: 'center', border: '1px solid var(--color-border)' }}>
                            {displayCover
                                ? <img src={displayCover} alt={book.title} style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                                : <span style={{ fontSize: '64px', opacity: 0.5 }}>📖</span>
                            }
                        </div>

                        {/* Cover upload button — mod/admin in edit mode only */}
                        {canModerate && editing && (
                            <>
                                <button
                                    onClick={() => fileInputRef.current?.click()}
                                    title="Upload cover image"
                                    style={{ position: 'absolute', bottom: '8px', right: '8px', width: '32px', height: '32px', border: 'none', borderRadius: '50%', background: 'rgba(0,0,0,0.6)', color: '#fff', fontSize: '16px', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                                >
                                    📷
                                </button>
                                <input ref={fileInputRef} type="file" accept="image/*" style={{ display: 'none' }} onChange={handleCoverFile} />
                                {coverFile && (
                                    <p style={{ margin: '6px 0 0', fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', textAlign: 'center' }}>
                                        {coverFile.name}
                                    </p>
                                )}
                            </>
                        )}
                    </div>

                    {/* Edition details card */}
                    {primaryEdition && (
                        <Card padding="sm">
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                <p style={{ margin: 0, fontSize: '11px', fontWeight: 700, color: 'var(--color-text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', fontFamily: 'var(--font-body)' }}>
                                    Edition Details
                                </p>
                                {[
                                    { label: 'Format',    value: primaryEdition.format },
                                    { label: 'Language',  value: primaryEdition.language },
                                    { label: 'Publisher', value: primaryEdition.publisher },
                                    { label: 'Published', value: primaryEdition.published_at },
                                    { label: 'Pages',     value: primaryEdition.page_count },
                                    { label: 'ISBN',      value: primaryEdition.isbn },
                                ].filter(f => f.value).map(field => (
                                    <div key={field.label} style={{ display: 'flex', justifyContent: 'space-between', gap: '8px' }}>
                                        <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', flexShrink: 0 }}>{field.label}</span>
                                        <span style={{ fontSize: '11px', color: 'var(--color-text)', fontFamily: 'var(--font-body)', textAlign: 'right', textTransform: 'capitalize' }}>{String(field.value)}</span>
                                    </div>
                                ))}
                            </div>
                        </Card>
                    )}
                </div>

                {/* ── Right column — main info ── */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                    {/* Status row with mod/admin edit toggle */}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
                        <Badge label={book.status} variant={book.status === 'approved' ? 'success' : 'warning'} size="sm" />
                        {book.genres?.map(g => <Badge key={g.id} label={g.name} variant="info" size="sm" />)}

                        {canModerate && !editing && (
                            <button
                                onClick={startEditing}
                                style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '6px', padding: '5px 12px', background: 'var(--color-surface-alt)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '12px', color: 'var(--color-text)', cursor: 'pointer', fontFamily: 'var(--font-body)' }}
                            >
                                ✏️ Edit
                            </button>
                        )}

                        {canModerate && editing && (
                            <div style={{ marginLeft: 'auto', display: 'flex', gap: '6px' }}>
                                <button
                                    onClick={() => saveMutation.mutate()}
                                    disabled={saveMutation.isPending}
                                    style={{ padding: '5px 14px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '12px', fontWeight: 600, cursor: saveMutation.isPending ? 'not-allowed' : 'pointer', fontFamily: 'var(--font-body)', opacity: saveMutation.isPending ? 0.7 : 1 }}
                                >
                                    {saveMutation.isPending ? 'Saving…' : '✓ Save'}
                                </button>
                                <button
                                    onClick={cancelEditing}
                                    style={{ padding: '5px 12px', background: 'var(--color-surface-alt)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '12px', color: 'var(--color-text)', cursor: 'pointer', fontFamily: 'var(--font-body)' }}
                                >
                                    Cancel
                                </button>
                            </div>
                        )}
                    </div>

                    {saveError && (
                        <p style={{ margin: 0, fontSize: '12px', color: 'var(--color-error)', fontFamily: 'var(--font-body)' }}>{saveError}</p>
                    )}

                    {/* Title */}
                    {editing ? (
                        <input
                            value={editTitle}
                            onChange={e => setEditTitle(e.target.value)}
                            style={{ width: '100%', padding: '8px 10px', boxSizing: 'border-box', background: 'var(--input-bg)', border: '1px solid var(--color-primary)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '22px', fontWeight: 700, fontFamily: 'var(--font-heading)', outline: 'none' }}
                        />
                    ) : (
                        <h1 style={{ margin: 0, fontSize: '28px', fontWeight: 700, color: 'var(--color-text)', fontFamily: 'var(--font-heading)', lineHeight: 1.2 }}>
                            {book.title}
                        </h1>
                    )}

                    <p style={{ margin: 0, fontSize: '15px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                        {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                    </p>

                    {/* Description */}
                    {editing ? (
                        <textarea
                            value={editDesc}
                            onChange={e => setEditDesc(e.target.value)}
                            rows={5}
                            placeholder="Description…"
                            style={{ width: '100%', padding: '10px 12px', boxSizing: 'border-box', background: 'var(--input-bg)', border: '1px solid var(--color-primary)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '14px', fontFamily: 'var(--font-body)', outline: 'none', resize: 'vertical', lineHeight: '1.6' }}
                        />
                    ) : book.description ? (
                        <Card padding="md">
                            <p style={{ margin: 0, fontSize: '14px', color: 'var(--color-text)', lineHeight: '1.7', fontFamily: 'var(--font-body)' }}>
                                {book.description}
                            </p>
                        </Card>
                    ) : null}

                    {/* All editions list */}
                    {book.editions && book.editions.length > 1 && (
                        <div>
                            <h3 style={{ margin: '0 0 12px', fontSize: '14px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-heading)' }}>
                                All Editions ({book.editions.length})
                            </h3>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                {book.editions.map((edition: Edition) => (
                                    <Card key={edition.id} padding="sm">
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
                                            <Badge label={edition.format} variant="default" size="sm" />
                                            {edition.language && <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>{edition.language.toUpperCase()}</span>}
                                            {edition.publisher && <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>{edition.publisher}</span>}
                                            {edition.isbn && <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', marginLeft: 'auto' }}>ISBN: {edition.isbn}</span>}
                                            <button
                                                onClick={() => {
                                                    setSelectedEditionId(edition.id)
                                                    setSelectedCondition('good')
                                                    setAddError('')
                                                    setAddOpen(true)
                                                }}
                                                style={{ marginLeft: 'auto', padding: '4px 10px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '11px', fontWeight: 600, cursor: 'pointer', fontFamily: 'var(--font-body)' }}
                                            >
                                                + Add
                                            </button>
                                        </div>
                                    </Card>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Action buttons */}
                    <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
                        <button
                            onClick={openAddModal}
                            style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 20px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', transition: 'var(--transition)', fontFamily: 'var(--font-body)' }}
                            onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary-hover)'}
                            onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary)'}
                        >
                            ➕ Add to my library
                        </button>

                        <button
                            onClick={() => navigate('/books')}
                            style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 20px', background: 'var(--color-surface-alt)', color: 'var(--color-text)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', transition: 'var(--transition)', fontFamily: 'var(--font-body)' }}
                        >
                            ← Back
                        </button>
                    </div>
                </div>
            </div>

            {/* ── Add to library modal ── */}
            <Modal
                isOpen={addOpen}
                onClose={() => setAddOpen(false)}
                title={`Add to library — ${book.title}`}
                confirmLabel="Add to library"
                onConfirm={() => addCopyMutation.mutate()}
                isLoading={addCopyMutation.isPending}
                size="sm"
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    {/* Edition picker */}
                    <div>
                        <label style={{ display: 'block', fontSize: '11px', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', marginBottom: '6px' }}>
                            Choose Edition
                        </label>
                        <select
                            value={selectedEditionId}
                            onChange={e => setSelectedEditionId(e.target.value)}
                            style={{ width: '100%', padding: '8px 10px', background: 'var(--input-bg)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '13px', fontFamily: 'var(--font-body)' }}
                        >
                            {book.editions?.map((ed: Edition) => (
                                <option key={ed.id} value={ed.id}>
                                    {ed.format}{ed.language ? ` · ${ed.language.toUpperCase()}` : ''}{ed.publisher ? ` · ${ed.publisher}` : ''}{ed.isbn ? ` · ${ed.isbn}` : ''}
                                </option>
                            ))}
                        </select>
                        <p style={{ margin: '6px 0 0', fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                            Don't see your edition?{' '}
                            <button
                                onClick={() => { setAddOpen(false); navigate('/books/add') }}
                                style={{ background: 'none', border: 'none', color: 'var(--color-primary)', fontSize: '11px', cursor: 'pointer', fontFamily: 'var(--font-body)', padding: 0, textDecoration: 'underline' }}
                            >
                                Add a new edition
                            </button>
                        </p>
                    </div>

                    {/* Condition */}
                    <div>
                        <label style={{ display: 'block', fontSize: '11px', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', marginBottom: '6px' }}>
                            Condition
                        </label>
                        <div style={{ display: 'flex', gap: '6px' }}>
                            {['new', 'good', 'fair', 'poor'].map(c => (
                                <button
                                    key={c}
                                    onClick={() => setSelectedCondition(c)}
                                    style={{ flex: 1, padding: '7px 0', border: `1px solid ${selectedCondition === c ? 'var(--color-primary)' : 'var(--color-border)'}`, borderRadius: 'var(--border-radius)', background: selectedCondition === c ? 'var(--color-primary)' : 'var(--color-surface-alt)', color: selectedCondition === c ? 'var(--color-primary-text)' : 'var(--color-text)', fontSize: '12px', fontFamily: 'var(--font-body)', fontWeight: selectedCondition === c ? 600 : 400, cursor: 'pointer', textTransform: 'capitalize', transition: 'var(--transition)' }}
                                >
                                    {c}
                                </button>
                            ))}
                        </div>
                    </div>

                    {addError && <p style={{ margin: 0, fontSize: '12px', color: 'var(--color-error)', fontFamily: 'var(--font-body)' }}>{addError}</p>}
                </div>
            </Modal>
        </div>
    )
}