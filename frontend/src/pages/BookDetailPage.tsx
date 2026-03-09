import { useRef, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getBook, addCopy, updateBook, uploadBookCover } from '../api/books'
import { Badge, Card, Modal, Spinner } from '../components'
import { useAuth } from '../context/AuthContext'
import type { Book, Edition } from '../types'

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
    const navigate    = useNavigate()
    const queryClient = useQueryClient()
    const { user }    = useAuth()
    const canModerate = user?.role === 'moderator' || user?.role === 'admin'

    const [addOpen, setAddOpen]                     = useState(false)
    const [selectedEditionId, setSelectedEditionId] = useState('')
    const [selectedCondition, setSelectedCondition] = useState('good')
    const [addError, setAddError]                   = useState('')

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

    const { data: book, isLoading, isError } = useQuery({ queryKey: ['book', id], queryFn: () => getBook(id!), enabled: !!id })

    const addCopyMutation = useMutation({
        mutationFn: () => addCopy(selectedEditionId, { condition: selectedCondition }),
        onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['my-library'] }); setAddOpen(false); setAddError('') },
        onError: () => setAddError('Failed to add to library. Please try again.'),
    })

    const saveMutation = useMutation({
        mutationFn: async () => {
            let newCoverUrl: string | undefined
            if (coverFile && id) newCoverUrl = await uploadBookCover(id, coverFile)
            const ed = book?.editions?.[0]
            return updateBook(id!, {
                title:       editTitle.trim() || undefined,
                description: editDesc.trim()  || undefined,
                cover_url:   newCoverUrl,
                authors:     editAuthors,
                genres:      editGenres,
                edition: ed ? {
                    id:               ed.id,
                    format:           editFormat,
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
            queryClient.invalidateQueries({ queryKey: ['books'] })
            setEditing(false); setSaveError(''); setCoverFile(null); setCoverPreview(undefined)
        },
        onError: () => setSaveError('Failed to save changes. Please try again.'),
    })

    function startEditing() {
        if (!book) return
        const ed = book.editions?.[0]
        setEditTitle(book.title)
        setEditDesc(book.description ?? '')
        setEditAuthors(book.authors?.map(a => a.name) ?? [])
        setEditGenres(book.genres?.map(g => g.name) ?? [])
        setEditFormat(ed?.format ?? '')
        setEditLanguage(ed?.language ?? 'en')
        setEditIsbn(ed?.isbn ?? '')
        setEditAsin((ed as any)?.asin ?? '')
        setEditPublisher(ed?.publisher ?? '')
        setEditEditionLabel((ed as any)?.edition ?? '')
        setEditPublishedAt(ed?.published_at ? String(ed.published_at).slice(0, 4) : '')
        setEditPageCount(ed?.page_count ? String(ed.page_count) : '')
        setEditFileFormat((ed as any)?.file_format ?? '')
        setEditDuration((ed as any)?.duration_minutes ? String((ed as any).duration_minutes) : '')
        setEditAudioFormat((ed as any)?.audio_format ?? '')
        setEditTranslators((ed as any)?.translators?.map((t: any) => t.name) ?? [])
        setCoverFile(null); setCoverPreview(undefined); setSaveError(''); setEditing(true)
    }

    function cancelEditing() { setEditing(false); setSaveError(''); setCoverFile(null); setCoverPreview(undefined) }

    function openAddModal() {
        if (!book) return
        setSelectedEditionId(book.editions?.[0]?.id ?? ''); setSelectedCondition('good'); setAddError(''); setAddOpen(true)
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
    const storedCover   = (book as any).cover_image_url ?? book.cover_url
    const displayCover  = coverPreview ?? storedCover
    const primaryEdition = book.editions?.[0]
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
                                    { label: 'Published',   value: primaryEdition.published_at },
                                    { label: 'Pages',       value: primaryEdition.page_count },
                                    { label: 'ISBN',        value: primaryEdition.isbn },
                                    { label: 'ASIN',        value: (primaryEdition as any).asin },
                                    { label: 'Edition',     value: (primaryEdition as any).edition },
                                    { label: 'File format', value: (primaryEdition as any).file_format },
                                    { label: 'Duration',    value: (primaryEdition as any).duration_minutes ? `${(primaryEdition as any).duration_minutes} min` : undefined },
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
                            {book.description && <Card padding="md"><p style={{ margin: 0, fontSize: '14px', color: 'var(--color-text)', lineHeight: '1.7', fontFamily: 'var(--font-body)' }}>{book.description}</p></Card>}
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

                    {/* All editions (view mode) */}
                    {!editing && book.editions && book.editions.length > 1 && (
                        <div>
                            <h3 style={{ margin: '0 0 12px', fontSize: '14px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-heading)' }}>All Editions ({book.editions.length})</h3>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                                {book.editions.map((edition: Edition) => (
                                    <Card key={edition.id} padding="sm">
                                        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
                                            <Badge label={edition.format} variant="default" size="sm" />
                                            {edition.language && <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>{edition.language.toUpperCase()}</span>}
                                            {edition.publisher && <span style={{ fontSize: '12px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>{edition.publisher}</span>}
                                            {edition.isbn && <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', marginLeft: 'auto' }}>ISBN: {edition.isbn}</span>}
                                            <button onClick={() => { setSelectedEditionId(edition.id); setSelectedCondition('good'); setAddError(''); setAddOpen(true) }} style={{ marginLeft: 'auto', padding: '4px 10px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '11px', fontWeight: 600, cursor: 'pointer', fontFamily: 'var(--font-body)' }}>+ Add</button>
                                        </div>
                                    </Card>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* Action buttons (view mode) */}
                    {!editing && (
                        <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
                            <button onClick={openAddModal} style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 20px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', transition: 'var(--transition)', fontFamily: 'var(--font-body)' }} onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary-hover)'} onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary)'}>➕ Add to my library</button>
                            <button onClick={() => navigate('/books')} style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 20px', background: 'var(--color-surface-alt)', color: 'var(--color-text)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', transition: 'var(--transition)', fontFamily: 'var(--font-body)' }}>← Back</button>
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
        </div>
    )
}