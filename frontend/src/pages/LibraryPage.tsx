import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query'
import { getMyLibrary, updateReadingStatus, removeCopy } from '../api/books'
import { Card, Badge, Spinner, EmptyState, Modal } from '../components'
import { useTheme } from '../context/ThemeContext'
import type { UserBook } from '../types'

// ─── Constants ────────────────────────────────────────────────────────────────

const STATUS_LABELS: Record<string, string> = {
    want_to_read: 'Want to Read',
    reading: 'Reading',
    read: 'Read',
}
const STATUS_VARIANTS: Record<string, 'default' | 'info' | 'success'> = {
    want_to_read: 'default',
    reading: 'info',
    read: 'success',
}

const PALETTE = [
    '#2563eb', '#16a34a', '#dc2626', '#9333ea',
    '#ea580c', '#0891b2', '#d97706', '#4f46e5',
    '#0d9488', '#be185d',
]

const SHELF_COLORS: Record<string, { shelf: string; shelfShadow: string; wall: string; frame: string }> = {
    'default-light':    { shelf: '#d4b896', shelfShadow: '#a07850', wall: '#f5efe6', frame: '#c4a882' },
    'woody':            { shelf: '#5c3317', shelfShadow: '#2d1608', wall: '#1a0c05', frame: '#4a2810' },
    'nordic':           { shelf: '#c8d8e8', shelfShadow: '#8fa8c0', wall: '#eef2f8', frame: '#b8ccd8' },
    'metallic':         { shelf: '#3a3a3a', shelfShadow: '#181818', wall: '#141414', frame: '#2a2a2a' },
    'futuristic':       { shelf: '#0a1a2a', shelfShadow: '#020810', wall: '#020510', frame: '#050f1a' },
    'post-apocalyptic': { shelf: '#3d2e10', shelfShadow: '#1a1206', wall: '#0d0904', frame: '#2d2008' },
    'dark-academia':    { shelf: '#3d2e18', shelfShadow: '#1a1208', wall: '#100d08', frame: '#2d2010' },
    'ocean':            { shelf: '#0a2040', shelfShadow: '#040f20', wall: '#030d20', frame: '#061830' },
    'space':            { shelf: '#0d0820', shelfShadow: '#04030a', wall: '#030108', frame: '#08051a' },
    'candlelight':      { shelf: '#2e2620', shelfShadow: '#141008', wall: '#100e0a', frame: '#231e18' },
}

function getShelfColors(themeId: string) {
    return SHELF_COLORS[themeId] ?? SHELF_COLORS['default-light']
}

function pickColor(title: string): string {
    const idx = title.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % PALETTE.length
    return PALETTE[idx]
}

// 4 view modes
type ViewMode = 'catalogue' | 'card' | 'list' | 'bookcase'

// ─── Main page ────────────────────────────────────────────────────────────────

