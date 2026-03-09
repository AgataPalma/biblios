import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listBooks, addToWishlist, addCopy } from '../api/books'
import { Card, Badge, Spinner, EmptyState, Modal } from '../components'
import { useAuth } from '../context/AuthContext'
import type { Book, Edition } from '../types'

// ─── Types ────────────────────────────────────────────────────────────────────

type ViewMode = 'grid' | 'list'
type SortKey  = 'default' | 'title_asc' | 'title_desc' | 'newest' | 'oldest'

// ─── Deterministic cover colour ───────────────────────────────────────────────

const PALETTE = [
    '#2563eb', '#16a34a', '#dc2626', '#9333ea',
    '#ea580c', '#0891b2', '#d97706', '#4f46e5',
    '#0d9488', '#be185d',
]
function pickColor(title: string): string {
    const idx = title.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % PALETTE.length
    return PALETTE[idx]
}

// ─── Main page ────────────────────────────────────────────────────────────────

export default function BooksPage() {
    const navigate    = useNavigate()
    const { user }    = useAuth()
    const queryClient = useQueryClient()
    const canModerate = user?.role === 'moderator' || user?.role === 'admin'

    // Pagination
    const [page, setPage] = useState<number>(1)
    const limit = 20

    // Filter / sort / view
    const [search,      setSearch]      = useState('')
    const [genre,       setGenre]       = useState('')
    const [sort,        setSort]        = useState<SortKey>('default')
    const [viewMode,    setViewMode]    = useState<ViewMode>('grid')
    const [noCoverOnly, setNoCoverOnly] = useState(false)

    // Add-to-library modal
    const [addTarget,          setAddTarget]          = useState<Book | null>(null)
    const [selectedEditionId,  setSelectedEditionId]  = useState('')
    const [selectedCondition,  setSelectedCondition]  = useState('good')
    const [addError,           setAddError]           = useState('')

    const { data, isLoading, isError } = useQuery({
        queryKey: ['books', page],
        queryFn: () => listBooks(page, limit),
        placeholderData: prev => prev,
    })

    const addCopyMutation = useMutation({
        mutationFn: () => addCopy(selectedEditionId, { condition: selectedCondition }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['my-library'] })
            setAddTarget(null)
            setAddError('')
        },
        onError: () => setAddError('Failed to add to library. Please try again.'),
    })

    // Client-side filter + sort
    const books = data?.books ?? []
    const allGenres = useMemo(
        () => [...new Set(books.flatMap(b => (b.genres ?? []).map(g => g.name)))].sort(),
        [books],
    )

    const filtered = useMemo(() => {
        let list = books
        if (search)      list = list.filter(b =>
            b.title.toLowerCase().includes(search.toLowerCase()) ||
            b.authors.some(a => a.name.toLowerCase().includes(search.toLowerCase()))
        )
        if (genre) list = list.filter(b => (b.genres ?? []).some(g => g.name === genre))
        if (noCoverOnly) list = list.filter(b => !((b as any).cover_image_url ?? b.cover_url))
        switch (sort) {
            case 'title_asc':  list = [...list].sort((a, b) => a.title.localeCompare(b.title)); break
            case 'title_desc': list = [...list].sort((a, b) => b.title.localeCompare(a.title)); break
            case 'newest':     list = [...list].sort((a, b) => b.created_at.localeCompare(a.created_at)); break
            case 'oldest':     list = [...list].sort((a, b) => a.created_at.localeCompare(b.created_at)); break
        }
        return list
    }, [books, search, genre, noCoverOnly, sort])

    const totalPages = data ? Math.ceil(data.total / limit) : 1

    function openAddModal(book: Book) {
        setAddTarget(book)
        setSelectedEditionId(book.editions?.[0]?.id ?? '')
        setSelectedCondition('good')
        setAddError('')
    }

    return (
        <div style={{
            minHeight: 'calc(100vh - 56px)',
            padding: '32px 24px',
            maxWidth: '1100px',
            margin: '0 auto',
        }}>
            {/* ── Header ── */}
            <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: '20px',
                flexWrap: 'wrap',
                gap: '12px',
            }}>
                <div>
                    <h1 style={{
                        margin: 0,
                        fontSize: '24px',
                        fontWeight: 700,
                        color: 'var(--color-text)',
                        fontFamily: 'var(--font-heading)',
                    }}>
                        Book Catalogue
                    </h1>
                    <p style={{
                        margin: '4px 0 0',
                        fontSize: '13px',
                        color: 'var(--color-text-muted)',
                        fontFamily: 'var(--font-body)',
                    }}>
                        {data ? `${data.total} approved book${data.total !== 1 ? 's' : ''}` : ''}
                    </p>
                </div>

                <button
                    onClick={() => navigate('/books/add')}
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                        padding: '10px 18px',
                        background: 'var(--color-primary)',
                        color: 'var(--color-primary-text)',
                        border: 'none',
                        borderRadius: 'var(--border-radius)',
                        fontSize: '13px',
                        fontWeight: 600,
                        cursor: 'pointer',
                        transition: 'var(--transition)',
                        fontFamily: 'var(--font-body)',
                    }}
                    onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary-hover)'}
                    onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.background = 'var(--color-primary)'}
                >
                    <span>➕</span>
                    <span>Add a book</span>
                </button>
            </div>

            {/* ── Filter bar ── */}
            <div style={{
                display: 'flex',
                gap: '10px',
                marginBottom: '20px',
                flexWrap: 'wrap',
                alignItems: 'center',
                background: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--border-radius)',
                padding: '12px 14px',
            }}>
                {/* Search */}
                <input
                    type="text"
                    placeholder="Search title or author…"
                    value={search}
                    onChange={e => setSearch(e.target.value)}
                    style={{
                        flex: '1 1 180px',
                        padding: '8px 10px',
                        background: 'var(--input-bg)',
                        border: '1px solid var(--color-border)',
                        borderRadius: 'var(--border-radius)',
                        color: 'var(--color-text)',
                        fontSize: '13px',
                        fontFamily: 'var(--font-body)',
                        outline: 'none',
                    }}
                />

                {/* Genre */}
                <select
                    value={genre}
                    onChange={e => setGenre(e.target.value)}
                    style={{
                        padding: '8px 10px',
                        background: 'var(--input-bg)',
                        border: '1px solid var(--color-border)',
                        borderRadius: 'var(--border-radius)',
                        color: 'var(--color-text)',
                        fontSize: '13px',
                        fontFamily: 'var(--font-body)',
                    }}
                >
                    <option value="">All genres</option>
                    {allGenres.map(g => <option key={g} value={g}>{g}</option>)}
                </select>

                {/* Sort */}
                <select
                    value={sort}
                    onChange={e => setSort(e.target.value as SortKey)}
                    style={{
                        padding: '8px 10px',
                        background: 'var(--input-bg)',
                        border: '1px solid var(--color-border)',
                        borderRadius: 'var(--border-radius)',
                        color: 'var(--color-text)',
                        fontSize: '13px',
                        fontFamily: 'var(--font-body)',
                    }}
                >
                    <option value="default">Default order</option>
                    <option value="title_asc">Title A → Z</option>
                    <option value="title_desc">Title Z → A</option>
                    <option value="newest">Newest first</option>
                    <option value="oldest">Oldest first</option>
                </select>

                {/* Mod/admin: no-cover filter */}
                {canModerate && (
                    <label style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '6px',
                        fontSize: '12px',
                        color: 'var(--color-text-muted)',
                        fontFamily: 'var(--font-body)',
                        cursor: 'pointer',
                        whiteSpace: 'nowrap',
                    }}>
                        <input
                            type="checkbox"
                            checked={noCoverOnly}
                            onChange={e => setNoCoverOnly(e.target.checked)}
                            style={{ cursor: 'pointer' }}
                        />
                        No cover only
                    </label>
                )}

                {/* View toggle */}
                <div style={{ display: 'flex', gap: '4px', marginLeft: 'auto' }}>
                    {(['grid', 'list'] as ViewMode[]).map(mode => (
                        <button
                            key={mode}
                            onClick={() => setViewMode(mode)}
                            title={mode === 'grid' ? 'Grid view' : 'List view'}
                            style={{
                                width: '34px',
                                height: '34px',
                                background: viewMode === mode ? 'var(--color-primary)' : 'var(--color-surface-alt)',
                                color:      viewMode === mode ? 'var(--color-primary-text)' : 'var(--color-text-muted)',
                                border: `1px solid ${viewMode === mode ? 'var(--color-primary)' : 'var(--color-border)'}`,
                                borderRadius: 'var(--border-radius)',
                                cursor: 'pointer',
                                fontSize: '16px',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                transition: 'var(--transition)',
                            }}
                        >
                            {mode === 'grid' ? '⊞' : '☰'}
                        </button>
                    ))}
                </div>
            </div>

            {/* ── Content ── */}
            {isLoading ? (
                <div style={{ display: 'flex', justifyContent: 'center', paddingTop: '80px' }}>
                    <Spinner size="lg" label="Loading catalogue..." />
                </div>
            ) : isError ? (
                <Card padding="lg">
                    <EmptyState
                        icon="⚠️"
                        title="Failed to load books"
                        description="Something went wrong. Please try again."
                        action={{ label: 'Retry', onClick: () => window.location.reload() }}
                    />
                </Card>
            ) : filtered.length === 0 ? (
                <Card padding="lg">
                    <EmptyState
                        icon="📭"
                        title={data?.books?.length ? 'No books match your filters' : 'No books yet'}
                        description={data?.books?.length
                            ? 'Try adjusting the search or filter.'
                            : 'The catalogue is empty. Be the first to add a book!'}
                        action={!data?.books?.length
                            ? { label: 'Add a book', onClick: () => navigate('/books/add') }
                            : undefined}
                    />
                </Card>
            ) : (
                <>
                    {viewMode === 'grid' ? (
                        <div style={{
                            display: 'grid',
                            gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
                            gap: '16px',
                            marginBottom: '32px',
                        }}>
                            {filtered.map(book => (
                                <BookCard
                                    key={book.id}
                                    book={book}
                                    onClick={() => navigate(`/books/${book.id}`)}
                                    onWishlist={() => addToWishlist(book.id)}
                                    onAddToLibrary={() => openAddModal(book)}
                                />
                            ))}
                        </div>
                    ) : (
                        <div style={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: '8px',
                            marginBottom: '32px',
                        }}>
                            {filtered.map(book => (
                                <BookListRow
                                    key={book.id}
                                    book={book}
                                    onClick={() => navigate(`/books/${book.id}`)}
                                    onWishlist={() => addToWishlist(book.id)}
                                    onAddToLibrary={() => openAddModal(book)}
                                />
                            ))}
                        </div>
                    )}

                    {totalPages > 1 && (
                        <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
                    )}
                </>
            )}

            {/* ── Add to library modal ── */}
            <Modal
                isOpen={!!addTarget}
                onClose={() => setAddTarget(null)}
                title={`Add to library — ${addTarget?.title ?? ''}`}
                confirmLabel="Add to library"
                onConfirm={() => addCopyMutation.mutate()}
                isLoading={addCopyMutation.isPending}
                size="sm"
            >
                <div style={{ display: 'flex', flexDirection: 'column', gap: '14px' }}>
                    {/* Edition picker */}
                    <div>
                        <label style={{
                            display: 'block',
                            fontSize: '11px',
                            fontWeight: 700,
                            textTransform: 'uppercase',
                            letterSpacing: '0.06em',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                            marginBottom: '6px',
                        }}>
                            Edition
                        </label>
                        <select
                            value={selectedEditionId}
                            onChange={e => setSelectedEditionId(e.target.value)}
                            style={{
                                width: '100%',
                                padding: '8px 10px',
                                background: 'var(--input-bg)',
                                border: '1px solid var(--color-border)',
                                borderRadius: 'var(--border-radius)',
                                color: 'var(--color-text)',
                                fontSize: '13px',
                                fontFamily: 'var(--font-body)',
                            }}
                        >
                            {addTarget?.editions?.map((ed: Edition) => (
                                <option key={ed.id} value={ed.id}>
                                    {ed.format}
                                    {ed.language ? ` · ${ed.language.toUpperCase()}` : ''}
                                    {ed.publisher ? ` · ${ed.publisher}` : ''}
                                    {ed.isbn ? ` · ${ed.isbn}` : ''}
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* Condition */}
                    <div>
                        <label style={{
                            display: 'block',
                            fontSize: '11px',
                            fontWeight: 700,
                            textTransform: 'uppercase',
                            letterSpacing: '0.06em',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                            marginBottom: '6px',
                        }}>
                            Condition
                        </label>
                        <div style={{ display: 'flex', gap: '6px' }}>
                            {['new', 'good', 'fair', 'poor'].map(c => (
                                <button
                                    key={c}
                                    onClick={() => setSelectedCondition(c)}
                                    style={{
                                        flex: 1,
                                        padding: '6px 0',
                                        border: `1px solid ${selectedCondition === c ? 'var(--color-primary)' : 'var(--color-border)'}`,
                                        borderRadius: 'var(--border-radius)',
                                        background: selectedCondition === c ? 'var(--color-primary)' : 'var(--color-surface-alt)',
                                        color: selectedCondition === c ? 'var(--color-primary-text)' : 'var(--color-text)',
                                        fontSize: '12px',
                                        fontFamily: 'var(--font-body)',
                                        fontWeight: selectedCondition === c ? 600 : 400,
                                        cursor: 'pointer',
                                        textTransform: 'capitalize',
                                        transition: 'var(--transition)',
                                    }}
                                >
                                    {c}
                                </button>
                            ))}
                        </div>
                    </div>

                    {addError && (
                        <p style={{
                            margin: 0,
                            fontSize: '12px',
                            color: 'var(--color-error)',
                            fontFamily: 'var(--font-body)',
                        }}>
                            {addError}
                        </p>
                    )}
                </div>
            </Modal>
        </div>
    )
}

