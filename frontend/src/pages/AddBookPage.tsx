import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import apiClient from '../api/client'
import { Card, Badge, Spinner } from '../components'
import type { LookupResultsPage, LookupFilters } from '../types'
import { lookupByISBN, lookupByTitleAuthor, checkDuplicate, submitBook as submitBookApi } from '../api/books'
import { useAuth } from '../context/AuthContext'

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
type Step = 'pick' | 'search' | 'results' | 'preview' | 'manual' | 'details' | 'done'
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
    { key: 'results', label: 'Results' },
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



function SecondaryButton({ label, onClick, disabled }: { label: string; onClick: () => void; disabled?: boolean }) {
    return (
        <button onClick={onClick} disabled={disabled} style={{
            opacity: disabled ? 0.4 : 1,
            padding: '10px 18px',
            background: 'var(--color-surface-alt)', color: 'var(--color-text)',
            border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)',
            fontSize: '13px', fontWeight: 500, fontFamily: 'var(--font-body)',
            cursor: disabled ? 'not-allowed' : 'pointer', transition: 'var(--transition)',
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

// ── TagInput: comma/enter to add, × to remove ─────────────────────────────────

function TagInput({ values, onChange, placeholder }: {
    values: string[]
    onChange: (v: string[]) => void
    placeholder?: string
}) {
    const [input, setInput] = useState('')

    function commit() {
        const trimmed = input.trim()
        if (trimmed && !values.includes(trimmed)) {
            onChange([...values, trimmed])
        }
        setInput('')
    }

    return (
        <div style={{
            display: 'flex', flexWrap: 'wrap', gap: '6px', alignItems: 'center',
            padding: '8px 10px', minHeight: '42px',
            background: 'var(--input-bg)', border: '1px solid var(--color-border)',
            borderRadius: 'var(--border-radius)',
        }}>
            {values.map(v => (
                <span key={v} style={{
                    display: 'inline-flex', alignItems: 'center', gap: '4px',
                    padding: '3px 8px', background: 'var(--color-primary)',
                    color: 'var(--color-primary-text)', borderRadius: '999px',
                    fontSize: '12px', fontFamily: 'var(--font-body)', fontWeight: 500,
                }}>
                    {v}
                    <button onClick={() => onChange(values.filter(x => x !== v))} style={{
                        background: 'none', border: 'none', cursor: 'pointer',
                        color: 'inherit', padding: '0 0 0 2px', fontSize: '13px', lineHeight: 1,
                    }}>×</button>
                </span>
            ))}
            <input
                value={input}
                onChange={e => setInput(e.target.value)}
                onKeyDown={e => {
                    if (e.key === 'Enter' || e.key === ',') { e.preventDefault(); commit() }
                    if (e.key === 'Backspace' && !input && values.length) {
                        onChange(values.slice(0, -1))
                    }
                }}
                onBlur={commit}
                placeholder={values.length === 0 ? placeholder : 'Add another…'}
                style={{
                    flex: 1, minWidth: '120px', background: 'none', border: 'none',
                    outline: 'none', color: 'var(--color-text)',
                    fontSize: '13px', fontFamily: 'var(--font-body)', padding: '2px 4px',
                }}
            />
        </div>
    )
}

// ── Main page ─────────────────────────────────────────────────────────────────

export default function AddBookPage() {
    const navigate  = useNavigate()
    const { user }  = useAuth()
    const isModerator = user?.role === 'moderator' || user?.role === 'admin'

    const [catalogueOnly, setCatalogueOnly] = useState(false)
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

    const [manualTitle,       setManualTitle]       = useState('')
    const [manualAuthors,     setManualAuthors]     = useState<string[]>([])
    const [manualTranslators, setManualTranslators] = useState<string[]>([])
    const [manualIsbn,        setManualIsbn]        = useState('')
    const [manualAsin,        setManualAsin]        = useState('')
    const [manualPublisher,   setManualPublisher]   = useState('')
    const [manualEditionLabel,setManualEditionLabel]= useState('')
    const [manualPublishedAt, setManualPublishedAt] = useState('')
    const [manualPages,       setManualPages]       = useState('')
    const [manualFileFormat,  setManualFileFormat]  = useState('')
    const [manualAudioFormat, setManualAudioFormat] = useState('')
    const [manualDesc,        setManualDesc]        = useState('')
    const [manualCoverUrl,    setManualCoverUrl]    = useState('')
    const [manualError,       setManualError]       = useState('')

    const [searchResults, setSearchResults]     = useState<BookLookup[]>([])
    const [searchTotal, setSearchTotal]         = useState(0)
    const [searchPage, setSearchPage]           = useState(1)

    const [filters, setFilters] = useState<LookupFilters>({})
    const [showFilters, setShowFilters] = useState(false)
    const [resultsSort, setResultsSort] = useState<'default' | 'title_asc' | 'year_desc' | 'year_asc'>('default')

    // ── Search ──────────────────────────────────────────────────────────────────

    const searchMutation = useMutation({
        mutationFn: async (): Promise<BookLookup | LookupResultsPage | null> => {
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
                const data = await performTitleSearch(searchPage)
                return data
            }
        },
        // CHANGE onSuccess:
        onSuccess: (result) => {
            if (searchMode === 'title') {
                const data = result as unknown as { results: BookLookup[]; total: number }
                if (!data?.results?.length && step !== 'results') {
                    // Only redirect to manual if we're not already showing results
                    setManualTitle(titleQuery)
                    setManualAuthors(authorQuery ? [authorQuery] : [])
                    setStep('manual')
                } else {
                    setStep('results')
                }
                return
            }
            // ISBN path unchanged
            if (!result) {
                setManualIsbn(isbn.replace(/[-\s]/g, ''))
                setStep('manual')
                return
            }
            const book = result as BookLookup
            setLookupResult(book)
            setLanguage(book.language ?? 'en')
            setStep('preview')
        },
        onError: () => {
            // API error — also offer manual entry
            if (searchMode === 'isbn') setManualIsbn(isbn.replace(/[-\s]/g, ''))
            else { setManualTitle(titleQuery); setManualAuthors(authorQuery ? [authorQuery] : []) }
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
            const dbFormat = mode === 'audiobook' ? 'audiobook' : format
            return submitBookApi({
                title:       lookupResult.title,
                authors:     lookupResult.authors ?? [],
                description: lookupResult.description,
                cover_url:   lookupResult.cover_url,
                genres:      [],
                catalogue_only: catalogueOnly,
                edition: {
                    format:           dbFormat,
                    isbn:             getIsbn(lookupResult),
                    language,
                    publisher:        lookupResult.publisher,
                    page_count:       mode === 'book' ? lookupResult.page_count : undefined,
                    duration_minutes: mode === 'audiobook' && duration ? parseInt(duration) : undefined,
                    narrator:         mode === 'audiobook' ? narrator : undefined,
                    file_format:      mode === 'book' && format === 'ebook' ? manualFileFormat || undefined : undefined,
                },
                condition: (mode === 'book' && !catalogueOnly && format !== 'ebook') ? condition : undefined,
            })
        },
        onSuccess: () => {
            setDoneMessage(catalogueOnly
                ? 'Book added to the catalogue and is now publicly visible.'
                : 'Book submitted! It will appear in the catalogue once a moderator approves it.')
            setStep('done')
        },
        onError: () => setDetailsError('Submission failed. Please try again.'),
    })

    const submitManualMutation = useMutation({
        mutationFn: () => {
            return submitBookApi({
                title:       manualTitle.trim(),
                authors:     manualAuthors,
                description: manualDesc || undefined,
                cover_url:   manualCoverUrl || undefined,
                genres:      [],
                catalogue_only: catalogueOnly,
                edition: {
                    format:           mode === 'audiobook' ? 'audiobook' : format,
                    isbn:             manualIsbn || undefined,
                    asin:             manualAsin || undefined,
                    language,
                    publisher:        manualPublisher || undefined,
                    edition:          manualEditionLabel || undefined,
                    published_at:     manualPublishedAt || undefined,
                    page_count:       mode === 'book' && manualPages ? parseInt(manualPages) : undefined,
                    file_format:      mode === 'book' && format === 'ebook' ? manualFileFormat || undefined : undefined,
                    duration_minutes: mode === 'audiobook' && duration ? parseInt(duration) : undefined,
                    narrator:         mode === 'audiobook' ? narrator || undefined : undefined,
                    audio_format:     mode === 'audiobook' ? manualAudioFormat || undefined : undefined,
                    translators:      manualTranslators.length > 0 ? manualTranslators : undefined,
                },
                condition: (mode === 'book' && !catalogueOnly && format !== 'ebook') ? condition : undefined,
            })
        },
        onSuccess: () => {
            setDoneMessage(catalogueOnly
                ? 'Book added to the catalogue and is now publicly visible.'
                : 'Book submitted! It will appear in the catalogue once a moderator approves it.')
            setStep('done')
        },
        onError: () => setManualError('Submission failed. Please try again.'),
    })


    function handleResultsPageChange(newPage: number) {
        setSearchPage(newPage)
        lookupByTitleAuthor(titleQuery.trim(), authorQuery.trim() || undefined, newPage)
            .then(data => {
                setSearchResults(data.results as unknown as BookLookup[])
                setSearchTotal(data.total)
            })
    }

    async function performTitleSearch(page: number) {
        const author = authorQuery.trim() || undefined
        const data = await lookupByTitleAuthor(titleQuery.trim(), author, page)
        setSearchResults(data.results as unknown as BookLookup[])
        setSearchTotal(data.total)
        return data
    }

    // ── Reset ────────────────────────────────────────────────────────────────────

    function reset() {
        setStep('pick')
        setCatalogueOnly(false)
        setIsbn(''); setTitleQuery(''); setAuthorQuery(''); setSearchError('')
        setLookupResult(null); setExistingEdition(null)
        setFormat('paperback'); setCondition('good')
        setNarrator(''); setDuration('')
        setLanguage('en'); setDetailsError(''); setDoneMessage('')
        setManualTitle(''); setManualAuthors([]); setManualTranslators([]); setManualIsbn('')
        setManualAsin(''); setManualPublisher(''); setManualEditionLabel(''); setManualPublishedAt('')
        setManualPages(''); setManualFileFormat(''); setManualAudioFormat('')
        setManualDesc(''); setManualCoverUrl(''); setManualError('')
        setFilters({})
        setShowFilters(false)
        setSearchResults([])
        setSearchTotal(0)
        setSearchPage(1)
    }

    // ── Render ───────────────────────────────────────────────────────────────────

    const isAudiobook = mode === 'audiobook'
    const filteredResults = (() => {
        let list = searchResults.filter(book => {
            if (filters.language && book.language !== filters.language) return false
            if (filters.publisher && !book.publisher?.toLowerCase().includes(filters.publisher.toLowerCase())) return false
            if (filters.author && !book.authors?.some(a => a.toLowerCase().includes(filters.author!.toLowerCase()))) return false
            if (filters.yearFrom && (!book.published_date || parseInt(book.published_date) < filters.yearFrom)) return false
            if (filters.yearTo && (!book.published_date || parseInt(book.published_date) > filters.yearTo)) return false
            return true
        })
        switch (resultsSort) {
            case 'title_asc':  list = [...list].sort((a, b) => a.title.localeCompare(b.title)); break
            case 'year_desc':  list = [...list].sort((a, b) => (b.published_date ?? '').localeCompare(a.published_date ?? '')); break
            case 'year_asc':   list = [...list].sort((a, b) => (a.published_date ?? '').localeCompare(b.published_date ?? '')); break
        }
        return list
    })()

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
                    {catalogueOnly ? 'Add to Catalogue' : 'Add to my Library'}
                </h1>
                <p style={{
                    margin: '6px 0 0', fontSize: '13px',
                    color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                }}>
                    Choose what you're adding, then search for it.
                </p>

                {/* Intent toggle — mod/admin only */}
                {isModerator && (
                    <div style={{
                        display: 'flex', marginTop: '16px',
                        background: 'var(--color-surface-alt)',
                        borderRadius: 'var(--border-radius)', padding: '3px',
                        width: 'fit-content', border: '1px solid var(--color-border)',
                    }}>
                        {([
                            { value: false, label: '🏠 Add to my Library' },
                            { value: true,  label: '📚 Catalogue only'    },
                        ] as { value: boolean; label: string }[]).map(opt => (
                            <button
                                key={String(opt.value)}
                                onClick={() => setCatalogueOnly(opt.value)}
                                style={{
                                    padding: '8px 18px', border: 'none',
                                    borderRadius: 'var(--border-radius)',
                                    background: catalogueOnly === opt.value ? 'var(--color-primary)' : 'transparent',
                                    color: catalogueOnly === opt.value ? 'var(--color-primary-text)' : 'var(--color-text-muted)',
                                    fontSize: '13px', fontWeight: catalogueOnly === opt.value ? 600 : 400,
                                    cursor: 'pointer', transition: 'var(--transition)',
                                    fontFamily: 'var(--font-body)',
                                }}
                            >
                                {opt.label}
                            </button>
                        ))}
                    </div>
                )}
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
                            else { setManualTitle(titleQuery); setManualAuthors(authorQuery ? [authorQuery] : []) }
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

            {/* ── STEP: Results ── */}
            {step === 'results' && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                    <Card padding="lg">
                        <p style={{
                            margin: '0 0 4px', fontSize: '13px',
                            color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                        }}>
                            {filteredResults.length} of {searchTotal} result{searchTotal !== 1 ? 's' : ''} for "{titleQuery}"
                            {authorQuery ? ` by "${authorQuery}"` : ''}
                        </p>
                    </Card>

                    {/* Sort + Filters toolbar */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: '8px', flexWrap: 'wrap' }}>
                        <select
                            value={resultsSort}
                            onChange={e => setResultsSort(e.target.value as typeof resultsSort)}
                            style={{ padding: '6px 10px', background: 'var(--input-bg)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', color: 'var(--color-text)', fontSize: '12px', fontFamily: 'var(--font-body)' }}
                        >
                            <option value="default">Sort: Relevance</option>
                            <option value="title_asc">Sort: Title A → Z</option>
                            <option value="year_desc">Sort: Newest first</option>
                            <option value="year_asc">Sort: Oldest first</option>
                        </select>
                        <button
                            onClick={() => setShowFilters(v => !v)}
                            style={{
                                background: 'none', border: '1px solid var(--color-border)',
                                borderRadius: 'var(--border-radius)', padding: '6px 12px',
                                fontSize: '12px', cursor: 'pointer', fontFamily: 'var(--font-body)',
                                color: 'var(--color-text-muted)',
                            }}
                        >
                            {showFilters ? 'Hide filters ▲' : 'Filter results ▼'}
                        </button>
                    </div>

                    {showFilters && (
                        <Card padding="md">
                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                                <div>
                                    <FieldLabel label="Language" />
                                    <select
                                        value={filters.language ?? ''}
                                        onChange={e => setFilters(f => ({ ...f, language: e.target.value || undefined }))}
                                        style={{
                                            width: '100%', padding: '8px 10px',
                                            background: 'var(--input-bg)', border: '1px solid var(--color-border)',
                                            borderRadius: 'var(--border-radius)', color: 'var(--color-text)',
                                            fontSize: '13px', fontFamily: 'var(--font-body)',
                                        }}
                                    >
                                        <option value="">Any</option>
                                        <option value="en">English</option>
                                        <option value="pt">Portuguese (Portugal)</option>
                                        <option value="pt-BR">Portuguese (Brazil)</option>
                                        <option value="es">Spanish</option>
                                        <option value="fr">French</option>
                                        <option value="de">German</option>
                                        <option value="it">Italian</option>
                                        <option value="ja">Japanese</option>
                                        <option value="zh">Chinese</option>
                                        <option value="ru">Russian</option>
                                        <option value="ar">Arabic</option>
                                        <option value="nl">Dutch</option>
                                        <option value="ko">Korean</option>
                                    </select>
                                </div>

                                <div>
                                    <FieldLabel label="Publisher" />
                                    <TextInput
                                        value={filters.publisher ?? ''}
                                        onChange={v => setFilters(f => ({ ...f, publisher: v || undefined }))}
                                        placeholder="e.g. Bloomsbury"
                                    />
                                </div>

                                {!authorQuery.trim() && (
                                    <div>
                                        <FieldLabel label="Author" />
                                        <TextInput
                                            value={filters.author ?? ''}
                                            onChange={v => setFilters(f => ({ ...f, author: v || undefined }))}
                                            placeholder="e.g. Rowling"
                                        />
                                    </div>
                                )}

                                <div>
                                    <FieldLabel label="Year from" />
                                    <TextInput
                                        value={filters.yearFrom?.toString() ?? ''}
                                        onChange={v => setFilters(f => ({ ...f, yearFrom: v ? parseInt(v) : undefined }))}
                                        placeholder="e.g. 1990"
                                        type="number"
                                    />
                                </div>

                                <div>
                                    <FieldLabel label="Year to" />
                                    <TextInput
                                        value={filters.yearTo?.toString() ?? ''}
                                        onChange={v => setFilters(f => ({ ...f, yearTo: v ? parseInt(v) : undefined }))}
                                        placeholder="e.g. 2024"
                                        type="number"
                                    />
                                </div>
                            </div>

                            <div style={{ marginTop: '12px', display: 'flex', gap: '8px' }}>
                                <SecondaryButton
                                    label="Clear filters"
                                    onClick={() => setFilters({})}
                                />
                            </div>
                        </Card>
                    )}

                    {filteredResults.map((book, i) => (
                        <Card key={i} padding="md">
                            <button
                                onClick={() => {
                                    setLookupResult(book)
                                    setLanguage(book.language ?? 'en')
                                    setStep('preview')
                                }}
                                style={{
                                    display: 'flex', gap: '12px', width: '100%',
                                    background: 'none', border: 'none', cursor: 'pointer',
                                    textAlign: 'left', padding: 0,
                                }}
                            >
                                <CoverPreview coverUrl={book.cover_url} title={book.title} size="sm" />
                                <div style={{ flex: 1 }}>
                                    <p style={{
                                        margin: '0 0 4px', fontSize: '14px', fontWeight: 600,
                                        color: 'var(--color-text)', fontFamily: 'var(--font-heading)',
                                    }}>
                                        {book.title}
                                    </p>
                                    <p style={{
                                        margin: '0 0 6px', fontSize: '12px',
                                        color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                                    }}>
                                        {book.authors?.join(', ') || 'Unknown author'}
                                        {book.published_date ? ` · ${book.published_date.slice(0, 4)}` : ''}
                                    </p>
                                    <div style={{ display: 'flex', gap: '4px', flexWrap: 'wrap' }}>
                                        {book.publisher && <Badge label={book.publisher} variant="muted" size="sm" />}
                                        {book.isbn_13 && <Badge label={`ISBN ${book.isbn_13}`} variant="default" size="sm" />}
                                    </div>
                                </div>
                            </button>
                        </Card>
                    ))}

                    {/* Pagination */}
                    {searchTotal > 20 && (
                        <div style={{ display: 'flex', justifyContent: 'center', gap: '8px', marginTop: '8px' }}>
                            <SecondaryButton
                                label="← Prev"
                                onClick={() => handleResultsPageChange(searchPage - 1)}
                                disabled={searchPage === 1}
                            />
                            <span style={{
                                alignSelf: 'center', fontSize: '13px',
                                color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)',
                            }}>
                    Page {searchPage} of {Math.ceil(searchTotal / 20)}
                </span>
                            <SecondaryButton
                                label="Next →"
                                onClick={() => handleResultsPageChange(searchPage + 1)}
                                disabled={searchPage >= Math.ceil(searchTotal / 20)}
                            />
                        </div>
                    )}

                    <div style={{ display: 'flex', gap: '10px', marginTop: '4px' }}>
                        <SecondaryButton label="← Back" onClick={() => setStep('search')} />
                        <button
                            onClick={() => { setManualTitle(titleQuery); setManualAuthors(authorQuery ? [authorQuery] : []); setStep('manual') }}
                            style={{
                                background: 'none', border: 'none', cursor: 'pointer',
                                color: 'var(--color-text-muted)', fontSize: '12px',
                                fontFamily: 'var(--font-body)', textDecoration: 'underline', padding: '4px',
                            }}
                        >
                            None of these — enter manually
                        </button>
                    </div>
                </div>
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
                            <FieldLabel label="Author(s)" hint="Required — press Enter or comma to add each name" />
                            <TagInput values={manualAuthors} onChange={setManualAuthors} placeholder="e.g. Patrick Rothfuss" />
                        </div>
                        {!isAudiobook && (
                            <div>
                                <FieldLabel label="Translator(s)" hint="Optional — press Enter or comma to add each name" />
                                <TagInput values={manualTranslators} onChange={setManualTranslators} placeholder="e.g. Margaret Jull Costa" />
                            </div>
                        )}
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                            <div>
                                <FieldLabel label="ISBN" hint="Optional" />
                                <TextInput value={manualIsbn} onChange={setManualIsbn} placeholder="e.g. 9780756404079" />
                            </div>
                            <div>
                                <FieldLabel label="ASIN" hint="Optional — Amazon ID" />
                                <TextInput value={manualAsin} onChange={setManualAsin} placeholder="e.g. B00AIUUXS4" />
                            </div>
                        </div>
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                            <div>
                                <FieldLabel label="Publisher" hint="Optional" />
                                <TextInput value={manualPublisher} onChange={setManualPublisher} placeholder="e.g. DAW Books" />
                            </div>
                            <div>
                                <FieldLabel label="Published year" hint="Optional" />
                                <TextInput value={manualPublishedAt} onChange={setManualPublishedAt} placeholder="e.g. 2007" />
                            </div>
                        </div>
                        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                            <div>
                                <FieldLabel label="Edition label" hint="Optional" />
                                <TextInput value={manualEditionLabel} onChange={setManualEditionLabel} placeholder="e.g. 10th Anniversary" />
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
                            <>
                                <div>
                                    <FieldLabel label="Narrator" hint="Optional" />
                                    <TextInput value={narrator} onChange={setNarrator} placeholder="e.g. Stephen Fry" />
                                </div>
                                <div>
                                    <FieldLabel label="Audio format" hint="Optional" />
                                    <ToggleGroup
                                        value={manualAudioFormat}
                                        onChange={setManualAudioFormat}
                                        options={[
                                            { value: 'MP3',  label: 'MP3'  },
                                            { value: 'AAC',  label: 'AAC'  },
                                            { value: 'FLAC', label: 'FLAC' },
                                            { value: 'WMA',  label: 'WMA'  },
                                        ]}
                                    />
                                </div>
                            </>
                        )}
                        {!isAudiobook && format === 'ebook' && !catalogueOnly && (
                            <div>
                                <FieldLabel label="File format" hint="Optional" />
                                <ToggleGroup
                                    value={manualFileFormat}
                                    onChange={setManualFileFormat}
                                    options={[
                                        { value: 'EPUB', label: 'EPUB' },
                                        { value: 'PDF',  label: 'PDF'  },
                                        { value: 'MOBI', label: 'MOBI' },
                                        { value: 'AZW3', label: 'AZW3' },
                                    ]}
                                />
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

                        {/* Cover image — file upload (base64 placeholder until real upload endpoint exists) */}
                        <div>
                            <FieldLabel label="Cover Image" hint="Optional — upload a photo of the cover" />
                            <div style={{ display: 'flex', gap: '12px', alignItems: 'flex-start' }}>
                                {manualCoverUrl && (
                                    <div style={{ width: '60px', height: '82px', flexShrink: 0, borderRadius: '4px', overflow: 'hidden', border: '1px solid var(--color-border)', background: `url(${manualCoverUrl}) center/cover no-repeat` }} />
                                )}
                                <div style={{ flex: 1 }}>
                                    <input
                                        type="file"
                                        accept="image/*"
                                        id="manual-cover-upload"
                                        style={{ display: 'none' }}
                                        onChange={e => {
                                            const file = e.target.files?.[0]
                                            if (!file) return
                                            const reader = new FileReader()
                                            reader.onload = ev => setManualCoverUrl(ev.target?.result as string)
                                            reader.readAsDataURL(file)
                                        }}
                                    />
                                    <label
                                        htmlFor="manual-cover-upload"
                                        style={{ display: 'inline-flex', alignItems: 'center', gap: '6px', padding: '8px 14px', background: 'var(--color-surface-alt)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', fontSize: '12px', fontFamily: 'var(--font-body)', color: 'var(--color-text)', cursor: 'pointer' }}
                                    >
                                        📷 {manualCoverUrl ? 'Change image' : 'Choose image'}
                                    </label>
                                    {manualCoverUrl && (
                                        <button onClick={() => setManualCoverUrl('')} style={{ marginLeft: '8px', background: 'none', border: 'none', color: 'var(--color-error)', fontSize: '11px', cursor: 'pointer', fontFamily: 'var(--font-body)' }}>
                                            Remove
                                        </button>
                                    )}
                                </div>
                            </div>
                        </div>

                        <hr style={{ border: 'none', borderTop: '1px solid var(--color-border)' }} />

                        {!isAudiobook && !catalogueOnly && (
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
                        {!isAudiobook && !catalogueOnly && format !== 'ebook' && (
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
                                    { code: 'en', label: 'English' }, { code: 'pt',    label: 'Portuguese (Portugal)' },
                                    { code: 'pt-BR', label: 'Portuguese (Brazil)'   },
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
                                label={
                                    catalogueOnly
                                        ? (submitManualMutation.isPending ? 'Adding...'     : 'Add to catalogue')
                                        : (submitManualMutation.isPending ? 'Submitting...' : 'Submit for review')
                                }
                                onClick={() => submitManualMutation.mutate()}
                                isLoading={submitManualMutation.isPending}
                                disabled={!manualTitle.trim() || manualAuthors.length === 0}
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
                            {!isAudiobook && !catalogueOnly && (
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
                                    {catalogueOnly
                                        ? '📚 This book will be added directly to the catalogue with no copy linked to any user.'
                                        : existingEdition
                                            ? '📚 This copy will be added to your library immediately.'
                                            : '🔍 This book will be reviewed by a moderator before appearing in the catalogue. Once approved, it will be added to your library automatically.'}
                                </p>
                            </div>

                            {detailsError && <ErrorBox message={detailsError} />}

                            <div style={{ display: 'flex', gap: '10px' }}>
                                <SecondaryButton label="← Back" onClick={() => setStep('preview')} />
                                <PrimaryButton
                                    label={
                                        catalogueOnly
                                            ? (submitMutation.isPending     ? 'Adding...'      : 'Add to catalogue')
                                            : existingEdition
                                                ? (addCopyMutation.isPending ? 'Adding...'      : 'Add to my library')
                                                : (submitMutation.isPending  ? 'Submitting...'  : 'Submit for review')
                                    }
                                    onClick={() => existingEdition && !catalogueOnly ? addCopyMutation.mutate() : submitMutation.mutate()}
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
                            <PrimaryButton
                                label={catalogueOnly ? 'Go to catalogue' : 'Go to my library'}
                                onClick={() => navigate(catalogueOnly ? '/books' : '/library')}
                            />
                        </div>
                    </div>
                </Card>
            )}
        </div>
    )
}