export default function LibraryPage() {
    const navigate    = useNavigate()
    const queryClient = useQueryClient()
    const { themeId } = useTheme()

    const [page,         setPage]         = useState(1)
    const [filter,       setFilter]       = useState<string>('all')
    const [viewMode,     setViewMode]     = useState<ViewMode>('catalogue')
    const [removeTarget, setRemoveTarget] = useState<UserBook | null>(null)
    const LIMIT = 20

    const { data, isLoading } = useQuery({
        queryKey: ['my-library', page],
        queryFn: () => getMyLibrary(page, LIMIT),
        placeholderData: keepPreviousData,
    })

    const statusMutation = useMutation({
        mutationFn: ({ copyId, status }: { copyId: string; status: 'want_to_read' | 'reading' | 'read' }) =>
            updateReadingStatus(copyId, { status }),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['my-library'] }),
    })

    const removeMutation = useMutation({
        mutationFn: (copyId: string) => removeCopy(copyId),
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: ['my-library'] })
            await queryClient.invalidateQueries({ queryKey: ['my-books'] })
            setRemoveTarget(null)
        },
    })

    const books: UserBook[] = data?.books ?? []
    const filtered = filter === 'all' ? books : books.filter(b => b.reading_status === filter)

    const counts = books.reduce((acc, b) => {
        acc[b.reading_status] = (acc[b.reading_status] ?? 0) + 1
        return acc
    }, {} as Record<string, number>)

    const totalPages = data ? Math.ceil(data.total / LIMIT) : 1

    return (
        <div style={{ minHeight: 'calc(100vh - 56px)', padding: '32px 24px', maxWidth: viewMode === 'bookcase' ? '100%' : '1100px', margin: '0 auto' }}>
            {/* ── Header ── */}
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '20px', flexWrap: 'wrap', gap: '12px' }}>
                <div>
                    <h1 style={{ margin: 0, fontSize: '24px', fontWeight: 700, color: 'var(--color-text)', fontFamily: 'var(--font-heading)' }}>
                        My Library
                    </h1>
                    <p style={{ margin: '4px 0 0', fontSize: '13px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                        {data?.total ?? 0} book{data?.total !== 1 ? 's' : ''} in your collection
                    </p>
                </div>
                <button
                    onClick={() => navigate('/books')}
                    style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '10px 18px', background: 'var(--color-primary)', color: 'var(--color-primary-text)', border: 'none', borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: 600, cursor: 'pointer', fontFamily: 'var(--font-body)' }}
                >
                    ➕ Add books
                </button>
            </div>

            {/* ── Filter tabs + view toggle ── */}
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '20px', flexWrap: 'wrap', gap: '10px' }}>
                {/* Status filter tabs */}
                <div style={{ display: 'flex', gap: '4px', background: 'var(--color-surface-alt)', padding: '4px', borderRadius: 'var(--border-radius)', width: 'fit-content' }}>
                    {[
                        { key: 'all',          label: `All (${books.length})` },
                        { key: 'want_to_read', label: `Want to Read (${counts.want_to_read ?? 0})` },
                        { key: 'reading',      label: `Reading (${counts.reading ?? 0})` },
                        { key: 'read',         label: `Read (${counts.read ?? 0})` },
                    ].map(tab => (
                        <button
                            key={tab.key}
                            onClick={() => setFilter(tab.key)}
                            style={{ padding: '6px 14px', border: 'none', borderRadius: 'var(--border-radius)', background: filter === tab.key ? 'var(--color-surface)' : 'transparent', color: filter === tab.key ? 'var(--color-text)' : 'var(--color-text-muted)', fontSize: '12px', fontWeight: filter === tab.key ? 600 : 400, cursor: 'pointer', transition: 'var(--transition)', boxShadow: filter === tab.key ? 'var(--shadow-sm)' : 'none', fontFamily: 'var(--font-body)', whiteSpace: 'nowrap' }}
                        >
                            {tab.label}
                        </button>
                    ))}
                </div>

                {/* View mode toggle — 4 options */}
                <div style={{ display: 'flex', gap: '4px' }}>
                    {([
                        { mode: 'catalogue', icon: '⊞', title: 'Catalogue grid' },
                        { mode: 'card',      icon: '🃏', title: 'Card view'      },
                        { mode: 'list',      icon: '☰', title: 'List view'      },
                        { mode: 'bookcase',  icon: '📚', title: 'Bookcase view'  },
                    ] as { mode: ViewMode; icon: string; title: string }[]).map(({ mode, icon, title }) => (
                        <button
                            key={mode}
                            onClick={() => setViewMode(mode)}
                            title={title}
                            style={{ width: '34px', height: '34px', background: viewMode === mode ? 'var(--color-primary)' : 'var(--color-surface-alt)', color: viewMode === mode ? 'var(--color-primary-text)' : 'var(--color-text-muted)', border: `1px solid ${viewMode === mode ? 'var(--color-primary)' : 'var(--color-border)'}`, borderRadius: 'var(--border-radius)', cursor: 'pointer', fontSize: '16px', display: 'flex', alignItems: 'center', justifyContent: 'center', transition: 'var(--transition)' }}
                        >
                            {icon}
                        </button>
                    ))}
                </div>
            </div>

            {/* ── Content ── */}
            {isLoading ? (
                <div style={{ display: 'flex', justifyContent: 'center', paddingTop: '80px' }}>
                    <Spinner size="lg" label="Loading your library..." />
                </div>
            ) : filtered.length === 0 ? (
                <Card padding="lg">
                    <EmptyState
                        icon="📚"
                        title={filter === 'all' ? 'Your library is empty' : `No books with status "${STATUS_LABELS[filter]}"`}
                        description={filter === 'all' ? 'Browse the catalogue and add books to your collection.' : 'Change the filter to see other books.'}
                        action={filter === 'all' ? { label: 'Browse catalogue', onClick: () => navigate('/books') } : undefined}
                    />
                </Card>
            ) : viewMode === 'catalogue' ? (
                <CatalogueView
                    books={filtered}
                    onNavigate={(bookId, copyId) => navigate(`/books/${bookId}?copy_id=${copyId}`)}
                    onStatusChange={(copyId, status) => statusMutation.mutate({ copyId, status })}
                    onRemove={b => setRemoveTarget(b)}
                />
            ) : viewMode === 'card' ? (
                <CardView
                    books={filtered}
                    onStatusChange={(copyId, status) => statusMutation.mutate({ copyId, status })}
                    onNavigate={(bookId, copyId) => navigate(`/books/${bookId}?copy_id=${copyId}`)}
                    onRemove={b => setRemoveTarget(b)}
                    isUpdating={statusMutation.isPending}
                />
            ) : viewMode === 'bookcase' ? (
                <BookcaseView
                    books={filtered}
                    themeId={themeId}
                    onNavigate={(bookId, copyId) => navigate(`/books/${bookId}?copy_id=${copyId}`)}
                />
            ) : (
                <ListView
                    books={filtered}
                    onStatusChange={(copyId, status) => statusMutation.mutate({ copyId, status })}
                    onNavigate={(bookId, copyId) => navigate(`/books/${bookId}?copy_id=${copyId}`)}
                    onRemove={b => setRemoveTarget(b)}
                    isUpdating={statusMutation.isPending}
                />
            )}

            {totalPages > 1 && (
                <div style={{ marginTop: '32px' }}>
                    <Pagination
                        page={page}
                        totalPages={totalPages}
                        onPageChange={p => { setPage(p); window.scrollTo({ top: 0, behavior: 'smooth' }) }}
                    />
                </div>
            )}

            {/* ── Confirm remove modal ── */}
            <Modal
                isOpen={!!removeTarget}
                onClose={() => setRemoveTarget(null)}
                title="Remove from library"
                confirmLabel="Remove"
                confirmVariant="danger"
                onConfirm={() => removeTarget && removeMutation.mutate(removeTarget.copy_id)}
                isLoading={removeMutation.isPending}
                size="sm"
            >
                <p style={{ margin: 0, fontSize: '14px', color: 'var(--color-text)', fontFamily: 'var(--font-body)', lineHeight: '1.6' }}>
                    Remove <strong>{removeTarget?.book.title}</strong> from your library? This cannot be undone.
                </p>
            </Modal>
        </div>
    )
}