// ─── Book Card (grid view) ────────────────────────────────────────────────────

function BookCard({
                      book, onClick, onWishlist, onAddToLibrary,
                  }: {
    book: Book
    onClick: () => void
    onWishlist: () => void
    onAddToLibrary: () => void
}) {
    const color    = pickColor(book.title)
    const coverSrc = (book as any).cover_image_url ?? book.cover_url

    return (
        <div
            onClick={onClick}
            style={{
                background: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--border-radius)',
                overflow: 'hidden',
                cursor: 'pointer',
                transition: 'var(--transition)',
                boxShadow: 'var(--shadow-sm)',
                display: 'flex',
                flexDirection: 'column',
            }}
            onMouseEnter={e => {
                (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-md)'
                ;(e.currentTarget as HTMLDivElement).style.transform = 'translateY(-3px)'
            }}
            onMouseLeave={e => {
                (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-sm)'
                ;(e.currentTarget as HTMLDivElement).style.transform = 'translateY(0)'
            }}
        >
            {/* Cover */}
            <div style={{
                height: '160px',
                background: coverSrc ? undefined : `linear-gradient(135deg, ${color}dd, ${color}88)`,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
                position: 'relative',
                overflow: 'hidden',
            }}>
                {coverSrc
                    ? <img src={coverSrc} alt={book.title} style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                    : <span style={{ fontSize: '40px', opacity: 0.6 }}>📖</span>
                }

                {/* Genre badge */}
                {book.genres && book.genres.length > 0 && (
                    <div style={{ position: 'absolute', bottom: '8px', left: '8px' }}>
                        <Badge label={book.genres[0].name} variant="info" size="sm" />
                    </div>
                )}

                {/* Action icons */}
                <div
                    onClick={e => e.stopPropagation()}
                    style={{
                        position: 'absolute',
                        top: '6px',
                        right: '6px',
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '4px',
                    }}
                >
                    <ActionIcon icon="♥" title="Add to wishlist"  onClick={onWishlist} />
                    <ActionIcon icon="➕" title="Add to library"   onClick={onAddToLibrary} />
                </div>
            </div>

            {/* Info */}
            <div style={{
                padding: '12px',
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                gap: '4px',
            }}>
                <p style={{
                    margin: 0,
                    fontSize: '13px',
                    fontWeight: 600,
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-body)',
                    overflow: 'hidden',
                    display: '-webkit-box',
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: 'vertical',
                    lineHeight: '1.4',
                }}>
                    {book.title}
                </p>
                <p style={{
                    margin: 0,
                    fontSize: '11px',
                    color: 'var(--color-text-muted)',
                    fontFamily: 'var(--font-body)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                }}>
                    {book.authors?.length > 0 ? book.authors.map(a => a.name).join(', ') : 'Unknown author'}
                </p>
            </div>
        </div>
    )
}

