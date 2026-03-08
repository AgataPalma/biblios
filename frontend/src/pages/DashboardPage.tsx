import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '../context/AuthContext'
import { Spinner, EmptyState } from '../components'
import { getMyLibrary, listBooks } from '../api/books'
import type { UserBook } from '../types'

// ── Book cover placeholder ─────────────────────────────────────────────────────
// When cover_url is available it renders as an <img>; otherwise a tinted rect.
function BookCover({
                       coverUrl,
                       title,
                       height,
                       selected = false,
                       onClick,
                   }: {
    coverUrl?: string
    title: string
    height: number
    selected?: boolean
    onClick?: () => void
}) {
    const width = Math.floor(height * 0.65)

    const baseStyle: React.CSSProperties = {
        width,
        height,
        borderRadius: '3px 5px 5px 3px',
        flexShrink: 0,
        cursor: onClick ? 'pointer' : 'default',
        transition: 'box-shadow 0.15s, transform 0.15s',
        boxShadow: selected
            ? '0 0 0 2px var(--color-primary), 2px 3px 10px rgba(0,0,0,0.5)'
            : 'inset -2px 0 4px rgba(0,0,0,0.25), 2px 3px 8px rgba(0,0,0,0.35)',
        overflow: 'hidden',
        position: 'relative',
    }

    const handleEnter = (e: React.MouseEvent<HTMLDivElement>) => {
        if (onClick) (e.currentTarget as HTMLDivElement).style.transform = 'translateY(-3px)'
    }
    const handleLeave = (e: React.MouseEvent<HTMLDivElement>) => {
        if (onClick) (e.currentTarget as HTMLDivElement).style.transform = 'translateY(0)'
    }

    if (coverUrl) {
        return (
            <div style={baseStyle} title={title} onClick={onClick} onMouseEnter={handleEnter} onMouseLeave={handleLeave}>
                <img src={coverUrl} alt={title} style={{ width: '100%', height: '100%', objectFit: 'cover', display: 'block' }} />
            </div>
        )
    }

    // Fallback: deterministic colour from title
    const palette = ['#7c6f5e','#5a7a8a','#7a5a6a','#6a7a5a','#8a6a5a','#6a5a8a','#8a7a5a','#5a6a7a']
    const color = palette[title.charCodeAt(0) % palette.length]

    return (
        <div
            style={{ ...baseStyle, background: `linear-gradient(to right, ${color}cc, ${color}, ${color}dd)` }}
            title={title}
            onClick={onClick}
            onMouseEnter={handleEnter}
            onMouseLeave={handleLeave}
        >
            <div style={{ position: 'absolute', left: '3px', top: 0, bottom: 0, width: '2px', background: 'rgba(255,255,255,0.08)' }} />
        </div>
    )
}

// ── Status pill ────────────────────────────────────────────────────────────────
function StatusPill({ status }: { status: string }) {
    const map: Record<string, { bg: string; text: string; label: string }> = {
        reading:      { bg: 'color-mix(in srgb, var(--color-success) 15%, transparent)', text: 'var(--color-success)',  label: 'Reading'       },
        read:         { bg: 'color-mix(in srgb, var(--color-primary) 15%, transparent)', text: 'var(--color-primary)', label: 'Read'          },
        want_to_read: { bg: 'color-mix(in srgb, var(--color-accent)  15%, transparent)', text: 'var(--color-accent)',  label: 'Want to read'  },
    }
    const s = map[status] ?? map.read
    return (
        <span style={{
            padding: '2px 8px', borderRadius: '999px',
            fontSize: '10px', fontWeight: 600,
            background: s.bg, color: s.text,
            fontFamily: 'var(--font-body)', whiteSpace: 'nowrap',
        }}>
            {s.label}
        </span>
    )
}