// ─── View 1: Catalogue (tall cover cards, like BooksPage) ─────────────────────

function CatalogueView({
                           books, onNavigate, onStatusChange, onRemove,
                       }: {
    books: UserBook[]
    onNavigate: (bookId: string, copyId: string) => void
    onStatusChange: (copyId: string, status: 'reading' | 'read' | 'want_to_read') => void
    onRemove: (b: UserBook) => void
}) {
    return (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: '16px' }}>
            {books.map(ub => {
                const { book } = ub
                const color = pickColor(book.title)
                const coverSrc = ub.cover_url

                return (
                    <div
                        key={ub.copy_id}
                        style={{ background: 'var(--color-surface)', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', overflow: 'hidden', cursor: 'pointer', transition: 'var(--transition)', boxShadow: 'var(--shadow-sm)', display: 'flex', flexDirection: 'column' }}
                        onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-md)'; (e.currentTarget as HTMLDivElement).style.transform = 'translateY(-3px)' }}
                        onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.boxShadow = 'var(--shadow-sm)'; (e.currentTarget as HTMLDivElement).style.transform = 'translateY(0)' }}
                    >
                        {/* Cover */}
                        <div
                            onClick={() => onNavigate(book.id, ub.copy_id)}
                            style={{ height: '160px', background: coverSrc ? `url(${coverSrc}) center/cover no-repeat` : `linear-gradient(135deg, ${color}dd, ${color}88)`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0, position: 'relative' }}
                        >
                            {!coverSrc && <span style={{ fontSize: '36px', opacity: 0.6 }}>📖</span>}
                            {/* Status badge */}
                            <div style={{ position: 'absolute', bottom: '6px', left: '6px' }}>
                                <Badge label={STATUS_LABELS[ub.reading_status]} variant={STATUS_VARIANTS[ub.reading_status]} size="sm" />
                            </div>
                        </div>
                        {/* Info */}
                        <div style={{ padding: '10px', flex: 1, display: 'flex', flexDirection: 'column', gap: '4px' }}>
                            <p onClick={() => onNavigate(book.id, ub.copy_id)} style={{ margin: 0, fontSize: '12px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-body)', overflow: 'hidden', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', lineHeight: '1.4', cursor: 'pointer' }}>
                                {book.title}
                            </p>
                            <p style={{ margin: 0, fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                                {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                            </p>
                            {ub.format && <Badge label={ub.format} variant="muted" size="sm" />}
                        </div>
                        {/* Quick status bar */}
                        <div style={{ borderTop: '1px solid var(--color-border)', display: 'flex' }}>
                            {(['want_to_read', 'reading', 'read'] as const).map(s => (
                                <button
                                    key={s}
                                    onClick={() => onStatusChange(ub.copy_id, s)}
                                    title={STATUS_LABELS[s]}
                                    style={{ flex: 1, padding: '6px 0', border: 'none', background: ub.reading_status === s ? 'var(--color-primary)' : 'transparent', color: ub.reading_status === s ? 'var(--color-primary-text)' : 'var(--color-text-muted)', fontSize: '13px', cursor: 'pointer', transition: 'var(--transition)' }}
                                >
                                    {s === 'want_to_read' ? '🔖' : s === 'reading' ? '📖' : '✅'}
                                </button>
                            ))}
                            <button
                                onClick={() => onRemove(ub)}
                                title="Remove"
                                style={{ padding: '6px 8px', border: 'none', background: 'transparent', color: 'var(--color-error)', fontSize: '13px', cursor: 'pointer', opacity: 0.5, transition: 'var(--transition)' }}
                                onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.opacity = '1'}
                                onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.opacity = '0.5'}
                            >
                                🗑️
                            </button>
                        </div>
                    </div>
                )
            })}
        </div>
    )
}

