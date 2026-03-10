import { useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getBook, addCopy, updateBook, uploadEditionCover, updateReadingStatus, removeCopy, deleteBook, deleteEdition, type UpdateCopyPayload } from '../api/books'
import { Badge, Card, Modal, Spinner } from '../components'
import { useAuth } from '../context/AuthContext'
import type { Book, Edition, UserBook } from '../types'

const PALETTE = ['#2563eb','#16a34a','#dc2626','#9333ea','#ea580c','#0891b2','#d97706','#4f46e5','#0d9488','#be185d']
function pickColor(title: string): string {
    const idx = title.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % PALETTE.length
    return PALETTE[idx]
}

function FieldLabel({ label, hint }: { label: string; hint?: string }) {
    return (
        <div style={{ marginBottom: '6px' }}>
            <label style={{ display: 'block', fontSize: '11px', fontWeight: 700, color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</label>
            {hint && <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>{hint}</span>}
        </div>
    )
}

function TextInput({ value, onChange, placeholder, type = 'text' }: { value: string; onChange: (v: string) => void; placeholder?: string; type?: string }) {
    return (
        <input type={type} value={value} onChange={e => onChange(e.target.value)} placeholder={placeholder}
               style={{ width: '100%', padding: '8px 10px', boxSizing: 'border-box', background: 'var(--input-bg)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '13px', fontFamily: 'var(--font-body)', outline: 'none' }} />
    )
}

function TagInput({ values, onChange, placeholder }: { values: string[]; onChange: (v: string[]) => void; placeholder?: string }) {
    const [input, setInput] = useState('')
    function commit() { const t = input.trim(); if (t && !values.includes(t)) onChange([...values, t]); setInput('') }
    return (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '6px', alignItems: 'center', padding: '6px 8px', minHeight: '38px', background: 'var(--input-bg)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)' }}>
            {values.map(v => (
                <span key={v} style={{ display: 'inline-flex', alignItems: 'center', gap: '4px', padding: '2px 8px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', borderRadius: '999px', fontSize: '12px', fontFamily: 'var(--font-body)', fontWeight: 500 }}>
                    {v}
                    <button onClick={() => onChange(values.filter(x => x !== v))} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'inherit', padding: '0 0 0 2px', fontSize: '13px', lineHeight: 1 }}>×</button>
                </span>
            ))}
            <input value={input} onChange={e => setInput(e.target.value)}
                   onKeyDown={e => { if (e.key === 'Enter' || e.key === ',') { e.preventDefault(); commit() } if (e.key === 'Backspace' && !input && values.length) onChange(values.slice(0, -1)) }}
                   onBlur={commit} placeholder={values.length === 0 ? placeholder : 'Add another…'}
                   style={{ flex: 1, minWidth: '100px', background: 'none', border: 'none', outline: 'none', color: 'var(--color-text)', fontSize: '13px', fontFamily: 'var(--font-body)', padding: '2px 4px' }} />
        </div>
    )
}

function ToggleGroup({ options, value, onChange }: { options: { value: string; label: string }[]; value: string; onChange: (v: string) => void }) {
    return (
        <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
            {options.map(opt => (
                <button key={opt.value} onClick={() => onChange(opt.value)} style={{ padding: '6px 14px', border: `1px solid ${value === opt.value ? 'var(--color-primary)' : 'var(--color-border)'}`, borderRadius: 'var(--border-radius)', background: value === opt.value ? 'var(--color-primary)' : 'var(--color-surface-alt)', color: value === opt.value ? 'var(--color-primary-text)' : 'var(--color-text)', fontSize: '12px', fontFamily: 'var(--font-body)', fontWeight: value === opt.value ? 600 : 400, cursor: 'pointer', transition: 'var(--transition)' }}>
                    {opt.label}
                </button>
            ))}
        </div>
    )
}

const LANGUAGES = [
    { code: 'en', label: 'English' }, { code: 'pt', label: 'Portuguese (Portugal)' },
    { code: 'pt-BR', label: 'Portuguese (Brazil)' }, { code: 'es', label: 'Spanish' },
    { code: 'fr', label: 'French' }, { code: 'de', label: 'German' }, { code: 'it', label: 'Italian' },
    { code: 'ja', label: 'Japanese' }, { code: 'zh', label: 'Chinese' }, { code: 'ru', label: 'Russian' },
    { code: 'ar', label: 'Arabic' }, { code: 'nl', label: 'Dutch' }, { code: 'ko', label: 'Korean' },
    { code: 'pl', label: 'Polish' }, { code: 'sv', label: 'Swedish' }, { code: 'other', label: 'Other' },
]

