import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import apiClient from '../api/client'
import { Card, Badge, Spinner } from '../components'
import { lookupByISBN, lookupByTitleAuthor, checkDuplicate, submitBook as submitBookApi } from '../api/books'

// ── Types ─────────────────────────────────────────────────────────────────────

interface BookLookup {
    title: string
    authors?: string[]
    publisher?: string
    published_date?: string
    description?: string
    isbn_10?: string
    isbn_13?: string
    page_count?: number
    language?: string
    cover_url?: string
    categories?: string[]
}

interface ExistingEdition {
    id: string
    book_id: string
    format: string
    isbn?: string
    language?: string
    publisher?: string
}

// 'mode' drives the entire page: book or audiobook
type Mode       = 'book' | 'audiobook'
type Step = 'pick' | 'search' | 'preview' | 'manual' | 'details' | 'done'
type SearchMode = 'isbn' | 'title'

// ── Helpers ───────────────────────────────────────────────────────────────────

function getIsbn(result: BookLookup): string | undefined {
    return result.isbn_13 ?? result.isbn_10
}

async function addCopyToLibrary(
    editionId: string,
    opts?: { condition?: string; narrator?: string; duration?: number }
): Promise<void> {
    await apiClient.post('/books/copies', { edition_id: editionId, ...opts })
}

// ── Step indicator ────────────────────────────────────────────────────────────

const STEPS: { key: Step; label: string }[] = [
    { key: 'pick',    label: 'Type'    },
    { key: 'search',  label: 'Search'  },
    { key: 'preview', label: 'Preview' },
    { key: 'details', label: 'Details' },
    { key: 'done',    label: 'Done'    },
]

function StepIndicator({ current }: { current: Step }) {
    const currentIndex = STEPS.findIndex(s => s.key === current)
    return (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '28px' }}>
            {STEPS.map((step, i) => {
                const done   = i < currentIndex
                const active = step.key === current
                return (
                    <div key={step.key} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <div style={{
                            width: '26px', height: '26px', borderRadius: '50%', flexShrink: 0,
                            background: done ? 'var(--color-success)' : active ? 'var(--color-primary)' : 'var(--color-surface-alt)',
                            border: `2px solid ${active ? 'var(--color-primary)' : done ? 'var(--color-success)' : 'var(--color-border)'}`,
                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                            fontSize: '11px', fontWeight: 700,
                            color: (done || active) ? '#fff' : 'var(--color-text-muted)',
                            transition: 'var(--transition)',
                        }}>
                            {done ? '✓' : i + 1}
                        </div>
                        <span style={{
                            fontSize: '12px', fontFamily: 'var(--font-body)',
                            color: active ? 'var(--color-text)' : 'var(--color-text-muted)',
                            fontWeight: active ? 600 : 400,
                        }}>
              {step.label}
            </span>
                        {i < STEPS.length - 1 && (
                            <div style={{ width: '16px', height: '1px', background: 'var(--color-border)', marginLeft: '4px' }} />
                        )}
                    </div>
                )
            })}
        </div>
    )
}

// ── Shared UI components ──────────────────────────────────────────────────────

function PrimaryButton({ label, onClick, isLoading, disabled, fullWidth }: {
    label: string; onClick?: () => void; isLoading?: boolean; disabled?: boolean; fullWidth?: boolean
}) {
    return (
        <button onClick={onClick} disabled={isLoading || disabled} style={{
            padding: '10px 20px',
            background: (isLoading || disabled) ? 'var(--color-surface-alt)' : 'var(--color-primary)',
            color: (isLoading || disabled) ? 'var(--color-text-muted)' : 'var(--color-primary-text)',
            border: 'none', borderRadius: 'var(--border-radius)',
            fontSize: '13px', fontWeight: 600, fontFamily: 'var(--font-body)',
            cursor: (isLoading || disabled) ? 'not-allowed' : 'pointer',
            transition: 'var(--transition)',
            width: fullWidth ? '100%' : undefined,
            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px',
        }}>
            {isLoading && <Spinner size="sm" />}
            {label}
        </button>
    )
}

function SecondaryButton({ label, onClick }: { label: string; onClick: () => void }) {
    return (
        <button onClick={onClick} style={{
            padding: '10px 18px',
            background: 'var(--color-surface-alt)', color: 'var(--color-text)',
            border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)',
            fontSize: '13px', fontWeight: 500, fontFamily: 'var(--font-body)',
            cursor: 'pointer', transition: 'var(--transition)',
        }}>
            {label}
        </button>
    )
}