// ─── View 2: Card (cover + info side by side) ─────────────────────────────────

function CardView({
                      books, onStatusChange, onNavigate, onRemove, isUpdating,
                  }: {
    books: UserBook[]
    onStatusChange: (copyId: string, status: 'reading' | 'read' | 'want_to_read') => void
    onNavigate: (bookId: string, copyId: string) => void
    onRemove: (b: UserBook) => void
    isUpdating: boolean
}) {
    return (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '16px' }}>
            {books.map(ub => (
                <LibraryBookCard
                    key={ub.copy_id}
                    userBook={ub}
                    onStatusChange={status => onStatusChange(ub.copy_id, status)}
                    onNavigate={() => onNavigate(ub.book.id, ub.copy_id)}
                    onRemove={() => onRemove(ub)}
                    isUpdating={isUpdating}
                />
            ))}
        </div>
    )
}

// ─── View 3: Bookcase ─────────────────────────────────────────────────────────

const BOOKS_PER_SHELF = 14

function BookcaseView({
                          books, themeId, onNavigate,
                      }: {
    books: UserBook[]
    themeId: string
    onNavigate: (bookId: string, copyId: string) => void
}) {
    const sc = getShelfColors(themeId)
    const SHELF_H      = 130
    const SHELF_THICK  = 14
    const FRAME_TOP    = 20
    const FRAME_BOTTOM = 24

    const shelves = useMemo(() => {
        const rows: UserBook[][] = []
        for (let i = 0; i < books.length; i += BOOKS_PER_SHELF) {
            rows.push(books.slice(i, i + BOOKS_PER_SHELF))
        }
        if (rows.length === 0) rows.push([])
        return rows
    }, [books])

    return (
        <div style={{ borderRadius: 'var(--border-radius)', overflow: 'hidden', boxShadow: 'var(--shadow-lg)', border: `4px solid ${sc.frame}` }}>
            <div style={{ height: `${FRAME_TOP}px`, background: sc.shelf, borderBottom: `2px solid ${sc.shelfShadow}` }} />
            {shelves.map((shelfBooks, si) => (
                <div key={si} style={{ display: 'flex', flexDirection: 'column', background: sc.wall }}>
                    <div style={{ height: `${SHELF_H - SHELF_THICK}px`, display: 'flex', alignItems: 'flex-end', padding: '0 8px', gap: '2px', overflow: 'hidden' }}>
                        {shelfBooks.map(ub => (
                            <BookSpine
                                key={ub.copy_id}
                                userBook={ub}
                                shelfHeight={SHELF_H - SHELF_THICK - 4}
                                onClick={() => onNavigate(ub.book.id, ub.copy_id)}
                            />
                        ))}
                        {shelfBooks.length < BOOKS_PER_SHELF && <div style={{ flex: 1 }} />}
                    </div>
                    <div style={{ height: `${SHELF_THICK}px`, background: `linear-gradient(to bottom, ${sc.shelf}, ${sc.shelfShadow})`, borderTop: `2px solid ${lighten(sc.shelf)}`, boxShadow: '0 4px 12px rgba(0,0,0,0.4)' }} />
                </div>
            ))}
            <div style={{ height: `${FRAME_BOTTOM}px`, background: sc.shelf, borderTop: `2px solid ${sc.shelfShadow}` }} />
        </div>
    )
}