// ─── Book List Row (list view) ────────────────────────────────────────────────

function BookListRow({
                         book, onClick, onWishlist, onAddToLibrary,
                     }: {
    book: Book
    onClick: () => void
    onWishlist: () => void
    onAddToLibrary: () => void
}) {
    const color = pickColor(book.title)

    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                gap: '12px',
                padding: '10px 12px',
                background: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--border-radius)',
                boxShadow: 'var(--shadow-sm)',
                transition: 'var(--transition)',
            }}
            onMouseEnter={e => (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-md)'}
            onMouseLeave={e => (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-sm)'}
        >
            {/* Thumbnail */}
            <div
                onClick={onClick}
                style={{
                    width: '40px',
                    height: '56px',
                    borderRadius: '4px',
                    flexShrink: 0,
                    background: ((book as any).cover_image_url ?? book.cover_url)
                        ? undefined
                        : `linear-gradient(135deg, ${color}dd, ${color}88)`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontSize: '18px',
                    cursor: 'pointer',
                    border: '1px solid var(--color-border)',
                    overflow: 'hidden',
                }}
            >
                {((book as any).cover_image_url ?? book.cover_url)
                    ? <img src={(book as any).cover_image_url ?? book.cover_url!} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                    : '📖'
                }
            </div>

            {/* Title + authors */}
            <div onClick={onClick} style={{ flex: 1, minWidth: 0, cursor: 'pointer' }}>
                <p style={{
                    margin: '0 0 2px',
                    fontSize: '14px',
                    fontWeight: 600,
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-body)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                }}>
                    {book.title}
                </p>
                <p style={{
                    margin: 0,
                    fontSize: '12px',
                    color: 'var(--color-text-muted)',
                    fontFamily: 'var(--font-body)',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                }}>
                    {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                </p>
            </div>

            {/* Genres */}
            <div style={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                {book.genres?.slice(0, 2).map(g => (
                    <Badge key={g.id} label={g.name} variant="info" size="sm" />
                ))}
            </div>

            {/* Action icons */}
            <div style={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                <ActionIcon icon="♥" title="Add to wishlist"  onClick={onWishlist} />
                <ActionIcon icon="➕" title="Add to library"   onClick={onAddToLibrary} />
            </div>
        </div>
    )
}

// ─── Small icon button ────────────────────────────────────────────────────────

function ActionIcon({ icon, title, onClick }: { icon: string; title: string; onClick: () => void }) {
    return (
        <button
            title={title}
            onClick={e => { e.stopPropagation(); onClick() }}
            style={{
                width: '28px',
                height: '28px',
                border: 'none',
                borderRadius: 'var(--border-radius)',
                background: 'rgba(0,0,0,0.45)',
                backdropFilter: 'blur(4px)',
                color: '#fff',
                fontSize: '13px',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                transition: 'var(--transition)',
                flexShrink: 0,
            }}
            onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.background = 'rgba(0,0,0,0.70)'}
            onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.background = 'rgba(0,0,0,0.45)'}
        >
            {icon}
        </button>
    )
}