function ToggleGroup({ options, value, onChange }: {
    options: { value: string; label: string; icon?: string }[]
    value: string
    onChange: (v: string) => void
}) {
    return (
        <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            {options.map(opt => (
                <button key={opt.value} onClick={() => onChange(opt.value)} style={{
                    padding: '8px 16px',
                    border: `1px solid ${value === opt.value ? 'var(--color-primary)' : 'var(--color-border)'}`,
                    borderRadius: 'var(--border-radius)',
                    background: value === opt.value ? 'var(--color-primary)' : 'var(--color-surface-alt)',
                    color: value === opt.value ? 'var(--color-primary-text)' : 'var(--color-text)',
                    fontSize: '13px', fontFamily: 'var(--font-body)',
                    fontWeight: value === opt.value ? 600 : 400,
                    cursor: 'pointer', transition: 'var(--transition)',
                }}>
                    {opt.icon && <span style={{ marginRight: '6px' }}>{opt.icon}</span>}
                    {opt.label}
                </button>
            ))}
        </div>
    )
}

function FieldLabel({ label, hint }: { label: string; hint?: string }) {
    return (
        <div style={{ marginBottom: '8px' }}>
            <label style={{
                display: 'block', fontSize: '11px', fontWeight: 700,
                color: 'var(--color-text-muted)',
                fontFamily: 'var(--font-body)', textTransform: 'uppercase', letterSpacing: '0.06em',
            }}>
                {label}
            </label>
            {hint && (
                <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
          {hint}
        </span>
            )}
        </div>
    )
}

function TextInput({ value, onChange, placeholder, type = 'text' }: {
    value: string; onChange: (v: string) => void; placeholder?: string; type?: string
}) {
    return (
        <input
            type={type}
            value={value}
            onChange={e => onChange(e.target.value)}
            placeholder={placeholder}
            style={{
                width: '100%', padding: '10px 12px', boxSizing: 'border-box',
                background: 'var(--input-bg)', border: '1px solid var(--color-border)',
                borderRadius: 'var(--border-radius)', color: 'var(--color-text)',
                fontSize: '13px', fontFamily: 'var(--font-body)', outline: 'none',
            }}
        />
    )
}

function CoverPreview({ coverUrl, title, size = 'md' }: {
    coverUrl?: string; title?: string; size?: 'sm' | 'md' | 'lg'
}) {
    const dims: Record<string, [number, number]> = { sm: [56, 76], md: [100, 140], lg: [160, 220] }
    const [w, h] = dims[size]
    const colors = ['#2563eb','#16a34a','#dc2626','#9333ea','#ea580c','#0891b2','#d97706','#4f46e5']
    const safeTitle = title ?? ''
    const color = colors[safeTitle.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % colors.length]
    return (
        <div style={{
            width: w, height: h, borderRadius: '4px', flexShrink: 0,
            background: coverUrl
                ? `url(${coverUrl}) center/cover no-repeat`
                : `linear-gradient(135deg, ${color}dd, ${color}88)`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            border: '1px solid var(--color-border)',
            fontSize: size === 'lg' ? '48px' : '24px',
            boxShadow: 'var(--shadow-sm)',
        }}>
            {!coverUrl && '📖'}
        </div>
    )
}

function ErrorBox({ message }: { message: string }) {
    return (
        <div style={{
            padding: '10px 14px',
            background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.2)',
            borderRadius: 'var(--border-radius)',
        }}>
            <p style={{ margin: 0, fontSize: '13px', color: 'var(--color-error)', fontFamily: 'var(--font-body)' }}>
                {message}
            </p>
        </div>
    )
}

// ── Main page ─────────────────────────────────────────────────────────────────