function lighten(hex: string): string {
    const n = parseInt(hex.replace('#', ''), 16)
    const r = Math.min(255, ((n >> 16) & 0xff) + 40)
    const g = Math.min(255, ((n >> 8)  & 0xff) + 40)
    const b = Math.min(255, (n         & 0xff) + 40)
    return `#${r.toString(16).padStart(2,'0')}${g.toString(16).padStart(2,'0')}${b.toString(16).padStart(2,'0')}`
}

function BookSpine({ userBook, shelfHeight, onClick }: { userBook: UserBook; shelfHeight: number; onClick: () => void }) {
    const format = userBook.format?.toLowerCase() ?? ''
    const isHardcover = format === 'hardcover'
    const isEbook     = format === 'ebook'
    const isAudiobook = format === 'audiobook'

    const spineWidth  = isHardcover ? 26 : 22
    const spineHeight = Math.round(shelfHeight * (isHardcover ? 0.92 : 0.78))
    const baseColor   = pickColor(userBook.book.title)

    const authorInitials = (userBook.book.authors ?? [])
        .map(a => a.name.split(' ').map(w => w[0]).join('').slice(0, 3).toUpperCase())
        .join(' ')

    const titleShort = userBook.book.title.length > 22
        ? userBook.book.title.slice(0, 20) + '…'
        : userBook.book.title

    return (
        <div
            title={`${userBook.book.title}\n${userBook.book.authors?.map(a => a.name).join(', ')}`}
            onClick={onClick}
            style={{ width: `${spineWidth}px`, height: `${spineHeight}px`, flexShrink: 0, borderRadius: '2px 2px 0 0', cursor: 'pointer', position: 'relative', overflow: 'hidden', background: `linear-gradient(to right, ${adjustBrightness(baseColor, 20)}, ${baseColor}, ${adjustBrightness(baseColor, -20)})`, boxShadow: '2px 0 4px rgba(0,0,0,0.3)', transition: 'transform 0.15s, box-shadow 0.15s', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'flex-start', paddingTop: '4px' }}
            onMouseEnter={e => { (e.currentTarget as HTMLDivElement).style.transform = 'translateY(-6px)'; (e.currentTarget as HTMLDivElement).style.boxShadow = '4px 0 12px rgba(0,0,0,0.5)' }}
            onMouseLeave={e => { (e.currentTarget as HTMLDivElement).style.transform = 'translateY(0)'; (e.currentTarget as HTMLDivElement).style.boxShadow = '2px 0 4px rgba(0,0,0,0.3)' }}
        >
            <div style={{ position: 'absolute', top: 0, left: 0, width: '3px', height: '100%', background: 'rgba(255,255,255,0.2)', borderRadius: '2px 0 0 0' }} />
            <div style={{ position: 'absolute', top: 0, right: 0, width: '3px', height: '100%', background: 'rgba(0,0,0,0.25)' }} />
            <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: '2px', background: 'rgba(255,255,255,0.35)', borderRadius: '2px 2px 0 0' }} />
            <div style={{ position: 'absolute', top: 0, bottom: 0, left: 0, right: 0, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: '2px', padding: '4px 2px' }}>
                <span style={{ writingMode: 'vertical-rl', textOrientation: 'mixed', transform: 'rotate(180deg)', fontSize: '7px', fontWeight: 700, color: 'rgba(255,255,255,0.7)', fontFamily: 'var(--font-body)', letterSpacing: '0.05em', overflow: 'hidden', maxHeight: '36px', whiteSpace: 'nowrap' }}>
                    {authorInitials}
                </span>
                <span style={{ writingMode: 'vertical-rl', textOrientation: 'mixed', transform: 'rotate(180deg)', fontSize: '8px', fontWeight: 600, color: 'rgba(255,255,255,0.9)', fontFamily: 'var(--font-body)', overflow: 'hidden', flex: 1, whiteSpace: 'nowrap', maxHeight: `${spineHeight - 50}px` }}>
                    {titleShort}
                </span>
            </div>
            {(isAudiobook || isEbook) && (
                <div style={{ position: 'absolute', bottom: '3px', left: 0, right: 0, textAlign: 'center', fontSize: '9px' }}>
                    {isAudiobook ? '🎧' : '📱'}
                </div>
            )}
            {isHardcover && (
                <div style={{ position: 'absolute', top: '3px', left: '2px', right: '2px', height: '1px', background: 'rgba(255,255,255,0.3)' }} />
            )}
        </div>
    )
}