// ─── Pagination ───────────────────────────────────────────────────────────────

function Pagination({
                        page, totalPages, onPageChange,
                    }: {
    page: number
    totalPages: number
    onPageChange: (p: number) => void
}) {
    const pages: (number | '...')[] = []

    if (totalPages <= 7) {
        for (let i = 1; i <= totalPages; i++) pages.push(i)
    } else {
        pages.push(1)
        if (page > 3) pages.push('...')
        for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) pages.push(i)
        if (page < totalPages - 2) pages.push('...')
        pages.push(totalPages)
    }

    return (
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '6px' }}>
            <PageBtn label="←" onClick={() => onPageChange(page - 1)} disabled={page === 1} />
            {pages.map((p, i) =>
                p === '...'
                    ? <span key={`e-${i}`} style={{ color: 'var(--color-text-muted)', padding: '0 4px' }}>…</span>
                    : <PageBtn key={p} label={String(p)} onClick={() => onPageChange(p as number)} active={p === page} />
            )}
            <PageBtn label="→" onClick={() => onPageChange(page + 1)} disabled={page === totalPages} />
        </div>
    )
}

function PageBtn({
                     label, onClick, active = false, disabled = false,
                 }: {
    label: string; onClick: () => void; active?: boolean; disabled?: boolean
}) {
    return (
        <button
            onClick={onClick}
            disabled={disabled}
            style={{
                minWidth: '36px',
                height: '36px',
                padding: '0 10px',
                background: active ? 'var(--color-primary)' : 'var(--color-surface)',
                color: active ? 'var(--color-primary-text)' : 'var(--color-text)',
                border: `1px solid ${active ? 'var(--color-primary)' : 'var(--color-border)'}`,
                borderRadius: 'var(--border-radius)',
                fontSize: '13px',
                fontWeight: active ? 600 : 400,
                cursor: disabled ? 'not-allowed' : 'pointer',
                opacity: disabled ? 0.4 : 1,
                transition: 'var(--transition)',
                fontFamily: 'var(--font-body)',
            }}
        >
            {label}
        </button>
    )
}