export default function AddBookPage() {
    const navigate = useNavigate()

    const [mode, setMode]               = useState<Mode>('book')
    const [step, setStep]               = useState<Step>('pick')
    const [searchMode, setSearchMode]   = useState<SearchMode>('isbn')
    const [isbn, setIsbn]               = useState('')
    const [titleQuery, setTitleQuery]   = useState('')
    const [authorQuery, setAuthorQuery] = useState('')
    const [searchError, setSearchError] = useState('')

    const [lookupResult, setLookupResult]       = useState<BookLookup | null>(null)
    const [existingEdition, setExistingEdition] = useState<ExistingEdition | null>(null)

    // Book-specific
    const [format, setFormat]       = useState<'hardcover' | 'paperback' | 'ebook'>('paperback')
    const [condition, setCondition] = useState('good')

    // Audiobook-specific
    const [narrator, setNarrator]   = useState('')
    const [duration, setDuration]   = useState('')  // minutes as string

    // Shared
    const [language, setLanguage]         = useState('en')
    const [detailsError, setDetailsError] = useState('')
    const [doneMessage, setDoneMessage]   = useState('')

    const [manualTitle,     setManualTitle]     = useState('')
    const [manualAuthors,   setManualAuthors]   = useState('')
    const [manualIsbn,      setManualIsbn]      = useState('')
    const [manualPublisher, setManualPublisher] = useState('')
    const [manualPages,     setManualPages]     = useState('')
    const [manualDesc,      setManualDesc]      = useState('')
    const [manualError,     setManualError]     = useState('')

    // ── Search ──────────────────────────────────────────────────────────────────

    const searchMutation = useMutation({
        mutationFn: async (): Promise<BookLookup | null> => {
            setSearchError('')
            if (searchMode === 'isbn') {
                const cleanIsbn = isbn.replace(/[-\s]/g, '')
                const dupCheck = await checkDuplicate(cleanIsbn)
                setExistingEdition(
                    dupCheck.exists && dupCheck.edition ? dupCheck.edition as ExistingEdition : null
                )
                const result = await lookupByISBN(cleanIsbn)
                return result as unknown as BookLookup | null
            } else {
                setExistingEdition(null)
                // Don't send empty author — causes 500 on backend
                const author = authorQuery.trim() || ''
                const result = await lookupByTitleAuthor(titleQuery.trim(), author)
                return result as unknown as BookLookup | null
            }
        },
        onSuccess: (result) => {
            if (!result) {
                // Pre-fill manual form with what user typed, go to manual entry
                if (searchMode === 'isbn') setManualIsbn(isbn.replace(/[-\s]/g, ''))
                else { setManualTitle(titleQuery); setManualAuthors(authorQuery) }
                setStep('manual')
                return
            }
            setLookupResult(result)
            setLanguage(result.language ?? 'en')
            setStep('preview')
        },
        onError: () => {
            // API error — also offer manual entry
            if (searchMode === 'isbn') setManualIsbn(isbn.replace(/[-\s]/g, ''))
            else { setManualTitle(titleQuery); setManualAuthors(authorQuery) }
            setSearchError('Nothing found. Fill in the details manually below.')
        },
    })

    // ── Add copy (edition already exists) ───────────────────────────────────────

    const addCopyMutation = useMutation({
        mutationFn: () => addCopyToLibrary(existingEdition!.id, {
            condition: mode === 'book' && format !== 'ebook' ? condition : undefined,
            narrator:  mode === 'audiobook' ? narrator : undefined,
            duration:  mode === 'audiobook' && duration ? parseInt(duration) : undefined,
        }),
        onSuccess: () => { setDoneMessage('Added to your library!'); setStep('done') },
        onError:   () => setDetailsError('Failed to add copy. Please try again.'),
    })

    // ── Submit new book ──────────────────────────────────────────────────────────

    const submitMutation = useMutation({
        mutationFn: () => {
            if (!lookupResult) throw new Error('No book data')
            const dbFormat = mode === 'audiobook' ? 'audiobook' : format  // hardcover | paperback | ebook
            return submitBookApi({
                title:       lookupResult.title,
                authors:     lookupResult.authors ?? [],
                description: lookupResult.description,
                cover_url:   lookupResult.cover_url,
                genres:      [],
                edition: {
                    format:           dbFormat,
                    isbn:             getIsbn(lookupResult),
                    language,
                    publisher:        lookupResult.publisher,
                    page_count:       mode === 'book' ? lookupResult.page_count : undefined,
                    duration_minutes: mode === 'audiobook' && duration ? parseInt(duration) : undefined,
                    narrator:         mode === 'audiobook' ? narrator : undefined,
                },
                condition: (mode === 'book' && format !== 'ebook') ? condition : undefined,
            })
        },
        onSuccess: () => {
            setDoneMessage('Book submitted! It will appear in the catalogue once a moderator approves it.')
            setStep('done')
        },
        onError: () => setDetailsError('Submission failed. Please try again.'),
    })

    const submitManualMutation = useMutation({
        mutationFn: () => {
            const authors = manualAuthors.split(',').map(a => a.trim()).filter(Boolean)
            return submitBookApi({
                title:       manualTitle.trim(),
                authors,
                description: manualDesc || undefined,
                cover_url:   undefined,
                genres:      [],
                edition: {
                    format:           mode === 'audiobook' ? 'audiobook' : format,
                    isbn:             manualIsbn || undefined,
                    language,
                    publisher:        manualPublisher || undefined,
                    page_count:       mode === 'book' && manualPages ? parseInt(manualPages) : undefined,
                    duration_minutes: mode === 'audiobook' && duration ? parseInt(duration) : undefined,
                    narrator:         mode === 'audiobook' ? narrator || undefined : undefined,
                },
                condition: (mode === 'book' && format !== 'ebook') ? condition : undefined,
            })
        },
        onSuccess: () => {
            setDoneMessage('Book submitted! It will appear in the catalogue once a moderator approves it.')
            setStep('done')
        },
        onError: () => setManualError('Submission failed. Please try again.'),
    })

    // ── Reset ────────────────────────────────────────────────────────────────────

    function reset() {
        setStep('pick')
        setIsbn(''); setTitleQuery(''); setAuthorQuery(''); setSearchError('')
        setLookupResult(null); setExistingEdition(null)
        setFormat('paperback'); setCondition('good')
        setNarrator(''); setDuration('')
        setLanguage('en'); setDetailsError(''); setDoneMessage('')
        setManualTitle(''); setManualAuthors(''); setManualIsbn('')
        setManualPublisher(''); setManualPages(''); setManualDesc(''); setManualError('')
    }

    // ── Render ───────────────────────────────────────────────────────────────────

    const isAudiobook = mode === 'audiobook'

    return (
        <div style={{
            minHeight: 'calc(100vh - 56px)',
            padding: '32px 24px',
            maxWidth: '680px',
            margin: '0 auto',
        }}>
            {/* Header */}
            <div style={{ marginBottom: '24px' }}>
                <button onClick={() => navigate('/books')} style={{
                    background: 'none', border: 'none', cursor: 'pointer',
                    color: 'var(--color-text-muted)', fontSize: '13px',
                    fontFamily: 'var(--font-body)', padding: 0, marginBottom: '12px',
                }}>
                    ← Back to catalogue
                </button>
                <h1 style={{
                    margin: 0, fontSize: '24px', fontWeight: 700,
                    color: 'var(--color-text)', fontFamily: 'var(--font-heading)',
                }}>
                    Add to my Library
                </h1>
                <p style={{
                    margin: '6px 0 0', fontSize: '13px',
                    color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                }}>
                    Choose what you're adding, then search for it.
                </p>
            </div>

            <StepIndicator current={step} />

            {/* ── STEP: Pick mode ── */}
            {step === 'pick' && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    {/* Book card */}
                    <button
                        onClick={() => { setMode('book'); setStep('search') }}
                        style={{
                            display: 'flex', alignItems: 'center', gap: '20px',
                            padding: '24px', textAlign: 'left',
                            background: 'var(--color-surface)', border: '2px solid var(--color-border)',
                            borderRadius: 'var(--border-radius)', cursor: 'pointer',
                            transition: 'var(--transition)', width: '100%',
                        }}
                        onMouseEnter={e => (e.currentTarget.style.borderColor = 'var(--color-primary)')}
                        onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--color-border)')}
                    >
                        <span style={{ fontSize: '40px', flexShrink: 0 }}>📘</span>
                        <div>
                            <p style={{
                                margin: '0 0 4px', fontSize: '16px', fontWeight: 700,
                                color: 'var(--color-text)', fontFamily: 'var(--font-heading)',
                            }}>
                                Add a Book
                            </p>
                            <p style={{
                                margin: 0, fontSize: '13px',
                                color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                            }}>
                                Hardcover, paperback, or eBook. Search by the book's own ISBN.
                            </p>
                        </div>
                    </button>

                    {/* Audiobook card */}
                    <button
                        onClick={() => { setMode('audiobook'); setStep('search') }}
                        style={{
                            display: 'flex', alignItems: 'center', gap: '20px',
                            padding: '24px', textAlign: 'left',
                            background: 'var(--color-surface)', border: '2px solid var(--color-border)',
                            borderRadius: 'var(--border-radius)', cursor: 'pointer',
                            transition: 'var(--transition)', width: '100%',
                        }}
                        onMouseEnter={e => (e.currentTarget.style.borderColor = 'var(--color-primary)')}
                        onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--color-border)')}
                    >
                        <span style={{ fontSize: '40px', flexShrink: 0 }}>🎧</span>
                        <div>
                            <p style={{
                                margin: '0 0 4px', fontSize: '16px', fontWeight: 700,
                                color: 'var(--color-text)', fontFamily: 'var(--font-heading)',
                            }}>
                                Add an Audiobook
                            </p>
                            <p style={{
                                margin: 0, fontSize: '13px',
                                color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                            }}>
                                Audiobooks have their own ISBN, narrator, and duration. Search using the audiobook's ISBN.
                            </p>
                        </div>
                    </button>
                </div>
            )}

            {/* ── STEP: Search ── */}
            {step === 'search' && (
                <Card padding="lg">
                    {/* Mode badge */}
                    <div style={{
                        display: 'inline-flex', alignItems: 'center', gap: '6px',
                        padding: '4px 10px', marginBottom: '18px',
                        background: 'var(--color-surface-alt)', borderRadius: '999px',
                        border: '1px solid var(--color-border)',
                    }}>
                        <span>{isAudiobook ? '🎧' : '📘'}</span>
                        <span style={{ fontSize: '12px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-body)' }}>
              {isAudiobook ? 'Audiobook' : 'Book'}
            </span>
                        <button
                            onClick={() => setStep('pick')}
                            style={{
                                background: 'none', border: 'none', cursor: 'pointer',
                                color: 'var(--color-text-muted)', fontSize: '11px',
                                fontFamily: 'var(--font-body)', padding: '0 0 0 4px',
                            }}
                        >
                            change
                        </button>
                    </div>

                    {/* ISBN / Title toggle */}
                    <div style={{
                        display: 'flex', background: 'var(--color-surface-alt)',
                        borderRadius: 'var(--border-radius)', padding: '3px',
                        marginBottom: '20px', width: 'fit-content',
                    }}>
                        {([
                            { key: 'isbn'  as SearchMode, label: '🔢 ISBN' },
                            { key: 'title' as SearchMode, label: '🔤 Title / Author' },
                        ]).map(m => (
                            <button key={m.key}
                                    onClick={() => { setSearchMode(m.key); setSearchError('') }}
                                    style={{
                                        padding: '7px 18px', border: 'none',
                                        borderRadius: 'var(--border-radius)',
                                        background: searchMode === m.key ? 'var(--color-surface)' : 'transparent',
                                        color: searchMode === m.key ? 'var(--color-text)' : 'var(--color-text-muted)',
                                        fontSize: '13px', fontWeight: searchMode === m.key ? 600 : 400,
                                        cursor: 'pointer', transition: 'var(--transition)',
                                        boxShadow: searchMode === m.key ? 'var(--shadow-sm)' : 'none',
                                        fontFamily: 'var(--font-body)',
                                    }}
                            >
                                {m.label}
                            </button>
                        ))}
                    </div>

                    {searchMode === 'isbn' ? (
                        <div>
                            <FieldLabel
                                label="ISBN"
                                hint={isAudiobook ? 'Use the audiobook ISBN — not the print edition.' : undefined}
                            />
                            <TextInput value={isbn} onChange={setIsbn} placeholder="e.g. 9780743273565" />
                        </div>
                    ) : (
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                            <div>
                                <FieldLabel label="Title" />
                                <TextInput value={titleQuery} onChange={setTitleQuery} placeholder="e.g. The Name of the Wind" />
                            </div>
                            <div>
                                <FieldLabel label="Author" />
                                <TextInput value={authorQuery} onChange={setAuthorQuery} placeholder="e.g. Patrick Rothfuss" />
                            </div>
                        </div>
                    )}

                    {searchError && (
                        <div style={{ marginTop: '14px' }}>
                            <ErrorBox message={searchError} />
                        </div>
                    )}

                    <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
                        <SecondaryButton label="← Back" onClick={() => setStep('pick')} />
                        <PrimaryButton
                            label={searchMutation.isPending ? 'Searching...' : 'Search'}
                            onClick={() => searchMutation.mutate()}
                            isLoading={searchMutation.isPending}
                            disabled={searchMode === 'isbn' ? !isbn.trim() : !titleQuery.trim()}
                            fullWidth
                        />
                    </div>
                    <button
                        onClick={() => {
                            if (searchMode === 'isbn') setManualIsbn(isbn.replace(/[-\s]/g, ''))
                            else { setManualTitle(titleQuery); setManualAuthors(authorQuery) }
                            setStep('manual')
                        }}
                        style={{
                            background: 'none', border: 'none', cursor: 'pointer',
                            color: 'var(--color-text-muted)', fontSize: '12px',
                            fontFamily: 'var(--font-body)', marginTop: '8px',
                            textDecoration: 'underline', padding: '4px',
                        }}
                    >
                        Can't find it? Enter details manually →
                    </button>
                </Card>
            )}

            {/* ── STEP: Preview ── */}
            {step === 'preview' && lookupResult && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    <Card padding="lg">
                        <p style={{
                            margin: '0 0 16px', fontSize: '11px', fontWeight: 700,
                            color: 'var(--color-text-muted)', textTransform: 'uppercase',
                            letterSpacing: '0.06em', fontFamily: 'var(--font-body)',
                        }}>
                            Found via Google Books / OpenLibrary
                        </p>

                        <div style={{ display: 'flex', gap: '16px' }}>
                            <CoverPreview coverUrl={lookupResult.cover_url} title={lookupResult.title} size="md" />
                            <div style={{ flex: 1 }}>
                                <h2 style={{
                                    margin: '0 0 6px', fontSize: '17px', fontWeight: 700,
                                    color: 'var(--color-text)', fontFamily: 'var(--font-heading)', lineHeight: 1.3,
                                }}>
                                    {lookupResult.title}
                                </h2>
                                <p style={{
                                    margin: '0 0 8px', fontSize: '13px',
                                    color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                                }}>
                                    {lookupResult.authors?.join(', ') || 'Unknown author'}
                                </p>
                                <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                                    {getIsbn(lookupResult) && (
                                        <Badge label={`ISBN ${getIsbn(lookupResult)}`} variant="default" size="sm" />
                                    )}
                                    {lookupResult.page_count != null && !isAudiobook && (
                                        <Badge label={`${lookupResult.page_count} pages`} variant="muted" size="sm" />
                                    )}
                                    {lookupResult.language && (
                                        <Badge label={lookupResult.language.toUpperCase()} variant="muted" size="sm" />
                                    )}
                                    {lookupResult.publisher && (
                                        <Badge label={lookupResult.publisher} variant="muted" size="sm" />
                                    )}
                                </div>
                            </div>
                        </div>

                        {lookupResult.description && (
                            <p style={{
                                margin: '16px 0 0', fontSize: '12px', color: 'var(--color-text-muted)',
                                lineHeight: '1.6', fontFamily: 'var(--font-body)',
                                display: '-webkit-box', WebkitLineClamp: 4,
                                WebkitBoxOrient: 'vertical', overflow: 'hidden',
                            }}>
                                {lookupResult.description}
                            </p>
                        )}
                    </Card>

                    {existingEdition && (
                        <Card padding="md">
                            <p style={{
                                margin: '0 0 4px', fontSize: '13px', fontWeight: 600,
                                color: 'var(--color-text)', fontFamily: 'var(--font-body)',
                            }}>
                                📚 This edition is already in the catalogue
                            </p>
                            <p style={{
                                margin: '0 0 12px', fontSize: '12px',
                                color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                            }}>
                                You can add a copy directly to your library — no moderation needed.
                            </p>
                            <PrimaryButton label="Add copy to my library →" onClick={() => setStep('details')} />
                        </Card>
                    )}

                    <div style={{ display: 'flex', gap: '10px' }}>
                        <SecondaryButton label="← Search again" onClick={() => setStep('search')} />
                        {!existingEdition && (
                            <PrimaryButton label="Continue →" onClick={() => setStep('details')} fullWidth />
                        )}
                    </div>
                </div>
            )}
            {/* ── STEP: Manual entry ── */}
            {step === 'manual' && (
                <Card padding="lg">
                    <div style={{
                        padding: '10px 14px', marginBottom: '20px',
                        background: 'rgba(234,179,8,0.08)', border: '1px solid rgba(234,179,8,0.25)',
                        borderRadius: 'var(--border-radius)',
                    }}>
                        <p style={{ margin: 0, fontSize: '12px', color: 'var(--color-text)', fontFamily: 'var(--font-body)', lineHeight: '1.6' }}>
                            📝 <strong>Manual entry</strong> — fill in what you know. Your submission will go to a moderator for review.
                        </p>
                    </div>

                    <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                        <div>
                            <FieldLabel label="Title" hint="Required" />
                            <TextInput value={manualTitle} onChange={setManualTitle} placeholder="e.g. The Name of the Wind" />
                        </div>
                        <div>
                            <FieldLabel label="Author(s)" hint="Required — separate multiple with commas" />
                            <TextInput value={manualAuthors} onChange={setManualAuthors} placeholder="e.g. Patrick Rothfuss" />
                        </div>
                        <div>
                            <FieldLabel label="ISBN" hint="Optional" />
                            <TextInput value={manualIsbn} onChange={setManualIsbn} placeholder="e.g. 9780756404079" />
                        </div>
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                            <div>
                                <FieldLabel label="Publisher" hint="Optional" />
                                <TextInput value={manualPublisher} onChange={setManualPublisher} placeholder="e.g. DAW Books" />
                            </div>
                            <div>
                                <FieldLabel label={isAudiobook ? 'Duration (min)' : 'Page count'} hint="Optional" />
                                <TextInput
                                    value={isAudiobook ? duration : manualPages}
                                    onChange={isAudiobook ? setDuration : setManualPages}
                                    placeholder={isAudiobook ? 'e.g. 640' : 'e.g. 662'}
                                    type="number"
                                />
                            </div>
                        </div>
                        {isAudiobook && (
                            <div>
                                <FieldLabel label="Narrator" hint="Optional" />
                                <TextInput value={narrator} onChange={setNarrator} placeholder="e.g. Stephen Fry" />
                            </div>
                        )}
                        <div>
                            <FieldLabel label="Description" hint="Optional" />
                            <textarea
                                value={manualDesc}
                                onChange={e => setManualDesc(e.target.value)}
                                placeholder="A brief description…"
                                rows={3}
                                style={{
                                    width: '100%', padding: '10px 12px', boxSizing: 'border-box',
                                    background: 'var(--input-bg)', border: '1px solid var(--color-border)',
                                    borderRadius: 'var(--border-radius)', color: 'var(--color-text)',
                                    fontSize: '13px', fontFamily: 'var(--font-body)', outline: 'none', resize: 'vertical',
                                }}
                            />
                        </div>

                        <hr style={{ border: 'none', borderTop: '1px solid var(--color-border)' }} />

                        {!isAudiobook && (
                            <div>
                                <FieldLabel label="Format" />
                                <ToggleGroup
                                    value={format}
                                    onChange={v => setFormat(v as typeof format)}
                                    options={[
                                        { value: 'paperback', label: 'Paperback', icon: '📗' },
                                        { value: 'hardcover', label: 'Hardcover', icon: '📘' },
                                        { value: 'ebook',     label: 'eBook',     icon: '💻' },
                                    ]}
                                />
                            </div>
                        )}
                        {!isAudiobook && format !== 'ebook' && (
                            <div>
                                <FieldLabel label="Condition" />
                                <ToggleGroup
                                    value={condition}
                                    onChange={setCondition}
                                    options={[
                                        { value: 'new',  label: 'New'  },
                                        { value: 'good', label: 'Good' },
                                        { value: 'fair', label: 'Fair' },
                                        { value: 'poor', label: 'Poor' },
                                    ]}
                                />
                            </div>
                        )}
                        <div>
                            <FieldLabel label="Language" />
                            <select
                                value={language}
                                onChange={e => setLanguage(e.target.value)}
                                style={{
                                    width: '100%', padding: '10px 12px',
                                    background: 'var(--input-bg)', border: '1px solid var(--color-border)',
                                    borderRadius: 'var(--border-radius)', color: 'var(--color-text)',
                                    fontSize: '13px', fontFamily: 'var(--font-body)', outline: 'none',
                                }}
                            >
                                {[
                                    { code: 'en', label: 'English' }, { code: 'pt', label: 'Portuguese' },
                                    { code: 'es', label: 'Spanish' }, { code: 'fr', label: 'French' },
                                    { code: 'de', label: 'German' },  { code: 'it', label: 'Italian' },
                                    { code: 'ja', label: 'Japanese' },{ code: 'zh', label: 'Chinese' },
                                    { code: 'ru', label: 'Russian' }, { code: 'ar', label: 'Arabic' },
                                    { code: 'nl', label: 'Dutch' },   { code: 'ko', label: 'Korean' },
                                    { code: 'pl', label: 'Polish' },  { code: 'sv', label: 'Swedish' },
                                    { code: 'other', label: 'Other' },
                                ].map(lang => (
                                    <option key={lang.code} value={lang.code}>{lang.label}</option>
                                ))}
                            </select>
                        </div>

                        {manualError && <ErrorBox message={manualError} />}

                        <div style={{ display: 'flex', gap: '10px' }}>
                            <SecondaryButton label="← Back to search" onClick={() => setStep('search')} />
                            <PrimaryButton
                                label={submitManualMutation.isPending ? 'Submitting...' : 'Submit for review'}
                                onClick={() => submitManualMutation.mutate()}
                                isLoading={submitManualMutation.isPending}
                                disabled={!manualTitle.trim() || !manualAuthors.trim()}
                                fullWidth
                            />
                        </div>
                    </div>
                </Card>
            )}
            {/* ── STEP: Details ── */}
            {step === 'details' && lookupResult && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
                    {/* Summary strip */}
                    <Card padding="md">
                        <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
                            <CoverPreview coverUrl={lookupResult.cover_url} title={lookupResult.title} size="sm" />
                            <div>
                                <p style={{
                                    margin: 0, fontSize: '14px', fontWeight: 600,
                                    color: 'var(--color-text)', fontFamily: 'var(--font-body)',
                                }}>
                                    {lookupResult.title}
                                </p>
                                <p style={{
                                    margin: '2px 0 0', fontSize: '12px',
                                    color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                                }}>
                                    {lookupResult.authors?.join(', ')}
                                </p>
                            </div>
                        </div>
                    </Card>

                    <Card padding="lg">
                        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>

                            {/* ── BOOK fields ── */}
                            {!isAudiobook && (
                                <>
                                    <div>
                                        <FieldLabel label="Format" />
                                        <ToggleGroup
                                            value={format}
                                            onChange={v => setFormat(v as typeof format)}
                                            options={[
                                                { value: 'paperback', label: 'Paperback', icon: '📗' },
                                                { value: 'hardcover', label: 'Hardcover', icon: '📘' },
                                                { value: 'ebook',     label: 'eBook',     icon: '💻' },
                                            ]}
                                        />
                                    </div>

                                    {format !== 'ebook' && (
                                        <div>
                                            <FieldLabel label="Condition" />
                                            <ToggleGroup
                                                value={condition}
                                                onChange={setCondition}
                                                options={[
                                                    { value: 'new',  label: 'New'  },
                                                    { value: 'good', label: 'Good' },
                                                    { value: 'fair', label: 'Fair' },
                                                    { value: 'poor', label: 'Poor' },
                                                ]}
                                            />
                                        </div>
                                    )}
                                </>
                            )}

                            {/* ── AUDIOBOOK fields ── */}
                            {isAudiobook && (
                                <>
                                    <div>
                                        <FieldLabel label="Narrator" hint="Optional" />
                                        <TextInput
                                            value={narrator}
                                            onChange={setNarrator}
                                            placeholder="e.g. Stephen Fry"
                                        />
                                    </div>
                                    <div>
                                        <FieldLabel label="Duration" hint="Total length in minutes — optional" />
                                        <TextInput
                                            value={duration}
                                            onChange={setDuration}
                                            placeholder="e.g. 640"
                                            type="number"
                                        />
                                    </div>
                                </>
                            )}

                            {/* Language — both modes */}
                            <div>
                                <FieldLabel label="Language" />
                                <select
                                    value={language}
                                    onChange={e => setLanguage(e.target.value)}
                                    style={{
                                        width: '100%', padding: '10px 12px',
                                        background: 'var(--input-bg)', border: '1px solid var(--color-border)',
                                        borderRadius: 'var(--border-radius)', color: 'var(--color-text)',
                                        fontSize: '13px', fontFamily: 'var(--font-body)', outline: 'none',
                                    }}
                                >
                                    {[
                                        { code: 'en',    label: 'English'    },
                                        { code: 'pt',    label: 'Portuguese' },
                                        { code: 'es',    label: 'Spanish'    },
                                        { code: 'fr',    label: 'French'     },
                                        { code: 'de',    label: 'German'     },
                                        { code: 'it',    label: 'Italian'    },
                                        { code: 'ja',    label: 'Japanese'   },
                                        { code: 'zh',    label: 'Chinese'    },
                                        { code: 'ru',    label: 'Russian'    },
                                        { code: 'ar',    label: 'Arabic'     },
                                        { code: 'nl',    label: 'Dutch'      },
                                        { code: 'ko',    label: 'Korean'     },
                                        { code: 'pl',    label: 'Polish'     },
                                        { code: 'sv',    label: 'Swedish'    },
                                        { code: 'other', label: 'Other'      },
                                    ].map(lang => (
                                        <option key={lang.code} value={lang.code}>{lang.label}</option>
                                    ))}
                                </select>
                            </div>

                            {/* Info box */}
                            <div style={{
                                padding: '12px 14px',
                                background: 'var(--color-surface-alt)',
                                borderRadius: 'var(--border-radius)',
                                border: '1px solid var(--color-border)',
                            }}>
                                <p style={{
                                    margin: 0, fontSize: '12px', color: 'var(--color-text-muted)',
                                    fontFamily: 'var(--font-body)', lineHeight: '1.6',
                                }}>
                                    {existingEdition
                                        ? '📚 This copy will be added to your library immediately.'
                                        : '🔍 This book will be reviewed by a moderator before appearing in the catalogue. Once approved, it will be added to your library automatically.'}
                                </p>
                            </div>

                            {detailsError && <ErrorBox message={detailsError} />}

                            <div style={{ display: 'flex', gap: '10px' }}>
                                <SecondaryButton label="← Back" onClick={() => setStep('preview')} />
                                <PrimaryButton
                                    label={
                                        existingEdition
                                            ? (addCopyMutation.isPending ? 'Adding...'     : 'Add to my library')
                                            : (submitMutation.isPending  ? 'Submitting...' : 'Submit for review')
                                    }
                                    onClick={() => existingEdition ? addCopyMutation.mutate() : submitMutation.mutate()}
                                    isLoading={addCopyMutation.isPending || submitMutation.isPending}
                                    fullWidth
                                />
                            </div>
                        </div>
                    </Card>
                </div>
            )}

            {/* ── STEP: Done ── */}
            {step === 'done' && (
                <Card padding="lg">
                    <div style={{
                        display: 'flex', flexDirection: 'column', alignItems: 'center',
                        gap: '16px', padding: '24px 0', textAlign: 'center',
                    }}>
                        <span style={{ fontSize: '64px' }}>{isAudiobook ? '🎧' : '🎉'}</span>
                        <h2 style={{
                            margin: 0, fontSize: '20px', fontWeight: 700,
                            color: 'var(--color-text)', fontFamily: 'var(--font-heading)',
                        }}>
                            All done!
                        </h2>
                        <p style={{
                            margin: 0, fontSize: '14px', color: 'var(--color-text-muted)',
                            maxWidth: '360px', lineHeight: '1.6', fontFamily: 'var(--font-body)',
                        }}>
                            {doneMessage}
                        </p>
                        <div style={{ display: 'flex', gap: '10px', marginTop: '8px' }}>
                            <SecondaryButton label="Add another" onClick={reset} />
                            <PrimaryButton label="Go to my library" onClick={() => navigate('/library')} />
                        </div>
                    </div>
                </Card>
            )}
        </div>
    )
}