function adjustBrightness(hex: string, delta: number): string {
    const n = parseInt(hex.replace('#', ''), 16)
    const clamp = (v: number) => Math.max(0, Math.min(255, v))
    const r = clamp(((n >> 16) & 0xff) + delta)
    const g = clamp(((n >> 8)  & 0xff) + delta)
    const b = clamp((n         & 0xff) + delta)
    return `#${r.toString(16).padStart(2,'0')}${g.toString(16).padStart(2,'0')}${b.toString(16).padStart(2,'0')}`
}

// ─── View 4: List ─────────────────────────────────────────────────────────────

function ListView({
                      books, onStatusChange, onNavigate, onRemove, isUpdating,
                  }: {
    books: UserBook[]
    onStatusChange: (copyId: string, status: 'reading' | 'read' | 'want_to_read') => void
    onNavigate: (bookId: string, copyId: string) => void
    onRemove: (b: UserBook) => void
    isUpdating: boolean
}) {
    return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0', border: '1px solid var(--color-border)', borderRadius: 'var(--border-radius)', overflow: 'hidden' }}>
            {/* Header row */}
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '8px 14px', fontSize: '11px', fontWeight: 700, color: 'var(--color-text-muted)', textTransform: 'uppercase', letterSpacing: '0.06em', fontFamily: 'var(--font-body)', background: 'var(--color-surface-alt)', borderBottom: '1px solid var(--color-border)' }}>
                <div style={{ width: '44px', flexShrink: 0 }} />
                <div style={{ flex: 1 }}>Title / Author</div>
                <div style={{ width: '90px', flexShrink: 0 }}>Format</div>
                <div style={{ width: '110px', flexShrink: 0 }}>Status</div>
                <div style={{ width: '70px', flexShrink: 0 }}>Condition</div>
                <div style={{ width: '28px', flexShrink: 0 }} />
            </div>

            {books.map((ub, i) => (
                <ListRow
                    key={ub.copy_id}
                    userBook={ub}
                    isLast={i === books.length - 1}
                    onStatusChange={status => onStatusChange(ub.copy_id, status)}
                    onNavigate={() => onNavigate(ub.book.id, ub.copy_id)}
                    onRemove={() => onRemove(ub)}
                    isUpdating={isUpdating}
                />
            ))}
        </div>
    )
}