// ── Main page ──────────────────────────────────────────────────────────────────
export default function DashboardPage() {
    const { user } = useAuth()
    const navigate = useNavigate()

    const { data: libraryData, isLoading: loadingLibrary } = useQuery({
        queryKey: ['my-library-dashboard'],
        queryFn: () => getMyLibrary(1, 50),
    })

    const { data: catalogueData, isLoading: loadingCatalogue } = useQuery({
        queryKey: ['books-dashboard'],
        queryFn: () => listBooks(1, 6),
    })

    const allBooks: UserBook[] = libraryData?.books ?? []
    const inProgress  = allBooks.filter(b => b.reading_status === 'reading')
    const recentBooks = allBooks.slice(0, 5)

    const stats = {
        reading:     allBooks.filter(b => b.reading_status === 'reading').length,
        read:        allBooks.filter(b => b.reading_status === 'read').length,
        wantToRead:  allBooks.filter(b => b.reading_status === 'want_to_read').length,
    }

    const [selectedId, setSelectedId] = useState<string | null>(null)
    const selectedBook = selectedId
        ? inProgress.find(b => b.copy_id === selectedId)
        : inProgress[0] ?? null

    const hour = new Date().getHours()
    const greeting = hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening'

    return (
        <div style={{
            minHeight: '100vh',
            padding: '32px 40px',
            maxWidth: '1400px',
            margin: '0 auto',
        }}>

            {/* Greeting */}
            <div style={{ marginBottom: '28px' }}>
                <h1 style={{
                    margin: 0,
                    fontSize: '26px',
                    fontWeight: 400,
                    color: 'var(--color-text)',
                    fontFamily: 'var(--font-heading)',
                    fontStyle: 'italic',
                }}>
                    {greeting}, <strong style={{ fontStyle: 'normal' }}>{user?.username}</strong> 👋
                </h1>
                <p style={{
                    margin: '6px 0 0',
                    fontSize: '13px',
                    color: 'var(--color-text-muted)',
                    fontFamily: 'var(--font-body)',
                }}>
                    Here's where your reading stands today.
                </p>
            </div>

            {/* ── Currently reading hero ─────────────────────────────── */}
            {loadingLibrary ? (
                <div style={{
                    background: 'var(--color-surface)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--border-radius)',
                    padding: '32px',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    marginBottom: '20px',
                }}>
                    <Spinner size="sm" />
                </div>
            ) : inProgress.length === 0 ? (
                <div style={{
                    background: 'var(--color-surface)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--border-radius)',
                    padding: '32px',
                    marginBottom: '20px',
                }}>
                    <EmptyState
                        icon="📖"
                        title="Nothing in progress"
                        description="Add a book to your library and set it to 'Reading' to see it here."
                        action={{ label: 'Browse catalogue', onClick: () => navigate('/books') }}
                    />
                </div>
            ) : (
                <div style={{
                    position: 'relative',
                    overflow: 'hidden',
                    background: 'var(--color-surface)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--border-radius)',
                    padding: '28px',
                    display: 'flex',
                    gap: '28px',
                    alignItems: 'center',
                    marginBottom: '20px',
                    // Subtle ambient tint from selected book — CSS can't read JS vars so we use a fixed overlay
                    backgroundImage: 'linear-gradient(135deg, color-mix(in srgb, var(--color-primary) 6%, transparent) 0%, transparent 60%)',
                }}>
                    {/* Ambient glow orb */}
                    <div style={{
                        position: 'absolute', right: '-40px', top: '-40px',
                        width: '200px', height: '200px', borderRadius: '50%',
                        background: 'radial-gradient(circle, color-mix(in srgb, var(--color-primary) 12%, transparent), transparent 70%)',
                        pointerEvents: 'none',
                    }} />

                    {/* Cascading book covers — one per in-progress book */}
                    <div style={{ display: 'flex', gap: '8px', alignItems: 'flex-end', flexShrink: 0 }}>
                        {inProgress.map((book, i) => {
                            const isSelected = book.copy_id === (selectedBook?.copy_id ?? inProgress[0]?.copy_id)
                            const selectedIdx = inProgress.findIndex(
                                b => b.copy_id === (selectedBook?.copy_id ?? inProgress[0]?.copy_id)
                            )
                            const distance = Math.abs(i - selectedIdx)
                            const height = isSelected ? 80 : Math.max(80 - distance * 20, 44)
                            return (
                                <div key={book.copy_id} style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '6px' }}>
                                    <BookCover
                                        coverUrl={book.book.cover_url}
                                        title={book.book.title}
                                        height={height}
                                        selected={isSelected}
                                        onClick={() => setSelectedId(book.copy_id)}
                                    />
                                    {inProgress.length > 1 && (
                                        <div style={{
                                            width: '4px', height: '4px', borderRadius: '50%',
                                            background: isSelected ? 'var(--color-primary)' : 'var(--color-text-muted)',
                                            transition: 'background 0.15s',
                                        }} />
                                    )}
                                </div>
                            )
                        })}
                    </div>

                    {/* Selected book info */}
                    {selectedBook && (
                        <div style={{ flex: 1, minWidth: 0 }}>
                            <p style={{
                                margin: '0 0 6px',
                                fontSize: '10px',
                                color: 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                                letterSpacing: '0.12em',
                                textTransform: 'uppercase',
                            }}>
                                Currently reading
                                {inProgress.length > 1 && ` · ${inProgress.findIndex(b => b.copy_id === selectedBook.copy_id) + 1} of ${inProgress.length}`}
                            </p>
                            <h2 style={{
                                margin: '0 0 3px',
                                fontSize: '22px',
                                fontWeight: 700,
                                color: 'var(--color-text)',
                                fontFamily: 'var(--font-heading)',
                                lineHeight: 1.2,
                            }}>
                                {selectedBook.book.title}
                            </h2>
                            <p style={{
                                margin: '0 0 6px',
                                fontSize: '13px',
                                color: 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                            }}>
                                {selectedBook.book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                            </p>
                            <p style={{
                                margin: 0,
                                fontSize: '11px',
                                color: 'var(--color-text-muted)',
                                fontFamily: 'var(--font-body)',
                                fontStyle: 'italic',
                            }}>
                                {selectedBook.format} edition
                                {selectedBook.language && ` · ${selectedBook.language}`}
                            </p>
                        </div>
                    )}

                    {/* In-progress count badge */}
                    <div style={{
                        flexShrink: 0,
                        textAlign: 'center',
                        padding: '16px 20px',
                        background: 'var(--color-surface-alt)',
                        borderRadius: 'var(--border-radius)',
                        border: '1px solid var(--color-border)',
                    }}>
                        <p style={{
                            margin: 0,
                            fontSize: '28px',
                            fontWeight: 700,
                            color: 'var(--color-primary)',
                            fontFamily: 'var(--font-heading)',
                            lineHeight: 1,
                        }}>
                            {stats.reading}
                        </p>
                        <p style={{
                            margin: '4px 0 0',
                            fontSize: '10px',
                            color: 'var(--color-text-muted)',
                            fontFamily: 'var(--font-body)',
                            textTransform: 'uppercase',
                            letterSpacing: '0.08em',
                        }}>
                            in progress
                        </p>
                    </div>
                </div>
            )}

            {/* ── Stats row ─────────────────────────────────────────── */}
            <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(3, 1fr)',
                gap: '12px',
                marginBottom: '20px',
            }}>
                {[
                    { label: 'Reading',      value: loadingLibrary ? '…' : String(stats.reading),    icon: '📖', accent: 'var(--color-success)' },
                    { label: 'Finished',     value: loadingLibrary ? '…' : String(stats.read),       icon: '✅', accent: 'var(--color-primary)' },
                    { label: 'Want to read', value: loadingLibrary ? '…' : String(stats.wantToRead), icon: '🔖', accent: 'var(--color-accent)'  },
                ].map(s => (
                    <div
                        key={s.label}
                        onClick={() => navigate('/library')}
                        style={{
                            background: 'var(--color-surface)',
                            border: '1px solid var(--color-border)',
                            borderRadius: 'var(--border-radius)',
                            padding: '16px 20px',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '14px',
                            cursor: 'pointer',
                            transition: 'var(--transition)',
                        }}
                        onMouseEnter={e => (e.currentTarget as HTMLDivElement).style.borderColor = s.accent}
                        onMouseLeave={e => (e.currentTarget as HTMLDivElement).style.borderColor = 'var(--color-border)'}
                    >
                        <div style={{
                            width: '38px', height: '38px',
                            borderRadius: 'var(--border-radius)',
                            background: `color-mix(in srgb, ${s.accent} 15%, transparent)`,
                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                            fontSize: '18px', flexShrink: 0,
                        }}>
                            {s.icon}
                        </div>
                        <div>
                            <p style={{ margin: 0, fontSize: '24px', fontWeight: 700, color: s.accent, fontFamily: 'var(--font-heading)', lineHeight: 1 }}>
                                {s.value}
                            </p>
                            <p style={{ margin: '3px 0 0', fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                                {s.label}
                            </p>
                        </div>
                    </div>
                ))}
            </div>

            {/* ── Two-column: My library + Catalogue ────────────────── */}
            <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(420px, 1fr))',
                gap: '20px',
            }}>
                {/* My library */}
                <div style={{
                    background: 'var(--color-surface)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--border-radius)',
                    overflow: 'hidden',
                }}>
                    <div style={{
                        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                        padding: '14px 20px',
                        borderBottom: '1px solid var(--color-border)',
                    }}>
                        <h2 style={{ margin: 0, fontSize: '14px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-heading)' }}>
                            My Library
                        </h2>
                        <button
                            onClick={() => navigate('/library')}
                            style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '12px', color: 'var(--color-primary)', fontFamily: 'var(--font-body)' }}
                        >
                            View all →
                        </button>
                    </div>

                    {/* Shelf strip */}
                    {recentBooks.length > 0 && (
                        <div style={{
                            padding: '12px 20px 10px',
                            display: 'flex', gap: '8px', alignItems: 'flex-end',
                            borderBottom: '1px solid var(--color-border)',
                            background: 'color-mix(in srgb, var(--color-surface-alt) 60%, transparent)',
                        }}>
                            {recentBooks.map(book => (
                                <BookCover
                                    key={book.copy_id}
                                    coverUrl={book.book.cover_url}
                                    title={book.book.title}
                                    height={44}
                                    onClick={() => navigate(`/books/${book.book.id}`)}
                                />
                            ))}
                            {allBooks.length > 5 && (
                                <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)', alignSelf: 'center', marginLeft: '4px' }}>
                                    +{allBooks.length - 5} more
                                </span>
                            )}
                        </div>
                    )}

                    {loadingLibrary ? (
                        <div style={{ padding: '24px', display: 'flex', justifyContent: 'center' }}>
                            <Spinner size="sm" />
                        </div>
                    ) : recentBooks.length === 0 ? (
                        <div style={{ padding: '16px' }}>
                            <EmptyState
                                icon="📭"
                                title="No books yet"
                                description="Add your first book to get started."
                                action={{ label: 'Browse catalogue', onClick: () => navigate('/books') }}
                            />
                        </div>
                    ) : (
                        recentBooks.map((book, i) => (
                            <div
                                key={book.copy_id}
                                onClick={() => navigate(`/books/${book.book.id}`)}
                                style={{
                                    display: 'flex', alignItems: 'center', gap: '12px',
                                    padding: '10px 20px',
                                    borderBottom: i < recentBooks.length - 1 ? '1px solid var(--color-border)' : 'none',
                                    cursor: 'pointer', transition: 'var(--transition)',
                                }}
                                onMouseEnter={e => (e.currentTarget as HTMLDivElement).style.background = 'var(--color-surface-alt)'}
                                onMouseLeave={e => (e.currentTarget as HTMLDivElement).style.background = 'transparent'}
                            >
                                <BookCover
                                    coverUrl={book.book.cover_url}
                                    title={book.book.title}
                                    height={36}
                                />
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <p style={{ margin: 0, fontSize: '13px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-body)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                        {book.book.title}
                                    </p>
                                    <p style={{ margin: '2px 0 0', fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                                        {book.book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                                    </p>
                                </div>
                                <StatusPill status={book.reading_status} />
                            </div>
                        ))
                    )}
                </div>

                {/* Recently added to catalogue */}
                <div style={{
                    background: 'var(--color-surface)',
                    border: '1px solid var(--color-border)',
                    borderRadius: 'var(--border-radius)',
                    overflow: 'hidden',
                }}>
                    <div style={{
                        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                        padding: '14px 20px',
                        borderBottom: '1px solid var(--color-border)',
                    }}>
                        <h2 style={{ margin: 0, fontSize: '14px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-heading)' }}>
                            Recently Added
                        </h2>
                        <button
                            onClick={() => navigate('/books')}
                            style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '12px', color: 'var(--color-primary)', fontFamily: 'var(--font-body)' }}
                        >
                            Browse all →
                        </button>
                    </div>

                    {loadingCatalogue ? (
                        <div style={{ padding: '24px', display: 'flex', justifyContent: 'center' }}>
                            <Spinner size="sm" />
                        </div>
                    ) : !catalogueData?.books?.length ? (
                        <div style={{ padding: '16px' }}>
                            <EmptyState
                                icon="🔍"
                                title="Catalogue is empty"
                                description="Be the first to submit a book."
                                action={{ label: 'Submit a book', onClick: () => navigate('/books/add') }}
                            />
                        </div>
                    ) : (
                        catalogueData.books.map((book, i) => (
                            <div
                                key={book.id}
                                onClick={() => navigate(`/books/${book.id}`)}
                                style={{
                                    display: 'flex', alignItems: 'center', gap: '12px',
                                    padding: '10px 20px',
                                    borderBottom: i < catalogueData.books.length - 1 ? '1px solid var(--color-border)' : 'none',
                                    cursor: 'pointer', transition: 'var(--transition)',
                                }}
                                onMouseEnter={e => (e.currentTarget as HTMLDivElement).style.background = 'var(--color-surface-alt)'}
                                onMouseLeave={e => (e.currentTarget as HTMLDivElement).style.background = 'transparent'}
                            >
                                <BookCover
                                    coverUrl={book.cover_url}
                                    title={book.title}
                                    height={36}
                                />
                                <div style={{ flex: 1, minWidth: 0 }}>
                                    <p style={{ margin: 0, fontSize: '13px', fontWeight: 600, color: 'var(--color-text)', fontFamily: 'var(--font-body)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                                        {book.title}
                                    </p>
                                    <p style={{ margin: '2px 0 0', fontSize: '11px', color: 'var(--color-text-muted)', fontFamily: 'var(--font-body)' }}>
                                        {book.authors?.map(a => a.name).join(', ') || 'Unknown author'}
                                    </p>
                                </div>
                            </div>
                        ))
                    )}
                </div>
            </div>
        </div>
    )
}