export default function BookDetailPage() {
    const { id }      = useParams<{ id: string }>()
    const [searchParams] = useSearchParams()
    const copyIdParam = searchParams.get('copy_id')   // set when navigating from the library
    const navigate    = useNavigate()
    const queryClient = useQueryClient()
    const { user }    = useAuth()
    const canModerate = user?.role === 'moderator' || user?.role === 'admin'

    const [addOpen, setAddOpen]                     = useState(false)
    const [selectedEditionId, setSelectedEditionId] = useState('')
    const [selectedCondition, setSelectedCondition] = useState('good')
    const [addError, setAddError]                   = useState('')

    const [selectedEditionIdx, setSelectedEditionIdx] = useState(0)   // which edition is shown in detail/edit

    const [editing, setEditing]                     = useState(false)
    const [editTitle, setEditTitle]                 = useState('')
    const [editDesc, setEditDesc]                   = useState('')
    const [editAuthors, setEditAuthors]             = useState<string[]>([])
    const [editGenres, setEditGenres]               = useState<string[]>([])
    const [editFormat, setEditFormat]               = useState('')
    const [editLanguage, setEditLanguage]           = useState('')
    const [editIsbn, setEditIsbn]                   = useState('')
    const [editAsin, setEditAsin]                   = useState('')
    const [editPublisher, setEditPublisher]         = useState('')
    const [editEditionLabel, setEditEditionLabel]   = useState('')
    const [editPublishedAt, setEditPublishedAt]     = useState('')
    const [editPageCount, setEditPageCount]         = useState('')
    const [editFileFormat, setEditFileFormat]       = useState('')
    const [editDuration, setEditDuration]           = useState('')
    const [editAudioFormat, setEditAudioFormat]     = useState('')
    const [editTranslators, setEditTranslators]     = useState<string[]>([])
    const [saveError, setSaveError]                 = useState('')
    const [coverFile, setCoverFile]                 = useState<File | null>(null)
    const [coverPreview, setCoverPreview]           = useState<string | undefined>()
    const fileInputRef = useRef<HTMLInputElement>(null)

    // Copy management (reading progress + ownership) — only when arriving from library
    const [copyEditing, setCopyEditing]             = useState(false)
    const [copyStatus, setCopyStatus]               = useState<'want_to_read'|'reading'|'read'>('want_to_read')
    const [copyPage, setCopyPage]                   = useState('')
    const [copyStarted, setCopyStarted]             = useState('')
    const [copyFinished, setCopyFinished]           = useState('')
    const [copyOwnedByUser, setCopyOwnedByUser]     = useState(true)
    const [copyBorrowedFrom, setCopyBorrowedFrom]   = useState('')
    const [copyLocation, setCopyLocation]           = useState('')
    const [copySaveError, setCopySaveError]         = useState('')
    const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false)
    const [deleteEditionId, setDeleteEditionId]     = useState<string | null>(null)

    const { data: book, isLoading, isError } = useQuery({ queryKey: ['book', id], queryFn: () => getBook(id!), enabled: !!id })

    // Find the user's copy from cache (populated when arriving from library)
    const libraryData = queryClient.getQueryData<{ books: UserBook[] }>(['my-library', 1])
    const userBook = copyIdParam ? libraryData?.books?.find(b => b.copy_id === copyIdParam) : undefined

    function startCopyEditing() {
        if (!userBook) return
        setCopyStatus(userBook.reading_status)
        setCopyPage(userBook.current_page ? String(userBook.current_page) : '')
        setCopyStarted(userBook.started_reading_at ? userBook.started_reading_at.slice(0, 10) : '')
        setCopyFinished(userBook.finished_reading_at ? userBook.finished_reading_at.slice(0, 10) : '')
        setCopyOwnedByUser(userBook.owned_by_user)
        setCopyBorrowedFrom(userBook.borrowed_from ?? '')
        setCopyLocation(userBook.location ?? '')
        setCopySaveError(''); setCopyEditing(true)
    }

    const copySaveMutation = useMutation({
        mutationFn: () => {
            const payload: UpdateCopyPayload = {
                status:             copyStatus,
                current_page:       copyPage ? parseInt(copyPage) : null,
                started_reading_at: copyStarted || null,
                finished_reading_at: copyFinished || null,
                owned_by_user:      copyOwnedByUser,
                borrowed_from:      copyBorrowedFrom || null,
                location:           copyLocation || null,
            }
            return updateReadingStatus(copyIdParam!, payload)
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['my-library'] })
            setCopyEditing(false); setCopySaveError('')
        },
        onError: () => setCopySaveError('Failed to save changes. Please try again.'),
    })

    const removeCopyMutation = useMutation({
        mutationFn: () => removeCopy(copyIdParam!),
        onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['my-library'] }); navigate('/library') },
    })

    const deleteBookMutation = useMutation({
        mutationFn: (force: boolean) => deleteBook(id!, force),
        onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['books'] }); navigate('/books') },
    })

    const deleteEditionMutation = useMutation({
        mutationFn: (editionId: string) => deleteEdition(id!, editionId),
        onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['book', id] }); setDeleteEditionId(null) },
    })

    const addCopyMutation = useMutation({
        mutationFn: () => addCopy(selectedEditionId, { condition: selectedCondition }),
        onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['my-library'] }); setAddOpen(false); setAddError('') },
        onError: () => setAddError('Failed to add to library. Please try again.'),
    })

    const saveMutation = useMutation({
        mutationFn: async () => {
            const ed = book?.editions?.[selectedEditionIdx] ?? book?.editions?.[0]
            let newCoverUrl: string | undefined
            if (coverFile && ed?.id) newCoverUrl = await uploadEditionCover(ed.id, coverFile)
            return updateBook(id!, {
                title:   editTitle.trim() || undefined,
                authors: editAuthors,
                genres:  editGenres,
                edition: ed ? {
                    id:               ed.id,
                    format:           editFormat,
                    description:      editDesc.trim()    || undefined,
                    cover_url:        newCoverUrl,
                    language:         editLanguage,
                    isbn:             editIsbn            || undefined,
                    asin:             editAsin            || undefined,
                    publisher:        editPublisher       || undefined,
                    edition:          editEditionLabel    || undefined,
                    published_at:     editPublishedAt     || undefined,
                    page_count:       editPageCount       ? parseInt(editPageCount) : undefined,
                    file_format:      editFileFormat      || undefined,
                    duration_minutes: editDuration        ? parseInt(editDuration)  : undefined,
                    audio_format:     editAudioFormat     || undefined,
                    translators:      editTranslators,
                } : undefined,
            })
        },
        onSuccess: (updated: Book) => {
            queryClient.setQueryData(['book', id], updated)
            queryClient.invalidateQueries({ queryKey: ['book', id] })
            queryClient.invalidateQueries({ queryKey: ['books'] })
            setEditing(false); setSaveError(''); setCoverFile(null); setCoverPreview(undefined)
        },
        onError: () => setSaveError('Failed to save changes. Please try again.'),
    })

    function startEditing() {
        if (!book) return
        const ed = book.editions?.[selectedEditionIdx] ?? book.editions?.[0]
        setEditTitle(book.title)
        setEditDesc(primaryEdition?.description ?? '')
        setEditAuthors(book.authors?.map(a => a.name) ?? [])
        setEditGenres(book.genres?.map(g => g.name) ?? [])
        setEditFormat(ed?.format ?? '')
        setEditLanguage(ed?.language ?? 'en')
        setEditIsbn(ed?.isbn ?? '')
        setEditAsin(ed?.asin ?? '')
        setEditPublisher(ed?.publisher ?? '')
        setEditEditionLabel(ed?.edition ?? '')
        setEditPublishedAt(ed?.published_at ? String(ed.published_at).slice(0, 4) : '')
        setEditPageCount(ed?.page_count ? String(ed.page_count) : '')
        setEditFileFormat(ed?.file_format ?? '')
        setEditDuration(ed?.duration_minutes ? String(ed.duration_minutes) : '')
        setEditAudioFormat(ed?.audio_format ?? '')
        setEditTranslators(ed?.translators?.map((t: { name: string }) => t.name) ?? [])
        setCoverFile(null); setCoverPreview(undefined); setSaveError(''); setEditing(true)
    }

    function cancelEditing() { setEditing(false); setSaveError(''); setCoverFile(null); setCoverPreview(undefined) }

    function openAddModal() {
        if (!book) return
        setSelectedEditionId(primaryEdition?.id ?? book.editions?.[0]?.id ?? ''); setSelectedCondition('good'); setAddError(''); setAddOpen(true)
    }

    function handleCoverFile(e: React.ChangeEvent<HTMLInputElement>) {
        const file = e.target.files?.[0]; if (!file) return
        setCoverFile(file)
        const reader = new FileReader(); reader.onload = ev => setCoverPreview(ev.target?.result as string); reader.readAsDataURL(file)
    }

    if (isLoading) return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 'calc(100vh - 56px)' }}><Spinner size="lg" label="Loading book..." /></div>
    if (isError || !book) return (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: 'calc(100vh - 56px)', gap: '16px' }}>
            <span style={{ fontSize: '48px' }}>📭</span>
            <h2 style={{ margin: 0, color: 'var(--color-text)', fontFamily: 'var(--font-heading)' }}>Book not found</h2>
            <button onClick={() => navigate('/books')} style={{ background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', padding: '10px 20px', cursor: 'pointer', fontFamily: 'var(--font-body)', fontSize: '14px' }}>Back to catalogue</button>
        </div>
    )

    const coverColor    = pickColor(book.title)
    const primaryEdition = book.editions?.[selectedEditionIdx] ?? book.editions?.[0]
    const storedCover   = primaryEdition?.cover_url
    const displayCover  = coverPreview ?? storedCover
    const isAudiobook   = editing ? editFormat === 'audiobook' : primaryEdition?.format === 'audiobook'

    return (
        <div style={{ minHeight: 'calc(100vh - 56px)', padding: '32px 24px', maxWidth: '900px', margin: '0 auto' }}>
            <button onClick={() => navigate('/books')} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', fontSize: '13px', fontFamily: 'var(--font-body)', padding: 0, marginBottom: '24px', display: 'flex', alignItems: 'center', gap: '6px' }}>← Back to catalogue</button>

            <div style={{ display: 'grid', gridTemplateColumns: '220px 1fr', gap: '32px', alignItems: 'start' }}>
                {/* Left — cover */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <div style={{ position: 'relative' }}>
                        <div style={{ width: '220px', height: '300px', borderRadius: '6px', overflow: 'hidden', boxShadow: 'var(--shadow-lg)', background: displayCover ? undefined : `linear-gradient(135deg, ${coverColor}dd, ${coverColor}88)`, display: 'flex', alignItems: 'center', justifyContent: 'center', border: '1px solid var(--color-border)' }}>
                            {displayCover ? <img src={displayCover} alt={book.title} style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} /> : <span style={{ fontSize: '64px', opacity: 0.5 }}>📖</span>}
                        </div>
                        {canModerate && editing && (
                            <>
                                <button onClick={() => fileInputRef.current?.click()} title="Upload cover image" style={{ position: 'absolute', bottom: '8px', right: '8px', width: '32px', height: '32px', border: 'none', borderRadius: '50%', background: 'rgba(0,0,0,0.6)', color: '#fff', fontSize: '16px', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>📷</button>
                                <input ref={fileInputRef} type="file" accept="image/*" style={{ display: 'none' }} onChange={handleCoverFile} />
                                {coverFile && <p style={{ margin: '6px 0 0', fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', textAlign: 'center' }}>{coverFile.name}</p>}
                            </>
                        )}
                    </div>

                    {/* Edition details (view mode) */}
                    {!editing && primaryEdition && (
                        <Card padding="sm">
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                <p style={{ margin: 0, fontSize: '11px', fontWeight: 700, color: 'var(--color-text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', fontFamily: 'var(--font-body)' }}>Edition Details</p>
                                {[
                                    { label: 'Format',      value: primaryEdition.format },
                                    { label: 'Language',    value: primaryEdition.language },
                                    { label: 'Publisher',   value: primaryEdition.publisher },
                                    { label: 'Published',   value: primaryEdition.published_at ? new Date(primaryEdition.published_at).getFullYear() : undefined },                                    { label: 'Pages',       value: primaryEdition.page_count },
                                    { label: 'ISBN',        value: primaryEdition.isbn },
                                    { label: 'ASIN',        value: primaryEdition.asin },
                                    { label: 'Edition',     value: primaryEdition.edition },
                                    { label: 'File format', value: primaryEdition.file_format },
                                    { label: 'Duration',    value: primaryEdition.duration_minutes ? `${primaryEdition.duration_minutes} min` : undefined },
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

                {/* Right — content */}
                <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                    {/* Toolbar */}
                    <div style={{ display: 'flex', alignItems: 'center', gap: '10px', flexWrap: 'wrap' }}>
                        <Badge label={book.status} variant={book.status === 'approved' ? 'success' : 'warning'} size="sm" />
                        {!editing && book.genres?.map(g => <Badge key={g.id} label={g.name} variant="info" size="sm" />)}
                        {canModerate && !editing && (
                            <button onClick={startEditing} style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '6px', padding: '5px 12px', background: 'var(--color-surface-alt)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '12px', color: 'var(--color-text)', cursor: 'pointer', fontFamily: 'var(--font-body)' }}>✏️ Edit</button>
                        )}
                        {canModerate && editing && (
                            <div style={{ marginLeft: 'auto', display: 'flex', gap: '6px' }}>
                                <button onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending} style={{ padding: '5px 14px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '12px', fontWeight: 600, cursor: saveMutation.isPending ? 'not-allowed' : 'pointer', fontFamily: 'var(--font-body)', opacity: saveMutation.isPending ? 0.7 : 1 }}>
                                    {saveMutation.isPending ? 'Saving…' : '✓ Save'}
                                </button>
                                <button onClick={cancelEditing} style={{ padding: '5px 12px', background: 'var(--color-surface-alt)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '12px', color: 'var(--color-text)', cursor: 'pointer', fontFamily: 'var(--font-body)' }}>Cancel</button>
                            </div>
                        )}
                    </div>

                    {saveError && <p style={{ margin: 0, fontSize: '12px', color: 'var(--color-error)', fontFamily: 'var(--font-body)' }}>{saveError}</p>}

                    {/* View mode */}
                    {!editing && (
                        <>
                            <h1 style={{ margin: 0, fontSize: '28px', fontWeight: 700, color: 'var(--color-text)', fontFamily: 'var(--font-heading)', lineHeight: 1.2 }}>{book.title}</h1>
                            <p style={{ margin: 0, fontSize: '15px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>{book.authors?.map(a => a.name).join(', ') || 'Unknown author'}</p>
                            {primaryEdition?.description && <Card padding="md"><p style={{ margin: 0, fontSize: '14px', color: 'var(--color-text)', lineHeight: '1.7', fontFamily: 'var(--font-body)' }}>{primaryEdition.description}</p></Card>}
                        </>
                    )}

                    {/* Edit mode */}
                    {editing && (
                        <Card padding="lg">
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                                <div>
                                    <FieldLabel label="Title" hint="Required" />
                                    <input value={editTitle} onChange={e => setEditTitle(e.target.value)} style={{ width: '100%', padding: '8px 10px', boxSizing: 'border-box', background: 'var(--input-bg)', border: '1px solid var(--color-primary)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '18px', fontWeight: 700, fontFamily: 'var(--font-heading)', outline: 'none' }} />
                                </div>
                                <div><FieldLabel label="Authors" hint="Enter to add" /><TagInput values={editAuthors} onChange={setEditAuthors} placeholder="e.g. J.R.R. Tolkien" /></div>
                                <div><FieldLabel label="Genres" hint="Enter to add" /><TagInput values={editGenres} onChange={setEditGenres} placeholder="e.g. Fantasy" /></div>
                                <div>
                                    <FieldLabel label="Description" />
                                    <textarea value={editDesc} onChange={e => setEditDesc(e.target.value)} rows={4} placeholder="Description…" style={{ width: '100%', padding: '10px 12px', boxSizing: 'border-box', background: 'var(--input-bg)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '13px', fontFamily: 'var(--font-body)', outline: 'none', resize: 'vertical', lineHeight: '1.6' }} />
                                </div>

                                {primaryEdition && <>
                                    <hr style={{ border: 'none', borderTop: '1px solid var(--color-border)', margin: '4px 0' }} />
                                    <p style={{ margin: 0, fontSize: '11px', fontWeight: 700, color: 'var(--color-text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', fontFamily: 'var(--font-body)' }}>Edition</p>

                                    <div>
                                        <FieldLabel label="Format" />
                                        <ToggleGroup value={editFormat} onChange={setEditFormat} options={[
                                            { value: 'paperback', label: '📗 Paperback' },
                                            { value: 'hardcover', label: '📘 Hardcover' },
                                            { value: 'ebook',     label: '💻 eBook'     },
                                            { value: 'audiobook', label: '🎧 Audiobook' },
                                        ]} />
                                    </div>

                                    <div>
                                        <FieldLabel label="Language" />
                                        <select value={editLanguage} onChange={e => setEditLanguage(e.target.value)} style={{ width: '100%', padding: '8px 10px', background: 'var(--input-bg)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '13px', fontFamily: 'var(--font-body)' }}>
                                            {LANGUAGES.map(l => <option key={l.code} value={l.code}>{l.label}</option>)}
                                        </select>
                                    </div>

                                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                                        <div><FieldLabel label="ISBN" /><TextInput value={editIsbn} onChange={setEditIsbn} placeholder="e.g. 9780756404079" /></div>
                                        <div><FieldLabel label="ASIN" /><TextInput value={editAsin} onChange={setEditAsin} placeholder="e.g. B00AIUUXS4" /></div>
                                    </div>

                                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                                        <div><FieldLabel label="Publisher" /><TextInput value={editPublisher} onChange={setEditPublisher} placeholder="e.g. DAW Books" /></div>
                                        <div><FieldLabel label="Published year" /><TextInput value={editPublishedAt} onChange={setEditPublishedAt} placeholder="e.g. 2007" /></div>
                                    </div>

                                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                                        <div><FieldLabel label="Edition label" /><TextInput value={editEditionLabel} onChange={setEditEditionLabel} placeholder="e.g. 10th Anniversary" /></div>
                                        {!isAudiobook
                                            ? <div><FieldLabel label="Page count" /><TextInput value={editPageCount} onChange={setEditPageCount} placeholder="e.g. 662" type="number" /></div>
                                            : <div><FieldLabel label="Duration (min)" /><TextInput value={editDuration} onChange={setEditDuration} placeholder="e.g. 640" type="number" /></div>
                                        }
                                    </div>

                                    {editFormat === 'ebook' && (
                                        <div>
                                            <FieldLabel label="File format" />
                                            <ToggleGroup value={editFileFormat} onChange={setEditFileFormat} options={[{ value: 'EPUB', label: 'EPUB' }, { value: 'PDF', label: 'PDF' }, { value: 'MOBI', label: 'MOBI' }, { value: 'AZW3', label: 'AZW3' }]} />
                                        </div>
                                    )}

                                    {isAudiobook && (
                                        <div>
                                            <FieldLabel label="Audio format" />
                                            <ToggleGroup value={editAudioFormat} onChange={setEditAudioFormat} options={[{ value: 'MP3', label: 'MP3' }, { value: 'AAC', label: 'AAC' }, { value: 'FLAC', label: 'FLAC' }, { value: 'WMA', label: 'WMA' }]} />
                                        </div>
                                    )}

                                    {!isAudiobook && (
                                        <div><FieldLabel label="Translators" hint="Enter to add" /><TagInput values={editTranslators} onChange={setEditTranslators} placeholder="e.g. Margaret Jull Costa" /></div>
                                    )}
                                </>}
                            </div>
                        </Card>
                    )}

                    {/* Edition selector — shown when there are multiple editions */}
                    {!editing && book.editions && book.editions.length > 1 && (
                        <div>
                            <h3 style={{ margin: '0 0 10px', fontSize: '13px', fontWeight: 700, color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>
                                Editions ({book.editions.length})
                            </h3>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                {book.editions.map((edition: Edition, idx: number) => {
                                    const isSelected = idx === selectedEditionIdx
                                    return (
                                        <button
                                            key={edition.id}
                                            onClick={() => setSelectedEditionIdx(idx)}
                                            style={{
                                                display: 'flex', alignItems: 'center', gap: '8px',
                                                padding: '8px 12px', width: '100%', textAlign: 'left', cursor: 'pointer',
                                                background: isSelected ? 'var(--color-primary)' : 'var(--color-surface-alt)',
                                                color: isSelected ? 'var(--color-primary-text)' : 'var(--color-text)',
                                                border: `1px solid ${isSelected ? 'var(--color-primary)' : 'var(--color-border)'}`,
                                                borderRadius: 'var(--border-radius)',
                                                fontFamily: 'var(--font-body)', fontSize: '12px', fontWeight: isSelected ? 600 : 400,
                                                transition: 'var(--transition)',
                                            }}
                                        >
                                            <span style={{ fontWeight: 600, textTransform: 'capitalize' }}>{edition.format}</span>
                                            {edition.language && <span style={{ opacity: 0.85 }}>{edition.language.toUpperCase()}</span>}
                                            {edition.published_at && <span style={{ opacity: 0.75 }}>{new Date(edition.published_at).getFullYear()}</span>}
                                            {edition.publisher && <span style={{ opacity: 0.75, flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{edition.publisher}</span>}
                                            {edition.isbn && <span style={{ opacity: 0.6, marginLeft: 'auto', flexShrink: 0, fontSize: '11px' }}>{edition.isbn}</span>}
                                        </button>
                                    )
                                })}
                            </div>
                        </div>
                    )}

                    {/* Action buttons (view mode) */}
                    {!editing && (
                        <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
                            {!copyIdParam && <button onClick={openAddModal} style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 20px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', transition: 'var(--transition)', fontFamily: 'var(--font-body)' }} onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary-hover)'} onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary)'}>➕ Add to my library</button>}
                            <button onClick={() => navigate(copyIdParam ? '/library' : '/books')} style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 20px', background: 'var(--color-surface-alt)', color: 'var(--color-text)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', transition: 'var(--transition)', fontFamily: 'var(--font-body)' }}>← {copyIdParam ? 'Back to library' : 'Back'}</button>
                            {canModerate && (
                                <button onClick={() => setDeleteConfirmOpen(true)} style={{ marginLeft: 'auto', padding: '10px 20px', background: 'transparent', color: 'var(--color-error, #dc2626)', border: '1px solid var(--color-error, #dc2626)', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', fontFamily: 'var(--font-body)' }}>🗑 Delete book</button>
                            )}
                        </div>
                    )}

                    {/* ── Copy management panel (library context) ── */}
                    {userBook && !editing && (
                        <Card padding="lg">
                            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px' }}>
                                <h3 style={{ margin: 0, fontSize: '14px', fontWeight: 700, color: 'var(--color-text)', fontFamily: 'var(--font-heading)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>My Copy</h3>
                                <div style={{ display: 'flex', gap: '8px' }}>
                                    {!copyEditing && <button onClick={startCopyEditing} style={{ padding: '5px 12px', background: 'var(--color-surface-alt)', color: 'var(--color-text)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '12px', fontWeight: 600, cursor: 'pointer', fontFamily: 'var(--font-body)' }}>Edit</button>}
                                    <button onClick={() => removeCopyMutation.mutate()} disabled={removeCopyMutation.isPending} style={{ padding: '5px 12px', background: 'transparent', color: 'var(--color-error, #dc2626)', border: '1px solid var(--color-error, #dc2626)', borderRadius: 'var(--border-radius)', fontSize: '12px', fontWeight: 600, cursor: 'pointer', fontFamily: 'var(--font-body)' }}>Remove from library</button>
                                </div>
                            </div>
                            {copyEditing ? (
                                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                                    <div>
                                        <FieldLabel label="Reading status" />
                                        <ToggleGroup value={copyStatus} onChange={v => setCopyStatus(v as typeof copyStatus)} options={[{ value: 'want_to_read', label: 'Want to read' }, { value: 'reading', label: 'Reading' }, { value: 'read', label: 'Read' }]} />
                                    </div>
                                    {copyStatus === 'reading' && (
                                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                                            <div><FieldLabel label="Current page" hint="Optional" /><TextInput value={copyPage} onChange={setCopyPage} placeholder="e.g. 142" type="number" /></div>
                                            <div><FieldLabel label="Started reading" /><TextInput value={copyStarted} onChange={setCopyStarted} placeholder="YYYY-MM-DD" type="date" /></div>
                                        </div>
                                    )}
                                    {copyStatus === 'read' && (
                                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                                            <div><FieldLabel label="Started reading" /><TextInput value={copyStarted} onChange={setCopyStarted} placeholder="YYYY-MM-DD" type="date" /></div>
                                            <div><FieldLabel label="Finished reading" /><TextInput value={copyFinished} onChange={setCopyFinished} placeholder="YYYY-MM-DD" type="date" /></div>
                                        </div>
                                    )}
                                    <div>
                                        <FieldLabel label="Ownership" />
                                        <ToggleGroup value={copyOwnedByUser ? 'owned' : 'borrowed'} onChange={v => setCopyOwnedByUser(v === 'owned')} options={[{ value: 'owned', label: '📦 I own it' }, { value: 'borrowed', label: '🤝 Borrowed' }]} />
                                    </div>
                                    {!copyOwnedByUser && (
                                        <div><FieldLabel label="Borrowed from (user ID)" hint="Optional" /><TextInput value={copyBorrowedFrom} onChange={setCopyBorrowedFrom} placeholder="User ID of the lender" /></div>
                                    )}
                                    <div><FieldLabel label="Location" hint="Optional — e.g. shelf, box, on loan" /><TextInput value={copyLocation} onChange={setCopyLocation} placeholder="e.g. Living room shelf" /></div>
                                    {copySaveError && <p style={{ margin: 0, fontSize: '12px', color: 'var(--color-error)', fontFamily: 'var(--font-body)' }}>{copySaveError}</p>}
                                    <div style={{ display: 'flex', gap: '8px' }}>
                                        <button onClick={() => copySaveMutation.mutate()} disabled={copySaveMutation.isPending} style={{ padding: '8px 16px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', fontFamily: 'var(--font-body)' }}>{copySaveMutation.isPending ? 'Saving…' : '✓ Save'}</button>
                                        <button onClick={() => setCopyEditing(false)} style={{ padding: '8px 16px', background: 'var(--color-surface-alt)', color: 'var(--color-text)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '13px', cursor: 'pointer', fontFamily: 'var(--font-body)' }}>Cancel</button>
                                    </div>
                                </div>
                            ) : (
                                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: '12px' }}>
                                    <div>
                                        <FieldLabel label="Status" />
                                        <Badge label={{ want_to_read: 'Want to read', reading: 'Reading', read: 'Read' }[userBook.reading_status] ?? userBook.reading_status} variant={({ want_to_read: 'default', reading: 'warning', read: 'success' } as Record<string, 'default' | 'warning' | 'success'>)[userBook.reading_status] ?? 'default'} size="sm" />
                                    </div>
                                    {userBook.reading_status === 'reading' && userBook.current_page && (
                                        <div><FieldLabel label="Current page" /><span style={{ fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>p. {userBook.current_page}</span></div>
                                    )}
                                    {(userBook.reading_status === 'reading' || userBook.reading_status === 'read') && userBook.started_reading_at && (
                                        <div><FieldLabel label="Started" /><span style={{ fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>{new Date(userBook.started_reading_at).toLocaleDateString()}</span></div>
                                    )}
                                    {userBook.reading_status === 'read' && userBook.finished_reading_at && (
                                        <div><FieldLabel label="Finished" /><span style={{ fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>{new Date(userBook.finished_reading_at).toLocaleDateString()}</span></div>
                                    )}
                                    <div>
                                        <FieldLabel label="Ownership" />
                                        <span style={{ fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>{userBook.owned_by_user ? '📦 Owned' : '🤝 Borrowed'}</span>
                                    </div>
                                    {!userBook.owned_by_user && userBook.borrowed_from && (
                                        <div><FieldLabel label="Borrowed from" /><span style={{ fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>{userBook.borrowed_from}</span></div>
                                    )}
                                    {userBook.location && (
                                        <div><FieldLabel label="Location" /><span style={{ fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>{userBook.location}</span></div>
                                    )}
                                    {userBook.condition && (
                                        <div><FieldLabel label="Condition" /><span style={{ fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)', textTransform: 'capitalize' }}>{userBook.condition}</span></div>
                                    )}
                                </div>
                            )}
                        </Card>
                    )}

                    {/* Admin: edition management */}
                    {canModerate && !editing && book.editions && book.editions.length > 0 && (
                        <div>
                            <h3 style={{ margin: '0 0 10px', fontSize: '13px', fontWeight: 700, color: 'var(--color-text-muted)', fontFamily: 'var(--font-heading)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>Edition Management</h3>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                                {book.editions.map((ed: Edition) => (
                                    <div key={ed.id} style={{ display: 'flex', alignItems: 'center', gap: '10px', padding: '8px 12px', background: 'var(--color-surface-alt)', borderRadius: 'var(--border-radius)', border: '1px solid var(--color-border)' }}>
                                        <Badge label={ed.format} variant="default" size="sm" />
                                        <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', flex: 1 }}>{[ed.language?.toUpperCase(), ed.publisher, ed.isbn].filter(Boolean).join(' · ')}</span>
                                        <button onClick={() => setDeleteEditionId(ed.id)} style={{ padding: '3px 8px', background: 'transparent', color: 'var(--color-error, #dc2626)', border: '1px solid var(--color-error, #dc2626)', borderRadius: 'var(--border-radius)', fontSize: '11px', cursor: 'pointer', fontFamily: 'var(--font-body)' }}>Delete</button>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {/* Add to library modal */}
            <Modal isOpen={addOpen} onClose={() => setAddOpen(false)} title={`Add to library — ${book.title}`} confirmLabel="Add to library" onConfirm={() => addCopyMutation.mutate()} isLoading={addCopyMutation.isPending} size="sm">
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <div>
                        <label style={{ display: 'block', fontSize: '11px', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', marginBottom: '6px' }}>Choose Edition</label>
                        <select value={selectedEditionId} onChange={e => setSelectedEditionId(e.target.value)} style={{ width: '100%', padding: '8px 10px', background: 'var(--input-bg)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '13px', fontFamily: 'var(--font-body)' }}>
                            {book.editions?.map((ed: Edition) => <option key={ed.id} value={ed.id}>{ed.format}{ed.language ? ` · ${ed.language.toUpperCase()}` : ''}{ed.publisher ? ` · ${ed.publisher}` : ''}{ed.isbn ? ` · ${ed.isbn}` : ''}</option>)}
                        </select>
                        <p style={{ margin: '6px 0 0', fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                            Don't see your edition?{' '}
                            <button onClick={() => { setAddOpen(false); navigate('/books/add') }} style={{ background: 'none', border: 'none', color: 'var(--color-primary)', fontSize: '11px', cursor: 'pointer', fontFamily: 'var(--font-body)', padding: 0, textDecoration: 'underline' }}>Add a new edition</button>
                        </p>
                    </div>
                    <div>
                        <label style={{ display: 'block', fontSize: '11px', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', marginBottom: '6px' }}>Condition</label>
                        <div style={{ display: 'flex', gap: '6px' }}>
                            {['new', 'good', 'fair', 'poor'].map(c => <button key={c} onClick={() => setSelectedCondition(c)} style={{ flex: 1, padding: '7px 0', border: `1px solid ${selectedCondition === c ? 'var(--color-primary)' : 'var(--color-border)'}`, borderRadius: 'var(--border-radius)', background: selectedCondition === c ? 'var(--color-primary)' : 'var(--color-surface-alt)', color: selectedCondition === c ? 'var(--color-primary-text)' : 'var(--color-text)', fontSize: '12px', fontFamily: 'var(--font-body)', fontWeight: selectedCondition === c ? 600 : 400, cursor: 'pointer', textTransform: 'capitalize', transition: 'var(--transition)' }}>{c}</button>)}
                        </div>
                    </div>
                    {addError && <p style={{ margin: 0, fontSize: '12px', color: 'var(--color-error)', fontFamily: 'var(--font-body)' }}>{addError}</p>}
                </div>
            </Modal>

            {/* Delete book confirmation */}
            <Modal isOpen={deleteConfirmOpen} onClose={() => setDeleteConfirmOpen(false)} title="Delete book" confirmLabel="Delete (force)" onConfirm={() => { deleteBookMutation.mutate(true); setDeleteConfirmOpen(false) }} isLoading={deleteBookMutation.isPending} size="sm">
                <p style={{ margin: 0, fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>
                    This will soft-delete <strong>{book.title}</strong> and all its editions and copies. Members who own this edition will lose it from their library.
                </p>
            </Modal>

            {/* Delete edition confirmation */}
            <Modal isOpen={!!deleteEditionId} onClose={() => setDeleteEditionId(null)} title="Delete edition" confirmLabel="Delete edition" onConfirm={() => deleteEditionId && deleteEditionMutation.mutate(deleteEditionId)} isLoading={deleteEditionMutation.isPending} size="sm">
                <p style={{ margin: 0, fontSize: '13px', color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>
                    This will soft-delete the selected edition and all its copies. Members who own this edition will lose it from their library.
                </p>
            </Modal>
        </div>
    )
}