function ListRow({
                     userBook, isLast, onStatusChange, onNavigate, onRemove, isUpdating,
                 }: {
    userBook: UserBook
    isLast: boolean
    onStatusChange: (status: 'reading' | 'read' | 'want_to_read') => void
    onNavigate: () => void
    onRemove: () => void
    isUpdating: boolean
}) {
    const { book } = userBook
    const color    = pickColor(book.title)
    const coverSrc = userBook.cover_url

    return (
        <div
            style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '9px 14px', background: 'rgba(0,0,0,0.45)', backgroundImage: 'none', borderImage: 'none', backdropFilter: 'blur(10px)', WebkitBackdropFilter: 'blur(10px)', borderBottom: isLast ? 'none' : '1px solid var(--color-border)', transition: 'background 0.2s ease, backdrop-filter 0.2s ease' }}
            onMouseEnter={e => { const el = e.currentTarget as HTMLDivElement; el.style.background = 'rgba(0,0,0,0.65)'; el.style.backdropFilter = 'blur(2px)'; el.style.backdropFilter = 'blur(2px)' }}
            onMouseLeave={e => { const el = e.currentTarget as HTMLDivElement; el.style.background = 'rgba(0,0,0,0.45)'; el.style.backdropFilter = 'blur(10px)'; el.style.backdropFilter = 'blur(10px)' }}
        >
            {/* Thumbnail — solid color fallback, no CSS background-image bleed */}
            <div
                onClick={onNavigate}
                style={{ width: '32px', height: '44px', flexShrink: 0, borderRadius: '3px', cursor: 'pointer', border: '1px solid var(--color-border)', overflow: 'hidden', background: `linear-gradient(135deg, ${color}dd, ${color}88)`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '16px' }}
            >
                {coverSrc
                    ? <img src={coverSrc} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                    : '📖'}
            </div>

            {/* Title + author */}
            <div onClick={onNavigate} style={{ flex: 1, minWidth: 0, cursor: 'pointer' }}>
                <p style={{ margin: 0, fontSize: '13px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-body)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {book.title}
                </p>
                <p style={{ margin: 0, fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                </p>
            </div>

            {/* Format */}
            <div style={{ width: '90px', flexShrink: 0 }}>
                {userBook.format && <Badge label={userBook.format} variant="muted" size="sm" />}
            </div>

            {/* Status selector */}
            <div style={{ width: '110px', flexShrink: 0, display: 'flex', gap: '2px' }}>
                {(['want_to_read', 'reading', 'read'] as const).map(s => (
                    <button
                        key={s}
                        disabled={isUpdating}
                        onClick={() => onStatusChange(s)}
                        title={STATUS_LABELS[s]}
                        style={{ flex: 1, padding: '3px 0', border: `1px solid ${userBook.reading_status === s ? 'var(--color-primary)' : 'var(--color-border)'}`, borderRadius: 'var(--border-radius)', background: userBook.reading_status === s ? 'var(--color-primary)' : 'transparent', color: userBook.reading_status === s ? 'var(--color-primary-text)' : 'var(--color-text-muted)', fontSize: '11px', cursor: isUpdating ? 'not-allowed' : 'pointer', fontFamily: 'var(--font-body)', transition: 'var(--transition)' }}
                    >
                        {s === 'want_to_read' ? '🔖' : s === 'reading' ? '📖' : '✅'}
                    </button>
                ))}
            </div>

            {/* Condition */}
            <div style={{ width: '70px', flexShrink: 0, fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', textTransform: 'capitalize' }}>
                {userBook.condition ?? '—'}
            </div>

            {/* Remove */}
            <button
                onClick={onRemove}
                style={{ width: '28px', height: '28px', flexShrink: 0, background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-error)', fontSize: '14px', opacity: 0.5, transition: 'var(--transition)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.opacity = '1'}
                onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.opacity = '0.5'}
                title="Remove from library"
            >
                🗑️
            </button>
        </div>
    )
}

// ─── Library Book Card (card view) ───────────────────────────────────────────

function LibraryBookCard({
                             userBook, onStatusChange, onNavigate, onRemove, isUpdating,
                         }: {
    userBook: UserBook
    onStatusChange: (status: 'reading' | 'read' | 'want_to_read') => void
    onNavigate: () => void
    onRemove: () => void
    isUpdating: boolean
}) {
    const { book } = userBook
    const color    = pickColor(book.title)
    const coverSrc = userBook.cover_url

    return (
        <Card padding="sm" hover>
            <div style={{ display: 'flex', gap: '12px' }}>
                {/* Cover */}
                <div
                    onClick={onNavigate}
                    style={{ width: '56px', height: '76px', borderRadius: '4px', flexShrink: 0, border: '1px solid var(--color-border)', cursor: 'pointer', overflow: 'hidden', background: `linear-gradient(135deg, ${color}dd, ${color}88)`, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '22px' }}
                >
                    {coverSrc
                        ? <img src={coverSrc} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
                        : '📖'}
                </div>

                {/* Info */}
                <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: '6px' }}>
                    <p onClick={onNavigate} style={{ margin: 0, fontSize: '13px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-body)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', cursor: 'pointer' }}>
                        {book.title}
                    </p>
                    <p style={{ margin: 0, fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                        {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                    </p>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                        <Badge label={STATUS_LABELS[userBook.reading_status]} variant={STATUS_VARIANTS[userBook.reading_status]} size="sm" />
                        {userBook.format && <Badge label={userBook.format} variant="muted" size="sm" />}
                    </div>
                </div>
            </div>

            {/* Status selector */}
            <div style={{ marginTop: '10px', paddingTop: '10px', borderTop: '1px solid var(--color-border)', display: 'flex', gap: '4px', justifyContent: 'space-between', alignItems: 'center' }}>
                <div style={{ display: 'flex', gap: '4px' }}>
                    {(['want_to_read', 'reading', 'read'] as const).map(status => (
                        <button
                            key={status}
                            disabled={isUpdating}
                            onClick={() => onStatusChange(status)}
                            style={{ padding: '4px 8px', border: `1px solid ${userBook.reading_status === status ? 'var(--color-primary)' : 'var(--color-border)'}`, borderRadius: 'var(--border-radius)', background: userBook.reading_status === status ? 'var(--color-primary)' : 'transparent', color: userBook.reading_status === status ? 'var(--color-primary-text)' : 'var(--color-text-muted)', fontSize: '10px', cursor: isUpdating ? 'not-allowed' : 'pointer', fontFamily: 'var(--font-body)', fontWeight: userBook.reading_status === status ? 600 : 400, transition: 'var(--transition)', whiteSpace: 'nowrap' }}
                        >
                            {status === 'want_to_read' ? '🔖' : status === 'reading' ? '📖' : '✅'}
                        </button>
                    ))}
                </div>
                <button
                    onClick={onRemove}
                    style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-error)', fontSize: '14px', padding: '4px', opacity: 0.6, transition: 'var(--transition)' }}
                    onMouseEnter={e => (e.currentTarget as HTMLButtonElement).style.opacity = '1'}
                    onMouseLeave={e => (e.currentTarget as HTMLButtonElement).style.opacity = '0.6'}
                    title="Remove from library"
                >
                    🗑️
                </button>
            </div>
        </Card>
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

function PageBtn({ label, onClick, active = false, disabled = false }: { label: string; onClick: () => void; active?: boolean; disabled?: boolean }) {
    return (
        <button
            onClick={onClick}
            disabled={disabled}
            style={{ minWidth: '36px', height: '36px', padding: '0 10px', background: active ? 'var(--color-primary)' : 'var(--color-surface)', color: active ? 'var(--color-primary-text)' : 'var(--color-text)', border: `1px solid ${active ? 'var(--color-primary)' : 'var(--color-border)'}`, borderRadius: 'var(--border-radius)', fontSize: '13px', fontWeight: active ? 600 : 400, cursor: disabled ? 'not-allowed' : 'pointer', opacity: disabled ? 0.4 : 1, transition: 'var(--transition)', fontFamily: 'var(--font-body)' }}
        >
            {label}
        </button>
    )
}